package handlers

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/sirupsen/logrus"
)

// RequireRole 验证用户角色（支持多个允许的角色）
// 使用示例：
//   RequireRole("admin", "super_admin") - 允许管理员和超级管理员访问
//   RequireRole("super_admin") - 只允许超级管理员访问
//   RequireRole("user", "admin", "super_admin") - 所有角色都可以访问
func RequireRole(allowedRoles ...string) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从context获取role（由AuthMiddleware设置）
		roleValue, exists := ctx.Get("role")
		if !exists {
			logrus.Warn("role not found in context, authentication middleware may not be applied")
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "未找到用户角色信息",
			})
			ctx.Abort()
			return
		}

		// 检查角色是否在允许列表中
		roleStr, ok := roleValue.(string)
		if !ok {
			logrus.Warn("role in context is not a string")
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "角色信息格式错误",
			})
			ctx.Abort()
			return
		}

		for _, allowedRole := range allowedRoles {
			if roleStr == allowedRole {
				// 角色匹配，继续处理请求
				ctx.Next(c)
				return
			}
		}

		// 角色不匹配，返回403 Forbidden
		uid, _ := ctx.Get("uid")
		logrus.Warnf("user %s with role %s attempted to access resource requiring roles: %v",
			uid, roleStr, allowedRoles)
		ctx.JSON(consts.StatusForbidden, map[string]interface{}{
			"code":    403,
			"message": "权限不足，需要" + strings.Join(allowedRoles, "或") + "角色",
		})
		ctx.Abort()
	}
}
