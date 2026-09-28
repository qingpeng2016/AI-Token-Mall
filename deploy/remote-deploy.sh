#!/usr/bin/env bash
# AI-Token-Mall 远程部署：Go（HTTP API / Bot）+ web-pc（Vue）
#
# 用法:
#   ./remote-deploy.sh <目标> <分支> [动作] [模式]
#
# 目标 (target):
#   mall | api | go | backend     Go 服务（默认 deploy 时编译 bin/ai-token-mall）
#   web | web-pc | front          前端 pnpm build:pc
#
# 动作 (action，默认 deploy):
#   deploy   拉代码 + 构建 + 启停（web 为 install + build）
#   restart  不 git pull；Go 重新编译并重启，web 仍 rebuild
#   stop     仅停止 Go 进程（web 无托管进程）
#   status   查看状态
#
# 模式 (mode，仅 Go):
#   user | api     HTTP API  --run_bot=false  (默认)
#   bot | robot    定时 Bot   --run_bot=true
#
# 示例:
#   ./remote-deploy.sh mall master
#   ./remote-deploy.sh api master deploy user
#   ./remote-deploy.sh go master restart bot
#   ./remote-deploy.sh web-pc master deploy
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.deploy.env"

usage() {
  cat <<'EOF'
用法: ./remote-deploy.sh <目标> <分支> [动作] [模式]

目标: mall | api | go | backend | web | web-pc | front
动作: deploy(默认) | restart | stop | status
模式 (仅 Go): user | api (HTTP，默认) | bot | robot (Bot)

示例:
  ./remote-deploy.sh mall master deploy user
  ./remote-deploy.sh api master restart bot
  ./remote-deploy.sh web-pc master deploy
EOF
  exit 1
}

TARGET_RAW="${1:-}"
BRANCH="${2:-}"
ACTION="${3:-deploy}"
MODE="${4:-}"

[[ -n "$TARGET_RAW" && -n "$BRANCH" ]] || usage

case "$ACTION" in
  deploy|restart|stop|status) ;;
  user|api|bot|robot)
    MODE="$ACTION"
    ACTION="deploy"
    ;;
  *)
    echo "未知动作: $ACTION" >&2
    usage
    ;;
esac

[[ -f "$ENV_FILE" ]] || {
  echo "缺少 ${ENV_FILE}，请先: cp .deploy.env.example .deploy.env" >&2
  exit 1
}

# shellcheck source=/dev/null
source "$ENV_FILE"

: "${SSH_HOST:?请在 .deploy.env 设置 SSH_HOST}"
: "${SSH_USER:?请在 .deploy.env 设置 SSH_USER}"
: "${SSH_PASS:?请在 .deploy.env 设置 SSH_PASS}"
SSH_PORT="${SSH_PORT:-22}"

MALL_APP_DIR="${MALL_APP_DIR:-/www/wwwroot/ai-token-mall}"
MALL_USER_RUN_CONF="${MALL_USER_RUN_CONF:-prod}"
MALL_BOT_RUN_CONF="${MALL_BOT_RUN_CONF:-prod}"
MALL_USER_PORT="${MALL_USER_PORT:-8886}"
MALL_WEB_ROOT="${MALL_WEB_ROOT:-front/apps}"
MALL_WEB_DIST="${MALL_WEB_DIST:-front/apps/web-pc/dist}"
WEB_BUILD_CMD="${WEB_BUILD_CMD:-corepack enable && pnpm install && pnpm build:pc}"
MALL_BIN_NAME="${MALL_BIN_NAME:-ai-token-mall}"
GIT_FETCH_TIMEOUT="${GIT_FETCH_TIMEOUT:-120}"

normalize_target() {
  case "$1" in
    mall|api|go|backend|ai-token-mall) echo "mall" ;;
    web|web-pc|front|frontend|pc) echo "web" ;;
    *)
      echo "未知目标: $1（支持 mall/api/go/backend 或 web/web-pc/front）" >&2
      exit 1
      ;;
  esac
}

TARGET="$(normalize_target "$TARGET_RAW")"

if [[ -z "$MODE" ]]; then
  case "$TARGET" in
    mall) MODE="user" ;;
    *) MODE="" ;;
  esac
fi

APP_DIR="$MALL_APP_DIR"

case "$TARGET" in
  mall)
    case "$MODE" in
      user|api|bot|robot) ;;
      *)
        echo "Go 模式只能是 user/api 或 bot/robot，当前: ${MODE}" >&2
        exit 1
        ;;
    esac
    ;;
  web) MODE="" ;;
