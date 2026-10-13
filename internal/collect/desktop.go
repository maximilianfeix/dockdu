package collect

import "github.com/maximilianfeix/dockdu/internal/report"

// desktopExplanation is printed verbatim; it paraphrases the Docker Desktop
// FAQ on the VM disk image.
const desktopExplanation = "Space inside the VM disk file is freed when images, containers, volumes or build cache are removed. The file itself only shrinks when Docker Desktop compacts it; lowering the disk usage limit under Settings → Resources → Advanced deletes the file and everything in it."

// desktopCandidate is one VM disk image location to probe.
type desktopCandidate struct {
	product string
	path    string
}

// buildDesktop fills the section from measured sizes.
func buildDesktop(c desktopCandidate, apparent, allocated, dockerReported int64) *report.Desktop {
	return &report.Desktop{
		Product: c.product, DiskFile: c.path, ApparentBytes: apparent, AllocatedBytes: allocated,
		DockerReportedBytes: dockerReported, Explanation: desktopExplanation,
	}
}
