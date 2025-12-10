#!/bin/bash

# XBoard MySQL 数据库升级脚本
# 交互式引导升级，保留所有数据
# 用法: bash upgrade-mysql.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${CONFIG_FILE:-configs/config.yaml}"
BACKUP_DIR="backups"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_success() { echo -e "${PURPLE}[SUCCESS]${NC} $1"; }
log_step() { echo -e "${CYAN}[STEP]${NC} $1"; }

# 显示 Banner
show_banner() {
    clear
    echo -e "${CYAN}"
    cat << 'EOF'
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   ██╗  ██╗██████╗  ██████╗  █████╗ ██████╗ ██████╗      ║
║   ╚██╗██╔╝██╔══██╗██╔═══██╗██╔══██╗██╔══██╗██╔══██╗     ║
║    ╚███╔╝ ██████╔╝██║   ██║███████║██████╔╝██║  ██║     ║
║    ██╔██╗ ██╔══██╗██║   ██║██╔══██║██╔══██╗██║  ██║     ║
║   ██╔╝ ██╗██████╔╝╚██████╔╝██║  ██║██║  ██║██████╔╝     ║
║   ╚═╝  ╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝      ║
║                                                           ║
║           MySQL 数据库升级向导                            ║
║           安全升级 · 保留数据 · 交互引导                  ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# 读取数据库配置
read_db_config() {
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "配置文件不存在: $CONFIG_FILE"
        echo ""
        read -p "请输入配置文件路径 [configs/config.yaml]: " custom_config
        CONFIG_FILE="${custom_config:-configs/config.yaml}"
        
        if [ ! -f "$CONFIG_FILE" ]; then
            log_error "配置文件仍然不存在，退出"
            exit 1
        fi
    fi
    
    DB_HOST=$(grep "host:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "host:" | awk '{print $2}' | tr -d '"' | head -1)
    DB_PORT=$(grep "port:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "port:" | awk '{print $2}' | tr -d '"' | head -1)
    DB_USER=$(grep "username:" "$CONFIG_FILE" | awk '{print $2}' | tr -d '"' | head -1)
    DB_PASS=$(grep "password:" "$CONFIG_FILE" | grep -A 5 "database:" | grep "password:" | awk '{print $2}' | tr -d '"' | head -1)
    DB_NAME=$(grep "database:" "$CONFIG_FILE" | grep -A 5 "database:" | tail -1 | awk '{print $2}' | tr -d '"')
    
    # 默认值
    DB_HOST="${DB_HOST:-localhost}"
    DB_PORT="${DB_PORT:-3306}"
    DB_USER="${DB_USER:-root}"
    DB_NAME="${DB_NAME:-xboard}"
}

# 测试数据库连接
test_db_connection() {
    log_step "测试数据库连接..."
    
    if ! command -v mysql &>/dev/null; then
        log_error "未检测到 mysql 命令"
        log_info "请先安装 MySQL 客户端"
        echo ""
        echo "Ubuntu/Debian: sudo apt-get install mysql-client"
        echo "CentOS/RHEL:   sudo yum install mysql"
        echo "macOS:         brew install mysql-client"
        exit 1
    fi
    
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" -e "USE $DB_NAME;" 2>/dev/null; then
        log_success "数据库连接成功"
        return 0
    else
        log_error "数据库连接失败"
        echo ""
        echo "请检查以下信息："
        echo "  主机: $DB_HOST"
        echo "  端口: $DB_PORT"
        echo "  用户: $DB_USER"
        echo "  数据库: $DB_NAME"
        echo ""
        read -p "是否重新输入数据库信息? [y/N]: " retry
        
        if [ "$retry" = "y" ] || [ "$retry" = "Y" ]; then
            input_db_config
            test_db_connection
        else
            exit 1
        fi
    fi
}

# 手动输入数据库配置
input_db_config() {
    echo ""
    log_info "请输入数据库连接信息："
    
    read -p "主机地址 [$DB_HOST]: " input_host
    DB_HOST="${input_host:-$DB_HOST}"
    
    read -p "端口 [$DB_PORT]: " input_port
    DB_PORT="${input_port:-$DB_PORT}"
    
    read -p "用户名 [$DB_USER]: " input_user
    DB_USER="${input_user:-$DB_USER}"
    
    read -sp "密码: " input_pass
    echo ""
    DB_PASS="${input_pass:-$DB_PASS}"
    
    read -p "数据库名 [$DB_NAME]: " input_name
    DB_NAME="${input_name:-$DB_NAME}"
}

# 显示当前数据统计
show_current_data() {
    log_step "获取当前数据统计..."
    
    local user_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_user;" 2>/dev/null || echo "0")
    local plan_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_plan;" 2>/dev/null || echo "0")
    local order_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_order;" 2>/dev/null || echo "0")
    local server_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_server;" 2>/dev/null || echo "0")
    
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║        当前数据库统计                  ║"
    echo "╠════════════════════════════════════════╣"
    echo "║  用户数:   $(printf '%4s' $user_count) 个                        ║"
    echo "║  套餐数:   $(printf '%4s' $plan_count) 个                        ║"
    echo "║  订单数:   $(printf '%4s' $order_count) 个                        ║"
    echo "║  节点数:   $(printf '%4s' $server_count) 个                        ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    # 保存数据
    mkdir -p "$BACKUP_DIR"
    cat > "$BACKUP_DIR/data_before_upgrade.txt" << EOF
升级前数据统计
时间: $(date)
用户数: $user_count
套餐数: $plan_count
订单数: $order_count
节点数: $server_count
EOF
}

# 备份数据库
backup_database() {
    log_step "备份数据库..."
    
    mkdir -p "$BACKUP_DIR"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/backup_before_upgrade_${timestamp}.sql"
    
    echo ""
    log_info "备份文件: $backup_file"
    
    if mysqldump -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" > "$backup_file" 2>/dev/null; then
        local size=$(du -h "$backup_file" | cut -f1)
        log_success "备份完成！文件大小: $size"
        echo "$backup_file" > "$BACKUP_DIR/latest_backup.txt"
        BACKUP_FILE="$backup_file"
    else
        log_error "备份失败"
        exit 1
    fi
}

# 显示待执行的迁移
show_pending_migrations() {
    log_step "检查待执行的迁移..."
    echo ""
    
    bash migrate.sh status
    
    echo ""
}

# 执行升级
run_upgrade() {
    log_step "执行数据库升级..."
    echo ""
    
    bash migrate.sh up
    
    echo ""
    log_success "数据库升级完成！"
}

# 验证升级结果
verify_upgrade() {
    log_step "验证升级结果..."
    
    local user_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_user;" 2>/dev/null || echo "0")
    local plan_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_plan;" 2>/dev/null || echo "0")
    local order_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_order;" 2>/dev/null || echo "0")
    local group_count=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM v2_user_group;" 2>/dev/null || echo "0")
    
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║        升级后数据统计                  ║"
    echo "╠════════════════════════════════════════╣"
    echo "║  用户数:     $(printf '%4s' $user_count) 个                      ║"
    echo "║  套餐数:     $(printf '%4s' $plan_count) 个                      ║"
    echo "║  订单数:     $(printf '%4s' $order_count) 个                      ║"
    echo "║  用户组数:   $(printf '%4s' $group_count) 个 (新增)              ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    # 保存数据
    cat > "$BACKUP_DIR/data_after_upgrade.txt" << EOF
升级后数据统计
时间: $(date)
用户数: $user_count
套餐数: $plan_count
订单数: $order_count
用户组数: $group_count
EOF
}

# 配置用户组
configure_user_groups() {
    log_step "配置用户组..."
    echo ""
    
    # 获取节点列表
    log_info "获取节点列表..."
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SELECT id, name, host FROM v2_server;" 2>/dev/null || true
    
    echo ""
    read -p "请输入普通用户组可访问的节点ID（逗号分隔，如: 1,2,3）: " normal_servers
    
    if [ -n "$normal_servers" ]; then
        # 转换为 JSON 数组格式
        local server_json="[${normal_servers}]"
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "UPDATE v2_user_group SET server_ids = '$server_json' WHERE id = 2;" 2>/dev/null
        log_success "普通用户组节点配置完成"
    fi
    
    echo ""
    read -p "请输入VIP用户组可访问的节点ID（逗号分隔，如: 1,2,3,4,5）: " vip_servers
    
    if [ -n "$vip_servers" ]; then
        local server_json="[${vip_servers}]"
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "UPDATE v2_user_group SET server_ids = '$server_json' WHERE id = 3;" 2>/dev/null
        log_success "VIP用户组节点配置完成"
    fi
    
    echo ""
    # 获取套餐列表
    log_info "获取套餐列表..."
    mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "SELECT id, name, month_price FROM v2_plan;" 2>/dev/null || true
    
    echo ""
    read -p "请输入普通用户组可购买的套餐ID（逗号分隔，如: 1,2,3）: " normal_plans
    
    if [ -n "$normal_plans" ]; then
        local plan_json="[${normal_plans}]"
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "UPDATE v2_user_group SET plan_ids = '$plan_json' WHERE id = 2;" 2>/dev/null
        log_success "普通用户组套餐配置完成"
    fi
    
    echo ""
    read -p "请输入VIP用户组可购买的套餐ID（逗号分隔，如: 10,11,12）: " vip_plans
    
    if [ -n "$vip_plans" ]; then
        local plan_json="[${vip_plans}]"
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "UPDATE v2_user_group SET plan_ids = '$plan_json' WHERE id = 3;" 2>/dev/null
        log_success "VIP用户组套餐配置完成"
    fi
    
    echo ""
    read -p "是否配置套餐的升级组? [Y/n]: " config_upgrade
    
    if [ "$config_upgrade" != "n" ] && [ "$config_upgrade" != "N" ]; then
        echo ""
        log_info "配置套餐升级组..."
        echo "  ID 1 = 试用用户"
        echo "  ID 2 = 普通用户"
        echo "  ID 3 = VIP用户"
        echo ""
        
        read -p "请输入要配置的套餐ID（逗号分隔）: " plan_ids
        read -p "购买这些套餐后升级到哪个组? [2]: " target_group
        target_group="${target_group:-2}"
        
        if [ -n "$plan_ids" ]; then
            IFS=',' read -ra PLAN_ARRAY <<< "$plan_ids"
            for plan_id in "${PLAN_ARRAY[@]}"; do
                mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "UPDATE v2_plan SET upgrade_group_id = $target_group WHERE id = $plan_id;" 2>/dev/null
            done
            log_success "套餐升级组配置完成"
        fi
    fi
}

# 显示最终结果
show_final_result() {
    clear
    echo -e "${GREEN}"
    cat << 'EOF'
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║                  ✓ 升级完成！                             ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    echo "备份文件位置:"
    echo "  $BACKUP_FILE"
    echo ""
    
    if [ -f "$BACKUP_DIR/data_before_upgrade.txt" ] && [ -f "$BACKUP_DIR/data_after_upgrade.txt" ]; then
        echo "数据对比:"
        echo ""
        echo "升级前:"
        cat "$BACKUP_DIR/data_before_upgrade.txt" | grep -E "用户数|套餐数|订单数"
        echo ""
        echo "升级后:"
        cat "$BACKUP_DIR/data_after_upgrade.txt" | grep -E "用户数|套餐数|订单数|用户组数"
    fi
    
    echo ""
    echo "下一步操作:"
    echo "  1. 重启服务: docker compose restart"
    echo "  2. 查看日志: docker compose logs -f"
    echo "  3. 测试功能: 访问后台管理"
    echo ""
    
    echo -e "${YELLOW}如果遇到问题，可以使用备份恢复:${NC}"
    echo "  mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p $DB_NAME < $BACKUP_FILE"
    echo ""
    
    echo "详细文档:"
    echo "  用户组设计: docs/user-group-design.md"
    echo "  升级指南:   docs/upgrade-guide.md"
    echo ""
}

# 主函数
main() {
    show_banner
    
    cd "$SCRIPT_DIR"
    
    # 步骤 1: 读取配置
    log_step "步骤 1/7: 读取数据库配置"
    read_db_config
    
    echo ""
    echo "数据库信息:"
    echo "  主机: $DB_HOST"
    echo "  端口: $DB_PORT"
    echo "  用户: $DB_USER"
    echo "  数据库: $DB_NAME"
    echo ""
    
    read -p "配置是否正确? [Y/n]: " config_ok
    if [ "$config_ok" = "n" ] || [ "$config_ok" = "N" ]; then
        input_db_config
    fi
    
    # 步骤 2: 测试连接
    echo ""
    log_step "步骤 2/7: 测试数据库连接"
    test_db_connection
    
    # 步骤 3: 显示当前数据
    echo ""
    log_step "步骤 3/7: 获取当前数据"
    show_current_data
    
    read -p "按 Enter 继续..."
    
    # 步骤 4: 备份数据库
    echo ""
    log_step "步骤 4/7: 备份数据库"
    backup_database
    
    echo ""
    read -p "按 Enter 继续..."
    
    # 步骤 5: 显示待执行的迁移
    echo ""
    log_step "步骤 5/7: 检查待执行的迁移"
    show_pending_migrations
    
    read -p "确认执行以上迁移? [Y/n]: " confirm
    if [ "$confirm" = "n" ] || [ "$confirm" = "N" ]; then
        log_info "已取消升级"
        exit 0
    fi
    
    # 步骤 6: 执行升级
    echo ""
    log_step "步骤 6/7: 执行升级"
    run_upgrade
    
    # 验证结果
    verify_upgrade
    
    echo ""
    read -p "按 Enter 继续..."
    
    # 步骤 7: 配置用户组
    echo ""
    log_step "步骤 7/7: 配置用户组"
    echo ""
    read -p "是否现在配置用户组? [Y/n]: " config_now
    
    if [ "$config_now" != "n" ] && [ "$config_now" != "N" ]; then
        configure_user_groups
    else
        log_info "跳过配置，稍后可以手动配置"
        echo ""
        echo "手动配置命令:"
        echo "  mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p $DB_NAME"
        echo "  UPDATE v2_user_group SET server_ids = '[1,2,3]' WHERE id = 2;"
        echo "  UPDATE v2_user_group SET plan_ids = '[1,2,3]' WHERE id = 2;"
    fi
    
    # 显示最终结果
    echo ""
    read -p "按 Enter 查看升级结果..."
    show_final_result
}

main "$@"
