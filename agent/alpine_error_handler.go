package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// AlpineErrorHandler Alpine 错误处理器
type AlpineErrorHandler struct {
	logger        *DebugLogger
	systemChecker *AlpineSystemChecker
}

// NewAlpineErrorHandler 创建新的 Alpine 错误处理器
func NewAlpineErrorHandler(logger *DebugLogger, systemChecker *AlpineSystemChecker) *AlpineErrorHandler {
	return &AlpineErrorHandler{
		logger:        logger,
		systemChecker: systemChecker,
	}
}

// HandleError 处理错误
func (aeh *AlpineErrorHandler) HandleError(err error, context string) *AlpineError {
	if err == nil {
		return nil
	}
	
	aeh.logger.Debug("处理错误: %v (上下文: %s)", err, context)
	
	// 如果已经是 AlpineError，直接返回
	if alpineErr, ok := err.(*AlpineError); ok {
		return alpineErr
	}
	
	// 创建新的 AlpineError
	alpineErr := &AlpineError{
		Message:   err.Error(),
		Timestamp: time.Now(),
	}
	
	// 分析错误类别
	alpineErr.Category = aeh.categorizeError(err, context)
	
	// 生成建议
	alpineErr.Suggestion = aeh.GenerateSuggestion(err, context, alpineErr.Category)
	
	// 捕获系统上下文
	if systemInfo, err := aeh.CaptureSystemContext(); err == nil {
		alpineErr.SystemInfo = systemInfo
	}
	
	// 捕获堆栈跟踪
	alpineErr.StackTrace = aeh.captureStackTrace()
	
	aeh.logger.LogError(alpineErr, context)
	
	return alpineErr
}

// HandlePanic 处理 panic
func (aeh *AlpineErrorHandler) HandlePanic(r interface{}) *AlpineError {
	aeh.logger.Error("捕获到 panic: %v", r)
	
	alpineErr := &AlpineError{
		Category:   ErrorCategorySystem,
		Message:    fmt.Sprintf("程序崩溃: %v", r),
		Timestamp:  time.Now(),
		StackTrace: aeh.captureStackTrace(),
	}
	
	// 生成崩溃建议
	alpineErr.Suggestion = aeh.generatePanicSuggestion(r)
	
	// 捕获系统上下文
	if systemInfo, err := aeh.CaptureSystemContext(); err == nil {
		alpineErr.SystemInfo = systemInfo
	}
	
	return alpineErr
}

// categorizeError 分类错误
func (aeh *AlpineErrorHandler) categorizeError(err error, context string) ErrorCategory {
	errMsg := strings.ToLower(err.Error())
	contextLower := strings.ToLower(context)
	
	// 网络相关错误 - 映射到系统错误，因为我们移除了 ErrorCategoryNetwork
	if strings.Contains(errMsg, "connection") ||
	   strings.Contains(errMsg, "network") ||
	   strings.Contains(errMsg, "dns") ||
	   strings.Contains(errMsg, "timeout") ||
	   strings.Contains(errMsg, "unreachable") ||
	   strings.Contains(contextLower, "network") ||
	   strings.Contains(contextLower, "api") {
		return ErrorCategorySystem
	}
	
	// 权限相关错误
	if strings.Contains(errMsg, "permission") ||
	   strings.Contains(errMsg, "access denied") ||
	   strings.Contains(errMsg, "forbidden") ||
	   strings.Contains(errMsg, "not permitted") {
		return ErrorCategoryPermission
	}
	
	// 依赖项相关错误
	if strings.Contains(errMsg, "not found") ||
	   strings.Contains(errMsg, "no such file") ||
	   strings.Contains(errMsg, "command not found") ||
	   strings.Contains(errMsg, "executable file not found") ||
	   strings.Contains(contextLower, "dependency") ||
	   strings.Contains(contextLower, "sing-box") {
		return ErrorCategoryDependency
	}
	
	// 配置相关错误
	if strings.Contains(errMsg, "config") ||
	   strings.Contains(errMsg, "configuration") ||
	   strings.Contains(errMsg, "invalid") ||
	   strings.Contains(errMsg, "parse") ||
	   strings.Contains(contextLower, "config") {
		return ErrorCategoryConfiguration
	}
	
	// 资源相关错误
	if strings.Contains(errMsg, "memory") ||
	   strings.Contains(errMsg, "disk") ||
	   strings.Contains(errMsg, "space") ||
	   strings.Contains(errMsg, "resource") ||
	   strings.Contains(errMsg, "limit") {
		return ErrorCategoryResource
	}
	
	// 兼容性相关错误
	if strings.Contains(errMsg, "musl") ||
	   strings.Contains(errMsg, "libc") ||
	   strings.Contains(errMsg, "library") ||
	   strings.Contains(errMsg, "symbol") ||
	   strings.Contains(errMsg, "version") {
		return ErrorCategoryCompatibility
	}
	
	// 默认为系统错误
	return ErrorCategorySystem
}

