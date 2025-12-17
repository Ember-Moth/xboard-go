#!/bin/bash

# dashGO 安全功能启用脚本
# 用于现有安装启用新的安全功能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_success() { echo -e "${BLUE}[SUCCESS]${NC} $1"; }

INSTALL_DIR="/opt/dashgo"

# 检查安装目录
check_installation() {
    if [ ! -d "$INSTALL_DIR" ]; then
        log_error "dashGO 未安装在 $INSTALL_DIR"
        exit 1
    fi
    
    if [ ! -f "$INSTALL_DIR/docker-compose.yaml" ]; then
        log_error "未找到 docker-compose.yaml 文件"
        exit 1
    fi
    
    log_info "检测到 dashGO 安装目录: $INSTALL_DIR"
}

# 备份现有配置
backup_config() {
    log_info "备份现有配置..."
    
    cd "$INSTALL_DIR"
    
    # 创建备份目录
    backup_dir="backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # 备份配置文件
    cp configs/config.yaml "$backup_dir/" 2>/dev/null || cp config.yaml "$backup_dir/" 2>/dev/null || true
    cp docker-compose.yaml "$backup_dir/" 2>/dev/null || true
    cp .env "$backup_dir/" 2>/dev/null || true
    
    log_success "配置已备份到: $backup_dir"
}

# 运行数据库迁移
run_migrations() {
    log_info "运行数据库迁移..."
    
    cd "$INSTALL_DIR"
    
    # 检查是否有迁移文件
    if [ ! -d "migrations" ]; then
        log_warn "未找到迁移目录，跳过数据库迁移"
        return 0
    fi
    
    # 停止服务
    docker compose down
    
    # 运行迁移
    if [ -f "migrate-linux-amd64" ]; then
        log_info "使用内置迁移工具..."
        ./migrate-linux-amd64 -path migrations -database "sqlite3://data/dashgo.db" up
    else
        log_info "使用 Docker 运行迁移..."
        docker run --rm -v "$PWD/migrations:/migrations" -v "$PWD/data:/data" \
            migrate/migrate -path=/migrations -database="sqlite3:///data/dashgo.db" up
    fi
    
    log_success "数据库迁移完成"
}

# 更新配置文件
update_config() {
    log_info "更新配置文件..."
    
    cd "$INSTALL_DIR"
    
    # 检查配置文件位置
    config_file=""
    if [ -f "configs/config.yaml" ]; then
        config_file="configs/config.yaml"
    elif [ -f "config.yaml" ]; then
        config_file="config.yaml"
    else
        log_error "未找到配置文件"
        exit 1
    fi
    
    # 添加安全配置（如果不存在）
    if ! grep -q "security:" "$config_file"; then
        log_info "添加安全配置..."
        cat >> "$config_file" << 'EOF'

# 安全配置
security:
  # 启用安全日志记录
  enable_logging: true
  
  # 认证失败保护
  auth_failure_protection:
    max_attempts: 5
    lockout_duration: 300  # 5分钟
    progressive_delay: true
  
  # 输入验证
  input_validation:
    enable_xss_protection: true
    enable_sql_injection_protection: true
    enable_path_traversal_protection: true
    max_input_length: 10000
  
  # 速率限制
  rate_limiting:
    enable: true
    requests_per_minute: 60
    burst_size: 10
  
  # 安全警报
  security_alerts:
    enable: true
    email_notifications: false
    telegram_notifications: false
EOF
        log_success "安全配置已添加"
    else
        log_info "安全配置已存在，跳过"
    fi
}

