package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/omengye/wechat_aiagent/types"
	"github.com/sirupsen/logrus"
)

// RegisterOAuthClientHandler 注册 OAuth 客户端（超级管理员）
func RegisterOAuthClientHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 获取操作者 UID
		uid, _ := ctx.Get("uid")
		uidStr := uid.(string)

		// 解析请求
		var req types.RegisterOAuthClientRequest
		if err := ctx.BindJSON(&req); err != nil {
			logrus.Errorf("bind register oauth client request failed: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "请求参数错误",
			})
			return
		}

		// 验证 scope
		scopeService := server.NewOAuthScopeService()
		if err := scopeService.ValidateScopes(req.AllowedScopes); err != nil {
			logrus.Errorf("invalid scopes: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": err.Error(),
			})
			return
		}

		// 注册客户端
		clientService := server.NewOAuthClientService(wChat.UserService.Redis)
		response, err := clientService.RegisterClient(
			req.Name,
			req.Description,
			req.RedirectURLs,
			req.AllowedScopes,
			uidStr,
		)
		if err != nil {
			logrus.Errorf("register oauth client failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "注册客户端失败",
			})
			return
		}

		// 记录审计日志
		auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
		ip := ctx.ClientIP()
		if err := auditService.LogClientRegistered(response.ClientID, uidStr, ip); err != nil {
			logrus.Errorf("log audit failed: %v", err)
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "注册成功，请立即保存 client_secret，此密钥只显示一次",
			"data":    response,
		})
	}
}

// ListOAuthClientsHandler 查看 OAuth 客户端列表（超级管理员）
func ListOAuthClientsHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 注意：这里简化实现，实际生产环境应使用专门的数据结构存储客户端列表
		// 由于 Redis 的限制，这里返回空列表提示
		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "获取客户端列表成功",
			"data":    []interface{}{},
			"note":    "完整实现需要使用 Redis SCAN 或专门的索引结构",
		})
	}
}

// GetOAuthClientHandler 获取单个 OAuth 客户端详情（超级管理员）
func GetOAuthClientHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		clientID := ctx.Param("client_id")
		if clientID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "client_id 不能为空",
			})
			return
		}

		clientService := server.NewOAuthClientService(wChat.UserService.Redis)
		client, err := clientService.GetClient(clientID)
		if err != nil {
			logrus.Errorf("get oauth client failed: %v", err)
			ctx.JSON(consts.StatusNotFound, map[string]interface{}{
				"code":    404,
				"message": "客户端不存在",
			})
			return
		}

		// 不返回密钥哈希
		response := map[string]interface{}{
			"client_id":      client.ClientID,
			"name":           client.Name,
			"description":    client.Description,
			"redirect_urls":  client.RedirectURLs,
			"allowed_scopes": client.AllowedScopes,
			"created_at":     client.CreatedAt,
			"created_by":     client.CreatedBy,
			"status":         client.Status,
			"rate_limit":     client.RateLimit,
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "获取客户端成功",
			"data":    response,
		})
	}
}

// UpdateOAuthClientHandler 更新 OAuth 客户端（超级管理员）
func UpdateOAuthClientHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		clientID := ctx.Param("client_id")
		if clientID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "client_id 不能为空",
			})
			return
		}

		var req types.UpdateOAuthClientRequest
		if err := ctx.BindJSON(&req); err != nil {
			logrus.Errorf("bind update oauth client request failed: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "请求参数错误",
			})
			return
		}

		// 如果更新了 scopes，验证其有效性
		if len(req.AllowedScopes) > 0 {
			scopeService := server.NewOAuthScopeService()
			if err := scopeService.ValidateScopes(req.AllowedScopes); err != nil {
				logrus.Errorf("invalid scopes: %v", err)
				ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
					"code":    400,
					"message": err.Error(),
				})
				return
			}
		}

		clientService := server.NewOAuthClientService(wChat.UserService.Redis)
		if err := clientService.UpdateClient(clientID, &req); err != nil {
			logrus.Errorf("update oauth client failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "更新客户端失败",
			})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "更新客户端成功",
		})
	}
}

// DeleteOAuthClientHandler 删除 OAuth 客户端（超级管理员）
func DeleteOAuthClientHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		clientID := ctx.Param("client_id")
		if clientID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "client_id 不能为空",
			})
			return
		}

		// 获取操作者 UID
		uid, _ := ctx.Get("uid")
		uidStr := uid.(string)

		clientService := server.NewOAuthClientService(wChat.UserService.Redis)
		if err := clientService.DeleteClient(clientID); err != nil {
			logrus.Errorf("delete oauth client failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "删除客户端失败",
			})
			return
		}

		// 记录审计日志
		auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
		ip := ctx.ClientIP()
		if err := auditService.LogClientRevoked(clientID, uidStr, ip); err != nil {
			logrus.Errorf("log audit failed: %v", err)
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "客户端已撤销，所有已签发的 token 将失效",
		})
	}
}

// ResetOAuthClientSecretHandler 重置 OAuth 客户端密钥（超级管理员）
func ResetOAuthClientSecretHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		clientID := ctx.Param("client_id")
		if clientID == "" {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "client_id 不能为空",
			})
			return
		}

		// 获取操作者 UID
		uid, _ := ctx.Get("uid")
		uidStr := uid.(string)

		clientService := server.NewOAuthClientService(wChat.UserService.Redis)
		newSecret, err := clientService.ResetClientSecret(clientID)
		if err != nil {
			logrus.Errorf("reset oauth client secret failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "重置密钥失败",
			})
			return
		}

		// 记录审计日志
		auditService := server.NewOAuthAuditService(wChat.UserService.Redis)
		ip := ctx.ClientIP()
		if err := auditService.LogClientSecretReset(clientID, uidStr, ip); err != nil {
			logrus.Errorf("log audit failed: %v", err)
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "密钥已重置，请立即保存新密钥。警告：旧密钥立即失效",
			"data": map[string]interface{}{
				"client_secret": newSecret,
			},
		})
	}
}
