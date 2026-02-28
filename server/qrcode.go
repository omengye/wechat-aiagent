package server

import (
	"time"

	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/types"
	"github.com/silenceper/wechat/v2/officialaccount/basic"
	"github.com/sirupsen/logrus"
)

// QRCodeRequest 创建二维码的请求参数
type QRCodeRequest struct {
	SceneStr   string `json:"scene_str"`   // 场景值字符串,临时二维码时为32字符字符串,永久二维码时最大64字符
	SceneID    int    `json:"scene_id"`    // 场景值ID,临时二维码时为32位非0整型,永久二维码时最大值为100000
	ExpireTime int    `json:"expire_time"` // 过期时间,单位秒,最大不超过2592000(30天),0表示永久二维码
	ActionName string `json:"action_name"` // 二维码类型:QR_SCENE为临时整型,QR_STR_SCENE为临时字符串,QR_LIMIT_SCENE为永久整型,QR_LIMIT_STR_SCENE为永久字符串
}

// QRCodeResponse 创建二维码的响应
type QRCodeResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Ticket        string `json:"ticket"`         // 获取的二维码ticket
		ExpireSeconds int    `json:"expire_seconds"` // 二维码有效时间,单位秒
		QRCodeURL     string `json:"qrcode_url"`     // 可以直接展示的二维码图片URL
	} `json:"data,omitempty"`
}

func (wChat *WeChat) CreateTmpQRCode(project *types.Project, expire int) (string, error) {
	// first query redis cache
	if url, err := wChat.UserService.Redis.Get(constants.RedisProjectQrcodePrefix + project.ProjectId); err != nil {
		return "", err
	} else if url != "" {
		return url, nil
	}

	// 如果 expire 为 0，使用配置的默认过期时间
	if expire == 0 {
		expire = wChat.Config.Wechat.QrcodeDefaultExpire
	}

	qrCode := wChat.CreateQRCode(QRCodeRequest{
		SceneStr:   project.ProjectId,
		ExpireTime: expire,
	})
	if qrCode.Code != 200 {
		return "", errors.NewPrivate("failed to generate QRCode")
	}

	// save redis cache
	if err := wChat.UserService.Redis.Set(constants.RedisProjectQrcodePrefix+project.ProjectId, qrCode.Data.QRCodeURL, int64(expire)); err != nil {
		return "", err
	}
	return qrCode.Data.QRCodeURL, nil
}

// CreateQRCode 创建带参数的二维码
func (wChat *WeChat) CreateQRCode(req QRCodeRequest) QRCodeResponse {
	// 验证参数
	if req.ActionName == "" {
		req.ActionName = "QR_STR_SCENE" // 默认使用临时字符串二维码
	}

	if req.ExpireTime == 0 && (req.ActionName == "QR_SCENE" || req.ActionName == "QR_STR_SCENE") {
		req.ExpireTime = wChat.Config.Wechat.QrcodeDefaultExpire // 使用配置的默认过期时间
	}

	if req.ActionName == "QR_SCENE" || req.ActionName == "QR_LIMIT_SCENE" {
		if req.SceneID == 0 {
			return QRCodeResponse{
				Code:    400,
				Message: "scene_id 不能为空",
			}
		}
	} else {
		if req.SceneStr == "" {
			return QRCodeResponse{
				Code:    400,
				Message: "scene_str 不能为空",
			}
		}
	}

	// 获取二维码管理器
	qrcode := wChat.Account.GetBasic()

	var ticket *basic.Ticket
	var err error

	// 根据类型创建二维码
	switch req.ActionName {
	case "QR_SCENE":
		// 临时整型参数二维码
		reqs := basic.NewTmpQrRequest(time.Second*time.Duration(req.ExpireTime), req.SceneID)
		ticket, err = qrcode.GetQRTicket(reqs)
	case "QR_STR_SCENE":
		// 临时字符串参数二维码
		reqs := basic.NewTmpQrRequest(time.Second*time.Duration(req.ExpireTime), req.SceneStr)
		ticket, err = qrcode.GetQRTicket(reqs)
	case "QR_LIMIT_SCENE":
		// 永久整型参数二维码
		reqs := basic.NewLimitQrRequest(req.SceneID)
		ticket, err = qrcode.GetQRTicket(reqs)
	case "QR_LIMIT_STR_SCENE":
		// 永久字符串参数二维码
		reqs := basic.NewLimitQrRequest(req.SceneStr)
		ticket, err = qrcode.GetQRTicket(reqs)
	default:
		return QRCodeResponse{
			Code:    400,
			Message: "不支持的二维码类型: " + req.ActionName,
		}
	}

	if err != nil {
		// 详细错误记录到日志（内部使用）
		logrus.Errorf("创建二维码失败 - ActionName: %s, SceneStr: %s, Error: %v",
			req.ActionName, req.SceneStr, err)
		// 返回通用错误给客户端（避免泄露内部信息）
		return QRCodeResponse{
			Code:    500,
			Message: "创建二维码失败，请稍后重试",
		}
	}

	// 获取二维码图片URL
	qrcodeURL := basic.ShowQRCode(ticket)

	// 记录成功日志，对Ticket进行脱敏
	logrus.Infof("成功创建二维码 - ActionName: %s, Ticket: %s",
		req.ActionName, maskSensitiveData(ticket.Ticket))

	return QRCodeResponse{
		Code:    200,
		Message: "success",
		Data: &struct {
			Ticket        string `json:"ticket"`
			ExpireSeconds int    `json:"expire_seconds"`
			QRCodeURL     string `json:"qrcode_url"`
		}{
			Ticket:        ticket.Ticket,
			ExpireSeconds: req.ExpireTime,
			QRCodeURL:     qrcodeURL,
		},
	}
}
