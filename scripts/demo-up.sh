#!/usr/bin/env bash
# Creates the three dockdu demo scenarios on the current Docker daemon.
# Every object is named dockdu-demo* and labelled dockdu.demo=1; nothing
# existing is touched. Remove with scripts/demo-down.sh (no prune involved).
#
#   A  container logs: one chatty container with the default json-file driver
#      (unbounded) and one with --log-opt max-size=1m --log-opt max-file=3
#   B  build cache: an image built with BuildKit whose layers stay in the
#      cache as records shared with the image
#   C  compose: project dockdu-demo with two services on one image, a named
#      volume, plus an orphan volume outside compose
set -euo pipefail

LOG_MB=${LOG_MB:-200}   # how much log output the chatty container writes
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

echo "== B: build an image with BuildKit (layers stay in the build cache)"
cat > "$work/Dockerfile" <<'DOCKERFILE'
FROM alpine:3.22
RUN dd if=/dev/urandom of=/blob bs=1M count=64 status=none
LABEL dockdu.demo=1
DOCKERFILE
docker buildx build --load -t dockdu-demo/app:1 "$work"

echo "== A: unbounded and bounded json-file logs"
docker run -d --name dockdu-demo-chatty --label dockdu.demo=1 alpine:3.22 \
  sh -c "head -c $((LOG_MB * 1024 * 1024)) /dev/zero | tr '\0' 'x' | fold -w 1023; sleep 86400"
docker run -d --name dockdu-demo-bounded --label dockdu.demo=1 \
  --log-opt max-size=1m --log-opt max-file=3 alpine:3.22 \
  sh -c "head -c $((20 * 1024 * 1024)) /dev/zero | tr '\0' 'y' | fold -w 1023; sleep 86400"

echo "== C: compose project with a shared image and volumes"
cat > "$work/compose.yaml" <<'COMPOSE'
name: dockdu-demo
services:
  web:
    image: dockdu-demo/app:1
    command: ["sleep", "86400"]
    labels: { dockdu.demo: "1" }
    volumes: [data:/data]
  worker:
    image: dockdu-demo/app:1
    command: ["sh", "-c", "dd if=/dev/zero of=/data/fill bs=1M count=32 status=none; sleep 86400"]
    labels: { dockdu.demo: "1" }
    volumes: [data:/data]
volumes:
  data:
    labels: { dockdu.demo: "1" }
COMPOSE
docker compose -f "$work/compose.yaml" up -d
docker volume create --label dockdu.demo=1 dockdu-demo-orphan >/dev/null

echo "waiting for the log writer..."
sleep 10
echo "done. Compare:"
echo "  docker system df"
echo "  dockdu                       # Linux host, as root"
echo "  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro -v /var/lib/docker:/var/lib/docker:ro ghcr.io/maximilianfeix/dockdu   # Docker Desktop"
