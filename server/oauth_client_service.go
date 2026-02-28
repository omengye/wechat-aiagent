package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/tools"
	"github.com/omengye/wechat_aiagent/types"
	"golang.org/x/crypto/bcrypt"
)

// OAuthClientService OAuth 客户端管理服务
type OAuthClientService struct {
	Redis *tools.RedisMem
}

// NewOAuthClientService 创建 OAuth 客户端服务
func NewOAuthClientService(redis *tools.RedisMem) *OAuthClientService {
	return &OAuthClientService{
		Redis: redis,
	}
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// GenerateClientID 生成客户端 ID
// 格式：app_{random_16_chars}
func (s *OAuthClientService) GenerateClientID() (string, error) {
	randomStr, err := generateRandomString(16)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("app_%s", randomStr), nil
}

// GenerateClientSecret 生成客户端密钥
// 格式：cs_{random_32_chars}
func (s *OAuthClientService) GenerateClientSecret() (string, error) {
	randomStr, err := generateRandomString(32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("cs_%s", randomStr), nil
}

// HashClientSecret 使用 bcrypt 加密客户端密钥
func (s *OAuthClientService) HashClientSecret(secret string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyClientSecret 验证客户端密钥
func (s *OAuthClientService) VerifyClientSecret(secret, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret))
}

// RegisterClient 注册新的 OAuth 客户端
func (s *OAuthClientService) RegisterClient(name, description string, redirectURLs, allowedScopes []string, createdBy string) (*types.RegisterOAuthClientResponse, error) {
	// 1. 生成 client_id 和 client_secret
	clientID, err := s.GenerateClientID()
	if err != nil {
		return nil, fmt.Errorf("生成 client_id 失败: %w", err)
	}

	clientSecret, err := s.GenerateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("生成 client_secret 失败: %w", err)
	}

	// 2. 加密 client_secret
	hashedSecret, err := s.HashClientSecret(clientSecret)
	if err != nil {
		return nil, fmt.Errorf("加密 client_secret 失败: %w", err)
	}

	// 3. 创建客户端对象
	client := &types.OAuthClient{
		ClientID:         clientID,
		ClientSecretHash: hashedSecret,
		Name:             name,
		Description:      description,
		RedirectURLs:     redirectURLs,
		AllowedScopes:    allowedScopes,
		CreatedAt:        time.Now(),
		CreatedBy:        createdBy,
		Status:           constants.OAuthClientStatusActive,
		RateLimit:        constants.RateLimitTokenPerHour,
	}

	// 4. 存储到 Redis
	clientJSON, err := json.Marshal(client)
	if err != nil {
		return nil, fmt.Errorf("序列化客户端信息失败: %w", err)
	}

	key := constants.RedisOAuthClientPrefix + clientID
	// 设置为永不过期（10 年）
	if err := s.Redis.Set(key, string(clientJSON), constants.OAuthClientExpireSeconds); err != nil {
		return nil, fmt.Errorf("存储客户端信息失败: %w", err)
	}

	// 5. 返回响应（包含明文 client_secret，只显示一次）
	return &types.RegisterOAuthClientResponse{
		ClientID:      clientID,
		ClientSecret:  clientSecret, // 明文，只在注册时返回
		Name:          name,
		Description:   description,
		RedirectURLs:  redirectURLs,
		AllowedScopes: allowedScopes,
		CreatedAt:     client.CreatedAt,
	}, nil
}

// GetClient 获取客户端信息
func (s *OAuthClientService) GetClient(clientID string) (*types.OAuthClient, error) {
	key := constants.RedisOAuthClientPrefix + clientID
	clientJSON, err := s.Redis.Get(key)
	if err != nil {
		return nil, err
	}
	if clientJSON == "" {
		return nil, errors.New("客户端不存在")
	}

	var client types.OAuthClient
	if err := json.Unmarshal([]byte(clientJSON), &client); err != nil {
		return nil, fmt.Errorf("解析客户端信息失败: %w", err)
	}

	return &client, nil
}

// saveClient 保存客户端信息到 Redis（内部方法）
func (s *OAuthClientService) saveClient(client *types.OAuthClient) error {
	clientJSON, err := json.Marshal(client)
	if err != nil {
		return fmt.Errorf("序列化客户端信息失败: %w", err)
	}

	key := constants.RedisOAuthClientPrefix + client.ClientID
	// 设置为永不过期（10 年）
	if err := s.Redis.Set(key, string(clientJSON), constants.OAuthClientExpireSeconds); err != nil {
		return fmt.Errorf("保存客户端信息失败: %w", err)
	}

	return nil
}

