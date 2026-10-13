//go:build darwin

package collect

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/maximilianfeix/dockdu/internal/report"
)

// Desktop measures the Docker Desktop or OrbStack VM disk image on macOS.
// It returns nil when neither file exists. Only Stat is used.
func Desktop(home string, dockerReported int64) (*report.Desktop, []string) {
	if home == "" {
		return nil, nil
	}
	cands := []desktopCandidate{
		{"Docker Desktop", filepath.Join(home, "Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw")},
		{"OrbStack", filepath.Join(home, "Library/Group Containers/HUAQ24HBR6.dev.orbstack/data/data.img")},
	}
	var warns []string
	for _, c := range cands {
		fi, err := os.Stat(c.path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			warns = append(warns, fmt.Sprintf("desktop: cannot stat %s: %v", c.path, err))
			continue
		}
		allocated := int64(-1)
		if st, ok := fi.Sys().(*syscall.Stat_t); ok {
			allocated = st.Blocks * 512 // Blocks is in 512-byte units on macOS
		}
		return buildDesktop(c, fi.Size(), allocated, dockerReported), warns
	}
	return nil, warns
}
