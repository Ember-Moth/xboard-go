package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DiagnosticTool 诊断工具
type DiagnosticTool struct {
	logger        *DebugLogger
	systemChecker *AlpineSystemChecker
	scriptPath    string
}

// NewDiagnosticTool 创建新的诊断工具
func NewDiagnosticTool(logger *DebugLogger, systemChecker *AlpineSystemChecker) *DiagnosticTool {
	return &DiagnosticTool{
		logger:        logger,
		systemChecker: systemChecker,
		scriptPath:    "./debug-alpine.sh",
	}
}

// RunFullDiagnostic 运行完整诊断
func (dt *DiagnosticTool) RunFullDiagnostic() (*DiagnosticResult, error) {
	dt.logger.Info("开始运行完整诊断...")
	
	result := &DiagnosticResult{
		Timestamp: time.Now(),
	}
	
	// 1. 系统信息检查
	dt.logger.Debug("检查系统信息...")
	if systemInfo, err := dt.systemChecker.CheckSystemInfo(); err == nil {
		result.SystemInfo = systemInfo
		dt.logger.LogSystemInfo(*systemInfo)
	} else {
		dt.logger.Error("系统信息检查失败: %v", err)
		result.Errors = append(result.Errors, AlpineError{
			Category:  ErrorCategorySystem,
			Message:   fmt.Sprintf("系统信息检查失败: %v", err),
			Timestamp: time.Now(),
		})
	}
	
	// 2. 依赖项检查
	dt.logger.Debug("检查依赖项...")
	if deps, err := dt.systemChecker.CheckDependencies(); err == nil {
		result.Dependencies = deps
		dt.logger.LogDependencyCheck(deps)
		
		// 检查关键依赖项
		dt.checkCriticalDependencies(deps, result)
	} else {
		dt.logger.Error("依赖项检查失败: %v", err)
		result.Errors = append(result.Errors, AlpineError{
			Category:  ErrorCategoryDependency,
			Message:   fmt.Sprintf("依赖项检查失败: %v", err),
			Timestamp: time.Now(),
		})
	}
	
	// 3. 网络连接测试
	dt.logger.Debug("测试网络连接...")
	networkTests := dt.runNetworkTests()
	result.NetworkTests = networkTests
	
	// 4. 文件检查
	dt.logger.Debug("检查重要文件...")
	fileChecks := dt.runFileChecks()
	result.FileChecks = fileChecks
	
	// 5. 容器环境检查
	dt.logger.Debug("检查容器环境...")
	if isContainer, err := dt.systemChecker.CheckContainerEnvironment(); err == nil {
		if isContainer {
			result.Suggestions = append(result.Suggestions, "检测到容器环境，已调整诊断策略")
		}
	}
	
	// 6. 生成建议
	dt.generateSuggestions(result)
	
	dt.logger.Info("完整诊断完成")
	return result, nil
}

// RunQuickCheck 运行快速检查
func (dt *DiagnosticTool) RunQuickCheck() (*DiagnosticResult, error) {
	dt.logger.Info("开始运行快速检查...")
	
	result := &DiagnosticResult{
		Timestamp: time.Now(),
	}
	
	// 基本系统信息
	if systemInfo, err := dt.systemChecker.CheckSystemInfo(); err == nil {
		result.SystemInfo = systemInfo
	}
	
	// 关键依赖项检查
	criticalDeps := []string{"sing-box"}
	for _, dep := range criticalDeps {
		depInfo := dt.systemChecker.checkSingleDependency(dep)
		result.Dependencies = append(result.Dependencies, depInfo)
		
		if !depInfo.Available {
			result.Errors = append(result.Errors, AlpineError{
				Category:  ErrorCategoryDependency,
				Message:   fmt.Sprintf("关键依赖项缺失: %s", dep),
				Suggestion: fmt.Sprintf("请安装 %s", dep),
				Timestamp: time.Now(),
			})
		}
	}
	
	// 基本网络测试
	basicNetworkTest := dt.testBasicNetworkConnectivity()
	result.NetworkTests = append(result.NetworkTests, basicNetworkTest)
	
	dt.logger.Info("快速检查完成")
	return result, nil
}

