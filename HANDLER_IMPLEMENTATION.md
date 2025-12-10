# Handler 和路由实现完成

## ✅ 已完成的工作

### 1. 新增 Handler 文件

#### `internal/handler/user_group.go`
用户组管理的所有 Handler：
- `AdminListUserGroups` - 获取用户组列表
- `AdminGetUserGroup` - 获取用户组详情
- `AdminCreateUserGroup` - 创建用户组
- `AdminUpdateUserGroup` - 更新用户组
- `AdminDeleteUserGroup` - 删除用户组
- `AdminSetUserGroupServers` - 设置用户组节点列表
- `AdminSetUserGroupPlans` - 设置用户组套餐列表
- `AdminAddServerToUserGroup` - 添加节点到用户组
- `AdminRemoveServerFromUserGroup` - 从用户组移除节点
- `AdminAddPlanToUserGroup` - 添加套餐到用户组
- `AdminRemovePlanFromUserGroup` - 从用户组移除套餐

#### `internal/handler/traffic.go`
流量管理的所有 Handler：
- `AdminGetTrafficStats` - 获取流量统计
- `AdminGetTrafficWarnings` - 获取流量预警用户
- `AdminResetTraffic` - 重置用户流量
- `AdminResetAllTraffic` - 重置所有用户流量
- `AdminGetUserTrafficDetail` - 获取用户流量详情
- `AdminSendTrafficWarning` - 发送流量预警通知
- `AdminBatchSendTrafficWarnings` - 批量发送流量预警
- `AdminAutobanOverTrafficUsers` - 自动封禁超流量用户

### 2. 更新的文件

#### `internal/handler/handler.go`
添加了新的路由：

**用户组管理路由：**
```go
// User Group management
admin.GET("/user-groups", AdminListUserGroups(services))
admin.GET("/user-group/:id", AdminGetUserGroup(services))
admin.POST("/user-group", AdminCreateUserGroup(services))
admin.PUT("/user-group/:id", AdminUpdateUserGroup(services))
admin.DELETE("/user-group/:id", AdminDeleteUserGroup(services))

// User Group - Server management
admin.POST("/user-group/:id/servers", AdminSetUserGroupServers(services))
admin.POST("/user-group/:id/server", AdminAddServerToUserGroup(services))
admin.DELETE("/user-group/:id/server/:server_id", AdminRemoveServerFromUserGroup(services))

// User Group - Plan management
admin.POST("/user-group/:id/plans", AdminSetUserGroupPlans(services))
admin.POST("/user-group/:id/plan", AdminAddPlanToUserGroup(services))
admin.DELETE("/user-group/:id/plan/:plan_id", AdminRemovePlanFromUserGroup(services))
```

**流量管理路由：**
```go
// Traffic management
admin.GET("/traffic/stats", AdminGetTrafficStats(services))
admin.GET("/traffic/warnings", AdminGetTrafficWarnings(services))
admin.POST("/traffic/reset/:id", AdminResetTraffic(services))
admin.POST("/traffic/reset-all", AdminResetAllTraffic(services))
admin.GET("/traffic/detail/:id", AdminGetUserTrafficDetail(services))
admin.POST("/traffic/warning/:id", AdminSendTrafficWarning(services))
admin.POST("/traffic/warnings/send", AdminBatchSendTrafficWarnings(services))
admin.POST("/traffic/autoban", AdminAutobanOverTrafficUsers(services))
```

#### `internal/service/service.go`
添加了新的服务：
```go
type Services struct {
    // ... 其他服务
    UserGroup   *UserGroupService
    Traffic     *TrafficService
}
```

#### `internal/repository/repository.go`
添加了新的 Repository：
```go
type Repositories struct {
    // ... 其他 Repository
    UserGroup     *UserGroupRepository
}
```

#### `cmd/server/main.go`
添加了 UserGroup 模型的自动迁移：
```go
db.AutoMigrate(
    // ... 其他模型
    &model.UserGroup{},
)
```

