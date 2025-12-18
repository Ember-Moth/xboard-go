//go:build windows
// +build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpace = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// getDiskSpace 获取磁盘空间（Windows 实现）
func (asc *AlpineSystemChecker) getDiskSpace(path string) (int64, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("无法转换路径: %w", err)
	}
	
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
	
	ret, _, err := getDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)
	
	if ret == 0 {
		return 0, fmt.Errorf("GetDiskFreeSpaceEx 失败: %w", err)
	}
	
	return int64(freeBytesAvailable), nil
}