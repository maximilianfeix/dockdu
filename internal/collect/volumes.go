package collect

import (
	"sort"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// anonymousVolumeLabel marks volumes Docker created without a name (moby
// daemon/volume/service/service.go AnonymousLabel).
const anonymousVolumeLabel = "com.docker.volume.anonymous"

// volumeReference classifies a volume as referenced, unreferenced or
// unknown. A missing usage block or a RefCount of -1 is unknown, never
// unreferenced: dockdu must not call something an orphan on a guess.
func volumeReference(v engine.Volume) string {
	switch {
	case !v.HasUsageData || v.RefCount < 0:
		return "unknown"
	case v.RefCount == 0:
		return "unreferenced"
	default:
		return "referenced"
	}
}

// Volumes applies SPEC §6.8. Orphan = unreferenced and not created by
// compose. Suggestions name each orphan individually; there is no bulk form.
func Volumes(s *engine.Snapshot) (report.Volumes, []string) {
	out := report.Volumes{Items: make([]report.VolumeItem, 0, len(s.Volumes)), Suggestions: []string{}}
	seen := make(map[string]bool, len(s.Volumes))
	for _, v := range s.Volumes {
		if seen[v.Name] {
			continue
		}
		seen[v.Name] = true
		out.Count++
		it := report.VolumeItem{
			Name: v.Name, Driver: v.Driver, Project: v.Labels[composeProjectLabel],
			Bytes: -1, RefCount: v.RefCount, Reference: volumeReference(v),
			BuildxNode: buildxNode(v.Name),
		}
		_, it.Anonymous = v.Labels[anonymousVolumeLabel]
		if v.HasUsageData && v.Size >= 0 {
			it.Bytes = v.Size
			out.TotalBytes += v.Size
		} else {
			out.UnknownSizes++
		}
		it.Orphan = it.Reference == "unreferenced" && it.Project == ""
		out.Items = append(out.Items, it)
	}
	sort.SliceStable(out.Items, func(i, j int) bool {
		a, b := out.Items[i], out.Items[j]
		if a.Bytes != b.Bytes {
			return a.Bytes > b.Bytes
		}
		return a.Name < b.Name
	})
	for _, it := range out.Items {
		if it.Orphan {
			out.Suggestions = append(out.Suggestions, report.SuggestVolumeRm(it.Name))
		}
	}
	return out, nil
}
