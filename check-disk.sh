#!/bin/bash

# 检查磁盘空间并提供清理建议

echo "=== 磁盘空间检查 ==="
echo ""

# 显示磁盘使用情况
echo "1. 磁盘使用情况:"
df -h

echo ""
echo "2. 检查根目录空间:"
ROOT_USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "$ROOT_USAGE" -gt 90 ]; then
    echo "   ⚠️  警告: 根目录使用率 ${ROOT_USAGE}% (超过 90%)"
else
    echo "   ✓ 根目录使用率 ${ROOT_USAGE}%"
fi

echo ""
echo "3. Docker 磁盘使用:"
docker system df 2>/dev/null || echo "   Docker 未运行或未安装"

echo ""
echo "4. 大文件查找 (前 10 个):"
du -sh /* 2>/dev/null | sort -rh | head -10

echo ""
echo "=== 清理建议 ==="
echo ""

if [ "$ROOT_USAGE" -gt 80 ]; then
    echo "磁盘空间不足，建议执行以下清理操作:"
    echo ""
    echo "1. 清理 Docker 未使用的资源:"
    echo "   docker system prune -a --volumes"
    echo ""
    echo "2. 清理系统日志:"
    echo "   journalctl --vacuum-time=3d"
    echo ""
    echo "3. 清理 APT 缓存 (Debian/Ubuntu):"
    echo "   apt-get clean"
    echo "   apt-get autoclean"
    echo ""
    echo "4. 清理 YUM 缓存 (CentOS/RHEL):"
    echo "   yum clean all"
    echo ""
    echo "5. 查找并删除大文件:"
    echo "   find / -type f -size +100M -exec ls -lh {} \\; 2>/dev/null"
    echo ""
    echo "6. 清理临时文件:"
    echo "   rm -rf /tmp/*"
    echo "   rm -rf /var/tmp/*"
else
    echo "✓ 磁盘空间充足"
fi

echo ""
echo "=== 快速清理命令 ==="
echo ""
echo "# 一键清理 Docker (会删除所有未使用的镜像、容器、卷)"
echo "docker system prune -a --volumes -f"
echo ""
echo "# 清理日志"
echo "journalctl --vacuum-time=3d"
echo ""
echo "# 清理包管理器缓存"
echo "apt-get clean || yum clean all"