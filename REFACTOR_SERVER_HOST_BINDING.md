# 节点-主机绑定关系重构完成

## 改动说明

将主机绑定节点的逻辑改为节点绑定主机，使关系更清晰、更符合实际使用场景。

## 修改内容

### 1. AdminCreateServer - 创建节点时可以绑定主机

**文件**: `internal/handler/admin.go`

**改动**:
- 请求参数添加 `host_id` 字段（可选）
- 如果设置了 `host_id`，验证主机是否存在
- 创建节点时设置 `HostID` 字段

**API 示例**:
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
    "method": "2022-blake3-aes-128-gcm"
  }
}
```

### 2. AdminUpdateServer - 更新节点时可以修改主机绑定

**文件**: `internal/handler/admin.go`

**改动**:
- 请求参数添加 `host_id` 字段（可选）
- 如果设置了 `host_id`，验证主机是否存在
- 更新节点时设置 `HostID` 字段

**API 示例**:
```json
PUT /api/admin/server/1
{
  "name": "香港节点1",
  "host_id": 2,  // 修改绑定到主机ID=2，设置为 null 可以解除绑定
  // ... 其他字段
}
```

### 3. AdminListServers - 列表中显示主机信息

**文件**: `internal/handler/admin.go`

**改动**:
- 获取所有主机信息
- 为每个节点添加 `host_name` 字段（如果绑定了主机）
- 返回增强的响应数据

**API 响应示例**:
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
      "rate": 1.0,
      "show": true,
      // ... 其他字段
    }
  ]
}
```

### 4. HostService - 移除主动绑定方法

**文件**: `internal/service/host.go`

**改动**:
- ❌ 删除 `BindServerToHost` 方法（不再需要）
- ❌ 删除 `UnbindServerFromHost` 方法（不再需要）
- ✅ 保留 `GetServersByHostID` 方法（只读查询）
- ✅ 保留 `Delete` 方法中的解绑逻辑（删除主机时自动解绑节点）

## 使用场景

### 场景1：创建节点并绑定主机

```bash
# 1. 创建主机
POST /api/admin/host
{
  "name": "香港主机1"
}
# 返回: { "data": { "id": 1, "token": "xxx" } }

# 2. 创建节点并绑定到主机
POST /api/admin/server
{
  "name": "香港节点1",
  "type": "shadowsocks",
  "host": "hk1.example.com",
  "port": "443",
  "host_id": 1,  // 绑定到主机1
  // ...
}
```

### 场景2：修改节点的主机绑定

```bash
# 将节点1从主机1迁移到主机2
PUT /api/admin/server/1
{
  "host_id": 2  // 修改绑定
}

# 解除节点1的主机绑定
PUT /api/admin/server/1
{
  "host_id": null  // 解除绑定
}
```

### 场景3：查看主机绑定的节点

```bash
# 获取节点列表，可以看到每个节点绑定的主机
GET /api/admin/servers
# 返回包含 host_id 和 host_name 字段
```

### 场景4：删除主机

```bash
# 删除主机时，自动解除所有绑定的节点
DELETE /api/admin/host/1
# 所有 host_id=1 的节点会被设置为 host_id=null
```

## 数据库变更

**无需迁移**，`host_id` 字段已经存在于 `v2_server` 表中。

## 前端需要的改动

### 1. 节点创建/编辑表单

添加"绑定主机"选择框：

```jsx
<FormItem label="绑定主机" name="host_id">
  <Select allowClear placeholder="选择主机（可选）">
    <Option value={null}>不绑定</Option>
    {hosts.map(host => (
      <Option key={host.id} value={host.id}>
        {host.name}
      </Option>
    ))}
  </Select>
</FormItem>
```

### 2. 节点列表

显示主机信息：

```jsx
<Table>
  <Column title="节点名称" dataIndex="name" />
  <Column title="类型" dataIndex="type" />
  <Column 
    title="绑定主机" 
    render={(record) => record.host_name || '未绑定'}
  />
  {/* ... 其他列 */}
</Table>
```

### 3. 主机详情页

显示绑定的节点列表（只读）：

```jsx
<Card title="绑定的节点">
  <List
    dataSource={servers}
    renderItem={server => (
      <List.Item>
        <Link to={`/server/${server.id}`}>{server.name}</Link>
      </List.Item>
    )}
  />
</Card>
```

## 优势

1. ✅ **逻辑清晰**: 节点主动选择运行在哪个主机上
2. ✅ **易于管理**: 在节点界面统一管理绑定关系
3. ✅ **灵活性高**: 节点可以不绑定主机，作为逻辑配置存在
4. ✅ **向后兼容**: 不需要数据迁移，现有数据保持不变
5. ✅ **自动清理**: 删除主机时自动解除节点绑定

## 测试建议

1. 创建节点时绑定主机
2. 创建节点时不绑定主机
3. 更新节点的主机绑定
4. 解除节点的主机绑定
5. 删除主机，验证节点的 host_id 被设置为 null
6. 绑定不存在的主机ID，验证返回错误

## 相关文件

```
modified:   internal/handler/admin.go
modified:   internal/service/host.go
new file:   docs/server-host-binding.md
new file:   REFACTOR_SERVER_HOST_BINDING.md
```
