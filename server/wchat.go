package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/hertz-contrib/websocket"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/tools"
	"github.com/omengye/wechat_aiagent/types"
	wechat "github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/officialaccount"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"github.com/silenceper/wechat/v2/officialaccount/server"
	"github.com/sirupsen/logrus"
)

type WeChat struct {
	Account   *officialaccount.OfficialAccount
	AuthToken string
	Config    *types.Config

	// user service
	UserService *UserService
}

type WSSendMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

// maskSensitiveData 对敏感数据进行脱敏处理，保留前4位和后4位
func maskSensitiveData(data string) string {
	if len(data) <= 8 {
		return "****"
	}
	return data[:4] + "****" + data[len(data)-4:]
}

func NewWechatServer(ctx context.Context, config *types.Config) *WeChat {
	wc := wechat.NewWechat()
	// use redis mem
	redisMem := tools.NewRedisMem(ctx, config)
	cfg := &offConfig.Config{
		AppID:     config.Wechat.AppId,
		AppSecret: config.Wechat.AppSecret,
		Token:     config.Wechat.Token,
		//EncodingAESKey: "xxxx",
		Cache:       redisMem.Client,
		UseStableAK: true,
	}

	userService := NewUserService(redisMem)

	return &WeChat{
		Account: wc.GetOfficialAccount(cfg),
		Config:  config,

		UserService: userService,
	}
}

func (wChat *WeChat) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	serv := wChat.Account.GetServer(req, rw)
	wChat.setHandler(serv)
	err := serv.Serve()
	if err != nil {
		logrus.Error(err)
		return
	}
	//发送回复的消息
	err = serv.Send()
	if err != nil {
		logrus.Error(err)
		return
	}
}

// HandleOAuthAuthorization 处理 OAuth 授权（生成授权码并返回重定向 URL）
func (wChat *WeChat) HandleOAuthAuthorization(uid, sessionID string) (string, error) {
	// 1. 获取授权会话信息
	codeService := NewOAuthCodeService(wChat.UserService.Redis)
	session, err := codeService.GetSession(sessionID)
	if err != nil {
		return "", err
	}

	// 2. 验证用户角色是否满足 scope 要求
	role, err := wChat.UserService.GetUserRole(uid, wChat.Config.Wechat.SuperAdmin)
	if err != nil {
		logrus.Warnf("get user role failed: %v, using default role", err)
		role = constants.RoleUser
	}

	scopes := ParseScopes(session.Scope)
	scopeService := NewOAuthScopeService()
	if err := scopeService.ValidateScopesForRole(scopes, role); err != nil {
		logrus.Errorf("user role does not satisfy scopes: %v", err)
		return "", err
	}

	// 3. 生成授权码
	code, err := codeService.CreateAuthorizationCode(
		uid,
		session.ClientID,
		session.RedirectURL,
		session.Scope,
		session.CodeChallenge,
		session.CodeChallengeMethod,
	)
	if err != nil {
		return "", err
	}

	// 4. 删除会话（一次性使用）
	if err := codeService.DeleteSession(sessionID); err != nil {
		logrus.Errorf("delete session failed: %v", err)
	}

	// 5. 记录审计日志
	auditService := NewOAuthAuditService(wChat.UserService.Redis)
	if err := auditService.LogAuthorizationGranted(session.ClientID, uid, "", session.Scope); err != nil {
		logrus.Errorf("log audit failed: %v", err)
	}

	// 6. 构建重定向 URL
	redirectURL := session.RedirectURL + "?code=" + code
	if session.State != "" {
		redirectURL += "&state=" + session.State
	}

	return redirectURL, nil
}

