# XBoard Go 快速开始指南

## 1. 数据库迁移

```bash
# 执行用户组迁移
mysql -u root -p xboard < migrations/003_create_user_group.sql

# 验证迁移
mysql -u root -p xboard -e "SELECT * FROM v2_user_group;"
```

应该看到3个默认用户组：
- ID 1: 试用用户
- ID 2: 普通用户  
- ID 3: VIP用户

## 2. 配置用户组

### 方式1：通过 SQL（临时方案）

```sql
-- 为"普通用户"组添加可访问的节点
UPDATE v2_user_group 
SET server_ids = '[1, 2, 3]'  -- 节点ID列表
WHERE id = 2;

-- 为"普通用户"组添加可购买的套餐
UPDATE v2_user_group 
SET plan_ids = '[1, 2, 3]'    -- 套餐ID列表
WHERE id = 2;

-- 为"VIP用户"组添加所有节点
UPDATE v2_user_group 
SET server_ids = '[1, 2, 3, 4, 5]'
WHERE id = 3;

-- 为"VIP用户"组添加高级套餐
UPDATE v2_user_group 
SET plan_ids = '[10, 11, 12]'
WHERE id = 3;
```

### 方式2：通过 API（推荐，需要先实现 Handler）

```bash
# 创建用户组
curl -X POST http://localhost:8080/api/v2/admin/user-group \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "高级VIP",
    "description": "高级用户组",
    "server_ids": [1,2,3,4,5],
    "plan_ids": [10,11,12],
    "default_transfer_enable": 107374182400
  }'

# 为用户组添加节点
curl -X POST http://localhost:8080/api/v2/admin/user-group/2/servers \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"server_ids": [1,2,3]}'
```

## 3. 配置套餐升级

```sql
-- 设置套餐购买后升级到的用户组
UPDATE v2_plan 
SET upgrade_group_id = 2  -- 购买后升级到"普通用户"组
WHERE id = 1;

UPDATE v2_plan 
SET upgrade_group_id = 3  -- 购买后升级到"VIP用户"组
WHERE id = 10;
```

## 4. 测试流控

### 创建测试用户

```sql
-- 创建一个测试用户，设置小流量
INSERT INTO v2_user (
  email, password, uuid, token, group_id, 
  transfer_enable, u, d, 
  created_at, updated_at
) VALUES (
  'test@example.com',
  '$2a$10$...', -- 密码哈希
  'xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx',
  'test_token_123',
  2,  -- 普通用户组
  10485760,  -- 10MB 流量（用于测试）
  0, 0,
  UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
);
```

### 测试流程

1. **查看用户可访问的节点**
   ```bash
   curl http://localhost:8080/api/v1/user/subscribe \
     -H "Authorization: Bearer test_token_123"
   ```
   应该只返回用户组配置的节点

2. **使用节点产生流量**
   - 连接到节点
   - 下载一些数据（超过 10MB）

3. **等待 Agent 上报流量**
   - Agent 每60秒上报一次
   - 查看日志：`tail -f /var/log/xboard-agent.log`

4. **验证流控生效**
   ```sql
   -- 查看用户流量
   SELECT email, u, d, transfer_enable, 
          (u + d) as used,
          (u + d) >= transfer_enable as is_over
   FROM v2_user 
   WHERE email = 'test@example.com';
   ```

5. **验证自动断开**
   - 超流量后，再次获取订阅
   - 应该看不到该用户（或无法连接）

## 5. 部署 Agent（Alpine）

```bash
# 在 Alpine 服务器上执行
curl -sL https://your-repo/agent/install.sh | sh -s -- \
  https://your-panel.com \
  your_host_token

# 查看日志
tail -f /var/log/xboard-agent.log
tail -f /var/log/xboard-agent.err

# 管理服务
rc-service xboard-agent status
rc-service xboard-agent restart
```

## 6. 监控流量

### 查看总流量统计

