package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/omengye/wechat_aiagent/types"
	"github.com/sirupsen/logrus"
)

// GetOAuthUserGrantsHandler 获取用户的授权列表
// GET /api/user/oauth/grants
func GetOAuthUserGrantsHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, _ := ctx.Get("uid")
		uidStr := uid.(string)

		codeService := server.NewOAuthCodeService(wChat.UserService.Redis)
		grants, err := codeService.GetUserGrants(uidStr)
		if err != nil {
			logrus.Errorf("get user grants failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "获取授权列表失败",
			})
			return
		}

		// 转换为列表格式
		grantList := make([]map[string]interface{}, 0, len(grants))
		for clientID, grant := range grants {
			grantList = append(grantList, map[string]interface{}{
				"client_id":   clientID,
				"client_name": grant.ClientName,
				"scope":       grant.Scope,
				"granted_at":  grant.GrantedAt,
				"last_used_at": grant.LastUsedAt,
				"token_count": grant.TokenCount,
			})
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "获取授权列表成功",
			"data":    grantList,
		})
	}
}

// RevokeOAuthUserGrantHandler 撤销用户对某个客户端的授权
// DELETE /api/user/oauth/grants/:client_id
func RevokeOAuthUserGrantHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		uid, _ := ctx.Get("uid")
		uidStr := uid.(string)

		clientID := ctx.Param("client_id")
		if clientID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "client_id 不能为空",
			})
			return
		}

		codeService := server.NewOAuthCodeService(wChat.UserService.Redis)
		if err := codeService.RevokeUserGrant(uidStr, clientID); err != nil {
			logrus.Errorf("revoke user grant failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "撤销授权失败",
			})
			return
		}

		// 记录审计日志
		auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
		ip := ctx.ClientIP()
		if err := auditService.LogGrantRevoked(clientID, uidStr, ip); err != nil {
			logrus.Errorf("log audit failed: %v", err)
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "授权已撤销，该应用的所有 token 将失效",
		})
	}
}

// OAuthUserInfoHandler OAuth 标准的 UserInfo 端点
// GET /api/oauth/userinfo
func OAuthUserInfoHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从上下文获取用户信息
		uid, _ := ctx.Get("uid")
		uidStr := uid.(string)

		role, _ := ctx.Get("role")
		roleStr := role.(string)

		scopesInterface, exists := ctx.Get("scopes")
		if !exists {
			// 非 OAuth token，返回错误
			ctx.JSON(consts.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "此端点仅支持 OAuth token",
			})
			return
		}

		scopes := scopesInterface.([]string)
		scopeService := server.NewOAuthScopeService()

		// 检查是否有 user:read scope
		if !scopeService.HasScope(scopes, constants.ScopeUserRead) {
			ctx.JSON(consts.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "缺少必需的权限: user:read",
			})
			return
		}

		// 构建响应
		userInfo := types.OAuthUserInfo{
			Sub: uidStr, // OAuth 标准字段
		}

		// 根据 scope 返回不同的字段
		if scopeService.HasScope(scopes, constants.ScopeUserRead) {
			// 获取用户信息（昵称、头像）
			// 这里简化实现，实际应该从数据库或微信 API 获取
			userInfo.Name = "微信用户" // TODO: 从微信 API 获取真实昵称
			userInfo.Picture = ""      // TODO: 从微信 API 获取头像
		}

		if scopeService.HasScope(scopes, constants.ScopeUserEmail) {
			// TODO: 如果系统中有邮箱信息，填充邮箱字段
			userInfo.Email = "" // 根据实际情况填充
		}

		if scopeService.HasScope(scopes, constants.ScopeUserRole) {
			userInfo.Role = roleStr
		}

		ctx.JSON(consts.StatusOK, userInfo)
	}
}
