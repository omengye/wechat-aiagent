package server

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/tools"
	"github.com/omengye/wechat_aiagent/types"
)

// OAuthCodeService OAuth 授权码服务
type OAuthCodeService struct {
	Redis *tools.RedisMem
}

// NewOAuthCodeService 创建 OAuth 授权码服务
func NewOAuthCodeService(redis *tools.RedisMem) *OAuthCodeService {
	return &OAuthCodeService{
		Redis: redis,
	}
}

// GenerateAuthorizationCode 生成授权码
// 格式：ac_{uuid_without_dashes}
func (s *OAuthCodeService) GenerateAuthorizationCode() string {
	id := uuid.New()
	return fmt.Sprintf("ac_%s", strings.ReplaceAll(id.String(), "-", ""))
}

// GenerateSessionID 生成会话 ID
func (s *OAuthCodeService) GenerateSessionID() string {
	id := uuid.New()
	return fmt.Sprintf("session_%s", strings.ReplaceAll(id.String(), "-", ""))
}

// CreateSession 创建授权会话
func (s *OAuthCodeService) CreateSession(clientID, redirectURL, scope, state, codeChallenge, codeChallengeMethod string) (string, error) {
	sessionID := s.GenerateSessionID()

	session := &types.OAuthSession{
		SessionID:           sessionID,
		ClientID:            clientID,
		RedirectURL:         redirectURL,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		CreatedAt:           time.Now(),
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("序列化会话失败: %w", err)
	}

	key := constants.RedisOAuthSessionPrefix + sessionID
	if err := s.Redis.Set(key, string(sessionJSON), constants.OAuthSessionExpireSeconds); err != nil {
		return "", fmt.Errorf("存储会话失败: %w", err)
	}

	return sessionID, nil
}

// GetSession 获取授权会话
func (s *OAuthCodeService) GetSession(sessionID string) (*types.OAuthSession, error) {
	key := constants.RedisOAuthSessionPrefix + sessionID
	sessionJSON, err := s.Redis.Get(key)
	if err != nil {
		return nil, err
	}
	if sessionJSON == "" {
		return nil, errors.New("会话不存在或已过期")
	}

	var session types.OAuthSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("解析会话失败: %w", err)
	}

	return &session, nil
}

// DeleteSession 删除会话
func (s *OAuthCodeService) DeleteSession(sessionID string) error {
	key := constants.RedisOAuthSessionPrefix + sessionID
	return s.Redis.Del(key)
}

// CreateAuthorizationCode 创建授权码
func (s *OAuthCodeService) CreateAuthorizationCode(uid, clientID, redirectURL, scope, codeChallenge, codeChallengeMethod string) (string, error) {
	code := s.GenerateAuthorizationCode()

	authCode := &types.OAuthAuthorizationCode{
		Code:                code,
		UID:                 uid,
		ClientID:            clientID,
		RedirectURL:         redirectURL,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		CreatedAt:           time.Now(),
		Used:                false,
	}

	codeJSON, err := json.Marshal(authCode)
	if err != nil {
		return "", fmt.Errorf("序列化授权码失败: %w", err)
	}

	key := constants.RedisOAuthCodePrefix + code
	if err := s.Redis.Set(key, string(codeJSON), constants.OAuthCodeExpireSeconds); err != nil {
		return "", fmt.Errorf("存储授权码失败: %w", err)
	}

	return code, nil
}

// GetAuthorizationCode 获取授权码
func (s *OAuthCodeService) GetAuthorizationCode(code string) (*types.OAuthAuthorizationCode, error) {
	key := constants.RedisOAuthCodePrefix + code
	codeJSON, err := s.Redis.Get(key)
	if err != nil {
		return nil, err
	}
	if codeJSON == "" {
		return nil, errors.New("授权码不存在或已过期")
	}

	var authCode types.OAuthAuthorizationCode
	if err := json.Unmarshal([]byte(codeJSON), &authCode); err != nil {
		return nil, fmt.Errorf("解析授权码失败: %w", err)
	}

	return &authCode, nil
}

// MarkAuthorizationCodeAsUsed 标记授权码为已使用
func (s *OAuthCodeService) MarkAuthorizationCodeAsUsed(code string) error {
	authCode, err := s.GetAuthorizationCode(code)
	if err != nil {
		return err
	}

	authCode.Used = true
	codeJSON, err := json.Marshal(authCode)
	if err != nil {
		return fmt.Errorf("序列化授权码失败: %w", err)
	}

	key := constants.RedisOAuthCodePrefix + code
	// 保留剩余的 TTL
	if err := s.Redis.Set(key, string(codeJSON), constants.OAuthCodeExpireSeconds); err != nil {
		return fmt.Errorf("更新授权码失败: %w", err)
	}

	return nil
}

// DeleteAuthorizationCode 删除授权码
func (s *OAuthCodeService) DeleteAuthorizationCode(code string) error {
	key := constants.RedisOAuthCodePrefix + code
	return s.Redis.Del(key)
}

