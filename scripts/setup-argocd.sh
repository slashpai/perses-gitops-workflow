#!/usr/bin/env bash
# Install Argo CD (if needed) and apply the perses-dashboards Application
# with a user-provided Git repo URL.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
APP_TEMPLATE="${ROOT_DIR}/deploy/argocd/application.yaml"
REPO_URL="${REPO_URL:-}"
TARGET_REVISION="${TARGET_REVISION:-main}"
YES="${YES:-false}"
ARGOCD_NAMESPACE="${ARGOCD_NAMESPACE:-argocd}"
APP_NAME="${APP_NAME:-perses-dashboards}"
SYNC_TIMEOUT_S="${SYNC_TIMEOUT_S:-180}"
ENABLE_PRESYNC_CHECK="${ENABLE_PRESYNC_CHECK:-}"

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
    echo "error: non-interactive shell; re-run with YES=true" >&2
    exit 1
  fi
  local reply
  read -r -p "${prompt} [y/N] " reply
  [[ "${reply}" =~ ^[Yy]$ ]]
}

# Accept https://host/org/repo(.git)? or git@host:org/repo(.git)?
# Reject empty, placeholders, and local paths.
validate_repo_url() {
  local url="$1"

  if [[ -z "${url}" ]]; then
    echo "error: repo URL is empty" >&2
    return 1
  fi

  if [[ "${url}" == *"<"* || "${url}" == *">"* || "${url}" == *your-user* ]]; then
    echo "error: replace the placeholder with a real Git URL (got: ${url})" >&2
    return 1
  fi

  if [[ "${url}" == *" "* ]]; then
    echo "error: repo URL must not contain spaces" >&2
    return 1
  fi

  # Require org/repo (at least two path segments).
  if [[ "${url}" =~ ^https://[A-Za-z0-9._-]+/[A-Za-z0-9._-]+/[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*(\.git)?$ ]]; then
    return 0
  fi
  if [[ "${url}" =~ ^git@[A-Za-z0-9._-]+:[A-Za-z0-9._-]+/[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*(\.git)?$ ]]; then
    return 0
  fi

  echo "error: invalid Git repo URL: ${url}" >&2
  echo "  expected e.g. https://github.com/<org>/<repo>.git" >&2
  echo "           or git@github.com:<org>/<repo>.git" >&2
  return 1
}

prompt_repo_url() {
  if [[ -n "${REPO_URL}" ]]; then
    validate_repo_url "${REPO_URL}"
    return 0
  fi

  if [[ ! -t 0 ]]; then
    echo "error: set REPO_URL=... when running non-interactively" >&2
    exit 1
  fi

  echo
  echo "Enter the Git repo URL Argo CD should sync (your fork of this repo)."
  echo "Example: https://github.com/ORG/REPO.git"
  echo
  local url
  while true; do
    read -r -p "repoURL: " url
    if validate_repo_url "${url}"; then
      REPO_URL="${url}"
      return 0
    fi
  done
}

# Fail early if Argo would get "revision … must be resolved" because the branch
# is not on the remote yet (common when local commits are not pushed).
check_remote_revision() {
  if ! command -v git >/dev/null 2>&1; then
    echo "warn: git not found; skipping remote revision check for ${TARGET_REVISION}"
    return 0
  fi

  echo "==> Checking remote has revision ${TARGET_REVISION}"
  if ! git ls-remote --exit-code "${REPO_URL}" "refs/heads/${TARGET_REVISION}" >/dev/null 2>&1; then
    echo "error: cannot resolve ${TARGET_REVISION} on ${REPO_URL}" >&2
    echo "  Push your branch first, e.g.: git push -u origin ${TARGET_REVISION}" >&2
    echo "  Or set TARGET_REVISION to an existing remote branch/tag/commit." >&2
    return 1
  fi
}

wait_for_deployment() {
  local name="$1"
  echo "    waiting for deployment/${name}"
  kubectl wait --for=condition=available \
    "deployment/${name}" -n "${ARGOCD_NAMESPACE}" --timeout=300s
}

# Avoid applying the Application while repo-server is still starting — that often
# surfaces as dial tcp …:8081: connection refused (8081 is the correct port).
wait_for_repo_server_endpoints() {
  local deadline=$((SECONDS + 180))
  echo "    waiting for argocd-repo-server endpoints"
  while (( SECONDS < deadline )); do
    local addrs
    addrs="$(kubectl get endpoints argocd-repo-server -n "${ARGOCD_NAMESPACE}" \
      -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null || true)"
    if [[ -n "${addrs}" ]]; then
      return 0
    fi
    sleep 2
  done
  echo "error: timed out waiting for argocd-repo-server endpoints" >&2
  return 1
}

wait_for_application_sync() {
  local deadline=$((SECONDS + SYNC_TIMEOUT_S))
  local refreshed=false

  echo "==> Waiting for Application ${APP_NAME} to sync (timeout ${SYNC_TIMEOUT_S}s)"
  while (( SECONDS < deadline )); do
    local sync health message
    sync="$(kubectl get application "${APP_NAME}" -n "${ARGOCD_NAMESPACE}" \
      -o jsonpath='{.status.sync.status}' 2>/dev/null || true)"
    health="$(kubectl get application "${APP_NAME}" -n "${ARGOCD_NAMESPACE}" \
      -o jsonpath='{.status.health.status}' 2>/dev/null || true)"
    message="$(kubectl get application "${APP_NAME}" -n "${ARGOCD_NAMESPACE}" \
      -o jsonpath='{.status.conditions[0].message}' 2>/dev/null || true)"

    if [[ "${sync}" == "Synced" ]]; then
      echo "    sync=${sync} health=${health:-n/a}"
      return 0
    fi

    # One hard refresh after components are up — clears stale Unknown / refused.
    if [[ "${refreshed}" == "false" ]] && (( SECONDS > 15 )); then
      echo "    hard-refreshing Application (current sync=${sync:-pending})"
      kubectl annotate application "${APP_NAME}" -n "${ARGOCD_NAMESPACE}" \
        argocd.argoproj.io/refresh=hard --overwrite >/dev/null
      refreshed=true
    fi

    if [[ -n "${message}" ]]; then
      echo "    sync=${sync:-pending} — ${message}"
    else
      echo "    sync=${sync:-pending} health=${health:-n/a}"
    fi
    sleep 5
  done

  echo "warn: Application did not reach Synced within ${SYNC_TIMEOUT_S}s" >&2
  echo "  Check: kubectl get application ${APP_NAME} -n ${ARGOCD_NAMESPACE}" >&2
  echo "  See README Troubleshooting if sync stays Unknown / connection refused." >&2
  return 0
}

need kubectl

if [[ ! -f "${APP_TEMPLATE}" ]]; then
  echo "error: missing Application template: ${APP_TEMPLATE}" >&2
  exit 1
fi

prompt_repo_url
check_remote_revision

echo
echo "Will:"
echo "  - install Argo CD into namespace ${ARGOCD_NAMESPACE} (if missing)"
echo "  - apply Application ${APP_NAME}"
echo "  - repoURL:         ${REPO_URL}"
echo "  - targetRevision:  ${TARGET_REVISION}"
echo "  - path:            manifests/dashboards"
echo "  - destination ns:  perses-dev"
echo
echo "Note: the repo must be reachable by Argo CD (public fork, or private repo credentials configured)."
echo

if ! confirm "Proceed?"; then
  echo "Aborted."
  exit 0
fi

echo "==> Ensuring Argo CD is installed"
if ! kubectl get namespace "${ARGOCD_NAMESPACE}" >/dev/null 2>&1; then
  kubectl create namespace "${ARGOCD_NAMESPACE}"
fi

if ! kubectl get deploy argocd-server -n "${ARGOCD_NAMESPACE}" >/dev/null 2>&1; then
  # Server-side apply avoids CRD annotation size limits with client-side apply
  # (metadata.annotations may not be more than 262144 bytes).
  kubectl apply --server-side --force-conflicts -n "${ARGOCD_NAMESPACE}" \
    -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
fi

echo "==> Waiting for Argo CD control plane"
wait_for_deployment argocd-server
wait_for_deployment argocd-repo-server
if kubectl get deploy argocd-redis -n "${ARGOCD_NAMESPACE}" >/dev/null 2>&1; then
  wait_for_deployment argocd-redis
fi
# application-controller is a StatefulSet in the default install
echo "    waiting for statefulset/argocd-application-controller"
kubectl rollout status statefulset/argocd-application-controller \
  -n "${ARGOCD_NAMESPACE}" --timeout=300s
wait_for_repo_server_endpoints

echo "==> Applying PersesDashboard Application"
# Keep deploy/argocd/application.yaml as a template; substitute for apply only.
tmp="$(mktemp)"
trap 'rm -f "${tmp}"' EXIT
sed \
  -e "s|repoURL:.*|repoURL: ${REPO_URL}|" \
  -e "s|targetRevision:.*|targetRevision: ${TARGET_REVISION}|" \
  "${APP_TEMPLATE}" > "${tmp}"

# Sanity: substituted file must not still contain the placeholder.
if grep -q '<your-user>' "${tmp}"; then
  echo "error: failed to substitute repoURL in Application manifest" >&2
  exit 1
fi

kubectl apply -f "${tmp}"
wait_for_application_sync

echo
echo "Argo CD Application applied."
echo "  Check sync:  kubectl get application ${APP_NAME} -n ${ARGOCD_NAMESPACE}"
echo "  Dashboards:  kubectl get persesdashboard -n perses-dev"

# --- Optional: PreSync metric validation ---
PRESYNC_SRC="${ROOT_DIR}/deploy/metrics-usage/presync-check.yaml"
PRESYNC_DST="${ROOT_DIR}/manifests/dashboards/presync-check.yaml"

if [[ -n "${ENABLE_PRESYNC_CHECK}" ]]; then
  enable_presync="${ENABLE_PRESYNC_CHECK}"
elif [[ "${YES}" == "true" ]]; then
  enable_presync="false"
else
  echo
  read -r -p "Enable PreSync metric validation? (requires metrics-usage running) [y/N] " enable_presync
  enable_presync="${enable_presync:-n}"
fi

if [[ "${enable_presync}" =~ ^[Yy] ]]; then
  if ! kubectl get deploy metrics-usage -n perses-dev >/dev/null 2>&1; then
    echo "  metrics-usage not found — deploying it first..."
    make -C "${ROOT_DIR}" setup-metrics-usage
  fi
  cp "${PRESYNC_SRC}" "${PRESYNC_DST}"
  echo "  PreSync check copied to manifests/dashboards/presync-check.yaml"
  echo "  Commit and push to enable it on next Argo CD sync."
else
  # Remove if previously enabled
  rm -f "${PRESYNC_DST}"
  echo "  PreSync metric validation: skipped"
fi

echo
echo "Optional UI:"
echo "  kubectl -n ${ARGOCD_NAMESPACE} port-forward svc/argocd-server 8443:443"
echo "  # admin password: kubectl -n ${ARGOCD_NAMESPACE} get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
echo "  # open https://localhost:8443  (accept the self-signed cert)"
