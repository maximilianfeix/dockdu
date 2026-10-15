package collect

import (
	"strings"

	"github.com/maximilianfeix/dockdu/internal/report"
)

// buildx's docker-container driver runs BuildKit in a container named
// buildx_buildkit_<node> and keeps its state, the builder's whole build
// cache, in the volume buildx_buildkit_<node>_state (buildx
// driver/driver.go builderNamePrefix, driver/docker-container/driver.go
// volumeStateSuffix). That cache is not in df's build cache row, which
// covers only the daemon's own builder.
const (
	buildxPrefix      = "buildx_buildkit_"
	buildxStateSuffix = "_state"
)

// buildxNode returns the node name of a buildx state volume, or "".
func buildxNode(volume string) string {
	if !strings.HasPrefix(volume, buildxPrefix) || !strings.HasSuffix(volume, buildxStateSuffix) {
		return ""
	}
	if len(volume) <= len(buildxPrefix)+len(buildxStateSuffix) {
		return ""
	}
	return volume[len(buildxPrefix) : len(volume)-len(buildxStateSuffix)]
}

// buildxBuilderName derives the builder from buildx's default node name,
// <builder><index> with index 0 for the first node (store/nodegroup.go).
// Only the first node's name is unambiguous, so anything else is "".
func buildxBuilderName(node string) string {
	if len(node) < 2 || !strings.HasSuffix(node, "0") {
		return ""
	}
	return node[:len(node)-1]
}

// BuildxBuilders lists the buildx state volumes in vols, in the volume
// list's order, with one prune hint per builder whose container exists
// and whose name is known. An unreferenced state volume already has its
// own `docker volume rm` hint in the volume section.
func BuildxBuilders(vols report.Volumes) ([]report.BuildxBuilder, []string) {
	out := []report.BuildxBuilder{}
	var sugg []string
	needLs := false
	for _, v := range vols.Items {
		if v.BuildxNode == "" {
			continue
		}
		b := report.BuildxBuilder{Node: v.BuildxNode, Volume: v.Name, Builder: buildxBuilderName(v.BuildxNode), Bytes: v.Bytes, Reference: v.Reference}
		out = append(out, b)
		switch {
		case b.Reference == "unreferenced":
		case b.Builder != "":
			sugg = append(sugg, report.SuggestBuildxPrune(b.Builder))
		default:
			needLs = true
		}
	}
	if needLs {
		sugg = append(sugg, report.SuggestBuildxLs)
	}
	return out, sugg
}
