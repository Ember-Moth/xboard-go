# Agent Auto-Update Design Document

## Overview

本设计文档描述 XBoard Agent 自动更新功能的实现方案。该功能允许 Agent 自动检测新版本、下载更新、安全替换自身，并在失败时自动回滚。

## Architecture

### 组件架构

```
┌─────────────────────────────────────────────────────────────┐
│                        XBoard Panel                          │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ Version API    │  │ Download URL │  │ Update Strategy │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              ▲
                              │ HTTPS
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        XBoard Agent                          │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ Version Check  │  │ Downloader   │  │ Updater         │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
│  ┌────────────────┐  ┌──────────────┐  ┌─────────────────┐ │
│  │ File Verifier  │  │ Rollback     │  │ Update History  │ │
│  └────────────────┘  └──────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 更新流程

```
启动 Agent
    ↓
发送心跳（包含当前版本）
    ↓
Panel 返回最新版本信息
    ↓
比较版本号
    ↓
┌───────────────┐
│ 版本相同？    │ ──Yes──→ 继续正常运行
└───────────────┘
    │ No
    ↓
检查更新策略
    ↓
┌───────────────┐
│ 自动更新？    │ ──No──→ 记录日志，等待手动触发
└───────────────┘
    │ Yes
    ↓
下载新版本（带重试）
    ↓
验证文件（大小、SHA256、权限）
    ↓
备份当前版本
    ↓
原子替换可执行文件
    ↓
┌───────────────┐
│ 替换成功？    │ ──No──→ 回滚到备份版本
└───────────────┘
    │ Yes
    ↓
重启 Agent
    ↓
验证新版本运行正常
    ↓
删除备份文件
    ↓
发送更新成功通知
```

## Components and Interfaces

### 1. Version Manager

**职责**：管理版本信息和版本比较

```go
type VersionManager struct {
    currentVersion string
}

// GetCurrentVersion 获取当前版本
func (vm *VersionManager) GetCurrentVersion() string

// CompareVersion 比较两个版本号
// 返回: -1 (当前版本更旧), 0 (版本相同), 1 (当前版本更新)
func (vm *VersionManager) CompareVersion(remote string) (int, error)

// ParseVersion 解析版本号
func (vm *VersionManager) ParseVersion(version string) (*semver.Version, error)
```

### 2. Update Checker

**职责**：检查是否有新版本可用

```go
type UpdateChecker struct {
    panelURL string
    token    string
    client   *http.Client
}

// CheckUpdate 检查更新
func (uc *UpdateChecker) CheckUpdate(currentVersion string) (*UpdateInfo, error)

type UpdateInfo struct {
    LatestVersion string `json:"latest_version"`
    DownloadURL   string `json:"download_url"`
    SHA256        string `json:"sha256"`
    FileSize      int64  `json:"file_size"`
    Strategy      string `json:"strategy"` // "auto" or "manual"
    ReleaseNotes  string `json:"release_notes"`
}
```

### 3. Downloader

**职责**：下载新版本文件

```go
type Downloader struct {
    client      *http.Client
    maxRetries  int
    retryDelay  time.Duration
}

// Download 下载文件
func (d *Downloader) Download(url, destPath string, progressCallback func(downloaded, total int64)) error

// DownloadWithRetry 带重试的下载
func (d *Downloader) DownloadWithRetry(url, destPath string) error
```

### 4. File Verifier

**职责**：验证下载文件的完整性和安全性

```go
type FileVerifier struct{}

// VerifySize 验证文件大小
func (fv *FileVerifier) VerifySize(filePath string, expectedSize int64) error

// VerifySHA256 验证 SHA256 哈希
func (fv *FileVerifier) VerifySHA256(filePath, expectedHash string) error

// VerifyExecutable 验证文件可执行性
func (fv *FileVerifier) VerifyExecutable(filePath string) error

// VerifyAll 执行所有验证
func (fv *FileVerifier) VerifyAll(filePath string, expectedSize int64, expectedHash string) error
```

### 5. Updater

**职责**：执行更新操作

```go
type Updater struct {
    execPath    string
    backupPath  string
    newPath     string
}

// Backup 备份当前可执行文件
func (u *Updater) Backup() error

// Replace 替换可执行文件（原子操作）
func (u *Updater) Replace() error

// Rollback 回滚到备份版本
func (u *Updater) Rollback() error

// Restart 重启 Agent
func (u *Updater) Restart() error

