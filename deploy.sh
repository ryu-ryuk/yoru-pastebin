# #!/bin/bash
# set -euo pipefail

# #        __                   __                   
# #       |  \                 |  \                  
# #   ____| ▓▓ ______   ______ | ▓▓ ______  __    __ 
# #  /      ▓▓/      \ /      \| ▓▓/      \|  \  |  \
# # |  ▓▓▓▓▓▓▓  ▓▓▓▓▓▓\  ▓▓▓▓▓▓\ ▓▓  ▓▓▓▓▓▓\ ▓▓  | ▓▓
# # | ▓▓  | ▓▓ ▓▓    ▓▓ ▓▓  | ▓▓ ▓▓ ▓▓  | ▓▓ ▓▓  | ▓▓
# # | ▓▓__| ▓▓ ▓▓▓▓▓▓▓▓ ▓▓__/ ▓▓ ▓▓ ▓▓__/ ▓▓ ▓▓__/ ▓▓
# #  \▓▓    ▓▓\▓▓     \ ▓▓    ▓▓ ▓▓\▓▓    ▓▓\▓▓    ▓▓
# #   \▓▓▓▓▓▓▓ \▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓ \▓▓ \▓▓▓▓▓▓ _\▓▓▓▓▓▓▓
# #                   | ▓▓                 |  \__| ▓▓
# #                   | ▓▓                  \▓▓    ▓▓
# #                    \▓▓                   \▓▓▓▓▓▓ 

# SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# PROJECT_DIR="${SCRIPT_DIR}"
# COMPOSE_FILE="docker-compose.prod.yml"
# ENV_FILE=".env"

# # color logging
# RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
# log()   { echo -e "${GREEN}[$(date +'%F %T')] $1${NC}"; }
# warn()  { echo -e "${YELLOW}[$(date +'%F %T')] WARNING: $1${NC}"; }
# error() { echo -e "${RED}[$(date +'%F %T')] ERROR: $1${NC}" >&2; exit 1; }
# info()  { echo -e "${BLUE}[$(date +'%F %T')] $1${NC}"; }

# check_env_vars() {
#   local vars=("$@")
#   for var in "${vars[@]}"; do
#     [[ -z "${!var:-}" ]] && error "required env variable '$var' is not set in ${ENV_FILE}"
#   done
# }

# check_requirements() {
#   log "checking system requirements..."
#   command -v docker >/dev/null || error "docker not installed"
#   docker info >/dev/null 2>&1 || error "cannot access docker daemon"
#   (command -v docker-compose >/dev/null || docker compose version >/dev/null) || error "docker-compose not installed"
#   log "✓ requirements ok"
# }

# setup() {
#   log "initializing production structure..."
#   mkdir -p data/{postgres,redis,uploads} logs backups
#   mkdir -p traefik/{letsencrypt,logs,dynamic}

#   chmod 755 data logs backups
#   chmod 700 traefik/letsencrypt

#   if [[ ! -f "${ENV_FILE}" ]]; then
#     cp .env.example "${ENV_FILE}" && warn "created default .env file. edit it with real secrets"
#   fi

#   docker network create web 2>/dev/null || log "network 'web' already exists"
#   log "✓ setup complete"
#   info "edit ${ENV_FILE}, then run: ./deploy.sh deploy"
# }

# deploy() {
#   log "starting deployment..."
#   [[ -f "${ENV_FILE}" ]] || error ".env missing. run './deploy.sh setup' first."

#   set -a; source "${ENV_FILE}"; set +a
#   check_env_vars POSTGRES_PASSWORD SERVER_PORT BASE_URL

#   log "building containers..."
#   docker-compose -f "${COMPOSE_FILE}" build

#   log "bringing up services..."
#   docker-compose -f "${COMPOSE_FILE}" up -d

#   sleep 5 && log "waiting for services to stabilize..."
#   sleep 30

#   docker-compose -f "${COMPOSE_FILE}" ps | grep -q "healthy" \
#     && log "✓ deploy succeeded" \
#     || warn "deployment complete but some containers are not healthy"
# }

# backup() {
#   log "running backup..."
#   docker-compose -f "${COMPOSE_FILE}" --profile backup run --rm backup
# }

# restore() {
#   local file="$1"
#   [[ -z "${file}" ]] && error "specify backup file"
#   [[ ! -f "${file}" ]] && error "file not found: $file"

#   warn "this will overwrite the current DB"
#   read -p "continue? (yes/no): " confirm
#   [[ "${confirm}" != "yes" ]] && log "cancelled" && exit 0

#   docker-compose -f "${COMPOSE_FILE}" stop yoru-pastebin

#   set -a; source "${ENV_FILE}"; set +a
#   check_env_vars POSTGRES_USER POSTGRES_DB

#   log "restoring DB from ${file}..."
#   cat "${file}" | docker-compose -f "${COMPOSE_FILE}" exec -T db \
#     psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" \
#     || error "restore failed"

#   docker-compose -f "${COMPOSE_FILE}" start yoru-pastebin
#   log "✓ restore complete"
# }

# logs() {
#   local service="${1:-}"
#   docker-compose -f "${COMPOSE_FILE}" logs -f --tail=100 "${service}"
# }

# status() {
#   log "containers:"
#   docker-compose -f "${COMPOSE_FILE}" ps
#   echo
#   log "resource usage:"
#   docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"
# }

# stop() {
#   log "stopping services..."
#   docker-compose -f "${COMPOSE_FILE}" down
#   log "✓ stopped"
# }

# restart() {
#   log "restarting..."
#   docker-compose -f "${COMPOSE_FILE}" restart
# }

# cleanup() {
#   log "cleaning docker resources..."
#   docker image prune -f
#   read -p "remove volumes? (yes/no): " confirm
#   [[ "${confirm}" == "yes" ]] && docker volume prune -f
#   docker network prune -f
#   log "✓ cleanup done"
# }

# health() {
#   log "checking health..."
#   docker-compose -f "${COMPOSE_FILE}" ps | grep -i "unhealthy" && warn "some containers unhealthy"
#   df -h / | awk 'NR==2{print $5}' | grep -q '[8-9][0-9]%' && warn "disk usage high"
#   free | awk '/Mem:/ {print int($3/$2 * 100)}' | grep -q '^9[0-9]' && warn "memory usage high"
#   log "✓ basic health check complete"
# }

# show_help() {
#   cat <<EOF
# usage: $0 [command]

# commands:
#   setup       initialize folders, .env, and network
#   deploy      build and start app containers
#   backup      dump db via backup profile
#   restore     restore DB from a SQL file
#   logs        show logs (optionally pass service name)
#   status      show service status + docker stats
#   stop        stop all containers
#   restart     restart containers
#   cleanup     remove unused docker resources
#   health      basic disk/mem/container check
# EOF
# }

# main() {
#   cd "${PROJECT_DIR}"
#   case "${1:-}" in
#     setup)     check_requirements && setup ;;
#     deploy)    check_requirements && deploy ;;
#     backup)    backup ;;
#     restore)   restore "${2:-}" ;;
#     logs)      logs "${2:-}" ;;
#     status)    status ;;
#     stop)      stop ;;
#     restart)   restart ;;
#     cleanup)   cleanup ;;
#     health)    health ;;
#     help|--help|-h|"") show_help ;;
#     *) error "unknown command: $1" ;;
#   esac
# }

# main "$@"
