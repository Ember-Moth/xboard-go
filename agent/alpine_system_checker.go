package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// AlpineSystemChecker Alpine 系统检查器
type AlpineSystemChecker struct {
	logger *DebugLogger
}

// NewAlpineSystemChecker 创建新的 Alpine 系统检查器
func NewAlpineSystemChecker(logger *DebugLogger) *AlpineSystemChecker {
	return &AlpineSystemChecker{
		logger: logger,
	}
}

// CheckSystemInfo 检查系统信息
func (asc *AlpineSystemChecker) CheckSystemInfo() (*SystemInfo, error) {
	asc.logger.Debug("开始检查系统信息")
	
	info := &SystemInfo{
		Timestamp: time.Now(),
	}
	
	// 检查操作系统
	info.OS = runtime.GOOS
	info.Architecture = runtime.GOARCH
	info.CPUCores = runtime.NumCPU()
	
	// 检查内核版本
	if kernel, err := asc.getKernelVersion(); err == nil {
		info.Kernel = kernel
	} else {
		asc.logger.Warn("无法获取内核版本: %v", err)
	}
	
	// 检查 Alpine 版本
	if alpineVersion, err := asc.getAlpineVersion(); err == nil {
		info.AlpineVersion = alpineVersion
	} else {
		asc.logger.Warn("无法获取 Alpine 版本: %v", err)
	}
	
	// 检查内存信息
	if memTotal, memFree, err := asc.getMemoryInfo(); err == nil {
		info.MemoryTotal = memTotal
		info.MemoryFree = memFree
	} else {
		asc.logger.Warn("无法获取内存信息: %v", err)
	}
	
	// 检查磁盘空间
	if diskFree, err := asc.getDiskSpace("/"); err == nil {
		info.DiskFree = diskFree
	} else {
		asc.logger.Warn("无法获取磁盘空间信息: %v", err)
	}
	
	// 检查容器环境
	info.IsContainer = asc.isContainerEnvironment()
	
	// 检查初始化系统
	info.HasSystemd = asc.hasSystemd()
	info.HasOpenRC = asc.hasOpenRC()
	
	// 检查网络接口
	if interfaces, err := asc.getNetworkInterfaces(); err == nil {
		info.NetworkInterfaces = interfaces
	} else {
		asc.logger.Warn("无法获取网络接口信息: %v", err)
	}
	
	// 检查 DNS 服务器
	if dnsServers, err := asc.getDNSServers(); err == nil {
		info.DNSServers = dnsServers
	} else {
		asc.logger.Warn("无法获取 DNS 服务器信息: %v", err)
	}
	
	asc.logger.Debug("系统信息检查完成")
	return info, nil
}

// getKernelVersion 获取内核版本
func (asc *AlpineSystemChecker) getKernelVersion() (string, error) {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("执行 uname 命令失败: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getAlpineVersion 获取 Alpine 版本
func (asc *AlpineSystemChecker) getAlpineVersion() (string, error) {
	// 尝试读取 /etc/alpine-release
	if content, err := os.ReadFile("/etc/alpine-release"); err == nil {
		return strings.TrimSpace(string(content)), nil
	}
	
	// 尝试从 /etc/os-release 获取
	if content, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "VERSION_ID=") {
				version := strings.TrimPrefix(line, "VERSION_ID=")
				version = strings.Trim(version, "\"")
				return version, nil
			}
		}
	}
	
	return "", fmt.Errorf("无法确定 Alpine 版本")
}

// getMemoryInfo 获取内存信息
func (asc *AlpineSystemChecker) getMemoryInfo() (total, free int64, err error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, fmt.Errorf("无法打开 /proc/meminfo: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		switch fields[0] {
		case "MemTotal:":
			if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
				total = val * 1024 // 转换为字节
			}
		case "MemAvailable:":
			if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
				free = val * 1024 // 转换为字节
			}
		case "MemFree:":
			// 如果没有 MemAvailable，使用 MemFree
			if free == 0 {
				if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
					free = val * 1024 // 转换为字节
				}
			}
		}
	}
	
	if total == 0 {
		return 0, 0, fmt.Errorf("无法解析内存信息")
	}
	
	return total, free, nil
}



// isContainerEnvironment 检查是否在容器环境中
func (asc *AlpineSystemChecker) isContainerEnvironment() bool {
	// 检查 /.dockerenv 文件
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	
	// 检查 /proc/1/cgroup
	if content, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		cgroupContent := string(content)
		if strings.Contains(cgroupContent, "docker") || 
		   strings.Contains(cgroupContent, "containerd") ||
		   strings.Contains(cgroupContent, "podman") {
			return true
		}
	}
	
	// 检查环境变量
	if os.Getenv("container") != "" {
		return true
	}
	
	return false
}

// hasSystemd 检查是否有 systemd
func (asc *AlpineSystemChecker) hasSystemd() bool {
	// 检查 /run/systemd/system 目录
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	
	// 检查 systemctl 命令
	if _, err := exec.LookPath("systemctl"); err == nil {
		return true
	}
	
	return false
}

