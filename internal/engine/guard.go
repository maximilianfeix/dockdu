package engine

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

// ErrNotReadOnly is returned for any request the read-only guard refuses.
// Seeing it means a code path tried to reach a Docker endpoint outside the
// allowlist; that is a bug, never a condition to work around.
var ErrNotReadOnly = errors.New("dockdu read-only guard: request refused")

// readOnlyEndpoints is every Docker Engine API path dockdu may call, after
// the optional /v1.xx version prefix. The client's version ping uses
// /_ping; everything else maps to Info, DiskUsage and ContainerInspect.
var readOnlyEndpoints = regexp.MustCompile(`^(/v[0-9]+\.[0-9]+)?(/_ping|/info|/system/df|/containers/[0-9a-f]{12,64}/json)$`)

// readOnlyGuard is installed as a request hook on every client dockdu
// creates. It runs before a request leaves the process and refuses any
// method other than GET and HEAD and any path outside readOnlyEndpoints,
// so a mutating call cannot reach the daemon even if one is added by
// mistake. A read-only socket mount does not give this guarantee: `:ro`
// does not stop POST or DELETE requests over the socket.
func readOnlyGuard(req *http.Request) error {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return fmt.Errorf("%w: method %s %s", ErrNotReadOnly, req.Method, req.URL.Path)
	}
	if !readOnlyEndpoints.MatchString(req.URL.Path) {
		return fmt.Errorf("%w: path %s", ErrNotReadOnly, req.URL.Path)
	}
	return nil
}
