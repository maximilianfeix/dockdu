# dockdu v0.2.0

**If you run Docker Engine 28 or older, upgrade.** v0.1 subtracted build cache that is shared with images twice on those engines: the total was too low, private cache showed as unknown, and every run warned and exited 2. v0.2 detects how the daemon counts that row and reports it correctly. This was checked on Engine 28.5.2 and 29.9.0.

## Changes
- **Read-only, enforced.** The Docker client refuses anything but reading `/_ping`, `/info`, `/system/df` and `/containers/<id>/json`, before a request is sent. Mounting the socket with `:ro` never gave that guarantee.
- **Docker contexts.** With no `DOCKER_HOST`, dockdu connects to the daemon your `docker` CLI uses: Colima, OrbStack, rootless, Docker Desktop on Linux. `ssh://` contexts are not supported yet.
- **More honest totals:**
  - unknown volume sizes and a build cache row that matches no known rule make the summary a lower bound;
  - base layers shared between images are reported in the compose section;
  - the image reclaimable figure is labelled as the Docker API's.
- **Safer suggestions.** Volume names from the daemon are shell-quoted.
- **JSON** now has `schema_version: 1`. Only fields were added.
- **Faster on large hosts.** Container inspects run on four workers.
- **Docker Desktop.** A host run prints one warning, not one per container, and points to the container image.

Full list: [CHANGELOG](https://github.com/maximilianfeix/dockdu/blob/v0.2.0/CHANGELOG.md). Validation: `docs/verification/v0.2-validation-2026-10-14.md`.

## Install
```
brew install maximilianfeix/tap/dockdu
go install github.com/maximilianfeix/dockdu/cmd/dockdu@v0.2.0
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro -v /var/lib/docker:/var/lib/docker:ro ghcr.io/maximilianfeix/dockdu:v0.2.0
```
Archives and `checksums.txt` are attached below.
