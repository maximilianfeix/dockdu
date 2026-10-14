# How comparable tools got early traction

Researched 2026-10-14.

**Labels.** Each line is one of:
- **Fact:** with a link or the API it came from.
- **Inference:** my reading of the facts.
- **[S]:** search snippet only.

**Sources.**
- HN data: the public HN Algolia API (`hn.algolia.com/api/v1/search`) and the Firebase item API, for points, comment counts and submitters.
- Repo dates, stars and release lists: `gh api` on 2026-10-14.
- READMEs: shallow clones.
- Reddit could not be read: scripted access is blocked. No Reddit launch numbers appear below, and none are invented.

## Snapshot (2026-10-14, `gh api`)

| Tool | Repo created | First GitHub release | Releases | Latest release | Stars |
|---|---|---|---|---|---|
| lazygit | 2018-05-19 | n/a (not pulled) | n/a | n/a | 83,175 |
| dive | 2018-05-13 | v0.0.1, 2018-10-17 | 28 | v0.13.1, 2025-03-29 | 54,636 |
| lazydocker | 2019-05-18 | v0.1.3, 2019-06-18 | 47 | v0.25.2, 2026-04-19 | 53,122 |
| k9s | 2019-01-25 | 0.1.0, 2019-02-02 | 258 | v0.51.0, 2026-06-06 | 34,813 |
| ctop | 2016-12-27 | v0.3, 2017-03-07 | 16 | v0.7.7, 2022-03-22 | 17,843 |
| dust | 2018-03-16 | – | – | – | 12,497 |
| dua-cli | 2019-05-29 | – | – | – | 6,348 |
| gdu | 2018-02-24 | – | – | – | 6,107 |
| oxker | 2022-04-25 | v0.0.1, 2022-04-25 | 53 | v0.13.4, 2026-08-22 | 1,912 |
| Doku | 2022-06-13 | v0.0.14, 2022-07-23 | 7 | v1.3.0, 2025-12-26 | 451 |

"First GitHub release" is the earliest *GitHub release object*. Earlier git tags may exist.

## Hacker News launch and repost history (Algolia API)