esac

case "$MODE" in
  user|api) RUN_CONF="$MALL_USER_RUN_CONF"; RUN_BOT="false"; LISTEN_PORT="$MALL_USER_PORT" ;;
  bot|robot) RUN_CONF="$MALL_BOT_RUN_CONF"; RUN_BOT="true"; LISTEN_PORT="" ;;
  *) RUN_CONF=""; RUN_BOT=""; LISTEN_PORT="" ;;
esac

SSH_TARGET="${SSH_USER}@${SSH_HOST}"

read -r -d '' REMOTE_SCRIPT_BODY <<EOF || true
set -euo pipefail

TARGET="${TARGET}"
APP_DIR="${APP_DIR}"
BRANCH="${BRANCH}"
ACTION="${ACTION}"
MODE="${MODE}"
RUN_CONF="${RUN_CONF}"
RUN_BOT="${RUN_BOT}"
LISTEN_PORT="${LISTEN_PORT}"
MALL_WEB_ROOT="${MALL_WEB_ROOT}"
MALL_WEB_DIST="${MALL_WEB_DIST}"
WEB_BUILD_CMD="${WEB_BUILD_CMD}"
MALL_BIN_NAME="${MALL_BIN_NAME}"
GIT_FETCH_TIMEOUT="${GIT_FETCH_TIMEOUT}"

LOG_DIR="\${APP_DIR}/logs"
BIN_DIR="\${APP_DIR}/bin"
PID_NAME="ai-token-mall"
if [[ -n "\${MODE}" ]]; then
  PID_NAME="ai-token-mall-\${MODE}"
fi
PID_FILE="\${LOG_DIR}/\${PID_NAME}.pid"
LOG_FILE="\${LOG_DIR}/\${PID_NAME}-nohup.log"

mkdir -p "\${LOG_DIR}" 2>/dev/null || true

proc_cwd() {
  local pid="\$1"
  readlink -f "/proc/\${pid}/cwd" 2>/dev/null || true
}

is_mall_go_cmd() {
  local cmd="\$1"
  [[ "\${cmd}" == *"main.go"* || "\${cmd}" == *"/bin/\${MALL_BIN_NAME}"* || "\${cmd}" == *"\${MALL_BIN_NAME}"* || "\${cmd}" == *"ai-token-mall"* || "\${cmd}" == *"--run_conf="* ]]
}

cmdline_matches_mode() {
  local pid="\$1"
  local cmd
  cmd=\$(tr '\\0' ' ' <"/proc/\${pid}/cmdline" 2>/dev/null || true)
  [[ -n "\${cmd}" ]] || return 1
  is_mall_go_cmd "\${cmd}" || return 1
  case "\${MODE}" in
    user|api)
      [[ "\${cmd}" == *"--run_bot=false"* || "\${cmd}" == *"--run_bot=0"* ]] || return 1
      [[ "\${cmd}" == *"--run_conf=\${RUN_CONF}"* ]] || return 1
      ;;
    bot|robot)
      [[ "\${cmd}" == *"--run_bot=true"* || "\${cmd}" == *"--run_bot=1"* ]] || return 1
      [[ "\${cmd}" == *"--run_conf=\${RUN_CONF}"* ]] || return 1
      ;;
  esac
  return 0
}

cmdline_matches() {
  local pid="\$1"
  local cwd
  cwd="\$(proc_cwd "\${pid}")"
  [[ "\${cwd}" == "\${APP_DIR}" ]] || return 1
  cmdline_matches_mode "\${pid}"
}

pids_on_port() {
  local port="\$1"
  local pids=""
  if command -v ss >/dev/null 2>&1; then
    pids=\$(ss -lptn "sport = :\${port}" 2>/dev/null | grep -o 'pid=[0-9]*' | cut -d= -f2 | sort -u || true)
  elif command -v lsof >/dev/null 2>&1; then
    pids=\$(lsof -ti ":\${port}" 2>/dev/null | sort -u || true)
  elif command -v fuser >/dev/null 2>&1; then
    pids=\$(fuser "\${port}/tcp" 2>/dev/null || true)
  fi
  echo "\${pids}"
}

kill_port_listeners() {
  local port="\$1" sig="\${2:-TERM}" pid
  [[ -n "\${port}" ]] || return 0
  for pid in \$(pids_on_port "\${port}"); do
    [[ -n "\${pid}" ]] || continue
    echo ">>> kill -\${sig} pid=\${pid} (port \${port})"
    kill "-\${sig}" "\${pid}" 2>/dev/null || true
  done
}

