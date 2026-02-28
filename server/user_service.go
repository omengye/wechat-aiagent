package server

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/hertz-contrib/websocket"
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/tools"
	"github.com/omengye/wechat_aiagent/types"
	"github.com/sirupsen/logrus"
)

var (
	// Projects 全局项目配置映射
	Projects   map[string]*types.Project
	projectsMu sync.RWMutex // 保护 Projects 的并发访问
)

// InitProject 初始化项目配置
func InitProject(projects []types.Project) {
	projectsMu.Lock()
	defer projectsMu.Unlock()

	if Projects == nil {
		Projects = make(map[string]*types.Project)
	}
	for _, project := range projects {
		p := project // 避免循环变量引用问题
		Projects[p.ProjectId] = &p
	}
}

// GetProject 线程安全地获取项目配置
func GetProject(projectId string) *types.Project {
	projectsMu.RLock()
	defer projectsMu.RUnlock()
	return Projects[projectId]
}

// UserService 用户服务，管理用户token和websocket连接
type UserService struct {
	Redis *tools.RedisMem

	WSProjectConn    map[string]*websocket.Conn
	OAuthSessionMap  map[string]string // projectID -> oauthSessionID
	connMu           sync.RWMutex      // 保护 WSProjectConn 的并发访问
	oauthSessionMu   sync.RWMutex      // 保护 OAuthSessionMap 的并发访问
}

// NewUserService 创建用户服务实例
func NewUserService(redisMem *tools.RedisMem) *UserService {
	conn := make(map[string]*websocket.Conn)
	oauthSession := make(map[string]string)
	return &UserService{
		Redis:           redisMem,
		WSProjectConn:   conn,
		OAuthSessionMap: oauthSession,
	}
}

// SetWSConn 线程安全地设置websocket连接
func (us *UserService) SetWSConn(projectId string, conn *websocket.Conn) {
	us.connMu.Lock()
	defer us.connMu.Unlock()
	us.WSProjectConn[projectId] = conn
}

// GetWSConn 线程安全地获取websocket连接
func (us *UserService) GetWSConn(projectId string) *websocket.Conn {
	us.connMu.RLock()
	defer us.connMu.RUnlock()
	return us.WSProjectConn[projectId]
}

// RemoveWSConn 线程安全地移除websocket连接
func (us *UserService) RemoveWSConn(projectId string) {
	us.connMu.Lock()
	defer us.connMu.Unlock()
	delete(us.WSProjectConn, projectId)
}

// UserToken 用户token信息
type UserToken struct {
	AccessToken     string `json:"access_token"`
	AccessTokenExp  int64  `json:"access_token_exp"`
	RefreshToken    string `json:"refresh_token"`
	RefreshTokenExp int64  `json:"refresh_token_exp"`
}

// GenerateTokenFromQrCode 从二维码扫描生成用户token
func (us *UserService) GenerateTokenFromQrCode(uid string, config *types.Config) (*UserToken, error) {
	// 1. 检查是否在黑名单中
	blacklistKey := constants.RedisTokenBlacklistPrefix + uid
	if isBlacklisted, _ := us.Redis.Get(blacklistKey); isBlacklisted != "" {
		logrus.Warnf("user %s is blacklisted, generating new token", uid)
		// 清除旧缓存，生成新token
		us.Redis.Del(constants.RedisAccessTokenPrefix + uid)
	} else if tk, err := us.Redis.Get(constants.RedisAccessTokenPrefix + uid); err != nil {
		return nil, err
	} else if tk != "" {
		var token UserToken
		if err := json.Unmarshal([]byte(tk), &token); err != nil {
			return nil, err
		}

		// 2. 验证缓存的token是否仍然有效
		_, err := ParseToken(token.AccessToken, config.Jwt.Secret)
		if err != nil {
			// Token已过期或无效，删除缓存并重新生成
			logrus.Infof("cached token expired or invalid for uid: %s, regenerating", uid)
			us.Redis.Del(constants.RedisAccessTokenPrefix + uid)
		} else {
			// Token仍然有效，返回缓存的token
			return &token, nil
		}
	}

	// 获取用户角色
	role, err := us.GetUserRole(uid, config.Wechat.SuperAdmin)
	if err != nil {
		logrus.Warnf("get user role failed for uid %s: %v, using default role", uid, err)
		role = constants.RoleUser
	}

	token, err := GenerateAccessToken(uid, config.Jwt.Secret, config.Jwt.Expire, role)
	if err != nil {
		logrus.Errorf("generate token failed: %v", err)
		return nil, err
	}
	refreshToken, err := GenerateRefreshToken(uid, config.Jwt.Secret, config.Jwt.RefreshExpire, role)
	if err != nil {
		logrus.Errorf("generate refresh token failed: %v", err)
		return nil, err
	}
	tk := &UserToken{
		AccessToken:     token,
		AccessTokenExp:  config.Jwt.Expire,
		RefreshToken:    refreshToken,
		RefreshTokenExp: config.Jwt.RefreshExpire,
	}
	marshal, err := json.Marshal(tk)
	if err != nil {
		logrus.Errorf("marshal token failed: %v", err)
		return nil, err
	}
	if err = us.Redis.Set(constants.RedisAccessTokenPrefix+uid, string(marshal), config.Jwt.Expire); err != nil {
		return nil, err
	}
	return tk, nil
}

