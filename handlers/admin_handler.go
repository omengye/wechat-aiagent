package handlers

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/sirupsen/logrus"
)

// AddAdminRequest 添加管理员请求
type AddAdminRequest struct {
	UID  string `json:"uid" binding:"required"`
	Role string `json:"role" binding:"required"`
}

// AddAdminHandler 添加管理员（只有超级管理员可以调用）
func AddAdminHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req AddAdminRequest
		if err := ctx.BindJSON(&req); err != nil {
			logrus.Warnf("bind json failed: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "请求参数错误",
			})
			return
		}

		// 验证角色类型
		if req.Role != constants.RoleAdmin && req.Role != constants.RoleSuperAdmin {
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "无效的角色类型，只能是 admin 或 super_admin",
			})
			return
		}

		// 添加管理员
		if err := wChat.UserService.AddAdmin(req.UID, req.Role); err != nil {
			logrus.Errorf("add admin failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "添加管理员失败",
			})
			return
		}

		logrus.Infof("admin added successfully: uid=%s, role=%s", req.UID, req.Role)
		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "添加管理员成功",
			"data": map[string]string{
				"uid":  req.UID,
				"role": req.Role,
			},
		})
	}
}

// RemoveAdminRequest 移除管理员请求
type RemoveAdminRequest struct {
	UID string `json:"uid" binding:"required"`
}

// RemoveAdminHandler 移除管理员（只有超级管理员可以调用）
func RemoveAdminHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		var req RemoveAdminRequest
		if err := ctx.BindJSON(&req); err != nil {
			logrus.Warnf("bind json failed: %v", err)
			ctx.JSON(consts.StatusBadRequest, map[string]interface{}{
				"code":    400,
				"message": "请求参数错误",
			})
			return
		}

		// 检查是否是超级管理员（不允许移除配置文件中的超级管理员）
		if req.UID == wChat.Config.Wechat.SuperAdmin {
			ctx.JSON(consts.StatusForbidden, map[string]interface{}{
				"code":    403,
				"message": "不能移除配置文件中的超级管理员",
			})
			return
		}

		// 移除管理员
		if err := wChat.UserService.RemoveAdmin(req.UID); err != nil {
			logrus.Errorf("remove admin failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "移除管理员失败",
			})
			return
		}

		logrus.Infof("admin removed successfully: uid=%s", req.UID)
		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "移除管理员成功",
		})
	}
}

// ListAdminsHandler 列出所有管理员（只有超级管理员可以调用）
func ListAdminsHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		admins, err := wChat.UserService.ListAdmins()
		if err != nil {
			logrus.Errorf("list admins failed: %v", err)
			ctx.JSON(consts.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "获取管理员列表失败",
			})
			return
		}

		// 添加配置文件中的超级管理员
		if wChat.Config.Wechat.SuperAdmin != "" {
			admins[wChat.Config.Wechat.SuperAdmin] = constants.RoleSuperAdmin
		}

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "获取管理员列表成功",
			"data":    admins,
		})
	}
}

// GetMyRoleHandler 获取当前用户角色（所有登录用户可调用）
func GetMyRoleHandler(wChat *server.WeChat) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		// 从context获取用户信息
		uidValue, _ := ctx.Get("uid")
		roleValue, _ := ctx.Get("role")

		uid, _ := uidValue.(string)
		role, _ := roleValue.(string)

		ctx.JSON(consts.StatusOK, map[string]interface{}{
			"code":    200,
			"message": "获取角色成功",
			"data": map[string]string{
				"uid":  uid,
				"role": role,
			},
		})
	}
}
