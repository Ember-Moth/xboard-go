#!/bin/bash

# 快速修复配置文件问题

INSTALL_DIR="/opt/dashgo"

echo "=== 修复 dashGO 配置文件 ==="
echo ""

# 检查安装目录
if [ ! -d "$INSTALL_DIR" ]; then
    echo "错误: 安装目录不存在: $INSTALL_DIR"
    exit 1
fi

cd "$INSTALL_DIR"

# 创建 configs 目录
mkdir -p configs

# 生成随机密码
DB_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
REDIS_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
JWT_SECRET=$(openssl rand -base64 32)
NODE_TOKEN=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)

# 创建配置文件
cat > configs/config.yaml << EOF
# dashGO Configuration for Docker
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

# 更新 .env 文件
cat > .env << EOF
MYSQL_ROOT_PASSWORD=${DB_PASS}
MYSQL_DATABASE=dashgo
REDIS_PASSWORD=${REDIS_PASS}
JWT_SECRET=${JWT_SECRET}
NODE_TOKEN=${NODE_TOKEN}
EOF

echo "✓ 配置文件已创建"
echo ""
echo "重要信息:"
echo "  数据库密码: ${DB_PASS}"
echo "  Redis 密码: ${REDIS_PASS}"
echo "  节点 Token: ${NODE_TOKEN}"
echo ""
echo "现在重启服务:"
echo "  cd $INSTALL_DIR"
echo "  docker compose down"
echo "  docker compose up -d --build"
echo ""
echo "注意: 如果磁盘空间不足，请先清理磁盘空间！"
echo "  检查磁盘空间: df -h"
echo "  清理 Docker: docker system prune -a"