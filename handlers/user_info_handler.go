package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/sirupsen/logrus"
)

// GetUserInfoHandler 获取微信用户基本信息
func GetUserInfoHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从context获取uid（由AuthMiddleware设置）
		uidValue, exists := ctx.Get("uid")
		if !exists {
			ctx.JSON(consts.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "未找到用户信息",
			})
			return
		}

		uid, ok := uidValue.(string)
		if !ok {
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "用户信息格式错误",
			})
			return
		}

		// 调用微信API获取用户信息
		userInfo, err := wChat.GetUserInfo(uid)
		if err != nil {
			logrus.Errorf("get user info failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "获取用户信息失败",
			})
			return
		}

		// 检查用户是否已关注
		if userInfo.Subscribe == 0 {
			ctx.JSON(consts.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "用户未关注公众号",
			})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "获取用户信息成功",
			"data":    userInfo,
		})
	}
}

// GetUserInfoByOpenidRequest 通过openid获取用户信息的请求
type GetUserInfoByOpenidRequest struct {
	Openid string `json:"openid" binding:"required"`
}

// GetUserInfoByOpenidHandler 通过openid获取微信用户基本信息（管理员接口）
func GetUserInfoByOpenidHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req GetUserInfoByOpenidRequest
		if err := ctx.BindJSON(&req); err != nil {
			logrus.Warnf("bind json failed: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "请求参数错误",
			})
			return
		}

		// 调用微信API获取用户信息
		userInfo, err := wChat.GetUserInfo(req.Openid)
		if err != nil {
			logrus.Errorf("get user info failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "获取用户信息失败",
			})
			return
		}

		// 检查用户是否已关注
		if userInfo.Subscribe == 0 {
			ctx.JSON(consts.StatusOK, map[string]interface{}{
				"code":    200,
				"message": "用户未关注公众号",
				"data":    userInfo,
			})
			return
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "获取用户信息成功",
			"data":    userInfo,
		})
	}
}