// RefreshUserToken 使用 refresh token 生成新的 token 对（Token Rotation）
// 这个方法实现了 Token Rotation 策略：
// 1. 验证旧的 refresh token
// 2. 生成新的 access token 和 refresh token
// 3. 将旧的 refresh token 加入撤销列表
// 4. 更新 Redis 缓存
func (us *UserService) RefreshUserToken(refreshToken string, config *types.Config) (*UserToken, error) {
	// 1. 解析并验证 refresh token
	claims, err := ParseToken(refreshToken, config.Jwt.Secret)
	if err != nil {
		logrus.Warnf("parse refresh token failed: %v", err)
		return nil, err
	}

	// 2. 检查 token 类型
	if claims.TokenType != constants.TokenTypeRefresh {
		logrus.Warnf("invalid token type: %s, expected refresh", claims.TokenType)
		return nil, errors.New("token类型错误")
	}

	uid := claims.Uid

	// 3. 检查用户是否在黑名单中（已取消关注）
	if isBlacklisted, _ := us.IsTokenBlacklisted(uid); isBlacklisted {
		logrus.Warnf("user %s is blacklisted, cannot refresh token", uid)
		return nil, errors.New("用户已被拉黑")
	}

	// 4. 检查 refresh token 是否已被撤销（防止重复使用）
	if isRevoked, _ := us.IsRefreshTokenRevoked(refreshToken); isRevoked {
		logrus.Warnf("refresh token has been revoked, possible token theft detected")
		// 安全措施：如果检测到已撤销的 token 被使用，将该用户的所有 token 加入黑名单
		us.Redis.Set(constants.RedisTokenBlacklistPrefix+uid, "revoked", config.Jwt.RefreshExpire)
		return nil, errors.New("refresh token已被撤销")
	}

	// 5. 获取用户最新角色（可能在token有效期内角色被修改）
	role, err := us.GetUserRole(uid, config.Wechat.SuperAdmin)
	if err != nil {
		logrus.Warnf("get user role failed for uid %s: %v, using role from old token", uid, err)
		role = claims.Role // 使用旧token中的角色
	}

	// 6. 生成新的 access token 和 refresh token
	newAccessToken, err := GenerateAccessToken(uid, config.Jwt.Secret, config.Jwt.Expire, role)
	if err != nil {
		logrus.Errorf("generate new access token failed: %v", err)
		return nil, err
	}

	newRefreshToken, err := GenerateRefreshToken(uid, config.Jwt.Secret, config.Jwt.RefreshExpire, role)
	if err != nil {
		logrus.Errorf("generate new refresh token failed: %v", err)
		return nil, err
	}

	// 6. 将旧的 refresh token 加入撤销列表（Token Rotation 的核心）
	// 有效期设置为原 refresh token 的剩余有效时间
	remainingTime := claims.ExpiresAt.Unix() - claims.IssuedAt.Unix()
	if err := us.RevokeRefreshToken(refreshToken, remainingTime); err != nil {
		logrus.Errorf("revoke old refresh token failed: %v", err)
		// 这里不返回错误，因为新 token 已经生成，只是记录日志
	}

	// 7. 创建新的 token 对象
	tk := &UserToken{
		AccessToken:     newAccessToken,
		AccessTokenExp:  config.Jwt.Expire,
		RefreshToken:    newRefreshToken,
		RefreshTokenExp: config.Jwt.RefreshExpire,
	}

	// 8. 更新 Redis 缓存
	marshal, err := json.Marshal(tk)
	if err != nil {
		logrus.Errorf("marshal token failed: %v", err)
		return nil, err
	}

	if err = us.Redis.Set(constants.RedisAccessTokenPrefix+uid, string(marshal), config.Jwt.Expire); err != nil {
		logrus.Errorf("update redis cache failed: %v", err)
		return nil, err
	}

	logrus.Infof("token refreshed successfully for uid: %s", uid)
	return tk, nil
}

