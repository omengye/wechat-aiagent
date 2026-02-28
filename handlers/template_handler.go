package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/sirupsen/logrus"
)

// SendTemplateMessageHandler 发送模板消息的处理函数
func SendTemplateMessageHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从上下文获取认证用户信息
		uid, exists := ctx.Get("uid")
		if !exists {
			logrus.Warn("uid not found in context")
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "未认证",
			})
			return
		}

		var req server.TemplateMessageRequest

		// 解析请求参数
		if err := ctx.BindAndValidate(&req); err != nil {
			logrus.Warnf("invalid request parameters: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "请求参数错误: " + err.Error(),
			})
			return
		}

		// 记录请求日志（包含用户信息）
		logrus.Infof("user %s sending template message to %s", uid, req.ToUser)

		// 发送模板消息
		response, err := wChat.SendTemplateMessage(&req)
		if err != nil {
			logrus.Errorf("send template message failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, response)
			return
		}

		// 返回响应
		if response.Code == 200 {
			ctx.JSON(consts.StatusOK, response)
		} else {
			ctx.JSON(consts.StatusBadRequest, response)
		}
	}
}
