# 用户组管理 API 文档

## 概述

用户组是新架构的核心，用于管理用户可以访问的节点和购买的套餐。

## 认证

所有 API 需要管理员权限：
```
Authorization: Bearer <admin_token>
```

## 用户组管理

### 1. 获取用户组列表

```http
GET /api/v2/admin/user-groups
```

**响应示例：**
```json
{
  "data": [
    {
      "id": 1,
      "name": "试用用户",
      "description": "新注册用户默认组",
      "server_ids": [1, 2],
      "plan_ids": [1, 2, 3],
      "default_transfer_enable": 1073741824,
      "default_speed_limit": null,
      "default_device_limit": null,
      "sort": 1,
      "servers": [
        {"id": 1, "name": "香港节点1", "type": "shadowsocks", "host": "hk1.example.com"},
        {"id": 2, "name": "美国节点1", "type": "shadowsocks", "host": "us1.example.com"}
      ],
      "plans": [
        {"id": 1, "name": "入门套餐", "transfer_enable": 10737418240},
        {"id": 2, "name": "基础套餐", "transfer_enable": 53687091200}
      ]
    }
  ]
}
```

### 2. 获取用户组详情

```http
GET /api/v2/admin/user-group/:id
```

**响应示例：**
```json
{
  "data": {
    "id": 2,
    "name": "普通用户",
    "description": "购买基础套餐的用户",
    "server_ids": [1, 2, 3],
    "plan_ids": [1, 2, 3, 4],
    "servers": [...],
    "plans": [...]
  }
}
```

### 3. 创建用户组

```http
POST /api/v2/admin/user-group
Content-Type: application/json

{
  "name": "VIP用户",
  "description": "高级用户组",
  "server_ids": [1, 2, 3, 4, 5],
  "plan_ids": [10, 11, 12],
  "default_transfer_enable": 107374182400,
  "default_speed_limit": 1000,
  "default_device_limit": 5,
  "sort": 3
}
```

**响应示例：**
```json
{
  "data": {
    "id": 4,
    "name": "VIP用户",
    ...
  }
}
```

### 4. 更新用户组

```http
PUT /api/v2/admin/user-group/:id
Content-Type: application/json

{
  "name": "高级VIP",
  "description": "更新后的描述",
  "server_ids": [1, 2, 3, 4, 5, 6],
  "plan_ids": [10, 11, 12, 13]
}
```

### 5. 删除用户组

```http
DELETE /api/v2/admin/user-group/:id
```

**注意：** 如果该组下还有用户，将无法删除。

## 用户组节点管理

### 1. 设置用户组的节点列表（覆盖）

```http
POST /api/v2/admin/user-group/:id/servers
Content-Type: application/json

{
  "server_ids": [1, 2, 3, 4, 5]
}
```

### 2. 为用户组添加单个节点

```http
POST /api/v2/admin/user-group/:id/server
Content-Type: application/json

{
  "server_id": 6
}
```

### 3. 从用户组移除节点

```http
DELETE /api/v2/admin/user-group/:id/server/:server_id
```

## 用户组套餐管理

### 1. 设置用户组的套餐列表（覆盖）

```http
POST /api/v2/admin/user-group/:id/plans
Content-Type: application/json

{
  "plan_ids": [10, 11, 12]
}
```

### 2. 为用户组添加单个套餐

```http
POST /api/v2/admin/user-group/:id/plan
Content-Type: application/json

{
  "plan_id": 13
}
```

### 3. 从用户组移除套餐

```http
DELETE /api/v2/admin/user-group/:id/plan/:plan_id
```

## 流量管理 API

### 1. 获取流量统计

```http
GET /api/v2/admin/traffic/stats
```

**响应示例：**
```json
{
  "data": {
    "total_upload": 1073741824000,
    "total_download": 2147483648000,
    "total_traffic": 3221225472000,
    "active_users": 150,
    "over_traffic_users": 5,
    "upload_gb": 1000.0,
    "download_gb": 2000.0,
    "total_gb": 3000.0
  }
}
```

### 2. 获取流量预警用户

```http
GET /api/v2/admin/traffic/warnings?threshold=80
```

