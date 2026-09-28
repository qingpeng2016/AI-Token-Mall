# 仅在远程服务器上执行（由 deploy-*.sh 通过 SSH 注入，勿在本机 source）

git_prepare_for_pull() {
  local clean_bin="${1:-0}"
  cd "${APP_DIR}"
  if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git restore --worktree log/info.log log/error.log 2>/dev/null \
      || git checkout HEAD -- log/info.log log/error.log 2>/dev/null \
      || true
  fi
  if [[ "${clean_bin}" == "1" ]]; then
    git clean -fd logs/ bin/ 2>/dev/null || true
  fi
}

git_fetch_with_timeout() {
  export GIT_TERMINAL_PROMPT=0
  if command -v timeout >/dev/null 2>&1; then
    timeout "${GIT_FETCH_TIMEOUT}" git fetch origin --prune --progress
  else
    git fetch origin --prune --progress
  fi
}

git_sync_branch() {
  cd "${APP_DIR}"
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "ERROR: ${APP_DIR} 不是 git 仓库" >&2
    exit 1
  fi
  git_prepare_for_pull "${GIT_CLEAN_BIN:-0}"
  echo ">>> [服务器] 目录: ${APP_DIR}"
  echo ">>> [服务器] git fetch + checkout ${BRANCH} + pull"
  git_fetch_with_timeout
  if git show-ref --verify --quiet "refs/heads/${BRANCH}"; then
    git checkout "${BRANCH}"
  else
    git checkout -B "${BRANCH}" "origin/${BRANCH}"
  fi
  git pull --ff-only origin "${BRANCH}" || {
    git stash push -u -m "deploy stash" || true
    git pull --ff-only origin "${BRANCH}"
  }
  git log -1 --oneline
}

setup_node_path() {
  export PATH="${PATH}:/usr/local/node/bin:/usr/local/bin:/usr/bin"
  if [[ -d /www/server/nvm/versions/node ]]; then
    latest=$(ls -1 /www/server/nvm/versions/node 2>/dev/null | sort -V | tail -1 || true)
    if [[ -n "${latest}" ]]; then
      export PATH="/www/server/nvm/versions/node/${latest}/bin:${PATH}"
    fi
  fi
  if [[ -s "${HOME}/.nvm/nvm.sh" ]]; then
    # shellcheck disable=SC1091
    source "${HOME}/.nvm/nvm.sh"
  fi
}

ensure_remote_pnpm() {
  setup_node_path
  if ! command -v node >/dev/null 2>&1; then
    echo "ERROR: [服务器] 未找到 node，请在宝塔安装 Node 18+" >&2
    exit 1
  fi
  echo ">>> [服务器] node $(node -v)"
  if command -v pnpm >/dev/null 2>&1; then
    echo ">>> [服务器] pnpm $(pnpm -v)"
    return 0
  fi
  if command -v corepack >/dev/null 2>&1; then
    corepack enable 2>/dev/null || true
  fi
  if command -v pnpm >/dev/null 2>&1; then
    echo ">>> [服务器] pnpm $(pnpm -v)"
    return 0
  fi
  echo ">>> [服务器] npm install -g pnpm@9"
  npm install -g pnpm@9
  setup_node_path
  command -v pnpm >/dev/null 2>&1 || {
    echo "ERROR: [服务器] 无法安装 pnpm" >&2
    exit 1
  }
  echo ">>> [服务器] pnpm $(pnpm -v)"
}
