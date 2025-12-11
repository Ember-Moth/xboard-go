# Requirements Document

## Introduction

XBoard Agent 需要支持自动更新功能，以便在有新版本发布时能够自动下载并更新自身，无需手动干预。这将大大简化运维工作，确保所有节点运行最新版本。

## Glossary

- **Agent**: XBoard 节点代理程序，运行在各个节点服务器上
- **Panel**: XBoard 管理面板，提供 Agent 版本信息和下载地址
- **Binary**: 编译后的可执行文件
- **Update Check**: 版本检查过程，比较当前版本和最新版本
- **Hot Update**: 热更新，在不中断服务的情况下更新程序

## Requirements

### Requirement 1

**User Story:** 作为系统管理员，我希望 Agent 能够自动检查更新，这样我就不需要手动检查每个节点的版本

#### Acceptance Criteria

1. WHEN Agent 启动时 THEN 系统应该记录当前版本号
2. WHEN Agent 发送心跳时 THEN 系统应该将当前版本号发送给 Panel
3. WHEN Panel 返回版本信息时 THEN Agent 应该比较本地版本和远程版本
4. WHEN 检测到新版本时 THEN Agent 应该记录日志并准备更新
5. WHEN 版本号相同时 THEN Agent 应该跳过更新检查

### Requirement 2

**User Story:** 作为系统管理员，我希望 Agent 能够自动下载新版本，这样我就不需要手动上传文件到每个节点

#### Acceptance Criteria

1. WHEN 检测到新版本时 THEN Agent 应该从指定 URL 下载新版本二进制文件
2. WHEN 下载过程中 THEN Agent 应该显示下载进度
3. WHEN 下载完成时 THEN Agent 应该验证文件完整性（文件大小、SHA256）
4. WHEN 下载失败时 THEN Agent 应该重试最多 3 次
5. WHEN 重试失败时 THEN Agent 应该记录错误并继续运行当前版本

### Requirement 3

**User Story:** 作为系统管理员，我希望 Agent 能够安全地替换自身，这样更新过程不会导致服务中断

#### Acceptance Criteria

1. WHEN 下载完成时 THEN Agent 应该将新文件保存为临时文件（如 xboard-agent.new）
2. WHEN 准备更新时 THEN Agent 应该备份当前可执行文件（如 xboard-agent.old）
3. WHEN 执行更新时 THEN Agent 应该使用原子操作替换可执行文件
4. WHEN 更新失败时 THEN Agent 应该自动回滚到备份版本
5. WHEN 更新成功时 THEN Agent 应该重启自身并删除备份文件

### Requirement 4

**User Story:** 作为系统管理员，我希望能够控制更新策略，这样我可以选择立即更新或延迟更新

#### Acceptance Criteria

1. WHEN Panel 返回更新信息时 THEN 信息应该包含更新策略（auto/manual）
2. WHEN 更新策略为 auto 时 THEN Agent 应该自动下载并安装更新
3. WHEN 更新策略为 manual 时 THEN Agent 应该仅记录日志，等待手动触发
4. WHEN 收到手动更新命令时 THEN Agent 应该立即执行更新
5. WHEN 更新过程中 THEN Agent 应该保持 sing-box 服务运行

### Requirement 5

**User Story:** 作为系统管理员，我希望更新过程是安全的，这样不会因为更新失败导致节点不可用

#### Acceptance Criteria

1. WHEN 下载新版本时 THEN Agent 应该验证下载 URL 的合法性（HTTPS）
2. WHEN 验证文件时 THEN Agent 应该检查文件权限和可执行性
3. WHEN 更新失败时 THEN Agent 应该自动回滚并记录详细错误信息
4. WHEN 回滚完成时 THEN Agent 应该发送告警通知到 Panel
5. WHEN 更新成功时 THEN Agent 应该发送成功通知到 Panel

### Requirement 6

**User Story:** 作为开发者，我希望版本号遵循语义化版本规范，这样可以清晰地表达版本变化

#### Acceptance Criteria

1. WHEN 定义版本号时 THEN 版本号应该遵循 SemVer 格式（如 v1.2.3）
2. WHEN 比较版本时 THEN 系统应该正确解析主版本号、次版本号和修订号
3. WHEN 比较版本时 THEN 系统应该按照 SemVer 规则判断新旧
4. WHEN 版本号包含预发布标识时 THEN 系统应该正确处理（如 v1.2.3-beta.1）
5. WHEN 版本号格式错误时 THEN 系统应该拒绝更新并记录错误

### Requirement 7

**User Story:** 作为系统管理员，我希望能够查看更新历史，这样可以追踪每次更新的时间和结果

#### Acceptance Criteria

1. WHEN 执行更新时 THEN Agent 应该记录更新开始时间、版本号
2. WHEN 更新完成时 THEN Agent 应该记录更新结束时间、结果（成功/失败）
3. WHEN 更新失败时 THEN Agent 应该记录失败原因和错误堆栈
4. WHEN 查询更新历史时 THEN Panel 应该返回最近 10 次更新记录
5. WHEN 更新记录过多时 THEN 系统应该自动清理 30 天前的记录

### Requirement 8

**User Story:** 作为系统管理员，我希望更新过程对用户透明，这样不会影响正在使用的连接

#### Acceptance Criteria

1. WHEN 准备更新时 THEN Agent 应该等待所有活跃连接关闭或超时
2. WHEN 等待超时（5分钟）时 THEN Agent 应该强制执行更新
3. WHEN 更新过程中 THEN sing-box 服务应该继续运行
4. WHEN Agent 重启时 THEN 新 Agent 应该接管 sing-box 管理
5. WHEN 重启完成时 THEN 用户连接应该无感知切换
