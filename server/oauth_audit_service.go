package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/tools"
	"github.com/omengye/wechat_aiagent/types"
)

// OAuthAuditService OAuth 审计日志服务
type OAuthAuditService struct {
	Redis *tools.RedisMem
}

// NewOAuthAuditService 创建审计日志服务
func NewOAuthAuditService(redis *tools.RedisMem) *OAuthAuditService {
	return &OAuthAuditService{
		Redis: redis,
	}
}

// LogEvent 记录审计事件
func (s *OAuthAuditService) LogEvent(event, clientID, uid, ip, userAgent, details string) error {
	log := types.OAuthAuditLog{
		Event:     event,
		ClientID:  clientID,
		UID:       uid,
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
		Details:   details,
	}

	logJSON, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("序列化审计日志失败: %w", err)
	}

	// 使用 List 存储审计日志
	key := constants.RedisOAuthAuditPrefix + clientID

	// 将日志添加到列表头部
	// 注意：这里简化实现，实际生产环境建议使用专门的日志系统
	// 由于 RedisMem 可能不支持 LPUSH，这里使用简单的 key-value 存储
	// 每个事件一个 key，使用时间戳作为唯一标识
	logKey := fmt.Sprintf("%s:%d", key, time.Now().UnixNano())
	if err := s.Redis.Set(logKey, string(logJSON), constants.OAuthAuditLogExpireSeconds); err != nil {
		return fmt.Errorf("存储审计日志失败: %w", err)
	}

	return nil
}

// LogClientRegistered 记录客户端注册事件
func (s *OAuthAuditService) LogClientRegistered(clientID, createdBy, ip string) error {
	return s.LogEvent(
		constants.AuditEventClientRegistered,
		clientID,
		createdBy,
		ip,
		"",
		fmt.Sprintf("客户端 %s 被注册", clientID),
	)
}

// LogClientSecretReset 记录密钥重置事件
func (s *OAuthAuditService) LogClientSecretReset(clientID, operatorUID, ip string) error {
	return s.LogEvent(
		constants.AuditEventClientSecretReset,
		clientID,
		operatorUID,
		ip,
		"",
		fmt.Sprintf("客户端 %s 的密钥被重置", clientID),
	)
}

// LogClientRevoked 记录客户端撤销事件
func (s *OAuthAuditService) LogClientRevoked(clientID, operatorUID, ip string) error {
	return s.LogEvent(
		constants.AuditEventClientRevoked,
		clientID,
		operatorUID,
		ip,
		"",
		fmt.Sprintf("客户端 %s 被撤销", clientID),
	)
}

// LogAuthorizationGranted 记录授权事件
func (s *OAuthAuditService) LogAuthorizationGranted(clientID, uid, ip, scope string) error {
	return s.LogEvent(
		constants.AuditEventAuthorizationGranted,
		clientID,
		uid,
		ip,
		"",
		fmt.Sprintf("用户 %s 授权了 scope: %s", uid, scope),
	)
}

// LogTokenIssued 记录 Token 签发事件
func (s *OAuthAuditService) LogTokenIssued(clientID, uid, ip, userAgent string) error {
	return s.LogEvent(
		constants.AuditEventTokenIssued,
		clientID,
		uid,
		ip,
		userAgent,
		"Token 已签发",
	)
}

// LogTokenRefreshed 记录 Token 刷新事件
func (s *OAuthAuditService) LogTokenRefreshed(clientID, uid, ip, userAgent string) error {
	return s.LogEvent(
		constants.AuditEventTokenRefreshed,
		clientID,
		uid,
		ip,
		userAgent,
		"Token 已刷新",
	)
}

// LogGrantRevoked 记录授权撤销事件
func (s *OAuthAuditService) LogGrantRevoked(clientID, uid, ip string) error {
	return s.LogEvent(
		constants.AuditEventGrantRevoked,
		clientID,
		uid,
		ip,
		"",
		fmt.Sprintf("用户 %s 撤销了对客户端 %s 的授权", uid, clientID),
	)
}

// LogInvalidClientSecret 记录无效密钥事件
func (s *OAuthAuditService) LogInvalidClientSecret(clientID, ip string) error {
	return s.LogEvent(
		constants.AuditEventInvalidClientSecret,
		clientID,
		"",
		ip,
		"",
		"客户端密钥验证失败",
	)
}

// LogInvalidAuthCode 记录无效授权码事件
func (s *OAuthAuditService) LogInvalidAuthCode(clientID, ip, code string) error {
	return s.LogEvent(
		constants.AuditEventInvalidAuthCode,
		clientID,
		"",
		ip,
		"",
		fmt.Sprintf("授权码 %s 无效或已过期", maskSensitiveData(code)),
	)
}

// LogRateLimitExceeded 记录限流事件
func (s *OAuthAuditService) LogRateLimitExceeded(clientID, ip string) error {
	return s.LogEvent(
		constants.AuditEventRateLimitExceeded,
		clientID,
		"",
		ip,
		"",
		"触发限流",
	)
}