port_in_use() {
  local port="\$1"
  [[ -n "\$(pids_on_port "\${port}" | tr -d '[:space:]')" ]]
}

git_prepare_for_pull() {
  local dir="\$1"
  cd "\${dir}"
  echo ">>> 清理运行时日志改动（避免 pull 冲突）"
  if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git restore --worktree log/info.log log/error.log 2>/dev/null \
      || git checkout HEAD -- log/info.log log/error.log 2>/dev/null \
      || git checkout -- log/ 2>/dev/null || true
  fi
  git clean -fd logs/ bin/ 2>/dev/null || true
}

git_fetch_with_timeout() {
  export GIT_TERMINAL_PROMPT=0
  echo ">>> git fetch origin --prune (timeout \${GIT_FETCH_TIMEOUT}s)"
  if command -v timeout >/dev/null 2>&1; then
    if ! timeout "\${GIT_FETCH_TIMEOUT}" git fetch origin --prune --progress; then
      echo "ERROR: git fetch 失败或超时（\${GIT_FETCH_TIMEOUT}s）" >&2
      exit 1
    fi
  elif ! git fetch origin --prune --progress; then
    echo "ERROR: git fetch 失败" >&2
    exit 1
  fi
}

git_pull_at() {
  local dir="\$1"
  local branch="\$2"
  if [[ ! -d "\${dir}" ]]; then
    echo "ERROR: 目录不存在: \${dir}" >&2
    exit 1
  fi
  cd "\${dir}"
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "ERROR: \${dir} 不是 git 仓库" >&2
    exit 1
  fi
  git_prepare_for_pull "\${dir}"
  echo ">>> git pull (\${branch}) @ \${dir}"
  git_fetch_with_timeout
  if git show-ref --verify --quiet "refs/heads/\${branch}"; then
    git checkout "\${branch}"
  elif git show-ref --verify --quiet "refs/remotes/origin/\${branch}"; then
    git checkout -B "\${branch}" "origin/\${branch}"
  else
    echo "ERROR: 找不到分支 \${branch}" >&2
    git branch -r | head -n 20 >&2 || true
    exit 1
  fi
  export GIT_TERMINAL_PROMPT=0
  if ! git pull --ff-only origin "\${branch}"; then
    git stash push -u -m "remote-deploy auto-stash \$(date +%F_%T)" || true
    if ! git pull --ff-only origin "\${branch}"; then
      echo "ERROR: git pull --ff-only 失败" >&2
      git status -sb >&2 || true
      exit 1
    fi
  fi
  git log -1 --oneline
}

git_pull() {
  git_pull_at "\${APP_DIR}" "\${BRANCH}"
}

build_mall_binary() {
  cd "\${APP_DIR}"
  mkdir -p "\${BIN_DIR}"
  local bin_path="\${BIN_DIR}/\${MALL_BIN_NAME}"
  echo ">>> go build -o \${bin_path} ."
  export PATH="\${PATH}:/usr/local/go/bin"
  export GOPATH="\${GOPATH:-\$HOME/go}"
  export GO111MODULE=on
  export GOPROXY="\${GOPROXY:-https://goproxy.cn,direct}"
  if ! go build -o "\${bin_path}" . ; then
    echo ">>> go build 失败" >&2
    exit 1
  fi
  echo ">>> 编译成功: \${bin_path}"
}

