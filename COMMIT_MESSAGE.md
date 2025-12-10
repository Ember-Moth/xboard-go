# Fix: 修复配置文件字段名称错误

## 问题描述

安装脚本生成的配置文件使用了错误的字段名 `type`，但代码期望的是 `driver`，导致数据库初始化失败，报错：
```
unsupported database driver:
```

## 根本原因

- `internal/config/config.go` 中 `DatabaseConfig` 结构体定义的字段是 `driver`
- `pkg/database/database.go` 读取的也是 `cfg.Driver`
- 但所有安装脚本生成配置时使用的是 `type: "mysql"`
- 导致 `cfg.Driver` 为空字符串，不匹配任何驱动

## 修复内容

### 修复的脚本文件
1. ✅ `install-existing-db.sh` - 修改 `type` 为 `driver`
2. ✅ `local-install.sh` - 修改两处（SQLite 和 MySQL）
3. ✅ `install.sh` - 修改 `type` 为 `driver`
4. ✅ `upgrade.sh` - 修改读取配置的逻辑，从 `grep "type:"` 改为 `grep "driver:"`

### 修复的文档文件
1. ✅ `docs/local-installation.md` - 更新配置示例
2. ✅ `QUICK_INSTALL.md` - 更新配置示例
3. ✅ `UPGRADE_MYSQL.md` - 更新配置示例

### 新增文件
1. ✅ `FIX_CONFIG.md` - 用户修复指南，说明如何修复现有配置文件

## 影响范围

- 所有使用这些脚本安装的用户都会受到影响
- 现有用户需要手动修改配置文件，将 `type` 改为 `driver`
- 新用户使用修复后的脚本将不会遇到此问题

## 测试建议

1. 使用 `install-existing-db.sh` 安装到现有 MySQL 数据库
2. 使用 `local-install.sh` 进行本地开发环境安装（SQLite 和 MySQL 模式）
3. 验证生成的配置文件使用 `driver` 字段
4. 验证数据库连接和迁移正常执行

## 用户操作指南

对于已经遇到此问题的用户：

1. 编辑 `configs/config.yaml`，将 `type: "mysql"` 改为 `driver: "mysql"`
2. 清理失败的迁移记录：`mysql -u root -p your_db -e "DELETE FROM schema_migrations;"`
3. 重新运行迁移：`bash migrate.sh up`

详细步骤请参考 `FIX_CONFIG.md`

## 相关文件

```
modified:   install-existing-db.sh
modified:   local-install.sh
modified:   install.sh
modified:   upgrade.sh
modified:   docs/local-installation.md
modified:   QUICK_INSTALL.md
modified:   UPGRADE_MYSQL.md
new file:   FIX_CONFIG.md
```
