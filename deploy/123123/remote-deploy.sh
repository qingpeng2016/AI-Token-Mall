#!/usr/bin/env bash
# BOX 三仓库远程部署：box（Go API/Bot）/ realbox（Vue 前端）/ box_mis（PHP 后台）
#
# 用法:
#   ./remote-deploy.sh <仓库> <分支> [动作] [模式]
#
# 仓库 (repo):
#   box | realbox | mis
#   别名: api/backend, front/frontend, box_mis
#
# 动作 (action，默认 deploy):
#   deploy   拉代码 + 停旧进程 + 启动（前端/Mis 为构建/清缓存）
#   restart  仅停旧进程 + 启动（不 git pull；realbox/mis 仍会 rebuild）
#   stop     仅停止匹配进程（realbox/mis 无托管进程则跳过）
#   status   查看状态
#
# 模式 (mode，仅 box):
#   user | api     API 用户服务  --run_bot=false  (默认)
#   robot | bot    机器人服务    --run_bot=true
#
# 示例:
#   ./remote-deploy.sh box master
#   ./remote-deploy.sh box master deploy user
#   ./remote-deploy.sh box master restart robot
#   ./remote-deploy.sh realbox v1 deploy
#   ./remote-deploy.sh mis master deploy
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.deploy.env"

usage() {
  cat <<'EOF'
用法: ./remote-deploy.sh <仓库> <分支> [动作] [模式]

仓库: box | realbox | mis
动作: deploy(默认) | restart | stop | status
模式 (仅 box):
  user | api   API 用户服务 (默认)
  robot | bot  机器人 Bot

示例:
  ./remote-deploy.sh box master deploy user
  ./remote-deploy.sh box master restart robot
  ./remote-deploy.sh realbox v1 deploy
  ./remote-deploy.sh mis master deploy
EOF
  exit 1
}

REPO_RAW="${1:-}"
BRANCH="${2:-}"
ACTION="${3:-deploy}"
MODE="${4:-}"

[[ -n "$REPO_RAW" && -n "$BRANCH" ]] || usage

case "$ACTION" in
  deploy|restart|stop|status) ;;
  user|api|robot|bot)
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

BOX_APP_DIR="${BOX_APP_DIR:-/www/wwwroot/box}"
REALBOX_APP_DIR="${REALBOX_APP_DIR:-/www/wwwroot/realbox}"
BOX_MIS_APP_DIR="${BOX_MIS_APP_DIR:-/www/wwwroot/box_mis}"
BOX_USER_RUN_CONF="${BOX_USER_RUN_CONF:-prod-user}"
BOX_ROBOT_RUN_CONF="${BOX_ROBOT_RUN_CONF:-prod-robot}"
BOX_USER_PORT="${BOX_USER_PORT:-8882}"
BOX_ROBOT_PORT="${BOX_ROBOT_PORT:-8881}"
REALBOX_BUILD_CMD="${REALBOX_BUILD_CMD:-npm run build:prod}"
GIT_FETCH_TIMEOUT="${GIT_FETCH_TIMEOUT:-120}"

normalize_repo() {
  case "$1" in
    box|api|backend|box-prod) echo "box" ;;
    realbox|front|frontend|box-front) echo "realbox" ;;
    mis|box_mis|box-mis) echo "mis" ;;
    *)
      echo "未知仓库: $1（支持 box / realbox / mis）" >&2
      exit 1
      ;;
  esac
}

REPO="$(normalize_repo "$REPO_RAW")"

if [[ -z "$MODE" ]]; then
  case "$REPO" in
    box) MODE="user" ;;
    *) MODE="" ;;
  esac
fi

case "$REPO" in
  box)
    APP_DIR="$BOX_APP_DIR"
    case "$MODE" in
      user|api|robot|bot) ;;
      *)
        echo "box 模式只能是 user/api 或 robot/bot，当前: ${MODE}" >&2
        exit 1
        ;;
    esac
    ;;
  realbox)
    APP_DIR="$REALBOX_APP_DIR"
    MODE=""
    ;;
  mis)
    APP_DIR="$BOX_MIS_APP_DIR"
    MODE=""
    ;;
