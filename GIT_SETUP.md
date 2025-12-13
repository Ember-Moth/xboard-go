# Git 配置指南 - dashGO 项目

## 问题描述
```
remote: Invalid username or token. Password authentication is not supported for Git operations.
```

这个错误表示 GitHub 不再支持使用密码进行 Git 操作，需要使用 Personal Access Token (PAT)。

## 解决方案

### 方法 1：使用 Personal Access Token (推荐)

#### 步骤 1：创建 GitHub Personal Access Token

1. 登录 GitHub
2. 点击右上角头像 → **Settings**
3. 左侧菜单最底部 → **Developer settings**
4. 左侧菜单 → **Personal access tokens** → **Tokens (classic)**
5. 点击 **Generate new token** → **Generate new token (classic)**
6. 填写信息：
   - **Note**: `dashGO Project`
   - **Expiration**: 选择过期时间（建议 90 days 或 No expiration）
   - **Select scopes**: 勾选 `repo` (完整的仓库访问权限)
7. 点击 **Generate token**
8. **重要**：复制生成的 token（只显示一次！）

#### 步骤 2：配置 Git 使用 Token

**Windows (PowerShell):**
```powershell
# 更新远程仓库 URL（使用 token）
git remote set-url origin https://YOUR_TOKEN@github.com/ZYHUO/dashGO.git

# 或者使用用户名和 token
git remote set-url origin https://ZYHUO:YOUR_TOKEN@github.com/ZYHUO/dashGO.git
```

**Linux/macOS:**
```bash
# 更新远程仓库 URL（使用 token）
git remote set-url origin https://YOUR_TOKEN@github.com/ZYHUO/dashGO.git

# 或者使用用户名和 token
git remote set-url origin https://ZYHUO:YOUR_TOKEN@github.com/ZYHUO/dashGO.git
```

**注意**：将 `YOUR_TOKEN` 替换为你刚才复制的 token

#### 步骤 3：验证配置

```bash
# 查看当前远程仓库配置
git remote -v

# 测试推送
git push
```

### 方法 2：使用 SSH 密钥（更安全）

#### 步骤 1：生成 SSH 密钥

```bash
# 生成新的 SSH 密钥
ssh-keygen -t ed25519 -C "your_email@example.com"

# 或者使用 RSA（如果系统不支持 ed25519）
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# 按提示操作，可以直接按 Enter 使用默认路径
```

#### 步骤 2：添加 SSH 密钥到 ssh-agent

**Windows (PowerShell):**
```powershell
# 启动 ssh-agent
Start-Service ssh-agent

# 添加密钥
ssh-add ~\.ssh\id_ed25519
```

**Linux/macOS:**
```bash
# 启动 ssh-agent
eval "$(ssh-agent -s)"

# 添加密钥
ssh-add ~/.ssh/id_ed25519
```

#### 步骤 3：添加 SSH 公钥到 GitHub

1. 复制公钥内容：
```bash
# Windows
type ~\.ssh\id_ed25519.pub

# Linux/macOS
cat ~/.ssh/id_ed25519.pub
```

2. 在 GitHub 上：
   - 点击右上角头像 → **Settings**
   - 左侧菜单 → **SSH and GPG keys**
   - 点击 **New SSH key**
   - **Title**: `dashGO Development`
   - **Key**: 粘贴刚才复制的公钥
   - 点击 **Add SSH key**

#### 步骤 4：更改远程仓库 URL 为 SSH

```bash
git remote set-url origin git@github.com:ZYHUO/dashGO.git
```

#### 步骤 5：验证 SSH 连接

```bash
ssh -T git@github.com
```

应该看到：`Hi ZYHUO! You've successfully authenticated...`

### 方法 3：使用 Git Credential Manager (Windows 推荐)

#### 安装 Git Credential Manager

Git for Windows 通常已包含，如果没有：

```powershell
winget install --id Git.Git -e --source winget
```

#### 配置

```bash
git config --global credential.helper manager-core
```

下次推送时会弹出登录窗口，使用 GitHub 账号登录即可。

## 快速修复命令

### 如果你已经有 Token：

```bash
# 替换 YOUR_TOKEN 为你的实际 token
git remote set-url origin https://YOUR_TOKEN@github.com/ZYHUO/dashGO.git
git push
```

### 如果你想使用 SSH：

```bash
# 生成密钥（如果还没有）
ssh-keygen -t ed25519 -C "your_email@example.com"

# 添加到 GitHub（复制公钥内容）
cat ~/.ssh/id_ed25519.pub  # Linux/macOS
type ~\.ssh\id_ed25519.pub  # Windows

# 更改为 SSH URL
git remote set-url origin git@github.com:ZYHUO/dashGO.git

# 测试
ssh -T git@github.com
git push
```

## 当前项目状态

在修复 Git 认证之前，你需要：

1. **提交所有更改**：
```bash
git add .
git commit -m "重命名项目为 dashGO 并修复 UTF-8 编码错误"
```

2. **修复认证**（使用上述任一方法）

3. **推送到 GitHub**：
```bash
git push origin main
```

## 常见问题

### Q: Token 在哪里输入？
A: Token 直接放在 URL 中，或者在推送时输入（用户名输入 GitHub 用户名，密码输入 Token）

### Q: Token 忘记了怎么办？
A: 在 GitHub Settings → Developer settings → Personal access tokens 中删除旧的，重新生成新的

### Q: SSH 密钥生成失败？
A: 确保已安装 OpenSSH，Windows 用户可以在"设置 → 应用 → 可选功能"中安装

### Q: 推送时仍然要求密码？
A: 检查远程 URL 是否正确配置：`git remote -v`

## 安全建议

1. **不要**将 Token 提交到代码仓库
2. **不要**在公共场合分享 Token
3. **定期**更换 Token
4. **使用** SSH 密钥比 Token 更安全
5. **启用** GitHub 两步验证

## 验证清单

- [ ] 创建了 Personal Access Token 或 SSH 密钥
- [ ] 更新了 Git 远程仓库 URL
- [ ] 测试了 `git push` 命令
- [ ] 成功推送到 GitHub
- [ ] 在 GitHub 上验证了更改

## 相关链接

- [GitHub Personal Access Tokens](https://github.com/settings/tokens)
- [GitHub SSH Keys](https://github.com/settings/keys)
- [Git Credential Manager](https://github.com/GitCredentialManager/git-credential-manager)

---

完成 Git 配置后，你就可以正常推送代码了！
