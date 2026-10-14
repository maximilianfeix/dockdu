# v0.2.0 release verification, 2026-10-14

Tag `v0.2.0` → `d054f47` (annotated, by Maxi). First release run 38612540882 stopped at the preflight: the tap token had no write access (HTTP 403), and nothing was published. After the owner fixed the token, the failed job was re-run from the tag. Result: success; preflight, GoReleaser and image publish all passed.

| Check | Result |
|---|---|
| GitHub release: not a draft, not a prerelease; 4 archives plus `checksums.txt` | PASS |
| `shasum -a 256 -c checksums.txt` for darwin_arm64 and linux_amd64 | PASS |
| darwin_arm64 binary: `dockdu --version` → `dockdu 0.2.0` | PASS |
| `ghcr.io/maximilianfeix/dockdu:v0.2.0` and `:latest`: linux/amd64 and linux/arm64 | PASS |
| Homebrew cask `version "0.2.0"`, all 4 sha256 match `checksums.txt` | PASS |
| `brew upgrade --cask dockdu` 0.1.0 → 0.2.0, no quarantine flag, a run exits 0 | PASS |
| `go install github.com/maximilianfeix/dockdu/cmd/dockdu@v0.2.0` → `dockdu 0.2.0` (v0.1.0 printed `dev`) | PASS |

## Follow-up found during the check

Homebrew 7.0.9 warns `Calling postflight is deprecated! Use postflight_steps instead.` for the generated cask (the quarantine-removal hook). GoReleaser 2.18.3 only offers `hooks.post.install`, which renders as `postflight`. The cask still installs. Options: a `custom_block` with `postflight_steps`, or a GoReleaser release that supports the new stanza. Fix it before Homebrew turns the warning into an error.