// ExecuteShellScript 执行诊断脚本
func (dt *DiagnosticTool) ExecuteShellScript() (string, error) {
	dt.logger.Debug("执行诊断脚本: %s", dt.scriptPath)
	
	// 检查脚本是否存在
	if _, err := os.Stat(dt.scriptPath); os.IsNotExist(err) {
		dt.logger.Warn("诊断脚本不存在: %s", dt.scriptPath)
		return "", fmt.Errorf("诊断脚本不存在: %s", dt.scriptPath)
	}
	
	// 确保脚本可执行
	if err := os.Chmod(dt.scriptPath, 0755); err != nil {
		dt.logger.Warn("无法设置脚本权限: %v", err)
	}
	
	// 执行脚本
	cmd := exec.Command("/bin/sh", dt.scriptPath)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		dt.logger.Error("脚本执行失败: %v", err)
		return string(output), fmt.Errorf("脚本执行失败: %w", err)
	}
	
	dt.logger.Debug("脚本执行成功，输出长度: %d", len(output))
	return string(output), nil
}

// GenerateReport 生成诊断报告
func (dt *DiagnosticTool) GenerateReport(result *DiagnosticResult) string {
	dt.logger.Debug("生成诊断报告...")
	
	report := &DiagnosticReport{
		ReportID:     dt.generateReportID(),
		Timestamp:    time.Now(),
		AgentVersion: Version,
	}
	
	if result.SystemInfo != nil {
		report.SystemInfo = *result.SystemInfo
	}
	
	report.Dependencies = result.Dependencies
	report.NetworkTests = result.NetworkTests
	report.FileChecks = result.FileChecks
	report.Errors = result.Errors
	report.Warnings = result.Warnings
	report.Suggestions = result.Suggestions
	
	// 确定整体状态
	report.OverallStatus = dt.determineOverallStatus(result)
	
	// 格式化为可读文本
	return dt.formatReportAsText(report)
}

// checkCriticalDependencies 检查关键依赖项
func (dt *DiagnosticTool) checkCriticalDependencies(deps []DependencyInfo, result *DiagnosticResult) {
	criticalDeps := map[string]string{
		"sing-box": "核心代理软件",
		"curl":     "HTTP 客户端（可选，可用 wget 替代）",
		"wget":     "HTTP 客户端（可选，可用 curl 替代）",
	}
	
	for _, dep := range deps {
		if description, isCritical := criticalDeps[dep.Name]; isCritical {
			if !dep.Available {
				if dep.Name == "sing-box" {
					// sing-box 是必需的
					result.Errors = append(result.Errors, AlpineError{
						Category:   ErrorCategoryDependency,
						Message:    fmt.Sprintf("缺少关键依赖项: %s (%s)", dep.Name, description),
						Suggestion: "请安装 sing-box，参考官方文档",
						Timestamp:  time.Now(),
					})
				} else {
					// 其他工具是可选的
					result.Warnings = append(result.Warnings, 
						fmt.Sprintf("缺少工具: %s (%s)", dep.Name, description))
				}
			}
		}
	}
}

// runNetworkTests 运行网络测试
func (dt *DiagnosticTool) runNetworkTests() []NetworkTest {
	var tests []NetworkTest
	
	// DNS 解析测试
	dnsTest := dt.testDNSResolution("google.com")
	tests = append(tests, dnsTest)
	dt.logger.LogNetworkTest(dnsTest)
	
	// HTTPS 连接测试
	httpsTest := dt.testHTTPSConnection("https://www.google.com")
	tests = append(tests, httpsTest)
	dt.logger.LogNetworkTest(httpsTest)
	
	// 基本连通性测试
	basicTest := dt.testBasicNetworkConnectivity()
	tests = append(tests, basicTest)
	dt.logger.LogNetworkTest(basicTest)
	
	return tests
}

