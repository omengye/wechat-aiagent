package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/sirupsen/logrus"
)

// CorsMiddleware CORS跨域中间件
// 支持配置化的白名单，提供灵活的跨域访问控制
func CorsMiddleware(allowedOrigins []string) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 优先使用浏览器自动设置的 Origin header
		origin := string(ctx.Request.Header.Peek("Origin"))

		// 如果 Origin 为空，尝试使用自定义的 X-Requested-Origin header
		if origin == "" {
			origin = string(ctx.Request.Header.Peek("X-Requested-Origin"))
			if origin != "" {
				logrus.Infof("CORS: Using X-Requested-Origin header: %s", origin)
			}
		}

		method := string(ctx.Request.Header.Method())

		// 检查是否允许该来源
		allowed := false

		// 如果白名单为空，默认允许所有来源（开发环境）
		if len(allowedOrigins) == 0 {
			allowed = true
			logrus.Warnf("CORS: No allowed origins configured, allowing all origins (development mode)")
		} else {
			// 检查白名单
			for _, allowedOrigin := range allowedOrigins {
				// 精确匹配
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			logrus.Warnf("CORS: Blocked request from unauthorized origin: %s", origin)
			// 返回 403 Forbidden，阻止请求处理
			ctx.AbortWithStatus(403)
			return
		}

		// 设置 CORS 响应头
		ctx.Response.Header.Set("Access-Control-Allow-Origin", origin)
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", "false")
		ctx.Response.Header.Set("Access-Control-Max-Age", constants.DefaultAccessControlMaxAge)

		// 处理预检请求（OPTIONS）
		if method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next(c)
	}
}
