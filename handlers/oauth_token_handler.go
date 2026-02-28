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

// OAuthTokenHandler OAuth Token 端点
// POST /api/oauth/token
func OAuthTokenHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 解析请求
		var req types.OAuthTokenRequest
		if err := ctx.Bind(&req); err != nil {
			logrus.Errorf("bind oauth token request failed: %v", err)
			ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
				Error:            constants.OAuthErrorInvalidRequest,
				ErrorDescription: "请求参数错误",
			})
			return
		}

		// 根据 grant_type 分发处理
		switch req.GrantType {
		case constants.OAuthGrantTypeAuthorizationCode:
			handleAuthorizationCodeGrant(wChat, ctx, &req)
		case constants.OAuthGrantTypeRefreshToken:
			handleRefreshTokenGrant(wChat, ctx, &req)
		default:
			ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
				Error:            constants.OAuthErrorUnsupportedGrantType,
				ErrorDescription: "不支持的 grant_type",
			})
		}
	}
}

// handleAuthorizationCodeGrant 处理授权码换取 token
func handleAuthorizationCodeGrant(wChat *server.WeChat, ctx *app.RequestContext, req *types.OAuthTokenRequest) {
	// 1. 验证必需参数
	if req.Code == "" || req.ClientID == "" || req.ClientSecret == "" || req.RedirectURL == "" {
		ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidRequest,
			ErrorDescription: "缺少必需参数",
		})
		return
	}

	ip := ctx.ClientIP()
	userAgent := string(ctx.UserAgent())

	// 2. 验证客户端凭证
	clientService := server.NewOAuthClientService(wChat.UserService.Redis)
	if err := clientService.VerifyClientCredentials(req.ClientID, req.ClientSecret); err != nil {
		logrus.Errorf("verify client credentials failed: %v", err)

		// 记录审计日志
		auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
		auditService.LogInvalidClientSecret(req.ClientID, ip)

		ctx.JSON(consts.StatusUnauthorized, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidClient,
			ErrorDescription: "客户端认证失败",
		})
		return
	}

	// 3. 获取授权码信息
	codeService := server.NewOAuthCodeService(wChat.UserService.Redis)
	authCode, err := codeService.GetAuthorizationCode(req.Code)
	if err != nil {
		logrus.Errorf("get authorization code failed: %v", err)

		// 记录审计日志
		auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
		auditService.LogInvalidAuthCode(req.ClientID, ip, req.Code)

		ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "授权码无效或已过期",
		})
		return
	}

	// 4. 验证授权码
	if authCode.Used {
		logrus.Warnf("authorization code already used: %s", req.Code)
		ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "授权码已被使用",
		})
		return
	}

	if authCode.ClientID != req.ClientID {
		logrus.Warnf("client_id mismatch for code %s", req.Code)
		ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "授权码与客户端不匹配",
		})
		return
	}

	if authCode.RedirectURL != req.RedirectURL {
		logrus.Warnf("redirect_url mismatch for code %s", req.Code)
		ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "redirect_url 不匹配",
		})
		return
	}

	// 5. PKCE 验证
	if authCode.CodeChallenge != "" {
		if err := codeService.VerifyPKCE(req.CodeVerifier, authCode.CodeChallenge, authCode.CodeChallengeMethod); err != nil {
			logrus.Errorf("PKCE verification failed: %v", err)
			ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
				Error:            constants.OAuthErrorInvalidGrant,
				ErrorDescription: "PKCE 验证失败",
			})
			return
		}
	}

	// 6. 获取用户角色
	role, err := wChat.UserService.GetUserRole(authCode.UID, wChat.Config.Wechat.SuperAdmin)
	if err != nil {
		logrus.Warnf("get user role failed: %v, using default role", err)
		role = constants.RoleUser
	}

	// 7. 验证用户角色是否满足 scope 要求
	scopes := server.ParseScopes(authCode.Scope)
	scopeService := server.NewOAuthScopeService()
	if err := scopeService.ValidateScopesForRole(scopes, role); err != nil {
		logrus.Errorf("user role does not satisfy scopes: %v", err)
		ctx.JSON(consts.StatusForbidden, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidScope,
			ErrorDescription: err.Error(),
		})
		return
	}

	// 8. 生成 OAuth Token
	accessToken, err := server.GenerateOAuthAccessToken(
		authCode.UID,
		wChat.Config.Jwt.Secret,
		wChat.Config.Jwt.Expire,
		role,
		scopes,
		req.ClientID,
	)
	if err != nil {
		logrus.Errorf("generate access token failed: %v", err)
		ctx.JSON(consts.StatusInternalServerError, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorServerError,
			ErrorDescription: "生成 token 失败",
		})
		return
	}

	// 9. 生成 refresh token（仅当 scope 包含 offline_access）
	var refreshToken string
	if scopeService.HasScope(scopes, constants.ScopeOfflineAccess) {
		refreshToken, err = server.GenerateOAuthRefreshToken(
			authCode.UID,
			wChat.Config.Jwt.Secret,
			wChat.Config.Jwt.RefreshExpire,
			role,
			scopes,
			req.ClientID,
		)
		if err != nil {
			logrus.Errorf("generate refresh token failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, types.OAuthErrorResponse{
				Error:            constants.OAuthErrorServerError,
				ErrorDescription: "生成 refresh token 失败",
			})
			return
		}
	}

	// 10. 标记授权码为已使用
	if err := codeService.MarkAuthorizationCodeAsUsed(req.Code); err != nil {
		logrus.Errorf("mark authorization code as used failed: %v", err)
	}

	// 11. 记录用户授权
	client, _ := clientService.GetClient(req.ClientID)
	clientName := req.ClientID
	if client != nil {
		clientName = client.Name
	}
	if err := codeService.RecordUserGrant(authCode.UID, req.ClientID, clientName, authCode.Scope); err != nil {
		logrus.Errorf("record user grant failed: %v", err)
	}

	// 12. 记录审计日志
	auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
	auditService.LogTokenIssued(req.ClientID, authCode.UID, ip, userAgent)

	// 13. 返回 token
	response := types.OAuthTokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    wChat.Config.Jwt.Expire,
		RefreshToken: refreshToken,
		Scope:        authCode.Scope,
	}

	ctx.JSON(consts.StatusOK, response)
}

