/**
 * OAuth API Service
 * Handles OAuth authorization flow
 */

import type { OAuthAuthorizeParams, OAuthAuthorizeResponse, OAuthError } from '@/types/oauth';
import { ENV_CONFIG } from '@/config/env';
import { httpClient } from '@/utils/httpClient';

/**
 * Request OAuth authorization
 * @param params OAuth authorization parameters
 * @returns OAuth authorization response with session_id
 */
export async function requestOAuthAuthorization(
  params: OAuthAuthorizeParams
): Promise<OAuthAuthorizeResponse> {
  // Build URL with query parameters
  const baseUrl = ENV_CONFIG.oauth.authorizeBaseUrl;
  const path = '/api/oauth/authorize';

  // 构造查询字符串
  const queryParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value) {
      queryParams.set(key, value);
    }
  });

  // 根据 baseUrl 是否为空来构造完整 URL
  const fullUrl = baseUrl
    ? `${baseUrl}${path}?${queryParams.toString()}`  // 生产环境：完整 URL
    : `${path}?${queryParams.toString()}`;            // 开发环境：相对路径（走代理）

  console.log('[OAuth API] Request URL:', fullUrl);

  try {
    // Use httpClient which automatically includes Origin header
    const data = await httpClient.get<OAuthAuthorizeResponse>(fullUrl);

    // Validate response
    if (!data.session_id || !data.project_id) {
      throw new Error('Invalid OAuth authorization response');
    }

    console.log('[OAuth API] Authorization successful:', {
      session_id: data.session_id,
      client_name: data.client_name,
      project_id: data.project_id,
    });

    return data;
  } catch (error) {
    console.error('[OAuth API] Authorization request failed:', error);
    throw error;
  }
}

/**
 * Generate random state for CSRF protection
 */
export function generateState(): string {
  return Math.random().toString(36).substring(2, 15) +
         Math.random().toString(36).substring(2, 15);
}

/**
 * Get OAuth parameters from environment
 */
export function getDefaultOAuthParams(): Partial<OAuthAuthorizeParams> {
  return {
    client_id: ENV_CONFIG.oauth.clientId,
    redirect_url: ENV_CONFIG.oauth.redirectUrl,
    response_type: ENV_CONFIG.oauth.responseType,
    scope: ENV_CONFIG.oauth.scope,
  };
}

/**
 * Exchange authorization code for access token
 * @param code Authorization code from callback URL
 * @param state State parameter for CSRF validation
 * @returns Access token and refresh token
 */
export async function exchangeCodeForToken(
  code: string,
  state?: string
): Promise<{ access_token: string; refresh_token: string; expires_in: number }> {
  const baseUrl = ENV_CONFIG.api.baseUrl;
  const path = '/api/oauth/token';

  try {
    console.log('[OAuth API] Exchanging code for token:', { code, state });

    // 构造请求体
    const requestBody = {
      grant_type: 'authorization_code',
      code: code,
      client_id: ENV_CONFIG.oauth.clientId,
      redirect_url: ENV_CONFIG.oauth.redirectUrl,
      state: state,
    };

    // 构造完整 URL
    const fullUrl = baseUrl ? `${baseUrl}${path}` : path;

    console.log('[OAuth API] Token request URL:', fullUrl);
    console.log('[OAuth API] Token request body:', requestBody);

    // 使用 httpClient 发送 POST 请求
    const data = await httpClient.post<{
      access_token: string;
      refresh_token: string;
      expires_in: number;
      token_type: string;
    }>(fullUrl, requestBody);

    console.log('[OAuth API] Token exchange successful:', {
      token_type: data.token_type,
      expires_in: data.expires_in,
    });

    return {
      access_token: data.access_token,
      refresh_token: data.refresh_token,
      expires_in: data.expires_in,
    };
  } catch (error) {
    console.error('[OAuth API] Token exchange failed:', error);
    throw error;
  }
}
