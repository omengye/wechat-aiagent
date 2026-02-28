package handlers

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/json"
	"github.com/hertz-contrib/websocket"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/server"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var (
	// projectIDRegex ProjectID验证正则表达式（只允许字母、数字、下划线、连字符）
	projectIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type           string `json:"type"`
	Msg            string `json:"msg"`
	OAuthSessionID string `json:"oauth_session_id,omitempty"` // OAuth 授权会话 ID
}

// IPRateLimiter IP速率限制器
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewIPRateLimiter 创建新的IP速率限制器
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter 获取指定IP的限流器
func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.mu.Lock()
		rl.limiters[ip] = limiter
		rl.mu.Unlock()
	}

	return limiter
}

// isValidProjectID 验证ProjectID格式
func isValidProjectID(id string, maxLen int) bool {
	if len(id) == 0 || len(id) > maxLen {
		return false
	}
	return projectIDRegex.MatchString(id)
}

// sendErrorMessage 向WebSocket连接发送错误消息
func sendErrorMessage(conn *websocket.Conn, message string) {
	response := map[string]string{
		"type":  "error",
		"error": message,
	}
	data, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("marshal error message failed: %v", err)
		return
	}
	sendMessage(conn, data)
}

func sendMessage(conn *websocket.Conn, data []byte) {
	if data == nil {
		logrus.Errorf("send message failed, data is nil")
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		logrus.Errorf("write error message failed: %v", err)
	}
}

// WSHandler WebSocket处理函数
func WSHandler(wChat *server.WeChat) app.HandlerFunc {
	expireTime := wChat.Config.Wechat.QrcodeExpire
	allowedOrigins := wChat.Config.Server.AllowedOrigins
	maxMessageSize := wChat.Config.Server.MaxMessageSize
	maxProjectIDLen := wChat.Config.Server.MaxProjectIDLen

	// 创建速率限制器，使用配置中的参数
	qrcodeRateLimiter := NewIPRateLimiter(
		rate.Limit(wChat.Config.Server.RateLimitRate),
		wChat.Config.Server.RateLimitBurst,
	)

	// 创建upgrader，使用配置中的allowedOrigins
	upgrader := websocket.HertzUpgrader{
		CheckOrigin: func(ctx *app.RequestContext) bool {
			origin := string(ctx.Request.Header.Peek("Origin"))

			// 检查白名单
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}

			// 如果白名单为空（首次部署），记录警告但允许通过
			if len(allowedOrigins) == 0 {
				logrus.Warnf("Warning: No allowed origins configured. Origin: %s", origin)
				return true
			}

			logrus.Warnf("Blocked WebSocket connection from unauthorized origin: %s", origin)
			return false
		},
	}

	return func(c context.Context, ctx *app.RequestContext) {
		// 获取客户端IP用于速率限制
		clientIP := ctx.ClientIP()

		err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
			// 确保连接关闭时清理资源
			defer func() {
				if err := conn.Close(); err != nil {
					logrus.Errorf("close websocket connection error: %v", err)
				}
			}()

			for {
				mt, message, err := conn.ReadMessage()
				if err != nil {
					logrus.Errorf("read: %v", err)
					break
				}
				logrus.Infof("recv message type: %d, size: %d bytes", mt, len(message))
				if mt == websocket.TextMessage {
					filterMessage(message, conn, wChat, clientIP, maxMessageSize, maxProjectIDLen, qrcodeRateLimiter)
				}

				// 修复: 使用秒作为单位,与配置文件保持一致
				if err := conn.SetReadDeadline(time.Now().Add(time.Duration(expireTime) * time.Second)); err != nil {
					logrus.Errorf("set read deadline error: %v", err)
					return
				}
			}
		})
		if err != nil {
			logrus.Errorf("upgrade: %v", err)
			return
		}
	}
}

func filterMessage(message []byte, conn *websocket.Conn, wChat *server.WeChat, clientIP string, maxMessageSize, maxProjectIDLen int, rateLimiter *IPRateLimiter) {
	// 1. 检查消息大小
	if len(message) > maxMessageSize {
		logrus.Warnf("message size exceeds limit: %d bytes from IP: %s", len(message), clientIP)
		sendErrorMessage(conn, "消息大小超出限制")
		return
	}

	// 2. 解析消息
	var msg WSMessage
	err := json.Unmarshal(message, &msg)
	if err != nil {
		logrus.Errorf("unmarshal message error: %v", err)
		sendErrorMessage(conn, "消息格式错误")
		return
	}

	if msg.Type == "ping" {
		sendMessage(conn, []byte("pong"))
		return
	}

	// 3. 验证消息类型（白名单）
	if msg.Type != "qrcode" {
		logrus.Warnf("invalid message type: %s from IP: %s", msg.Type, clientIP)
		sendErrorMessage(conn, "不支持的消息类型")
		return
	}

	// 4. 速率限制检查
	limiter := rateLimiter.GetLimiter(clientIP)
	if !limiter.Allow() {
		logrus.Warnf("rate limit exceeded for IP: %s", clientIP)
		sendErrorMessage(conn, "请求过于频繁，请稍后再试")
		return
	}

	// 5. 验证ProjectID格式
	if !isValidProjectID(msg.Msg, maxProjectIDLen) {
		logrus.Warnf("invalid project ID format: %s from IP: %s", msg.Msg, clientIP)
		sendErrorMessage(conn, "项目ID格式无效")
		return
	}

	// 6. 检查项目是否存在
	project := server.GetProject(msg.Msg)
	if project == nil {
		logrus.Warnf("project not found: %s from IP: %s", msg.Msg, clientIP)
		sendErrorMessage(conn, "项目不存在")
		return
	}

	// 7. 创建二维码
	url, err := wChat.CreateTmpQRCode(project, wChat.Config.Wechat.QrcodeExpire)
	if err != nil {
		logrus.Errorf("create qrcode error: %v", err)
		sendErrorMessage(conn, "创建二维码失败，请稍后重试")
		return
	}

	// 8. 发送二维码URL
	sendMsg := &server.WSSendMessage{
		Type: constants.WSQrcodeType,
		Data: url,
	}
	sendMsgJson, err := json.Marshal(sendMsg)
	if err != nil {
		logrus.Errorf("marshal send message error: %v", err)
		sendErrorMessage(conn, "系统错误")
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, sendMsgJson)
	if err != nil {
		logrus.Errorf("write message error: %v", err)
		return
	}

	// 9. 保存WebSocket连接
	wChat.UserService.SetWSConn(project.ProjectId, conn)

	// 10. 如果是 OAuth 授权流程，保存 session ID 映射
	if msg.OAuthSessionID != "" {
		wChat.UserService.SetOAuthSession(project.ProjectId, msg.OAuthSessionID)
		logrus.Infof("OAuth authorization session created: %s for project: %s", msg.OAuthSessionID, msg.Msg)
	}

	logrus.Infof("QR code created successfully for project: %s, IP: %s", msg.Msg, clientIP)
}
