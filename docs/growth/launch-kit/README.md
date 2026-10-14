# Launch kit (v0.3.0)

Everything here is checked against the code (v0.3.0) and the measurements in [../../verification/demo-2026-10-14.md](../../verification/demo-2026-10-14.md). Nothing has been posted.

**Rule:** posts on Hacker News, Reddit, Lobsters, the Docker forums and dev.to are written by the maintainer, from these facts and outlines. Several of these venues ban or require disclosure of AI-written text; see [../community-research.md](../community-research.md). The GitHub release text is the only finished draft, since it goes into this repository.

| File | Use |
|---|---|
| [fact-sheet.md](fact-sheet.md) | Every claim you may make, with its source |
| [outlines.md](outlines.md) | Technical article, Show HN, Reddit, X, LinkedIn, dev.to: structure and points, not prose |
| [directories.md](directories.md) | awesome-docker, console.dev, later lists: criteria and status |
| [release-v0.3.0.md](release-v0.3.0.md) | GitHub release text for v0.3.0 (v0.2.0 text kept for reference) |
| [../../articles/docker-disk-usage-vs-prune.md](../../articles/docker-disk-usage-vs-prune.md) | The technical article: problem first, measured, references; dockdu only at the end |
| [../measurement.md](../measurement.md) | How to measure the launch without telemetry, with the baseline |
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

`LOG_MB=500 scripts/demo-up.sh` on a disposable Engine 29.9.0 daemon, dockdu v0.3.0 release binary, 2026-10-14:
- `docker system df` reports **313.6 MiB**.
- dockdu finds **537.0 MiB of container logs** that no df row contains. `du -ch` on the log files: 535.1M.
- **All 140.7 MiB of build cache is shared** with an image. On this containerd image store, `docker builder prune -af` printed 147.6MB and freed 72.7 MiB on disk. On Engine 28 / overlay2 the same kind of records freed about 0.

GIF: `docs/demo.gif` (23 s); screenshot: `docs/demo.png`; recipe: `demo.tape`.

## Before posting anywhere

- [ ] The search for existing posts was clean on 2026-10-14: no dockdu post on Hacker News (Algolia search) and no mention on GitHub outside this repo. Reddit could not be searched by script; check your own Reddit account history for an earlier dockdu post before posting again.
- [ ] Upload `docs/social-preview.png` under Settings → Social preview.
- [ ] Post once per venue, written by you; link the article or the repo, not both in every comment.