esac

case "$MODE" in
  user|api) RUN_CONF="$BOX_USER_RUN_CONF"; RUN_BOT="false"; LISTEN_PORT="$BOX_USER_PORT" ;;
  robot|bot) RUN_CONF="$BOX_ROBOT_RUN_CONF"; RUN_BOT="true"; LISTEN_PORT="$BOX_ROBOT_PORT" ;;
  *) RUN_CONF=""; RUN_BOT=""; LISTEN_PORT="" ;;
esac

SSH_TARGET="${SSH_USER}@${SSH_HOST}"

read -r -d '' REMOTE_SCRIPT_BODY <<EOF || true
set -euo pipefail

REPO="${REPO}"
APP_DIR="${APP_DIR}"
BRANCH="${BRANCH}"
ACTION="${ACTION}"
MODE="${MODE}"
RUN_CONF="${RUN_CONF}"
RUN_BOT="${RUN_BOT}"
LISTEN_PORT="${LISTEN_PORT}"
REALBOX_BUILD_CMD="${REALBOX_BUILD_CMD}"
GIT_FETCH_TIMEOUT="${GIT_FETCH_TIMEOUT}"

LOG_DIR="\${APP_DIR}/logs"
BIN_DIR="\${APP_DIR}/bin"
PID_NAME="\${REPO}"
if [[ -n "\${MODE}" ]]; then
  PID_NAME="\${REPO}-\${MODE}"
fi
PID_FILE="\${LOG_DIR}/\${PID_NAME}.pid"
LOG_FILE="\${LOG_DIR}/\${PID_NAME}-nohup.log"

mkdir -p "\${LOG_DIR}" 2>/dev/null || true

proc_cwd() {
  local pid="\$1"
  readlink -f "/proc/\${pid}/cwd" 2>/dev/null || true
}

is_box_cmd() {
  local cmd="\$1"
  [[ "\${cmd}" == *"main.go"* || "\${cmd}" == *"/bin/box"* || "\${cmd}" == *" box "* || "\${cmd}" == *"/box "* || "\${cmd}" == *"box-prod"* || "\${cmd}" == *"--run_conf="* ]]
}

cmdline_matches_mode() {
  local pid="\$1"
  local cmd
  cmd=\$(tr '\\0' ' ' <"/proc/\${pid}/cmdline" 2>/dev/null || true)
  [[ -n "\${cmd}" ]] || return 1
  is_box_cmd "\${cmd}" || return 1
  case "\${MODE}" in
    user|api)
      [[ "\${cmd}" == *"--run_bot=false"* || "\${cmd}" == *"--run_bot=0"* ]] || return 1
      [[ "\${cmd}" == *"--run_conf=\${RUN_CONF}"* ]] || return 1
      ;;
    robot|bot)
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
      echo "ERROR: 常见原因: HTTPS 未配置凭据、网络不通、Gitee/GitHub 不可达" >&2
      echo "ERROR: 请在服务器执行: cd \$(pwd) && git remote -v && git fetch origin -v" >&2
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
    echo "ERROR: 请在 .deploy.env 修正 BOX_APP_DIR / REALBOX_APP_DIR / BOX_MIS_APP_DIR" >&2
    exit 1
  fi
  cd "\${dir}"
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "ERROR: \${dir} 不是 git 仓库（无 .git）" >&2
    echo "ERROR: 请 SSH 登录服务器后 clone，或修改 .deploy.env 中的目录路径" >&2
    echo "ERROR: 示例: git clone git@github.com:qingpeng2016/box.git \${dir}" >&2
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
    echo "ERROR: 找不到分支 \${branch} @ \${dir}" >&2
    echo "ERROR: 可用远程分支:" >&2
    git branch -r | head -n 20 >&2 || true
    exit 1
  fi
  export GIT_TERMINAL_PROMPT=0
  if ! git pull --ff-only origin "\${branch}"; then
    echo ">>> pull 仍失败，尝试 stash 后重试" >&2
    git stash push -u -m "remote-deploy auto-stash \$(date +%F_%T)" || true
    if ! git pull --ff-only origin "\${branch}"; then
      echo "ERROR: git pull --ff-only 失败，可能有本地改动或需要 merge" >&2
      git status -sb >&2 || true
      exit 1
    fi
  fi
  git log -1 --oneline
}

