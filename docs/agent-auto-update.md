# XBoard Agent 自动更新功能

## 概述

XBoard Agent 支持自动更新功能，可以在有新版本发布时自动下载并更新自身，无需手动干预。这大大简化了运维工作，确保所有节点运行最新版本。

## 功能特性

- ✅ **自动版本检查**：定期检查是否有新版本可用
- ✅ **安全下载**：仅支持 HTTPS 下载，验证文件完整性（SHA256）
- ✅ **原子更新**：使用原子操作替换文件，确保更新过程安全
- ✅ **自动回滚**：更新失败时自动回滚到原版本
- ✅ **更新策略**：支持自动更新和手动更新两种策略
- ✅ **更新历史**：记录所有更新操作的历史
- ✅ **零停机**：更新过程中 sing-box 服务继续运行

## 配置选项

### 命令行参数

```bash
xboard-agent \
  -panel https://panel.example.com \
  -token abc123 \
  -auto-update=true \              # 是否启用自动更新检查（默认: true）
  -update-check-interval=3600 \    # 更新检查间隔（秒，默认: 3600）
  -update                          # 手动触发更新
```

### 参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-auto-update` | bool | true | 是否启用自动更新检查 |
| `-update-check-interval` | int | 3600 | 更新检查间隔（秒） |
| `-update` | bool | false | 手动触发更新 |

## 更新流程

### 自动更新流程

```
1. Agent 启动时记录当前版本
   ↓
2. 定期发送心跳到 Panel（包含当前版本）
   ↓
3. Panel 返回最新版本信息
   ↓
4. 比较版本号
   ↓
5. 如果有新版本且策略为 "auto"
   ↓
6. 下载新版本（带重试，最多3次）
   ↓
7. 验证文件（大小、SHA256、权限）
   ↓
8. 备份当前版本
   ↓
9. 原子替换可执行文件
   ↓
10. 重启 Agent
   ↓
11. 验证新版本运行正常
   ↓
12. 删除备份文件
   ↓
13. 发送更新成功通知到 Panel
```

### 手动更新流程

当更新策略为 "manual" 时：

1. Agent 检测到新版本后仅记录日志
2. 管理员使用 `-update` 参数手动触发更新
3. 执行与自动更新相同的更新流程

## 使用示例

### 启用自动更新（默认）

```bash
# 使用默认配置（每小时检查一次）
xboard-agent -panel https://panel.example.com -token abc123

# 自定义检查间隔（每30分钟检查一次）
xboard-agent -panel https://panel.example.com -token abc123 -update-check-interval=1800
```

### 禁用自动更新

```bash
xboard-agent -panel https://panel.example.com -token abc123 -auto-update=false
```

### 手动触发更新

```bash
# 方式1：使用 -update 参数
xboard-agent -panel https://panel.example.com -token abc123 -update

# 方式2：重启 Agent（如果有待处理的更新）
systemctl restart xboard-agent
```

## 更新策略

### 自动更新（auto）

- Panel 返回 `strategy: "auto"` 时启用
- Agent 检测到新版本后自动下载并安装
- 适用于大多数场景

### 手动更新（manual）

- Panel 返回 `strategy: "manual"` 时启用
- Agent 检测到新版本后仅记录日志，等待手动触发
- 适用于需要人工审核的场景

## 安全机制

### 下载安全

- ✅ 仅支持 HTTPS 协议下载
- ✅ 可选的域名白名单验证
- ✅ 防止路径遍历攻击
- ✅ 防止访问系统敏感目录

### 文件验证

- ✅ 验证文件大小
- ✅ 验证 SHA256 哈希
- ✅ 验证文件可执行权限
- ✅ 防止世界可写文件（Unix）

### Token 认证

- ✅ 最小长度验证（16字符）
- ✅ 非法字符检测
- ✅ 所有 API 请求需要 Token

### 更新安全

- ✅ 原子操作替换文件
- ✅ 更新失败自动回滚
- ✅ 更新互斥锁防止并发更新
- ✅ 文件大小限制（最大 500MB）

## 更新历史

Agent 会记录所有更新操作的历史，存储在 `/opt/xboard-agent/update-history.json`：

```json
{
  "records": [
    {
      "timestamp": "2024-12-11T18:30:00Z",
      "from_version": "v1.0.0",
      "to_version": "v1.2.0",
      "status": "success",
      "error_message": ""
    }
  ]
}
```

### 历史记录字段

| 字段 | 说明 |
|------|------|
| `timestamp` | 更新时间 |
| `from_version` | 原版本 |
| `to_version` | 目标版本 |
| `status` | 状态（success/failed/rollback） |
| `error_message` | 错误信息（如果失败） |

### 自动清理

- 系统会自动清理 30 天前的更新记录
- 最多保留最近 10 条记录

## 故障排查

### 更新失败

如果更新失败，Agent 会自动回滚到原版本并继续运行。查看日志：

