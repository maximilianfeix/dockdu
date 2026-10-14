# Campaign and engineering log

Statuses:
- **drafted:** text prepared, not submitted;
- **submitted:** sent to a third party, pending;
- **published:** publicly visible;
- **accepted / rejected:** the third party's decision;
- **blocked:** not possible with the available access or rules.

Times are CEST.

| Time (2026-10-14) | Channel | Action | URL | Status | Follow-up |
|---|---|---|---|---|---|
| ~20:30 | dockdu | Measured `docker builder prune` on overlay2 vs containerd; found that v0.2.0 overstated "prune would NOT free" on the containerd store | docs/verification/demo-2026-10-14.md | fixed in v0.3.0 (#29) | — |
| ~20:45 | GitHub release | v0.3.0 (containerd fix, `--disk`) | https://github.com/maximilianfeix/dockdu/releases/tag/v0.3.0 | published, verified | — |
| ~21:00 | GitHub | Social preview uploaded (Settings) | repo page | published | — |
| ~21:05 | GitHub | Topics added: container-logs, disk-space, docker-desktop, self-hosted, sysadmin, linux, golang | repo page | published | — |
| ~21:10 | awesome-docker | PR "Add dockdu to Observability" (validator passed locally) | https://github.com/veggiemonk/awesome-docker/pull/1578 | submitted | answer review comments; the maintainer merges in batches every ~3–4 weeks |
| ~21:20 | GitHub release | v0.3.1 (80-column output, least-privilege workflows, x/sys) | https://github.com/maximilianfeix/dockdu/releases/tag/v0.3.1 | published, verified | — |
| ~21:30 | GitHub release | v0.3.2 (a size was split from its unit when wrapping, found in the real v0.3.1 output) | https://github.com/maximilianfeix/dockdu/releases/tag/v0.3.2 | published, verified; curated release notes | — |
| ~21:35 | README | Demo GIF, screenshot and sample re-recorded with v0.3.2 on a disposable Engine 29.9 host (#36) | README | published | — |
| ~21:40 | dev.to | Article "Why docker system df doesn't explain your full disk, and why prune frees less than it says", AI disclosure "Fully Autonomous" plus an in-text disclosure | https://dev.to/taventix/why-docker-system-df-doesnt-explain-your-full-disk-and-why-prune-frees-less-than-it-says-27gc | published | watch comments |
| ~21:45 | GitHub | Ruleset on `main`: PR required, checks `test`/`integration`/`analyze`, no force push or deletion | repo settings | active | — |
| ~21:48 | awesome-cli-apps-in-a-csv | PR adding dockdu to `data/apps.csv` (category vm, next to ctop/dive/lazydocker) | https://github.com/toolleeo/awesome-cli-apps-in-a-csv/pull/509 | submitted | answer review comments |
| ~21:50 | Terminal Trove | Web submission form, partly filled | https://terminaltrove.com/submit/ | **blocked**: the agent's permission check refused further form input ("public data-sharing upload"); the draft is saved in the browser | maintainer finishes it: tagline, technical details, image `docs/demo.gif`, categories, install, author confirmation |
| ~21:40 | Reddit | Read rules / post | — | **blocked**: reddit.com is not reachable through the browser tooling | maintainer posts by hand, see launch-kit/outlines.md |
| — | X (@44errrr) | Post about the measured finding | — | **deferred**: the account posted a hexvero announcement at ~21:40; spacing posts | one post later, see below |
| — | Console.dev, newsletters | Email pitches | launch-kit/submissions.md | **drafted**: email is excluded by the maintainer's rules | maintainer sends |

## Measurements

| Time | Stars | Forks | Views 14d (unique) | Referrers | Release downloads (v0.1.0/0.2.0/0.3.0/0.3.1/0.3.2) |
|---|---|---|---|---|---|
| 2026-10-14 21:51 | 0 | 0 | 7 (1) | github.com 3 | 7 / 4 / 6 / 6 / 5 |

**Download caveat:** nearly all downloads so far are the agent's own post-release verifications (it fetches every archive once per release, plus one linux binary for the demo). They are not adoption.

## X post (ready, not yet published)

> docker system df on a test host: 313.6 MiB.
> Container logs on the same host: 536.9 MiB, in no df row.
> And `docker builder prune` printed 147.6MB but freed 72.7 MiB (containerd image store; ~0 on overlay2).
> Measured, with commands to reproduce:
> https://dev.to/taventix/why-docker-system-df-doesnt-explain-your-full-disk-and-why-prune-frees-less-than-it-says-27gc
