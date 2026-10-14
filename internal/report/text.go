package report

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// safe replaces control characters in daemon-sourced strings so a crafted
// label or name cannot inject terminal escape sequences into the report.
func safe(s string) string {
	return strings.Map(func(r rune) rune {
		if unsafeRune(r) {
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
	// para writes prose that fits as is, or word-wrapped at textWidth with
	// first before the first line and cont before the others.
	para := func(first, cont, text string) {
		if runeLen(first)+runeLen(text) <= textWidth {
			p("%s%s\n", first, text)
			return
		}
		for i, l := range wrapTokens(words(text), " ", textWidth-runeLen(first), textWidth-runeLen(cont)) {
			if i == 0 {
				p("%s%s\n", first, l)
			} else {
				p("%s%s\n", cont, l)
			}
		}
	}
	// list writes " · "-separated items, breaking only between items.
	list := func(first, cont string, items ...string) {
		for i, l := range wrapTokens(items, " · ", textWidth-runeLen(first), textWidth-runeLen(cont)) {
			if i == 0 {
				p("%s%s\n", first, l)
			} else {
				p("%s%s\n", cont, l)
			}
		}
	}
	// row writes a summary line: label, value, then the note on the same
	// line when it fits and on indented lines below when it does not.
	row := func(label, value string, note ...string) {
		head := fmt.Sprintf("  %-28s%-11s", label, value)
		joined := strings.Join(note, " · ")
		if joined == "" {
			p("%s\n", strings.TrimRight(head, " "))
			return
		}
		if runeLen(head)+1+runeLen(joined) <= textWidth {
			p("%s %s\n", head, joined)
			return
		}
		p("%s\n", strings.TrimRight(head, " "))
		list("    ", "    ", note...)
	}
	// header writes "\nTITLE  (a · b · c)", or the title alone with the
	// details on the next line when the whole does not fit.
	header := func(title string, info ...string) {
		p("\n")
		joined := strings.Join(info, " · ")
		if runeLen(title)+4+runeLen(joined) <= textWidth {
			p("%s  (%s)\n", title, joined)
			return
		}
		p("%s\n", title)
		list("  ", "  ", info...)
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
	list("", "  ", "dockdu "+r.Version, "docker "+safe(r.Daemon.ServerVersion)+" (API "+safe(r.Daemon.APIVersion)+")",
		safe(r.Daemon.OperatingSystem), "data root "+root)
	p("\n")

	s := r.Summary
	p("SUMMARY\n")
	row("docker system df reports", Bytes(s.DfReportedBytes), "images "+Bytes(s.DfImages), "containers "+Bytes(s.DfContainers),
		"volumes "+Bytes(s.DfVolumes), "build cache "+Bytes(s.DfBuildCache))
	logNote := []string{"not shown by df"}
	if !s.LogBytesComplete {
		logNote = append(logNote, "incomplete, lower bound (see CONTAINER LOGS)")
	}
	row("container logs on disk", Bytes(s.LogBytes), logNote...)
	var partial []string
	if !s.LogBytesComplete {
		partial = append(partial, "logs incomplete")
	}
	if r.BuildCache.DfRowMismatch {
		partial = append(partial, "build cache row disagrees with its records")
	}
	if r.BuildCache.SharedMayBeFreedByPrune && s.BuildCacheSharedBytes > 0 {
		partial = append(partial, "shared build cache may keep its own copy (containerd store)")
	}
	if s.VolumeSizesUnknown > 0 {
		partial = append(partial, fmt.Sprintf("%d volume size(s) unknown", s.VolumeSizesUnknown))
	}
	if r.BuildCache.DfRowExcludesShared {
		// Older daemons: df's build cache row already leaves shared records out.
		row("Docker resources in total", Bytes(s.DockerBytesTotal), "= df + logs")
		para("    ", "    ", fmt.Sprintf("(df's build cache row already leaves out %s shared with images)", Bytes(r.BuildCache.SharedBytes)))
	} else {
		row("Docker resources in total", Bytes(s.DockerBytesTotal), fmt.Sprintf("= df − %s shared build cache + logs", Bytes(s.BuildCacheSharedBytes)))
	}
	if len(partial) > 0 {
		para("    ", "    ", "lower bound: "+strings.Join(partial, "; "))
	}
	if f := r.Filesystem; f != nil {
		row("data root filesystem", Bytes(f.UsedBytes), "used of "+Bytes(f.SizeBytes), Bytes(f.AvailableBytes)+" free")
		p("    allocated blocks, a different measure from Docker's sizes\n")
	}
	sug(s.Suggestions)

	if f := r.Filesystem; f != nil && f.DataRoot != nil {
		du := f.DataRoot
		status := "complete"
		if !du.Complete {
			status = "incomplete, lower bound (see WARNINGS)"
		}
		n, label := limit(len(du.Items))
		header("DATA ROOT ON DISK", safe(f.Path), "allocated blocks", fmt.Sprintf("%d entries", du.Entries), status)
		if n > 0 {
			tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "  ENTRY (%s)\tON DISK\tENTRIES\n", label)
			for _, it := range du.Items[:n] {
				fmt.Fprintf(tw, "  %s\t%s\t%d\n", safe(it.Name), Bytes(it.AllocatedBytes), it.Entries)
			}
			tw.Flush()
		}
		p("  data root on disk           %s\n", Bytes(du.AllocatedBytes))
		if du.OutsideDataRootBytes >= 0 {
			row("outside the data root", Bytes(du.OutsideDataRootBytes), "on the same filesystem, not under "+safe(f.Path), "approximate: includes filesystem metadata")
		} else {
			p("  outside the data root       unknown (the walk was incomplete)\n")
		}
		if du.SkippedMounts > 0 {
			para("  ", "  ", fmt.Sprintf("%d mount(s) inside the data root not entered (overlay mounts of running containers, separately mounted volumes)", du.SkippedMounts))
		}
		para("  ", "  ", "Docker's sizes above are logical sizes; these are blocks on disk. Compare the two side by side: if overlay2 or containerd is far larger than images + containers + build cache, the data root holds layers or content Docker does not list.")
	}

	bc := r.BuildCache
	p("\nBUILD CACHE\n")
	list("  ", "  ", fmt.Sprintf("%d records", bc.Records), "total "+Bytes(bc.TotalBytes), "shared "+Bytes(bc.SharedBytes),
		"private "+Bytes(bc.PrivateBytes), "reclaimable (as df counts it) "+Bytes(bc.ReclaimableBytes))
	if bc.SharedMayBeFreedByPrune {
		para("  ", "  ", fmt.Sprintf("prune frees an unknown part of the %s shared with images: on the containerd image store a shared record can keep its own copy, and Docker does not report how much (on overlay2 none of it is freed). In use, never freed: %s", Bytes(bc.SharedBytes), Bytes(bc.InUseBytes)))
	} else {
		para("  ", "  ", fmt.Sprintf("prune would NOT free %s (shared with images or in use)", Bytes(bc.NotFreedByPrune)))
	}
	sug(bc.Suggestions)

	lg := r.Logs
	n, label := limit(len(lg.Items))
	status := "complete"
	if !lg.Complete {
		status = "incomplete"
	}
	header("CONTAINER LOGS", label, "total "+Bytes(lg.TotalBytes), status)
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
	header("COMPOSE PROJECTS", "total "+Bytes(pr.TotalBytes), "images used by several projects counted once: "+Bytes(pr.SharedImageBytes))
	if pr.ImageOverlapBytes > 0 {
		para("  ", "  ", fmt.Sprintf("image sizes include base layers shared between images: they overlap by at least %s, so IMAGES and TOTAL are not additive", Bytes(pr.ImageOverlapBytes)))
	}
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
	header("IMAGES", label, "total "+Bytes(im.TotalBytes), fmt.Sprintf("dangling %d", im.Dangling), "reclaimable per the Docker API "+Bytes(im.ReclaimableBytes))
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
	header("VOLUMES", label, "total "+Bytes(vo.TotalBytes), fmt.Sprintf("size unknown for %d", vo.UnknownSizes))
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
		list("  ", "  ", "apparent size "+Bytes(d.ApparentBytes)+" (the configured maximum)", "allocated on disk "+Bytes(d.AllocatedBytes))
		para("  ", "  ", fmt.Sprintf("Docker reports %s of resources; a different measure that may overlap, not a disk-block count", Bytes(d.DockerReportedBytes)))
		para("  ", "  ", d.Explanation)
	}

	if len(r.Warnings) > 0 {
		p("\nWARNINGS\n")
		for _, wn := range r.Warnings {
			para("  - ", "    ", safe(wn))
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
