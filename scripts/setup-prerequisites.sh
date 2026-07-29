#!/usr/bin/env bash
# Install cluster prerequisites using the same layout as
# https://github.com/slashpai/perses-operator-examples
# (kind, cert-manager, perses-operator, minimal kube-prometheus-stack,
#  Perses server in perses-dev, Prometheus datasource).
#
set -euo pipefail

DEFAULT_CLUSTER_NAME="${DEFAULT_CLUSTER_NAME:-perses-demo}"
CLUSTER_NAME="${CLUSTER_NAME:-}"
SKIP_KIND="${SKIP_KIND:-false}"
YES="${YES:-false}"
EXAMPLES_REF="${EXAMPLES_REF:-main}"
EXAMPLES_RAW="https://raw.githubusercontent.com/slashpai/perses-operator-examples/${EXAMPLES_REF}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBE_PROM_VALUES="${SCRIPT_DIR}/kube-prometheus-values.yaml"

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
  echo "Select where to install prerequisites:"
  echo "  Current kubectl context: ${current_ctx}"
  echo
  echo "  1) Create new kind cluster (${DEFAULT_CLUSTER_NAME})  [recommended]"
  local create_opt=1
  local i=2
  local c
  for c in "${clusters[@]+"${clusters[@]}"}"; do
    echo "  ${i}) Use existing kind cluster: ${c}"
    i=$((i + 1))
  done
  echo "  ${i}) Use current kubectl context (no kind create/export)"
  local current_opt="${i}"
  echo

  local choice
  if [[ "${YES}" == "true" ]]; then
    if [[ "${SKIP_KIND}" == "true" ]]; then
      choice="${current_opt}"
    elif [[ -n "${CLUSTER_NAME}" ]]; then
      # Prefer create/reuse of CLUSTER_NAME without the menu.
      SKIP_KIND="false"
      return 0
    else
      choice="${create_opt}"
      CLUSTER_NAME="${DEFAULT_CLUSTER_NAME}"
    fi
  else
    if [[ ! -t 0 ]]; then
      echo "error: non-interactive shell; set CLUSTER_NAME=... or SKIP_KIND=true and YES=true" >&2
      exit 1
    fi
    read -r -p "Choice [1-${i}] (default 1): " choice
    choice="${choice:-1}"
  fi

  if [[ "${choice}" == "${create_opt}" ]]; then
    SKIP_KIND="false"
    if [[ -z "${CLUSTER_NAME}" ]]; then
      if [[ "${YES}" == "true" ]]; then
        CLUSTER_NAME="${DEFAULT_CLUSTER_NAME}"
      else
        read -r -p "New cluster name [${DEFAULT_CLUSTER_NAME}]: " CLUSTER_NAME
        CLUSTER_NAME="${CLUSTER_NAME:-${DEFAULT_CLUSTER_NAME}}"
      fi
    fi
    return 0
  fi

  if [[ "${choice}" == "${current_opt}" ]]; then
    SKIP_KIND="true"
    CLUSTER_NAME=""
    return 0
  fi

  # Existing kind clusters are options 2 .. (current_opt - 1)
  if [[ "${choice}" =~ ^[0-9]+$ ]] && (( choice >= 2 && choice < current_opt )); then
    CLUSTER_NAME="${clusters[$((choice - 2))]}"
    SKIP_KIND="false"
    return 0
  fi

  echo "error: invalid choice: ${choice}" >&2
  exit 1
}

need kubectl
need helm

# Non-interactive:
#   SKIP_KIND=true YES=true ./scripts/setup-prerequisites.sh
#   CLUSTER_NAME=perses-demo YES=true ./scripts/setup-prerequisites.sh
if [[ "${SKIP_KIND}" == "true" && "${YES}" == "true" ]]; then
  :
elif [[ -n "${CLUSTER_NAME}" && "${YES}" == "true" ]]; then
  SKIP_KIND="false"
else
  choose_target
fi

echo
if [[ "${SKIP_KIND}" == "true" ]]; then
  echo "Will install into current kubectl context:"
  kubectl config current-context
  kubectl cluster-info | head -n 1 || true
else
  echo "Will use kind cluster: ${CLUSTER_NAME}"
