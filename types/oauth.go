package types

import "time"

// OAuthClient OAuth 客户端
type OAuthClient struct {
	ClientID         string    `json:"client_id"`
	ClientSecretHash string    `json:"client_secret_hash"` // bcrypt 加密后的密钥
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	RedirectURLs     []string  `json:"redirect_urls"`
	AllowedScopes    []string  `json:"allowed_scopes"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        string    `json:"created_by"` // 创建者 openid
	Status           string    `json:"status"`     // active, suspended
	RateLimit        int       `json:"rate_limit"` // 每小时请求次数限制
}

// OAuthClientResponse 客户端响应（用于列表展示）
type OAuthClientResponse struct {
	ClientID          string     `json:"client_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	AuthorizedUsers   int        `json:"authorized_users"`               // 授权用户数
	LastTokenIssuedAt *time.Time `json:"last_token_issued_at,omitempty"` // 最后一次签发 token 的时间
}

// OAuthAuthorizationCode 授权码
type OAuthAuthorizationCode struct {
	Code                string    `json:"code"`
	UID                 string    `json:"uid"`
	ClientID            string    `json:"client_id"`
	RedirectURL         string    `json:"redirect_url"`
	Scope               string    `json:"scope"` // 空格分隔
	CodeChallenge       string    `json:"code_challenge,omitempty"`
	CodeChallengeMethod string    `json:"code_challenge_method,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	Used                bool      `json:"used"`
}

// OAuthSession 授权会话（用于授权页面）
type OAuthSession struct {
	SessionID           string    `json:"session_id"`
	ClientID            string    `json:"client_id"`
	RedirectURL         string    `json:"redirect_url"`
	Scope               string    `json:"scope"`
	State               string    `json:"state,omitempty"`
	CodeChallenge       string    `json:"code_challenge,omitempty"`
	CodeChallengeMethod string    `json:"code_challenge_method,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

// OAuthUserGrant 用户授权记录
type OAuthUserGrant struct {
	ClientName string    `json:"client_name"`
	Scope      []string  `json:"scope"`
	GrantedAt  time.Time `json:"granted_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	TokenCount int       `json:"token_count"` // 已签发 token 次数
}

// OAuthAuditLog 审计日志
type OAuthAuditLog struct {
	Event     string    `json:"event"`
	ClientID  string    `json:"client_id"`
	UID       string    `json:"uid,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent,omitempty"`
	Details   string    `json:"details,omitempty"`
}

// OAuthTokenRequest Token 请求
type OAuthTokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type"`
	Code         string `json:"code,omitempty" form:"code"`
	RedirectURL  string `json:"redirect_url,omitempty" form:"redirect_url"`
	ClientID     string `json:"client_id" form:"client_id"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
	RefreshToken string `json:"refresh_token,omitempty" form:"refresh_token"`
	CodeVerifier string `json:"code_verifier,omitempty" form:"code_verifier"` // PKCE
}

// OAuthTokenResponse Token 响应
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"` // Bearer
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// OAuthErrorResponse OAuth 错误响应
type OAuthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// OAuthUserInfo 用户信息响应（/oauth/userinfo）
type OAuthUserInfo struct {
	Sub     string `json:"sub"`               // 用户 openid
	Name    string `json:"name,omitempty"`    // 昵称
	Picture string `json:"picture,omitempty"` // 头像
	Email   string `json:"email,omitempty"`   // 邮箱
	Role    string `json:"role,omitempty"`    // 角色
}

// RegisterOAuthClientRequest 注册客户端请求
type RegisterOAuthClientRequest struct {
	Name          string   `json:"name" binding:"required"`
	Description   string   `json:"description"`
	RedirectURLs  []string `json:"redirect_urls" binding:"required,min=1"`
	AllowedScopes []string `json:"allowed_scopes" binding:"required,min=1"`
}

// RegisterOAuthClientResponse 注册客户端响应
type RegisterOAuthClientResponse struct {
	ClientID      string    `json:"client_id"`
	ClientSecret  string    `json:"client_secret"` // 只在注册时显示一次
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	RedirectURLs  []string  `json:"redirect_urls"`
	AllowedScopes []string  `json:"allowed_scopes"`
	CreatedAt     time.Time `json:"created_at"`
}

// UpdateOAuthClientRequest 更新客户端请求
type UpdateOAuthClientRequest struct {
	Name          string   `json:"name,omitempty"`
	Description   string   `json:"description,omitempty"`
	RedirectURLs  []string `json:"redirect_urls,omitempty"`
	AllowedScopes []string `json:"allowed_scopes,omitempty"`
	Status        string   `json:"status,omitempty"` // active, suspended
}