func (wChat *WeChat) setHandler(server *server.Server) {
	//设置接收消息的处理方法
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		uid := string(msg.FromUserName)
		// 扫码事件(已关注用户为SCAN,未关注用户在关注后为subscribe)
		if msg.Event == "SCAN" || msg.Event == "subscribe" {
			logrus.Infof("get event scan, key: %s", maskSensitiveData(msg.EventKey))
			// 微信服务端推送subscribe事件会加上qrscene_
			projectId := strings.ReplaceAll(msg.EventKey, constants.WeChatQrScenePrefix, "")
			if projectId == "" {
				return nil
			}

			// 添加ProjectID验证
			maxProjectIDLen := wChat.Config.Server.MaxProjectIDLen
			if len(projectId) > maxProjectIDLen {
				logrus.Warnf("project ID too long: %d chars, max: %d", len(projectId), maxProjectIDLen)
				return nil
			}

			conn := wChat.UserService.GetWSConn(projectId)
			if conn == nil {
				logrus.Warnf("websocket connection not found for project: %s", projectId)
				return nil
			}

			// 获取项目配置，检查是否配置了 OAuth 客户端
			project := GetProject(projectId)
			if project != nil && len(project.OAuthClients) > 0 {
				// 配置了 OAuth 客户端 - 使用 OAuth 授权流程
				oauthSessionID := wChat.UserService.GetOAuthSession(projectId)
				if oauthSessionID == "" {
					logrus.Warnf("OAuth client configured but no session found for project: %s", projectId)
					return nil
				}

				// OAuth 授权流程
				logrus.Infof("OAuth authorization flow detected for session: %s", oauthSessionID)

				// 处理 OAuth 授权
				redirectURL, err := wChat.HandleOAuthAuthorization(uid, oauthSessionID)
				if err != nil {
					logrus.Errorf("handle oauth authorization failed: %v", err)
					return nil
				}

				// 发送重定向指令给前端
				sendMsgJson, err := json.Marshal(WSSendMessage{
					Type: constants.WSOAuthRedirectType,
					Data: redirectURL,
				})
				if err != nil {
					logrus.Errorf("marshal redirect message failed: %v", err)
					return nil
				}

				if err := conn.WriteMessage(websocket.TextMessage, sendMsgJson); err != nil {
					logrus.Errorf("write websocket message failed: %v", err)
				}

				// 清理连接和会话映射
				if err := conn.Close(); err != nil {
					logrus.Errorf("close websocket connection failed: %v", err)
				}
				wChat.UserService.RemoveWSConn(projectId)
				wChat.UserService.RemoveOAuthSession(projectId)

				text := message.NewText("授权成功！正在跳转...")
				return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
			}

			// 普通登录流程
			userToken, err := wChat.UserService.GenerateTokenFromQrCode(uid, wChat.Config)
			if err != nil {
				logrus.Errorf("generate token failed: %v", err)
				return nil
			}
			logrus.Infof("generate token success: %s", maskSensitiveData(userToken.AccessToken))
			token, err := json.Marshal(userToken)
			if err != nil {
				logrus.Errorf("marshal user token failed: %v", err)
				return nil
			}
			sendMsgJson, err := json.Marshal(WSSendMessage{
				Type: constants.WSTokenType,
				Data: string(token),
			})
			if err != nil {
				logrus.Errorf("marshal send message failed: %v", err)
				return nil
			}

			if err := conn.WriteMessage(websocket.TextMessage, sendMsgJson); err != nil {
				logrus.Errorf("write websocket message failed: %v", err)
			}
			if err := conn.Close(); err != nil {
				logrus.Errorf("close websocket connection failed: %v", err)
			}
			// 清理连接映射
			wChat.UserService.RemoveWSConn(projectId)

			text := message.NewText("login success!🙂")
			return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
		} else if msg.Event == "unsubscribe" {
			// 用户取消关注
			logrus.Infof("uid(%s) unsubscribe", uid)

			// 1. 删除令牌缓存
			if err := wChat.UserService.Redis.Del(constants.RedisAccessTokenPrefix + uid); err != nil {
				logrus.Errorf("redis delete token cache failed: %v", err)
				return nil
			}

			// 2. 将令牌加入黑名单，有效期设为令牌的过期时间
			blacklistKey := constants.RedisTokenBlacklistPrefix + uid
			expiry := wChat.Config.Jwt.Expire
			if err := wChat.UserService.Redis.Set(blacklistKey, "revoked", expiry); err != nil {
				logrus.Errorf("failed to blacklist token for uid %s: %v", uid, err)
			} else {
				logrus.Infof("token blacklisted for uid: %s, expiry: %d seconds", uid, expiry)
			}
		}
		logrus.Infof("get message '%s' from uid: %s", msg.Content, uid)
		return nil
	})
}