// GenerateSuggestion 生成建议
func (aeh *AlpineErrorHandler) GenerateSuggestion(err error, context string, category ErrorCategory) string {
	// 检查是否是网络相关错误
	errMsg := strings.ToLower(err.Error())
	contextLower := strings.ToLower(context)
	
	if strings.Contains(errMsg, "connection") ||
	   strings.Contains(errMsg, "network") ||
	   strings.Contains(errMsg, "dns") ||
	   strings.Contains(errMsg, "timeout") ||
	   strings.Contains(contextLower, "network") ||
	   strings.Contains(contextLower, "api") {
		return aeh.generateNetworkSuggestion(err, context)
	}
	
	switch category {
	case ErrorCategoryPermission:
		return aeh.generatePermissionSuggestion(err, context)
	case ErrorCategoryDependency:
		return aeh.generateDependencySuggestion(err, context)
	case ErrorCategoryConfiguration:
		return aeh.generateConfigurationSuggestion(err, context)
	case ErrorCategoryResource:
		return aeh.generateResourceSuggestion(err, context)
	case ErrorCategoryCompatibility:
		return aeh.generateCompatibilitySuggestion(err, context)
	case ErrorCategorySystem:
		return aeh.generateSystemSuggestion(err, context)
	default:
		return "请检查系统日志获取更多信息，或联系技术支持。"
	}
}

// generateNetworkSuggestion 生成网络相关建议
func (aeh *AlpineErrorHandler) generateNetworkSuggestion(err error, context string) string {
	suggestions := []string{
		"1. 检查网络连接是否正常",
		"2. 验证 DNS 配置: cat /etc/resolv.conf",
		"3. 测试网络连通性: ping 8.8.8.8",
		"4. 检查防火墙设置",
		"5. 如果在容器中，检查容器网络配置",
	}
	
	errMsg := strings.ToLower(err.Error())
	
	if strings.Contains(errMsg, "dns") {
		suggestions = append(suggestions, "6. 尝试使用不同的 DNS 服务器: echo 'nameserver 8.8.8.8' > /etc/resolv.conf")
	}
	
	if strings.Contains(errMsg, "timeout") {
		suggestions = append(suggestions, "6. 增加超时时间或检查网络延迟")
	}
	
	if strings.Contains(context, "panel") {
		suggestions = append(suggestions, "6. 验证面板地址和端口是否正确")
		suggestions = append(suggestions, "7. 检查面板服务是否正常运行")
	}
	
	return "网络连接问题建议:\n" + strings.Join(suggestions, "\n")
}

// generatePermissionSuggestion 生成权限相关建议
func (aeh *AlpineErrorHandler) generatePermissionSuggestion(err error, context string) string {
	suggestions := []string{
		"1. 检查文件/目录权限: ls -la",
		"2. 确保以适当的用户身份运行",
		"3. 检查 SELinux/AppArmor 设置（如果适用）",
	}
	
	if strings.Contains(context, "config") {
		suggestions = append(suggestions, "4. 确保配置目录可写: chmod 755 /etc/sing-box")
		suggestions = append(suggestions, "5. 检查配置文件权限: chmod 644 /etc/sing-box/config.json")
	}
	
	if strings.Contains(context, "log") {
		suggestions = append(suggestions, "4. 确保日志目录可写: chmod 755 /tmp")
	}
	
	suggestions = append(suggestions, "6. 在容器中，检查是否需要特权模式")
	
	return "权限问题建议:\n" + strings.Join(suggestions, "\n")
}

