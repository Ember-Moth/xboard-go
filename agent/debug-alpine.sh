#!/bin/sh

# Alpine 环境下 Agent 增强诊断脚本
# 版本: 2.0
# 更新时间: 2025-12-18

echo "=========================================="
echo "XBoard Agent Alpine 环境增强诊断 v2.0"
echo "=========================================="
echo "诊断时间: $(date)"
echo ""

# 检查系统信息
echo "1. 系统信息:"
echo "   OS: $(uname -s)"
echo "   Kernel: $(uname -r)"
echo "   Architecture: $(uname -m)"
echo "   Alpine Version: $(cat /etc/alpine-release 2>/dev/null || echo 'Not Alpine')"
echo "   Uptime: $(uptime)"
echo "   Load Average: $(cat /proc/loadavg 2>/dev/null || echo 'N/A')"
echo "   CPU Cores: $(nproc 2>/dev/null || echo 'N/A')"
echo ""

# 检查容器环境
echo "1.1 容器环境检查:"
if [ -f /.dockerenv ]; then
    echo "   ✓ 检测到 Docker 容器环境"
elif grep -q docker /proc/1/cgroup 2>/dev/null; then
    echo "   ✓ 检测到容器环境 (cgroup)"
elif [ -n "$container" ]; then
    echo "   ✓ 检测到容器环境 (环境变量)"
else
    echo "   ✗ 未检测到容器环境"
fi

# 检查初始化系统
if [ -d /run/systemd/system ]; then
    echo "   ✓ 使用 systemd"
elif command -v rc-service >/dev/null 2>&1; then
    echo "   ✓ 使用 OpenRC"
else
    echo "   ? 未知初始化系统"
fi
echo ""

# 检查 Agent 文件
echo "2. Agent 文件检查:"
AGENT_PATH="./xboard-agent"
if [ -f "$AGENT_PATH" ]; then
    echo "   ✓ Agent 文件存在: $AGENT_PATH"
    echo "   文件大小: $(ls -lh $AGENT_PATH | awk '{print $5}')"
    echo "   权限: $(ls -l $AGENT_PATH | awk '{print $1}')"
    echo "   文件类型: $(file $AGENT_PATH)"
else
    echo "   ✗ Agent 文件不存在: $AGENT_PATH"
    echo "   可用的 Agent 文件:"
    ls -la xboard-agent* 2>/dev/null || echo "     没有找到 Agent 文件"
fi
echo ""

# 检查依赖库
echo "3. 依赖库检查:"
if command -v ldd >/dev/null 2>&1; then
    if [ -f "$AGENT_PATH" ]; then
        echo "   动态库依赖:"
        ldd "$AGENT_PATH" 2>/dev/null || echo "     静态编译或无法检查依赖"
    fi
else
    echo "   ldd 命令不可用"
fi
echo ""

# 检查 sing-box
echo "4. sing-box 检查:"
SINGBOX_PATH=$(which sing-box 2>/dev/null)
if [ -n "$SINGBOX_PATH" ]; then
    echo "   ✓ sing-box 路径: $SINGBOX_PATH"
    echo "   版本: $(sing-box version 2>/dev/null | head -1 || echo '无法获取版本')"
    echo "   权限: $(ls -l $SINGBOX_PATH | awk '{print $1}')"
else
    echo "   ✗ sing-box 未找到在 PATH 中"
    echo "   尝试查找 sing-box:"
    find /usr -name "sing-box" 2>/dev/null || echo "     未找到"
fi
echo ""

# 检查网络连接
echo "5. 网络连接检查:"
echo "   网络接口:"
ip addr show 2>/dev/null || ifconfig 2>/dev/null || echo "     无法获取网络接口信息"

echo "   DNS 配置:"
if [ -f /etc/resolv.conf ]; then
    echo "     DNS 服务器:"
    grep nameserver /etc/resolv.conf | head -3
else
    echo "     ✗ /etc/resolv.conf 不存在"
fi

echo "   DNS 解析测试:"
for domain in google.com cloudflare.com github.com; do
    if nslookup $domain >/dev/null 2>&1; then
        echo "     ✓ $domain"
    elif host $domain >/dev/null 2>&1; then
        echo "     ✓ $domain (使用 host)"
    else
        echo "     ✗ $domain"
    fi
done

echo "   网络连通性测试:"
if command -v ping >/dev/null 2>&1; then
    if ping -c 1 -W 3 8.8.8.8 >/dev/null 2>&1; then
        echo "     ✓ 可以 ping 通 8.8.8.8"
    else
        echo "     ✗ 无法 ping 通 8.8.8.8"
    fi
