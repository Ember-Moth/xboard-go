# 架构说明 - 节点与主机的关系

## 误解澄清

之前我说要"删除旧的绑定逻辑"，但实际上**没有旧逻辑需要删除**。所有现有代码都是正确的核心功能。

## 当前架构（完全正确）

### 数据模型

```go
// Server - 逻辑节点
type Server struct {
    ID     int64
    Name   string
    Type   string
    HostID *int64  // 可以绑定到主机（用于自动部署）
    // ...
}

// Host - 物理主机
type Host struct {
    ID     int64
    Name   string
    Token  string
    // ...
}

// ServerNode - 主机上的节点实例
type ServerNode struct {
    ID       int64
    HostID   int64   // 运行在哪个主机上
    ServerID *int64  // 可以绑定到 Server（继承配置）
    // ...
}
```

### 绑定关系

```
┌─────────┐
│ Server  │ 逻辑节点（定义协议、用户组等）
└────┬────┘
     │ host_id (可选)
     ↓
┌─────────┐
│  Host   │ 物理主机（运行 sing-box）
└────┬────┘
     │ 运行多个
     ↓
┌──────────────┐
│ ServerNode   │ 节点实例（实际运行的服务）
└──────────────┘
     ↑
     │ server_id (可选)
     │ 继承配置
┌─────────┐
│ Server  │
└─────────┘
```

## 核心方法说明

### 1. GetByHostID - 查询绑定到主机的节点

```go
// 用途：生成主机配置时，获取所有绑定到该主机的 Server
func (r *ServerRepository) GetByHostID(hostID int64) ([]model.Server, error) {
    var servers []model.Server
    err := r.db.Where("host_id = ?", hostID).Find(&servers).Error
    return servers, err
}
```

**使用场景**：
```go
// 生成 sing-box 配置
func (s *HostService) GenerateSingBoxConfig(hostID int64) {
    // 获取绑定到这个主机的所有 Server
    servers, _ := s.serverRepo.GetByHostID(hostID)
    
    // 为每个 Server 生成 inbound 配置
    for _, server := range servers {
        inbound := buildInboundFromServer(&server)
        // ...
    }
}
```

### 2. UnbindFromHost - 删除主机时解绑节点

```go
// 用途：删除主机时，将所有绑定的 Server 的 host_id 设为 null
func (r *ServerRepository) UnbindFromHost(hostID int64) error {
    return r.db.Model(&model.Server{}).
        Where("host_id = ?", hostID).
        Update("host_id", nil).Error
}
```

**使用场景**：
```go
// 删除主机
func (s *HostService) Delete(hostID int64) error {
    // 1. 先解绑所有 Server（防止孤立引用）
    if err := s.serverRepo.UnbindFromHost(hostID); err != nil {
        return err
    }
    
    // 2. 删除所有 ServerNode
    s.nodeRepo.DeleteByHostID(hostID)
    
    // 3. 删除主机
    return s.hostRepo.Delete(hostID)
}
```

### 3. GetServersByHostID - Service 层封装

```go
// 用途：Service 层封装，提供统一接口
func (s *HostService) GetServersByHostID(hostID int64) ([]model.Server, error) {
    return s.serverRepo.GetByHostID(hostID)
}
```

**使用场景**：
```go
// 可能用于管理界面查询
// 或者其他需要获取主机绑定节点的地方
```

## 新增功能（已实现）

### 节点创建/编辑时可以选择绑定主机

**后端 API**：
```go
// AdminCreateServer - 创建节点时可以设置 host_id
func AdminCreateServer(services *service.Services) gin.HandlerFunc {
    var req struct {
        Name   string `json:"name"`
        Type   string `json:"type"`
        HostID *int64 `json:"host_id"` // 新增：可以选择绑定主机
        // ...
    }
    
    server := &model.Server{
        Name:   req.Name,
        HostID: req.HostID, // 设置绑定
        // ...
    }
}

// AdminUpdateServer - 更新节点时可以修改 host_id
func AdminUpdateServer(services *service.Services) gin.HandlerFunc {
    var req struct {
        HostID *int64 `json:"host_id"` // 可以修改绑定
        // ...
    }
    
    server.HostID = req.HostID // 更新绑定
}
```

**前端界面**（Servers.vue）：
```vue
<div>
  <label>绑定主机</label>
  <select v-model="editingServer!.host_id">
    <option :value="null">不绑定（手动配置）</option>
    <option v-for="h in hosts" :key="h.id" :value="h.id">
      {{ h.name }} ({{ h.ip || '未知IP' }})
    </option>
  </select>
  <p>绑定后将自动部署到主机</p>
</div>
```

## 完整工作流程

### 场景1：自动部署

```
1. 管理员创建 Host（香港主机1）
   ↓
2. 管理员创建 Server（香港节点1），选择绑定到 Host
   ↓
3. Agent 调用 /api/v2/agent/config 获取配置
   ↓
4. 后端调用 GenerateSingBoxConfig(hostID)
   ├─ 调用 GetByHostID(hostID) 获取绑定的 Server
   └─ 为每个 Server 生成 inbound 配置
   ↓
5. Agent 应用配置，启动 sing-box
   ↓
6. 用户通过订阅获取 Server 信息
```

### 场景2：手动配置

```
1. 管理员创建 Server，不绑定 Host
   ↓
2. Server 仅用于生成订阅链接
   ↓
3. 管理员手动在服务器上配置 sing-box
```

### 场景3：删除主机

```
1. 管理员删除 Host
   ↓
2. 后端调用 Delete(hostID)
   ├─ 调用 UnbindFromHost(hostID) 解绑所有 Server
   ├─ 调用 DeleteByHostID(hostID) 删除所有 ServerNode
   └─ 删除 Host
   ↓
3. 原本绑定的 Server 变为未绑定状态
   ↓
4. 这些 Server 仍然存在，可以重新绑定到其他 Host
```

## 总结

### ✅ 现有代码都是正确的

- `GetByHostID` - 必需（生成配置）
- `UnbindFromHost` - 必需（删除主机）
- `GetServersByHostID` - 必需（查询接口）

### ✅ 新功能已实现

- 节点创建时可以选择绑定主机
- 节点编辑时可以修改绑定
- 节点列表显示绑定的主机名称

### ❌ 不存在"旧逻辑"

- 从来没有实现过"从主机界面主动绑定节点"的功能
- 所有代码都是核心功能，无需删除

### ✅ 架构清晰

- Server 可以选择绑定到 Host（自动部署）
- Host 通过查询获取绑定的 Server（生成配置）
- 删除 Host 时自动解绑 Server（防止孤立引用）

## 可以安全提交

所有代码都是正确的，前后端逻辑一致，可以安全提交到 GitHub。