fi
echo
echo "This will install (if missing):"
echo "  - cert-manager"
echo "  - perses-operator"
echo "  - minimal kube-prometheus-stack (operator, Prometheus, Alertmanager, node-exporter;"
echo "  - Perses (perses-dev) + prometheus-datasource"
echo

if ! confirm "Proceed?"; then
  echo "Aborted."
  exit 0
fi

echo "==> Helm repos"
helm repo add jetstack https://charts.jetstack.io >/dev/null 2>&1 || true
helm repo add perses https://perses.github.io/helm-charts >/dev/null 2>&1 || true
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null 2>&1 || true
helm repo update

if [[ "${SKIP_KIND}" != "true" ]]; then
  need kind
  if ! kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
    echo "==> Creating kind cluster ${CLUSTER_NAME}"
    kind create cluster --name "${CLUSTER_NAME}"
  else
    echo "==> kind cluster ${CLUSTER_NAME} already exists — exporting kubeconfig"
    kind export kubeconfig --name "${CLUSTER_NAME}"
  fi
  kubectl cluster-info --context "kind-${CLUSTER_NAME}" >/dev/null
  kubectl config use-context "kind-${CLUSTER_NAME}" >/dev/null
fi

echo "==> cert-manager"
if ! helm status cert-manager -n cert-manager >/dev/null 2>&1; then
  helm install cert-manager jetstack/cert-manager \
    --namespace cert-manager \
    --create-namespace \
    --set crds.enabled=true
fi
kubectl wait --for=condition=Ready pods --all -n cert-manager --timeout=180s

echo "==> perses-operator"
if ! helm status perses-operator -n perses-operator-system >/dev/null 2>&1; then
  helm install perses-operator perses/perses-operator \
    --namespace perses-operator-system \
    --create-namespace
fi
kubectl wait --for=condition=Ready pods --all -n perses-operator-system --timeout=180s

echo "==> kube-prometheus-stack (minimal)"
if ! helm status kube-prometheus-stack -n monitoring >/dev/null 2>&1; then
  helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
    --namespace monitoring \
    --create-namespace \
    -f "${KUBE_PROM_VALUES}"
else
  echo "    release already exists — leaving it unchanged"
  echo "    to apply minimal values: helm upgrade kube-prometheus-stack prometheus-community/kube-prometheus-stack -n monitoring -f ${KUBE_PROM_VALUES}"
fi

# Operator creates Prometheus/Alertmanager pods asynchronously — wait for each.
echo "==> Waiting for kube-prometheus-stack pods"
kubectl wait --for=condition=Ready pod \
  -l app.kubernetes.io/name=prometheus-operator \
  -n monitoring --timeout=180s
for _ in $(seq 1 60); do
  if kubectl get pods -n monitoring -l app.kubernetes.io/name=prometheus --no-headers 2>/dev/null | grep -q .; then
    break
  fi
  sleep 5
done
kubectl wait --for=condition=Ready pod \
  -l app.kubernetes.io/name=prometheus \
  -n monitoring --timeout=300s
kubectl wait --for=condition=Ready pod \
  -l app.kubernetes.io/name=alertmanager \
  -n monitoring --timeout=180s
kubectl wait --for=condition=Ready pod \
  -l app.kubernetes.io/name=prometheus-node-exporter \
  -n monitoring --timeout=180s

echo "==> perses-dev namespace + Perses server + datasource"
kubectl create namespace perses-dev --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f "${EXAMPLES_RAW}/basic/perses.yaml"
kubectl apply -f "${EXAMPLES_RAW}/kube-prometheus/prometheus-datasource.yaml"

echo "==> Waiting for Perses to become Available"
kubectl wait --for=condition=Available perses/perses-sample -n perses-dev --timeout=180s || \
  echo "warning: Perses not Available yet — check: kubectl get perses -n perses-dev"

echo
echo "Prerequisites ready (same config as perses-operator-examples)."
echo "  kubectl context:     $(kubectl config current-context 2>/dev/null || echo unknown)"
echo "  Perses namespace:     perses-dev"
echo "  Datasource name:     prometheus-datasource"
echo "  Prometheus:          monitoring (minimal stack)"
echo
echo "Next:"
echo "  make render"
echo "  kubectl apply -f manifests/dashboards/"
echo "  # or configure deploy/argocd/application.yaml and sync with Argo CD"
echo
echo "UI: kubectl -n perses-dev port-forward svc/perses-sample 8080:8080"
