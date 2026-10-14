# Outlines for posts written by the maintainer

## Technical article (dev.to canonical; Hashnode cross-post with a canonical URL)

- **Title:** "Why `docker system df` doesn't explain your full disk".
- **Sections:**
  1. The symptom, using real numbers from the demo.
  2. Logs: the json-file defaults and how to check by hand with `du`.
  3. Build cache: what Shared means, the Engine 28 vs 29 row difference, `buildx du` vs the daemon.
  4. Docker.raw: apparent vs allocated size.
  5. What no API shows: overlay2 leftovers, containerd content.
  6. One-command version: dockdu, with its limits.
- **Disclosure line** if AI helped with drafting (dev.to guideline).

## Show HN

- **Title:** `Show HN: dockdu – where did your Docker disk space go? (read-only CLI)`.
- **Link:** the repository.
- **First comment, your words, covering:**
  - why you built it;
  - the Engine 28 vs 29 build-cache row difference (concrete and verifiable);
  - read-only enforcement (`:ro` on the socket does not make the API read-only; dockdu's request guard does);
  - limits (no per-layer attribution, default builder only, no Windows).
- **Logistics:** be online for about 3 hours. Answer the expected questions: why not `du`/ncdu, why not `buildx du`, containerd.

## Reddit

- **r/docker:** ask the moderators first (self-promotion reportedly needs approval; the rules could not be verified first-hand). Then a text post: problem first, demo numbers, link at the end.
- **r/selfhosted:** check the age and AI-built-project rules on the live sidebar. Likely the new-project megathread until the project is 3 months old.
- **r/devops:** only if the rules allow; CI-runner angle (logs and build cache on runners, JSON plus exit code 2 in scripts).

## X / LinkedIn

- **X:** 4–5 posts. The symptom → logs → shared build cache (with the 96 B vs 140.7 MiB example) → read-only → link. No hashtags spam, no star requests.
- **LinkedIn:** a single post about the Engine 28 vs 29 accounting difference as a lesson in measuring before cleaning, with the link in the first comment.

## dev.to tutorial

- **Title:** "Find out which compose project fills your Docker disk".
- **Steps:** run `scripts/demo-up.sh` on a test daemon, read each section, use the jq examples from the README, run `scripts/demo-down.sh`.