// generateDependencySuggestion 生成依赖项相关建议
func (aeh *AlpineErrorHandler) generateDependencySuggestion(err error, context string) string {
	suggestions := []string{
		"1. 更新软件包索引: apk update",
	}
	
	errMsg := strings.ToLower(err.Error())
	
	if strings.Contains(errMsg, "sing-box") {
		suggestions = append(suggestions, "2. 安装 sing-box: 请参考官方安装文档")
		suggestions = append(suggestions, "3. 检查 sing-box 版本兼容性")
		suggestions = append(suggestions, "4. 验证 sing-box 可执行文件路径")
	} else if strings.Contains(errMsg, "curl") {
		suggestions = append(suggestions, "2. 安装 curl: apk add curl")
	} else if strings.Contains(errMsg, "wget") {
		suggestions = append(suggestions, "2. 安装 wget: apk add wget")
	} else {
		suggestions = append(suggestions, "2. 安装缺失的软件包: apk add <package-name>")
	}
	
	suggestions = append(suggestions, "3. 检查 PATH 环境变量")
	suggestions = append(suggestions, "4. 验证可执行文件权限: chmod +x <file>")
	
	return "依赖项问题建议:\n" + strings.Join(suggestions, "\n")
}

// generateConfigurationSuggestion 生成配置相关建议
func (aeh *AlpineErrorHandler) generateConfigurationSuggestion(err error, context string) string {
	suggestions := []string{
		"1. 检查配置文件语法",
		"2. 验证配置文件路径是否正确",
		"3. 确保配置文件格式正确（JSON/YAML）",
	}
	
	if strings.Contains(context, "sing-box") {
		suggestions = append(suggestions, "4. 使用 sing-box check 验证配置: sing-box check -c /etc/sing-box/config.json")
		suggestions = append(suggestions, "5. 参考 sing-box 官方配置文档")
	}
	
	suggestions = append(suggestions, "6. 检查配置文件权限和所有权")
	suggestions = append(suggestions, "7. 备份并重置为默认配置进行测试")
	
	return "配置问题建议:\n" + strings.Join(suggestions, "\n")
}

// generateResourceSuggestion 生成资源相关建议
func (aeh *AlpineErrorHandler) generateResourceSuggestion(err error, context string) string {
	suggestions := []string{
		"1. 检查系统资源使用情况:",
		"   - 内存: free -h",
		"   - 磁盘: df -h",
		"   - CPU: top",
	}
	
	errMsg := strings.ToLower(err.Error())
	
	if strings.Contains(errMsg, "memory") {
		suggestions = append(suggestions, "2. 释放内存:")
		suggestions = append(suggestions, "   - 停止不必要的进程")
		suggestions = append(suggestions, "   - 清理缓存: sync && echo 3 > /proc/sys/vm/drop_caches")
		suggestions = append(suggestions, "3. 考虑增加容器内存限制")
	}
	
	if strings.Contains(errMsg, "disk") || strings.Contains(errMsg, "space") {
		suggestions = append(suggestions, "2. 清理磁盘空间:")
		suggestions = append(suggestions, "   - 删除临时文件: rm -rf /tmp/*")
		suggestions = append(suggestions, "   - 清理日志文件")
		suggestions = append(suggestions, "   - 清理软件包缓存: apk cache clean")
	}
	
	suggestions = append(suggestions, "4. 在容器环境中，检查资源限制设置")
	
	return "资源问题建议:\n" + strings.Join(suggestions, "\n")
}

// generateCompatibilitySuggestion 生成兼容性相关建议
func (aeh *AlpineErrorHandler) generateCompatibilitySuggestion(err error, context string) string {
	suggestions := []string{
		"1. 检查 Alpine Linux 版本兼容性",
		"2. 验证可执行文件是否为正确的架构版本",
	}
	
	errMsg := strings.ToLower(err.Error())
	
	if strings.Contains(errMsg, "musl") {
		suggestions = append(suggestions, "3. musl libc 兼容性问题:")
		suggestions = append(suggestions, "   - 确保使用 musl 编译的版本")
		suggestions = append(suggestions, "   - 检查是否需要安装 gcompat: apk add gcompat")
		suggestions = append(suggestions, "   - 考虑使用静态编译版本")
	}
	
	if strings.Contains(errMsg, "library") || strings.Contains(errMsg, "symbol") {
		suggestions = append(suggestions, "3. 库依赖问题:")
		suggestions = append(suggestions, "   - 检查缺失的库: ldd <executable>")
		suggestions = append(suggestions, "   - 安装缺失的库文件")
	}
	
	suggestions = append(suggestions, "4. 尝试使用 Alpine 专用版本")
	suggestions = append(suggestions, "5. 检查环境变量设置")
	
	return "兼容性问题建议:\n" + strings.Join(suggestions, "\n")
}

