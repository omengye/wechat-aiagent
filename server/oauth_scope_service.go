package server

import (
	"errors"
	"fmt"

	"github.com/omengye/wechat_aiagent/constants"
)

// OAuthScopeService OAuth Scope 验证服务
type OAuthScopeService struct{}

// NewOAuthScopeService 创建 Scope 服务
func NewOAuthScopeService() *OAuthScopeService {
	return &OAuthScopeService{}
}

// ValidateScopesForRole 验证用户角色是否满足所有请求的 scopes
func (s *OAuthScopeService) ValidateScopesForRole(scopes []string, userRole string) error {
	for _, scope := range scopes {
		requiredRole, exists := constants.ScopeRequiredRole[scope]
		if !exists {
			return fmt.Errorf("未知的 scope: %s", scope)
		}

		if !s.hasPermission(userRole, requiredRole) {
			return fmt.Errorf("scope '%s' 需要 %s 角色，当前角色为 %s", scope, requiredRole, userRole)
		}
	}
	return nil
}

// hasPermission 检查用户角色是否满足所需角色
// 角色层级：super_admin > admin > user
func (s *OAuthScopeService) hasPermission(userRole, requiredRole string) bool {
	roleLevel := map[string]int{
		constants.RoleUser:       1,
		constants.RoleAdmin:      2,
		constants.RoleSuperAdmin: 3,
	}

	userLevel, userExists := roleLevel[userRole]
	requiredLevel, requiredExists := roleLevel[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// HasScope 检查 token 的 scopes 中是否包含指定的 scope
func (s *OAuthScopeService) HasScope(tokenScopes []string, requiredScope string) bool {
	for _, scope := range tokenScopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

// HasAllScopes 检查 token 的 scopes 中是否包含所有指定的 scopes
func (s *OAuthScopeService) HasAllScopes(tokenScopes []string, requiredScopes []string) bool {
	scopeMap := make(map[string]bool)
	for _, scope := range tokenScopes {
		scopeMap[scope] = true
	}

	for _, required := range requiredScopes {
		if !scopeMap[required] {
			return false
		}
	}
	return true
}

// HasAnyScope 检查 token 的 scopes 中是否包含任意一个指定的 scope
func (s *OAuthScopeService) HasAnyScope(tokenScopes []string, requiredScopes []string) bool {
	scopeMap := make(map[string]bool)
	for _, scope := range tokenScopes {
		scopeMap[scope] = true
	}

	for _, required := range requiredScopes {
		if scopeMap[required] {
			return true
		}
	}
	return false
}

// ValidateScopes 验证 scope 字符串是否有效
func (s *OAuthScopeService) ValidateScopes(scopes []string) error {
	if len(scopes) == 0 {
		return errors.New("scope 不能为空")
	}

	for _, scope := range scopes {
		if _, exists := constants.ScopeRequiredRole[scope]; !exists {
			return fmt.Errorf("无效的 scope: %s", scope)
		}
	}

	return nil
}

// GetScopeDisplayNames 获取 scopes 的显示名称
func (s *OAuthScopeService) GetScopeDisplayNames(scopes []string) []string {
	displayNames := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		if displayName, exists := constants.ScopeDisplayNames[scope]; exists {
			displayNames = append(displayNames, displayName)
		} else {
			displayNames = append(displayNames, scope)
		}
	}
	return displayNames
}