**参数：**
- `threshold`: 流量使用百分比阈值（默认80）

**响应示例：**
```json
{
  "data": [
    {
      "id": 123,
      "email": "user@example.com",
      "upload": 42949672960,
      "download": 85899345920,
      "total_used": 128849018880,
      "transfer_enable": 161061273600,
      "usage_percent": 80.0,
      "is_over_limit": false,
      "total_gb": 120.0,
      "limit_gb": 150.0
    }
  ],
  "total": 15
}
```

### 3. 重置用户流量

```http
POST /api/v2/admin/traffic/reset/:id
```

### 4. 重置所有用户流量

```http
POST /api/v2/admin/traffic/reset-all
```

**响应示例：**
```json
{
  "data": true,
  "message": "已重置流量",
  "count": 150
}
```

### 5. 获取用户流量详情

```http
GET /api/v2/admin/traffic/detail/:id
```

**响应示例：**
```json
{
  "data": {
    "user_id": 123,
    "email": "user@example.com",
    "upload": 42949672960,
    "download": 85899345920,
    "total_used": 128849018880,
    "transfer_enable": 161061273600,
    "remaining": 32212254720,
    "usage_percent": 80.0,
    "is_over_limit": false,
    "upload_gb": 40.0,
    "download_gb": 80.0,
    "total_gb": 120.0,
    "limit_gb": 150.0,
    "remaining_gb": 30.0,
    "last_used_at": 1702345678
  }
}
```

### 6. 发送流量预警通知

```http
POST /api/v2/admin/traffic/warning/:id
```

### 7. 批量发送流量预警

```http
POST /api/v2/admin/traffic/warnings/send?threshold=80
```

**响应示例：**
```json
{
  "data": true,
  "message": "批量发送完成",
  "total": 15,
  "success": 14
}
```

### 8. 自动封禁超流量用户

```http
POST /api/v2/admin/traffic/autoban
```

**响应示例：**
```json
{
  "data": true,
  "message": "已封禁超流量用户",
  "count": 5
}
```

## 使用场景

### 场景1：创建新用户组

```bash
# 1. 创建用户组
curl -X POST http://localhost:8080/api/v2/admin/user-group \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "高级VIP",
    "description": "高级用户组，享受所有节点和套餐",
    "server_ids": [1,2,3,4,5],
    "plan_ids": [10,11,12],
    "default_transfer_enable": 107374182400,
    "sort": 4
  }'

# 2. 查看创建结果
curl http://localhost:8080/api/v2/admin/user-groups \
  -H "Authorization: Bearer $TOKEN"
```

### 场景2：为用户组添加新节点

```bash
# 添加节点ID为6的节点到用户组2
curl -X POST http://localhost:8080/api/v2/admin/user-group/2/server \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"server_id": 6}'
```

### 场景3：查看流量预警用户并发送通知

```bash
# 1. 查看流量使用超过80%的用户
curl http://localhost:8080/api/v2/admin/traffic/warnings?threshold=80 \
  -H "Authorization: Bearer $TOKEN"

# 2. 批量发送预警通知
curl -X POST http://localhost:8080/api/v2/admin/traffic/warnings/send?threshold=80 \
  -H "Authorization: Bearer $TOKEN"
```

### 场景4：每月重置流量

```bash
# 重置所有用户流量（建议在定时任务中执行）
curl -X POST http://localhost:8080/api/v2/admin/traffic/reset-all \
  -H "Authorization: Bearer $TOKEN"
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权（Token无效或过期） |
| 403 | 权限不足（非管理员） |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 注意事项

1. **用户组删除限制**
   - 如果用户组下还有用户，无法删除
   - 需要先将用户移动到其他组

2. **流量统计限制**
   - 当前采用平均分配算法
   - 单用户流量可能不够精确
   - 详见 `docs/traffic-limitation.md`

3. **节点和套餐关联**
   - 节点和套餐不再直接关联用户组
   - 通过用户组的 `server_ids` 和 `plan_ids` 管理
   - 更灵活，更易维护

4. **默认用户组**
   - 新注册用户默认分配到 ID=1 的组
   - 建议保留 ID=1 作为试用组
   - 不要删除默认组
