#!/usr/bin/env bash
# Tear down resources installed by setup-prerequisites.sh.
set -euo pipefail

DEFAULT_CLUSTER_NAME="${DEFAULT_CLUSTER_NAME:-perses-demo}"
CLUSTER_NAME="${CLUSTER_NAME:-}"
SKIP_KIND="${SKIP_KIND:-false}"
DELETE_KIND_CLUSTER="${DELETE_KIND_CLUSTER:-}"
YES="${YES:-false}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required tool not found: $1" >&2
    exit 1
  fi
}

confirm() {
  local prompt="$1"
  if [[ "${YES}" == "true" ]]; then
    return 0
  fi
  if [[ ! -t 0 ]]; then
    echo "error: non-interactive shell; re-run with YES=true or answer prompts in a terminal" >&2
    exit 1
  fi
  local reply
  read -r -p "${prompt} [y/N] " reply
  [[ "${reply}" =~ ^[Yy]$ ]]
}

helm_uninstall() {
  local release="$1"
  local ns="$2"
  if helm status "${release}" -n "${ns}" >/dev/null 2>&1; then
    echo "==> helm uninstall ${release} -n ${ns}"
    helm uninstall "${release}" -n "${ns}" || true
  else
    echo "==> skip ${release} (not found in ${ns})"
  fi
}

choose_target() {
  need kind

  local clusters=()
  local line
  while IFS= read -r line; do
    [[ -n "${line}" ]] && clusters+=("${line}")
  done < <(kind get clusters 2>/dev/null || true)

  local current_ctx
  current_ctx="$(kubectl config current-context 2>/dev/null || echo "(none)")"

  echo
  echo "Select where to clean up:"
  echo "  Current kubectl context: ${current_ctx}"
  echo

  local i=1
  local c
  for c in "${clusters[@]+"${clusters[@]}"}"; do
    echo "  ${i}) kind cluster: ${c}"
    i=$((i + 1))
  done
  echo "  ${i}) Use current kubectl context"
  local current_opt="${i}"
  echo

  local choice
  if [[ "${YES}" == "true" ]]; then
    if [[ "${SKIP_KIND}" == "true" ]]; then
      choice="${current_opt}"
    elif [[ -n "${CLUSTER_NAME}" ]]; then
      choice=""
      local idx=1
      for c in "${clusters[@]+"${clusters[@]}"}"; do
        if [[ "${c}" == "${CLUSTER_NAME}" ]]; then
          choice="${idx}"
          break
        fi
        idx=$((idx + 1))
      done
      if [[ -z "${choice}" ]]; then
        echo "error: kind cluster not found: ${CLUSTER_NAME}" >&2
        exit 1
      fi
    elif ((${#clusters[@]} > 0)); then
      choice=1
      CLUSTER_NAME="${clusters[0]}"
    else
      choice="${current_opt}"
    fi
  else
    if [[ ! -t 0 ]]; then
      echo "error: non-interactive shell; set CLUSTER_NAME=... or SKIP_KIND=true and YES=true" >&2
      exit 1
    fi
    read -r -p "Choice [1-${i}]: " choice
  fi

  if [[ "${choice}" == "${current_opt}" ]]; then
    SKIP_KIND="true"
    CLUSTER_NAME=""
    return 0
  fi

  if [[ "${choice}" =~ ^[0-9]+$ ]] && (( choice >= 1 && choice < current_opt )); then
    CLUSTER_NAME="${clusters[$((choice - 1))]}"
    SKIP_KIND="false"
    return 0
  fi

  echo "error: invalid choice: ${choice}" >&2
  exit 1
}

choose_delete_kind() {
  if [[ "${SKIP_KIND}" == "true" ]]; then
    DELETE_KIND_CLUSTER="false"
    return 0
  fi
  if [[ "${DELETE_KIND_CLUSTER}" == "true" || "${DELETE_KIND_CLUSTER}" == "false" ]]; then
    return 0
  fi
  if [[ "${YES}" == "true" ]]; then
    DELETE_KIND_CLUSTER="true"
    return 0
  fi
  echo
  echo "Cleanup mode for kind cluster ${CLUSTER_NAME}:"
  echo "  1) Delete the entire kind cluster (recommended for demos)"
  echo "  2) Uninstall Helm releases / namespaces only (keep the cluster)"
  echo
  local choice
  read -r -p "Choice [1-2] (default 1): " choice
  choice="${choice:-1}"
  case "${choice}" in
    1) DELETE_KIND_CLUSTER="true" ;;
    2) DELETE_KIND_CLUSTER="false" ;;
    *)
      echo "error: invalid choice: ${choice}" >&2
      exit 1
      ;;
  esac
}

uninstall_stack() {
  echo "==> Removing Argo CD Application / namespace"
  kubectl delete application perses-dashboards -n argocd --ignore-not-found=true 2>/dev/null || true
  kubectl delete namespace argocd --ignore-not-found=true --wait=false 2>/dev/null || true

  echo "==> Removing Perses resources / namespace"
  kubectl delete persesdashboard --all -n perses-dev --ignore-not-found=true 2>/dev/null || true
  kubectl delete persesdatasource --all -n perses-dev --ignore-not-found=true 2>/dev/null || true
  kubectl delete perses --all -n perses-dev --ignore-not-found=true 2>/dev/null || true
  kubectl delete namespace perses-dev --ignore-not-found=true --wait=false 2>/dev/null || true

  helm_uninstall kube-prometheus-stack monitoring
  kubectl delete namespace monitoring --ignore-not-found=true --wait=false 2>/dev/null || true

  helm_uninstall perses-operator perses-operator-system
  kubectl delete namespace perses-operator-system --ignore-not-found=true --wait=false 2>/dev/null || true

  helm_uninstall cert-manager cert-manager
  kubectl delete namespace cert-manager --ignore-not-found=true --wait=false 2>/dev/null || true
}

need kubectl
need helm

if [[ "${SKIP_KIND}" == "true" && "${YES}" == "true" ]]; then
  :
elif [[ -n "${CLUSTER_NAME}" && "${YES}" == "true" ]]; then
  SKIP_KIND="false"
else
  choose_target
fi

choose_delete_kind

echo
if [[ "${SKIP_KIND}" == "true" ]]; then
  echo "Will clean up current kubectl context:"
  kubectl config current-context
elif [[ "${DELETE_KIND_CLUSTER}" == "true" ]]; then
  echo "Will delete kind cluster: ${CLUSTER_NAME}"
else
  echo "Will uninstall demo stack from kind cluster: ${CLUSTER_NAME} (cluster kept)"
  echo "  (Argo CD, Perses, kube-prometheus-stack, perses-operator, cert-manager)"
fi
echo

if ! confirm "Proceed with cleanup?"; then
  echo "Aborted."
  exit 0
fi

if [[ "${SKIP_KIND}" != "true" ]]; then
  need kind
  if ! kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
    echo "error: kind cluster not found: ${CLUSTER_NAME}" >&2
    exit 1
  fi
  kind export kubeconfig --name "${CLUSTER_NAME}" >/dev/null
  kubectl config use-context "kind-${CLUSTER_NAME}" >/dev/null
fi

if [[ "${SKIP_KIND}" != "true" && "${DELETE_KIND_CLUSTER}" == "true" ]]; then
  echo "==> kind delete cluster --name ${CLUSTER_NAME}"
  kind delete cluster --name "${CLUSTER_NAME}"
  echo
  echo "Cleanup complete (kind cluster deleted)."
  exit 0
fi

uninstall_stack

echo
echo "Cleanup complete."
echo "  Remaining namespaces may take a minute to terminate."
echo "  Check: kubectl get ns"
