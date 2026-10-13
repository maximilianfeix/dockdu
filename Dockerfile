# dockdu ships as a static binary on an empty base: nothing else is in the
# image. GoReleaser builds per platform and copies the matching binary.
FROM scratch
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/dockdu /usr/bin/dockdu
ENTRYPOINT ["/usr/bin/dockdu"]
