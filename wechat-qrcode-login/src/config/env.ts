/**
 * Environment Configuration
 * Centralized configuration management from .env
 */

export const ENV_CONFIG = {
  // ==================== API Configuration ====================
  api: {
    // 开发环境：空字符串使用 Vite 代理（/api -> http://127.0.0.1:8443）
    // 生产环境：设置完整 URL
    baseUrl: import.meta.env.VITE_API_BASE_URL || '',
  },

  // ==================== WebSocket Configuration ====================
  websocket: {
    // 生产环境：实际的 WebSocket 服务器地址
    host: import.meta.env.VITE_WS_HOST,
    path: import.meta.env.VITE_WS_PATH || '/api/ws',
    timeout: Number(import.meta.env.VITE_WS_TIMEOUT) || 300000,
    getFullUrl(): string {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const url = `${protocol}//${this.host}${this.path}`;
      console.log('[ENV Config] WebSocket URL:', url);
      return url;
    },
  },

  // ==================== OAuth Configuration ====================
  oauth: {
    clientId: import.meta.env.VITE_OAUTH_CLIENT_ID || '',
    authorizeBaseUrl: import.meta.env.VITE_OAUTH_AUTHORIZE_BASE_URL || '',
    redirectUrl: import.meta.env.VITE_OAUTH_REDIRECT_URL || 'http://localhost:5173/#/oauth/callback',
    scope: import.meta.env.VITE_OAUTH_SCOPE || 'user:read offline_access',
    responseType: (import.meta.env.VITE_OAUTH_RESPONSE_TYPE || 'code') as 'code',
  },

  // ==================== WeChat Configuration ====================
  wechat: {
    loginProjectId: import.meta.env.VITE_LOGIN_PROJECT_ID || '',
    oauthProjectId: import.meta.env.VITE_OAUTH_PROJECT_ID || '',
  },
};

/**
 * Validate required environment variables
 * @throws Error if required variables are missing
 */
export function validateEnvConfig(): void {
  const errors: string[] = [];

  // Check OAuth configuration
  if (!ENV_CONFIG.oauth.clientId) {
    errors.push('VITE_OAUTH_CLIENT_ID is required');
  }
  if (!ENV_CONFIG.oauth.authorizeBaseUrl) {
    errors.push('VITE_OAUTH_AUTHORIZE_BASE_URL is required');
  }
  if (!ENV_CONFIG.oauth.redirectUrl) {
    errors.push('VITE_OAUTH_REDIRECT_URL is required');
  }

  // Check WeChat configuration
  if (!ENV_CONFIG.wechat.loginProjectId) {
    errors.push('VITE_LOGIN_PROJECT_ID is required');
  }
  if (!ENV_CONFIG.wechat.oauthProjectId) {
    errors.push('VITE_OAUTH_PROJECT_ID is required');
  }

  if (errors.length > 0) {
    const errorMessage = `Environment configuration errors:\n${errors.join('\n')}`;
    console.error(errorMessage);
    throw new Error(errorMessage);
  }
}

/**
 * Log current configuration (for debugging)
 */
export function logEnvConfig(): void {
  console.log('[ENV] Configuration loaded:', {
    api: ENV_CONFIG.api,
    websocket: {
      host: ENV_CONFIG.websocket.host,
      path: ENV_CONFIG.websocket.path,
      timeout: ENV_CONFIG.websocket.timeout,
      fullUrl: ENV_CONFIG.websocket.getFullUrl(),
    },
    oauth: ENV_CONFIG.oauth,
    wechat: ENV_CONFIG.wechat,
  });
}