// handleRefreshTokenGrant 处理 refresh token 刷新
func handleRefreshTokenGrant(wChat *server.WeChat, ctx *app.RequestContext, req *types.OAuthTokenRequest) {
	// 1. 验证必需参数
	if req.RefreshToken == "" || req.ClientID == "" || req.ClientSecret == "" {
		ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidRequest,
			ErrorDescription: "缺少必需参数",
		})
		return
	}

	ip := ctx.ClientIP()
	userAgent := string(ctx.UserAgent())

	// 2. 验证客户端凭证
	clientService := server.NewOAuthClientService(wChat.UserService.Redis)
	if err := clientService.VerifyClientCredentials(req.ClientID, req.ClientSecret); err != nil {
		logrus.Errorf("verify client credentials failed: %v", err)

		// 记录审计日志
		auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
		auditService.LogInvalidClientSecret(req.ClientID, ip)

		ctx.JSON(consts.StatusUnauthorized, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidClient,
			ErrorDescription: "客户端认证失败",
		})
		return
	}

	// 3. 解析 refresh token
	claims, err := server.ParseToken(req.RefreshToken, wChat.Config.Jwt.Secret)
	if err != nil {
		logrus.Errorf("parse refresh token failed: %v", err)
		ctx.JSON(consts.StatusUnauthorized, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "refresh token 无效或已过期",
		})
		return
	}

	// 4. 验证 token 类型
	if claims.TokenType != constants.TokenTypeRefresh {
		logrus.Warnf("invalid token type: %s", claims.TokenType)
		ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "token 类型错误",
		})
		return
	}

	// 5. 验证 client_id
	if claims.ClientID != req.ClientID {
		logrus.Warnf("client_id mismatch in refresh token")
		ctx.JSON(consts.StatusBadRequest, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "token 与客户端不匹配",
		})
		return
	}

	// 6. 检查用户黑名单
	if isBlacklisted, _ := wChat.UserService.IsTokenBlacklisted(claims.Uid); isBlacklisted {
		ctx.JSON(consts.StatusUnauthorized, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "用户已被拉黑",
		})
		return
	}

	// 7. 检查 refresh token 是否已被撤销
	if isRevoked, _ := wChat.UserService.IsRefreshTokenRevoked(req.RefreshToken); isRevoked {
		logrus.Warnf("refresh token has been revoked")
		// 安全措施：标记用户为黑名单
		wChat.UserService.Redis.Set(constants.RedisTokenBlacklistPrefix+claims.Uid, "revoked", wChat.Config.Jwt.RefreshExpire)

		ctx.JSON(consts.StatusUnauthorized, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorInvalidGrant,
			ErrorDescription: "refresh token 已被撤销",
		})
		return
	}

	// 8. 获取用户最新角色
	role, err := wChat.UserService.GetUserRole(claims.Uid, wChat.Config.Wechat.SuperAdmin)
	if err != nil {
		logrus.Warnf("get user role failed: %v, using role from token", err)
		role = claims.Role
	}

	// 9. 生成新的 token pair
	newAccessToken, err := server.GenerateOAuthAccessToken(
		claims.Uid,
		wChat.Config.Jwt.Secret,
		wChat.Config.Jwt.Expire,
		role,
		claims.Scope,
		req.ClientID,
	)
	if err != nil {
		logrus.Errorf("generate new access token failed: %v", err)
		ctx.JSON(consts.StatusInternalServerError, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorServerError,
			ErrorDescription: "生成 token 失败",
		})
		return
	}

	newRefreshToken, err := server.GenerateOAuthRefreshToken(
		claims.Uid,
		wChat.Config.Jwt.Secret,
		wChat.Config.Jwt.RefreshExpire,
		role,
		claims.Scope,
		req.ClientID,
	)
	if err != nil {
		logrus.Errorf("generate new refresh token failed: %v", err)
		ctx.JSON(consts.StatusInternalServerError, types.OAuthErrorResponse{
			Error:            constants.OAuthErrorServerError,
			ErrorDescription: "生成 refresh token 失败",
		})
		return
	}

	// 10. 撤销旧的 refresh token
	remainingTime := claims.ExpiresAt.Unix() - claims.IssuedAt.Unix()
	if err := wChat.UserService.RevokeRefreshToken(req.RefreshToken, remainingTime); err != nil {
		logrus.Errorf("revoke old refresh token failed: %v", err)
	}

	// 11. 记录审计日志
	auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
	auditService.LogTokenRefreshed(req.ClientID, claims.Uid, ip, userAgent)

	// 12. 返回新 token
	response := types.OAuthTokenResponse{
		AccessToken:  newAccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    wChat.Config.Jwt.Expire,
		RefreshToken: newRefreshToken,
		Scope:        server.JoinScopes(claims.Scope),
	}

	ctx.JSON(consts.StatusOK, response)
}
