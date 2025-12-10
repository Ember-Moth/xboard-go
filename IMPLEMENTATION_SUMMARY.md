# XBoard Go 流控和用户组实现总结

## 已完成的工作

### 1. 用户组架构重构 ✅

#### 核心改动
- **用户组为核心**：节点和套餐都归属于用户组
- 创建了 `v2_user_group` 表作为核心管理表
- 用户组包含：
  - `server_ids`: 该组可访问的节点列表
  - `plan_ids`: 该组可购买的套餐列表
  - 默认流量、速度、设备限制等

#### 新增文件
- `internal/model/user_group.go` - 用户组模型
- `internal/repository/user_group.go` - 用户组数据访问层
- `internal/service/user_group.go` - 用户组业务逻辑层
- `migrations/003_create_user_group.sql` - 数据库迁移脚本
- `migrations/003_create_user_group_rollback.sql` - 回滚脚本

#### 数据关系
```
UserGroup (用户组)
  ├─ server_ids: [1, 2, 3]  # 该组可访问的节点
  └─ plan_ids: [10, 11]      # 该组可购买的套餐

User
  └─ group_id: 2  # 用户属于哪个组

Server (节点)
  # 不再有 group_ids 字段

Plan (套餐)
  └─ upgrade_group_id: 3  # 购买后升级到哪个组（可选）
```

### 2. 流控功能实现 ✅

#### 核心功能
1. **实时流量检查**
   - 在 `GetAvailableUsers` 时自动过滤超流量用户
   - 超流量用户不会出现在节点的用户列表中
   - 自动断开连接（通过不返回用户实现）

2. **流量上报优化**
   - 修改 `TrafficFetch` 添加流量检查
   - 上报流量后立即检查用户状态
   - 超流量用户缓存失效，下次拉取时自动排除

3. **流量统计服务**
   - 创建 `TrafficService` 提供完整的流量管理
   - 流量预警（80%、90%）
   - 流量统计和报表
   - 批量重置流量

#### ⚠️ 流量统计限制
**重要说明：由于 sing-box 技术限制，当前无法精确统计单个用户的流量。**

**当前方案：平均分配算法**
- 统计节点总流量（准确）
- 按用户数平均分配（不够精确）
- 适合小规模部署（< 100 用户）

**建议：**
- 设置较大的流量包（如 100GB/月）
- 采用宽松流控策略（超流量限速而不是断开）
- 在用户协议中说明统计方式

**详细说明：** 请查看 `docs/traffic-limitation.md`

#### 新增文件
- `internal/service/traffic.go` - 流量管理服务

#### 修改文件
- `internal/repository/user.go` - 添加流控检查
- `internal/service/user.go` - 优化流量上报逻辑

### 3. Agent Alpine 兼容性修复 ✅

#### 问题诊断
- Agent 在 Alpine 下无法启动
- 缺少日志输出
- go.mod 缺少依赖

#### 修复内容
1. **添加依赖**
   - 修复 `agent/go.mod`，添加必要的依赖

2. **日志输出**
   - 修改 `agent/install.sh`
   - Alpine 使用 OpenRC，添加日志文件配置
   - 日志路径：`/var/log/xboard-agent.log` 和 `/var/log/xboard-agent.err`

3. **启动检查**
   - 添加启动失败检测
   - 显示错误日志路径

## 使用指南

### 1. 数据库迁移

```bash
# 执行迁移
mysql -u root -p xboard < migrations/003_create_user_group.sql

# 如需回滚
mysql -u root -p xboard < migrations/003_create_user_group_rollback.sql
```

### 2. 用户组管理

#### 创建用户组
```go
group := &model.UserGroup{
    Name:        "VIP用户",
    Description: "高级用户组",
    ServerIDs:   model.JSONArray{1, 2, 3, 5},  // 可访问节点1,2,3,5
    PlanIDs:     model.JSONArray{10, 11, 12},  // 可购买套餐10,11,12
    DefaultTransferEnable: 107374182400,        // 默认100GB
    DefaultSpeedLimit:     &speedLimit,         // 默认速度限制
}
userGroupService.Create(group)
```

#### 为用户组添加节点
```go
// 方式1：单个添加
userGroupService.AddServerToGroup(groupID, serverID)

// 方式2：批量设置（覆盖）
userGroupService.SetServersForGroup(groupID, []int64{1, 2, 3, 5})
```

#### 为用户组添加套餐
```go
// 方式1：单个添加
userGroupService.AddPlanToGroup(groupID, planID)

// 方式2：批量设置（覆盖）
userGroupService.SetPlansForGroup(groupID, []int64{10, 11, 12})
```

#### 获取用户可访问的节点
```go
servers, err := userGroupService.GetAvailableServersForUser(user)
// 返回用户所在组可以访问的所有节点
```

#### 获取用户可购买的套餐
```go
plans, err := userGroupService.GetAvailablePlansForUser(user)
// 返回用户所在组可以购买的所有套餐
```

### 3. 流控管理

#### 检查用户流量
```go
isOver, percentage := trafficService.CheckUserTrafficLimit(user)
// isOver: 是否超流量
// percentage: 使用百分比
```

#### 获取流量预警用户
```go
// 获取流量使用超过80%的用户
users, err := trafficService.GetTrafficWarningUsers(80)

// 发送预警通知
for _, user := range users {
    _, percentage := trafficService.CheckUserTrafficLimit(&user)
    trafficService.SendTrafficWarning(&user, percentage)
}
```

