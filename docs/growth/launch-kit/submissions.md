# Submissions tracker

Statuses: prepared → approved → sent/opened → response → listed/covered. Nothing is marked sent unless it actually was.

| Channel | Mechanism | Status | Link / date |
|---|---|---|---|
| awesome-docker | Pull request (Observability, between DLIA and docker-exporter; `./awesome-docker validate` passed locally) | **PR opened** 2026-10-14 (authorized). CI waits for maintainer approval of first-time-contributor workflows | https://github.com/veggiemonk/awesome-docker/pull/1578 |
| Console.dev | Email to hello@console.dev (only pre-1.0 tools) | **Prepared, not sent.** Emails need the maintainer | text below |
| dev.to article | Publish `docs/articles/docker-disk-usage-vs-prune.md` with an AI-assistance disclosure (dev.to guideline) | **Blocked:** no dev.to session in the browser; dockdu's automation may not log in or create accounts | — |
| Reddit | Text posts written by the maintainer | **Blocked for automation:** reddit.com is not reachable through the browser tooling (safety restriction), and rules could not be read; see community-research.md | — |
| Hacker News (Show HN) | Maintainer writes the submission and the first comment (generated text is banned) | Outline in outlines.md | — |

## Console.dev pitch (draft, for hello@console.dev)

> **Subject:** dockdu (v0.3, beta): where Docker's disk space went, read-only CLI
>
> Hi Console team,
>
> dockdu is an open-source Go CLI (MIT, v0.3.0, pre-1.0) for a recurring Docker problem: the disk is full, but `docker system df` reports a fraction of it. It measures what df leaves out and never deletes anything:
> - container log files, with their rotation settings;
> - build cache shared with images;
> - storage per Compose project;
> - optionally, the data root on disk.
>
> It's an interesting case because the numbers depend on the Docker version. On a disposable Engine 29.9 host, df reported 313.6 MiB while the container logs alone took 537 MiB. `docker builder prune` printed "147.6MB" and freed 72.7 MiB on the containerd image store, versus about 0 on overlay2 (Engine 28). The measurements and a 23-second demo are in the repository.
>
> Self-service: `brew install maximilianfeix/tap/dockdu`, `go install`, or a container image. No signup and no telemetry; the Docker client refuses any non-GET request.
>
> https://github.com/maximilianfeix/dockdu
>
> Thanks for considering it,
> Maximilian
