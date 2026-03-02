# WeChat AI Agent

基于 Go 的微信扫码登录与 OAuth 2.0 授权服务，支持角色权限控制、第三方应用接入和实时 WebSocket 推送。

## 功能特性

- **微信扫码登录**：用户通过微信扫描二维码完成身份认证，Token 通过 WebSocket 实时推送
- **OAuth 2.0 授权服务器**：完整实现授权码模式（Authorization Code Flow），支持 PKCE
- **角色权限控制**：三级角色体系（user / admin / super_admin），Scope 与角色绑定
- **Token 轮换**：每次刷新 Refresh Token 时自动轮换，旧 Token 立即失效
- **Token 吊销**：用户取消关注时自动吊销所有 Token 并加入黑名单
- **第三方应用管理**：Super Admin 可注册、管理 OAuth 客户端应用
- **微信模板消息**：Admin 可向指定用户发送微信模板消息
- **审计日志**：记录所有 OAuth 事件（授权、Token 颁发、吊销等）
- **IP 限流**：基于令牌桶算法的 QR 码请求速率限制
- **CORS 白名单**：严格的来源校验，支持开发模式

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端框架 | Go + [Hertz](https://github.com/cloudwego/hertz) |
| 认证 | JWT (HS256) + OAuth 2.0 |
| 缓存/存储 | Redis |
| 实时通信 | WebSocket |
| 微信集成 | [silenceper/wechat](https://github.com/silenceper/wechat) |
| 日志 | Logrus |
| 前端 | React 19 + TypeScript + Vite |
| UI 组件 | Radix UI + shadcn/ui + Tailwind CSS |

## 项目结构

```
wechat-aiagent/
├── main.go                      # 入口：路由注册、服务初始化
├── config.yaml                  # 生产配置（不提交到版本控制）
├── config.yaml.example          # 配置模板
├── go.mod / go.sum
├── constants/
│   ├── vars.go                  # Token 类型、Redis 前缀、角色、过期时间
│   └── oauth.go                 # OAuth 错误码、Scope 定义、授权事件
├── types/
│   ├── yaml_config.go           # 配置结构体
│   └── oauth.go                 # OAuth 数据模型
├── tools/
│   └── redis_mem.go             # Redis 客户端封装
├── server/
│   ├── hertz_server.go          # HTTP 服务器封装
│   ├── wchat.go                 # 微信消息处理（扫码事件、取关事件）
│   ├── qrcode.go                # 二维码生成与缓存
│   ├── oauth.go                 # JWT 生成与解析
│   ├── oauth_init_service.go    # 启动时从配置文件初始化 OAuth 客户端
│   ├── oauth_client_service.go  # 客户端注册、验证、Secret 哈希
│   ├── oauth_code_service.go    # 授权码、Session、用户授权管理
│   ├── oauth_scope_service.go   # Scope 校验与角色检查
│   ├── oauth_audit_service.go   # 审计日志记录
│   ├── user_service.go          # 用户 Token 管理、WebSocket 连接
│   ├── user_info.go             # 从微信 API 获取用户信息
│   └── template_message.go      # 微信模板消息发送
├── handlers/
│   ├── auth_handler.go          # POST /api/auth/refresh
│   ├── auth_middleware.go       # JWT 校验、角色提取
│   ├── role_middleware.go       # 角色授权中间件
│   ├── cors_middleware.go       # CORS 中间件
│   ├── wshandler.go             # WebSocket /api/ws
│   ├── oauth_authorize_handler.go
│   ├── oauth_token_handler.go
│   ├── oauth_user_handler.go
│   ├── oauth_client_handler.go
│   ├── admin_handler.go
│   ├── template_handler.go
│   └── user_info_handler.go
└── wechat-qrcode-login/         # React 前端
    ├── src/
    │   ├── App.tsx              # 路由配置
    │   └── pages/              # Landing、Login、Home、OAuthAuthorize、OAuthCallback
    └── package.json
```

## 快速开始

### 前置要求

- Go 1.21+
- Redis 6+
- 微信公众号（服务号，已开启网页授权）
- Node.js 18+（前端构建）

### 1. 克隆并安装依赖

```bash
git clone <repo-url>
cd wechat-aiagent
go mod download
```

### 2. 配置文件

```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml`，填入以下必填项：

```yaml
wechat:
  appId: "your_app_id"
  appSecret: "your_app_secret"
  token: "your_wechat_token"          # 微信服务器验证 Token
  superAdmin: "openid_of_super_admin" # 超级管理员的 OpenID
  qrcode:
    - projectId: "unique_project_id"  # 项目唯一标识（建议 MD5 格式）
      projectName: "your-project"
      tmpStr: "scene_string"          # 二维码场景值

jwt:
  secret: "your_jwt_secret_key"

redis:
  addr: "127.0.0.1:6379"
  password: ""

server:
  port: ":8443"
  allowedOrigins:
    - "http://localhost:5173"
  corsOrigins:
    - "http://localhost:5173"
```

### 3. 启动后端

```bash
go run main.go
```

### 4. 启动前端（开发模式）

```bash
cd wechat-qrcode-login
npm install
npm run dev
```

前端默认运行在 `http://localhost:5173`。

### 5. 配置微信服务器

在微信公众平台「开发 → 基本配置」中填写：
- **服务器地址 (URL)**：`https://your-domain/api/wechat/`
- **Token**：与 `config.yaml` 中 `wechat.token` 一致
- **消息加解密方式**：明文模式

## API 接口

完整 OAuth API 文档见 [OAUTH_API.md](./OAUTH_API.md)。

### 公开接口（无需认证）

| 方法 | 路径 | 描述 |
|------|------|------|
| `GET` | `/api/ws` | WebSocket 连接，用于推送二维码和 Token |
| `POST` | `/api/auth/refresh` | 刷新 Access Token |
| `GET` | `/api/oauth/authorize` | OAuth 授权页面初始化 |
| `POST` | `/api/oauth/token` | 授权码换取 Token |
| `GET` | `/api/oauth/userinfo` | 获取 OAuth 用户信息（需 Bearer Token） |
| `*` | `/api/wechat/` | 微信服务器消息接收 |

### 用户接口（需 JWT 认证）

| 方法 | 路径 | 描述 |
|------|------|------|
| `POST` | `/api/user/token/validate` | 校验 Token 有效性 |
| `GET` | `/api/user/role` | 获取当前用户角色 |
| `GET` | `/api/user/info` | 获取当前用户微信信息 |
| `GET` | `/api/user/oauth/grants` | 查看已授权的第三方应用 |
| `DELETE` | `/api/user/oauth/grants/:client_id` | 撤销第三方应用授权 |

### 管理接口（需 admin 角色）

| 方法 | 路径 | 描述 |
|------|------|------|
| `POST` | `/api/admin/template/send` | 发送微信模板消息 |
| `POST` | `/api/admin/user/info` | 按 OpenID 查询用户信息 |

### 超级管理员接口（需 super_admin 角色）

| 方法 | 路径 | 描述 |
|------|------|------|
| `POST` | `/api/super/admin/add` | 添加管理员 |
| `POST` | `/api/super/admin/remove` | 移除管理员 |
| `GET` | `/api/super/admin/list` | 查看管理员列表 |
| `GET` | `/api/super/oauth/client/list` | 查看 OAuth 客户端列表 |
| `GET` | `/api/super/oauth/client/:client_id` | 查看客户端详情 |

## 业务流程

### 扫码登录流程

```
前端 ──WebSocket──▶ 后端：发送 {"type":"qrcode","msg":"projectId"}
后端 ──────────────▶ 调用微信 API 生成临时二维码
后端 ──WebSocket──▶ 前端：下发二维码 URL
用户 ──微信扫码──▶ 微信服务器 ──POST──▶ 后端 /api/wechat/
后端 ──────────────▶ 生成 JWT Access Token + Refresh Token
后端 ──WebSocket──▶ 前端：下发 Token，完成登录
```

### OAuth 授权流程

```
第三方应用 ──────▶ GET /api/oauth/authorize?client_id=&redirect_url=&scope=
后端 ────────────▶ 校验客户端，创建 Session，返回 session_id + projectId
前端展示授权页 ──▶ 用户扫码确认
后端 ────────────▶ 校验用户角色满足 Scope 要求，生成授权码
前端 ────────────▶ 重定向到 redirect_url?code=&state=
第三方后端 ──────▶ POST /api/oauth/token（换取 Access/Refresh Token）
第三方后端 ──────▶ GET /api/oauth/userinfo（使用 Bearer Token 访问）
```

## OAuth Scope 说明

| Scope | 权限 | 最低角色 |
|-------|------|---------|
| `user:read` | 读取基本信息（昵称、头像） | user |
| `user:email` | 读取邮箱 | user |
| `user:role` | 读取用户角色 | user |
| `offline_access` | 长期访问（Refresh Token） | user |
| `message:send` | 发送模板消息 | admin |
| `admin:user:read` | 读取其他用户信息 | admin |
| `admin:user:manage` | 管理用户 | super_admin |

## 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `wechat.qrcodeExpire` | 二维码有效期（秒） | 300 |
| `wechat.qrcodeDefaultExpire` | 临时二维码过期时间 | 1800 |
| `jwt.expire` | Access Token 有效期（秒） | 1440（24小时） |
| `jwt.refreshExpire` | Refresh Token 有效期（秒） | 10080（7天） |
| `server.rateLimitRate` | QR 请求限流速率（次/秒/IP） | 5 |
| `server.rateLimitBurst` | 限流突发容量 | 10 |
| `server.maxProjectIdLen` | ProjectID 最大长度 | 64 |

## 安全设计

- **Token 轮换**：Refresh Token 使用一次即失效，新旧 Token 不可同时使用
- **Token 黑名单**：用户取消关注时，所有关联 Token 立即写入 Redis 黑名单
- **Secret 哈希存储**：OAuth 客户端 Secret 使用 bcrypt 哈希，不存明文
- **Scope 与角色绑定**：申请高权限 Scope 时后端校验用户角色，防止越权
- **CORS 白名单**：生产环境严格校验请求来源
- **审计日志**：所有 OAuth 操作均记录时间戳、IP、User-Agent
- **敏感数据脱敏**：日志中 Token 和 OpenID 仅显示首尾各 4 位

## License

[MIT](./LICENSE)
