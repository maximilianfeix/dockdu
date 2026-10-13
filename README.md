# dockdu

**Find out where Docker's disk space actually went, and what `docker system prune` would NOT give back.**

`docker system df` says 11 GB. Your disk says 138 GB. dockdu explains the difference, grouped by compose project, including the BuildKit cache and container logs that `df` does not show.

Read-only. One command. dockdu never removes, stops or prunes anything; every suggestion is printed as text for you to run.

Status: v0.1 in progress. See [SPEC.md](SPEC.md) for the scope and [PROGRESS.md](PROGRESS.md) for the current state.
