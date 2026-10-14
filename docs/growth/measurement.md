# Measuring adoption

No telemetry, now or later. Everything below is data GitHub, the Go proxy or the registries already have.

## Weekly (5 minutes, maintainer)

```
R=maximilianfeix/dockdu
gh repo view $R --json stargazerCount,forkCount --jq '"stars \(.stargazerCount) forks \(.forkCount)"'
gh api repos/$R/traffic/views --jq '"views 14d \(.count) unique \(.uniques)"'
gh api repos/$R/traffic/popular/referrers --jq '.[] | "\(.referrer) \(.count)/\(.uniques)"'
gh api repos/$R/traffic/popular/paths --jq '.[] | "\(.path) \(.count)"'
gh api repos/$R/releases --jq '.[] | "\(.tag_name) " + ([.assets[]|select(.name!="checksums.txt")|.download_count]|add|tostring)'
gh issue list --repo $R --state all --json author,title --jq '.[]|select(.author.login!="maximilianfeix")|.title'
```

Copy the numbers into the table below. GitHub keeps traffic for only 14 days, so record weekly or lose it.

## What each number means

| Signal | Source | Caveat |
|---|---|---|
| Views, unique visitors, referrers | Traffic API, 14 days | Includes your own visits. Referrers show which post or list worked. |
| Release archive downloads | Releases API | Includes the maintainer's verification downloads (about 3 per release on 2026-10-14). Homebrew installs count here, because the cask downloads the archive. |
| `go install` | none | proxy.golang.org publishes no counts. |
| ghcr pulls | package page in the GitHub UI | The API needs `read:packages`. |
| Clones | Traffic API | Mostly CI and the Go module proxy; not a user signal. |
| Stars, forks | Repo API | Lagging and noisy; never the goal. |
| **Independent issues and PRs** | Issues | **The strongest signal**: someone ran it on their host and cared enough to write. |
| Organic mentions | search "dockdu" on HN, Reddit, GitHub | Reddit search is blocked for scripts; check by hand. |

## Baseline (2026-10-14, before any launch post)

| Date | Stars | Forks | Views 14d (unique) | Referrers | Downloads v0.1.0 / v0.2.0 / v0.3.0 | External issues |
|---|---|---|---|---|---|---|
| 2026-10-14 | 0 | 0 | 7 (1) | github.com 3 | 7 / 4 / 0 | 0 |

## Questions to answer after each launch step

1. **Where did visitors come from?** Referrers in the week after.
2. **Did visitors install?** Downloads of the newest release.
3. **Did anyone report a real host?** New issues, especially the "numbers look wrong" form.
4. **What was misunderstood?** Recurring questions in comments: fix the README or the output, not the post.
