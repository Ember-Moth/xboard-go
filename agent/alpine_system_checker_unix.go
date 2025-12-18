//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"syscall"
)

// getDiskSpace 获取磁盘空间（Unix 实现）
func (asc *AlpineSystemChecker) getDiskSpace(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("无法获取磁盘空间信息: %w", err)
	}
	
	// 计算可用空间（字节）
	free := int64(stat.Bavail) * int64(stat.Bsize)
	return free, nil
}