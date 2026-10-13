package collect

import (
	"sort"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// Images applies SPEC §6.7. Count/total/reclaimable come from the daemon's
// aggregate; dangling and unique-size rules from the records.
func Images(s *engine.Snapshot) (report.Images, []string) {
	out := report.Images{
		Count:            len(s.Images),
		TotalBytes:       s.Totals.Images.TotalSize,
		ReclaimableBytes: s.Totals.Images.Reclaimable,
		Items:            make([]report.ImageItem, 0, len(s.Images)),
		Suggestions:      []string{},
	}
	if s.Totals.Images.TotalCount > 0 {
		out.Count = int(s.Totals.Images.TotalCount)
	}
	unusedTagged := 0
	for _, im := range s.Images {
		it := report.ImageItem{
			ID:          im.ID,
			Tags:        append([]string{}, im.RepoTags...),
			Bytes:       im.Size,
			SharedBytes: im.SharedSize,
			Containers:  im.Containers,
			Dangling:    isDangling(im),
			UniqueBytes: uniqueBytes(im),
		}
		if it.Dangling {
			out.Dangling++
			if it.UniqueBytes >= 0 {
				out.DanglingBytes += it.UniqueBytes
			}
		} else if im.Containers == 0 {
			unusedTagged++
		}
		out.Items = append(out.Items, it)
	}
	sort.SliceStable(out.Items, func(i, j int) bool {
		if out.Items[i].Bytes != out.Items[j].Bytes {
			return out.Items[i].Bytes > out.Items[j].Bytes
		}
		return out.Items[i].ID < out.Items[j].ID
	})
	if out.Dangling > 0 {
		out.Suggestions = append(out.Suggestions, report.SuggestImagePruneDangling)
	}
	if unusedTagged > 0 {
		out.Suggestions = append(out.Suggestions, report.SuggestImagePruneAll)
	}
	return out, nil
}

// isDangling: no tags, or the literal <none>:<none> placeholder.
func isDangling(im engine.Image) bool {
	if len(im.RepoTags) == 0 {
		return true
	}
	for _, t := range im.RepoTags {
		if t != "<none>:<none>" {
			return false
		}
	}
	return true
}

// uniqueBytes is Size-SharedSize only when no container uses the image and
// both values are known; otherwise -1 (unknown), never 0.
func uniqueBytes(im engine.Image) int64 {
	if im.Containers != 0 || im.SharedSize < 0 || im.Size < 0 {
		return -1
	}
	u := im.Size - im.SharedSize
	if u < 0 {
		return -1
	}
	return u
}
