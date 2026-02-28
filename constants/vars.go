package constants

const (
	TokenTypeAccess                = "access"
	TokenTypeRefresh               = "refresh"
	RedisProjectQrcodePrefix       = "redis_project_qrcode-"
	RedisAccessTokenPrefix         = "redis_access_token-"
	RedisTokenBlacklistPrefix      = "token_blacklist-"
	RedisRevokedRefreshTokenPrefix = "revoked_refresh_token-"
	RedisAdminUserPrefix           = "admin_user-" // 管理员标识，值为角色(admin/super_admin)

	WSQrcodeType = "QRCODE"
	WSTokenType  = "TOKEN"

	WeChatQrScenePrefix = "qrscene_"

	DefaultAccessControlMaxAge = "5" // 5s

	// 用户角色
	RoleUser       = "user"        // 普通用户
	RoleAdmin      = "admin"       // 管理员
	RoleSuperAdmin = "super_admin" // 超级管理员

	// 过期时间
	AdminRoleExpireSeconds = 1 * 365 * 24 * 3600 // 管理员角色过期时间（1年）
)
