# XBoard Go 流控和用户组重构方案

## 一、架构澄清

### 1.1 正确的层级关系
```
Host (物理主机)
  └─ Server (sing-box 服务实例，一个协议配置)
       └─ Node (用户看到的节点，IP+Port 变体)
```

### 1.2 用户组设计
- **UserGroup（用户组）**：用户的权限分组（VIP、普通、试用等）
- **Server.group_ids**：该服务器允许哪些用户组访问
- **User.group_id**：用户属于哪个组
- **Plan.group_id**：购买套餐后用户会被分配到哪个组

## 二、流控实现方案

### 2.1 实时流控检查
在 `GetAvailableUsers` 时过滤：
- 检查用户是否超流量
- 检查用户是否过期
- 检查用户是否被封禁

### 2.2 流量上报优化
- Agent 上报流量时，面板立即检查用户状态
- 超流量用户标记为 `banned` 或设置特殊状态
- 下次 Agent 拉取用户列表时自动排除

### 2.3 流量统计改进
- 保留当前的增量上报机制
- 添加流量预警（80%、90%）
- 添加流量重置任务

## 三、用户组重构（核心改动）

### 3.1 新的用户组设计理念
**用户组是核心，节点和套餐都归属于用户组**

```
UserGroup (用户组)
  ├─ Servers (该组可用的节点列表)
  └─ Plans (该组可购买的套餐列表)
  
User (用户)
  └─ group_id -> 决定能访问哪些节点和购买哪些套餐
```

### 3.2 数据关系
```go
// UserGroup: 用户组（VIP、普通、试用等）
// UserGroup.server_ids: 该组可以访问的节点列表（JSON数组）
// UserGroup.plan_ids: 该组可以购买的套餐列表（JSON数组）

// User.group_id: 用户所属的组
// Server: 节点配置（不再有 group_ids）
// Plan: 套餐配置（不再有 group_id）

// 逻辑：
// 1. 用户注册 -> User.group_id = 默认组（如：试用组）
// 2. 用户购买套餐 -> 套餐可以指定升级到哪个组
// 3. 用户访问节点 -> 检查 User.group_id 对应的 UserGroup.server_ids
// 4. 用户查看套餐 -> 检查 User.group_id 对应的 UserGroup.plan_ids
```

### 3.3 用户组管理界面
管理员在用户组管理页面：
- 创建/编辑用户组
- 为用户组分配可用节点（多选）
- 为用户组分配可购买套餐（多选）
- 设置组的默认流量、速度限制等

### 3.4 创建用户时的组分配
- 注册时：`group_id = 1`（默认组，如"试用用户"）
- 购买套餐后：可选择是否升级组（如：购买VIP套餐 -> 升级到VIP组）
- 管理员手动分配：可以直接修改用户的 `group_id`

## 四、Agent Alpine 兼容性

### 4.1 问题分析
Alpine 下 sing-box 的 Clash API 可能不可用，导致：
- 无法获取连接信息
- 无法统计用户流量

### 4.2 解决方案
**方案 A：使用 sing-box 的 stats API（推荐）**
```json
{
  "experimental": {
    "cache_file": {
      "enabled": true
    }
  }
}
```

**方案 B：使用端口流量平均分配（备用）**
- 获取总流量
- 按用户数平均分配
- 不够精确但可用

**方案 C：使用 iptables 统计（最精确）**
- 为每个用户创建 iptables 规则
- 定期读取流量统计
- 需要 root 权限

## 五、实施步骤

### 5.1 第一阶段：流控实现（高优先级）
1. 修改 `GetAvailableUsers` 添加流量检查
2. 修改 `TrafficFetch` 添加超限检查
3. 添加流量预警功能
4. 测试流控是否生效

### 5.2 第二阶段：用户组优化（中优先级）
1. 明确用户组字段含义
2. 修改创建用户逻辑
3. 修改订单完成逻辑（分配用户组）
4. 更新前端界面

### 5.3 第三阶段：Agent 优化（中优先级）
1. 添加多种流量统计方式
2. 自动检测可用的统计方式
3. 添加 Alpine 特殊处理
4. 测试 Alpine 兼容性

### 5.4 第四阶段：Node 架构（低优先级）
1. 创建 Node 表（如果需要）
2. Node 指向 Server
3. 用户订阅时返回 Node 列表
4. 支持同一 Server 多个 Node（不同 IP）

## 六、数据库变更

