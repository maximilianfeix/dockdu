#!/usr/bin/env bash
# Removes what scripts/demo-up.sh created, by name, after checking the
# dockdu.demo=1 label. No prune command is used, so nothing else can go.
# The BuildKit cache records of the demo build stay (the cache cannot be
# removed by label without a prune); they are small and age out with your
# normal `docker builder prune` routine.
set -euo pipefail

# labelled TYPE NAME: true when the object exists and carries dockdu.demo=1.
labelled() {
  local field=.Config.Labels
  [ "$1" = volume ] && field=.Labels
  [ "$(docker "$1" inspect --format "{{index $field \"dockdu.demo\"}}" "$2" 2>/dev/null)" = 1 ]
}

for c in dockdu-demo-chatty dockdu-demo-bounded dockdu-demo-web-1 dockdu-demo-worker-1; do
  if labelled container "$c"; then docker rm -f "$c" >/dev/null && echo "removed container $c"; fi
done
for v in dockdu-demo_data dockdu-demo-orphan; do
  if labelled volume "$v"; then docker volume rm "$v" >/dev/null && echo "removed volume $v"; fi
done
docker network rm dockdu-demo_default >/dev/null 2>&1 && echo "removed network dockdu-demo_default" || true
if labelled image dockdu-demo/app:1; then docker image rm dockdu-demo/app:1 >/dev/null && echo "removed image dockdu-demo/app:1"; fi
echo "alpine:3.22 was left in place (it may have existed before)."