// VerifyPKCE 验证 PKCE code_verifier
func (s *OAuthCodeService) VerifyPKCE(codeVerifier, codeChallenge, method string) error {
	if method == "" {
		// 如果没有使用 PKCE，跳过验证
		return nil
	}

	if codeVerifier == "" {
		return errors.New("缺少 code_verifier")
	}

	switch method {
	case constants.PKCEMethodPlain:
		if codeVerifier != codeChallenge {
			return errors.New("PKCE 验证失败")
		}
	case constants.PKCEMethodS256:
		hash := sha256.Sum256([]byte(codeVerifier))
		computedChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
		if computedChallenge != codeChallenge {
			return errors.New("PKCE 验证失败")
		}
	default:
		return fmt.Errorf("不支持的 code_challenge_method: %s", method)
	}

	return nil
}

// RecordUserGrant 记录用户授权
func (s *OAuthCodeService) RecordUserGrant(uid, clientID, clientName, scope string) error {
	key := constants.RedisOAuthUserGrantPrefix + uid

	// 获取现有授权记录
	grantsJSON, err := s.Redis.Get(key)
	var grants map[string]types.OAuthUserGrant
	if err == nil && grantsJSON != "" {
		if err := json.Unmarshal([]byte(grantsJSON), &grants); err != nil {
			grants = make(map[string]types.OAuthUserGrant)
		}
	} else {
		grants = make(map[string]types.OAuthUserGrant)
	}

	// 更新或创建授权记录
	if grant, exists := grants[clientID]; exists {
		grant.LastUsedAt = time.Now()
		grant.TokenCount++
		grants[clientID] = grant
	} else {
		grants[clientID] = types.OAuthUserGrant{
			ClientName: clientName,
			Scope:      ParseScopes(scope),
			GrantedAt:  time.Now(),
			LastUsedAt: time.Now(),
			TokenCount: 1,
		}
	}

	// 存储到 Redis
	updatedJSON, err := json.Marshal(grants)
	if err != nil {
		return fmt.Errorf("序列化授权记录失败: %w", err)
	}

	// 永不过期（10 年）
	if err := s.Redis.Set(key, string(updatedJSON), constants.OAuthUserGrantExpireSeconds); err != nil {
		return fmt.Errorf("存储授权记录失败: %w", err)
	}

	return nil
}

// GetUserGrants 获取用户的所有授权记录
func (s *OAuthCodeService) GetUserGrants(uid string) (map[string]types.OAuthUserGrant, error) {
	key := constants.RedisOAuthUserGrantPrefix + uid
	grantsJSON, err := s.Redis.Get(key)
	if err != nil {
		return nil, err
	}
	if grantsJSON == "" {
		return make(map[string]types.OAuthUserGrant), nil
	}

	var grants map[string]types.OAuthUserGrant
	if err := json.Unmarshal([]byte(grantsJSON), &grants); err != nil {
		return nil, fmt.Errorf("解析授权记录失败: %w", err)
	}

	return grants, nil
}

// RevokeUserGrant 撤销用户对某个客户端的授权
func (s *OAuthCodeService) RevokeUserGrant(uid, clientID string) error {
	// 1. 删除授权记录
	key := constants.RedisOAuthUserGrantPrefix + uid
	grantsJSON, err := s.Redis.Get(key)
	if err != nil {
		return err
	}
	if grantsJSON == "" {
		return nil // 没有授权记录，直接返回
	}

	var grants map[string]types.OAuthUserGrant
	if err := json.Unmarshal([]byte(grantsJSON), &grants); err != nil {
		return fmt.Errorf("解析授权记录失败: %w", err)
	}

	delete(grants, clientID)

	updatedJSON, err := json.Marshal(grants)
	if err != nil {
		return fmt.Errorf("序列化授权记录失败: %w", err)
	}

	if err := s.Redis.Set(key, string(updatedJSON), constants.OAuthUserGrantExpireSeconds); err != nil {
		return fmt.Errorf("更新授权记录失败: %w", err)
	}

	// 2. 标记为已撤销（用于 token 验证时检查）
	revokedKey := constants.RedisOAuthRevokedGrantPrefix + uid + ":" + clientID
	if err := s.Redis.Set(revokedKey, "revoked", constants.OAuthRevokedExpireSeconds); err != nil {
		return fmt.Errorf("标记授权为已撤销失败: %w", err)
	}

	return nil
}

// IsGrantRevoked 检查用户对某个客户端的授权是否已撤销
func (s *OAuthCodeService) IsGrantRevoked(uid, clientID string) (bool, error) {
	revokedKey := constants.RedisOAuthRevokedGrantPrefix + uid + ":" + clientID
	value, err := s.Redis.Get(revokedKey)
	if err != nil {
		return false, err
	}
	return value != "", nil
}

// GenerateRandomState 生成随机 state 字符串（用于测试）
func GenerateRandomState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
