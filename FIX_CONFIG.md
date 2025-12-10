# 修复配置文件问题

## 问题说明

您遇到的错误 "unsupported database driver" 是因为配置文件使用了错误的字段名。

- **错误**: `type: "mysql"`
- **正确**: `driver: "mysql"`

## 修复步骤

### 1. 修改配置文件

编辑您的配置文件 `configs/config.yaml`，将：

```yaml
database:
  type: "mysql"
  host: "localhost"
  port: 3306
  database: "akari"  # 您的数据库名
  username: "root"
  password: "your_password"
```

改为：

```yaml
database:
  driver: "mysql"  # 改 type 为 driver
  host: "localhost"
  port: 3306
  database: "akari"  # 您的数据库名
  username: "root"
  password: "your_password"
```

### 2. 清理失败的迁移记录

由于之前的迁移失败，需要清理迁移记录：

```bash
# 连接到 MySQL
mysql -h localhost -u root -p

# 选择数据库
USE akari;

# 删除失败的迁移记录
DELETE FROM schema_migrations;

# 退出
EXIT;
```

### 3. 重新运行迁移

```bash
# 使用安装脚本（推荐）
bash install-existing-db.sh

# 或者手动运行迁移
bash migrate.sh up
```

### 4. 验证

```bash
# 检查迁移状态
bash migrate.sh status

# 应该看到所有迁移都已应用
```

## 快速修复命令

如果您想一键修复，可以运行：

```bash
# 1. 修改配置文件（假设配置文件在 configs/config.yaml）
sed -i 's/type: "mysql"/driver: "mysql"/g' configs/config.yaml

# 2. 清理迁移记录
mysql -h localhost -u root -p akari -e "DELETE FROM schema_migrations;"

# 3. 重新运行迁移
bash migrate.sh up
```

## 注意事项

1. **数据库名称**: 确保配置文件中的 `database` 字段与您实际的数据库名称一致（您的是 `akari`）
2. **备份**: 在执行任何操作前，建议先备份数据库
3. **权限**: 确保数据库用户有足够的权限创建表和修改结构

## 已修复的文件

以下脚本和文档已经修复，现在都使用正确的 `driver` 字段：

- ✅ `install-existing-db.sh`
- ✅ `local-install.sh`
- ✅ `install.sh`
- ✅ `upgrade.sh`
- ✅ `docs/local-installation.md`
- ✅ `QUICK_INSTALL.md`
- ✅ `UPGRADE_MYSQL.md`

## 如果还有问题

如果修复后仍然有问题，请检查：

1. 配置文件路径是否正确
2. 数据库连接信息是否正确
3. 数据库是否已创建
4. 用户权限是否足够

可以使用以下命令测试数据库连接：

```bash
mysql -h localhost -u root -p akari -e "SELECT 1;"
```