# 更新 Docker Compose 配置
update_docker_compose() {
    log_info "检查 Docker Compose 配置..."
    
    cd "$INSTALL_DIR"
    
    # 检查是否需要添加健康检查
    if ! grep -q "healthcheck:" docker-compose.yaml; then
        log_info "添加健康检查配置..."
        
        # 创建临时文件
        cat > docker-compose.tmp.yaml << 'EOF'
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
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

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
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    container_name: dashgo-nginx
    ports:
      - "${WEB_PORT}:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - dashgo
    restart: unless-stopped
    networks:
      - dashgo-net
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  redis_data:

networks:
  dashgo-net:
    driver: bridge
EOF
        
        # 备份原文件并替换
        cp docker-compose.yaml docker-compose.yaml.bak
        mv docker-compose.tmp.yaml docker-compose.yaml
        
        log_success "Docker Compose 配置已更新"
    else
        log_info "Docker Compose 配置已是最新版本"
    fi
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    
    cd "$INSTALL_DIR"
    
    # 重新构建并启动服务
    docker compose up -d --build
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 15
    
    # 检查服务状态
    if docker compose ps | grep -q "Up"; then
        log_success "服务启动成功"
    else
        log_error "服务启动失败，查看日志:"
        docker compose logs --tail=20
        exit 1
    fi
}

# 验证安全功能
verify_security_features() {
    log_info "验证安全功能..."
    
    cd "$INSTALL_DIR"
    
    # 检查安全表是否创建
    if [ -f "data/dashgo.db" ]; then
        log_info "检查安全数据表..."
        
        # 使用 sqlite3 检查表
        if command -v sqlite3 >/dev/null 2>&1; then
            tables=$(sqlite3 data/dashgo.db ".tables" | grep -E "v2_security|v2_auth_failure")
            if [ -n "$tables" ]; then
                log_success "安全数据表已创建: $tables"
            else
                log_warn "未找到安全数据表，可能需要手动运行迁移"
            fi
        else
            log_info "sqlite3 未安装，跳过表检查"
        fi
    fi
    
    # 检查服务健康状态
    log_info "检查服务健康状态..."
    
    # 等待服务完全启动
    sleep 5
    
    # 检查健康端点
    if curl -f -s http://localhost:8080/health >/dev/null 2>&1; then
        log_success "健康检查端点正常"
    else
        log_warn "健康检查端点不可用，服务可能仍在启动中"
    fi
    
    # 检查日志中是否有错误
    log_info "检查服务日志..."
    error_count=$(docker compose logs dashgo 2>&1 | grep -i error | wc -l)
    if [ "$error_count" -eq 0 ]; then
        log_success "服务日志中无错误"
    else
        log_warn "服务日志中发现 $error_count 个错误，请检查日志"
    fi
}

# 显示完成信息
show_completion_info() {
    log_success "安全功能启用完成！"
    echo ""
    echo "新增的安全功能:"
    echo "  ✓ 安全事件日志记录"
    echo "  ✓ 认证失败保护（渐进延迟）"
    echo "  ✓ 实时安全警报"
    echo "  ✓ 输入验证和清理"
    echo "  ✓ SQL注入防护"
    echo "  ✓ XSS攻击防护"
    echo "  ✓ 路径遍历防护"
    echo "  ✓ 网络弹性和重试机制"
    echo "  ✓ 数据库连接弹性"
    echo "  ✓ 优雅降级"
    echo ""
    echo "管理命令:"
    echo "  查看服务状态: cd $INSTALL_DIR && docker compose ps"
    echo "  查看安全日志: cd $INSTALL_DIR && docker compose logs dashgo | grep -i security"
    echo "  重启服务: cd $INSTALL_DIR && docker compose restart"
    echo ""
    echo "配置文件位置:"
    echo "  主配置: $INSTALL_DIR/configs/config.yaml"
    echo "  Docker: $INSTALL_DIR/docker-compose.yaml"
    echo ""
    log_warn "建议定期检查安全日志并根据需要调整安全配置"
}

# 主函数
main() {
    echo "dashGO 安全功能启用脚本"
    echo "=========================="
    echo ""
    
    # 检查 root 权限
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 用户运行此脚本"
        exit 1
    fi
    
    check_installation
    backup_config
    run_migrations
    update_config
    update_docker_compose
    restart_services
    verify_security_features
    show_completion_info
}

main "$@"