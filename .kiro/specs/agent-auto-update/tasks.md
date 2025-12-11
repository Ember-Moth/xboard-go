# Implementation Plan

- [x] 1. 添加版本管理和语义化版本支持






  - 在 agent/main.go 中定义版本常量
  - 引入 semver 库（github.com/Masterminds/semver/v3）
  - 实现 VersionManager 结构体和方法
  - 实现版本比较逻辑（CompareVersion, ParseVersion）
  - _Requirements: 6.1, 6.2, 6.3_

- [x] 2. 实现更新检查功能





  - 创建 UpdateChecker 结构体
  - 实现 CheckUpdate 方法，调用 Panel API
  - 定义 UpdateInfo 数据结构
  - 在心跳响应中包含版本信息
  - 实现版本比较和更新判断逻辑
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 3. 实现文件下载功能





  - 创建 Downloader 结构体
  - 实现基础下载方法（Download）
  - 添加下载进度回调支持
  - 实现重试逻辑（DownloadWithRetry，最多3次）
  - 添加超时和错误处理
  - _Requirements: 2.1, 2.2, 2.4, 2.5_

- [x] 4. 实现文件验证功能





  - 创建 FileVerifier 结构体
  - 实现文件大小验证（VerifySize）
  - 实现 SHA256 哈希验证（VerifySHA256）
  - 实现可执行权限验证（VerifyExecutable）
  - 实现综合验证方法（VerifyAll）
  - _Requirements: 2.3, 5.2_

- [x] 5. 实现更新和回滚功能





  - 创建 Updater 结构体
  - 实现备份当前文件（Backup）
  - 实现原子替换操作（Replace，使用 os.Rename）
  - 实现回滚功能（Rollback）
  - 实现重启逻辑（Restart，使用 exec.Command）
  - 实现清理备份文件（CleanupBackup）
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 6. 实现更新历史记录





  - 创建 UpdateHistory 结构体和 UpdateRecord 数据结构
  - 实现添加记录方法（AddRecord）
  - 实现获取记录方法（GetRecords）
  - 实现清理旧记录方法（Cleanup，30天）
  - 实现持久化到 JSON 文件
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 7. 实现更新策略控制





  - 在 UpdateInfo 中添加 Strategy 字段
  - 实现自动更新逻辑（strategy = "auto"）
  - 实现手动更新逻辑（strategy = "manual"）
  - 添加手动触发更新的命令支持
  - 确保更新过程中 sing-box 继续运行
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 8. 集成更新功能到 Agent 主循环





  - 在 Agent 结构体中添加更新相关字段
  - 在启动时记录当前版本
  - 在心跳时发送版本信息
  - 添加定期检查更新的 ticker（可配置间隔）
  - 实现更新流程的完整调用链
  - 添加更新互斥锁，防止并发更新
  - _Requirements: 1.1, 1.2, 8.3, 8.4_

- [x] 9. 实现错误处理和通知





  - 定义 UpdateError 错误类型
  - 实现统一错误处理函数（HandleError）
  - 实现更新成功通知到 Panel
  - 实现更新失败告警到 Panel
  - 添加详细的日志记录
  - _Requirements: 5.3, 5.4, 5.5_

- [x] 10. 添加安全检查





  - 验证下载 URL 必须是 HTTPS
  - 验证文件路径，防止路径遍历
  - 验证文件权限
  - 添加 Token 认证检查
  - _Requirements: 5.1, 5.2_

- [x] 11. 添加配置选项


  - 添加 -auto-update 启动参数
  - 添加 -update-check-interval 参数
  - 实现配置解析和验证
  - 更新 agent/install.sh 脚本
  - _Requirements: 4.1_

- [x] 12. Panel API 实现



  - 在 Panel 添加 GET /api/v1/agent/version 端点
  - 实现版本信息返回（latest_version, download_url, sha256, file_size, strategy）
  - 添加更新通知接收端点 POST /api/v1/agent/update-status
  - 在管理后台添加版本管理界面
  - _Requirements: 1.2, 5.4, 5.5_

- [x] 13. 更新文档


  - 更新 agent/install.sh 说明
  - 创建 docs/agent-auto-update.md 文档
  - 更新 CHANGELOG_v1.0.0.md
  - 更新 README_SETUP.md
  - 添加更新流程图和使用示例
  - _Requirements: All_

- [x] 14. Checkpoint - 确保所有测试通过



  - 确保所有测试通过，如有问题请询问用户
