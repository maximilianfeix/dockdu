# Launch and growth plan

Date: 2026-10-14. Nothing in this plan has been posted. Every external step needs the maintainer's go-ahead and is written by him.

## Ground rules

- **No manipulation.** No vote requests, no friends upvoting, no second accounts, no reposting the same text, no fake users or testimonials.
- **Write every post yourself.** HN ("don't post generated text"), the Docker Community Forums (accounts suspected of AI content are removed), Lobsters ("meaningful human authorship") and awesome-selfhosted (LLM submissions banned) all restrict machine-written text, and dev.to requires disclosure. Sources and quotes are in [community-research.md](community-research.md). The material below is a **fact sheet and outline** to write from, not text to paste.
- **Lead with the problem, not the tool.** Every piece must be useful to someone who never installs dockdu.
- **Ship v0.2 first.** The Engine 28 accounting fix (P0-1) affects many Linux servers, the exact audience a launch reaches. Launching v0.1.0 now would show those readers wrong numbers.

## Who adopts first

1. **Homelab and self-hosting operators on Linux** with compose stacks: logs per compose service is the clearest win, and they run as root.
2. **CI runner maintainers:** build cache and logs fill runners, and versioned JSON plus exit code 2 fit scripts.
3. **Docker Desktop users on macOS:** they hit Docker.raw confusion, but Docker Desktop's own Settings and Disk Usage extension already cover part of it. Treat them as secondary.

## Fact sheet (verified; quote from here)

- json-file has `max-size` -1 (unlimited) and `max-file` 1 by default ([docs](https://docs.docker.com/engine/logging/drivers/json-file/)). `docker system df` has no row for log files; upstream requests have been open since 2017 (moby #32205) and 2018 (moby #38341).
- Build cache records marked Shared are also part of images:
  - on Engine ≥ 29 (API ≥ 1.52), df's build cache row includes them;
  - docker CLI 28 leaves them out (`cli/command/system/df.go`);
  - `docker buildx du` counts them as reclaimable, the daemon does not;
  - pruning them drops metadata only.
- Docker.raw: Finder and `ls -l` show the configured maximum; `ls -ks` or Docker Desktop's Settings → Advanced show the real allocation (Docker Mac FAQ). The file shrinks only when Desktop compacts it.
- dockdu limits:
  - default builder only;
  - no overlay2-orphan or containerd-content attribution;
  - no Windows binary;
  - logs need read access to the data root (on Desktop, use the container one-liner).
- dockdu read-only claim: enforced by a request guard (GET/HEAD on four endpoints) plus tests; `:ro` on the socket is not what makes it safe.

## 30 / 60 / 90 days

| Window | Goal | Actions (maintainer) | Success signal |
|---|---|---|---|
| Days 0–30 | v0.2 out, honest and verified | Review and merge `v0.2-hardening`. Fix the tap token. Run the demo on a Linux Engine 28 host and on Engine 29 and record results. Regenerate the GIF. Tag v0.2.0. Then: awesome-docker PR (one line, by hand), console.dev email (pre-1.0 window), and one written article (topic 1 of the calendar) on dev.to with a disclosure if AI helped. | Install works on all four documented paths. First external issue or PR. awesome-docker PR reviewed. |
| Days 31–60 | One focused launch | Show HN with the repo link and a first comment you write (story, arithmetic, limits); be present for 3 hours. Answer existing Docker forum and r/docker disk-space questions where dockdu genuinely answers them, with the manual commands first. Ask r/docker mods before any post. | Comments with real-world reports ("found 40 GB of logs"). Bug reports from hosts unlike yours. |
| Days 61–90 | Earn word of mouth | Fix what launch users reported. Publish article 2 or 3 from real reports (with permission). Consider the containerd-store attribution (A4) if users ask. r/selfhosted full post only after the 3-month age rule (if it applies) and with real before/after numbers. | Returning issues and PRs from people you don't know, references from other people's posts. |

## Metrics

Public or owner-visible data only. No telemetry will be added.

| Metric | Source | Available? |
|---|---|---|
| Stars, forks, watchers | `gh repo view --json stargazerCount,forkCount` | yes |
| Views and clones (14 days) | `gh api repos/maximilianfeix/dockdu/traffic/views` (owner token) | yes, owner only |
| Release downloads | `gh api repos/maximilianfeix/dockdu/releases --jq '.[].assets[].download_count'` | yes |
| ghcr pulls | GitHub package page | partially (UI only) |
| Homebrew installs | analytics only for homebrew-core, not taps | **unavailable** |
| `go install` count | proxy.golang.org does not publish counts | **unavailable** |
| Successful first runs / misleading measurements | issues and discussions only | qualitative |

On 2026-10-14: 0 stars, 0 forks, 0 issues. No download numbers were recorded in this session.

## Assumptions not yet verified

- Engine 28 is still common on Linux servers. Plausible, but no data was gathered.
- Homelab users will run a CLI as root rather than a web UI like Doku.
- The containerd image store is not yet the default on Linux Engine 29 installs. If it is, the df-vs-disk gap there is partly outside dockdu.
- A technical first comment matters more on HN than the title.
