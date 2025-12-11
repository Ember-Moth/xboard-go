# Agent 自动更新 API 文档

## Agent API

### 1. 获取版本信息

**接口**: `GET /api/v1/agent/version`

**认证**: 需要 Agent Token（Header: Authorization）

**请求参数**:
- `version` (可选): 当前 Agent 版本号

**响应示例**:
```json
{
  "data": {
    "latest_version": "v1.2.0",
    "download_url": "https://download.sharon.wiki/xboard-agent-linux-amd64",
    "sha256": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
    "file_size": 6090936,
    "strategy": "auto",
    "release_notes": "XBoard Agent v1.2.0\n- 新增自动更新功能\n- 修复流量统计bug"
  }
}
```

**字段说明**:
- `latest_version`: 最新版本号
- `download_url`: 下载地址（必须是 HTTPS）
- `sha256`: 文件 SHA256 哈希（64字符）
- `file_size`: 文件大小（字节）
- `strategy`: 更新策略（auto=自动更新, manual=手动更新）
- `release_notes`: 发布说明

---

### 2. 上报更新状态

**接口**: `POST /api/v1/agent/update-status`

**认证**: 需要 Agent Token（Header: Authorization）

**请求体**:
```json
{
  "from_version": "v1.0.0",
  "to_version": "v1.2.0",
  "status": "success",
  "error_message": "",
  "timestamp": "2024-12-11T18:30:00Z"
}
```

**字段说明**:
- `from_version`: 原版本号（必填）
- `to_version`: 目标版本号（必填）
- `status`: 更新状态（必填）
  - `success`: 更新成功
  - `failed`: 更新失败
  - `rollback`: 已回滚
- `error_message`: 错误信息（失败时填写）
- `timestamp`: 时间戳（必填）

**响应示例**:
```json
{
  "data": "ok"
}
```

---

## 管理后台 API

### 1. 获取版本列表

**接口**: `GET /api/v2/admin/agent/versions`

**认证**: 需要管理员权限

**请求参数**:
- `page`: 页码（默认: 1）
- `page_size`: 每页数量（默认: 20）

**响应示例**:
```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "version": "v1.2.0",
        "download_url": "https://download.sharon.wiki/xboard-agent-linux-amd64",
        "sha256": "abc123...",
        "file_size": 6090936,
        "strategy": "auto",
        "release_notes": "更新说明",
        "is_latest": true,
        "created_at": "2024-12-11T18:30:00Z",
        "updated_at": "2024-12-11T18:30:00Z"
      }
    ],
    "total": 10,
    "page": 1
  }
}
```

---

### 2. 创建版本

**接口**: `POST /api/v2/admin/agent/version`

**认证**: 需要管理员权限

**请求体**:
```json
{
  "version": "v1.2.0",
  "download_url": "https://download.sharon.wiki/xboard-agent-linux-amd64",
  "sha256": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
  "file_size": 6090936,
  "strategy": "manual",
  "release_notes": "XBoard Agent v1.2.0\n- 新功能\n- Bug 修复"
}
```

**响应示例**:
```json
{
  "data": {
    "id": 2,
    "version": "v1.2.0",
    ...
  }
}
```

---

### 3. 更新版本

**接口**: `PUT /api/v2/admin/agent/version/:id`

**认证**: 需要管理员权限

**请求体**: 同创建版本

**响应示例**: 同创建版本

---

### 4. 删除版本

**接口**: `DELETE /api/v2/admin/agent/version/:id`

**认证**: 需要管理员权限

**注意**: 不能删除标记为最新的版本

**响应示例**:
```json
{
  "data": true
}
```

---

### 5. 设置最新版本

**接口**: `POST /api/v2/admin/agent/version/:id/set_latest`

**认证**: 需要管理员权限

**响应示例**:
```json
{
  "data": "ok"
}
```

---

### 6. 获取更新日志

**接口**: `GET /api/v2/admin/agent/update_logs`

**认证**: 需要管理员权限

**请求参数**:
- `host_id`: 主机 ID（可选，筛选特定主机）
- `page`: 页码（默认: 1）
- `page_size`: 每页数量（默认: 20）

**响应示例**:
```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "host_id": 1,
        "from_version": "v1.0.0",
        "to_version": "v1.2.0",
        "status": "success",
        "error_message": "",
        "created_at": "2024-12-11T18:30:00Z"
      }
    ],
    "total": 50,
    "page": 1
  }
}
```

---

## 数据库表结构

### agent_versions

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| version | VARCHAR(50) | 版本号（唯一） |
| download_url | VARCHAR(500) | 下载地址 |
| sha256 | VARCHAR(64) | SHA256 哈希 |
| file_size | BIGINT | 文件大小（字节） |
| strategy | VARCHAR(20) | 更新策略（auto/manual） |
| release_notes | TEXT | 发布说明 |
| is_latest | TINYINT(1) | 是否为最新版本 |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### agent_update_logs

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键 |
| host_id | BIGINT | 主机 ID |
| from_version | VARCHAR(50) | 原版本 |
| to_version | VARCHAR(50) | 目标版本 |
| status | VARCHAR(20) | 状态（success/failed/rollback） |
| error_message | TEXT | 错误信息 |
| created_at | TIMESTAMP | 创建时间 |

---

## 使用流程

### Agent 自动更新流程

1. Agent 定期调用 `GET /api/v1/agent/version` 获取最新版本信息
2. 比较版本号，判断是否需要更新
3. 如果 `strategy` 为 `auto`，自动下载并更新
4. 如果 `strategy` 为 `manual`，等待手动触发
5. 更新完成后，调用 `POST /api/v1/agent/update-status` 上报状态

### 管理员配置流程

1. 访问管理后台 Agent 版本管理页面
2. 点击"添加版本"创建新版本配置
3. 填写版本信息（版本号、下载地址、SHA256、文件大小等）
4. 选择更新策略（自动/手动）
5. 点击"设为最新"将该版本标记为最新版本
6. Agent 将在下次检查时获取到新版本信息

---

## 安全注意事项

1. **HTTPS Only**: 下载地址必须使用 HTTPS 协议
2. **SHA256 验证**: Agent 会验证下载文件的 SHA256 哈希
3. **Token 认证**: 所有 API 请求都需要有效的 Token
4. **文件大小限制**: Agent 限制最大文件大小为 500MB
5. **权限控制**: 管理后台 API 需要管理员权限

---

## 错误处理

### 常见错误码

- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 认证失败
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误

### 错误响应格式

```json
{
  "error": "错误信息描述"
}
```

---

## 测试示例

### 使用 curl 测试

```bash
# 获取版本信息
curl -H "Authorization: your-token" \
  https://panel.example.com/api/v1/agent/version

# 上报更新状态
curl -X POST \
  -H "Authorization: your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "from_version": "v1.0.0",
    "to_version": "v1.2.0",
    "status": "success",
    "error_message": "",
    "timestamp": "2024-12-11T18:30:00Z"
  }' \
  https://panel.example.com/api/v1/agent/update-status
```

---

## 相关文档

- [Agent 自动更新功能文档](./agent-auto-update.md)
- [Agent 安装指南](./agent-setup.md)