#### 获取流量统计
```go
stats, err := trafficService.GetTrafficStats()
// 返回：
// - total_upload: 总上传
// - total_download: 总下载
// - active_users: 活跃用户数
// - over_traffic_users: 超流量用户数
```

#### 重置用户流量
```go
// 重置单个用户
trafficService.ResetUserTraffic(userID)

// 重置所有用户（定时任务）
count, err := trafficService.ResetAllUsersTraffic()
```

### 4. Agent 部署（Alpine）

```bash
# 安装 Agent
curl -sL https://your-repo/agent/install.sh | bash -s -- https://panel.com token123

# 查看日志
tail -f /var/log/xboard-agent.log
tail -f /var/log/xboard-agent.err

# 管理服务
rc-service xboard-agent status
rc-service xboard-agent restart
rc-service xboard-agent stop
```

## 工作流程示例

### 场景1：新用户注册
```
1. 用户注册 -> User.group_id = 1 (试用用户组)
2. 试用组配置：
   - server_ids: [1, 2]  # 只能访问节点1和2
   - plan_ids: [1, 2, 3] # 可以购买套餐1,2,3
   - default_transfer_enable: 1GB
3. 用户查看节点 -> 只能看到节点1和2
4. 用户查看套餐 -> 只能看到套餐1,2,3
```

### 场景2：用户购买套餐
```
1. 用户购买"VIP月付套餐"
2. 套餐配置：upgrade_group_id = 3 (VIP用户组)
3. 订单完成 -> User.group_id = 3
4. VIP组配置：
   - server_ids: [1, 2, 3, 4, 5]  # 可访问所有节点
   - plan_ids: [10, 11, 12]        # 可购买高级套餐
5. 用户现在可以访问所有节点
```

### 场景3：流量控制
```
1. 用户使用节点产生流量
2. Agent 每60秒上报流量（平均分配算法）
   - 获取节点总流量：1000 MB
   - 节点有 10 个用户
   - 每个用户分配：1000 / 10 = 100 MB
3. 面板接收流量 -> 更新 User.u 和 User.d
4. 检查：User.u + User.d >= User.transfer_enable
5. 如果超流量：
   - 下次 GetAvailableUsers 时不返回该用户
   - 节点自动断开该用户连接
   - 可选：发送邮件通知

注意：由于采用平均分配，单用户流量可能不够精确
```

## API 接口（需要添加）

### 用户组管理
```
GET    /api/v2/admin/user-groups           # 获取用户组列表
POST   /api/v2/admin/user-group            # 创建用户组
GET    /api/v2/admin/user-group/:id        # 获取用户组详情
PUT    /api/v2/admin/user-group/:id        # 更新用户组
DELETE /api/v2/admin/user-group/:id        # 删除用户组

POST   /api/v2/admin/user-group/:id/servers    # 设置用户组节点
POST   /api/v2/admin/user-group/:id/plans      # 设置用户组套餐
```

### 流量管理
```
GET    /api/v2/admin/traffic/stats         # 获取流量统计
GET    /api/v2/admin/traffic/warnings      # 获取流量预警用户
POST   /api/v2/admin/traffic/reset/:id     # 重置用户流量
POST   /api/v2/admin/traffic/reset-all     # 重置所有用户流量
```

## 下一步工作

### 高优先级
1. [ ] 添加用户组管理的 Handler 和路由
2. [ ] 添加流量管理的 Handler 和路由
3. [ ] 修改前端界面，支持新的用户组管理
4. [ ] 修改订单完成逻辑，支持用户组升级
5. [ ] 测试流控是否正常工作

### 中优先级
1. [ ] 添加流量预警定时任务
2. [ ] 添加流量重置定时任务
3. [ ] 优化 Agent 流量统计（支持多种方式）
4. [ ] 添加流量统计图表

### 低优先级
1. [ ] 添加用户组权限管理（更细粒度）
2. [ ] 添加流量日志记录
3. [ ] 添加流量分析报表

## 注意事项

1. **向后兼容**
   - 保留了 Server.group_ids 和 Plan.group_id 字段
   - 旧数据不会丢失，但不再使用这些字段
   - 建议逐步迁移到新架构

2. **流控策略**
   - 当前采用"温和"策略：超流量用户不会被封禁，只是无法连接
   - 如需"激进"策略，可以在 `TrafficFetch` 中取消注释自动封禁代码

3. **流量统计限制（重要）**
   - ⚠️ 当前无法精确统计单个用户流量
   - 采用平均分配算法：总流量 ÷ 用户数
   - 建议设置较大的流量包，减少误差影响
   - 详细说明：`docs/traffic-limitation.md`

4. **性能考虑**
   - 流量检查在数据库层面完成，性能较好
   - 大量用户时建议添加索引：`CREATE INDEX idx_user_traffic ON v2_user(transfer_enable, u, d)`

5. **Agent 日志**
   - Alpine 下日志位置：`/var/log/xboard-agent.log`
   - 如果 Agent 无法启动，先查看错误日志：`tail -f /var/log/xboard-agent.err`

## 测试清单

- [ ] 创建用户组
- [ ] 为用户组添加节点
- [ ] 为用户组添加套餐
- [ ] 用户注册后自动分配到默认组
- [ ] 用户购买套餐后升级组
- [ ] 用户只能看到所在组的节点
- [ ] 用户只能购买所在组的套餐
- [ ] 用户超流量后无法连接节点
- [ ] 流量统计准确
- [ ] 流量预警正常发送
- [ ] Agent 在 Alpine 下正常运行
