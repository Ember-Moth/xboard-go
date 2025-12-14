#!/bin/bash

# 切换到 SQLite 数据库（避免 MySQL 问题）

INSTALL_DIR="/opt/dashgo"

echo "=========================================="
echo "   切换到 SQLite 数据库"
echo "=========================================="
echo ""

if [ ! -d "$INSTALL_DIR" ]; then
    echo "❌ 安装目录不存在: $INSTALL_DIR"
    exit 1
fi

cd "$INSTALL_DIR"

echo "1. 停止服务..."
docker compose down

echo ""
echo "2. 备份当前配置..."
cp configs/config.yaml configs/config.yaml.mysql.bak 2>/dev/null || true
cp docker-compose.yaml docker-compose.yaml.mysql.bak 2>/dev/null || true

echo ""
echo "3. 创建 SQLite 配置..."

# 生成随机密码
REDIS_PASS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
JWT_SECRET=$(openssl rand -base64 32)
NODE_TOKEN=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)

cat > configs/config.yaml << EOF
app:
  name: "dashGO"
  mode: "release"
  listen: ":8080"

database:
  driver: "sqlite"
  dsn: "data/dashgo.db"

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

echo ""
echo "4. 创建简化的 Docker Compose..."

cat > docker-compose.yaml << 'EOF'
services:
  dashgo:
    build: .
    container_name: dashgo
    ports:
      - "8080:8080"
    volumes:
      - ./configs/config.yaml:/app/configs/config.yaml
      - ./data:/app/data
      - ./web/dist:/app/web/dist
    depends_on:
      - redis
    restart: unless-stopped
    environment:
      - TZ=Asia/Shanghai
    networks:
      - dashgo-net

  redis:
    image: redis:7-alpine
    container_name: dashgo-redis
    command: redis-server --requirepass ${REDIS_PASSWORD:-}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped
    networks:
      - dashgo-net

  nginx:
    image: nginx:alpine
    container_name: dashgo-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - dashgo
    restart: unless-stopped
    networks:
      - dashgo-net

volumes:
  redis_data:

networks:
  dashgo-net:
    driver: bridge
EOF

# 更新 .env
cat > .env << EOF
REDIS_PASSWORD=${REDIS_PASS}
JWT_SECRET=${JWT_SECRET}
NODE_TOKEN=${NODE_TOKEN}
EOF

echo ""
echo "5. 创建数据目录..."
mkdir -p data

echo ""
echo "6. 启动服务..."
docker compose up -d --build

echo ""
echo "7. 等待服务启动..."
sleep 10

echo ""
echo "8. 检查状态..."
docker compose ps

echo ""
echo "=========================================="
echo "   切换完成！"
echo "=========================================="
echo ""
echo "✓ 现在使用 SQLite 数据库"
echo "✓ 数据库文件: $INSTALL_DIR/data/dashgo.db"
echo ""
echo "重要信息:"
echo "  Redis 密码: ${REDIS_PASS}"
echo "  节点 Token: ${NODE_TOKEN}"
echo ""
echo "默认管理员:"
echo "  邮箱: admin@example.com"
echo "  密码: admin123456"
echo ""
echo "如果需要恢复 MySQL:"
echo "  cp configs/config.yaml.mysql.bak configs/config.yaml"
echo "  cp docker-compose.yaml.mysql.bak docker-compose.yaml"
echo "  docker compose up -d --build"
echo ""
echo "查看日志:"
echo "  docker compose logs -f"