# Agent Update Strategy Control

## Overview

The XBoard Agent supports two update strategies: **automatic** and **manual**. This allows administrators to control when and how agents are updated.

## Update Strategies

### Auto Strategy

When the Panel returns `strategy: "auto"`, the agent will:
1. Detect the new version via heartbeat or update check
2. Automatically download the new version
3. Verify file integrity (SHA256, size, permissions)
4. Backup the current version
5. Replace the executable atomically
6. Restart the agent with the new version

**Note:** sing-box continues running during the update process. The new agent process will take over sing-box management after restart.

### Manual Strategy

When the Panel returns `strategy: "manual"`, the agent will:
1. Detect the new version via heartbeat or update check
2. Log the update information
3. Wait for manual trigger
4. Store the update information in `updatePending`

To manually trigger the update, restart the agent with the `-update` flag:

```bash
./xboard-agent -panel https://your-panel.com -token YOUR_TOKEN -update
```

## Command Line Flags

### `-update`

Manually trigger an update when a new version is available.

**Usage:**
```bash
./xboard-agent -panel https://panel.example.com -token abc123 -update
```

**Behavior:**
- If no update is pending: Agent runs normally
- If update is pending with `strategy: "manual"`: Executes the update immediately
- If update is pending with `strategy: "auto"`: Executes the update immediately (redundant, as auto updates happen automatically)

## Update Flow

### Automatic Update Flow

```
Agent Heartbeat
    ↓
Panel returns version info (strategy: "auto")
    ↓
Version comparison (newer version detected)
    ↓
Download new version
    ↓
Verify file (SHA256, size, permissions)
    ↓
Backup current version
    ↓
Replace executable (atomic operation)
    ↓
Restart agent
    ↓
New version running
```

### Manual Update Flow

```
Agent Heartbeat
    ↓
Panel returns version info (strategy: "manual")
    ↓
Version comparison (newer version detected)
    ↓
Log update information
    ↓
Store in updatePending
    ↓
Wait for manual trigger (-update flag)
    ↓
User restarts with -update flag
    ↓
Download new version
    ↓
Verify file (SHA256, size, permissions)
    ↓
Backup current version
    ↓
Replace executable (atomic operation)
    ↓
Restart agent
    ↓
New version running
```

## sing-box Service Continuity

During the update process:
- **sing-box continues running** - The agent does NOT stop sing-box
- User connections remain active
- Traffic continues to flow
- The new agent process takes over sing-box management after restart
- No service interruption for end users

## Error Handling

If any step fails during the update:
1. The agent logs the error
2. Automatically rolls back to the backup version (if replacement failed)
3. Continues running the current version
4. Reports the failure to the Panel (if configured)

## Examples

### Example 1: Auto Update

Panel API response:
```json
{
  "data": {
    "latest_version": "v1.2.0",
    "download_url": "https://download.example.com/xboard-agent-linux-amd64",
    "sha256": "abc123...",
    "file_size": 6090936,
    "strategy": "auto",
    "release_notes": "Bug fixes and improvements"
  }
}
```

Agent behavior:
```
🔔 检测到新版本: v1.2.0 (当前版本: v1.0.0)
📝 更新说明: Bug fixes and improvements
🚀 自动更新策略已启用，准备更新...
📥 开始下载新版本...
   下载到: /opt/xboard-agent/xboard-agent.new
✓ 下载完成
🔍 验证文件完整性...
✓ 文件验证通过
💾 备份当前版本...
✓ 备份完成
🔄 替换可执行文件...
✓ 替换完成
ℹ️  sing-box 服务继续运行中...
🔄 重启 Agent...
✓ 更新成功！正在启动新版本 v1.2.0
```

### Example 2: Manual Update

Panel API response:
```json
{
  "data": {
    "latest_version": "v1.2.0",
    "download_url": "https://download.example.com/xboard-agent-linux-amd64",
    "sha256": "abc123...",
    "file_size": 6090936,
    "strategy": "manual",
    "release_notes": "Major update - please review before updating"
  }
}
```

Agent behavior:
```
🔔 检测到新版本: v1.2.0 (当前版本: v1.0.0)
📝 更新说明: Major update - please review before updating
ℹ️  手动更新策略已启用，等待手动触发更新
   下载地址: https://download.example.com/xboard-agent-linux-amd64
   使用 -update 参数重启 Agent 以执行更新
```

Administrator manually triggers update:
```bash
./xboard-agent -panel https://panel.example.com -token abc123 -update
```

Then the update proceeds as in Example 1.

## Testing

The update strategy control feature includes comprehensive tests:

### Unit Tests
- `TestUpdateStrategy_Auto` - Tests auto strategy detection
- `TestUpdateStrategy_Manual` - Tests manual strategy detection
- `TestUpdateInfo_StrategyField` - Tests strategy field in UpdateInfo
- `TestAgent_ManualUpdateFlag` - Tests manual update flag

### Integration Tests
- `TestCompleteUpdateFlow_AutoStrategy` - Tests complete auto update flow
- `TestCompleteUpdateFlow_ManualStrategy` - Tests complete manual update flow
- `TestUpdateFlow_WithManualTrigger` - Tests manual trigger behavior
- `TestUpdateFlow_SingBoxContinuesRunning` - Verifies sing-box continuity
- `TestUpdateStrategy_StrategyEnforcement` - Tests strategy enforcement
- `TestUpdateFlow_HeartbeatIntegration` - Tests heartbeat integration

Run tests:
```bash
cd agent
go test -v -run TestUpdateStrategy
go test -v -run TestUpdateFlow
```

## Requirements Validation

This implementation satisfies the following requirements:

- **Requirement 4.1**: ✓ Panel returns update strategy (auto/manual)
- **Requirement 4.2**: ✓ Auto strategy triggers automatic update
- **Requirement 4.3**: ✓ Manual strategy waits for manual trigger
- **Requirement 4.4**: ✓ Manual update command support via `-update` flag
- **Requirement 4.5**: ✓ sing-box continues running during update

## Security Considerations

1. **HTTPS Only**: Downloads only from HTTPS URLs
2. **SHA256 Verification**: All downloads are verified
3. **Atomic Operations**: File replacement uses atomic rename operations
4. **Automatic Rollback**: Failed updates automatically roll back
5. **Permission Verification**: Executable permissions are verified

## Troubleshooting

### Update not triggering with auto strategy

Check:
1. Agent can reach the Panel API
2. Version comparison is working correctly
3. Download URL is accessible (HTTPS)
4. Sufficient disk space for download and backup

### Manual update not working

Check:
1. Agent was restarted with `-update` flag
2. Update information is available (check logs for "检测到新版本")
3. Download URL is accessible
4. File permissions allow writing to agent directory

### Update fails and rolls back

Check:
1. SHA256 hash matches the downloaded file
2. File size matches expected size
3. Sufficient disk space
4. File permissions are correct
5. No other process is using the agent executable