### 3. API 文档

创建了完整的 API 文档：`docs/api-user-group.md`

包含：
- 所有接口的详细说明
- 请求/响应示例
- 使用场景
- 错误码说明

## 📋 API 路由列表

### 用户组管理 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v2/admin/user-groups` | 获取用户组列表 |
| GET | `/api/v2/admin/user-group/:id` | 获取用户组详情 |
| POST | `/api/v2/admin/user-group` | 创建用户组 |
| PUT | `/api/v2/admin/user-group/:id` | 更新用户组 |
| DELETE | `/api/v2/admin/user-group/:id` | 删除用户组 |
| POST | `/api/v2/admin/user-group/:id/servers` | 设置用户组节点列表 |
| POST | `/api/v2/admin/user-group/:id/server` | 添加节点到用户组 |
| DELETE | `/api/v2/admin/user-group/:id/server/:server_id` | 从用户组移除节点 |
| POST | `/api/v2/admin/user-group/:id/plans` | 设置用户组套餐列表 |
| POST | `/api/v2/admin/user-group/:id/plan` | 添加套餐到用户组 |
| DELETE | `/api/v2/admin/user-group/:id/plan/:plan_id` | 从用户组移除套餐 |

### 流量管理 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v2/admin/traffic/stats` | 获取流量统计 |
| GET | `/api/v2/admin/traffic/warnings` | 获取流量预警用户 |
| POST | `/api/v2/admin/traffic/reset/:id` | 重置用户流量 |
| POST | `/api/v2/admin/traffic/reset-all` | 重置所有用户流量 |
| GET | `/api/v2/admin/traffic/detail/:id` | 获取用户流量详情 |
| POST | `/api/v2/admin/traffic/warning/:id` | 发送流量预警通知 |
| POST | `/api/v2/admin/traffic/warnings/send` | 批量发送流量预警 |
| POST | `/api/v2/admin/traffic/autoban` | 自动封禁超流量用户 |

## 🧪 测试步骤

### 1. 启动服务

```bash
# 编译
go build -o xboard ./cmd/server

# 运行（会自动创建 UserGroup 表）
./xboard -config configs/config.yaml
```

### 2. 测试用户组 API

```bash
# 获取管理员 Token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/guest/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123456"}' \
  | jq -r '.data.token')

# 获取用户组列表
curl http://localhost:8080/api/v2/admin/user-groups \
  -H "Authorization: Bearer $TOKEN"

# 创建用户组
curl -X POST http://localhost:8080/api/v2/admin/user-group \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试组",
    "description": "测试用户组",
    "server_ids": [1, 2],
    "plan_ids": [1, 2],
    "default_transfer_enable": 10737418240,
    "sort": 10
  }'
```

### 3. 测试流量管理 API

```bash
# 获取流量统计
curl http://localhost:8080/api/v2/admin/traffic/stats \
  -H "Authorization: Bearer $TOKEN"

# 获取流量预警用户
curl http://localhost:8080/api/v2/admin/traffic/warnings?threshold=80 \
  -H "Authorization: Bearer $TOKEN"

# 重置用户流量
curl -X POST http://localhost:8080/api/v2/admin/traffic/reset/1 \
  -H "Authorization: Bearer $TOKEN"
```

## 🔧 下一步工作

### 高优先级

1. **修改订单完成逻辑**
   - 在 `internal/service/order.go` 中
   - 订单完成后，根据套餐的 `upgrade_group_id` 升级用户组
   
   ```go
   // 在 CompleteOrder 函数中添加
   if plan.UpgradeGroupID != nil && *plan.UpgradeGroupID > 0 {
       user.GroupID = plan.UpgradeGroupID
       userRepo.Update(user)
   }
   ```

2. **修改用户订阅接口**
   - 在 `internal/handler/user.go` 的 `UserSubscribe` 中
   - 使用 `UserGroupService.GetAvailableServersForUser` 获取用户可访问的节点
   
   ```go
   servers, err := services.UserGroup.GetAvailableServersForUser(user)
   ```

