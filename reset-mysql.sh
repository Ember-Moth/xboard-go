#!/bin/bash

# 重置 MySQL 数据库

INSTALL_DIR="/opt/dashgo"

echo "=== 重置 dashGO MySQL 数据库 ==="
echo ""
echo "⚠️  警告: 这将删除所有数据库数据！"
echo ""
read -p "确定要继续吗? [y/N]: " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo "已取消"
    exit 0
fi

echo ""
echo "1. 检查磁盘空间..."
df -h /

ROOT_USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "$ROOT_USAGE" -gt 90 ]; then
    echo ""
    echo "❌ 错误: 磁盘空间不足 (使用率 ${ROOT_USAGE}%)"
    echo ""
    echo "请先清理磁盘空间:"
    echo "  bash check-disk.sh"
    echo ""
    echo "快速清理命令:"
    echo "  docker system prune -a --volumes -f"
    echo "  journalctl --vacuum-time=3d"
    exit 1
fi

echo "✓ 磁盘空间充足 (使用率 ${ROOT_USAGE}%)"

if [ ! -d "$INSTALL_DIR" ]; then
    echo "错误: 安装目录不存在: $INSTALL_DIR"
    exit 1
fi

cd "$INSTALL_DIR"

echo ""
echo "2. 停止所有服务..."
docker compose down

echo ""
echo "3. 删除 MySQL 数据卷..."
docker volume rm dashgo_mysql_data 2>/dev/null || echo "   数据卷不存在或已删除"

echo ""
echo "4. 清理 MySQL 容器..."
docker rm -f dashgo-mysql 2>/dev/null || true

echo ""
echo "5. 重新启动服务..."
docker compose up -d

echo ""
echo "6. 等待 MySQL 初始化..."
sleep 10

echo ""
echo "7. 检查服务状态..."
docker compose ps

echo ""
echo "8. 查看 MySQL 日志..."
docker compose logs dashgo-mysql | tail -20

echo ""
echo "=== 完成 ==="
echo ""
echo "如果 MySQL 仍然不健康，请查看日志:"
echo "  docker compose logs dashgo-mysql"
echo ""
echo "检查健康状态:"
echo "  docker compose ps"
echo ""
echo "如果问题持续，可能需要:"
echo "  1. 清理更多磁盘空间"
echo "  2. 检查 Docker 配置"
echo "  3. 查看系统日志: journalctl -xe"