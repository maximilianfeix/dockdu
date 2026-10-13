package collect

import "strings"

func containsVolumesFlag(s string) bool {
	return strings.Contains(s, "--volumes") || strings.Contains(s, "volume prune")
}
