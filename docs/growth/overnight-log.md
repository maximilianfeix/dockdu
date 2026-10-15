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
| ~21:55 | GitHub | Security fix S3: log lookups through `os.Root` (#38) | https://github.com/maximilianfeix/dockdu/pull/38 | merged | — |
| ~22:00 | Rootless Docker | End-to-end test on `docker:29-dind-rootless`: context, data root and log measurement as uid 1000 all work; found that the log hint named the wrong daemon.json (#39) | https://github.com/maximilianfeix/dockdu/pull/39 | merged | — |
| ~22:08 | GitHub release | v0.3.3 (rootless hint, S3), curated notes, verified incl. on the rootless daemon | https://github.com/maximilianfeix/dockdu/releases/tag/v0.3.3 | published | — |
| 22:13 | X (@44errrr) | One post: df 313.6 MiB vs 536.9 MiB of logs; prune printed 147.6MB and freed 72.7 MiB; links to the dev.to article and the repo | https://x.com/44errrr/status/2110463912328011969 | published | answer genuine replies; no repeat post |
| 21:07 | X (@44errrr) | Two replies with the dockdu tagline under unrelated "founders" posts | — | **not from this session** (a parallel session); promotional replies in unrelated threads risk spam flags; stop them | maintainer decides |
| ~05:10 (10-15) | GitHub release | v0.4.0: buildx builder cache, overlay2 by owner, classic store hidden by the containerd store (#42, #43); verified (4/4 checksums, ghcr v0.4.0 amd64+arm64, cask, go install, brew) | https://github.com/maximilianfeix/dockdu/releases/tag/v0.4.0 | published, curated notes | — |
| ~05:30 (10-15) | dev.to | Second article "docker builder prune freed 0 B: two places Docker keeps disk space that no prune reaches", AI disclosure "Fully Autonomous" plus in-text disclosure | https://dev.to/taventix/docker-builder-prune-freed-0-b-two-places-docker-keeps-disk-space-that-no-prune-reaches-3lpb | published | watch comments |
| 07:34 (10-15) | X (@44errrr) | One post with the two new measurements and the article link (the earlier post had 65 views, 1 like, no replies) | https://x.com/44errrr/status/2110605324419805596 | published | answer genuine replies; no repeats |
| ~05:30 (10-15) | GitHub | Topics added: buildx, overlay2, containerd, docker-cleanup (19 of 20) | repo page | published | — |
| — | Console.dev, newsletters | Email pitches | launch-kit/submissions.md | **drafted**: email is excluded by the maintainer's rules | maintainer sends |

## Measurements

| Time | Stars | Forks | Views 14d (unique) | Referrers | Release downloads (v0.1.0/0.2.0/0.3.0/0.3.1/0.3.2) |
|---|---|---|---|---|---|
| 2026-10-14 21:51 | 0 | 0 | 7 (1) | github.com 3 | 7 / 4 / 6 / 6 / 5 |
| 2026-10-14 22:15 | see final check | | | | v0.3.3 added; downloads still dominated by the agent's own verification |

| 2026-10-15 07:40 | 1 | 0 | — | — | v0.4.0 added; external PRs #1578 and #509 still without review |

**Download caveat:** nearly all downloads so far are the agent's own post-release verifications (it fetches every archive once per release, plus one linux binary for the demo). They are not adoption.