// CleanupBackup 清理备份文件
func (u *Updater) CleanupBackup() error
```

### 6. Update History

**职责**：记录更新历史

```go
type UpdateHistory struct {
    records []UpdateRecord
}

type UpdateRecord struct {
    Timestamp    time.Time `json:"timestamp"`
    FromVersion  string    `json:"from_version"`
    ToVersion    string    `json:"to_version"`
    Status       string    `json:"status"` // "success", "failed", "rollback"
    ErrorMessage string    `json:"error_message,omitempty"`
}

// AddRecord 添加更新记录
func (uh *UpdateHistory) AddRecord(record UpdateRecord)

// GetRecords 获取更新记录
func (uh *UpdateHistory) GetRecords(limit int) []UpdateRecord

// Cleanup 清理旧记录
func (uh *UpdateHistory) Cleanup(days int)
```

## Data Models

### Version Format

使用语义化版本规范（SemVer）：

```
v<major>.<minor>.<patch>[-<prerelease>][+<buildmetadata>]

示例:
- v1.0.0
- v1.2.3
- v2.0.0-beta.1
- v1.0.0+20231211
```

### Update Info Response

Panel API 返回的更新信息：

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

### Update History File

本地存储的更新历史（JSON 格式）：

```json
{
  "records": [
    {
      "timestamp": "2024-12-11T18:30:00Z",
      "from_version": "v1.0.0",
      "to_version": "v1.2.0",
      "status": "success",
      "error_message": ""
    },
    {
      "timestamp": "2024-12-10T10:15:00Z",
      "from_version": "v1.0.0",
      "to_version": "v1.1.0",
      "status": "failed",
      "error_message": "download failed: connection timeout"
    }
  ]
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Version comparison consistency

*For any* two valid SemVer version strings v1 and v2, comparing v1 to v2 and then v2 to v1 should produce opposite results (if v1 < v2, then v2 > v1)

**Validates: Requirements 6.3**

### Property 2: Download retry idempotence

*For any* download URL and destination path, retrying a failed download should eventually succeed or exhaust all retries, and the final result should be the same regardless of how many intermediate failures occurred

**Validates: Requirements 2.4**

### Property 3: Update atomicity

*For any* update operation, either the update completes successfully with the new version running, or it fails and the old version continues running - there should be no intermediate state where the agent is non-functional

**Validates: Requirements 3.3, 3.4**

### Property 4: File verification completeness

*For any* downloaded file, if all verification checks pass (size, SHA256, executable), then the file should be identical to the original file on the server

**Validates: Requirements 2.3**

### Property 5: Rollback correctness

*For any* failed update, after rollback, the agent should be running the exact same version and configuration as before the update attempt

**Validates: Requirements 5.3**

### Property 6: Update history consistency

*For any* sequence of update operations, the update history should contain exactly one record per update attempt, ordered by timestamp

**Validates: Requirements 7.1, 7.2**

### Property 7: Strategy enforcement

*For any* update check, if the strategy is "manual", the agent should never automatically download or install updates without explicit user command

**Validates: Requirements 4.2, 4.3**

### Property 8: Version monotonicity

*For any* successful update, the new version should be strictly greater than the old version according to SemVer rules

**Validates: Requirements 6.3**

## Error Handling

### Error Categories

1. **Network Errors**
   - Connection timeout
   - DNS resolution failure
   - HTTP errors (4xx, 5xx)
   - 处理：重试最多 3 次，间隔递增（1s, 2s, 4s）

2. **File Errors**
   - Disk full
   - Permission denied
   - File corruption
   - 处理：记录错误，不重试，保持当前版本

3. **Verification Errors**
   - Size mismatch
   - SHA256 mismatch
   - Not executable
   - 处理：删除下载文件，记录错误，不重试

4. **Update Errors**
   - Backup failed
   - Replace failed
   - Restart failed
   - 处理：立即回滚，记录错误，发送告警

### Error Recovery

```go
type UpdateError struct {
    Category string
    Message  string
    Retryable bool
}

func (e *UpdateError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

// HandleError 统一错误处理
func HandleError(err error) {
    switch e := err.(type) {
    case *UpdateError:
        if e.Retryable {
            // 重试逻辑
        } else {
            // 记录错误，放弃更新
        }
    default:
        // 未知错误，记录并放弃
    }
}
```

## Testing Strategy

### Unit Tests

1. **Version Comparison Tests**
   - 测试各种版本号格式的比较
   - 测试预发布版本的处理
   - 测试无效版本号的错误处理

2. **File Verification Tests**
   - 测试 SHA256 验证
   - 测试文件大小验证
   - 测试可执行权限验证

3. **Downloader Tests**
   - 测试正常下载
   - 测试下载失败重试
   - 测试进度回调

4. **Updater Tests**
   - 测试备份操作
   - 测试文件替换
   - 测试回滚操作

### Property-Based Tests

使用 Go 的 `testing/quick` 包或 `gopter` 库进行属性测试：

1. **Property Test: Version Comparison**
   - 生成随机版本号对
   - 验证比较结果的一致性和传递性

2. **Property Test: Update Atomicity**
   - 模拟各种更新失败场景
   - 验证系统始终处于一致状态

3. **Property Test: Rollback Correctness**
   - 随机触发更新失败
   - 验证回滚后状态与更新前完全一致

### Integration Tests

1. **End-to-End Update Test**
   - 启动测试 Agent
   - 模拟 Panel 返回新版本
   - 验证完整更新流程
   - 验证新版本正常运行

2. **Failure Recovery Test**
   - 模拟各种失败场景
   - 验证错误处理和回滚
   - 验证 Agent 继续正常运行

## Implementation Notes

### 文件路径约定

```
/opt/xboard-agent/
├── xboard-agent           # 当前运行的可执行文件
├── xboard-agent.new       # 下载的新版本（临时）
├── xboard-agent.old       # 备份的旧版本（临时）
└── update-history.json    # 更新历史记录
```

### 原子替换实现

在 Linux 上使用 `rename()` 系统调用实现原子替换：

```go
func (u *Updater) Replace() error {
    // 1. 备份当前文件
    if err := os.Rename(u.execPath, u.backupPath); err != nil {
        return err
    }
    
    // 2. 原子替换（rename 是原子操作）
    if err := os.Rename(u.newPath, u.execPath); err != nil {
        // 替换失败，回滚
        os.Rename(u.backupPath, u.execPath)
        return err
    }
    
    return nil
}
```

### 重启实现

使用 `exec.Command` 启动新进程，然后退出当前进程：

```go
func (u *Updater) Restart() error {
    // 获取当前进程的参数
    args := os.Args[1:]
    
    // 启动新进程
    cmd := exec.Command(u.execPath, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    if err := cmd.Start(); err != nil {
        return err
    }
    
    // 等待新进程启动
    time.Sleep(2 * time.Second)
    
    // 退出当前进程
    os.Exit(0)
    
    return nil
}
```

### 并发安全

更新过程中需要确保：
1. 只有一个更新操作在执行
2. 更新期间心跳和流量上报继续工作
3. sing-box 服务不受影响

使用互斥锁保护更新操作：

```go
type Agent struct {
    // ... 其他字段
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

## Security Considerations

1. **HTTPS Only**: 只允许从 HTTPS URL 下载更新
2. **SHA256 Verification**: 必须验证文件哈希
3. **File Permissions**: 验证文件权限，防止权限提升攻击
4. **Path Validation**: 验证文件路径，防止路径遍历攻击
5. **Token Authentication**: 更新 API 需要 Token 认证

## Performance Considerations

1. **下载不阻塞**: 下载在后台进行，不影响心跳和流量上报
2. **重启快速**: 重启过程应在 5 秒内完成
3. **内存占用**: 下载时使用流式处理，避免一次性加载整个文件到内存
4. **磁盘空间**: 更新前检查磁盘空间，确保有足够空间存储新版本和备份

## Deployment

### Panel API 变更

需要在 Panel 添加新的 API 端点：

```
GET /api/v1/agent/version
Response:
{
  "data": {
    "latest_version": "v1.2.0",
    "download_url": "https://download.sharon.wiki/xboard-agent-linux-amd64",
    "sha256": "abc123...",
    "file_size": 6090936,
    "strategy": "auto",
    "release_notes": "..."
  }
}
```

### 配置选项

在 Agent 启动参数中添加更新相关选项：

```bash
xboard-agent \
  -panel https://panel.example.com \
  -token abc123 \
  -auto-update true \           # 是否启用自动更新
  -update-check-interval 3600   # 检查更新间隔（秒）
```

### 发布流程

1. 编译新版本二进制
2. 计算 SHA256 哈希
3. 上传到 `https://download.sharon.wiki/`
4. 在 Panel 后台更新版本信息
5. Agent 自动检测并更新