// UpdateClient 更新客户端信息
func (s *OAuthClientService) UpdateClient(clientID string, updates *types.UpdateOAuthClientRequest) error {
	// 1. 获取现有客户端
	client, err := s.GetClient(clientID)
	if err != nil {
		return err
	}

	// 2. 更新字段
	if updates.Name != "" {
		client.Name = updates.Name
	}
	if updates.Description != "" {
		client.Description = updates.Description
	}
	if len(updates.RedirectURLs) > 0 {
		client.RedirectURLs = updates.RedirectURLs
	}
	if len(updates.AllowedScopes) > 0 {
		client.AllowedScopes = updates.AllowedScopes
	}
	if updates.Status != "" {
		if updates.Status != constants.OAuthClientStatusActive && updates.Status != constants.OAuthClientStatusSuspended {
			return errors.New("无效的状态值")
		}
		client.Status = updates.Status
	}

	// 3. 存储到 Redis
	clientJSON, err := json.Marshal(client)
	if err != nil {
		return fmt.Errorf("序列化客户端信息失败: %w", err)
	}

	key := constants.RedisOAuthClientPrefix + clientID
	if err := s.Redis.Set(key, string(clientJSON), constants.OAuthClientExpireSeconds); err != nil {
		return fmt.Errorf("更新客户端信息失败: %w", err)
	}

	return nil
}

// DeleteClient 删除客户端（软删除，将状态设为 suspended）
func (s *OAuthClientService) DeleteClient(clientID string) error {
	// 标记客户端为已撤销
	revokedKey := constants.RedisOAuthRevokedClientPrefix + clientID
	if err := s.Redis.Set(revokedKey, "revoked", constants.OAuthRevokedExpireSeconds); err != nil {
		return fmt.Errorf("标记客户端为已撤销失败: %w", err)
	}

	// 删除客户端信息
	key := constants.RedisOAuthClientPrefix + clientID
	return s.Redis.Del(key)
}

// IsClientRevoked 检查客户端是否已被撤销
func (s *OAuthClientService) IsClientRevoked(clientID string) (bool, error) {
	revokedKey := constants.RedisOAuthRevokedClientPrefix + clientID
	value, err := s.Redis.Get(revokedKey)
	if err != nil {
		return false, err
	}
	return value != "", nil
}

// ValidateRedirectURL 验证重定向 URI 是否在白名单中
func (s *OAuthClientService) ValidateRedirectURL(clientID, redirectURL string) error {
	client, err := s.GetClient(clientID)
	if err != nil {
		return err
	}

	// 完全匹配（不允许通配符、子域名、路径前缀匹配）
	for _, allowed := range client.RedirectURLs {
		if redirectURL == allowed {
			return nil
		}
	}

	return errors.New("redirect_url 不在白名单中")
}

// ValidateScopes 验证请求的 scopes 是否都在客户端的 allowed_scopes 中
func (s *OAuthClientService) ValidateScopes(clientID string, requestedScopes []string) error {
	client, err := s.GetClient(clientID)
	if err != nil {
		return err
	}

	allowedScopeMap := make(map[string]bool)
	for _, scope := range client.AllowedScopes {
		allowedScopeMap[scope] = true
	}

	for _, scope := range requestedScopes {
		if !allowedScopeMap[scope] {
			return fmt.Errorf("scope '%s' 不在客户端的允许列表中", scope)
		}
	}

	return nil
}

// VerifyClientCredentials 验证客户端凭证
func (s *OAuthClientService) VerifyClientCredentials(clientID, clientSecret string) error {
	// 1. 检查客户端是否被撤销
	if revoked, _ := s.IsClientRevoked(clientID); revoked {
		return errors.New("客户端已被撤销")
	}

	// 2. 获取客户端信息
	client, err := s.GetClient(clientID)
	if err != nil {
		return errors.New("客户端不存在")
	}

	// 3. 检查客户端状态
	if client.Status != constants.OAuthClientStatusActive {
		return errors.New("客户端已被暂停")
	}

	// 4. 验证密钥
	if err := s.VerifyClientSecret(clientSecret, client.ClientSecretHash); err != nil {
		return errors.New("客户端密钥错误")
	}

	return nil
}

// ResetClientSecret 重置客户端密钥
func (s *OAuthClientService) ResetClientSecret(clientID string) (string, error) {
	// 1. 获取客户端
	client, err := s.GetClient(clientID)
	if err != nil {
		return "", err
	}

	// 2. 生成新密钥
	newSecret, err := s.GenerateClientSecret()
	if err != nil {
		return "", fmt.Errorf("生成新密钥失败: %w", err)
	}

	// 3. 加密新密钥
	hashedSecret, err := s.HashClientSecret(newSecret)
	if err != nil {
		return "", fmt.Errorf("加密新密钥失败: %w", err)
	}

	// 4. 更新客户端
	client.ClientSecretHash = hashedSecret
	clientJSON, err := json.Marshal(client)
	if err != nil {
		return "", fmt.Errorf("序列化客户端信息失败: %w", err)
	}

	key := constants.RedisOAuthClientPrefix + clientID
	if err := s.Redis.Set(key, string(clientJSON), constants.OAuthClientExpireSeconds); err != nil {
		return "", fmt.Errorf("更新客户端信息失败: %w", err)
	}

	// 5. 返回明文密钥（只显示一次）
	return newSecret, nil
}

// ParseScopes 将空格分隔的 scope 字符串解析为数组
func ParseScopes(scopeStr string) []string {
	if scopeStr == "" {
		return []string{}
	}
	scopes := strings.Split(scopeStr, " ")
	// 去除空字符串
	result := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// JoinScopes 将 scope 数组连接为空格分隔的字符串
func JoinScopes(scopes []string) string {
	return strings.Join(scopes, " ")
}