// hasOpenRC 检查是否有 OpenRC
func (asc *AlpineSystemChecker) hasOpenRC() bool {
	// 检查 /sbin/openrc
	if _, err := os.Stat("/sbin/openrc"); err == nil {
		return true
	}
	
	// 检查 rc-service 命令
	if _, err := exec.LookPath("rc-service"); err == nil {
		return true
	}
	
	// 检查 /etc/init.d 目录
	if _, err := os.Stat("/etc/init.d"); err == nil {
		return true
	}
	
	return false
}

// getNetworkInterfaces 获取网络接口信息
func (asc *AlpineSystemChecker) getNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("无法获取网络接口: %w", err)
	}
	
	var result []NetworkInterface
	for _, iface := range interfaces {
		netIface := NetworkInterface{
			Name: iface.Name,
			IsUp: iface.Flags&net.FlagUp != 0,
		}
		
		// 获取接口地址
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				netIface.Addresses = append(netIface.Addresses, addr.String())
			}
		}
		
		result = append(result, netIface)
	}
	
	return result, nil
}

// getDNSServers 获取 DNS 服务器
func (asc *AlpineSystemChecker) getDNSServers() ([]string, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("无法打开 /etc/resolv.conf: %w", err)
	}
	defer file.Close()
	
	var dnsServers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "nameserver ") {
			dns := strings.TrimPrefix(line, "nameserver ")
			dns = strings.Fields(dns)[0] // 只取第一个字段
			dnsServers = append(dnsServers, dns)
		}
	}
	
	return dnsServers, nil
}

// CheckDependencies 检查依赖项
func (asc *AlpineSystemChecker) CheckDependencies() ([]DependencyInfo, error) {
	asc.logger.Debug("开始检查依赖项")
	
	dependencies := []string{
		"sing-box",
		"curl",
		"wget",
		"ping",
		"nslookup",
		"dig",
		"netstat",
		"ss",
		"ps",
		"top",
		"free",
		"df",
		"lsof",
		"strace",
	}
	
	var result []DependencyInfo
	for _, dep := range dependencies {
		info := asc.checkSingleDependency(dep)
		result = append(result, info)
	}
	
	asc.logger.Debug("依赖项检查完成，共检查 %d 个依赖项", len(result))
	return result, nil
}

// checkSingleDependency 检查单个依赖项
func (asc *AlpineSystemChecker) checkSingleDependency(name string) DependencyInfo {
	info := DependencyInfo{
		Name: name,
	}
	
	// 查找可执行文件路径
	path, err := exec.LookPath(name)
	if err != nil {
		info.Available = false
		info.Error = fmt.Errorf("命令不可用: %w", err)
		return info
	}
	
	info.Available = true
	info.Path = path
	
	// 尝试获取版本信息
	version := asc.getCommandVersion(name)
	info.Version = version
	
	return info
}

// getCommandVersion 获取命令版本
func (asc *AlpineSystemChecker) getCommandVersion(command string) string {
	// 常见的版本参数
	versionArgs := [][]string{
		{"--version"},
		{"-V"},
		{"-v"},
		{"version"},
	}
	
	for _, args := range versionArgs {
		cmd := exec.Command(command, args...)
		output, err := cmd.Output()
		if err == nil {
			// 取第一行作为版本信息
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
				return strings.TrimSpace(lines[0])
			}
		}
	}
	
	return "未知版本"
}

// CheckNetworkConnectivity 检查网络连接性
func (asc *AlpineSystemChecker) CheckNetworkConnectivity() error {
	asc.logger.Debug("开始检查网络连接性")
	
	// 测试 DNS 解析
	if err := asc.testDNSResolution(); err != nil {
		return fmt.Errorf("DNS 解析测试失败: %w", err)
	}
	
	// 测试 HTTPS 连接
	if err := asc.testHTTPSConnection(); err != nil {
		return fmt.Errorf("HTTPS 连接测试失败: %w", err)
	}
	
	asc.logger.Debug("网络连接性检查完成")
	return nil
}

// testDNSResolution 测试 DNS 解析
func (asc *AlpineSystemChecker) testDNSResolution() error {
	testDomains := []string{
		"google.com",
		"cloudflare.com",
		"github.com",
	}
	
	for _, domain := range testDomains {
		_, err := net.LookupHost(domain)
		if err == nil {
			asc.logger.Debug("DNS 解析测试成功: %s", domain)
			return nil
		}
		asc.logger.Debug("DNS 解析失败: %s - %v", domain, err)
	}
	
	return fmt.Errorf("所有 DNS 解析测试都失败")
}

// testHTTPSConnection 测试 HTTPS 连接
func (asc *AlpineSystemChecker) testHTTPSConnection() error {
	testURLs := []string{
		"https://www.google.com",
		"https://www.cloudflare.com",
		"https://www.github.com",
	}
	
	for _, url := range testURLs {
		if err := asc.testSingleHTTPSConnection(url); err == nil {
			asc.logger.Debug("HTTPS 连接测试成功: %s", url)
			return nil
		} else {
			asc.logger.Debug("HTTPS 连接失败: %s - %v", url, err)
		}
	}
	
	return fmt.Errorf("所有 HTTPS 连接测试都失败")
}