find_matching_pid() {
  local pid
  for pid_path in /proc/[0-9]*; do
    pid=\${pid_path#/proc/}
    cmdline_matches_mode "\${pid}" || continue
    echo "\${pid}"
    return 0
  done
  return 1
}

kill_matching() {
  local sig="\${1:-TERM}" use_loose="\${2:-}" pid cmd cwd
  for pid_path in /proc/[0-9]*; do
    pid=\${pid_path#/proc/}
    if [[ "\${use_loose}" == "loose" ]]; then
      cmdline_matches_mode "\${pid}" || continue
    else
      cmdline_matches "\${pid}" || continue
    fi
    cmd=\$(tr '\\0' ' ' <"/proc/\${pid}/cmdline" 2>/dev/null || true)
    cwd=\$(proc_cwd "\${pid}")
    echo ">>> kill -\${sig} pid=\${pid} cwd=\${cwd}"
    echo ">>>   \${cmd}"
    kill "-\${sig}" "\${pid}" 2>/dev/null || true
  done
}

stop_mall() {
  echo ">>> 停止 [ai-token-mall/\${MODE}] @ \${APP_DIR}"
  if [[ -f "\${PID_FILE}" ]]; then
    old_pid=\$(tr -d '[:space:]' <"\${PID_FILE}" 2>/dev/null || true)
    if [[ -n "\${old_pid}" ]] && kill -0 "\${old_pid}" 2>/dev/null; then
      echo ">>> kill PID 文件 pid=\${old_pid}"
      kill -TERM "\${old_pid}" 2>/dev/null || true
    fi
    rm -f "\${PID_FILE}"
  fi
  kill_matching TERM loose
  if [[ "\${RUN_BOT}" == "false" && -n "\${LISTEN_PORT}" ]]; then
    kill_port_listeners "\${LISTEN_PORT}" TERM
  fi
  sleep 2
  kill_matching KILL loose
  if [[ "\${RUN_BOT}" == "false" && -n "\${LISTEN_PORT}" ]]; then
    kill_port_listeners "\${LISTEN_PORT}" KILL
    sleep 1
    if port_in_use "\${LISTEN_PORT}"; then
      echo "ERROR: 端口 \${LISTEN_PORT} 仍被占用:" >&2
      ss -lptn "sport = :\${LISTEN_PORT}" 2>/dev/null >&2 || true
      exit 1
    fi
    echo ">>> 已停止，端口 \${LISTEN_PORT} 已释放"
  else
    echo ">>> 已发送停止信号（Bot 模式无 HTTP 端口）"
  fi
}

show_mall_status() {
  echo ">>> 状态 [ai-token-mall/\${MODE}] @ \${APP_DIR}"
  local found=0 pid cmd cwd
  for pid_path in /proc/[0-9]*; do
    pid=\${pid_path#/proc/}
    cmdline_matches_mode "\${pid}" || continue
    found=1
    cmd=\$(tr '\\0' ' ' <"/proc/\${pid}/cmdline" 2>/dev/null || true)
    cwd=\$(proc_cwd "\${pid}")
    echo ">>> RUNNING pid=\${pid} cwd=\${cwd}"
    echo ">>>   \${cmd}"
  done
  if [[ "\${RUN_BOT}" == "false" && -n "\${LISTEN_PORT}" ]] && port_in_use "\${LISTEN_PORT}"; then
    echo ">>> 端口 \${LISTEN_PORT} 监听中: \$(pids_on_port "\${LISTEN_PORT}" | tr '\\n' ' ')"
    found=1
  fi
  [[ -f "\${PID_FILE}" ]] && echo ">>> PID 文件: \$(cat "\${PID_FILE}" 2>/dev/null || echo '-')"
  [[ -f "\${LOG_FILE}" ]] && echo ">>> 日志: \${LOG_FILE}"
  [[ "\${found}" -eq 0 ]] && echo ">>> 无匹配进程"
}

wait_for_mall() {
  local i pid
  echo ">>> 等待进程启动（最多 30s）..."
  for i in \$(seq 1 30); do
    pid="\$(find_matching_pid || true)"
    if [[ -n "\${pid}" ]] && kill -0 "\${pid}" 2>/dev/null; then
      if [[ "\${RUN_BOT}" == "false" && -n "\${LISTEN_PORT}" ]]; then
        if port_in_use "\${LISTEN_PORT}"; then
          echo ">>> 启动成功 pid=\${pid}，端口 \${LISTEN_PORT} 已监听"
          show_mall_status
          return 0
        fi
      else
        echo ">>> Bot 启动成功 pid=\${pid}"
        show_mall_status
        return 0
      fi
    fi
    if [[ -f "\${LOG_FILE}" ]] && grep -qE 'panic:|FATAL|exit status' "\${LOG_FILE}"; then
      echo ">>> 启动失败，日志:" >&2
      tail -n 50 "\${LOG_FILE}" >&2 || true
      exit 1
    fi
    sleep 1
  done
  echo ">>> 启动超时" >&2
  tail -n 50 "\${LOG_FILE}" 2>/dev/null >&2 || true
  exit 1
}

start_mall() {
  cd "\${APP_DIR}"
  mkdir -p "\${BIN_DIR}" "\${LOG_DIR}"
  local bin_path="\${BIN_DIR}/\${MALL_BIN_NAME}"
  [[ -x "\${bin_path}" ]] || build_mall_binary

  echo ">>> 日志: \${LOG_FILE}"
  : >"\${LOG_FILE}"

  export PATH="\${PATH}:/usr/local/go/bin"
  echo ">>> 启动: \${bin_path} --run_conf=\${RUN_CONF} --run_bot=\${RUN_BOT}"
  nohup "\${bin_path}" --run_conf="\${RUN_CONF}" --run_bot="\${RUN_BOT}" >>"\${LOG_FILE}" 2>&1 &
  echo \$! >"\${PID_FILE}"
  wait_for_mall
}

deploy_web() {
  local web_dir="\${APP_DIR}/\${MALL_WEB_ROOT}"
  local dist_path="\${APP_DIR}/\${MALL_WEB_DIST}"
  echo ">>> web-pc 部署 @ \${web_dir}"
  if [[ "\${ACTION}" == "deploy" ]]; then
    git_pull
  fi
  if [[ ! -d "\${web_dir}" ]]; then
    echo "ERROR: 前端目录不存在: \${web_dir}" >&2
    exit 1
  fi
  cd "\${web_dir}"
  export PATH="\${PATH}:/usr/local/node/bin"
  if [[ -f "\${dist_path}/.user.ini" ]]; then
    chattr -i "\${dist_path}/.user.ini" 2>/dev/null || true
  fi
  rm -rf "\${dist_path}" node_modules/.cache 2>/dev/null || true

  echo ">>> \${WEB_BUILD_CMD}"
  eval "\${WEB_BUILD_CMD}"

  if [[ ! -f "\${dist_path}/index.html" ]]; then
    echo "ERROR: 构建后未找到 \${dist_path}/index.html" >&2
    exit 1
  fi
  echo ">>> 构建完成"
  ls -lh "\${dist_path}/" | head -n 15 || true
  echo ">>> Nginx 静态根目录应指向: \${dist_path}"
}

case "\${ACTION}" in
  deploy)
    case "\${TARGET}" in
      mall)
        git_pull
        build_mall_binary
        stop_mall
        start_mall
        ;;
      web) deploy_web ;;
    esac
    ;;
  restart)
    case "\${TARGET}" in
      mall)
        build_mall_binary
        stop_mall
        start_mall
        ;;
      web) deploy_web ;;
    esac
    ;;
  stop)
    case "\${TARGET}" in
      mall) stop_mall ;;
      web) echo ">>> web-pc 为静态资源，无托管进程" ;;
    esac
    ;;
  status)
    case "\${TARGET}" in
      mall) show_mall_status ;;
      web)
        echo ">>> web-pc @ \${APP_DIR}/\${MALL_WEB_DIST}"
        cd "\${APP_DIR}" && git log -1 --oneline 2>/dev/null || true
        ls -lh "\${APP_DIR}/\${MALL_WEB_DIST}/index.html" 2>/dev/null || echo ">>> 未构建，请: ./remote-deploy.sh web master deploy"
        ;;
    esac
    ;;
  *)
    echo "未知动作: \${ACTION}" >&2
    exit 1
    ;;
