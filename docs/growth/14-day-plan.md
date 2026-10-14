# 14-day launch plan (from 2026-10-15)

Built only on channels whose status is verified in [launch-kit/submissions.md](launch-kit/submissions.md) and [distribution-prospects.md](distribution-prospects.md). Each day is a small, finished step. Record measurements with [measurement.md](measurement.md).

| Day | Action | Who | Depends on | Success signal |
|---|---|---|---|---|
| 1 | Merge #31 (launch docs), #32 (Scorecard), #33 (80-column output); release v0.3.1 so the README demo matches the current binary | maintainer approves, automation executes | approval | v0.3.1 published, post-release checks PASS |
| 1 | Enable a branch ruleset on `main` requiring `test`, `integration`, `analyze` | maintainer (settings) | — | Scorecard BranchProtection alert closes |
| 2 | Send the Console.dev email (text in submissions.md) | maintainer | — | reply or listing within 2–4 weeks |
| 2 | Watch awesome-docker#1578; answer review comments the same day | automation can answer while a session runs | maintainer review requests | merged |
| 3 | Publish the article on dev.to with the AI-assistance disclosure; canonical link to the repo docs | maintainer logs in (or posts) | #31 merged | referrers from dev.to in traffic |
| 4–5 | Answer 2–3 existing, recent "docker disk full" questions where the manual commands from the article solve them; mention dockdu only where it fits | maintainer (forums ban AI text) | article live | helpful replies, no removals |
| 6 | Show HN, written by the maintainer, with the first comment on the Engine 28 vs 29 prune measurements; be online for 3 hours | maintainer | v0.3.1, article | HN referrers, issues from new hosts |
| 7 | Weekly measurement (measurement.md); fix the top misunderstanding in the README or output | automation | — | table row added |
| 8–9 | Outreach to 3–5 contacts from outreach-drafts.md via their public channels, one at a time, personalised | maintainer sends | v0.3.1 | technical feedback |
| 10 | r/docker: ask the moderators first, then a problem-first text post written by the maintainer | maintainer | mod approval | discussion quality, not upvotes |
| 11–12 | Triage every issue and PR from new users within 24 h; release fixes | automation + maintainer | — | issues closed, second-time users |
| 13 | Second article angle (e.g. "Engine 28 vs 29: what docker builder prune really frees"), only if the first found readers | maintainer publishes | referrer data | — |
| 14 | Review: which referrers produced installs (downloads) and issues; drop channels with no signal | automation | 2 weeks of data | decision for the next 30 days |

**Not in the plan:**
- vote requests, star requests, cross-posting the same text, or new accounts;
- Reddit or HN posts written by automation;
- any channel whose rules were not read first.