// testSingleHTTPSConnection 测试单个 HTTPS 连接
func (asc *AlpineSystemChecker) testSingleHTTPSConnection(url string) error {
	// 使用 curl 命令测试（如果可用）
	if _, err := exec.LookPath("curl"); err == nil {
		cmd := exec.Command("curl", "-s", "--connect-timeout", "5", "--max-time", "10", "-I", url)
		return cmd.Run()
	}
	
	// 使用 wget 命令测试（如果可用）
	if _, err := exec.LookPath("wget"); err == nil {
		cmd := exec.Command("wget", "--timeout=5", "--tries=1", "--spider", url)
		return cmd.Run()
	}
	
	return fmt.Errorf("没有可用的 HTTP 客户端工具")
}

// CheckFilePermissions 检查文件权限
func (asc *AlpineSystemChecker) CheckFilePermissions(paths []string) error {
	asc.logger.Debug("开始检查文件权限")
	
	for _, path := range paths {
		if err := asc.checkSingleFilePermission(path); err != nil {
			asc.logger.Warn("文件权限检查失败: %s - %v", path, err)
			return fmt.Errorf("文件权限检查失败 %s: %w", path, err)
		}
	}
	
	asc.logger.Debug("文件权限检查完成")
	return nil
}

// checkSingleFilePermission 检查单个文件权限
func (asc *AlpineSystemChecker) checkSingleFilePermission(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", path)
		}
		return fmt.Errorf("无法获取文件信息: %w", err)
	}
	
	mode := info.Mode()
	asc.logger.Debug("文件 %s 权限: %v", path, mode)
	
	// 检查是否可读
	if _, err := os.Open(path); err != nil {
		return fmt.Errorf("文件不可读: %w", err)
	}
	
	return nil
}

// CheckContainerEnvironment 检查容器环境
func (asc *AlpineSystemChecker) CheckContainerEnvironment() (bool, error) {
	asc.logger.Debug("开始检查容器环境")
	
	isContainer := asc.isContainerEnvironment()
	
	if isContainer {
		asc.logger.Info("检测到容器环境")
		
		// 检查容器特定的配置
		if err := asc.checkContainerSpecificSettings(); err != nil {
			asc.logger.Warn("容器特定设置检查失败: %v", err)
		}
	} else {
		asc.logger.Info("未检测到容器环境")
	}
	
	return isContainer, nil
}

// checkContainerSpecificSettings 检查容器特定设置
func (asc *AlpineSystemChecker) checkContainerSpecificSettings() error {
	// 检查是否有足够的权限
	if err := asc.checkContainerPrivileges(); err != nil {
		return fmt.Errorf("容器权限检查失败: %w", err)
	}
	
	// 检查网络配置
	if err := asc.checkContainerNetworking(); err != nil {
		return fmt.Errorf("容器网络检查失败: %w", err)
	}
	
	return nil
}

// checkContainerPrivileges 检查容器权限
func (asc *AlpineSystemChecker) checkContainerPrivileges() error {
	// 检查是否可以创建文件
	testFile := "/tmp/xboard-agent-test"
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("无法创建测试文件: %w", err)
	}
	file.Close()
	os.Remove(testFile)
	
	// 检查是否可以执行网络操作
	// 这里可以添加更多的权限检查
	
	return nil
}

// checkContainerNetworking 检查容器网络
func (asc *AlpineSystemChecker) checkContainerNetworking() error {
	// 检查网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("无法获取网络接口: %w", err)
	}
	
	hasValidInterface := false
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Name != "lo" {
			hasValidInterface = true
			break
		}
	}
	
	if !hasValidInterface {
		return fmt.Errorf("没有可用的网络接口")
	}
	
	return nil
}

// GetToolAvailability 获取工具可用性
func (asc *AlpineSystemChecker) GetToolAvailability() ([]ToolAvailability, error) {
	asc.logger.Debug("开始检查工具可用性")
	
	tools := map[string]string{
		"curl":     "wget 或内置 HTTP 客户端",
		"wget":     "curl 或内置 HTTP 客户端", 
		"ping":     "内置网络测试",
		"nslookup": "dig 或内置 DNS 解析",
		"dig":      "nslookup 或内置 DNS 解析",
		"netstat":  "ss 或 /proc/net 解析",
		"ss":       "netstat 或 /proc/net 解析",
		"lsof":     "/proc 文件系统解析",
		"strace":   "内置调试功能",
		"free":     "/proc/meminfo 解析",
		"df":       "syscall 获取磁盘信息",
	}
	
	var result []ToolAvailability
	for tool, alternative := range tools {
		availability := ToolAvailability{
			Name:        tool,
			Alternative: alternative,
		}
		
		if path, err := exec.LookPath(tool); err == nil {
			availability.Available = true
			availability.Path = path
			availability.Version = asc.getCommandVersion(tool)
		} else {
			availability.Available = false
		}
		
		result = append(result, availability)
	}
	
	asc.logger.Debug("工具可用性检查完成，共检查 %d 个工具", len(result))
	return result, nil
}