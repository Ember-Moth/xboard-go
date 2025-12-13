#!/bin/bash

# dashGO 一键修复脚本

set -e

INSTALL_DIR="/opt/dashgo"

echo "=========================================="
echo "   dashGO 一键修复脚本"
echo "=========================================="
echo ""

# 检查 root 权限
if [ "$EUID" -ne 0 ]; then
    echo "❌ 请使用 root 用户运行此脚本"
    exit 1
fi

# 1. 检查磁盘空间
echo "1. 检查磁盘空间..."
ROOT_USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
echo "   根目录使用率: ${ROOT_USAGE}%"

if [ "$ROOT_USAGE" -gt 90 ]; then
    echo ""
    echo "⚠️  磁盘空间不足，开始清理..."
    
    # 停止服务
    if [ -d "$INSTALL_DIR" ]; then
        cd "$INSTALL_DIR"
        docker compose down 2>/dev/null || true
    fi
    
    # 清理 Docker
    echo "   清理 Docker 资源..."
    docker system prune -a --volumes -f
    
    # 清理日志
    echo "   清理系统日志..."
    journalctl --vacuum-time=3d 2>/dev/null || true
    
    # 清理包管理器缓存
    echo "   清理包管理器缓存..."
    apt-get clean 2>/dev/null || yum clean all 2>/dev/null || true
    
    # 再次检查
    ROOT_USAGE=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    echo "   清理后使用率: ${ROOT_USAGE}%"
    
    if [ "$ROOT_USAGE" -gt 90 ]; then
        echo ""
        echo "❌ 磁盘空间仍然不足，请手动清理更多空间"
        echo ""
        echo "建议:"
        echo "  1. 删除不需要的文件"
        echo "  2. 清理 /tmp 目录"
        echo "  3. 检查大文件: du -sh /* | sort -rh | head -10"
        exit 1
    fi
fi

echo "   ✓ 磁盘空间充足"

# 2. 检查安装目录
echo ""
echo "2. 检查安装目录..."
if [ ! -d "$INSTALL_DIR" ]; then
    echo "   ❌ 安装目录不存在: $INSTALL_DIR"
    echo "   请先运行安装脚本: bash install.sh"
    exit 1
fi
echo "   ✓ 安装目录存在"

cd "$INSTALL_DIR"

# 3. 修复配置文件
echo ""
echo "3. 修复配置文件..."
mkdir -p configs

if [ ! -f "configs/config.yaml" ]; then
    echo "   创建配置文件..."
    
    # 生成随机密码
    DB_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
    REDIS_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
    JWT_SECRET=$(openssl rand -base64 32)
    NODE_TOKEN=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)
    
    cat > configs/config.yaml << EOF
app:
  name: "dashGO"
  mode: "release"
  listen: ":8080"

database:
  driver: "mysql"
  host: "mysql"
  port: 3306
  username: "root"
  password: "${DB_PASS}"
  database: "dashgo"

redis:
  host: "redis"
  port: 6379
  password: "${REDIS_PASS}"
  db: 0

jwt:
  secret: "${JWT_SECRET}"
  expire_hour: 24

node:
  token: "${NODE_TOKEN}"
  push_interval: 60
  pull_interval: 60
  enable_sync: false

mail:
  host: "smtp.example.com"
  port: 587
  username: ""
  password: ""
  from_name: "dashGO"
  from_addr: "noreply@example.com"
  encryption: "tls"

telegram:
  bot_token: ""
  chat_id: ""

admin:
  email: "admin@example.com"
  password: "admin123456"
EOF
    
    # 更新 .env
    cat > .env << EOF
MYSQL_ROOT_PASSWORD=${DB_PASS}
MYSQL_DATABASE=dashgo
REDIS_PASSWORD=${REDIS_PASS}
JWT_SECRET=${JWT_SECRET}
NODE_TOKEN=${NODE_TOKEN}
EOF
    
    echo "   ✓ 配置文件已创建"
    echo ""
    echo "   重要信息:"
    echo "     数据库密码: ${DB_PASS}"
    echo "     Redis 密码: ${REDIS_PASS}"
    echo "     节点 Token: ${NODE_TOKEN}"
else
    echo "   ✓ 配置文件已存在"
fi

# 4. 重置 MySQL
echo ""
echo "4. 重置 MySQL 数据库..."
echo "   停止服务..."
docker compose down

echo "   删除旧数据..."
docker volume rm dashgo_mysql_data 2>/dev/null || echo "   (数据卷不存在)"

# 5. 重新启动
echo ""
echo "5. 重新启动服务..."
docker compose up -d --build

# 6. 等待服务启动
echo ""
echo "6. 等待服务启动..."
echo "   (这可能需要 30-60 秒)"

for i in {1..30}; do
    sleep 2
    if docker compose ps | grep -q "Up"; then
        break
    fi
    echo -n "."
done
echo ""

# 7. 检查状态
echo ""
echo "7. 检查服务状态..."
docker compose ps

echo ""
echo "=========================================="
echo "   修复完成！"
echo "=========================================="
echo ""

# 检查是否有不健康的容器
if docker compose ps | grep -q "unhealthy"; then
    echo "⚠️  警告: 仍有容器不健康"
    echo ""
    echo "查看日志:"
    echo "  docker compose logs"
    echo ""
    echo "查看特定服务:"
    echo "  docker compose logs dashgo-mysql"
    echo "  docker compose logs dashgo"
else
    echo "✓ 所有服务运行正常"
    echo ""
    echo "访问地址:"
    IP=$(curl -s4 ip.sb 2>/dev/null || curl -s4 ifconfig.me 2>/dev/null || echo "YOUR_IP")
    echo "  http://${IP}:80"
    echo ""
    echo "默认管理员:"
    echo "  邮箱: admin@dashgo.local"
    echo "  密码: admin123"
fi

echo ""
echo "常用命令:"
echo "  查看状态: docker compose ps"
echo "  查看日志: docker compose logs -f"
echo "  重启服务: docker compose restart"