```bash
# systemd 系统
journalctl -u xboard-agent -f

# Alpine (OpenRC)
tail -f /var/log/xboard-agent.err
```

### 常见错误

#### 1. 下载失败

```
错误: download failed: connection timeout
```

**解决方案**：
- 检查网络连接
- 检查防火墙设置
- 等待自动重试（最多3次）

#### 2. 验证失败

```
错误: verification failed: SHA256 hash mismatch
```

**解决方案**：
- 文件可能已损坏
- 联系管理员检查下载源
- 等待下次更新检查

#### 3. 权限错误

```
错误: permission denied
```

**解决方案**：
- 确保 Agent 以 root 权限运行
- 检查文件权限：`ls -la /opt/xboard-agent/`

#### 4. 磁盘空间不足

```
错误: no space left on device
```

**解决方案**：
- 清理磁盘空间
- 检查磁盘使用：`df -h`

## 版本管理

### 版本号格式

Agent 使用语义化版本规范（SemVer）：

```
v<major>.<minor>.<patch>[-<prerelease>][+<buildmetadata>]

示例:
- v1.0.0
- v1.2.3
- v2.0.0-beta.1
```

### 版本比较规则

- 主版本号（major）：不兼容的 API 变更
- 次版本号（minor）：向后兼容的功能新增
- 修订号（patch）：向后兼容的问题修正

## Panel API

### 获取版本信息

```http
GET /api/v1/agent/version
Authorization: <token>
```

**响应示例**：

```json
{
  "data": {
    "latest_version": "v1.2.0",
    "download_url": "https://download.sharon.wiki/xboard-agent-linux-amd64",
    "sha256": "abc123...",
    "file_size": 6090936,
    "strategy": "auto",
    "release_notes": "- 新增自动更新功能\n- 修复流量统计bug"
  }
}
```

### 更新状态通知

```http
POST /api/v1/agent/update-status
Authorization: <token>
Content-Type: application/json

{
  "from_version": "v1.0.0",
  "to_version": "v1.2.0",
  "status": "success",
  "error_message": "",
  "timestamp": "2024-12-11T18:30:00Z"
}
```

## 最佳实践

### 1. 生产环境

- 使用手动更新策略（`strategy: "manual"`）
- 在测试环境验证新版本后再更新生产环境
- 定期检查更新历史

### 2. 测试环境

- 使用自动更新策略（`strategy: "auto"`）
- 设置较短的检查间隔（如 30 分钟）
- 及时发现和报告问题

### 3. 监控

- 监控更新成功率
- 设置更新失败告警
- 定期检查 Agent 版本分布

### 4. 回滚

如果新版本有问题，可以手动回滚：

```bash
# 停止 Agent
systemctl stop xboard-agent

# 恢复备份（如果存在）
cd /opt/xboard-agent
mv xboard-agent.old xboard-agent

# 启动 Agent
systemctl start xboard-agent
```

## 常见问题

### Q: 更新会中断用户连接吗？

A: 不会。更新过程中 sing-box 服务继续运行，用户连接不受影响。只有 Agent 进程会重启。

### Q: 更新失败会怎样？

A: Agent 会自动回滚到原版本并继续运行，不会影响服务。

### Q: 如何禁用自动更新？

A: 使用 `-auto-update=false` 参数启动 Agent。

### Q: 如何查看当前版本？

A: 查看日志中的启动信息，或使用 `xboard-agent -version`（如果支持）。

### Q: 更新需要多长时间？

A: 通常在 1-2 分钟内完成，取决于网络速度和文件大小。

### Q: 可以自定义下载源吗？

A: 下载 URL 由 Panel 提供，可以在 Panel 后台配置。

## 技术细节

### 文件路径

```
/opt/xboard-agent/
├── xboard-agent           # 当前运行的可执行文件
├── xboard-agent.new       # 下载的新版本（临时）
├── xboard-agent.old       # 备份的旧版本（临时）
└── update-history.json    # 更新历史记录
```

### 原子替换

使用 `rename()` 系统调用实现原子替换，确保更新过程的安全性：

```go
// 1. 备份当前文件
os.Rename(execPath, backupPath)

// 2. 原子替换（rename 是原子操作）
os.Rename(newPath, execPath)

// 3. 如果失败，自动回滚
os.Rename(backupPath, execPath)
```

### 并发控制

使用互斥锁防止并发更新：

```go
type Agent struct {
    updateMutex sync.Mutex
    updating    bool
}

func (a *Agent) performUpdate() error {
    a.updateMutex.Lock()
    defer a.updateMutex.Unlock()
    
    if a.updating {
        return errors.New("update already in progress")
    }
    
    a.updating = true
    defer func() { a.updating = false }()
    
    // 执行更新...
}
```

## 相关文档

- [Agent 安装指南](./agent-setup.md)
- [本地安装指南](./local-installation.md)
- [数据库迁移](./database-migration.md)

## 更新日志

查看 [CHANGELOG_v1.0.0.md](../CHANGELOG_v1.0.0.md) 了解版本更新历史。
