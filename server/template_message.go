package server

import (
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"github.com/sirupsen/logrus"
)

// TemplateMessageData 模板消息数据项
type TemplateMessageData struct {
	Value string `json:"value"` // 数据值
	Color string `json:"color"` // 颜色（可选，默认黑色）
}

// TemplateMessageRequest 发送模板消息请求
type TemplateMessageRequest struct {
	ToUser string                         `json:"touser"` // 接收者openid（必填）
	URL    string                         `json:"url"`    // 跳转链接（可选）
	Data   map[string]TemplateMessageData `json:"data"`   // 模板数据（必填，必须包含content字段）
}

// TemplateMessageResponse 发送模板消息响应
type TemplateMessageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	MsgID   int64  `json:"msgid,omitempty"` // 消息ID
}

// SendTemplateMessage 发送模板消息
func (wChat *WeChat) SendTemplateMessage(req *TemplateMessageRequest) (*TemplateMessageResponse, error) {
	// 参数验证
	if req.ToUser == "" {
		logrus.Warn("toUser is empty")
		return &TemplateMessageResponse{
			Code:    400,
			Message: "接收者openid不能为空",
		}, nil
	}

	if len(req.Data) == 0 {
		logrus.Warn("template data is empty")
		return &TemplateMessageResponse{
			Code:    400,
			Message: "模板数据不能为空",
		}, nil
	}

	// 验证data中必须包含content字段
	if _, exists := req.Data["content"]; !exists {
		logrus.Warn("data must contain 'content' field")
		return &TemplateMessageResponse{
			Code:    400,
			Message: "模板数据必须包含content字段",
		}, nil
	}

	// 构建模板消息
	templateMsg := &message.TemplateMessage{
		ToUser:     req.ToUser,
		TemplateID: wChat.Config.Wechat.TemplateId,
		URL:        req.URL,
		Data:       make(map[string]*message.TemplateDataItem),
	}

	// 转换模板数据格式
	for key, value := range req.Data {
		templateMsg.Data[key] = &message.TemplateDataItem{
			Value: value.Value,
			Color: value.Color,
		}
	}

	// 发送模板消息
	tmpl := wChat.Account.GetTemplate()
	msgID, err := tmpl.Send(templateMsg)
	if err != nil {
		logrus.Errorf("send template message failed: %v", err)
		return &TemplateMessageResponse{
			Code:    500,
			Message: "发送模板消息失败: " + err.Error(),
		}, err
	}

	logrus.Infof("template message sent successfully, msgID: %d, toUser: %s", msgID, maskSensitiveData(req.ToUser))
	return &TemplateMessageResponse{
		Code:    200,
		Message: "success",
		MsgID:   msgID,
	}, nil
}
