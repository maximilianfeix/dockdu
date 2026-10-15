# dockdu marketing plan (autopilot edition)

Written 2026-10-15. Consolidates `launch-plan.md`, `14-day-plan.md`, `distribution-prospects.md`,
`community-research.md` and `growth-case-studies.md` into one plan that an agent can execute on its own.

**North star:** GitHub stars and real installs from people other than the maintainer.
Baseline on 2026-10-15: 1 star, 5 unique visitors in 14 days, v0.5.0. Downloads so far are mostly our own
post-release checks, so they do not count.

## Positioning

- **One line:** "Docker says 2 GB is reclaimable. dockdu shows where the other 36 GB went."
- **For:** developers on Docker Desktop / Colima / OrbStack whose disk is full; homelab and small-server admins;
  CI maintainers whose runners run out of space.
- **Against the alternatives:**
  - `docker system df`: no log size, wrong reclaimable numbers ([docker/cli#2715](https://github.com/docker/cli/issues/2715),
    [#1751](https://github.com/docker/cli/issues/1751), [#4656](https://github.com/docker/cli/issues/4656)).
  - Doku: a web dashboard, unmaintained since 2025-12, open bugs with Docker 29 overlayfs.
  - dive / ncdu: per-image or per-path, no engine-wide attribution.
  - dockdu: one read-only command, every byte attributed (logs, build cache, buildx volumes, anonymous volumes,
    overlay2 leftovers), with the exact command to reclaim it.
- **Proof points (measured, keep the evidence links):**
  - df 313.6 MiB vs 536.9 MiB of logs
  - prune printed 147.6 MB but freed 72.7 MiB
  - `docker builder prune` frees ~half of "shared" cache on the containerd store

## What the agent does on its own

Rules: no spam, no replies under other people's posts, no AI-written promo on platforms that forbid it,
one PR per awesome list, everything committed as the owner without AI attribution.

### Week 1: make the repo convert visitors

1. **README above the fold:** the one-liner, the before/after numbers, a 20–30 s GIF, three install lines
   (brew / go install / docker run), a comparison table vs df/Doku/dive/ncdu.
2. **"Why is my Docker disk full?" section with anchors** that answer docker/cli #2715, #1751 and #4656, plus a
   Docker 29 overlayfs/containerd explainer. These pages are what search and AI answers will cite.
3. **GitHub Pages landing page** with the same story, an OpenGraph image and `SoftwareApplication` schema.
4. **Repo metadata:** description, topics (`docker`, `disk-usage`, `cli`, `devops`, `docker-desktop`, `colima`,
   `orbstack`, `homelab`), social preview, a pinned discussion "Share your dockdu output".

### Week 2: product hooks that bring people back

5. `--fail-over SIZE` exit code for CI/cron, so it stays installed.
6. A Markdown report (`--format md`) for pasting into issues and forums.
7. Per-compose-project totals; ignore lists; `--context a,b` (asked for by Doku users as "multiple hosts").
8. Real validation on Colima, OrbStack, Podman and WSL2 where possible, recorded in `docs/verification/`.
   Untested platforms are listed honestly in the README.

### Weeks 2–4: distribution the agent may do itself

9. **Package managers:** AUR, nixpkgs, Scoop, and homebrew-core once its notability rules are met.
10. **Awesome lists:** one PR each where CONTRIBUTING allows self-submission and the gates are met. Watch
    awesome-docker#1578 and awesome-cli-apps-in-a-csv#509 and answer reviews the same day.
    Never awesome-selfhosted (it bans LLM-written entries).
11. **X (@44errrr):** one informational post on a release day with real news and the GIF. No replies, no mentions.
12. **In-repo articles** (`docs/articles/`) that answer the upstream issues; linked from the README and Pages.

### Owner queue (drafts only, never posted by the agent)

Hacker News, Reddit (r/docker, r/selfhosted, r/homelab), Lobsters, the Docker forums, newsletters (selfh.st,
Console.dev) and dev.to promotion forbid or punish AI-written self-promotion. For these the agent keeps a fact sheet in
`~/autopilot/owner-queue.md`, with:
- the measured claims and their evidence links
- three title ideas, written as facts
- the best posting time (US morning on a weekday)

The owner writes and posts these in his own words. Earlier runs posted promotional dev.to articles and X replies;
the owner should rewrite or remove those.

## Measurement (every Monday in `docs/growth/measurement.md`)

Stars, forks, unique visitors and referrers (`gh api repos/maximilianfeix/dockdu/traffic/...`), brew/ghcr downloads
minus our own checks, issues opened by others.

| Target | Date |
|---|---|
| 10 stars and the first outside issue | 2026-10-29 |
| 50 stars | 2026-11-30 |
| 3 awesome lists merged | 2026-11-30 |

If a channel brings no visitors after two weeks, drop it.

## Release cadence

At most one release per week, only with a change users can see, and with notes that explain it. No patch bursts.