// testDNSResolution 测试 DNS 解析
func (dt *DiagnosticTool) testDNSResolution(domain string) NetworkTest {
	start := time.Now()
	test := NetworkTest{
		Target: domain,
		Type:   "dns",
	}
	
	// 使用 nslookup 或 dig 测试
	var cmd *exec.Cmd
	if _, err := exec.LookPath("nslookup"); err == nil {
		cmd = exec.Command("nslookup", domain)
	} else if _, err := exec.LookPath("dig"); err == nil {
		cmd = exec.Command("dig", "+short", domain)
	} else {
		// 使用内置 Go 解析
		if err := dt.systemChecker.testDNSResolution(); err == nil {
			test.Success = true
		} else {
			test.Success = false
			test.Error = err.Error()
		}
		test.Duration = time.Since(start)
		return test
	}
	
	if err := cmd.Run(); err == nil {
		test.Success = true
	} else {
		test.Success = false
		test.Error = err.Error()
	}
	
	test.Duration = time.Since(start)
	return test
}

// testHTTPSConnection 测试 HTTPS 连接
func (dt *DiagnosticTool) testHTTPSConnection(url string) NetworkTest {
	start := time.Now()
	test := NetworkTest{
		Target: url,
		Type:   "https",
	}
	
	if err := dt.systemChecker.testSingleHTTPSConnection(url); err == nil {
		test.Success = true
	} else {
		test.Success = false
		test.Error = err.Error()
	}
	
	test.Duration = time.Since(start)
	return test
}

// testBasicNetworkConnectivity 测试基本网络连通性
func (dt *DiagnosticTool) testBasicNetworkConnectivity() NetworkTest {
	start := time.Now()
	test := NetworkTest{
		Target: "network-interfaces",
		Type:   "basic",
	}
	
	// 检查网络接口
	if interfaces, err := dt.systemChecker.getNetworkInterfaces(); err == nil {
		hasActiveInterface := false
		for _, iface := range interfaces {
			if iface.IsUp && iface.Name != "lo" && len(iface.Addresses) > 0 {
				hasActiveInterface = true
				break
			}
		}
		
		if hasActiveInterface {
			test.Success = true
		} else {
			test.Success = false
			test.Error = "没有活动的网络接口"
		}
	} else {
		test.Success = false
		test.Error = err.Error()
	}
	
	test.Duration = time.Since(start)
	return test
}

// runFileChecks 运行文件检查
func (dt *DiagnosticTool) runFileChecks() []FileCheck {
	var checks []FileCheck
	
	importantPaths := []string{
		"/etc/sing-box",
		"/etc/sing-box/config.json",
		"/tmp",
		"/etc/resolv.conf",
	}
	
	for _, path := range importantPaths {
		check := dt.checkSingleFile(path)
		checks = append(checks, check)
	}
	
	return checks
}

// checkSingleFile 检查单个文件
func (dt *DiagnosticTool) checkSingleFile(path string) FileCheck {
	check := FileCheck{
		Path: path,
	}
	
	info, err := os.Stat(path)
	if err != nil {
		check.Exists = false
		if !os.IsNotExist(err) {
			check.Error = err.Error()
		}
		return check
	}
	
	check.Exists = true
	check.Permissions = info.Mode()
	check.Size = info.Size()
	
	return check
}

// generateSuggestions 生成建议
func (dt *DiagnosticTool) generateSuggestions(result *DiagnosticResult) {
	// 基于错误生成建议
	for _, err := range result.Errors {
		if err.Suggestion != "" {
			result.Suggestions = append(result.Suggestions, err.Suggestion)
		}
	}
	
	// 基于系统信息生成建议
	if result.SystemInfo != nil {
		if result.SystemInfo.MemoryFree < 100*1024*1024 { // 小于 100MB
			result.Suggestions = append(result.Suggestions, 
				"可用内存较低，建议释放内存或增加内存限制")
		}
		
		if result.SystemInfo.DiskFree < 500*1024*1024 { // 小于 500MB
			result.Suggestions = append(result.Suggestions, 
				"可用磁盘空间较低，建议清理临时文件")
		}
		
		if result.SystemInfo.IsContainer {
			result.Suggestions = append(result.Suggestions, 
				"检测到容器环境，确保容器有足够的权限和资源")
		}
	}
	
	// 基于网络测试生成建议
	allNetworkTestsFailed := true
	for _, test := range result.NetworkTests {
		if test.Success {
			allNetworkTestsFailed = false
			break
		}
	}
	
	if allNetworkTestsFailed && len(result.NetworkTests) > 0 {
		result.Suggestions = append(result.Suggestions, 
			"所有网络测试都失败，请检查网络配置和连接")
	}
}

