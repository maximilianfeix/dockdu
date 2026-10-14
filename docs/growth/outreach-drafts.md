# dockdu: outreach contacts and DRAFT messages

> **ALL MESSAGES BELOW ARE DRAFTS. NONE HAS BEEN SENT.** Nobody was contacted, and no account was used.

Researched 2026-10-14. These are the people who published about Docker disk usage, logs, build cache, Docker Desktop disk or Docker CLI tools in 2024–2026. Every channel listed is one the person or publication advertises publicly. I did not guess any email address.

**Before sending anything:**
- Rewrite each draft in your own voice. Several of these people publicly dislike AI-written outreach: selfh.st's editor mocked an "AI disclaimer", and SRE Weekly leaves "agent submissions" unread.
- Send at most one message per person.
- Do not follow up more than once.
- Never ask for stars or coverage. Ask for technical feedback.

**Verified numbers you may quote.** Sources: repo `docs/verification/demo-2026-10-14.md`, `CHANGELOG.md` 0.3.0, `docs/growth/launch-kit/fact-sheet.md`.
- **Engine 29.9 demo:** `docker system df` reports 313.6 MiB; dockdu finds 537.0 MiB of container logs that no df row contains (`du`: 535.1M).
- **Build cache shared with images:**
  - Engine 29.9, containerd image store: `docker builder prune -af` freed 72.7 of 140.7 MiB. A second run freed 56.8 of 108.7 MiB.
  - Engine 28.5, overlay2: about 0 freed.
- **Measurement method:** dockdu calls only GET/HEAD on `/_ping`, `/info`, `/system/df` and `/containers/<id>/json`. It measures logs with stat only.

## Priority order

| # | Contact | Why first | Channel |
|---|---|---|---|
| 1 | Ian L. Paterson | wrote about exactly this failure mode in June 2026 | site contact page |
| 2 | Ethan Sholly (selfh.st) | spotlighted oxker; core homelab audience | submission form |
| 3 | Console.dev team | pre-1.0 window, reviewed dive | hello@console.dev |
| 4 | Ákos Takács | Docker Captain and forum moderator; Desktop/VM expertise | it-sziget.hu, X @rimelek, dev.to |
| 5 | Peter Cooper (Go Weekly) | features Go CLIs, incl. Docker-related ones | editor@cooperpress.com |
| 6 | Bobby Borisov (Linuxiac) | 14 GB log-file article | linuxiac.com/contact |
| 7 | Kyle Galbraith (Depot) | build-cache cleanup guide updated Dec 2025 | X @kylegalbraith (from his GitHub profile) |
| 8 | Vladimir Chemeris (StorageRadar) | Docker-on-Mac disk article 2026 | storageradar.chama.pro/contacts/ |
| 9 | Jonas Scholz (Sliplane) | "5 Awesome Docker Tools" roundup | GitHub @code42cate / sliplane.io |
| 10 | amerkurev (Doku maintainer) | closest tool; README compares against Doku | GitHub Discussions on amerkurev/doku |

Alternates, if one of the above is not a fit:
- **Jonathan Hall, Cup o' Go:** the podcast has a "Lightning round" of Go tools; reach via `#cup-o-go` on the Gopher Slack.
- **Hamza Musa, Medevel:** covered Doku on 2025-07-31; reach via the site's "About & Contact".

---

## 1. Ian L. Paterson: homelab and infrastructure blogger
- **Published work:** [V] "Why docker image prune Won't Fix a Full Root Disk: The Containerd Snapshotter Trap", 2026-06-19, updated 2026-08-15. https://ianlpaterson.com/blog/docker-image-prune-containerd-disk-full/
  - It opens: "Root on my home server hit 99 percent with 2.1 GB free, and docker image prune freed nothing because it only deletes untagged images."
  - Per search snippets, the post also notes that image layers live in `/var/lib/containerd`, so moving the data-root did not help [S].
