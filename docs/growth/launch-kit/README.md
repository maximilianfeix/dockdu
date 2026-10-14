# Launch kit (v0.2.0)

Everything here is checked against the code and the 2026-10-14 validation runs. Nothing has been posted.

**Rule:** posts on Hacker News, Reddit, Lobsters, the Docker forums and dev.to are written by the maintainer, from these facts and outlines. Several of these venues ban or require disclosure of AI-written text; see [../community-research.md](../community-research.md). The GitHub release text is the only finished draft, since it goes into this repository.

| File | Use |
|---|---|
| [fact-sheet.md](fact-sheet.md) | Every claim you may make, with its source |
| [outlines.md](outlines.md) | Technical article, Show HN, Reddit, X, LinkedIn, dev.to: structure and points, not prose |
| [directories.md](directories.md) | awesome-docker, console.dev, later lists: criteria and status |
| [release-v0.2.0.md](release-v0.2.0.md) | GitHub release text |
| [faq.md](faq.md) | Answers to the questions launch threads will ask (security, read-only, accuracy) |

## Positioning

- **Headline:** Where did your Docker disk space go?
- **One line:** dockdu shows what `docker system df` misses (container logs, build cache shared with images, storage per compose project) without deleting anything.
- **When to use something else:**
  - interactive cleanup: lazydocker or the Docker Desktop Disk Usage extension;
  - continuous monitoring with overlay2 and bind-mount sizes: Doku;
  - per-builder cache detail: `docker buildx du`;
  - image layer efficiency: dive.

## Demo summary

`scripts/demo-up.sh` on a disposable daemon, recorded 2026-10-14:
- **Logs:** 109 MiB of container logs that `docker system df` does not show; one container at 107 MiB, unbounded.
- **Build cache, Engine 29:** 140.7 MiB that `docker builder prune` would not free, because it is shared with an image.
- **Build cache, Engine 28:** the same cache appears as 96 B in df's row; dockdu explains why.
- **Teardown** removes only labelled objects.

GIF: `docs/demo.gif`; full output: `docs/verification/v0.2-validation-2026-10-14.md`.