else
    echo "     ✗ ping 命令不可用"
fi

echo "   HTTPS 连接测试:"
if command -v wget >/dev/null 2>&1; then
    if wget --spider --timeout=5 https://www.google.com 2>/dev/null; then
        echo "     ✓ HTTPS 连接正常 (wget)"
    else
        echo "     ✗ HTTPS 连接失败 (wget)"
    fi
elif command -v curl >/dev/null 2>&1; then
    if curl -s --connect-timeout 5 https://www.google.com >/dev/null; then
        echo "     ✓ HTTPS 连接正常 (curl)"
    else
        echo "     ✗ HTTPS 连接失败 (curl)"
    fi
else
    echo "     ✗ wget 和 curl 都不可用"
fi

echo "   证书检查:"
if [ -d /etc/ssl/certs ]; then
    cert_count=$(ls /etc/ssl/certs/*.pem 2>/dev/null | wc -l)
    echo "     ✓ SSL 证书目录存在 ($cert_count 个证书)"
else
    echo "     ✗ SSL 证书目录不存在"
fi
echo ""

# 检查系统资源
echo "6. 系统资源:"
echo "   内存使用:"
free -h 2>/dev/null || echo "     free 命令不可用"
echo "   磁盘空间:"
df -h . 2>/dev/null || echo "     df 命令不可用"
echo "   临时目录空间:"
df -h /tmp 2>/dev/null || echo "     无法检查 /tmp"
echo "   inode 使用情况:"
df -i . 2>/dev/null || echo "     无法检查 inode"
echo ""

# 检查 Alpine 软件包管理
echo "6.1 Alpine 软件包检查:"
if command -v apk >/dev/null 2>&1; then
    echo "   ✓ apk 包管理器可用"
    echo "   已安装的关键软件包:"
    for pkg in musl libc6-compat gcompat ca-certificates; do
        if apk info -e $pkg >/dev/null 2>&1; then
            echo "     ✓ $pkg"
        else
            echo "     ✗ $pkg (未安装)"
        fi
    done
    echo "   软件包缓存状态:"
    apk stats 2>/dev/null || echo "     无法获取统计信息"
else
    echo "   ✗ apk 包管理器不可用"
fi
echo ""

# 检查 musl libc 兼容性
echo "6.2 musl libc 兼容性检查:"
if [ -f /lib/ld-musl-*.so.1 ]; then
    echo "   ✓ musl libc 已安装"
    echo "   musl 版本: $(ls -la /lib/ld-musl-*.so.1)"
else
    echo "   ✗ musl libc 未找到"
fi

if command -v ldd >/dev/null 2>&1; then
    echo "   ldd 版本: $(ldd --version 2>&1 | head -1)"
else
    echo "   ✗ ldd 不可用"
fi

# 检查 glibc 兼容层
if [ -f /lib/libc.so.6 ]; then
    echo "   ✓ glibc 兼容层存在"
elif apk info -e gcompat >/dev/null 2>&1; then
    echo "   ✓ gcompat 已安装"
else
    echo "   ✗ 无 glibc 兼容层"
fi
echo ""

# 检查进程和信号
echo "7. 进程检查:"
echo "   当前运行的 Agent 进程:"
ps aux | grep xboard-agent | grep -v grep || echo "     没有运行的 Agent 进程"
echo "   当前运行的 sing-box 进程:"
ps aux | grep sing-box | grep -v grep || echo "     没有运行的 sing-box 进程"
echo ""

# 尝试运行 Agent 并捕获错误
echo "8. Agent 启动测试:"
if [ -f "$AGENT_PATH" ]; then
    echo "   尝试显示帮助信息:"
    timeout 5 "$AGENT_PATH" -h 2>&1 || echo "     Agent 无法显示帮助或超时"
    echo ""
    
    echo "   尝试运行 Agent (5秒超时):"
    echo "   命令: $AGENT_PATH -panel https://example.com -token test123456789012345"
    timeout 5 "$AGENT_PATH" -panel https://example.com -token test123456789012345 2>&1 || echo "     Agent 运行失败或超时"
else
    echo "   跳过 Agent 测试 (文件不存在)"
fi
echo ""

# 检查日志和核心转储
echo "9. 错误信息检查:"
echo "   检查系统日志 (最近10行):"
if [ -f /var/log/messages ]; then
    tail -10 /var/log/messages | grep -i "xboard\|agent\|segfault\|killed" || echo "     没有相关错误信息"
elif [ -f /var/log/syslog ]; then
    tail -10 /var/log/syslog | grep -i "xboard\|agent\|segfault\|killed" || echo "     没有相关错误信息"
else
    echo "     系统日志文件不存在"
fi

echo "   检查核心转储:"
if [ -f core ]; then
    echo "     ✗ 发现核心转储文件: core"
    ls -la core
elif [ -f core.* ]; then
    echo "     ✗ 发现核心转储文件:"
    ls -la core.*
else
    echo "     ✓ 没有核心转储文件"
fi
echo ""

# 环境变量检查
echo "10. 环境变量:"
echo "    PATH: $PATH"
echo "    LD_LIBRARY_PATH: ${LD_LIBRARY_PATH:-未设置}"
echo "    TMPDIR: ${TMPDIR:-未设置}"
echo ""

# 提供调试建议
echo "=========================================="
echo "Alpine Linux 专用调试建议:"
echo "=========================================="
echo "1. 如果 Agent 立即崩溃:"
echo "   - 检查是否是正确的架构版本 (amd64/arm64)"
echo "   - 确保使用 musl 编译的版本或安装 gcompat:"
echo "     apk add gcompat"
echo "   - 尝试使用 strace 跟踪系统调用:"
echo "     apk add strace"
echo "     strace -f ./xboard-agent -panel <URL> -token <TOKEN>"
echo ""
echo "2. 如果是库依赖问题:"
echo "   - 安装必要的兼容库:"
echo "     apk add musl libc6-compat gcompat"
echo "   - 检查动态库依赖:"
echo "     ldd ./xboard-agent"
echo "   - 如果缺少库文件，尝试:"
echo "     apk add --no-cache ca-certificates tzdata"
echo ""
echo "3. 如果是 sing-box 问题:"
echo "   - 安装 sing-box (Alpine 版本):"
echo "     apk add sing-box"
echo "   - 或下载 Alpine 兼容版本"
echo "   - 检查版本兼容性:"
echo "     sing-box version"
echo "     sing-box check -c /etc/sing-box/config.json"
echo ""
echo "4. 如果是网络问题:"
echo "   - 安装网络工具:"
echo "     apk add curl wget bind-tools iputils"
echo "   - 检查 DNS 配置:"
echo "     cat /etc/resolv.conf"
echo "   - 更新 CA 证书:"
echo "     apk add ca-certificates && update-ca-certificates"
echo "   - 在容器中，检查网络模式和防火墙"
echo ""
echo "5. 如果是权限问题:"
echo "   - 确保 Agent 有执行权限:"
echo "     chmod +x xboard-agent"
echo "   - 创建必要的目录:"
echo "     mkdir -p /etc/sing-box /tmp"
echo "     chmod 755 /etc/sing-box"
echo "   - 在容器中，可能需要特权模式或适当的 capabilities"
echo ""
echo "6. 如果是资源问题:"
echo "   - 检查内存限制:"
echo "     cat /sys/fs/cgroup/memory/memory.limit_in_bytes 2>/dev/null"
echo "   - 清理空间:"
echo "     apk cache clean"
echo "     rm -rf /tmp/* /var/cache/apk/*"
echo ""
echo "7. 启用详细日志和调试:"
echo "   - 使用调试版本:"
echo "     ./xboard-agent-debug -debug -panel <URL> -token <TOKEN>"
echo "   - 设置环境变量:"
echo "     export DEBUG=1 TRACE=1"
echo "   - 运行完整诊断:"
echo "     ./xboard-agent-debug diagnose"
echo "   - 重定向输出:"
echo "     ./xboard-agent ... > agent.log 2>&1"
echo ""
echo "8. Alpine 容器特定建议:"
echo "   - 使用官方 Alpine 镜像"
echo "   - 确保容器有足够的内存和 CPU"
echo "   - 检查容器网络配置"
echo "   - 考虑使用 --privileged 或适当的 --cap-add"
echo ""
echo "9. 获取帮助:"
echo "   - 收集诊断信息:"
echo "     ./debug-alpine.sh > diagnostic-report.txt 2>&1"
echo "   - 包含系统信息、错误日志和配置文件"
echo "   - 联系技术支持时提供完整的诊断报告"
echo ""
echo "=========================================="
echo "诊断完成时间: $(date)"
echo "=========================================="