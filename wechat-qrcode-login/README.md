# 微信二维码登录示例项目

这是一个基于 React + TypeScript + Vite 构建的微信二维码登录前端演示项目，通过 WebSocket 实时通信实现扫码登录流程。

## 功能特性

- **实时二维码登录** - 通过 WebSocket 获取并显示二维码
- **登录状态管理** - 完整的登录流程状态追踪（连接中、等待扫描、已扫描、登录成功等）
- **自动重连机制** - WebSocket 断线自动重连，最多重试 3 次
- **心跳保活** - 定期发送心跳消息保持连接活跃
- **二维码过期处理** - 超时自动过期并提供刷新功能
- **主题切换** - 支持明暗主题切换
- **响应式设计** - 适配移动端和桌面端
- **Token 加密存储** - 使用 sessionStorage 和简单加密存储敏感信息

## 技术栈

- **框架**: React 19 + TypeScript
- **构建工具**: Vite 7
- **路由**: Wouter (轻量级路由库)
- **UI 组件**: Radix UI + Tailwind CSS 4
- **状态管理**: React Hooks
- **实时通信**: WebSocket
- **表单处理**: React Hook Form + Zod
- **通知提示**: Sonner

## 项目结构

```
wechat-qrcode-login/
├── src/
│   ├── components/          # UI 组件
│   │   ├── ui/             # 基础 UI 组件库
│   │   ├── ErrorBoundary.tsx
│   │   └── ThemeToggle.tsx
│   ├── contexts/           # React Context
│   │   └── ThemeContext.tsx
│   ├── hooks/              # 自定义 Hooks
│   │   ├── useWeChatLogin.ts
│   │   └── use-mobile.ts
│   ├── pages/              # 页面组件
│   │   ├── Login.tsx       # 登录页
│   │   ├── Home.tsx        # 登录后主页
│   │   └── NotFound.tsx
│   ├── services/           # 服务层
│   │   └── websocket.ts    # WebSocket 登录服务
│   ├── types/              # TypeScript 类型定义
│   │   └── auth.ts
│   ├── utils/              # 工具函数
│   │   └── secureStorage.ts # 安全存储工具
│   ├── constants/          # 常量配置
│   │   └── timing.ts
│   ├── App.tsx
│   └── main.tsx
├── public/
├── .env.example            # 环境变量示例
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## 快速开始

### 环境要求

- Node.js >= 18
- pnpm (推荐) 或 npm

### 安装依赖

```bash
pnpm install
# 或
npm install
```

### 配置环境变量

复制 `.env.example` 创建 `.env` 文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件配置 WebSocket 服务器地址：

```env
# WebSocket 服务器地址
VITE_WS_HOST=127.0.0.1:8443

# WebSocket 路径
VITE_WS_PATH=/api/ws