### 6.1 需要的表结构调整（新架构）
```sql
-- 1. 创建用户组表（核心表）
CREATE TABLE IF NOT EXISTS v2_user_group (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(64) NOT NULL COMMENT '组名称',
    description TEXT COMMENT '组描述',
    server_ids JSON COMMENT '该组可访问的节点ID列表',
    plan_ids JSON COMMENT '该组可购买的套餐ID列表',
    default_transfer_enable BIGINT DEFAULT 0 COMMENT '默认流量（字节）',
    default_speed_limit INT COMMENT '默认速度限制（Mbps）',
    default_device_limit INT COMMENT '默认设备数限制',
    sort INT DEFAULT 0 COMMENT '排序',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
) COMMENT='用户权限组（核心表）';

-- 2. 插入默认组
INSERT INTO v2_user_group (id, name, description, server_ids, plan_ids, default_transfer_enable, sort, created_at, updated_at) VALUES
(1, '试用用户', '新注册用户默认组，流量较少', '[]', '[]', 1073741824, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2, '普通用户', '购买基础套餐的用户', '[]', '[]', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(3, 'VIP用户', '购买高级套餐的用户', '[]', '[]', 0, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 3. 用户表：明确 group_id 含义
ALTER TABLE v2_user MODIFY COLUMN group_id BIGINT DEFAULT 1 COMMENT '用户所属权限组ID（关联v2_user_group）';

-- 4. 服务器表：移除 group_ids（不再需要）
-- ALTER TABLE v2_server DROP COLUMN group_ids; -- 保留以兼容旧数据，但不再使用

-- 5. 套餐表：添加 upgrade_group_id
ALTER TABLE v2_plan ADD COLUMN upgrade_group_id BIGINT COMMENT '购买后升级到的用户组ID（可选）';
-- ALTER TABLE v2_plan DROP COLUMN group_id; -- 保留以兼容旧数据，但不再使用

-- 6. 为现有用户设置默认组
UPDATE v2_user SET group_id = 1 WHERE group_id IS NULL OR group_id = 0;
```

### 6.2 Node 表（如果需要独立节点）
```sql
CREATE TABLE IF NOT EXISTS v2_node (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    server_id BIGINT NOT NULL COMMENT '关联的Server ID',
    name VARCHAR(128) NOT NULL COMMENT '节点名称',
    host VARCHAR(255) NOT NULL COMMENT '节点IP或域名',
    port INT NOT NULL COMMENT '节点端口',
    show BOOLEAN DEFAULT TRUE COMMENT '是否显示',
    sort INT DEFAULT 0 COMMENT '排序',
    tags JSON COMMENT '标签',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    INDEX idx_server_id (server_id)
) COMMENT='用户可见节点（指向Server）';
```

## 七、配置示例

### 7.1 用户组配置示例（新架构）
```yaml
# 1. 创建用户组
user_group:
  name: "VIP用户"
  server_ids: [1, 2, 3, 5]  # VIP可以访问这些节点
  plan_ids: [10, 11, 12]     # VIP可以购买这些套餐
  default_speed_limit: 1000  # 默认1Gbps
  
# 2. 创建服务器（不再需要指定组）
server:
  name: "香港节点"
  type: "shadowsocks"
  # 不再有 group_ids 字段
  
# 3. 创建套餐
plan:
  name: "VIP月付套餐"
  upgrade_group_id: 3  # 购买后升级到VIP组（可选）
  
# 4. 用户购买套餐后
user:
  group_id: 3  # 自动升级到VIP组
  # 现在可以访问 server_ids: [1,2,3,5]
  # 现在可以购买 plan_ids: [10,11,12]
```

### 7.2 流控配置示例
```yaml
# 配置文件添加
traffic:
  check_interval: 60  # 每60秒检查一次
  warning_threshold: 80  # 80%时预警
  auto_ban: true  # 超流量自动封禁
```

## 八、测试计划

### 8.1 流控测试
- [ ] 创建测试用户，设置小流量限制
- [ ] 使用节点产生流量
- [ ] 验证流量统计是否准确
- [ ] 验证超流量后是否自动断开
- [ ] 验证流量重置是否正常

### 8.2 用户组测试
- [ ] 创建不同组的用户
- [ ] 创建限制组的服务器
- [ ] 验证用户只能访问允许的服务器
- [ ] 验证购买套餐后组分配正确

### 8.3 Alpine 测试
- [ ] 在 Alpine 容器中部署 Agent
- [ ] 验证流量统计是否正常
- [ ] 验证用户同步是否正常
- [ ] 验证配置更新是否正常

## 九、向后兼容

### 9.1 保持兼容性
- 保留 ServerNode 表（向后兼容）
- Server.group_ids 为空时允许所有用户
- User.group_id 为 NULL 时视为无权限用户

### 9.2 迁移脚本
提供数据迁移脚本，将现有数据迁移到新结构。
