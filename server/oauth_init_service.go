package server

import (
	"github.com/omengye/wechat_aiagent/constants"
	"github.com/omengye/wechat_aiagent/tools"
	"github.com/omengye/wechat_aiagent/types"
	"github.com/sirupsen/logrus"
)

// InitOAuthClientsFromConfig 从配置文件初始化 OAuth 客户端
// 在系统启动时调用，将配置文件中的客户端信息同步到 Redis
func InitOAuthClientsFromConfig(config *types.Config, redis *tools.RedisMem) error {
	clientService := NewOAuthClientService(redis)

	// 遍历所有项目
	for _, project := range config.Wechat.Qrcode {
		// 遍历该项目下的所有 OAuth 客户端配置
		for _, clientConfig := range project.OAuthClients {
			if err := initSingleClient(clientService, &clientConfig, project.ProjectId); err != nil {
				logrus.Errorf("failed to init OAuth client %s for project %s: %v",
					clientConfig.ClientID, project.ProjectId, err)
				return err
			}
		}
	}

	return nil
}

// initSingleClient 初始化单个 OAuth 客户端
func initSingleClient(clientService *OAuthClientService, config *types.OAuthClientConfig, projectID string) error {
	// 1. 验证必填字段
	if config.ClientID == "" {
		logrus.Errorf("clientId is required for OAuth client in project %s", projectID)
		return nil // 跳过无效配置
	}
	if config.ClientSecret == "" {
		logrus.Errorf("clientSecret is required for OAuth client %s", config.ClientID)
		return nil
	}
	if len(config.RedirectURLs) == 0 {
		logrus.Errorf("redirectUrls is required for OAuth client %s", config.ClientID)
		return nil
	}
	if len(config.AllowedScopes) == 0 {
		logrus.Errorf("allowedScopes is required for OAuth client %s", config.ClientID)
		return nil
	}

	// 2. 检查客户端是否已存在
	existingClient, err := clientService.GetClient(config.ClientID)
	if err == nil && existingClient != nil {
		// 客户端已存在，检查是否需要更新
		logrus.Infof("OAuth client %s already exists, checking for updates...", config.ClientID)

		// 检查密钥是否匹配（验证配置文件中的明文密钥）
		if err := clientService.VerifyClientSecret(config.ClientSecret, existingClient.ClientSecretHash); err != nil {
			// 密钥不匹配，说明配置文件中的密钥已更新，需要重新加密
			logrus.Infof("OAuth client %s secret changed, updating...", config.ClientID)
			hashedSecret, err := clientService.HashClientSecret(config.ClientSecret)
			if err != nil {
				return err
			}

			// 更新客户端信息
			existingClient.ClientSecretHash = hashedSecret
			existingClient.Name = config.Name
			existingClient.Description = config.Description
			existingClient.RedirectURLs = config.RedirectURLs
			existingClient.AllowedScopes = config.AllowedScopes
			if config.RateLimit > 0 {
				existingClient.RateLimit = config.RateLimit
			}

			if err := clientService.saveClient(existingClient); err != nil {
				return err
			}
			logrus.Infof("✓ OAuth client %s (%s) updated successfully", config.ClientID, config.Name)
		} else {
			logrus.Infof("✓ OAuth client %s (%s) is up to date", config.ClientID, config.Name)
		}
		return nil
	}

	// 3. 客户端不存在，创建新客户端
	hashedSecret, err := clientService.HashClientSecret(config.ClientSecret)
	if err != nil {
		return err
	}

	// 设置默认限流值
	rateLimit := config.RateLimit
	if rateLimit == 0 {
		rateLimit = constants.RateLimitTokenPerHour
	}

	// 创建客户端对象
	client := &types.OAuthClient{
		ClientID:         config.ClientID,
		ClientSecretHash: hashedSecret,
		Name:             config.Name,
		Description:      config.Description,
		RedirectURLs:     config.RedirectURLs,
		AllowedScopes:    config.AllowedScopes,
		CreatedBy:        "config_file",
		Status:           constants.OAuthClientStatusActive,
		RateLimit:        rateLimit,
	}

	// 保存到 Redis
	if err := clientService.saveClient(client); err != nil {
		return err
	}

	logrus.Infof("✓ OAuth client %s (%s) initialized successfully", config.ClientID, config.Name)
	return nil
}