esac
EOF

REMOTE_B64=$(printf '%s' "$REMOTE_SCRIPT_BODY" | base64 | tr -d '\n')
REMOTE_ONE_LINER="echo ${REMOTE_B64} | base64 -d | bash"

run_with_sshpass() {
  sshpass -p "$SSH_PASS" ssh \
    -o StrictHostKeyChecking=accept-new \
    -o PreferredAuthentications=password \
    -o PubkeyAuthentication=no \
    -p "$SSH_PORT" \
    "$SSH_TARGET" \
    "$REMOTE_ONE_LINER"
  return $?
}

run_with_expect() {
  expect <<EOF
set timeout 900
log_user 1
spawn ssh -o StrictHostKeyChecking=accept-new -o PreferredAuthentications=password -o PubkeyAuthentication=no -p ${SSH_PORT} ${SSH_TARGET} "${REMOTE_ONE_LINER}"
expect {
  -re "(?i)password:" {
    send "${SSH_PASS}\r"
    exp_continue
  }
  eof
}
catch wait result
exit [lindex \$result 3]
EOF
}

echo ">>> 连接 ${SSH_TARGET}:${SSH_PORT}"
echo ">>> 目标=${TARGET} 分支=${BRANCH} 动作=${ACTION} 模式=${MODE:-n/a} 目录=${APP_DIR}"

if command -v sshpass >/dev/null 2>&1; then
  run_with_sshpass
elif command -v expect >/dev/null 2>&1; then
  run_with_expect
else
  echo "未找到 sshpass 或 expect，请安装其一。" >&2
  exit 1
fi

rc=$?
if [[ $rc -ne 0 ]]; then
  echo ">>> 远程操作失败 (exit ${rc})" >&2
  exit $rc
fi

echo ">>> 远程操作完成"
