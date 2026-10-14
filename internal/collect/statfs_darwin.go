//go:build darwin

package collect

import "syscall"

// statFS reads the filesystem holding path; on macOS block counts are in
// units of Bsize.
func statFS(path string) (fsInfo, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return fsInfo{}, err
	}
	unit := int64(st.Bsize)
	return fsInfo{
		size:  int64(st.Blocks) * unit,
		used:  int64(st.Blocks-st.Bfree) * unit,
		avail: int64(st.Bavail) * unit,
	}, nil
}