// generateSystemSuggestion 生成系统相关建议
func (aeh *AlpineErrorHandler) generateSystemSuggestion(err error, context string) string {
	suggestions := []string{
		"1. 检查系统日志: dmesg | tail -20",
		"2. 检查进程状态: ps aux",
		"3. 检查系统负载: uptime",
		"4. 验证系统完整性",
	}
	
	if strings.Contains(context, "startup") {
		suggestions = append(suggestions, "5. 检查启动脚本和服务")
		suggestions = append(suggestions, "6. 验证系统初始化完成")
	}
	
	suggestions = append(suggestions, "7. 重启服务或系统（如果可能）")
	suggestions = append(suggestions, "8. 联系系统管理员或技术支持")
	
	return "系统问题建议:\n" + strings.Join(suggestions, "\n")
}

// generatePanicSuggestion 生成 panic 相关建议
func (aeh *AlpineErrorHandler) generatePanicSuggestion(r interface{}) string {
	suggestions := []string{
		"程序崩溃建议:",
		"1. 检查系统资源是否充足",
		"2. 验证配置文件是否正确",
		"3. 检查是否存在权限问题",
		"4. 查看完整的错误日志",
		"5. 尝试以调试模式重新启动",
		"6. 如果问题持续，请联系技术支持并提供:",
		"   - 完整的错误信息",
		"   - 系统环境信息", 
		"   - 操作步骤",
	}
	
	return strings.Join(suggestions, "\n")
}

// CaptureSystemContext 捕获系统上下文
func (aeh *AlpineErrorHandler) CaptureSystemContext() (*SystemInfo, error) {
	if aeh.systemChecker == nil {
		return nil, fmt.Errorf("系统检查器未初始化")
	}
	
	// 获取基本系统信息
	systemInfo, err := aeh.systemChecker.CheckSystemInfo()
	if err != nil {
		aeh.logger.Warn("无法获取完整系统信息: %v", err)
		// 返回部分信息
		return &SystemInfo{
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
			CPUCores:     runtime.NumCPU(),
			Timestamp:    time.Now(),
		}, nil
	}
	
	return systemInfo, nil
}

// captureStackTrace 捕获堆栈跟踪
func (aeh *AlpineErrorHandler) captureStackTrace() string {
	buf := make([]byte, 1024*64)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// RecoverFromPanic 从 panic 中恢复
func (aeh *AlpineErrorHandler) RecoverFromPanic() {
	if r := recover(); r != nil {
		alpineErr := aeh.HandlePanic(r)
		aeh.logger.Error("程序崩溃已恢复: %v", alpineErr.Message)
		aeh.logger.Error("建议: %s", alpineErr.Suggestion)
	}
}

// IsRetryableError 判断错误是否可重试
func (aeh *AlpineErrorHandler) IsRetryableError(err error) bool {
	if alpineErr, ok := err.(*AlpineError); ok {
		switch alpineErr.Category {
		case ErrorCategoryResource:
			return true // 资源错误可能是临时的
		case ErrorCategorySystem:
			// 检查是否是网络相关的系统错误
			errMsg := strings.ToLower(alpineErr.Message)
			if strings.Contains(errMsg, "network") ||
			   strings.Contains(errMsg, "connection") ||
			   strings.Contains(errMsg, "timeout") {
				return true
			}
			return false // 其他系统错误通常不可重试
		case ErrorCategoryPermission:
			return false // 权限错误需要手动修复
		case ErrorCategoryDependency:
			return false // 依赖项错误需要安装
		case ErrorCategoryConfiguration:
			return false // 配置错误需要修复
		case ErrorCategoryCompatibility:
			return false // 兼容性错误需要更换版本
		}
	}
	
	// 对于普通错误，检查错误消息
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "timeout") ||
	   strings.Contains(errMsg, "connection") ||
	   strings.Contains(errMsg, "temporary") {
		return true
	}
	
	return false
}