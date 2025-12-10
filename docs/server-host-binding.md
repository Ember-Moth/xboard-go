# 节点-主机绑定关系重构

## 当前问题

当前设计中，主机（Host）和节点（Server）的绑定关系存在混乱：

1. **Server 模型**有 `host_id` 字段，表示节点可以绑定到主机
2. **HostService** 提供了 `BindServerToHost` 和 `UnbindServerFromHost` 方法
3. 但在 **AdminCreateServer/AdminUpdateServer** 接口中，没有 `host_id` 参数
4. 这导致绑定关系不清晰，用户不知道应该在哪里配置

## 设计目标

**节点应该主动绑定主机，而不是主机绑定节点**

理由：
1. 节点需要继承主机的配置（除了 IP 和端口）
2. 节点是服务的提供者，应该知道自己运行在哪个主机上
3. 一个主机可以运行多个节点，但一个节点只能运行在一个主机上
4. 从节点管理界面配置绑定关系更直观

## 实现方案

### 1. 保留 Server.host_id 字段

```go
type Server struct {
    // ...
    HostID *int64 `gorm:"column:host_id;index" json:"host_id"` // 绑定的主机ID
    // ...
}
```

### 2. 修改 AdminCreateServer 接口

添加 `host_id` 参数：

```go
var req struct {
    Name             string                 `json:"name" binding:"required"`
    Type             string                 `json:"type" binding:"required"`
    Host             string                 `json:"host" binding:"required"`
    Port             string                 `json:"port" binding:"required"`
    HostID           *int64                 `json:"host_id"` // 新增：绑定的主机ID
    Rate             float64                `json:"rate"`
    Show             bool                   `json:"show"`
    Tags             []string               `json:"tags"`
    GroupID          []int64                `json:"group_id"`
    ProtocolSettings map[string]interface{} `json:"protocol_settings"`
}

server := &model.Server{
    // ...
    HostID: req.HostID, // 设置主机绑定
    // ...
}
```

### 3. 修改 AdminUpdateServer 接口

同样添加 `host_id` 参数：

```go
var req struct {
    // ...
    HostID *int64 `json:"host_id"` // 新增：绑定的主机ID
    // ...
}

server.HostID = req.HostID // 更新主机绑定
```

### 4. 修改 AdminListServers 接口

返回数据中包含主机信息：

```go
type ServerResponse struct {
    ID               int64                  `json:"id"`
    Name             string                 `json:"name"`
    Type             string                 `json:"type"`
    Host             string                 `json:"host"`
    Port             string                 `json:"port"`
    HostID           *int64                 `json:"host_id"`
    HostName         string                 `json:"host_name,omitempty"` // 主机名称
    // ...
}
```

### 5. 移除 HostService 中的绑定方法

删除以下方法（不再需要）：
- `BindServerToHost`
- `UnbindServerFromHost`
- `GetServersByHostID`（可选保留，用于查询）

### 6. 前端界面调整

**节点管理界面**：
- 添加"绑定主机"下拉选择框
- 显示当前绑定的主机名称
- 可以选择"不绑定"（host_id = null）

**主机管理界面**：
- 显示绑定到该主机的节点列表（只读）
- 不提供绑定/解绑操作（改为在节点界面操作）

## 数据库迁移

不需要迁移，`host_id` 字段已经存在。

## API 变更

### 创建节点

**请求**：
```json
POST /api/admin/server
{
  "name": "香港节点1",
  "type": "shadowsocks",
  "host": "hk1.example.com",
  "port": "443",
  "host_id": 1,  // 新增：绑定到主机ID=1
  "rate": 1.0,
  "show": true,
  "group_id": [1, 2],
  "protocol_settings": {
    "method": "2022-blake3-aes-128-gcm",
    "password": "xxx"
  }
}
```

### 更新节点

**请求**：
```json
PUT /api/admin/server/1
{
  "name": "香港节点1",
  "host_id": 2,  // 新增：修改绑定到主机ID=2
  // ... 其他字段
}
```

### 获取节点列表

**响应**：
```json
{
  "data": [
    {
      "id": 1,
      "name": "香港节点1",
      "type": "shadowsocks",
      "host": "hk1.example.com",
      "port": "443",
      "host_id": 1,
      "host_name": "香港主机1",  // 新增：主机名称
      // ... 其他字段
    }
  ]
}
```

## 使用场景

### 场景1：创建节点并绑定主机

1. 管理员创建主机（Host）
2. 管理员创建节点（Server），选择绑定到某个主机
3. 主机的 Agent 会自动部署该节点的配置

### 场景2：节点迁移到其他主机

1. 管理员编辑节点
2. 修改 `host_id` 为新的主机ID
3. 旧主机的 Agent 会移除该节点配置
4. 新主机的 Agent 会部署该节点配置

### 场景3：节点不绑定主机

1. 管理员创建节点，不选择主机（`host_id` = null）
2. 节点作为逻辑配置存在，不会自动部署
3. 用户可以通过订阅链接获取该节点信息

## 优势

1. **逻辑清晰**：节点主动选择运行在哪个主机上
2. **配置继承**：节点可以继承主机的配置（未来扩展）
3. **易于管理**：在节点界面统一管理绑定关系
4. **灵活性高**：节点可以不绑定主机，作为逻辑配置存在

## 实施步骤

1. ✅ 修改 `AdminCreateServer` 接口，添加 `host_id` 参数
2. ✅ 修改 `AdminUpdateServer` 接口，添加 `host_id` 参数
3. ✅ 修改 `AdminListServers` 接口，返回主机信息
4. ✅ 移除 `HostService` 中不需要的绑定方法
5. ⏳ 前端界面调整（需要前端开发）
6. ⏳ 文档更新

## 注意事项

1. **向后兼容**：现有的 `host_id` 字段保持不变，不需要数据迁移
2. **可选绑定**：`host_id` 为可选字段，节点可以不绑定主机
3. **验证逻辑**：如果设置了 `host_id`，需要验证主机是否存在
4. **级联删除**：删除主机时，需要将绑定的节点的 `host_id` 设置为 null