git_pull() {
  git_pull_at "\${APP_DIR}" "\${BRANCH}"
}

build_box_binary() {
  cd "\${APP_DIR}"
  mkdir -p "\${BIN_DIR}"
  local bin_path="\${BIN_DIR}/box"
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

stop_box() {
  echo ">>> 停止 [box/\${MODE}] @ \${APP_DIR} (port \${LISTEN_PORT})"
  if [[ -f "\${PID_FILE}" ]]; then
    old_pid=\$(tr -d '[:space:]' <"\${PID_FILE}" 2>/dev/null || true)
    if [[ -n "\${old_pid}" ]] && kill -0 "\${old_pid}" 2>/dev/null; then
      echo ">>> kill PID 文件 pid=\${old_pid}"
      kill -TERM "\${old_pid}" 2>/dev/null || true
    fi
    rm -f "\${PID_FILE}"
  fi
  # 先按 run_conf 匹配（不限 cwd，兼容 go run 旧进程）
  kill_matching TERM loose
  kill_port_listeners "\${LISTEN_PORT}" TERM
  sleep 2
  kill_matching KILL loose
  kill_port_listeners "\${LISTEN_PORT}" KILL
  sleep 1
  if port_in_use "\${LISTEN_PORT}"; then
    echo "ERROR: 端口 \${LISTEN_PORT} 仍被占用:" >&2
    ss -lptn "sport = :\${LISTEN_PORT}" 2>/dev/null >&2 || lsof -i ":\${LISTEN_PORT}" 2>/dev/null >&2 || true
    exit 1
  fi
  echo ">>> 已停止，端口 \${LISTEN_PORT} 已释放"
}

show_box_status() {
  echo ">>> 状态 [box/\${MODE}] @ \${APP_DIR} (port \${LISTEN_PORT})"
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
  if port_in_use "\${LISTEN_PORT}"; then
    echo ">>> 端口 \${LISTEN_PORT} 监听中: \$(pids_on_port "\${LISTEN_PORT}" | tr '\\n' ' ')"
    found=1
  fi
  [[ -f "\${PID_FILE}" ]] && echo ">>> PID 文件: \$(cat "\${PID_FILE}" 2>/dev/null || echo '-')"
  [[ "\${found}" -eq 0 ]] && echo ">>> 无匹配进程"
}

wait_for_box() {
  local i pid
  echo ">>> 等待 box 进程启动（最多 20s）..."
  for i in \$(seq 1 20); do
    pid="\$(find_matching_pid || true)"
    if [[ -n "\${pid}" ]] && kill -0 "\${pid}" 2>/dev/null; then
      echo ">>> 启动成功 pid=\${pid}"
      show_box_status
      return 0
    fi
    if [[ -f "\${LOG_FILE}" ]] && grep -qE 'panic:|FATAL|exit status' "\${LOG_FILE}"; then
      echo ">>> 启动失败，日志:" >&2
      tail -n 40 "\${LOG_FILE}" >&2 || true
      exit 1
    fi
    sleep 1
  done
  echo ">>> 启动超时" >&2
  tail -n 40 "\${LOG_FILE}" 2>/dev/null >&2 || true
  exit 1
}

start_box() {
  cd "\${APP_DIR}"
  mkdir -p "\${BIN_DIR}" "\${LOG_DIR}"
  local bin_path="\${BIN_DIR}/box"
  [[ -x "\${bin_path}" ]] || build_box_binary

  echo ">>> 日志: \${LOG_FILE}"
  : >"\${LOG_FILE}"

  export PATH="\${PATH}:/usr/local/go/bin"
  echo ">>> 启动: \${bin_path} --run_conf=\${RUN_CONF} --run_bot=\${RUN_BOT}"
  nohup "\${bin_path}" --run_conf="\${RUN_CONF}" --run_bot="\${RUN_BOT}" >>"\${LOG_FILE}" 2>&1 &
  echo \$! >"\${PID_FILE}"
  wait_for_box
}

deploy_realbox() {
  cd "\${APP_DIR}"
  echo ">>> realbox 前端部署 @ \${APP_DIR}"
  if [[ "\${ACTION}" == "deploy" ]]; then
    git_pull
  fi

  export PATH="\${PATH}:/usr/local/node/bin"
  if [[ -f dist/.user.ini ]]; then
    chattr -i dist/.user.ini 2>/dev/null || true
  fi
  rm -rf dist node_modules/.cache

  echo ">>> npm install"
  npm install
  echo ">>> \${REALBOX_BUILD_CMD}"
  eval "\${REALBOX_BUILD_CMD}"

  echo ">>> 构建完成"
  ls -lh dist/ 2>/dev/null | head -n 15 || true
  echo ">>> Nginx 根目录应指向: \${APP_DIR}/dist"
}

deploy_mis() {
  cd "\${APP_DIR}"
  echo ">>> box_mis 部署 @ \${APP_DIR}"
  if [[ "\${ACTION}" == "deploy" ]]; then
    git_pull
  fi
  if [[ -f vendor/autoload.php ]]; then
    echo ">>> composer 依赖已存在，跳过 install"
  elif command -v composer >/dev/null 2>&1; then
    composer install --no-dev -o || true
  fi
  if [[ -f think ]]; then
    php think cache:clear >/dev/null 2>&1 || true
    echo ">>> 已执行 php think cache:clear"
  fi
  git log -1 --oneline 2>/dev/null || true
  echo ">>> Mis 部署完成（Web 由 Nginx/PHP-FPM 提供）"
}

case "\${ACTION}" in
  deploy)
    case "\${REPO}" in
      box)
        git_pull
        build_box_binary
        stop_box
        start_box
        ;;
      realbox) deploy_realbox ;;
      mis) deploy_mis ;;
    esac
    ;;
  restart)
    case "\${REPO}" in
      box)
        build_box_binary
        stop_box
        start_box
        ;;
      realbox) deploy_realbox ;;
      mis) deploy_mis ;;
    esac
    ;;
  stop)
    case "\${REPO}" in
      box) stop_box ;;
      realbox) echo ">>> realbox 为静态资源，无托管进程" ;;
      mis) echo ">>> mis 无托管 Go 进程" ;;
    esac
    ;;
  status)
    case "\${REPO}" in
      box) show_box_status ;;
      realbox)
        echo ">>> realbox @ \${APP_DIR}"
        cd "\${APP_DIR}" && git log -1 --oneline 2>/dev/null || true
        ls -lh dist/index.html 2>/dev/null || echo ">>> dist/index.html 不存在，需 deploy"
        ;;
      mis)
        echo ">>> mis @ \${APP_DIR}"
        cd "\${APP_DIR}" && git log -1 --oneline 2>/dev/null || true
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
set timeout 600
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
echo ">>> 仓库=${REPO} 分支=${BRANCH} 动作=${ACTION} 模式=${MODE:-n/a} 目录=${APP_DIR}"

if command -v sshpass >/dev/null 2>&1; then
  run_with_sshpass
elif command -v expect >/dev/null 2>&1; then
  run_with_expect
else
  echo "未找到 sshpass 或 expect。" >&2
  exit 1
fi

rc=$?
if [[ $rc -ne 0 ]]; then
  echo ">>> 远程操作失败 (exit ${rc})" >&2
  exit $rc
fi

echo ">>> 远程操作完成"
