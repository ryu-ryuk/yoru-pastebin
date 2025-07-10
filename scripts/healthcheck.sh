#!/bin/bash
#  __                         __   __     __                     __                         __       
# |  \                       |  \ |  \   |  \                   |  \                       |  \      
# | ▓▓____   ______   ______ | ▓▓_| ▓▓_  | ▓▓____        _______| ▓▓____   ______   _______| ▓▓   __ 
# | ▓▓    \ /      \ |      \| ▓▓   ▓▓ \ | ▓▓    \      /       \ ▓▓    \ /      \ /       \ ▓▓  /  \
# | ▓▓▓▓▓▓▓\  ▓▓▓▓▓▓\ \▓▓▓▓▓▓\ ▓▓\▓▓▓▓▓▓ | ▓▓▓▓▓▓▓\    |  ▓▓▓▓▓▓▓ ▓▓▓▓▓▓▓\  ▓▓▓▓▓▓\  ▓▓▓▓▓▓▓ ▓▓_/  ▓▓
# | ▓▓  | ▓▓ ▓▓    ▓▓/      ▓▓ ▓▓ | ▓▓ __| ▓▓  | ▓▓    | ▓▓     | ▓▓  | ▓▓ ▓▓    ▓▓ ▓▓     | ▓▓   ▓▓ 
# | ▓▓  | ▓▓ ▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓ ▓▓ | ▓▓|  \ ▓▓  | ▓▓    | ▓▓_____| ▓▓  | ▓▓ ▓▓▓▓▓▓▓▓ ▓▓_____| ▓▓▓▓▓▓\ 
# | ▓▓  | ▓▓\▓▓     \\▓▓    ▓▓ ▓▓  \▓▓  ▓▓ ▓▓  | ▓▓     \▓▓     \ ▓▓  | ▓▓\▓▓     \\▓▓     \ ▓▓  \▓▓\
#  \▓▓   \▓▓ \▓▓▓▓▓▓▓ \▓▓▓▓▓▓▓\▓▓   \▓▓▓▓ \▓▓   \▓▓      \▓▓▓▓▓▓▓\▓▓   \▓▓ \▓▓▓▓▓▓▓ \▓▓▓▓▓▓▓\▓▓   \▓▓


# Used by Docker healthcheck

set -euo pipefail

# Configuration
HOST="${HEALTH_CHECK_HOST:-localhost}"
PORT="${SERVER_PORT:-8080}"
TIMEOUT="${HEALTH_CHECK_TIMEOUT:-10}"

# Perform health check
check_health() {
    # Check if the server responds
    if curl -f -s --max-time "${TIMEOUT}" "http://${HOST}:${PORT}/health" > /dev/null 2>&1; then
        echo "✓ Health check passed"
        exit 0
    else
        echo "✗ Health check failed"
        exit 1
    fi
}

check_health
