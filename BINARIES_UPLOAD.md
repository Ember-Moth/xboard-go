# 二进制文件上传说明

## 需要上传的文件

请将以下编译好的二进制文件上传到 `https://download.sharon.wiki/`：

### 1. 服务器端 (Server)
- **文件名**: `xboard-server-linux-amd64`
- **位置**: 项目根目录
- **大小**: ~20MB
- **说明**: XBoard 主服务器程序

### 2. 数据库迁移工具 (Migrate)
- **文件名**: `migrate-linux-amd64`
- **位置**: 项目根目录
- **大小**: ~10MB
- **说明**: 数据库迁移工具

### 3. 节点代理 (Agent)
- **文件名**: `xboard-agent-linux-amd64`
- **位置**: `agent/` 目录
- **大小**: ~6MB
- **说明**: 节点代理程序

## 上传后的 URL

上传完成后，文件应该可以通过以下 URL 访问：

```
https://download.sharon.wiki/xboard-server-linux-amd64
https://download.sharon.wiki/migrate-linux-amd64
https://download.sharon.wiki/xboard-agent-linux-amd64
```

## 已更新的安装脚本

以下脚本已更新为从 `https://download.sharon.wiki/` 下载预编译二进制：

1. **setup.sh** - 主安装脚本
   - 添加了 `download_binaries()` 函数
   - 自动检测系统架构 (amd64/arm64)
   - 下载对应架构的二进制文件
   - 创建符号链接方便使用

2. **agent/install.sh** - Agent 安装脚本
   - 更新下载 URL 为 `https://download.sharon.wiki/`
   - 支持 amd64 和 arm64 架构
   - 下载失败时自动回退到源码编译

## 编译信息

- **编译时间**: 2025-12-11
- **Go 版本**: 1.21+
- **编译参数**: `-ldflags="-s -w"` (去除调试信息，减小体积)
- **目标平台**: Linux AMD64

## 测试验证

上传完成后，可以通过以下命令测试下载：

```bash
# 测试 server 下载
wget https://download.sharon.wiki/xboard-server-linux-amd64

# 测试 migrate 下载
wget https://download.sharon.wiki/migrate-linux-amd64

# 测试 agent 下载
wget https://download.sharon.wiki/xboard-agent-linux-amd64
```

## 使用新的安装脚本

上传完成后，用户可以直接使用更新后的安装脚本：

```bash
# 安装 Dashboard
bash setup.sh

# 安装 Agent
curl -sL https://raw.githubusercontent.com/ZYHUO/xboard-go/main/agent/install.sh | bash -s -- <面板地址> <Token>
```

安装脚本会自动从 `https://download.sharon.wiki/` 下载预编译二进制，无需本地编译。
