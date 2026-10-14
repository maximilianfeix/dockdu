> Research snapshot from 2026-10-14. Reddit rules could not be read first-hand; read every venue's live rules before posting. Nothing here has been posted.

# dockdu: community channels and evidence of demand

Researched 2026-10-14. I did not post anything, log in, or create accounts. **Verified** means I read the rule text on the official page that day. **Unverified** means secondary sources only.

**Reddit caveat:** reddit.com blocks both the web-search tool and unauthenticated `curl` (rules.json returned a challenge page). None of the Reddit rules below could be checked first-hand. Before posting, read each subreddit's live sidebar and `/about/rules`.

## Cross-cutting finding: AI-assisted content

Several venues now restrict machine-generated posts explicitly:
- HN bans generated text in posts and comments.
- Docker Community Forums bans AI-generated content and removes accounts suspected of it.
- awesome-selfhosted bans LLM-generated submissions, and its CONTRIBUTING.md tells AI agents not to draft entries or PR text.
- dev.to requires disclosure.
- Lobsters treats content "created without meaningful human authorship" as spam.

**Practical rule for dockdu:** every post, comment, and list submission should be written by Maximilian himself. Generated drafts should not be pasted. The tool itself being built with AI help is a separate question. Only HN's "quickly-generated one-offs" line and the Lobsters `vibecoding` tag touch on that.

## Channels