# 超时时间（毫秒）
VITE_WS_TIMEOUT=300000
```

### 运行开发服务器

```bash
pnpm dev
# 或
npm run dev
```

访问 http://localhost:5173

### 构建生产版本

```bash
pnpm build
# 或
npm run build
```

构建产物在 `dist/` 目录。

### 预览生产构建

```bash
pnpm preview
# 或
npm run preview
```

## 登录流程

1. **连接 WebSocket** - 页面加载时自动连接到配置的 WebSocket 服务器
2. **请求二维码** - 连接成功后发送请求获取登录二维码
3. **显示二维码** - 接收到二维码后展示给用户
4. **等待扫描** - 用户使用微信扫描二维码
5. **扫描确认** - 用户在微信中确认登录
6. **接收 Token** - 服务器返回 access_token 和 refresh_token
7. **存储跳转** - 加密存储 token 并跳转到主页

## WebSocket 协议

### 客户端发送

```json
{
  "type": "qrcode",
  "msg": "client_id_or_session"
}
```

### 服务器响应

**二维码数据**
```json
{
  "type": "QRCODE",
  "data": "data:image/png;base64,..."
}
```

**已扫描**
```json
{
  "type": "SCANNED",
  "data": ""
}
```

**Token**
```json
{
  "type": "TOKEN",
  "data": "{\"access_token\":\"xxx\",\"refresh_token\":\"yyy\"}"
}
```

**过期**
```json
{
  "type": "EXPIRED",
  "data": ""
}
```

**错误**
```json
{
  "type": "ERROR",
  "data": "错误信息"
}
```

## 安全考虑

### ✅ 高危漏洞已修复 (2026-01-14)

项目中的高危安全漏洞已修复，详见 `SECURITY_FIX.md`：

#### 1. ✅ 弱加密算法（已修复）
- **之前**: 使用简单的 XOR + Base64 编码
- **现在**: 使用 Web Crypto API 的 AES-GCM 256-bit 加密
- **改进**: 军事级加密标准，提供认证加密和完整性验证

#### 2. ✅ 硬编码密钥（已修复）
- **之前**: XOR 密钥硬编码在源代码中
- **现在**: 每个会话生成唯一的 256-bit 密钥，存储在内存中且不可提取
- **改进**: 会话隔离，密钥无法导出，浏览器关闭自动销毁

### ✅ 中危漏洞已修复 (2026-01-14)

继高危漏洞修复后，所有中危漏洞也已修复，详见 `SECURITY_MEDIUM_RISK_FIX.md`：

#### 3. ✅ Token 明文显示（已修复）
- **之前**: UI 直接显示完整 token
- **现在**: 默认遮蔽，只显示前6位和后4位，提供显示/隐藏和复制功能
- **改进**: 防止屏幕共享和截图泄露，添加用户控制

#### 4. ✅ 缺少 CSP（已修复）
- **之前**: 没有任何 Content Security Policy
- **现在**: 完整的 CSP 配置 + 多重安全头部（X-Frame-Options, X-Content-Type-Options 等）
- **改进**: 防止 XSS、点击劫持、数据注入等攻击

#### 5. ✅ 缺少输入验证（已修复）
- **之前**: WebSocket 消息直接使用无验证
- **现在**: 使用 Zod 对所有消息进行多层验证（结构、格式、大小限制）
- **改进**: 拒绝无效和恶意消息，防止注入攻击

#### 6. ✅ 缺少 CSRF 保护（已修复）
- **之前**: 没有 CSRF 保护机制
- **现在**: 完整的 CSRF Token 机制（生成、验证、重新生成）
- **改进**: 防止跨站请求伪造和会话劫持

### ⚠️ 剩余低危问题

#### 7. 硬编码敏感信息（低危）
- **问题**: `src/services/websocket.ts:169` 有硬编码的默认 client ID
- **影响**: 可能暴露业务逻辑
- **建议**: 使用环境变量或动态生成

### 生产环境建议

如果要在生产环境使用，需要：

1. **使用 HTTPS/WSS** - 强制使用加密连接
2. **实现真正的加密** - 使用 Web Crypto API 或成熟的加密库
3. **添加 CSRF 保护** - 实现 CSRF Token 机制
4. **Token 刷新机制** - 实现 access_token 过期自动刷新
5. **添加 CSP 策略** - 配置严格的内容安全策略
6. **输入验证和清理** - 对所有外部输入进行验证
7. **日志和监控** - 添加安全日志和异常监控
8. **使用 HttpOnly Cookie** - 考虑使用 HttpOnly Cookie 存储 token
9. **实现会话管理** - 添加会话超时和强制登出机制
10. **安全审计** - 定期进行安全审计和渗透测试

## 代码规范

- 使用 ESLint 进行代码检查
- 使用 Prettier 进行代码格式化
- 遵循 TypeScript 严格模式

```bash
# 运行 lint 检查
pnpm lint

# 格式化代码
pnpm format
```

## 浏览器支持

- Chrome >= 90
- Firefox >= 88
- Safari >= 14
- Edge >= 90

## 开发说明

### 添加新的 UI 组件

本项目使用 shadcn/ui 组件系统，可以通过以下命令添加新组件：

```bash
npx shadcn@latest add [component-name]
```

### 自定义主题

主题配置在 `src/contexts/ThemeContext.tsx` 中，可以自定义颜色、字体等。

### WebSocket 连接配置

WebSocket 连接逻辑在 `src/services/websocket.ts` 中，包含：

- 自动重连机制
- 心跳保活
- 超时处理
- 错误处理

## 常见问题

### Q: WebSocket 连接失败？
A: 检查 `.env` 文件中的 `VITE_WS_HOST` 配置是否正确，确保 WebSocket 服务器正在运行。

### Q: 二维码不显示？
A: 检查浏览器控制台是否有 CORS 错误，确保服务器允许跨域请求。

### Q: Token 存储在哪里？
A: Token 使用简单加密后存储在 sessionStorage 中，浏览器关闭后自动清除。

### Q: 如何修改超时时间？
A: 在 `.env` 文件中修改 `VITE_WS_TIMEOUT` 的值（单位：毫秒）。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关资源

- [React 文档](https://react.dev/)
- [Vite 文档](https://vitejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)
- [Radix UI](https://www.radix-ui.com/)
- [WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)

## 免责声明

本项目仅供学习和演示使用，不建议直接用于生产环境。使用本项目代码所产生的任何安全问题，作者不承担责任。在生产环境中使用前，请进行全面的安全审计和改进。