| Date | Tool | Points | Comments | Submitter | Title (shortened) |
|---|---|---|---|---|---|
| 2017-03-09 | ctop | 323 | 73 | vektorlab | "Ctop – commandline monitoring for containers" |
| 2018-04-13 | dust | 3 | 0 | andy_boot | "Show HN: Dust: A more intuitive version of du in rust" |
| 2018-08-05 | lazygit | 551 | 229 | jesseduffield | "Show HN: I made a tool that made me faster at Git" |
| 2018-10-28 | dive | 4 | 0 | kiyanwang | "Wagoodman/dive: A tool for exploring each layer…" |
| 2018-11-25 | dive | 684 | 42 | boyter | "Dive – A tool for exploring each layer in a Docker image" |
| 2019-06-30 | lazydocker | 340 | 47 | jesseduffield | "Lazydocker: a terminal GUI for Docker" (README link) |
| 2020-05-30 | k9s | 154 | 22 | sharjeelsayed | "K9s – Manage Your Kubernetes Clusters in Style" |
| 2021-07-25 | ncdu 2 | 236 | 42 | signa11 | "Ncdu 2: Less hungry and more Ziggy" (author's blog post) |
| 2021-11-30 | lazygit | 395 | 141 | tnorthcutt | "Lazygit: A simple terminal UI for Git commands" |
| 2021-12-01 | lazydocker | 259 | 17 | thunderbong | "Lazydocker: The lazier way to manage everything Docker" |
| 2022-12-07 | ncdu | 446 | 162 | thecosmicfrog | "Ncdu – NCurses Disk Usage" |
| 2023-07-18 | lazydocker | 481 | 73 | mottiden | "Lazydocker: a lazier way to manage everything Docker" |
| 2023-07-19 | k9s | 117 | 19 | emersonrsantos | "K9s: A lazier way to manage Kubernetes Clusters" |
| 2024-01-08 | dive | 464 | 61 | tomas789 | "Dive: A tool for exploring a Docker image, layer contents…" |
| 2024-11-22 | lazydocker | 74 | 9 | Gedxx | "Lazy Docker: The lazier way to manage everything Docker" |
| 2025-11-10 | lazygit | 436 | 222 | linhns | "The lazy Git UI you didn't know you need" (third-party blog) |

No HN story was found for oxker, Doku or dua-cli. gdu's largest story had 1 point (2021-12-22 and 2023-09-28).

---

## lazygit (Jesse Duffield)

**Facts:**
- Show HN by the author on 2018-08-05: 551 points, 229 comments. The title was benefit-framed: "I made a tool that made me faster at Git".
- In "Lazygit Turns 5" (https://jesseduffield.com/Lazygit-5-Years-On/, HN 2023-08-05, 211 points), the author says:
  - his attempts to promote it on Reddit (r/webdev) and a Facebook group "fell on deaf ears";
  - he posted to HN expecting nothing;
  - a flamewar about git UIs versus the CLI "likely boost[ed] the debut post to the frontpage" (his own guess).
- The same post says he kept all documentation in the repo, so users are "always one click away from starring", and that the in-app startup popup suggests starring the repo.
- He released once the MVP was usable, and early contributors then "accelerated" development.
- He had 37k stars at the time of writing. An Appwrite sponsorship funded him for a year.

**Inference:**
- The first audience came from one HN post, not from Reddit.
- A title that states a personal benefit invited debate.
- Later traffic came from *other people's* posts years later (2021, 2023, 2025).

## lazydocker (Jesse Duffield)

**Facts:**
- The repo was created on 2019-05-18. The author's own HN post on 2019-06-30 (340 points, 47 comments) linked the README directly.
- Three releases (v0.2, v0.2.1, v0.2.2) are dated the same day as the post.
- Third parties reposted it in 2021 (259 points), 2023 (481) and 2024 (74).
- The README:
  - opens with sponsor logos and a screenshot;
  - has an "Elevator Pitch" that starts "Minor rant incoming: Something's not working? Maybe a service is down…";
  - offers 9+ install methods (Homebrew, Scoop, Chocolatey, asdf, binary, Go, AUR, Docker, manual).
- It is listed in awesome-docker (Terminal), awesome-tuis, awesome-cli-apps and awesome-cli-apps-in-a-csv. Tecmint ranked it #1 in "11 Must-Have Docker Tools" (2024-11-13).

**Inference:**
- It launched to an audience the author already had from lazygit, about 11 months earlier, with the same "lazy" branding.
- Its repeated front-page reposts show that a tool with a clear pain-point pitch keeps being rediscovered.

## dive (Alex Goodman)

**Facts:**
- The repo was created on 2018-05-13, and the first GitHub release is v0.0.1 (2018-10-17).
- An early HN post by a third party (2018-10-28) got 4 points. Another third party's post on 2018-11-25 got 684 points, 42 comments, about 5–6 weeks after the first release.
- A 2024-01-08 repost by a third party got 464 points.
- Console.dev reviewed it on 2023-10-19 [S] (https://console.dev/tools/dive).
- Docker's blog published "Reduce Your Image Size with the Dive-In Docker Extension" on 2023-08-02. It is an HN entry linking docker.com, about an extension built on dive's idea. The relationship is beyond what I verified.
- The README puts a demo GIF (`.data/demo.gif`) right after the badges, and shows a CI-mode screenshot.
- It shipped 4 releases in its first 6 days (v0.0.1–v0.0.4, 2018-10-17 to 10-22).

**Inference:**
- The author did not need to launch it himself. A visual demo of a concrete problem (wasted image space) made it shareable.
- Fast early releases suggest the author responded quickly to first users.

## ctop (Bradley Cicenas)

**Facts:**
- The repo was created on 2016-12-27. Release v0.3 is dated 2017-03-07 and v0.4 2017-03-09.
- An HN post on 2017-03-09 by the account `vektorlab` got 323 points and 73 comments. The project was originally published under the vektorlab organization (inference from the account name).
- The README has a logo and then `grid.gif` near the top, with a Homebrew version badge.
- It has 16 releases. The last is from 2022-03-22.

**Inference:**
- It launched within days of a usable release, with a GIF.
- Despite no releases since 2022, it still holds 17.8k stars and is listed in awesome-docker, awesome-tuis and Terminal Trove's Docker category [S]. List placements keep paying off long after activity stops.

## oxker (Jack Wills)

**Facts:**
- The repo was created on 2022-04-25, with v0.0.1 the same day.
- It has 53 releases; the latest is v0.13.4 (2026-08-22).
- No HN story was found.
- selfh.st's 2023-10-13 newsletter was "a spotlight on Oxker, a text-based user interface for container management" (https://selfh.st/weekly/2023-10-13/).
- It is listed in awesome-docker (Terminal) and awesome-tuis. Terminal Trove has a page for it [S].
- It is distributed via crates.io, Docker Hub/ghcr, the AUR and MacPorts [S].
- The README shows an animated `.webp` demo near the top.

**Inference:** this is the closest growth pattern to dockdu's realistic path:
- no viral launch;
- steady releases;
- newsletter spotlights;
- awesome-list entries;
- broad packaging.

That combination reached about 1.9k stars over 4.5 years.

## k9s (Fernand Galiana)

**Facts:**
- The repo was created on 2019-01-25 and first released (0.1.0) on 2019-02-02.
- It has 258 GitHub releases.
- Early HN posts by third parties in 2019 got 2–14 points. A 2020-05-30 post linking its own site, k9scli.io, got 154 points.
- The README opens with a logo image, and a section says "K9s will always remain OSS and therefore free!"

**Inference:** early HN attempts did nothing. Growth came from very frequent releases and word of mouth in the Kubernetes community over a year or more.

## Doku (amerkurev): the closest competitor

**Facts:**
- The repo was created on 2022-06-13. GitHub releases run from v0.0.14 (2022-07-23) to v1.3.0 (2025-12-26), 7 in total.
- No HN story was found.
- It was covered by blogs, not launches:
  - noted.lol, 2023-02-09 (https://noted.lol/doku-docker-disk-usage-dashboard/);
  - Medevel, 2025-07-31 (https://medevel.com/doku-docker-disk-monitor/);
  - mariushosting Synology install guides [S];
  - an Unraid Community Apps listing, which a search snippet says has about 1.7M downloads [S, unverified].
- It is listed in awesome-docker (Observability) and awesome-homelab (Monitoring).
- The README has build, Coveralls and Docker-pulls badges and an SVG screenshot.

**Inference:**
- Homelab install-guide authors (Synology/Unraid) carried a container-first tool further than any launch post did.
- Its star count (451) understates its use, given the download figures.
- dockdu's ghcr image makes the same path possible, but a one-shot CLI fits Unraid/Synology "apps" less well than a dashboard.

## ncdu, dust, gdu, dua (disk analyzers)

**Facts:**
- ncdu is hosted outside GitHub. HN stories:
  - 2022-12-07: 446 points, 162 comments;
  - 2021-07-25 (author's "Ncdu 2" blog post, submitted by a third party): 236 points.
- dust's own Show HN (2018-04-13) got 3 points. It now has 12.5k stars and is listed in modern-unix.
- gdu (6.1k stars) and dua-cli (6.3k stars) had no meaningful HN stories.
- dua, gdu and ncdu are all in awesome-cli-apps-in-a-csv. dua-cli and gdu are in awesome-tuis.

**Inference:**
- Tools that do one familiar job better grow through lists ("modern unix replacements"), package managers and word of mouth, without a launch moment.
- An author-written blog post about the internals (ncdu 2's Zig rewrite) got more attention than release announcements.

## Patterns: verified vs inferred

**Verified across the tools above:**
1. **A demo above the fold.** dive (GIF), ctop (GIF), oxker (webp) and lazydocker (screenshot) all show the tool in the first screen of the README.
2. **The biggest HN spikes were often third-party posts.** Examples: dive 684, ncdu 446, lazydocker 481 (2023) and lazygit 436 (2025). The author-posted exceptions were lazygit's debut (551), lazydocker's debut (340) and ctop's (323, org account).
3. **Most of these tools are on several awesome lists at once.** lazydocker, ctop, dive and oxker each appear in 2–4 of the lists checked.
4. **Newsletters covered small tools without a launch.** Examples: selfh.st spotlighted oxker, and Console.dev reviewed dive [S].
5. **Release cadence varies widely.** k9s has 258 releases, oxker 53 and lazydocker 47, while ctop has 16 and none since 2022. ctop still holds 17.8k stars.

**Inferred (not proven by the data):**
- A benefit-framed or rant-framed pitch ("made me faster", "Minor rant incoming") outperforms a feature list.
- A strong first HN post seeds later third-party reposts.
- Homelab install guides matter for container-delivered tools.

## Tactics transferable to dockdu (realistic at 0 stars, solo maintainer)

1. **One Show HN, posted by Maximilian on the day of a substantive release, with a problem-framed title.** Precedents: lazygit's "I made a tool that made me faster at Git" (551 points) and lazydocker's post on its v0.2 release day. For dockdu, the measured story fits that pattern: df 313.6 MiB vs 537.0 MiB of logs. Write the first comment yourself, and do not repost a minor version.
2. **Keep the demo in the first screen, and keep a static image too.** dockdu already has `docs/demo.gif` and `demo.png`. Terminal Trove requires "an image preview (PNG, GIF or MP4)" [S], so reuse the same assets.
3. **Pursue list placements where the peers already are, not where gates block you.** awesome-docker (PR #1578 is open; Doku is in the same Observability section) and awesome-cli-apps-in-a-csv (ctop, dive and lazydocker are there) are open now. Calendar the rest: awesome-go about 2027-03-13, awesome-devops at 100 stars + 6 months, awesome-sysadmin not before 12 months and only via a third party.
4. **Aim for a newsletter spotlight, as oxker got from selfh.st.** Submit once to selfh.st, Console.dev (only while < 1.0) and Go Weekly, each with a different technical hook: homelab compose projects, a beta dev tool, and a Go read-only client respectively.
5. **Make the technical article the shareable unit, not the repo.** Third-party HN spikes went to concrete, visual or internals-focused content: dive's layers, ncdu 2's rewrite post. dockdu's "prune frees ~0 on overlay2 vs about half on containerd" measurement is that kind of content. Publish it under your own name, with AI disclosure where required.
6. **Package broadly before chasing traffic.** oxker and lazydocker both offer many install paths. Add the AUR and nixpkgs now, and Homebrew core once thresholds allow: 75 stars from others, or 225 self-submitted, and a repo at least 30 days old.
7. **Ship steady, meaningful releases with plain changelogs.** oxker (53 releases) and k9s (258) stayed visible through release feeds. selfh.st reportedly compiles its release lists from GitHub data [S]. Avoid version bumps purely for visibility.
8. **Serve the container path for homelab guide writers.** Doku spread through Synology/Unraid guides [S]. A short "run it once on Synology/Unraid" doc section (a read-only `docker run`) gives such authors something to cite. This is an inference, so test the demand before investing in it.
9. **Answer the evergreen questions, not just launch threads.** The Docker forums thread t/137291 has 44,747 views [V-repo]. Helpful, human-written answers with manual commands first, and dockdu as optional, compound over time. ctop and Doku kept getting found long after their peaks.
10. **Keep the honest comparison table and ask competitors to correct it.** This is an inference: there is no hard evidence that it drives stars. Credibility matters with skeptical forum regulars and editors (see the selfh.st "AI disclaimer" remark). Linking to Doku, buildx du and lazydocker where they fit better is already in the README; keep it that way.
