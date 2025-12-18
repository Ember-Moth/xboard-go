package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelError:
		return "ERROR"
	case LogLevelWarn:
		return "WARN"
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// DebugLogger 调试日志记录器
type DebugLogger struct {
	level      LogLevel
	output     io.Writer
	enableFile bool
	logFile    string
	fileWriter *os.File
	logger     *log.Logger
}

// NewDebugLogger 创建新的调试日志记录器
func NewDebugLogger(level LogLevel, enableFile bool) *DebugLogger {
	dl := &DebugLogger{
		level:      level,
		output:     os.Stdout,
		enableFile: enableFile,
		logFile:    "/tmp/xboard-agent-debug.log",
	}

	// 设置输出目标
	if enableFile {
		if err := dl.setupFileOutput(); err != nil {
			// 如果文件输出失败，回退到标准输出
			fmt.Printf("警告: 无法设置文件日志输出: %v\n", err)
			dl.enableFile = false
		}
	}

	// 创建日志记录器
	dl.logger = log.New(dl.output, "", 0)

	return dl
}

// setupFileOutput 设置文件输出
func (dl *DebugLogger) setupFileOutput() error {
	file, err := os.OpenFile(dl.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("无法打开日志文件 %s: %w", dl.logFile, err)
	}

	dl.fileWriter = file
	
	// 同时输出到文件和标准输出
	dl.output = io.MultiWriter(os.Stdout, file)
	
	return nil
}

// Close 关闭日志记录器
func (dl *DebugLogger) Close() error {
	if dl.fileWriter != nil {
		return dl.fileWriter.Close()
	}
	return nil
}

// shouldLog 检查是否应该记录指定级别的日志
func (dl *DebugLogger) shouldLog(level LogLevel) bool {
	return level <= dl.level
}

// formatMessage 格式化日志消息
func (dl *DebugLogger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(3)
	caller := "unknown"
	if ok {
		// 只保留文件名，不包含完整路径
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			caller = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}

	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level.String(), caller, message)
}

// log 内部日志记录方法
func (dl *DebugLogger) log(level LogLevel, format string, args ...interface{}) {
	if !dl.shouldLog(level) {
		return
	}

	message := dl.formatMessage(level, format, args...)
	dl.logger.Println(message)
}

// Error 记录错误级别日志
func (dl *DebugLogger) Error(format string, args ...interface{}) {
	dl.log(LogLevelError, format, args...)
}

// Warn 记录警告级别日志
func (dl *DebugLogger) Warn(format string, args ...interface{}) {
	dl.log(LogLevelWarn, format, args...)
}

// Info 记录信息级别日志
func (dl *DebugLogger) Info(format string, args ...interface{}) {
	dl.log(LogLevelInfo, format, args...)
}

// Debug 记录调试级别日志
func (dl *DebugLogger) Debug(format string, args ...interface{}) {
	dl.log(LogLevelDebug, format, args...)
}

// Trace 记录跟踪级别日志
func (dl *DebugLogger) Trace(format string, args ...interface{}) {
	dl.log(LogLevelTrace, format, args...)
}

// LogSystemInfo 记录系统信息
func (dl *DebugLogger) LogSystemInfo(info SystemInfo) {
	dl.Info("=== 系统信息 ===")
	dl.Info("操作系统: %s", info.OS)
	dl.Info("内核版本: %s", info.Kernel)
	dl.Info("架构: %s", info.Architecture)
	dl.Info("Alpine 版本: %s", info.AlpineVersion)
	dl.Info("CPU 核心数: %d", info.CPUCores)
	dl.Info("总内存: %d MB", info.MemoryTotal/1024/1024)
	dl.Info("可用内存: %d MB", info.MemoryFree/1024/1024)
	dl.Info("可用磁盘空间: %d MB", info.DiskFree/1024/1024)
	dl.Info("是否为容器环境: %v", info.IsContainer)
	dl.Info("是否有 systemd: %v", info.HasSystemd)
	dl.Info("是否有 OpenRC: %v", info.HasOpenRC)
	
	if len(info.NetworkInterfaces) > 0 {
		dl.Info("网络接口:")
		for _, iface := range info.NetworkInterfaces {
			dl.Info("  - %s: %v (状态: %v)", iface.Name, iface.Addresses, iface.IsUp)
		}
	}
	
	if len(info.DNSServers) > 0 {
		dl.Info("DNS 服务器: %v", info.DNSServers)
	}
	
	dl.Info("检查时间: %s", info.Timestamp.Format("2006-01-02 15:04:05"))
	dl.Info("================")
}

// LogAPIRequest 记录 API 请求
func (dl *DebugLogger) LogAPIRequest(method, url string, body interface{}) {
	dl.Debug("API 请求: %s %s", method, url)
	if body != nil {
		dl.Trace("请求体: %+v", body)
	}
}

// LogAPIResponse 记录 API 响应
func (dl *DebugLogger) LogAPIResponse(statusCode int, body interface{}) {
	dl.Debug("API 响应: HTTP %d", statusCode)
	if body != nil {
		dl.Trace("响应体: %+v", body)
	}
}

// LogError 记录错误信息（包含更多上下文）
func (dl *DebugLogger) LogError(err error, context string) {
	if err == nil {
		return
	}
	
	dl.Error("错误发生在 %s: %v", context, err)
	
	// 如果是 AlpineError，记录额外信息
	if alpineErr, ok := err.(*AlpineError); ok {
		dl.Error("错误类别: %s", alpineErr.Category)
		dl.Error("建议: %s", alpineErr.Suggestion)
		if alpineErr.SystemInfo != nil {
			dl.Debug("错误时系统信息: OS=%s, 内存=%dMB", 
				alpineErr.SystemInfo.OS, 
				alpineErr.SystemInfo.MemoryFree/1024/1024)
		}
		if alpineErr.StackTrace != "" {
			dl.Trace("堆栈跟踪:\n%s", alpineErr.StackTrace)
		}
	}
}

// LogDependencyCheck 记录依赖项检查结果
func (dl *DebugLogger) LogDependencyCheck(deps []DependencyInfo) {
	dl.Info("=== 依赖项检查结果 ===")
	for _, dep := range deps {
		status := "✓"
		if !dep.Available {
			status = "✗"
		}
		
		dl.Info("%s %s", status, dep.Name)
		if dep.Available {
			dl.Debug("  路径: %s", dep.Path)
			dl.Debug("  版本: %s", dep.Version)
		} else {
			dl.Warn("  错误: %v", dep.Error)
		}
	}
	dl.Info("=====================")
}

// LogNetworkTest 记录网络测试结果
func (dl *DebugLogger) LogNetworkTest(test NetworkTest) {
	status := "✓"
	if !test.Success {
		status = "✗"
	}
	
	dl.Info("%s 网络测试 [%s]: %s (耗时: %v)", status, test.Type, test.Target, test.Duration)
	if !test.Success && test.Error != "" {
		dl.Warn("  错误: %s", test.Error)
	}
}

// SetLevel 设置日志级别
func (dl *DebugLogger) SetLevel(level LogLevel) {
	dl.level = level
	dl.Debug("日志级别已设置为: %s", level.String())
}

// GetLevel 获取当前日志级别
func (dl *DebugLogger) GetLevel() LogLevel {
	return dl.level
}

// IsDebugEnabled 检查是否启用了调试级别
func (dl *DebugLogger) IsDebugEnabled() bool {
	return dl.shouldLog(LogLevelDebug)
}

// IsTraceEnabled 检查是否启用了跟踪级别
func (dl *DebugLogger) IsTraceEnabled() bool {
	return dl.shouldLog(LogLevelTrace)
}