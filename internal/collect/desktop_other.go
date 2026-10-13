//go:build !darwin

package collect

import "github.com/maximilianfeix/dockdu/internal/report"

// Desktop is only implemented on macOS; elsewhere there is no VM disk file
// to measure and the section is omitted.
func Desktop(home string, dockerReported int64) (*report.Desktop, []string) {
	return nil, nil
}
