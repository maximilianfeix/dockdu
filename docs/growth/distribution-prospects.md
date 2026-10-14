# dockdu: distribution prospects

Researched 2026-10-14 for dockdu v0.3.0 (0 stars, repo created 2026-10-13). Nothing was posted, submitted or sent, and no accounts were used.

**Markers**
- **[V]**: I read the rule or page myself on 2026-10-14, either via `gh`, a git clone, or a page fetch.
- **[V-repo]**: verified on the same day in `docs/growth/community-research.md` and not re-fetched here.
- **[S]**: from search-engine snippets only, because the site blocked scripted fetches.
- **[U]**: unverified.

**Ranking.** Prospects are ordered by fit × realistic access today × evidence that the venue covered similar tools (Doku, dive, lazydocker, ctop, oxker, or Docker CLI/TUI tools). Audience size is not part of the score.

**Cross-cutting rule.** These venues explicitly restrict AI-written submissions:
- HN, Docker forums, awesome-selfhosted, awesome-sysadmin, dev.to and Lobsters ([V-repo] / [V]);
- SRE Weekly: "I'd prefer to receive link submissions written by humans, not agents" [V].

The editor of selfh.st publicly mocked a repo's "AI disclaimer" (see #3). Every submission text should be written by Maximilian himself.

## Summary table

| # | Prospect | Type | Access today | Status / blocker |
|---|---|---|---|---|
| 1 | awesome-docker | list | open | **PR #1578 already open** (2026-10-14) |
| 2 | Console.dev | newsletter | email | eligible only while < 1.0 |
| 3 | selfh.st Self-Host Weekly | newsletter | form | precedent: oxker spotlight |
| 4 | Go Weekly (Cooperpress) | newsletter | email | precedent: Unregistry, Plow |
| 5 | Show HN | community | self-post | once only; human text only |
| 6 | Terminal Trove | directory | form | needs preview + package repo [S] |
| 7 | awesome-cli-apps-in-a-csv | list | issue/PR | no age/star gate found |
| 8 | Docker Community Forums | community | answers | no AI content; no promo category |
| 9 | dev.to article | blog | self-post | AI disclosure required |
| 10 | Changelog News | aggregator | form (sign-in) | own work allowed |
| 11 | Cup o' Go | podcast | Gopher Slack | process undocumented |
| 12 | Linuxiac | news site | contact form | covered Docker logs in 2024 |
| 13 | TerminalDock | directory | form (sign-in) | rules unknown |
| 14 | r/docker | community | mod approval [U] | Reddit rules unreadable |
| 15 | r/selfhosted | community | megathread [U] | < 3 months → megathread only [U] |
| 16 | r/golang | community | [U] | Reddit rules unreadable |
| 17 | Medevel | blog | contact [U] | covered Doku in 2025 |
| 18 | noted.lol | blog | contribute page [U] | covered Doku in 2023 |
| 19 | Tecmint | blog | [U] | Docker tools roundup 2024 |
| 20 | DevOps Weekly | newsletter | email [U] | no submission policy |
| 21 | DevOps'ish | newsletter | contact [U] | no submission policy |
| 22 | LibHunt (Go) | directory | submit (login) | thin evidence |
| 23 | AUR / nixpkgs / MacPorts | packaging | PR | [U] rules; oxker uses AUR + MacPorts [S] |
| 24 | Homebrew core | packaging | PR | **blocked**: 225 stars for self-submit |
| 25 | awesome-homelab (ccbikai) | list | none working | **issues disabled**, no PRs |
| 26 | OpenAlternative | directory | form (sign-in) | weak fit (alternatives to proprietary) |
| 27 | AlternativeTo | directory | [U] | weak fit |
| 28 | Lobsters | community | invite-only | needs invite; self-promo < 25% |
| 29 | r/devops, r/homelab | community | [U] | low fit / unverified |
| 30 | SRE Weekly | newsletter | email | low fit; humans-only links |
| 31 | TLDR DevOps | newsletter | none | **no submission route** (ads only) |
| 32 | awesome-go | list | PR | **blocked until ~2027-03-13** |
| 33 | awesome-devops | list | PR | **blocked**: 100 stars + 6 months |
| 34 | awesome-tuis | list | PR | **ineligible**: not a TUI; 6 months |
| 35 | awesome-sysadmin | list | PR | **blocked**: 12 months + "not your own" |
| 36 | awesome-shell | list | PR | **blocked**: 50 stars |
| 37 | awesome-cli-apps (agarrharr) | list | none | **archived** |
| 38 | awesome-selfhosted | list | PR | **ineligible** (network services, 4 months, LLM ban) |
| 39 | modern-unix, awesome-compose | lists | n/a | not a fit (inactive / samples only) |

