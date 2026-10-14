//go:build !linux && !darwin

package collect

import (
	"context"
	"errors"

	"github.com/maximilianfeix/dockdu/internal/report"
)

func walkAllocated(context.Context, string, DiskBudget) (*report.DiskUsage, error) {
	return nil, errors.New("--disk is only implemented on Linux and macOS")
}

func statFS(string) (fsInfo, error) {
	return fsInfo{}, errors.New("statfs is only implemented on Linux and macOS")
}
