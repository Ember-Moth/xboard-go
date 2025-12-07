# XBoard Go

XBoard Go 是一个用 Go 语言重写的代理面板系统，支持多种代理协议。

## 功能特性

- 用户管理：注册、登录、密码修改、流量统计
- 套餐管理：多周期定价、流量限制、速度限制
- 订单管理：创建订单、支付、取消
- 节点管理：支持 Shadowsocks、VMess、VLESS、Trojan、Hysteria2、TUIC 等协议
- 订阅管理：支持 Clash、sing-box、Base64 等多种订阅格式
- 工单系统：用户提交工单、管理员回复
- 邀请返利：邀请码、佣金统计
- 管理后台：完整的后台管理功能

## 快速开始

### 1. 配置文件

复制配置文件并修改：

```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml`，配置数据库、Redis、JWT 等信息。

### 2. 编译运行

```bash
# 编译后端
go build -o xboard ./cmd/server

# 编译前端
cd web
npm install
npm run build
cd ..

# 运行
./xboard -config configs/config.yaml
```

### 3. 初始管理员

在 `configs/config.yaml` 中配置初始管理员：

```yaml
admin:
  email: "admin@example.com"
  password: "your_password"
```

启动后会自动创建管理员账号。

## 配置说明

### 数据库配置

支持 MySQL 和 SQLite：

```yaml
database:
  driver: "mysql"  # mysql 或 sqlite
  host: "127.0.0.1"
  port: 3306
  database: "xboard"
  username: "root"
  password: "your_password"
```

### Redis 配置

```yaml
redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
```

### JWT 配置

```yaml
jwt:
  secret: "change-this-to-a-random-string"
  expire_hour: 24
```

## API 文档

### 用户端 API

- `POST /api/v1/guest/register` - 用户注册
- `POST /api/v1/guest/login` - 用户登录
- `GET /api/v1/guest/plans` - 获取套餐列表
- `GET /api/v1/user/info` - 获取用户信息
- `GET /api/v1/user/subscribe` - 获取订阅信息
- `GET /api/v1/client/subscribe` - 客户端订阅

### 管理端 API

- `GET /api/v2/admin/stats/overview` - 统计概览
- `GET /api/v2/admin/users` - 用户列表
- `GET /api/v2/admin/servers` - 节点列表
- `GET /api/v2/admin/plans` - 套餐列表
- `GET /api/v2/admin/orders` - 订单列表
- `GET /api/v2/admin/tickets` - 工单列表

## 目录结构

```
xboard-go/
├── cmd/server/          # 主程序入口
├── configs/             # 配置文件
├── internal/
│   ├── config/          # 配置加载
│   ├── handler/         # HTTP 处理器
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── protocol/        # 订阅协议生成
│   ├── repository/      # 数据访问层
│   └── service/         # 业务逻辑层
├── pkg/
│   ├── cache/           # Redis 缓存
│   ├── database/        # 数据库连接
│   └── utils/           # 工具函数
└── web/                 # 前端代码
    ├── src/
    │   ├── api/         # API 调用
    │   ├── layouts/     # 布局组件
    │   ├── router/      # 路由配置
    │   ├── stores/      # 状态管理
    │   └── views/       # 页面组件
    └── dist/            # 编译输出
```

## 致谢

本项目的开发离不开以下开源项目和工具的支持：

- [Xboard](https://github.com/cedar2025/Xboard) - 感谢 cedar2025 提供的 Xboard 原版项目，本项目参考了其设计理念和数据库结构
- [sing-box 脚本](https://github.com/fscarmen/sing-box) - 感谢 fscarmen 提供的 sing-box 一键安装脚本
- [AWS Kiro](https://kiro.dev) - 感谢 AWS Kiro 提供的 Claude AI 辅助开发

## 许可证

MIT License