---

## Tier A: act now

### 1. awesome-docker: https://github.com/veggiemonk/awesome-docker
- **Status:** [V] PR [#1578](https://github.com/veggiemonk/awesome-docker/pull/1578) "Add dockdu to Observability" was opened by maximilianfeix on 2026-10-14. It is open, has no comments yet, and changes only README.md.
- **Policy:** [V] `.github/CONTRIBUTING.md`:
  - "If you removed the Docker integration, would the project still have a reason to exist?"
  - "Tool that monitors containers, container resources, or container logs → Accept"
  - "one sentence, lead with the verb"
  - "The project has been active in the last 2 years."
  - The PR template is headed "IF THE PROJECT JUST RUNS IN DOCKER, AUTOMATICALLY DENIED".
  - No AI rule, minimum age or star requirement was found.
- **Similar tools already listed:** [V]
  - **Doku** in *Observability* (README line 335, "monitor Docker disk usage"), the same section as the dockdu PR;
  - **dive, dockly, dry, lazydocker, oxker** under *User Interfaces → Terminal*.
- **Evidence of throughput:** [V] [#1514](https://github.com/veggiemonk/awesome-docker/pull/1514) "Add bosun (Terminal)" was opened on 2026-07-28 and merged on 2026-08-18. #1526 and #1523 were merged the same day. Inference: the maintainer merges in batches, roughly every 3–4 weeks. Four other "Add …" PRs (#1574–#1577) are open ahead of dockdu.
- **Effort:** done. Respond to review only, in your own words.
- **Why it fits:** it is the closest list match, and the Doku precedent puts a disk-usage tool in the same section.

### 2. Console.dev: https://console.dev/selection-criteria
- **Mechanism:** [V] "Email hello@console.dev". The address is published on the criteria page.
- **Policy:** [V]
  - "We only list early access, alpha, or beta releases of tools where developers are the primary user."
  - "Any GA or stable releases are not eligible."
  - "We do not do sponsored reviews."
  - The criteria list "an API and/or CLI" as a plus.
  - No AI rule.
- **Evidence:** [S] console.dev reviewed **dive** on 2023-10-19 (https://console.dev/tools/dive).
- **Effort:** low. A short email draft exists in `docs/growth/launch-kit/submissions.md`; rewrite it in your own voice.
- **Why it fits:** v0.3.0 is pre-1.0, so the window closes at 1.0. Read-only, no signup and JSON output all match their criteria.

### 3. selfh.st Self-Host Weekly: https://selfh.st/contact/
- **Mechanism:** [V] The contact page says: "Interested in sharing a project you've been working on? Click the button below to be redirected to my submission form". There is also an "E-mail Me" button, Bluesky and Mastodon. The form URL is behind a JavaScript button and was not resolved.
- **Policy:** [V] From the 2026-01-23 issue: "I'm always looking for new and existing self-hosted content to share… Reach out using the button below". No written criteria.
- **Evidence:** [V]
  - The 2023-10-13 issue was "a spotlight on Oxker, a text-based user interface for container management" (https://selfh.st/weekly/2023-10-13/).
  - Its "New Software" section featured the Docker manager Dockhand (Dec 2025, discussed in https://selfh.st/weekly/2026-01-23/).
- **Caution:** [V] In that issue, editor Ethan Sholly said he had "mentally wrote the project off" partly because "its repo features an AI disclaimer that reminds me of the older trend…". Keep the pitch factual and human-written.
- **Effort:** low. **Why it fits:** homelab compose users are dockdu's core audience, and the oxker precedent shows Docker CLI/TUI tools get covered.

### 4. Go Weekly (formerly Golang Weekly): https://golangweekly.com
- **Mechanism:** [V] cooperpress.com/contact lists "Editorial editor@cooperpress.com". The editor is Peter Cooper (named in issue #620). There is no per-newsletter submission form [V].
- **Policy:** [U] No published criteria. Paid sponsorship exists, but it is not the route for coverage.
- **Evidence:** [V] the "Code & Tools" section regularly features Go CLIs:
  - #559: "Unregistry: Push Docker Images Directly to Remote Servers";
  - #599: "Plow 1.4… a TUI";
  - #620 (2026-10-09): "Boring: A Single-Binary SSH Tunnel Manager".
- **Timing:** [V] #620 says "we're taking another break next week… see you again on Friday, October 23".
- **Effort:** low. **Why it fits:** a Go single-binary CLI with two dependencies and a measured technical story (BuildKit shared records).

### 5. Hacker News: Show HN: https://news.ycombinator.com/showhn.html
- **Policy:** [V-repo]
  - "Don't post quickly-generated one-offs"
  - "Please don't put generated text in HN posts"
  - "Don't solicit upvotes"
  - A version bump is not a new Show HN.
- **Evidence (HN Algolia API) [V]:**
  - ctop: 323 points, 73 comments, on 2017-03-09, two days after its v0.3 release, posted by the vektorlab account.
  - lazydocker: 340 points, 47 comments, on 2019-06-30, posted by the author.
  - dive: 684 points on 2018-11-25, posted by a third party.
  - No dockdu post exists [V-repo].
- **Effort:** medium. Write the post and first comment yourself, and be online for 2–3 hours. **Fit:** strong, if the first comment is technical (the df-vs-logs measurement).

### 6. Terminal Trove: https://terminaltrove.com/submit/
- **Mechanism:** [S] web form. Requirements from snippets:
  - "Must have an image preview (PNG, GIF or MP4)";
  - "Must not exist already on Terminal Trove";
  - cross-platform and standalone binaries preferred;
  - install instructions per package manager;
  - "The tool also has to exist in the package repositories" [S]; whether a personal Homebrew tap counts is [U];
  - the form asks "are you the author?".
- **Evidence:** [S]
  - a Docker category (https://terminaltrove.com/categories/docker) lists ctop and dive;
  - oxker has a page (https://terminaltrove.com/oxker/);
  - "Tool of the Week" featured dcv, a docker-compose TUI (September 2025 wrap-up).
- **Note:** the site returned HTTP 403 to scripted fetches, so check the form in a browser. **Effort:** low (`docs/demo.gif` exists).

### 7. awesome-cli-apps-in-a-csv: https://github.com/toolleeo/awesome-cli-apps-in-a-csv
- **Mechanism:** [V] CONTRIBUTING.md:
  - "open an issue including… name, homepage, git, description", or a PR that edits `data/apps.csv` only;
  - "Please make changes to the CSV file only, not to the README file."
- **Policy:** [V] No age, star or AI rule found. Active: merges on 2026-09-14 and 2026-07-21. 2,641 stars.
- **Similar tools:** [V] **ctop, dive, Dockly, dry, lazydocker** in "Containerization and virtualization" (README ~l.1292). Disk analyzers dua, gdu and ncdu are also listed.
- **Effort:** low. **Fit:** good, as a Docker CLI next to existing Docker CLIs.

### 8. Docker Community Forums: https://forums.docker.com
- **Policy:** [V-repo] "Don't post AI generated content. We will remove accounts suspected of this." There is no show-and-tell category.
- **Evidence of demand:** [V-repo] threads t/151053 (2026-02), t/137291 (44,747 views) and others. [V] The regular answering t/151053 is "Metin Y." (meyay), a forum "Leader".
- **Approach:** answer existing questions with manual commands first, and mention dockdu only where it fits. **Effort:** ongoing.

### 9. dev.to article: https://dev.to
- **Policy:** [V-repo] AI-assisted articles must "Disclose the fact that they were generated or assisted by AI". The Terms ask for content "not designed primarily for the purposes of promotion".
- **Evidence:** [V] Jonas Scholz, "5 Awesome Docker Tools To Make Your Life Easier" (dev.to, 2024-11-12, edited 2025-02-01), which covers lazydocker and dive.
- **Effort:** medium. The draft is `docs/articles/docker-disk-usage-vs-prune.md` (repo).

### 10. Changelog News: https://changelog.com/news/submit
- **Policy:** [V]
  - "Please sign in / up to submit news."
  - Submissions of your own work are welcome: "if your fellow devs will find it interesting, submit it."
  - Not accepted: "How-to's and tutorials", "Commercial products or services".
  - You get an email only if it is published.
- **Evidence:** [U] no Changelog coverage of lazydocker, dive or ctop was found. **Effort:** low, but it needs an account, which only Maximilian can create.

## Tier B: plausible, lower evidence or harder access

### 11. Cup o' Go (weekly Go news podcast): https://cupogo.dev
- **Channel:** [V] "Join us on Slack: #cup-o-go on the Gopher Slack". The site also has an Email link.
- **Evidence:** [V] each episode has a "⚡ Lightning round" of releases and community tools. For example, E162 (June 2026) featured "gaal — Keep skills in sync across AI agents". Host Jonathan Hall appears in 170 episodes [V].
- **Submission process:** [U]. **Fit:** a Go CLI with a release; low effort.

### 12. Linuxiac: https://linuxiac.com/contact/
- **Channel:** [V] a contact form ("Your constructive criticism and honest feedback are invaluable…"). Its news section covers software releases [V nav].
- **Evidence:** [S] Bobby Borisov, "Reducing Docker Logs Size: A Practical Guide to Log Management" (2024-07-21). It starts from "a Docker container had accumulated a 14 GB log file over just a year". Republished on LinuxToday (2024-08-02).
- **Policy on project tips:** [U]. **Fit:** strong topic overlap with the logs finding.

### 13. TerminalDock: https://terminaldock.com/submit
- **Facts:** [V]
  - submitting requires sign-in ("Get access to the dashboard and manage your submitted tools");
  - its newsletter claims "5,000+ developers";
  - it lists lazydocker (https://terminaldock.com/lazydocker);
  - it is made by Piotr Kulpinski, who also runs OpenAlternative.
- **Rules:** [U], including whether there is a paid expedite. **Effort:** low, but it needs an account.

### 14–16. r/docker, r/selfhosted, r/golang
- **Rules:** [U]. Reddit blocks scripted access [V-repo]. A secondary source says r/docker allows self-promotion "only with prior moderator approval", and r/selfhosted puts projects younger than 3 months into a "New Project Megathread" [U].
- **Note:** r/selfhosted has a widely discussed AI/"vibe-coded" rule whose wording is unverified [V-repo]. Read each sidebar before posting.
- **Fit:** high audience match, but access is unknown.

### 17. Medevel: https://medevel.com
- **Evidence:** [V] Hamza Musa, "Doku: A Free and Open-source Docker Disk Monitor" (2025-07-31). The site has an "About & Contact" nav. Submission rules [U]. **Fit:** they cover open-source Docker disk tools; dockdu is the CLI counterpart.

### 18. noted.lol: https://noted.lol
- **Evidence:** [V] Jeremy, "Doku – A Simple Self Hosted Docker Disk Usage Dashboard" (2023-02-09). The nav has "Contribute" and "Contact". Rules [U]. Older evidence.

### 19. Tecmint: https://www.tecmint.com
- **Evidence:** [V] Ravi Saive, "11 Must-Have Docker Tools To Simplify Your Workflow" (updated 2024-11-13), with LazyDocker as #1. Submission route [U].

### 20. DevOps Weekly: https://www.devopsweekly.com
- **Facts:** [V] Gareth Rushgrove's email appears only in the unsubscribe/privacy sentence. No submission policy. Treat it as [U]. Do not use the address unless a submission invitation is found.

### 21. DevOps'ish: https://devopsish.com/about/
- **Facts:** [V] Chris Short's newsletter, "Cloud Native, DevOps, Open Source…". The site has a Contact link. Submission policy [U].

### 22. LibHunt Go: https://go.libhunt.com
- **Facts:** [V] It has "Submit" and "Login" links and a "Go newsletter". [U] It mostly mirrors awesome-go, so a listing there likely follows awesome-go eligibility (inference).

### 23. Linux packaging: AUR, nixpkgs, MacPorts
- **Evidence:** [S] oxker is packaged in the AUR and MacPorts.
- **Rules:** [U] for nixpkgs/MacPorts. Anyone can publish to the AUR (common knowledge, [U] here).
- **Fit:** this is reach, not promotion. It also helps meet Terminal Trove's "package repositories" requirement [S].

### 24. Homebrew core: https://docs.brew.sh/Package-Acceptance-Policy
- **Blocker:** [V]
  - "at least 30 forks, 30 watchers or 75 stars";
  - "at least 90 forks, 90 watchers or 225 stars for a self-submission by the repository owner";
  - "A code repository less than 30 days old is normally not eligible."
- Keep the personal tap for now.

## Tier C: low fit or blocked (do not spend time now)

### 25. awesome-homelab (ccbikai): https://github.com/ccbikai/awesome-homelab
- **Similar tools:** [V] Doku is listed in Monitoring (README l.710).
- **Access:** [V] The README "Submit App" link points to GitHub issues, but **"the 'ccbikai/awesome-homelab' repository has disabled issues"**. `gh pr list` returned no PRs. The README is generated from `data/*.yaml` (AGENTS.md). Effectively curator-only.

### 26. OpenAlternative: https://openalternative.co/submit
- [V] Sign-in required. Scope: "open source alternatives to proprietary software". A ranking "health score" weighs popularity and age [V]. Weak fit: dockdu has no proprietary counterpart beyond Docker Desktop's extension.

### 27. AlternativeTo: https://alternativeto.net
- [U] Rules not fetched (minimal page). Weak fit for the same reason as #26.

### 28. Lobsters: https://lobste.rs/about
- [V-repo] Invite-only. "self-promo should be less than a quarter of one's stories and comments". New users cannot use the `show` tag for 70 days.

### 29. r/devops, r/homelab
- [U] (Reddit). Lower audience fit than r/docker and r/selfhosted.

### 30. SRE Weekly: https://sreweekly.com/about/
- **Policy:** [V]
  - "Feel free to reply to the email newsletter directly, or email me directly at my first name at this domain."
  - "I do not accept embargoed/pre-release content."
  - "I'd prefer to receive link submissions written by humans, not agents. I tend to leave agent submissions unrea[d]."
- **Fit:** it covers reliability articles, not tools. A dev.to article about logs filling a disk would be the only candidate.

### 31. TLDR DevOps: https://tldr.tech/devops
- [V] No submission route on the page, only "Advertise" (advertise.tldr.tech). **Blocked**, short of paid placement, which you excluded.

### 32. awesome-go: https://github.com/avelino/awesome-go
- **Blocker:** [V-repo] "at least 5 months of history", "≥80% coverage" with a report link, and a SemVer release. Not before about **2027-03-13**.

### 33. awesome-devops: https://github.com/wmariuss/awesome-devops
- **Blocker:** [V] "at least 100 stars and 6 months of history". An automated review also checks "signs of inflated stars", and "an AI reviewer checks that the description, section and tags match".

### 34. awesome-tuis: https://github.com/rothgar/awesome-tuis
- **Ineligible:** [V] "Repos need to be at least 6 months old"; "Interfaces should be interactive or re-draw output. Output formatters are not considered TUIs." dockdu prints a report, so it does not qualify even though ctop, dive, lazydocker and oxker are listed there [V].

### 35. awesome-sysadmin: https://github.com/awesome-foss/awesome-sysadmin-data
- **Blocker:** [V] the PR template requires:
  - "first released more than 12 months ago";
  - "Software you are submitting is not your own, unless you have a healthy ecosystem with a few contributors";
  - "intended for professional setup".
- CONTRIBUTING also says: "Machine/LLM-generated contributions, that do not respect project guidelines are not allowed and will result in a ban."
- The earliest date is about 2027-10, and only if someone else submits it.

### 36. awesome-shell: https://github.com/alebcay/awesome-shell
- **Blocker:** [V] "GitHub projects must have at least 50 stars… This applies to… any projects that you developed or authored yourself." Self-promotion is explicitly OK once that bar is met. ncdu is listed.

### 37. awesome-cli-apps (agarrharr): https://github.com/agarrharr/awesome-cli-apps
- [V] The repo is **archived** (`archived: true`). It still lists lazydocker, ctop, dockly, dua-cli and NCDu, but no new entries are possible.

### 38. awesome-selfhosted
- [V-repo] Lists network services only. It requires a first release "more than four months old", and "Machine/LLM-generated contributions are not allowed and will result in a ban." Poor fit.

### 39. modern-unix / docker/awesome-compose
- **modern-unix:** [V] last push 2024-09-10. It lists dust and duf, but appears inactive.
- **awesome-compose:** [V] "A curated list of Docker Compose samples". It is not a tool list.

## Blocking rules at a glance
- **Age gates:** awesome-tuis 6 months, awesome-devops 6 months, awesome-go 5 months, awesome-selfhosted 4 months, awesome-sysadmin 12 months, Homebrew core 30 days.
- **Star gates:** awesome-shell 50, awesome-devops 100, Homebrew core 75 (third party) or 225 (self-submitted).
- **Self-submission banned:** awesome-sysadmin ("not your own").
- **AI-content bans:** awesome-sysadmin and awesome-selfhosted (ban); HN, Docker forums and Lobsters (spam/removal); SRE Weekly (unread); dev.to (disclosure).
- **No route:** TLDR DevOps (ads only), awesome-homelab (issues disabled), awesome-cli-apps (archived).
- **Account required (Maximilian only):** Changelog News, TerminalDock, OpenAlternative, LibHunt, Reddit, dev.to.