// determineOverallStatus 确定整体状态
func (dt *DiagnosticTool) determineOverallStatus(result *DiagnosticResult) string {
	if len(result.Errors) > 0 {
		return "error"
	}
	
	if len(result.Warnings) > 0 {
		return "warning"
	}
	
	return "healthy"
}

// generateReportID 生成报告 ID
func (dt *DiagnosticTool) generateReportID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// formatReportAsText 将报告格式化为文本
func (dt *DiagnosticTool) formatReportAsText(report *DiagnosticReport) string {
	var sb strings.Builder
	
	sb.WriteString("=== XBoard Agent 诊断报告 ===\n")
	sb.WriteString(fmt.Sprintf("报告 ID: %s\n", report.ReportID))
	sb.WriteString(fmt.Sprintf("生成时间: %s\n", report.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Agent 版本: %s\n", report.AgentVersion))
	sb.WriteString(fmt.Sprintf("整体状态: %s\n", report.OverallStatus))
	sb.WriteString("\n")
	
	// 系统信息
	sb.WriteString("=== 系统信息 ===\n")
	sb.WriteString(fmt.Sprintf("操作系统: %s\n", report.SystemInfo.OS))
	sb.WriteString(fmt.Sprintf("架构: %s\n", report.SystemInfo.Architecture))
	sb.WriteString(fmt.Sprintf("Alpine 版本: %s\n", report.SystemInfo.AlpineVersion))
	sb.WriteString(fmt.Sprintf("CPU 核心: %d\n", report.SystemInfo.CPUCores))
	sb.WriteString(fmt.Sprintf("内存: %d MB / %d MB\n", 
		report.SystemInfo.MemoryFree/1024/1024,
		report.SystemInfo.MemoryTotal/1024/1024))
	sb.WriteString(fmt.Sprintf("磁盘空间: %d MB\n", report.SystemInfo.DiskFree/1024/1024))
	sb.WriteString(fmt.Sprintf("容器环境: %v\n", report.SystemInfo.IsContainer))
	sb.WriteString("\n")
	
	// 依赖项检查
	if len(report.Dependencies) > 0 {
		sb.WriteString("=== 依赖项检查 ===\n")
		for _, dep := range report.Dependencies {
			status := "✓"
			if !dep.Available {
				status = "✗"
			}
			sb.WriteString(fmt.Sprintf("%s %s", status, dep.Name))
			if dep.Available && dep.Version != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", dep.Version))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	
	// 网络测试
	if len(report.NetworkTests) > 0 {
		sb.WriteString("=== 网络测试 ===\n")
		for _, test := range report.NetworkTests {
			status := "✓"
			if !test.Success {
				status = "✗"
			}
			sb.WriteString(fmt.Sprintf("%s %s [%s] (耗时: %v)\n", 
				status, test.Target, test.Type, test.Duration))
			if !test.Success && test.Error != "" {
				sb.WriteString(fmt.Sprintf("  错误: %s\n", test.Error))
			}
		}
		sb.WriteString("\n")
	}
	
	// 错误信息
	if len(report.Errors) > 0 {
		sb.WriteString("=== 错误信息 ===\n")
		for i, err := range report.Errors {
			sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, err.Category, err.Message))
			if err.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("   建议: %s\n", err.Suggestion))
			}
		}
		sb.WriteString("\n")
	}
	
	// 警告信息
	if len(report.Warnings) > 0 {
		sb.WriteString("=== 警告信息 ===\n")
		for i, warning := range report.Warnings {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, warning))
		}
		sb.WriteString("\n")
	}
	
	// 建议
	if len(report.Suggestions) > 0 {
		sb.WriteString("=== 建议 ===\n")
		for i, suggestion := range report.Suggestions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
		}
		sb.WriteString("\n")
	}
	
	sb.WriteString("=== 报告结束 ===\n")
	
	return sb.String()
}