// IsTokenBlacklisted 检查用户是否在黑名单中
func (us *UserService) IsTokenBlacklisted(uid string) (bool, error) {
	blacklistKey := constants.RedisTokenBlacklistPrefix + uid
	value, err := us.Redis.Get(blacklistKey)
	if err != nil {
		return false, err
	}
	return value != "", nil
}

// IsRefreshTokenRevoked 检查 refresh token 是否已被撤销
func (us *UserService) IsRefreshTokenRevoked(refreshToken string) (bool, error) {
	revokedKey := constants.RedisRevokedRefreshTokenPrefix + refreshToken
	value, err := us.Redis.Get(revokedKey)
	if err != nil {
		return false, err
	}
	return value != "", nil
}

// RevokeRefreshToken 将 refresh token 加入撤销列表
func (us *UserService) RevokeRefreshToken(refreshToken string, expireSeconds int64) error {
	revokedKey := constants.RedisRevokedRefreshTokenPrefix + refreshToken
	return us.Redis.Set(revokedKey, "revoked", expireSeconds)
}

// GetUserRole 获取用户角色
// 返回值：user, admin, super_admin
func (us *UserService) GetUserRole(uid string, superAdmin string) (string, error) {
	// 1. 检查是否是超级管理员（配置文件中硬编码）
	if uid == superAdmin {
		return constants.RoleSuperAdmin, nil
	}

	// 2. 检查Redis中是否是管理员
	adminKey := constants.RedisAdminUserPrefix + uid
	role, err := us.Redis.Get(adminKey)
	if err != nil {
		return constants.RoleUser, err
	}

	// 3. 如果Redis中存在，返回对应角色
	if role == constants.RoleAdmin || role == constants.RoleSuperAdmin {
		return role, nil
	}

	// 4. 默认返回普通用户
	return constants.RoleUser, nil
}

// IsAdmin 检查用户是否是管理员（包括超级管理员）
func (us *UserService) IsAdmin(uid string, superAdmin string) (bool, error) {
	role, err := us.GetUserRole(uid, superAdmin)
	if err != nil {
		return false, err
	}
	return role == constants.RoleAdmin || role == constants.RoleSuperAdmin, nil
}

// AddAdmin 添加管理员
// role: admin 或 super_admin
// 永不过期(设置为10年)
func (us *UserService) AddAdmin(uid string, role string) error {
	if role != constants.RoleAdmin && role != constants.RoleSuperAdmin {
		return errors.New("无效的角色类型")
	}
	adminKey := constants.RedisAdminUserPrefix + uid
	// 设置10年过期时间（相当于永不过期）
	return us.Redis.Set(adminKey, role, constants.AdminRoleExpireSeconds)
}

// RemoveAdmin 移除管理员
func (us *UserService) RemoveAdmin(uid string) error {
	adminKey := constants.RedisAdminUserPrefix + uid
	return us.Redis.Del(adminKey)
}

// ListAdmins 列出所有管理员（需要Redis支持扫描keys）
// 注意：这个方法在生产环境可能性能较差，建议使用专门的数据结构存储管理员列表
func (us *UserService) ListAdmins() (map[string]string, error) {
	// 这里简化实现，实际生产环境建议使用 Redis Set 存储管理员列表
	// 返回 map[openid]role
	admins := make(map[string]string)
	// 由于当前Redis工具类可能不支持SCAN，这里返回空实现
	// 生产环境应该使用 Redis SCAN 命令或维护一个单独的管理员列表
	return admins, nil
}

// SetOAuthSession 线程安全地设置 OAuth 会话映射
func (us *UserService) SetOAuthSession(projectId, sessionId string) {
	us.oauthSessionMu.Lock()
	defer us.oauthSessionMu.Unlock()
	us.OAuthSessionMap[projectId] = sessionId
}

// GetOAuthSession 线程安全地获取 OAuth 会话 ID
func (us *UserService) GetOAuthSession(projectId string) string {
	us.oauthSessionMu.RLock()
	defer us.oauthSessionMu.RUnlock()
	return us.OAuthSessionMap[projectId]
}

// RemoveOAuthSession 线程安全地移除 OAuth 会话映射
func (us *UserService) RemoveOAuthSession(projectId string) {
	us.oauthSessionMu.Lock()
	defer us.oauthSessionMu.Unlock()
	delete(us.OAuthSessionMap, projectId)
}
