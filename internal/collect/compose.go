package collect

import (
	"sort"
	"time"

	"github.com/maximilianfeix/dockdu/internal/engine"
	"github.com/maximilianfeix/dockdu/internal/report"
)

// noneProject is the bucket for everything without a compose project label.
const noneProject = "(none)"

// stoppedOldAfter is the age after which a non-running container counts as
// "stopped, old" in its project row.
const stoppedOldAfter = 7 * 24 * time.Hour

// projectOf returns the compose project of a labelled object or noneProject.
func projectOf(labels map[string]string) string {
	if p := labels[composeProjectLabel]; p != "" {
		return p
	}
	return noneProject
}

// imageProjects maps every image ID to the sorted set of projects whose
// containers reference it. Images no container uses are absent. A container
// without ImageID falls back to matching its image name against RepoTags.
func imageProjects(s *engine.Snapshot) map[string][]string {
	byTag := map[string]string{}
	for _, im := range s.Images {
		for _, t := range im.RepoTags {
			byTag[t] = im.ID
		}
	}
	set := map[string]map[string]bool{}
	for _, c := range s.Containers {
		id := c.ImageID
		if id == "" {
			id = byTag[c.Image]
		}
		if id == "" {
			continue
		}
		if set[id] == nil {
			set[id] = map[string]bool{}
		}
		set[id][projectOf(c.Labels)] = true
	}
	out := make(map[string][]string, len(set))
	for id, ps := range set {
		names := make([]string, 0, len(ps))
		for p := range ps {
			names = append(names, p)
		}
		sort.Strings(names)
		out[id] = names
	}
	return out
}

// Projects applies SPEC §6.6: containers, images, volumes and measured log
// bytes grouped by compose project. An image used by several projects is
// reported in each project's SharedImageBytes and excluded from TotalBytes,
// so the section total is the plain sum of the project totals.
func Projects(s *engine.Snapshot, logs report.Logs, now time.Time) (report.Projects, []string) {
	rows := map[string]*report.Project{}
	row := func(name string) *report.Project {
		if r, ok := rows[name]; ok {
			return r
		}
		r := &report.Project{Name: name}
		rows[name] = r
		return r
	}

	for _, c := range s.Containers {
		r := row(projectOf(c.Labels))
		r.Containers++
		if c.State != "running" && c.Created > 0 && now.Sub(time.Unix(c.Created, 0)) > stoppedOldAfter {
			r.StoppedOld++
		}
	}

	owners := imageProjects(s)
	var sharedOnce int64
	seenImage := make(map[string]bool, len(s.Images))
	for _, im := range s.Images {
		if im.ID != "" && seenImage[im.ID] {
			continue
		}
		seenImage[im.ID] = true
		size := im.Size
		if size < 0 {
			size = 0 // unknown image size: not added, never negative
		}
		ps, used := owners[im.ID]
		switch {
		case !used && isDangling(im):
			row(noneProject).DanglingImages++
			row(noneProject).ImageBytes += size
		case !used:
			row(noneProject).ImageBytes += size
		case len(ps) == 1:
			row(ps[0]).ImageBytes += size
		default:
			sharedOnce += size
			for _, p := range ps {
				r := row(p)
				r.SharedImageBytes += size
				r.ImagesShared++
			}
		}
	}

	seenVolume := make(map[string]bool, len(s.Volumes))
	for _, v := range s.Volumes {
		if seenVolume[v.Name] {
			continue
		}
		seenVolume[v.Name] = true
		r := row(projectOf(v.Labels))
		if v.HasUsageData && v.Size >= 0 {
			r.VolumeBytes += v.Size
		} else {
			r.VolumesUnknown++
		}
		if volumeReference(v) == "unreferenced" && v.Labels[composeProjectLabel] == "" {
			r.OrphanVolumes++
		}
	}

	for _, it := range logs.Items {
		name := it.Project
		if name == "" {
			name = noneProject
		}
		if !it.Measured {
			if it.Rotation != "not on disk" {
				row(name).LogsUnknown++
			}
			continue
		}
		row(name).LogBytes += it.Bytes
	}

	out := report.Projects{Items: make([]report.Project, 0, len(rows)), SharedImageBytes: sharedOnce, Suggestions: []string{}}
	for _, r := range rows {
		r.TotalBytes = r.ImageBytes + r.VolumeBytes + r.LogBytes
		out.TotalBytes += r.TotalBytes
		out.Items = append(out.Items, *r)
	}
	out.TotalBytes += sharedOnce // shared images counted exactly once at section level
	sort.SliceStable(out.Items, func(i, j int) bool {
		a, b := out.Items[i], out.Items[j]
		if (a.Name == noneProject) != (b.Name == noneProject) {
			return b.Name == noneProject
		}
		if a.TotalBytes != b.TotalBytes {
			return a.TotalBytes > b.TotalBytes
		}
		return a.Name < b.Name
	})
	return out, nil
}

// AttachImageProjects fills ImageItem.Projects from container references.
func AttachImageProjects(imgs *report.Images, s *engine.Snapshot) {
	owners := imageProjects(s)
	for i := range imgs.Items {
		imgs.Items[i].Projects = owners[imgs.Items[i].ID]
	}
}
