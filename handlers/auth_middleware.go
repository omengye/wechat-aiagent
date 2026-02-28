package handlers

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware 认证中间件，验证 access token
func AuthMiddleware(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从 Header 中获取 Authorization
		authHeader := string(ctx.GetHeader("Authorization"))
		if authHeader == "" {
			logrus.Warn("missing Authorization header")
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "缺少 Authorization header",
			})
			ctx.Abort()
			return
		}

		// 解析 Bearer token
		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}

		if tokenString == "" {
			logrus.Warn("empty token")
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "token 不能为空",
			})
			ctx.Abort()
			return
		}

		// 验证 token
		claims, err := server.ParseToken(tokenString, wChat.Config.Jwt.Secret)
		if err != nil {
			logrus.Warnf("invalid token: %v", err)
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "token 无效或已过期",
			})
			ctx.Abort()
			return
		}

		// 检查 token 类型
		if claims.TokenType != "access" {
			logrus.Warnf("invalid token type: %s", claims.TokenType)
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "token 类型错误",
			})
			ctx.Abort()
			return
		}

		// 检查用户是否在黑名单中
		if isBlacklisted, _ := wChat.UserService.IsTokenBlacklisted(claims.Uid); isBlacklisted {
			logrus.Warnf("user %s is blacklisted", claims.Uid)
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "token 已被撤销",
			})
			ctx.Abort()
			return
		}

		// OAuth 相关检查
		if claims.ClientID != "" {
			// 这是 OAuth token，需要额外验证

			// 1. 检查客户端是否被撤销
			clientService := server.NewOAuthClientService(wChat.UserService.Redis)
			if revoked, _ := clientService.IsClientRevoked(claims.ClientID); revoked {
				logrus.Warnf("OAuth client %s is revoked", claims.ClientID)
				ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
					"code":    401,
					"message": "OAuth 客户端已被撤销",
				})
				ctx.Abort()
				return
			}

			// 2. 检查用户是否撤销了对该客户端的授权
			codeService := server.NewOAuthCodeService(wChat.UserService.Redis)
			if revoked, _ := codeService.IsGrantRevoked(claims.Uid, claims.ClientID); revoked {
				logrus.Warnf("user %s revoked grant for client %s", claims.Uid, claims.ClientID)
				ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
					"code":    401,
					"message": "用户已撤销授权",
				})
				ctx.Abort()
				return
			}
		}

		// 将用户信息存入上下文
		ctx.Set("uid", claims.Uid)
		ctx.Set("token_type", claims.TokenType)
		ctx.Set("role", claims.Role)
		ctx.Set("scopes", claims.Scope)       // OAuth scope
		ctx.Set("client_id", claims.ClientID) // OAuth client_id

		// 继续处理请求
		ctx.Next(c)
	}
}