```sql
SELECT 
  COUNT(*) as total_users,
  SUM(u + d) as total_traffic,
  SUM(u + d) / 1024 / 1024 / 1024 as total_gb,
  AVG(u + d) as avg_traffic
FROM v2_user
WHERE group_id IS NOT NULL;
```

### 查看超流量用户

```sql
SELECT 
  email,
  (u + d) as used,
  transfer_enable as total,
  (u + d) * 100.0 / transfer_enable as usage_percent
FROM v2_user
WHERE transfer_enable > 0 
  AND (u + d) >= transfer_enable
ORDER BY usage_percent DESC;
```

### 查看流量预警用户（80%以上）

```sql
SELECT 
  email,
  (u + d) as used,
  transfer_enable as total,
  (u + d) * 100.0 / transfer_enable as usage_percent
FROM v2_user
WHERE transfer_enable > 0 
  AND (u + d) * 100.0 / transfer_enable >= 80
  AND (u + d) < transfer_enable
ORDER BY usage_percent DESC;
```

## 7. 常见问题

### Q1: 用户看不到任何节点？
**A:** 检查用户组配置
```sql
-- 查看用户所在组
SELECT u.email, u.group_id, g.name, g.server_ids
FROM v2_user u
LEFT JOIN v2_user_group g ON u.group_id = g.id
WHERE u.email = 'user@example.com';

-- 如果 server_ids 为空，添加节点
UPDATE v2_user_group 
SET server_ids = '[1,2,3]'
WHERE id = 用户的group_id;
```

### Q2: 流量统计不准确？
**A:** 这是正常的，当前采用平均分配算法
- 查看 `docs/traffic-limitation.md` 了解详情
- 建议设置较大的流量包
- 采用宽松流控策略

### Q3: Agent 无法启动？
**A:** 查看错误日志
```bash
# Alpine
tail -f /var/log/xboard-agent.err

# 其他系统
journalctl -u xboard-agent -f
```

常见原因：
- 面板地址错误
- Token 错误
- 网络不通
- sing-box 未安装

### Q4: 用户购买套餐后没有升级组？
**A:** 检查套餐配置
```sql
-- 查看套餐的升级组配置
SELECT id, name, upgrade_group_id 
FROM v2_plan;

-- 设置升级组
UPDATE v2_plan 
SET upgrade_group_id = 3  -- VIP组
WHERE id = 套餐ID;
```

然后在订单完成逻辑中添加升级代码（需要修改 `internal/service/order.go`）

### Q5: 如何重置所有用户流量？
**A:** 执行 SQL
```sql
-- 重置所有用户流量
UPDATE v2_user 
SET u = 0, d = 0, t = UNIX_TIMESTAMP()
WHERE plan_id IS NOT NULL;
```

或者通过 API（需要先实现）：
```bash
curl -X POST http://localhost:8080/api/v2/admin/traffic/reset-all \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## 8. 下一步

1. [ ] 实现用户组管理 API 和前端界面
2. [ ] 实现流量管理 API 和前端界面
3. [ ] 修改订单完成逻辑，支持用户组升级
4. [ ] 添加流量预警定时任务
5. [ ] 添加流量重置定时任务
6. [ ] 优化流量统计算法

## 9. 生产环境建议

### 流量包设置
```
试用用户：1-5 GB/月
普通用户：50-100 GB/月
VIP用户：200-500 GB/月
```

### 流控策略
```yaml
# 推荐配置
traffic_control:
  warning_threshold: 80%    # 80%时发送预警
  limit_threshold: 100%     # 100%时限速
  ban_threshold: 120%       # 120%时断开
  reset_day: 1              # 每月1号重置
```

### 用户组配置
```
试用组：
  - 节点：1-2个基础节点
  - 套餐：只能购买入门套餐
  - 流量：1-5 GB

普通组：
  - 节点：3-5个常规节点
  - 套餐：可购买所有基础套餐
  - 流量：根据套餐

VIP组：
  - 节点：所有节点（包括高速节点）
  - 套餐：可购买所有套餐
  - 流量：根据套餐
  - 速度：无限制或更高
