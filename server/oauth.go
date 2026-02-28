package server

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/omengye/wechat_aiagent/constants"
)

// Claims JWT声明结构
type Claims struct {
	Uid       string   `json:"uid"`
	TokenType string   `json:"token_type"` // access 或 refresh
	Role      string   `json:"role"`       // user, admin, super_admin
	Scope     []string `json:"scope,omitempty"` // OAuth scope
	ClientID  string   `json:"client_id,omitempty"` // OAuth 客户端 ID
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成访问令牌
// uid: 用户ID
// secret: JWT签名密钥
// expire: 过期时间(秒)
// role: 用户角色
func GenerateAccessToken(uid string, secret string, expire int64, role string) (string, error) {
	return generateToken(uid, secret, expire, constants.TokenTypeAccess, role)
}

// GenerateRefreshToken 生成刷新令牌
// uid: 用户ID
// secret: JWT签名密钥
// expire: 过期时间(秒)
// role: 用户角色
func GenerateRefreshToken(uid string, secret string, expire int64, role string) (string, error) {
	return generateToken(uid, secret, expire, constants.TokenTypeRefresh, role)
}

// RefreshAccessToken 使用刷新令牌生成新的访问令牌
// refreshToken: 刷新令牌字符串
// secret: JWT签名密钥
// expire: 新访问令牌的过期时间(秒)
func RefreshAccessToken(refreshToken string, secret string, expire int64) (string, error) {
	claims, err := ParseToken(refreshToken, secret)
	if err != nil {
		return "", err
	}

	if claims.TokenType != constants.TokenTypeRefresh {
		return "", errors.New("invalid token type")
	}

	return GenerateAccessToken(claims.Uid, secret, expire, claims.Role)
}

// generateToken 生成JWT令牌的内部函数
func generateToken(uid string, secret string, expire int64, tokenType string, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		Uid:       uid,
		TokenType: tokenType,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并验证JWT令牌
// tokenString: JWT令牌字符串
// secret: JWT签名密钥
func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateOAuthAccessToken 生成 OAuth 访问令牌（带 scope 和 client_id）
func GenerateOAuthAccessToken(uid string, secret string, expire int64, role string, scope []string, clientID string) (string, error) {
	return generateOAuthToken(uid, secret, expire, constants.TokenTypeAccess, role, scope, clientID)
}

// GenerateOAuthRefreshToken 生成 OAuth 刷新令牌（带 scope 和 client_id）
func GenerateOAuthRefreshToken(uid string, secret string, expire int64, role string, scope []string, clientID string) (string, error) {
	return generateOAuthToken(uid, secret, expire, constants.TokenTypeRefresh, role, scope, clientID)
}

// generateOAuthToken 生成 OAuth JWT 令牌的内部函数
func generateOAuthToken(uid string, secret string, expire int64, tokenType string, role string, scope []string, clientID string) (string, error) {
	now := time.Now()
	claims := Claims{
		Uid:       uid,
		TokenType: tokenType,
		Role:      role,
		Scope:     scope,
		ClientID:  clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
