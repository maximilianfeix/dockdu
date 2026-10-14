package report

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"unicode"
)

// safe replaces control characters in daemon-sourced strings so a crafted
// label or name cannot inject terminal escape sequences into the report.
func safe(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return '?'
		}
		return r
	}, s)
}

// WriteText renders the human-readable report (SPEC §7). top limits the rows
// of each list; 0 means all. Nothing here is coloured.
func WriteText(w io.Writer, r *Report, top int) error {
	var b strings.Builder
	p := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }
	sug := func(ss []string) {
		for _, s := range ss {
			p("  → %s\n", safe(s)) // defence in depth: names inside are daemon data
		}
	}
	limit := func(n int) (int, string) {
		if top <= 0 || n <= top {
			return n, fmt.Sprintf("%d", n)
		}
		return top, fmt.Sprintf("top %d of %d", top, n)
	}

	root := safe(r.Daemon.DataRoot)
	if r.Daemon.ReportedDataRoot != "" {
		root += " (daemon reports " + safe(r.Daemon.ReportedDataRoot) + ")"
	}
	p("dockdu %s · docker %s (API %s) · %s · data root %s\n\n", r.Version, safe(r.Daemon.ServerVersion), safe(r.Daemon.APIVersion), safe(r.Daemon.OperatingSystem), root)

	s := r.Summary
	p("SUMMARY\n")
	p("  docker system df reports    %-11s images %s · containers %s · volumes %s · build cache %s\n",
		Bytes(s.DfReportedBytes), Bytes(s.DfImages), Bytes(s.DfContainers), Bytes(s.DfVolumes), Bytes(s.DfBuildCache))
	logNote := "not shown by df"
	if !s.LogBytesComplete {
		logNote += " · incomplete, lower bound (see CONTAINER LOGS)"
	}
	p("  container logs on disk      %-11s %s\n", Bytes(s.LogBytes), logNote)
	totalNote := ""
	if !s.LogBytesComplete {
		totalNote = " (lower bound: logs incomplete)"
	}
	if s.BuildCacheSharedBytes == 0 && r.BuildCache.SharedBytes > 0 {
		// Older daemons: df's build cache row already leaves shared records out.
		p("  Docker resources in total   %-11s = df + logs (df's build cache row already leaves out %s shared with images)%s\n", Bytes(s.DockerBytesTotal), Bytes(r.BuildCache.SharedBytes), totalNote)
	} else {
		p("  Docker resources in total   %-11s = df − %s build cache shared with images + logs%s\n", Bytes(s.DockerBytesTotal), Bytes(s.BuildCacheSharedBytes), totalNote)
	}
	sug(s.Suggestions)

	bc := r.BuildCache
	p("\nBUILD CACHE\n")
	p("  %d records · total %s · shared %s · private %s · reclaimable (as df counts it) %s\n",
		bc.Records, Bytes(bc.TotalBytes), Bytes(bc.SharedBytes), Bytes(bc.PrivateBytes), Bytes(bc.ReclaimableBytes))
	p("  prune would NOT free %s (shared with images or in use)\n", Bytes(bc.NotFreedByPrune))
	sug(bc.Suggestions)

	lg := r.Logs
	n, label := limit(len(lg.Items))
	status := "complete"
	if !lg.Complete {
		status = "incomplete"
	}
	p("\nCONTAINER LOGS  (%s · total %s · %s)\n", label, Bytes(lg.TotalBytes), status)
	if n > 0 {
		tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  CONTAINER\tPROJECT\tDRIVER\tSIZE\tFILES\tROTATION\tNOTE")
		for _, it := range lg.Items[:n] {
			size, files, note := Bytes(it.Bytes), fmt.Sprint(it.Files), safe(it.Note)
			if !it.Measured {
				size, files = "unmeasured", "-"
				if it.Rotation != "not on disk" {
					note = "not measured, see WARNINGS"
				}
			}
			rot := it.Rotation
			if it.Rotation == "bounded" && it.MaxSize != "" {
				ms, mf := it.MaxSize, it.MaxFile
				defaults := strings.Contains(ms, "(driver default)") || strings.Contains(mf, "(driver default)")
				ms = strings.TrimSuffix(ms, " (driver default)")
				mf = strings.TrimSuffix(mf, " (driver default)")
				rot = fmt.Sprintf("bounded %s × %s", ms, mf)
				if defaults {
					rot += " (driver defaults)"
				}
			}
			fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\t%s\t%s\n", safe(it.Name), orDash(safe(it.Project)), safe(it.Driver), size, files, rot, note)
		}
		tw.Flush()
	}
	sug(lg.Suggestions)

	pr := r.Projects
	p("\nCOMPOSE PROJECTS  (total %s · shared images counted once: %s)\n", Bytes(pr.TotalBytes), Bytes(pr.SharedImageBytes))
	if len(pr.Items) > 0 {
		tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  PROJECT\tCONTAINERS\tIMAGES\tSHARED IMAGES\tVOLUMES\tLOGS\tTOTAL\tNOTES")
		for _, it := range pr.Items {
			var notes []string
			if it.OrphanVolumes > 0 {
				notes = append(notes, fmt.Sprintf("%d orphan volume(s)", it.OrphanVolumes))
			}
			if it.DanglingImages > 0 {
				notes = append(notes, fmt.Sprintf("%d dangling image(s)", it.DanglingImages))
			}
			if it.StoppedOld > 0 {
				notes = append(notes, fmt.Sprintf("%d stopped > 7 d", it.StoppedOld))
			}
			if it.VolumesUnknown > 0 {
				notes = append(notes, fmt.Sprintf("%d volume size(s) unknown", it.VolumesUnknown))
			}
			logs := Bytes(it.LogBytes)
			if it.LogsUnknown > 0 {
				logs += fmt.Sprintf(" (%d unmeasured)", it.LogsUnknown)
			}
			shared := "-"
			if it.ImagesShared > 0 {
				shared = fmt.Sprintf("+%s (%d)", Bytes(it.SharedImageBytes), it.ImagesShared)
			}
			fmt.Fprintf(tw, "  %s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n", safe(it.Name), it.Containers, Bytes(it.ImageBytes), shared,
				Bytes(it.VolumeBytes), logs, Bytes(it.TotalBytes), strings.Join(notes, ", "))
		}
		tw.Flush()
	}
	sug(pr.Suggestions)

	im := r.Images
	n, label = limit(len(im.Items))
	p("\nIMAGES  (%s · total %s · dangling %d · reclaimable (as df counts it) %s)\n", label, Bytes(im.TotalBytes), im.Dangling, Bytes(im.ReclaimableBytes))
	if n > 0 {
		tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  IMAGE\tSIZE\tSHARED\tUNIQUE IF REMOVED\tCONTAINERS\tPROJECTS")
		for _, it := range im.Items[:n] {
			name := shortID(it.ID)
			if len(it.Tags) > 0 && !it.Dangling {
				name = safe(it.Tags[0])
				if len(it.Tags) > 1 {
					name += fmt.Sprintf(" (+%d)", len(it.Tags)-1)
				}
			} else if it.Dangling {
				name = "<none> " + shortID(it.ID)
			}
			containers := fmt.Sprint(it.Containers)
			if it.Containers < 0 {
				containers = "unknown"
			}
			fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\t%s\n", name, Bytes(it.Bytes), Bytes(it.SharedBytes), Bytes(it.UniqueBytes), containers, orDash(safe(strings.Join(it.Projects, ","))))
		}
		tw.Flush()
	}
	sug(im.Suggestions)

	vo := r.Volumes
	n, label = limit(len(vo.Items))
	p("\nVOLUMES  (%s · total %s · size unknown for %d)\n", label, Bytes(vo.TotalBytes), vo.UnknownSizes)
	if n > 0 {
		tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "  VOLUME\tPROJECT\tDRIVER\tSIZE\tREFS\tSTATE")
		for _, it := range vo.Items[:n] {
			refs := fmt.Sprint(it.RefCount)
			if it.RefCount < 0 {
				refs = "unknown"
			}
			state := it.Reference
			if it.Orphan {
				state = "orphan (unreferenced, not from compose)"
			}
			fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\t%s\t%s\n", safe(it.Name), orDash(safe(it.Project)), safe(it.Driver), Bytes(it.Bytes), refs, state)
		}
		tw.Flush()
	}
	sug(vo.Suggestions)

	if d := r.Desktop; d != nil {
		p("\n%s\n", strings.ToUpper(d.Product))
		p("  %s\n", d.DiskFile)
		p("  apparent size %s (the configured maximum) · allocated on disk %s\n", Bytes(d.ApparentBytes), Bytes(d.AllocatedBytes))
		p("  Docker reports %s of resources; a different measure that may overlap, not a disk-block count\n", Bytes(d.DockerReportedBytes))
		p("  %s\n", d.Explanation)
	}

	if len(r.Warnings) > 0 {
		p("\nWARNINGS\n")
		for _, wn := range r.Warnings {
			p("  - %s\n", safe(wn))
		}
	}
	p("\ndockdu never deletes anything. Every → line is a suggestion for you to run.\n")
	// tabwriter pads empty trailing cells; strip the padding so the output
	// (and the golden file) has no trailing spaces.
	lines := strings.Split(b.String(), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	_, err := io.WriteString(w, strings.Join(lines, "\n"))
	return err
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func shortID(id string) string {
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}
