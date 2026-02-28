# OAuth 2.0 API 文档

## 目录

1. [OAuth 2.0 概述](#oauth-20-概述)
2. [授权流程](#授权流程)
3. [超级管理员接口](#超级管理员接口)
4. [OAuth 授权接口](#oauth-授权接口)
5. [用户授权管理接口](#用户授权管理接口)
6. [第三方应用接入指南](#第三方应用接入指南)

---

## OAuth 2.0 概述

本系统实现了标准的 OAuth 2.0 授权码流程（Authorization Code Flow），支持 PKCE（Proof Key for Code Exchange）增强安全性。

### 主要特性

- ✅ 标准 OAuth 2.0 授权码流程
- ✅ PKCE 支持（防止授权码截获）
- ✅ 细粒度 Scope 权限控制
- ✅ Token Rotation（refresh token 单次使用）
- ✅ 客户端撤销和授权撤销
- ✅ 完整的审计日志
- ✅ 限流保护
- ✅ 与现有登录系统兼容（不影响原有功能）

### Scope 定义

| Scope | 说明 | 所需角色 |
|-------|------|---------|
| `user:read` | 读取用户基本信息（昵称、头像） | user |
| `user:email` | 读取用户邮箱 | user |
| `user:role` | 读取用户角色 | user |
| `message:send` | 发送模板消息 | admin |
| `admin:user:read` | 读取其他用户信息 | admin |
| `admin:user:manage` | 管理用户 | super_admin |
| `offline_access` | 长期访问权限（获取 refresh token） | user |

---

## 授权流程

### 完整流程图

```
1. 第三方应用发起授权请求
   ↓
2. 用户跳转到授权页面
   ↓
3. 用户扫描微信二维码
   ↓
4. 系统生成授权码并重定向回第三方应用
   ↓
5. 第三方应用用授权码换取 access token 和 refresh token
   ↓
6. 第三方应用使用 access token 访问用户资源
```

---

## 超级管理员接口

⚠️ **重要说明：** 本系统的 OAuth 客户端**完全基于配置文件管理**，不提供动态注册接口。所有第三方应用必须在 `config.yaml` 中预先配置。详见 [OAUTH_CONFIG.md](./OAUTH_CONFIG.md)。

### 1. 获取客户端列表

**接口地址：** `GET /api/super/oauth/client/list`

**认证要求：** 超级管理员 Access Token

**成功响应：** `200 OK`
```json
{
  "code": 200,
  "message": "获取客户端列表成功",
  "data": [],
  "note": "完整实现需要使用 Redis SCAN 或专门的索引结构"
}
```

---

### 2. 获取客户端详情

**接口地址：** `GET /api/super/oauth/client/:client_id`

**认证要求：** 超级管理员 Access Token

**成功响应：** `200 OK`
```json
{
  "code": 200,
  "message": "获取客户端成功",
  "data": {
    "client_id": "app_xxx",
    "name": "第三方应用名称",
    "description": "应用描述",
    "redirect_urls": ["https://app.example.com/oauth/callback"],
    "allowed_scopes": ["user:read", "user:email"],
    "created_at": "2025-01-27T10:00:00Z",
    "created_by": "super_admin_openid",
    "status": "active",
    "rate_limit": 1000
  }
}
```

**如何管理客户端：**
- 添加客户端：编辑 `config.yaml` 并重启服务
- 修改客户端：编辑 `config.yaml` 并重启服务
- 删除客户端：从 `config.yaml` 移除并重启服务

详细配置说明请参考 [OAUTH_CONFIG.md](./OAUTH_CONFIG.md)。

---

## OAuth 授权接口

### 1. 授权请求

发起 OAuth 授权请求，用户将看到授权页面并扫码。

**接口地址：** `GET /api/oauth/authorize`

**请求参数：**

| 参数名                   | 类型 | 必填 | 说明 |
|-----------------------|------|------|------|
| client_id             | string | 是 | 客户端 ID |
| redirect_url          | string | 是 | 授权后的回调地址 |
| response_type         | string | 是 | 固定值：`code` |
| scope                 | string | 是 | 空格分隔的 scope 列表 |
| state                 | string | 推荐 | 防 CSRF 攻击的随机字符串 |
| code_challenge        | string | 可选 | PKCE code challenge |
| code_challenge_method | string | 可选 | PKCE 方法：`S256` 或 `plain` |

**示例请求：**
```
GET /api/oauth/authorize?
  client_id=app_xxx&
  redirect_url=https://app.com/callback&
  response_type=code&
  scope=user:read user:email offline_access&
  state=random_string&
  code_challenge=sha256_hash&
  code_challenge_method=S256
```

**成功响应：** `200 OK` (HTML 授权页面)

用户扫码成功后，浏览器将重定向到：
```
https://app.com/callback?code=ac_xxx&state=random_string
```

---

### 2. Token 端点（授权码换取 Token）

**接口地址：** `POST /api/oauth/token`

**Content-Type：** `application/x-www-form-urlencoded`

**请求参数（Authorization Code Grant）：**

| 参数名           | 类型 | 必填 | 说明 |
|---------------|------|------|------|
| grant_type    | string | 是 | 固定值：`authorization_code` |
| code          | string | 是 | 授权码 |
| redirect_url  | string | 是 | 回调地址（需与授权请求一致） |
| client_id     | string | 是 | 客户端 ID |
| client_secret | string | 是 | 客户端密钥 |
| code_verifier | string | 可选 | PKCE code verifier |

**示例请求：**
```bash
curl -X POST https://yourapp.com/api/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=ac_xxx" \
  -d "redirect_url=https://app.com/callback" \
  -d "client_id=app_xxx" \
  -d "client_secret=cs_xxx" \
  -d "code_verifier=xxx"
```

**成功响应：** `200 OK`
```json
{
  "access_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 1440,
  "refresh_token": "eyJhbGci...",
  "scope": "user:read user:email offline_access"
}
```

---

### 3. Token 端点（Refresh Token）

**请求参数（Refresh Token Grant）：**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| grant_type | string | 是 | 固定值：`refresh_token` |
| refresh_token | string | 是 | Refresh Token |
| client_id | string | 是 | 客户端 ID |
| client_secret | string | 是 | 客户端密钥 |

**示例请求：**
```bash
curl -X POST https://yourapp.com/api/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=eyJhbGci..." \
  -d "client_id=app_xxx" \
  -d "client_secret=cs_xxx"
```

**成功响应：** `200 OK`
```json
{
  "access_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 1440,
  "refresh_token": "eyJhbGci...",
  "scope": "user:read user:email offline_access"
}
```

⚠️ **重要：** 旧的 refresh token 会立即失效（Token Rotation）。

---

### 4. UserInfo 端点

获取用户信息（符合 OAuth 2.0 标准）。

**接口地址：** `GET /api/oauth/userinfo`

**认证要求：** OAuth Access Token（需包含 `user:read` scope）

**请求头：**
```
Authorization: Bearer <oauth_access_token>
```

**成功响应：** `200 OK`
```json
{
  "sub": "user_openid",
  "name": "微信用户",
  "picture": "https://wx.qlogo.cn/...",
  "email": "user@example.com",
  "role": "admin"
}
```

**响应字段取决于 scope：**
- `user:read` → `sub`, `name`, `picture`
- `user:email` → `email`
- `user:role` → `role`

---

## 用户授权管理接口

### 1. 查看我的授权列表

**接口地址：** `GET /api/user/oauth/grants`

**认证要求：** Access Token

**成功响应：** `200 OK`
```json
{
  "code": 200,
  "message": "获取授权列表成功",
  "data": [
    {
      "client_id": "app_xxx",
      "client_name": "第三方应用名称",
      "scope": ["user:read", "user:email"],
      "granted_at": "2025-01-27T10:00:00Z",
      "last_used_at": "2025-01-27T15:30:00Z",
      "token_count": 5
    }
  ]
}
```

---

### 2. 撤销授权

**接口地址：** `DELETE /api/user/oauth/grants/:client_id`

**认证要求：** Access Token

**成功响应：** `200 OK`
```json
{
  "code": 200,
  "message": "授权已撤销，该应用的所有 token 将失效"
}
```

---

## 第三方应用接入指南

### Node.js 示例

```javascript
const crypto = require('crypto');
const axios = require('axios');

const CLIENT_ID = 'app_xxx';
const CLIENT_SECRET = 'cs_xxx';
const REDIRECT_URL = 'https://myapp.com/oauth/callback';
const AUTHORIZE_URL = 'https://yourapp.com/api/oauth/authorize';
const TOKEN_URL = 'https://yourapp.com/api/oauth/token';

// 1. 生成 PKCE
function generatePKCE() {
  const verifier = crypto.randomBytes(32).toString('base64url');
  const challenge = crypto
    .createHash('sha256')
    .update(verifier)
    .digest('base64url');
  return { verifier, challenge };
}

// 2. 发起授权请求
app.get('/login', (req, res) => {
  const state = crypto.randomBytes(16).toString('hex');
  const { verifier, challenge } = generatePKCE();

  req.session.oauth_state = state;
  req.session.code_verifier = verifier;

  const authUrl = `${AUTHORIZE_URL}?${new URLSearchParams({
    client_id: CLIENT_ID,
    redirect_url: REDIRECT_URL,
    response_type: 'code',
    scope: 'user:read user:email offline_access',
    state: state,
    code_challenge: challenge,
    code_challenge_method: 'S256'
  })}`;

  res.redirect(authUrl);
});

// 3. 处理回调
app.get('/oauth/callback', async (req, res) => {
  const { code, state } = req.query;

  if (state !== req.session.oauth_state) {
    return res.status(400).send('Invalid state');
  }

  try {
    const response = await axios.post(TOKEN_URL, new URLSearchParams({
      grant_type: 'authorization_code',
      code: code,
      client_id: CLIENT_ID,
      client_secret: CLIENT_SECRET,
      redirect_url: REDIRECT_URL,
      code_verifier: req.session.code_verifier
    }));

    const { access_token, refresh_token } = response.data;
    req.session.access_token = access_token;
    req.session.refresh_token = refresh_token;

    res.send('登录成功！');
  } catch (error) {
    console.error(error.response.data);
    res.status(500).send('登录失败');
  }
});

// 4. 访问用户资源
app.get('/api/userinfo', async (req, res) => {
  try {
    const response = await axios.get('https://yourapp.com/api/oauth/userinfo', {
      headers: { 'Authorization': `Bearer ${req.session.access_token}` }
    });
    res.json(response.data);
  } catch (error) {
    if (error.response.status === 401) {
      // Token 过期，使用 refresh token 刷新
      // ...
    }
    res.status(500).send('请求失败');
  }
});
```

---

## 安全建议

### 1. 使用 PKCE

对于单页应用（SPA）和移动应用，**强烈建议**使用 PKCE 防止授权码截获攻击。

### 2. 验证 State 参数

始终验证回调中的 `state` 参数，防止 CSRF 攻击。

### 3. 安全存储密钥

- 后端应用：将 `client_secret` 存储在环境变量中
- 前端应用：使用后端代理，不要在前端暴露密钥

### 4. HTTPS Only

生产环境**必须**使用 HTTPS，避免中间人攻击。

### 5. Redirect URI 白名单

严格配置 `redirect_urls` 白名单，不允许通配符。

---

## OAuth 错误响应

所有 OAuth 错误响应符合 RFC 6749 标准：

```json
{
  "error": "invalid_grant",
  "error_description": "Authorization code is invalid or expired"
}
```

### 常见错误码

| 错误码 | 说明 |
|--------|------|
| `invalid_request` | 请求参数错误 |
| `invalid_client` | 客户端认证失败 |
| `invalid_grant` | 授权码无效 |
| `unauthorized_client` | 客户端无权限 |
| `unsupported_grant_type` | 不支持的 grant_type |
| `invalid_scope` | scope 无效 |
| `server_error` | 服务器错误 |

---

## 限流

- **授权端点**：10 次/分钟/IP
- **Token 端点**：1000 次/小时/客户端

触发限流将返回：
```json
{
  "error": "rate_limit_exceeded",
  "error_description": "请求过于频繁，请稍后再试"
}
```

---

## 与现有系统的兼容性

✅ **完全兼容**：OAuth 2.0 功能不影响现有的微信扫码登录功能。

- 现有的 `/api/ws` 登录流程保持不变
- 现有的 `/api/auth/refresh` 接口继续可用
- 所有现有接口正常工作

OAuth token 和普通 token 的区别：
- OAuth token 包含 `client_id` 和 `scope` 字段
- 普通 token 不包含这些字段
- 中间件会自动识别并处理两种类型的 token
