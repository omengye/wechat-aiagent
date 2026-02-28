package main

import (
	"context"
	"flag"
	"os"

	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/omengye/wechat_aiagent/handlers"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/omengye/wechat_aiagent/types"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func loadYaml() types.Config {
	configFile := flag.String("f", "config.yaml", "配置文件路径")
	flag.Parse()

	data, err := os.ReadFile(*configFile)
	if err != nil {
		logrus.Fatalf("读取配置文件失败: %v", err)
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		logrus.Fatalf("解析配置文件失败: %v", err)
	}
	return config
}

func main() {
	config := loadYaml()
	ctx := context.Background()

	// 初始化项目配置
	server.InitProject(config.Wechat.Qrcode)

	// 创建服务实例
	h := server.NewHertzHttp(config.Server.Port)

	wChat := server.NewWechatServer(ctx, &config)

	// 初始化 OAuth 客户端（从配置文件加载）
	if err := server.InitOAuthClientsFromConfig(&config, wChat.UserService.Redis); err != nil {
		logrus.Fatalf("初始化 OAuth 客户端失败: %v", err)
	}
	logrus.Info("OAuth 客户端初始化完成")

	// 配置路由
	setupRoutes(h, wChat, &config)

	// 启动服务
	h.Run()
}

// setupRoutes 配置所有路由
func setupRoutes(h *server.HertzHttp, wChat *server.WeChat, config *types.Config) {
	// 微信消息接收接口（不需要CORS）
	h.Server.Any("/api/wechat/", adaptor.HertzHandler(wChat))

	// API 路由分组（统一应用 CORS）
	apiGroup := h.Server.Group("/api")
	origins := append(config.Server.CorsOrigins, config.Server.AllowedOrigins...)
	apiGroup.Use(handlers.CorsMiddleware(origins))

	// 公开接口（无需认证）
	apiGroup.GET("/ws", handlers.WSHandler(wChat))
	apiGroup.POST("/auth/refresh", handlers.RefreshTokenHandler(wChat))

	// 用户接口（需要登录）
	userGroup := apiGroup.Group("/user")
	userGroup.Use(
		handlers.CorsMiddleware(config.Server.CorsOrigins),
		handlers.AuthMiddleware(wChat))
	userGroup.POST("/token/validate", handlers.ValidateAccessTokenHandler(wChat))
	userGroup.GET("/role", handlers.GetMyRoleHandler(wChat))
	userGroup.GET("/info", handlers.GetUserInfoHandler(wChat))

	// 用户 OAuth 授权管理接口（需要登录）
	userGroup.GET("/oauth/grants", handlers.GetOAuthUserGrantsHandler(wChat))
	userGroup.DELETE("/oauth/grants/:client_id", handlers.RevokeOAuthUserGrantHandler(wChat))

	// 管理员接口（需要 admin 或 super_admin 角色）
	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(
		handlers.CorsMiddleware(config.Server.CorsOrigins),
		handlers.AuthMiddleware(wChat),
		handlers.RequireRole("admin", "super_admin"))
	adminGroup.POST("/template/send", handlers.SendTemplateMessageHandler(wChat))
	adminGroup.POST("/user/info", handlers.GetUserInfoByOpenidHandler(wChat))

	// 超级管理员接口（需要 super_admin 角色）
	superAdminGroup := apiGroup.Group("/super")
	superAdminGroup.Use(
		handlers.CorsMiddleware(config.Server.CorsOrigins),
		handlers.AuthMiddleware(wChat),
		handlers.RequireRole("super_admin"))
	superAdminGroup.POST("/admin/add", handlers.AddAdminHandler(wChat))
	superAdminGroup.POST("/admin/remove", handlers.RemoveAdminHandler(wChat))
	superAdminGroup.GET("/admin/list", handlers.ListAdminsHandler(wChat))

	// OAuth 客户端管理接口（超级管理员） - 仅保留查询接口
	superAdminGroup.GET("/oauth/client/list", handlers.ListOAuthClientsHandler(wChat))
	superAdminGroup.GET("/oauth/client/:client_id", handlers.GetOAuthClientHandler(wChat))

	// OAuth 授权接口（公开）
	oauthGroup := apiGroup.Group("/oauth")
	oauthGroup.Use(handlers.CorsMiddleware(config.Server.CorsOrigins))
	oauthGroup.GET("/authorize", handlers.OAuthAuthorizeHandler(wChat)) // 授权页面
	oauthGroup.POST("/token", handlers.OAuthTokenHandler(wChat))        // Token 端点

	// OAuth UserInfo 端点（需要认证）
	oauthGroup.GET("/userinfo",
		handlers.AuthMiddleware(wChat),
		handlers.OAuthUserInfoHandler(wChat))
}
