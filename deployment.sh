#!/bin/bash
set -e

echo "💿 building latest image..."
docker build -t yoru-pastebin:latest .

echo "📤 pushing image to local docker engine..."
# push for  a registry

echo "🚀 redeploying stack..."
if ! docker service update --force yoru_yoru 2>/dev/null; then
  echo "⚠️  service update failed, removing and redeploying..."
  docker service rm yoru_yoru
  sleep 2
fi

docker stack deploy -c docker-compose.prod.yml yoru
