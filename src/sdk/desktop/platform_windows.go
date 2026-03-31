// +build windows

package desktop

import (
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpace = kernel32.NewProc("GetDiskFreeSpaceExW")
)

func getDiskUsage(path string) (total, free, used uint64, err error) {
	pathPtr, _ := syscall.UTF16PtrFromString(path)

	var freeBytes uint64
	var totalBytes uint64
	var freeAvailable uint64

	ret, _, callErr := getDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&freeBytes)),
	)

	if ret == 0 {
		err = callErr
		return
	}

	total = totalBytes
	free = freeBytes
	used = total - free
	return
}