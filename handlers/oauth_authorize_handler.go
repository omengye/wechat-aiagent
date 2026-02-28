package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/sirupsen/logrus"
)

// OAuthAuthorizeHandler OAuth 授权请求处理器
// GET /api/oauth/authorize
func OAuthAuthorizeHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 1. 解析查询参数
		clientID := ctx.Query("client_id")
		redirectURL := ctx.Query("redirect_url")
		responseType := ctx.Query("response_type")
		scope := ctx.Query("scope")
		state := ctx.Query("state")
		codeChallenge := ctx.Query("code_challenge")
		codeChallengeMethod := ctx.Query("code_challenge_method")

		// 2. 验证必需参数
		if clientID == "" || redirectURL == "" || responseType == "" || scope == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":             constants.OAuthErrorInvalidRequest,
				"error_description": "缺少必需参数：client_id, redirect_url, response_type, scope",
			})
			return
		}

		// 3. 验证 response_type
		if responseType != constants.OAuthResponseTypeCode {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":             constants.OAuthErrorInvalidRequest,
				"error_description": "不支持的 response_type，只支持 'code'",
			})
			return
		}

		// 4. 验证客户端
		clientService := server.NewOAuthClientService(wChat.UserService.Redis)
		client, err := clientService.GetClient(clientID)
		if err != nil {
			logrus.Errorf("get oauth client failed: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":             constants.OAuthErrorInvalidClient,
				"error_description": "客户端不存在",
			})
			return
		}

		// 5. 检查客户端状态
		if client.Status != constants.OAuthClientStatusActive {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":             constants.OAuthErrorUnauthorizedClient,
				"error_description": "客户端已被暂停",
			})
			return
		}

		// 6. 验证 redirect_url
		if err := clientService.ValidateRedirectURL(clientID, redirectURL); err != nil {
			logrus.Errorf("invalid redirect_url: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":             constants.OAuthErrorInvalidRequest,
				"error_description": "redirect_url 不在白名单中",
			})
			return
		}

		// 7. 验证 scopes
		scopes := server.ParseScopes(scope)
		if err := clientService.ValidateScopes(clientID, scopes); err != nil {
			logrus.Errorf("invalid scopes: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"error":             constants.OAuthErrorInvalidScope,
				"error_description": err.Error(),
			})
			return
		}

		// 8. 创建授权会话
		codeService := server.NewOAuthCodeService(wChat.UserService.Redis)
		sessionID, err := codeService.CreateSession(
			clientID,
			redirectURL,
			scope,
			state,
			codeChallenge,
			codeChallengeMethod,
		)
		if err != nil {
			logrus.Errorf("create oauth session failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"error":             constants.OAuthErrorServerError,
				"error_description": "创建授权会话失败",
			})
			return
		}

		// 9. 返回授权信息 JSON
		scopeService := server.NewOAuthScopeService()
		scopeDisplayNames := scopeService.GetScopeDisplayNames(scopes)

		// 获取 projectId（优先使用配置了 OAuth 客户端的项目）
		projectID := ""
		for _, project := range wChat.Config.Wechat.Qrcode {
			if len(project.OAuthClients) > 0 {
				projectID = project.ProjectId
				break
			}
		}
		if projectID == "" && len(wChat.Config.Wechat.Qrcode) > 0 {
			// 如果没有配置 OAuth 客户端的项目，使用第一个项目
			projectID = wChat.Config.Wechat.Qrcode[0].ProjectId
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"session_id":   sessionID,
			"client_name":  client.Name,
			"scope_names":  scopeDisplayNames,
			"scopes":       scope,
			"project_id":   projectID,
			"redirect_url": redirectURL,
			"state":        state,
		})
	}
}
