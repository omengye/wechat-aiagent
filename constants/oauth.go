package constants

// OAuth 相关常量
const (
	// Redis Key 前缀
	RedisOAuthClientPrefix       = "oauth:client:"        // OAuth 客户端信息
	RedisOAuthCodePrefix         = "oauth:code:"          // 授权码
	RedisOAuthSessionPrefix      = "oauth:session:"       // 授权会话
	RedisOAuthUserGrantPrefix    = "oauth:user:grants:"   // 用户授权记录
	RedisOAuthRevokedGrantPrefix = "oauth:revoked_grant:" // 已撤销的授权
	RedisOAuthRevokedClientPrefix = "oauth:revoked_client:" // 已撤销的客户端
	RedisOAuthAuditPrefix        = "oauth:audit:"         // 审计日志
	RedisOAuthRateLimitPrefix    = "rate_limit:oauth:"    // 限流

	// OAuth 客户端状态
	OAuthClientStatusActive    = "active"
	OAuthClientStatusSuspended = "suspended"

	// OAuth 错误码（符合 RFC 6749）
	OAuthErrorInvalidRequest         = "invalid_request"
	OAuthErrorInvalidClient          = "invalid_client"
	OAuthErrorInvalidGrant           = "invalid_grant"
	OAuthErrorUnauthorizedClient     = "unauthorized_client"
	OAuthErrorUnsupportedGrantType   = "unsupported_grant_type"
	OAuthErrorInvalidScope           = "invalid_scope"
	OAuthErrorAccessDenied           = "access_denied"
	OAuthErrorServerError            = "server_error"

	// OAuth Grant Types
	OAuthGrantTypeAuthorizationCode = "authorization_code"
	OAuthGrantTypeRefreshToken      = "refresh_token"

	// OAuth Response Types
	OAuthResponseTypeCode = "code"

	// PKCE Code Challenge Methods
	PKCEMethodPlain = "plain"
	PKCEMethodS256  = "S256"

	// Scope 定义
	ScopeUserRead        = "user:read"         // 读取用户基本信息
	ScopeUserEmail       = "user:email"        // 读取用户邮箱
	ScopeUserRole        = "user:role"         // 读取用户角色
	ScopeMessageSend     = "message:send"      // 发送模板消息
	ScopeAdminUserRead   = "admin:user:read"   // 读取其他用户信息
	ScopeAdminUserManage = "admin:user:manage" // 管理用户
	ScopeOfflineAccess   = "offline_access"    // 长期访问权限

	// WebSocket 消息类型
	WSOAuthRedirectType = "REDIRECT" // OAuth 重定向指令

	// 审计事件类型
	AuditEventClientRegistered      = "client_registered"
	AuditEventClientSecretReset     = "client_secret_reset"
	AuditEventClientRevoked         = "client_revoked"
	AuditEventAuthorizationGranted  = "authorization_granted"
	AuditEventTokenIssued           = "token_issued"
	AuditEventTokenRefreshed        = "token_refreshed"
	AuditEventGrantRevoked          = "grant_revoked"
	AuditEventInvalidClientSecret   = "invalid_client_secret"
	AuditEventInvalidAuthCode       = "invalid_authorization_code"
	AuditEventRateLimitExceeded     = "rate_limit_exceeded"

	// 限流配置
	RateLimitTokenPerHour      = 1000 // Token 端点每小时请求限制
	RateLimitAuthorizePerMin   = 10   // 授权端点每分钟请求限制

	// 过期时间
	OAuthCodeExpireSeconds       = 600                // 授权码 10 分钟
	OAuthSessionExpireSeconds    = 600                // 授权会话 10 分钟
	OAuthAuditLogExpireSeconds   = 2592000            // 审计日志 30 天
	OAuthClientExpireSeconds     = 1 * 365 * 24 * 3600 // OAuth 客户端缓存（1年）
	OAuthUserGrantExpireSeconds  = 1 * 365 * 24 * 3600 // 用户授权记录缓存（1年）
	OAuthRevokedExpireSeconds    = 1 * 365 * 24 * 3600 // 撤销记录缓存（1年）
)

// ScopeDisplayNames Scope 显示名称映射
var ScopeDisplayNames = map[string]string{
	ScopeUserRead:        "读取你的基本信息（昵称、头像）",
	ScopeUserEmail:       "读取你的邮箱地址",
	ScopeUserRole:        "读取你的角色信息",
	ScopeMessageSend:     "代你发送模板消息",
	ScopeAdminUserRead:   "读取其他用户信息",
	ScopeAdminUserManage: "管理用户账号",
	ScopeOfflineAccess:   "长期访问权限（即使你离线）",
}

// ScopeRequiredRole Scope 所需的最低角色
var ScopeRequiredRole = map[string]string{
	ScopeUserRead:        RoleUser,
	ScopeUserEmail:       RoleUser,
	ScopeUserRole:        RoleUser,
	ScopeMessageSend:     RoleAdmin,
	ScopeAdminUserRead:   RoleAdmin,
	ScopeAdminUserManage: RoleSuperAdmin,
	ScopeOfflineAccess:   RoleUser,
}
