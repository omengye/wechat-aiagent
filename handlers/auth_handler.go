package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/sirupsen/logrus"
)

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *server.UserToken `json:"data,omitempty"`
}

// RefreshTokenHandler 处理刷新令牌的请求
// 使用 Token Rotation 策略：每次刷新都生成新的 access token 和 refresh token
// 旧的 refresh token 立即失效，提高安全性
func RefreshTokenHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req RefreshTokenRequest

		// 1. 解析请求参数
		if err := ctx.BindAndValidate(&req); err != nil {
			logrus.Warnf("invalid request parameters: %v", err)
			response := RefreshTokenResponse{
				Code:    400,
				Message: "请求参数错误",
			}
			ctx.JSON(consts.StatusBadRequest, response)
			return
		}

		// 2. 验证 refresh token 并生成新的令牌对
		newToken, err := wChat.UserService.RefreshUserToken(req.RefreshToken, wChat.Config)
		if err != nil {
			logrus.Warnf("refresh token failed: %v", err)

			// 根据错误类型返回不同的状态码
			statusCode := consts.StatusUnauthorized
			message := err.Error()

			response := RefreshTokenResponse{
				Code:    statusCode,
				Message: message,
			}
			ctx.JSON(statusCode, response)
			return
		}

		// 3. 返回成功响应
		response := RefreshTokenResponse{
			Code:    200,
			Message: "success",
			Data:    newToken,
		}

		logrus.Infof("token refreshed successfully for user")
		ctx.JSON(consts.StatusOK, response)
	}
}

// ValidateAccessTokenHandler 验证 access token 的有效性（可选实现）
func ValidateAccessTokenHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从 Header 中获取 token
		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) == 0 {
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "缺少 Authorization header",
			})
			return
		}

		// 解析 Bearer token
		tokenString := string(authHeader)
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// 验证 token
		claims, err := server.ParseToken(tokenString, wChat.Config.Jwt.Secret)
		if err != nil {
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "token 无效或已过期",
			})
			return
		}

		// 检查是否是 access token
		if claims.TokenType != "access" {
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "token 类型错误",
			})
			return
		}

		// 检查黑名单
		if isBlacklisted, _ := wChat.UserService.IsTokenBlacklisted(claims.Uid); isBlacklisted {
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "token 已被撤销",
			})
			return
		}

		// 返回 token 信息
		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "token 有效",
			"data": map[string]interface{}{
				"uid":        claims.Uid,
				"token_type": claims.TokenType,
				"expires_at": claims.ExpiresAt.Unix(),
			},
		})
	}
}
