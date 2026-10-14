//go:build linux

package collect

import "syscall"

// statFS reads the filesystem holding path. Block counts are in units of
// the fragment size (Frsize); older kernels leave it 0 and use Bsize.
func statFS(path string) (fsInfo, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return fsInfo{}, err
	}
	unit := st.Frsize
	if unit <= 0 {
		unit = st.Bsize
	}
	return fsInfo{
		size:  int64(st.Blocks) * unit,
		used:  int64(st.Blocks-st.Bfree) * unit,
		avail: int64(st.Bavail) * unit,
	}, nil
}
