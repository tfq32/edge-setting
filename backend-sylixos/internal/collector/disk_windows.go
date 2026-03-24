//go:build windows

package collector

import (
	"syscall"
	"unsafe"
)

func getDiskUsagePct(path string) float64 {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpace := kernel32.NewProc("GetDiskFreeSpaceExW")

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		// Windows 下用 C:\ 作为默认
		pathPtr, _ = syscall.UTF16PtrFromString("C:\\")
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	ret, _, _ := getDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if ret == 0 || totalBytes == 0 {
		return 0
	}
	used := totalBytes - totalFreeBytes
	return round2(float64(used) / float64(totalBytes) * 100)
}