```

## 10. 监控和维护

### 每日检查
```bash
# 检查 Agent 状态
rc-service xboard-agent status

# 检查流量统计
mysql -u root -p xboard -e "
  SELECT 
    DATE(FROM_UNIXTIME(t)) as date,
    COUNT(*) as active_users,
    SUM(u + d) / 1024 / 1024 / 1024 as total_gb
  FROM v2_user
  WHERE t >= UNIX_TIMESTAMP(DATE_SUB(NOW(), INTERVAL 1 DAY))
  GROUP BY DATE(FROM_UNIXTIME(t));
"
```

### 每周检查
- 检查超流量用户数量
- 检查流量异常用户
- 检查 Agent 日志是否有错误
- 备份数据库

### 每月任务
- 重置用户流量（如果需要）
- 清理过期用户
- 生成流量报表
- 优化用户组配置


---

## 🎉 新功能：前端管理界面

### 访问管理后台

1. **启动服务**
```bash
./xboard -config configs/config.yaml
```

2. **登录管理后台**
- 访问：`http://localhost:8080/admin`
- 使用管理员账号登录

3. **用户组管理**
- 访问：`http://localhost:8080/admin/user-groups`
- 功能：
  - 查看所有用户组
  - 创建/编辑/删除用户组
  - 为用户组添加节点和套餐
  - 设置默认流量限制

4. **流量管理**
- 访问：`http://localhost:8080/admin/traffic-management`
- 功能：
  - 查看流量统计概览
  - 查看流量预警用户
  - 重置用户流量
  - 批量操作
  - 自动封禁超流量用户

### 快速测试流程

#### 1. 创建用户组
```
管理后台 → 用户组管理 → 创建用户组
- 名称：测试组
- 描述：用于测试的用户组
- 选择节点：勾选要包含的节点
- 选择套餐：勾选可购买的套餐
- 默认流量：10GB
```

#### 2. 创建套餐并关联用户组
```
管理后台 → 套餐管理 → 创建套餐
- 名称：测试套餐
- 流量：100GB
- 升级组ID：选择刚创建的用户组
```

#### 3. 测试订单流程
```
1. 用户购买套餐
2. 完成支付
3. 验证用户组自动升级
4. 检查用户订阅节点是否更新
```

#### 4. 测试流量管理
```
管理后台 → 流量管理
1. 查看流量统计
2. 设置预警阈值（如80%）
3. 查看预警用户列表
4. 测试重置流量功能
```

## 📚 相关文档

- [前端集成文档](FRONTEND_INTEGRATION.md) - 前端功能详细说明
- [最终检查清单](FINAL_CHECKLIST.md) - 完整的测试清单
- [Handler实现](HANDLER_IMPLEMENTATION.md) - API接口文档
- [流量限制说明](docs/traffic-limitation.md) - 流量统计的限制和说明

## 🔧 故障排查

### 前端页面无法访问
```bash
# 检查前端是否构建
cd web && npm run build

# 检查静态文件是否存在
ls -la web/dist
```

### API返回401错误
```bash
# 检查JWT配置
# 确保config.yaml中jwt_secret已设置
# 重新登录获取新token
```

### 用户组不生效
```bash
# 检查数据库
mysql -u root -p xboard -e "SELECT * FROM v2_user_group;"

# 检查用户的group_id
mysql -u root -p xboard -e "SELECT id, email, group_id FROM v2_user;"
```

### 流量统计不准确
- 这是已知限制，详见 `docs/traffic-limitation.md`
- 系统采用平均分配算法
- Agent会优先尝试用户级统计

## 🎯 下一步

1. 按照 [FINAL_CHECKLIST.md](FINAL_CHECKLIST.md) 进行全面测试
2. 根据实际需求调整用户组配置
3. 配置定时任务（流量预警、自动重置）
4. 部署到生产环境

---

**更新时间：** 2025-12-10
**版本：** v2.0.0