- **Channel:** [V] "Contact" link in the site navigation (https://ianlpaterson.com/contact/). The site footer also links LinkedIn and X. Whether the contact page has a form is not confirmed.
- **Why dockdu fits:** He hit the containerd image store trap that dockdu 0.3 now flags:
  - JSON field `daemon.image_store`;
  - `shared_may_be_freed_by_prune`;
  - `--disk`, which reports "outside the data root" on the same filesystem. That is where `/var/lib/containerd` would appear if it sits on the same filesystem as the data root (inference).
- **Finding to share:** On the containerd store, BuildKit records marked shared with images are partly freed by prune (72.7 of 140.7 MiB). On overlay2 they freed none.
- **DRAFT (not sent):**
  > Subject: Your containerd snapshotter post, and a read-only measurement tool
  >
  > Hi Ian, your post on `docker image prune` freeing nothing on a 99%-full root matched something I measured while building dockdu, a read-only Go CLI that explains Docker disk usage. On Engine 29.9's containerd store, `docker builder prune -af` freed 72.7 of 140.7 MiB of "shared" cache; on overlay2 it freed nothing. `dockdu --disk` also prints how much of the data-root filesystem lies outside `/var/lib/docker`. Would that number have pointed you at `/var/lib/containerd`, or would it mislead on your layout? Any correction is welcome. https://github.com/maximilianfeix/dockdu

## 2. Ethan Sholly: selfh.st / Self-Host Weekly
- **Published work:** [V]
  - "This Week in Self-Hosted (13 October 2023)": "a spotlight on Oxker, a text-based user interface for container management". https://selfh.st/weekly/2023-10-13/
  - Self-Host Weekly 2026-01-23, which discusses the Docker manager Dockhand, featured earlier in New Software. https://selfh.st/weekly/2026-01-23/
- **Channel:** [V] https://selfh.st/contact/: "Interested in sharing a project you've been working on? Click the button below to be redirected to my submission form". Also Bluesky and Mastodon.
- **Why dockdu fits:** It answers the homelab "compose stack filled my disk" question per compose project, and never deletes anything.
- **Finding to share:** In the Engine 29.9 demo, logs of 537 MiB were invisible to a df report of 313.6 MiB.
- **Caution:** He criticised a repo's "AI disclaimer". Keep the README and the pitch factual and avoid hype.
- **DRAFT (not sent; for the submission form):**
  > dockdu is a read-only CLI (Go, MIT, v0.3) for the "my Docker host is full but `docker system df` says otherwise" problem. It groups storage by Compose project, lists container log files with their rotation settings, and flags unreferenced non-Compose volumes, with one suggested command per finding, never executed. On a test host df reported 313.6 MiB while container logs took 537 MiB. It runs as a binary or as a `FROM scratch` container with read-only mounts. I'd value feedback from self-hosters on whether the per-project view matches how they think about their stacks. https://github.com/maximilianfeix/dockdu

## 3. Console.dev: developer tools newsletter
- **Published work:** [S] review of **dive**, 2023-10-19 (https://console.dev/tools/dive). [V] Criteria: "We only list early access, alpha, or beta releases"; "Any GA or stable releases are not eligible."
- **Channel:** [V] "Email hello@console.dev" (https://console.dev/selection-criteria).
- **Why dockdu fits:** It is pre-1.0, a CLI with JSON output, with no signup or telemetry, and every engine request is GET/HEAD.
- **Finding to share:** The same prune command behaves differently by image store: shared build cache freed on the containerd store, not on overlay2.
- **Note:** A longer pitch already exists in repo `docs/growth/launch-kit/submissions.md`. This version is shorter and different.
- **DRAFT (not sent):**
  > Subject: dockdu v0.3 (beta): what `docker system df` leaves out
  >
  > Hi Console team, dockdu is a pre-1.0, MIT-licensed Go CLI that reports the Docker disk usage df omits: container log files with rotation status, build cache shared with images, storage per Compose project, and the Docker Desktop VM file's real allocation. It is read-only by construction; its client rejects anything but GET/HEAD on four endpoints. One detail I found useful: the same `docker builder prune` frees shared cache on the containerd image store but not on overlay2. If you try it, I'd appreciate hearing where the output is unclear. https://github.com/maximilianfeix/dockdu

## 4. Ákos Takács (rimelek): Docker Captain, Docker forum moderator
- **Published work:** [V] dev.to/rimelek:
  - "Which Docker variant am I using and where is the daemon running?" (2024-12-26)
  - "You run containers, not dockers…" (2024-10-27)
  - "Everything about Docker volumes" (2024-01-07)

  His GitHub bio [V]: "Docker Captain, Docker forum moderator, content creator". He also answered Docker.raw growth reports on forums.docker.com (t/147832) [S].
- **Channel:** [V] blog https://it-sziget.hu (from his GitHub profile), X @rimelek, and dev.to. Do **not** DM him through the forum about a tool. The forum bans promotional/AI content.
- **Why dockdu fits:** dockdu resolves the daemon the same way the CLI does (DOCKER_HOST, then DOCKER_CONTEXT, then the current context). It treats Docker Desktop's VM file and its in-VM logs separately, which is exactly the "where is the daemon running" problem he writes about.
- **Finding to share:** On Docker Desktop, log files live inside the VM, so a host run exits 2 and marks them as unmeasured. The container one-liner with a read-only `/var/lib/docker` mount measures them.
- **DRAFT (not sent):**
  > Hi Ákos, your post on which Docker variant you're using and where the daemon runs helped me get context resolution right in dockdu, a read-only CLI that explains Docker disk usage. On Desktop it reports Docker.raw's apparent vs allocated size and marks log files as unmeasured from the host, since they live in the VM; a `docker run` with read-only mounts measures them from inside. You've answered many Docker.raw questions on the forum: is there a Desktop or context setup where this approach gives a wrong or confusing answer? Technical criticism very welcome. https://github.com/maximilianfeix/dockdu

## 5. Peter Cooper: Go Weekly (Cooperpress)
- **Published work:** [V] Go Weekly's "Code & Tools" section:
  - #559: "Unregistry: Push Docker Images Directly to Remote Servers" (https://golangweekly.com/issues/559)
  - #599: Plow 1.4, a benchmarking TUI
  - #620 (2026-10-09): "Boring: A Single-Binary SSH Tunnel Manager"
- **Channel:** [V] cooperpress.com/contact lists "Editorial editor@cooperpress.com". [V] Issue #620 says the next issue is 2026-10-23.
- **Why dockdu fits:** A single static Go binary with two direct dependencies (`moby/moby/client` v0.6.2 and api v1.56.1, not the deprecated `docker/docker` module). It has a request-guard hook that enforces read-only access.
- **DRAFT (not sent):**
  > Subject: Go CLI: dockdu, a read-only Docker disk usage report
  >
  > Hi Peter, in case it's useful for Code & Tools: dockdu is a Go CLI that explains Docker disk usage that `docker system df` doesn't show, such as container logs and build cache shared with images. It's a static binary with two direct dependencies, built on the new `moby/moby/client` module rather than the deprecated `docker/docker` one, and a transport hook rejects every request except GET/HEAD on four endpoints. I'd also be glad of feedback from any reader on that read-only enforcement approach. https://github.com/maximilianfeix/dockdu

## 6. Bobby Borisov: Linuxiac
- **Published work:** [S] "Reducing Docker Logs Size: A Practical Guide to Log Management", 2024-07-21. https://linuxiac.com/reducing-docker-logs-file-size/
  - It begins with "a Docker container had accumulated a 14 GB log file over just a year" and uses find/du to list log sizes.
  - It was republished on LinuxToday on 2024-08-02.
- **Channel:** [V] https://linuxiac.com/contact/ (contact form).
- **Why dockdu fits:** It automates his find/du step. It attributes the files to containers and Compose services, includes rotated `*-json.log.N` files that a simple glob misses, and shows whether rotation is bounded.
- **Finding to share:** In the demo, `du` on the `*-json.log` glob gave 535.1M and dockdu gave 537.0 MiB. The difference is the rotated files the glob skipped.
- **DRAFT (not sent):**
  > Hi Bobby, your 2024 article about a 14 GB container log matches the case I built dockdu for: a read-only CLI that lists each container's log files, including rotated ones, with driver and rotation limits, grouped by Compose project. In a test, a `*-json.log` du glob found 535.1M while the full set was 537.0 MiB; the difference was rotated files the glob missed. If you try it on a real server, I'd like to know whether the suggested `daemon.json` fix and "new containers only" note are clear enough. https://github.com/maximilianfeix/dockdu

## 7. Kyle Galbraith: Depot co-founder
- **Published work:** [V] "How to clear Docker cache and free up space on your system", Depot blog, 2024-06-14, updated 2025-12-02. https://depot.dev/blog/docker-clear-cache . It was cross-posted to dev.to. It recommends `docker buildx prune` for build cache.
- **Channel:** [V] His GitHub profile lists X `@kylegalbraith` and depot.dev. Public replies or DMs on X are the professional channel. No email is listed.
- **Why dockdu fits:** dockdu covers the step before pruning: how much of the build cache is shared with images, and whether pruning it frees anything on this daemon.
- **Finding to share:**
  - The engine's `Reclaimable` excludes shared records; `buildx du` counts them.
  - On Engine 29.9 / containerd, prune freed about half of the shared bytes (72.7 of 140.7 MiB); on Engine 28.5 / overlay2, none.
- **DRAFT (not sent):**
  > Hi Kyle, your Docker cache-clearing guide is one I point people to. While building dockdu, a read-only Docker disk report, I measured something that might interest you: build cache records marked shared freed nothing under `docker builder prune -af` on Engine 28.5/overlay2, but about half (72.7 of 140.7 MiB) on Engine 29.9's containerd store. Does that match what you see on BuildKit hosts, or is there a case I'm not modelling? Notes and numbers: https://github.com/maximilianfeix/dockdu

## 8. Vladimir Chemeris: StorageRadar blog
- **Published work:** [V] "Docker Disk Usage on Mac: What Actually Eats Space", published 2026-02-18, updated 2026-06-01. https://storageradar.chama.pro/blog/docker-disk-usage-on-mac-what-actually-eats-space . It covers where Docker Desktop stores data, prune, and why deleting Docker folders directly is risky.
- **Channel:** [V] site links `/contacts/` and `/support/` (https://storageradar.chama.pro/contacts/).
- **Context:** He sells a Mac storage app, so the overlap may be commercial. Ask only for technical review, never for a mention.
- **Finding to share:** Finder and `ls -l` show Docker.raw's configured maximum. dockdu shows apparent vs allocated size next to Docker's own resource total (Docker Mac FAQ). Logs inside the VM need the container run.
- **DRAFT (not sent):**
  > Hi Vladimir, I read your Docker-on-Mac disk article. I maintain dockdu, a read-only CLI that reports Docker.raw's apparent vs allocated size next to what Docker itself accounts for, and lists container logs, which on Desktop are only measurable from inside the VM. You clearly test macOS storage reporting carefully: does reading `st_blocks` for the allocated figure hold up on APFS in your experience, or are there cases (snapshots, clones) where it misleads? https://github.com/maximilianfeix/dockdu

## 9. Jonas Scholz: Sliplane co-founder
- **Published work:** [V] "5 Awesome Docker Tools To Make Your Life Easier" (lazydocker, dive, OrbStack, Watchtower, Sliplane). It was published on dev.to on 2024-11-12, edited 2025-02-01, and originally on sliplane.io. https://dev.to/code42cate/5-awesome-docker-tools-to-make-your-life-easier-22kl
- **Channel:** [V] GitHub @code42cate (bio "co-founder @sliplane, i like container and microvms"), sliplane.io, dev.to.
- **Why dockdu fits:** It is a read-only complement to lazydocker, which prunes on request. The output is pasteable into support threads, which is useful for a container-hosting company's users (inference).
- **DRAFT (not sent):**
  > Hi Jonas, your Docker tools roundup (lazydocker, dive and co.) is close to how I think about the toolbox. I built dockdu for the gap between them: a read-only report of what `docker system df` misses, such as log files, build cache shared with images, and per-Compose-project totals, with JSON and exit code 2 when data is partial. Since you run containers for customers, I'd value your take on one question: is the JSON output useful for support or CI, or is something essential missing? https://github.com/maximilianfeix/dockdu

## 10. amerkurev: Doku maintainer
- **Published work:** [V]
  - Doku v1.0.0 (2025-03-11) through v1.3.0 (2025-12-26). https://github.com/amerkurev/doku/releases
  - Doku is listed in awesome-docker *Observability* and awesome-homelab *Monitoring*, and was covered by Medevel (2025-07-31).
- **Channel:** [V] GitHub Discussions are enabled on amerkurev/doku. His GitHub profile lists https://python.study and no email.
- **Why reach out:** dockdu's README and `docs/audits/competitive-landscape.md` make claims about Doku, for example that it measures only the current `LogPath` file. Asking him to correct them is honest and useful. Do not pitch dockdu to Doku's users in his Discussions.
- **DRAFT (not sent; Discussions post, category "General" or similar):**
  > Hi, I maintain dockdu, a read-only CLI for Docker disk usage, and its README compares itself with Doku. From reading `app/scan/scanner.py` I wrote that Doku sizes each container's current log file but not rotated `-json.log.N` siblings, and that it has a per-record "shared" flag for build cache but no shared total. Are those accurate for v1.3? I'd rather fix my comparison than misdescribe your project. Also, Doku's overlay2 in-use/unused split covers something dockdu doesn't; I link to Doku for that case.
