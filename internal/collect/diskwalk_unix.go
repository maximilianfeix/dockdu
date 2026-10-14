//go:build linux || darwin

package collect

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	"github.com/maximilianfeix/dockdu/internal/report"
)

// rootFilesItem collects regular files that sit directly in the data root.
const rootFilesItem = "(files)"

// walkAllocated sums allocated blocks (st_blocks × 512) under root, the way
// `du -sx` does: lstat only, so symbolic links are never followed; an
// inode with several hard links is counted once; directories on another
// device (overlay mounts of running containers, separately mounted
// volumes) are not entered. File contents are never opened. The result is
// grouped by the top-level entry of root.
func walkAllocated(ctx context.Context, root string, budget DiskBudget) (*report.DiskUsage, error) {
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	rootStat, ok := rootInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, os.ErrInvalid
	}
	dev := uint64(rootStat.Dev)
	w := &walker{ctx: ctx, dev: dev, budget: budget, seen: map[[2]uint64]struct{}{},
		du: &report.DiskUsage{Complete: true, Items: []report.DiskItem{}}}
	w.du.AllocatedBytes += w.blocks(rootStat)
	w.du.Entries++

	top, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	items := map[string]*report.DiskItem{}
	item := func(name string) *report.DiskItem {
		if it, ok := items[name]; ok {
			return it
		}
		it := &report.DiskItem{Name: name}
		items[name] = it
		return it
	}
	for _, e := range top {
		if w.atDirectory() {
			break
		}
		path := filepath.Join(root, e.Name())
		fi, err := os.Lstat(path)
		if err != nil {
			continue // removed while walking
		}
		st, ok := fi.Sys().(*syscall.Stat_t)
		if !ok {
			continue
		}
		if fi.IsDir() {
			if uint64(st.Dev) != dev {
				w.du.SkippedMounts++
				continue
			}
			w.dir(path, st, item(e.Name()))
			continue
		}
		w.count(st, item(rootFilesItem))
	}
	for _, it := range items {
		w.du.Items = append(w.du.Items, *it)
	}
	sortDiskItems(w.du.Items)
	return w.du, nil
}

type walker struct {
	ctx    context.Context
	dev    uint64
	budget DiskBudget
	seen   map[[2]uint64]struct{}
	du     *report.DiskUsage
	done   bool
}

// stopped checks the entry budget and, every 4096 entries, the deadline.
// atDirectory checks the deadline unconditionally; it is called once per
// directory, where a ReadDir costs far more than the check. Once stopped,
// the result is marked incomplete.
func (w *walker) stopped() bool { return w.check(w.du.Entries%4096 == 0) }

func (w *walker) atDirectory() bool { return w.check(true) }

func (w *walker) check(deadline bool) bool {
	if w.done {
		return true
	}
	if (w.budget.MaxEntries > 0 && w.du.Entries >= w.budget.MaxEntries) || (deadline && w.ctx.Err() != nil) {
		w.done = true
		w.du.Complete = false
	}
	return w.done
}

// blocks returns the allocated bytes of an inode, or 0 when an inode with
// several links was already counted.
func (w *walker) blocks(st *syscall.Stat_t) int64 {
	if uint64(st.Nlink) > 1 && st.Mode&syscall.S_IFMT != syscall.S_IFDIR {
		key := [2]uint64{uint64(st.Dev), st.Ino}
		if _, dup := w.seen[key]; dup {
			return 0
		}
		w.seen[key] = struct{}{}
	}
	return st.Blocks * 512
}

func (w *walker) count(st *syscall.Stat_t, it *report.DiskItem) {
	b := w.blocks(st)
	w.du.AllocatedBytes += b
	w.du.Entries++
	it.AllocatedBytes += b
	it.Entries++
}

// dir walks one directory tree iteratively (no recursion depth limit to
// hit on deep trees).
func (w *walker) dir(path string, st *syscall.Stat_t, it *report.DiskItem) {
	w.count(st, it)
	stack := []string{path}
	for len(stack) > 0 && !w.atDirectory() {
		d := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		entries, err := os.ReadDir(d)
		if err != nil {
			w.du.UnreadableDirs++
			w.du.Complete = false
			continue
		}
		for _, e := range entries {
			if w.stopped() {
				return
			}
			p := filepath.Join(d, e.Name())
			fi, err := os.Lstat(p)
			if err != nil {
				continue // removed while walking
			}
			cst, ok := fi.Sys().(*syscall.Stat_t)
			if !ok {
				continue
			}
			if fi.IsDir() {
				if uint64(cst.Dev) != w.dev {
					w.du.SkippedMounts++
					continue
				}
				w.count(cst, it)
				stack = append(stack, p)
				continue
			}
			w.count(cst, it)
		}
	}
}