### Hacker News (Show HN)
- **Rules (verified, https://news.ycombinator.com/showhn.html):**
  - "Show HN is for something you've made that other people can play with."
  - "The project should be non-trivial. Don't post quickly-generated one-offs; anybody can do that now."
  - "The project must be something you've worked on personally and which you're around to discuss."
  - "Please make it easy for users to try your thing out, ideally without barriers such as signups or emails."
  - "New features and upgrades ("Foo 1.3.1 is out") generally aren't substantive enough to be Show HNs."
  - "Please don't ask friends to upvote or comment. That's not ok on HN."
- **General guidelines (verified, https://news.ycombinator.com/newsguidelines.html):**
  - "Please don't use HN primarily for promotion."
  - "Don't solicit upvotes, comments, or submissions."
  - "Please don't put generated text in HN posts."
  - "Don't post generated text or AI-edited text." (about comments)
- **AI:** restricted, both for generated post/comment text and for "quickly-generated one-offs".
- **Best format:**
  - Title in the form `Show HN: dockdu – find where Docker's disk space went (read-only CLI)`.
  - Link to the GitHub repo.
  - Add a first comment in his own words: the df-vs-disk story, how the shared build cache arithmetic works, what it does not do, and the limits (no Windows, default builder only).
  - Be online for the first 2–3 hours.
  - Post once. A v0.2 release does not justify a second Show HN.
- **Expectations (assumption):** most Show HNs get under 10 points. A niche, honest diagnostic tool can do well if the first comment is technical. Expect pointed questions about `du`, ncdu, and buildx du, and about overlay2/containerd.

### r/docker
- **Rules (unverified):** a mirror of the r/docker wiki rules (red.applefritter.com/r/docker/wiki/rules, returned 502 when I fetched it directly; the content comes from a search snippet) says self-promotion is "allowed only with prior moderator approval". It also says open-source tools may be shared if they provide genuine value, and link posts need a short description.
- **AI:** no rule found (unverified).
- **Format:** message the mods first and ask whether a post about an open-source, read-only diagnostic is OK. Then make a text post that leads with the problem and real numbers, with the link at the end.
- **Expectations:** moderate. Questions like "my disk is full" come up constantly there, so answering those threads helpfully, and linking only where it fits, may work better than one launch post. (Assumption.)

### r/selfhosted
- **Rules (unverified, secondary source https://launchit.fast/blog/where-to-promote-your-startup-on-reddit-2026, which says it checked on 2026-10-08):**
  - Projects younger than three months go only in the "New Project Megathread".
  - Promoted apps "must be production ready and have docs".
  - On Wednesdays, "dashboards and tools for self-hosters" are allowed even if they are not self-hosted.
- r/selfhosted has a widely discussed rule for AI-built ("vibe-coded") projects, but I **could not verify** its current wording. Check the sidebar.
- **Fit:** dockdu (created 2026-10-13) is under 3 months old, so it goes in the megathread only, if the secondary source is right.
- **Expectations:** a megathread entry gets little visibility. Homelab users with compose stacks are the core audience, so a full post after 3 months, with a real before/after, is the better shot.

### r/devops
- **Rules:** unverified (Reddit blocked). Commonly reported as strict on vendor and self-promotion posts. Treat this as an assumption.
- **Format:** a CI-runner angle ("our runners filled up: the logs and build cache df doesn't show") with JSON and exit codes. Expectations are low to moderate.

### r/homelab
- **Rules:** unverified. Assumption: it is focused on hardware and lab setups, and software promotion is a weaker fit than r/selfhosted. Low priority.

### Dev.to
- **Rules (verified):**
  - The Terms (https://dev.to/terms) ask for content "not designed primarily for the purposes of promotion or creating backlinks".
  - The AI guidelines (https://dev.to/guidelines-for-ai-assisted-articles-on-dev) say articles must "Disclose the fact that they were generated or assisted by AI in the post", "Be checked for factual accuracy before publishing", and must not "Contain educational content or information generated by AI" when the author lacks the knowledge. Bots or AI must not generate comments.
- **Format:** a tutorial-style article such as "Why `docker system df` says 11 GB when your disk says 138 GB". Explain json-file defaults, shared build cache, and Docker.raw sparse files, with dockdu as one of several ways to check, alongside the manual `du`/`ls -ks` commands. Use tags `docker`, `devops`, `go`, `selfhosted`.
- **Expectations:** steady long-tail search traffic. Spikes are unlikely.

### Hashnode
- **Rules (unverified, from search snippets of hashnode.com/code-of-conduct and the terms; the official guidelines URL returned 404):** they prohibit "using the platform primarily for self-promotion without contributing to the community". For AI: "If you use AI-assisted writing, review everything before publishing. You are responsible for the accuracy and originality".
- **Format:** cross-post of the dev.to article with a canonical URL. Expectations are low.

### Docker Community Forums (forums.docker.com)
- **Rules (verified, https://forums.docker.com/guidelines):**
  - "Don't post spam or otherwise vandalize the forum."
  - "Don't post AI generated content." The fuller sentence, as fetched: "Don't post AI generated content. We will remove accounts suspected of this."
  - There is no explicit self-promotion clause.
- **Categories (verified via /categories.json):** Announcements (Docker team only), General, Docker Desktop, Docker Hub, Docker Engine, Image Builds, Support. There is no "show and tell" category.
- **Format:** answer existing disk-space questions in his own words, with the manual commands first and a mention of dockdu as optional. Active regulars there (for example the moderator who answered t/151053) are skeptical of GUI tools and value precise explanations.
- **Expectations:** small but highly targeted, and good for credibility. The risk is being seen as spam if every answer links the tool.

### Lobsters
- **Rules (verified, https://lobste.rs/about):**
  - "As a rule of thumb, self-promo should be less than a quarter of one's stories and comments."
  - Membership is invite-only ("a user invitation tree to combat spam").
  - Users "are considered "new" for their first 70 days", and new users cannot use certain tags, including `show` and `vibecoding`.
  - The spam flag covers "content that is created without meaningful human authorship".
- **Fit:** he needs an invitation and should not post his own project as a first action. Expectations are low unless someone else submits it.

### awesome-docker (veggiemonk/awesome-docker, 37,032 stars; inspected @ `83cf001`, 2026-10-02)
- **Criteria (verified, .github/CONTRIBUTING.md):**
  - The "for Docker" test: "If you removed the Docker integration, would the project still have a reason to exist?"
  - "Tool that monitors containers, container resources, or container logs" → Accept.
  - "The project has been active in the last 2 years."
  - "one sentence, lead with the verb".
  - The PR template header reads "IF THE PROJECT JUST RUNS IN DOCKER, AUTOMATICALLY DENIED".
- **Fit:** strong. Doku, dive, and lazydocker are already listed there. Nothing about AI or a minimum age was found.
- **Action:** a one-line PR such as "Explain where Docker disk space went (logs, shared build cache, compose projects) without deleting anything." This is the **best near-term channel**.

### awesome-selfhosted (awesome-selfhosted/awesome-selfhosted, 326,523 stars; data repo @ `fdd1a1e`, 2026-10-14)
- **Criteria (verified, awesome-selfhosted-data CONTRIBUTING.md):**
  - "first release more than four months old".
  - "Machine/LLM-generated contributions are not allowed and will result in a ban."
  - Under "What does not qualify": "Software that is a desktop, mobile, or command-line application, which relies on a separate file synchronisation/server program" and "Anything that is a generic container/deployment automation/virtualization/... tool is better suited for awesome-sysadmin".
  - Its CONTRIBUTING.md has an explicit "AI AGENTS" block forbidding agents from drafting entries, PR text, or the human-attestation checkbox.
- **Fit:** poor. awesome-selfhosted lists self-hosted *network services*, and dockdu is a CLI diagnostic. It is not eligible before roughly 2027-02-13 in any case (v0.1.0 tagged around 2026-10-13). awesome-sysadmin is the more plausible list (its rules were not checked).

### awesome-go (avelino/awesome-go, 188,388 stars; inspected @ `9084f30`, 2026-10-14)
- **Criteria (verified, CONTRIBUTING.md):**
  - "at least 5 months of history"
  - "an open source license", "a `go.mod`", and "at least one SemVer release (`vX.Y.Z`)"
  - "Tests meet the coverage guideline (≥80% for non-data packages...)"
  - A pkg.go.dev link and a coverage report link (Codecov/Coveralls) in the PR body.
  - Automated "Repository maturity: First commit is at least 5 months old".
- **Fit:** eligible no earlier than about **2027-03-13**. Before then, add a coverage badge (Codecov/Coveralls) and make sure the doc comments render on pkg.go.dev. No AI rule was found.

### console.dev
- **Criteria (verified, https://console.dev/selection-criteria):**
  - "We only list early access, alpha, or beta releases of tools where developers are the primary user."
  - The version must be below 1.0 or carry a pre-release label. "Any GA or stable releases are not eligible."
  - Questions include "Is the primary user a developer?" and "Does it have good documentation?".
  - Submit by "Email hello@console.dev". "We do not do sponsored reviews."
- **Fit:** v0.1.0 qualifies as pre-1.0. The window closes at 1.0. Expectations: 2–3 tools are picked per week, so selection is uncertain.

### Golang Weekly (Cooperpress)
- **Submission:** golangweekly.com shows no submission instructions (verified). A 2020 issue says the editors are "only a 'Reply' button away" (https://golangweekly.com/issues/307). The current process is unverified. Paid sponsorship exists (cooperpress.com/mediakit).
- **Fit:** an article ("how dockdu reads BuildKit's shared records") is likelier to get picked than a bare repo link. Assumption.

### DevOps Weekly
- **Submission:** devopsweekly.com lists Gareth Rushgrove (gareth@morethanseven.net) only in a subscription context, with no submission instructions (verified). Emailing a short note is plausible but unverified as welcome.

### selfh.st newsletter
- **Submission:** selfh.st/weekly links "Submit Content" (https://selfh.st/submit/), but the fetched page showed only a subscribe form, so the criteria are unverified. They may require a form rendered by JavaScript. Check in a browser.
- **Fit:** good audience (homelab, compose). The weekly issue lists new software releases.

## Suggested order (assumption, based on the above)
1. awesome-docker PR (strong fit, no age gate). The PR text must be written by him.
2. console.dev email while the version is below 1.0.
3. A dev.to article written by him (disclose any AI assistance), then one Show HN that links the repo, with him available to answer comments.
4. Helpful answers on the Docker forums and in r/docker threads, linking the tool only where it fits.
5. Later: r/selfhosted full post after 3 months, awesome-go after 2027-03-13, and awesome-sysadmin (check its rules).

## Evidence of demand: real public threads

| # | Thread | Date / state | Signal |
|---|---|---|---|
| 1 | [moby/moby#52539](https://github.com/moby/moby/issues/52539) "`docker system df` shows 265.5GB used for images without any images" | opened 2026-05-05, open, 35 comments | df vs reality gap, very recent |
| 2 | [moby/moby#38848](https://github.com/moby/moby/issues/38848) "'docker system df' does not list all space being used" | 2019-03-10, open, 37 comments, 16 thumbs-up | long-standing df blind spot |
| 3 | [moby/moby#38341](https://github.com/moby/moby/issues/38341) "Include log space usage in `system df`..." | 2018-12-10, open | logs missing from df, requested upstream |
| 4 | [moby/moby#32205](https://github.com/moby/moby/issues/32205) "Docker system df shows wrong container size, -json.log size is not included?" | 2017-03-29, open | same, about logs |
| 5 | [moby/moby#45003](https://github.com/moby/moby/issues/45003) "docker system du and docker system prune reported sizes are confusing" | 2023-02-15, open | build cache / prune numbers confusing (a maintainer-style write-up) |
| 6 | [moby/moby#41472](https://github.com/moby/moby/issues/41472) "`docker system prune -a` isn't regaining reclaimable space" | 2020-09-18, open | prune frees less than shown |
| 7 | [moby/moby#47089](https://github.com/moby/moby/issues/47089) "Constantly growing /var/lib/docker/overlay2/" | 2024-01-17, open, 16 comments | disk grows while df looks fine |
| 8 | [forums.docker.com/t/151053](https://forums.docker.com/t/huge-disk-usage/151053) "Huge disk usage" | 2026-02-06, 15 posts, 2,939 views | homelab/LXC user, Portainer numbers vs disk; mentions containerd store |
| 9 | [forums.docker.com/t/151143](https://forums.docker.com/t/config-options-to-reduce-docker-raw-for-docker-sandboxes-on-docker-desktop-for-mac/151143) "Config options to reduce Docker.raw..." | 2026-02-19, 6 posts | Docker.raw size on Mac |
| 10 | [forums.docker.com/t/147964](https://forums.docker.com/t/docker-desktop-mac-ignoring-the-disk-usage-limit-i-specified-in-settings-resources/147964) "Docker Desktop (Mac) ignoring the disk usage limit" | 2025-05-02, 1,669 views | Docker.raw / VM disk confusion |
| 11 | [forums.docker.com/t/137291](https://forums.docker.com/t/docker-containers-taking-up-all-space-over-time/137291) "Docker containers taking up all space over time" | 2023-08-16, **44,747 views** | evergreen search demand |
| 12 | [forums.docker.com/t/138443](https://forums.docker.com/t/trouble-downsizing-docker-raw-using-docker-desktop/138443) "Trouble downsizing Docker.raw" | 2023-11-12, 7,848 views | Docker.raw does not shrink |
| 13 | [docker/for-mac#6865](https://github.com/docker/for-mac/issues/6865) "Docker Desktop reports incorrect VM Disk usage..." | 2023-06-02, open | Desktop VM disk reporting |

Thread metadata (dates, post counts, views) came from the Discourse `/t/<id>.json` API and `gh issue view` on 2026-10-14. I did not check Reddit threads because Reddit was inaccessible. Many content-farm articles from 2026 on "docker logs filling disk" (netdata.cloud, oneuptime.com, canadianwebhosting.com) indicate SEO demand for the log problem. They are not first-hand user reports.

**Reading of the evidence:**
- The df blind spot for logs has been requested upstream since 2017–2018 and is still open.
- Docker.raw confusion is recurring on the Docker forums.
- "prune frees less than expected" has open moby issues.
- Overlay2/containerd growth (threads 1, 7, 8) is the most active current pain, and it is **partly outside what dockdu v0.1 attributes**. Expect launch-thread commenters to raise it.
