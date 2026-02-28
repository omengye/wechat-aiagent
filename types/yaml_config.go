package types

// Project 项目配置
type Project struct {
	ProjectId    string              `yaml:"projectId"`
	ProjectName  string              `yaml:"projectName"`
	TmpStr       string              `yaml:"tmpStr"`
	OAuthClients []OAuthClientConfig `yaml:"oauthClients,omitempty"` // OAuth 客户端配置（可选）
}

// OAuthClientConfig OAuth 客户端配置
type OAuthClientConfig struct {
	ClientID      string   `yaml:"clientId"`      // 客户端 ID
	ClientSecret  string   `yaml:"clientSecret"`  // 客户端密钥（明文，启动时会加密）
	Name          string   `yaml:"name"`          // 应用名称
	Description   string   `yaml:"description"`   // 应用描述
	RedirectURLs  []string `yaml:"redirectUrls"`  // 回调地址白名单
	AllowedScopes []string `yaml:"allowedScopes"` // 允许的 scope 列表
	RateLimit     int      `yaml:"rateLimit"`     // 每小时请求限制（0表示使用默认值1000）
}

// Config 应用配置
type Config struct {
	Wechat struct {
		AppId               string    `yaml:"appId"`
		AppSecret           string    `yaml:"appSecret"`
		Token               string    `yaml:"token"`
		Qrcode              []Project `yaml:"qrcode"`
		QrcodeExpire        int       `yaml:"qrcodeExpire"`        // 单位：秒
		QrcodeDefaultExpire int       `yaml:"qrcodeDefaultExpire"` // 临时二维码默认过期时间，单位：秒
		TemplateId          string    `yaml:"templateId"`          // 模板消息ID
		SuperAdmin          string    `yaml:"superAdmin"`          // 超级管理员openid
	} `yaml:"wechat"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
	} `yaml:"redis"`

	Jwt struct {
		Secret        string `yaml:"secret"`
		Expire        int64  `yaml:"expire"`        // 单位：秒
		RefreshExpire int64  `yaml:"refreshExpire"` // 单位：秒
	} `yaml:"jwt"`

	Server struct {
		Port            string   `yaml:"port"`            // 服务器监听端口，如 ":8443"
		AllowedOrigins  []string `yaml:"allowedOrigins"`  // WebSocket允许的来源白名单
		CorsOrigins     []string `yaml:"corsOrigins"`     // CORS跨域白名单
		MaxMessageSize  int      `yaml:"maxMessageSize"`  // WebSocket消息最大大小，单位：字节
		MaxProjectIDLen int      `yaml:"maxProjectIdLen"` // ProjectID最大长度
		RateLimitRate   int      `yaml:"rateLimitRate"`   // 速率限制：每个IP每秒请求数
		RateLimitBurst  int      `yaml:"rateLimitBurst"`  // 速率限制：突发请求数
	} `yaml:"server"`
}
