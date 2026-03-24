//go:build !windows

package collector

import "syscall"

func getDiskUsagePct(path string) float64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0
	}
	bsize := uint64(stat.Bsize)
	total := uint64(stat.Blocks) * bsize
	free := uint64(stat.Bfree) * bsize
	if total == 0 {
		return 0
	}
	used := total - free
	return round2(float64(used) / float64(total) * 100)
}