3. **修改套餐列表接口**
   - 在 `internal/handler/guest.go` 的 `GuestGetPlans` 中
   - 使用 `UserGroupService.GetAvailablePlansForUser` 获取用户可购买的套餐

4. **前端界面开发**
   - 用户组管理页面
   - 流量管理页面
   - 修改节点管理（移除用户组选择）
   - 修改套餐管理（添加升级组选项）

### 中优先级

1. **添加定时任务**
   - 流量预警定时任务（每天检查）
   - 流量重置定时任务（每月1号）
   - 在 `internal/service/scheduler.go` 中添加

2. **优化流量统计**
   - 研究更精确的统计方案
   - 添加流量日志记录

3. **添加用户组权限**
   - 更细粒度的权限控制
   - 用户组继承机制

### 低优先级

1. **添加审计日志**
   - 记录用户组变更
   - 记录流量重置操作

2. **添加数据导出**
   - 导出流量报表
   - 导出用户组配置

3. **性能优化**
   - 添加缓存
   - 优化查询

## 📝 代码示例

### 修改订单完成逻辑

在 `internal/service/order.go` 中找到 `CompleteOrder` 函数，添加：

```go
func (s *OrderService) CompleteOrder(tradeNo, callbackNo string) error {
    // ... 现有代码 ...
    
    // 获取套餐信息
    plan, err := s.planRepo.FindByID(order.PlanID)
    if err != nil {
        return err
    }
    
    // 如果套餐配置了升级组，则升级用户组
    if plan.UpgradeGroupID != nil && *plan.UpgradeGroupID > 0 {
        user.GroupID = plan.UpgradeGroupID
        log.Printf("User %d upgraded to group %d", user.ID, *plan.UpgradeGroupID)
    }
    
    // ... 现有代码 ...
}
```

### 修改用户订阅接口

在 `internal/handler/user.go` 中找到 `UserSubscribe` 函数，修改为：

```go
func UserSubscribe(services *service.Services) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := getUserFromContext(c)
        
        // 使用新的用户组服务获取可访问的节点
        servers, err := services.UserGroup.GetAvailableServersForUser(user)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        // ... 生成订阅配置 ...
    }
}
```

## ⚠️ 注意事项

1. **数据库迁移**
   - 首次启动会自动创建 `v2_user_group` 表
   - 会插入3个默认用户组
   - 确保数据库有足够权限

2. **向后兼容**
   - 保留了旧的 `Server.group_ids` 和 `Plan.group_id` 字段
   - 不会影响现有数据
   - 可以逐步迁移

3. **权限控制**
   - 所有 API 都需要管理员权限
   - 确保 JWT 中间件正常工作

4. **错误处理**
   - 所有 Handler 都有完整的错误处理
   - 返回标准的 JSON 格式

## 📚 相关文档

- `REFACTOR_PLAN.md` - 重构方案
- `IMPLEMENTATION_SUMMARY.md` - 实施总结
- `QUICK_START.md` - 快速开始指南
- `docs/api-user-group.md` - API 文档
- `docs/traffic-limitation.md` - 流量统计限制说明

## ✅ 完成清单

- [x] 创建 UserGroup Handler
- [x] 创建 Traffic Handler
- [x] 注册路由
- [x] 更新 Services
- [x] 更新 Repositories
- [x] 添加自动迁移
- [x] 编写 API 文档
- [ ] 修改订单完成逻辑
- [ ] 修改用户订阅接口
- [ ] 修改套餐列表接口
- [ ] 开发前端界面
- [ ] 添加定时任务
- [ ] 测试所有功能

## 🎉 总结

Handler 和路由部分已经全部完成！现在你可以：

1. 启动服务测试 API
2. 修改订单和订阅逻辑
3. 开发前端界面
4. 部署到生产环境

所有的后端基础设施都已就绪，剩下的主要是业务逻辑的调整和前端开发。
