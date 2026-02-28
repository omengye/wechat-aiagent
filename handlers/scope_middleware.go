package handlers

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/sirupsen/logrus"
)

// RequireScope 中间件：验证 token 是否包含必需的 scope
// 必须在 AuthMiddleware 之后使用
func RequireScope(wChat *server.WeChat, requiredScopes ...string) app.HandlerFunc {
	scopeService := server.NewOAuthScopeService()

	return func(c context.Context, ctx *app.RequestContext) {
		// 从上下文中获取 scope（由 AuthMiddleware 设置）
		scopesInterface, exists := ctx.Get("scopes")
		if !exists {
			logrus.Warn("scopes not found in context, might be non-OAuth token")
			// 对于非 OAuth token（没有 scope 字段），允许通过
			// 因为现有的直接登录不包含 scope
			ctx.Next(c)
			return
		}

		scopes, ok := scopesInterface.([]string)
		if !ok {
			logrus.Error("invalid scopes type in context")
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "服务器内部错误",
			})
			ctx.Abort()
			return
		}

		// 检查是否包含所有必需的 scope
		if !scopeService.HasAllScopes(scopes, requiredScopes) {
			missingScopes := []string{}
			for _, required := range requiredScopes {
				if !scopeService.HasScope(scopes, required) {
					missingScopes = append(missingScopes, required)
				}
			}

			logrus.Warnf("missing required scopes: %v", missingScopes)
			ctx.JSON(consts.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": fmt.Sprintf("缺少必需的权限: %v", missingScopes),
			})
			ctx.Abort()
			return
		}

		// 验证通过，继续处理请求
		ctx.Next(c)
	}
}

// RequireAnyScope 中间件：验证 token 是否包含任意一个必需的 scope
func RequireAnyScope(wChat *server.WeChat, requiredScopes ...string) app.HandlerFunc {
	scopeService := server.NewOAuthScopeService()

	return func(c context.Context, ctx *app.RequestContext) {
		// 从上下文中获取 scope
		scopesInterface, exists := ctx.Get("scopes")
		if !exists {
			logrus.Warn("scopes not found in context, might be non-OAuth token")
			// 对于非 OAuth token，允许通过
			ctx.Next(c)
			return
		}

		scopes, ok := scopesInterface.([]string)
		if !ok {
			logrus.Error("invalid scopes type in context")
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "服务器内部错误",
			})
			ctx.Abort()
			return
		}

		// 检查是否包含任意一个必需的 scope
		if !scopeService.HasAnyScope(scopes, requiredScopes) {
			logrus.Warnf("missing any of required scopes: %v", requiredScopes)
			ctx.JSON(consts.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": fmt.Sprintf("需要以下任意一个权限: %v", requiredScopes),
			})
			ctx.Abort()
			return
		}

		// 验证通过，继续处理请求
		ctx.Next(c)
	}
}
