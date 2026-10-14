# FAQ for launch threads

**Does it delete anything?**
No. It never sends a request other than GET/HEAD to four read endpoints. The client refuses anything else before sending, and tests against a fake Engine API fail if a mutating call is ever reached. The `→` lines are text for you to read.

**The socket is mounted `:ro`, isn't that enough?**
No. `:ro` stops the socket file from being replaced, not POST or DELETE requests over it. Anything with the socket is effectively root on the host. The guarantee comes from dockdu's code, so use a checksummed release or build it from source.

**Why not just `du -sh /var/lib/docker`?**
`du` tells you how big the directory is, not which container's logs, which compose project or which build cache records it belongs to, or what prune would keep. It also double-counts overlay mounts if you don't use `-x`. dockdu uses Docker's own accounting and measures logs with `lstat`.

**Why not `docker buildx du`?**
It is great for builder detail and counts shared records as reclaimable. dockdu puts the shared figure next to `docker system df` and avoids counting those bytes twice against images.

**Does it read my logs or environment variables?**
No. Log files are only `lstat`ed. Container inspect data stays in memory; only the log driver, path and rotation options are used.

**Does it phone home?**
No. The only connection is the Docker endpoint you configured.

**Why are numbers different from `df -h`?**
Docker reports logical sizes with shared layers counted once. Disks fill in allocated blocks, and some of the space may not be Docker's. dockdu shows both side by side and never subtracts one from the other.

**Windows?**
Not supported. Inside a WSL2 distro running Docker Engine natively, the Linux binary should work, but that is untested.
