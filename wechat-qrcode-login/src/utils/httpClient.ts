/**
 * HTTP Client Utility
 * Centralized HTTP client with default headers including Origin
 */

import { ENV_CONFIG } from '@/config/env';

/**
 * Get default headers for all requests
 */
function getDefaultHeaders(): HeadersInit {
  // 从环境变量读取 Origin，默认使用 API base URL
  const origin = import.meta.env.VITE_API_ORIGIN || ENV_CONFIG.api.baseUrl || window.location.origin;

  console.log('[HTTP Client] Setting X-Requested-Origin header:', origin);

  return {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
    // 注意：浏览器会自动设置 Origin header，这里使用自定义 header
    // 后端可以同时检查 Origin 和 X-Requested-Origin
    'X-Requested-Origin': origin,
  };
}

/**
 * HTTP Client with default configuration
 */
export const httpClient = {
  /**
   * GET request
   */
  async get<T>(url: string, options?: RequestInit): Promise<T> {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        ...getDefaultHeaders(),
        ...options?.headers,
      },
      credentials: 'include',
      ...options,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Request failed' }));
      throw new Error(error.error_description || error.message || `HTTP ${response.status}`);
    }

    return response.json();
  },

  /**
   * POST request
   */
  async post<T>(url: string, data?: any, options?: RequestInit): Promise<T> {
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        ...getDefaultHeaders(),
        ...options?.headers,
      },
      credentials: 'include',
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Request failed' }));
      throw new Error(error.error_description || error.message || `HTTP ${response.status}`);
    }

    return response.json();
  },

  /**
   * PUT request
   */
  async put<T>(url: string, data?: any, options?: RequestInit): Promise<T> {
    const response = await fetch(url, {
      method: 'PUT',
      headers: {
        ...getDefaultHeaders(),
        ...options?.headers,
      },
      credentials: 'include',
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Request failed' }));
      throw new Error(error.error_description || error.message || `HTTP ${response.status}`);
    }

    return response.json();
  },

  /**
   * DELETE request
   */
  async delete<T>(url: string, options?: RequestInit): Promise<T> {
    const response = await fetch(url, {
      method: 'DELETE',
      headers: {
        ...getDefaultHeaders(),
        ...options?.headers,
      },
      credentials: 'include',
      ...options,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: 'Request failed' }));
      throw new Error(error.error_description || error.message || `HTTP ${response.status}`);
    }

    return response.json();
  },
};

/**
 * Build full API URL
 */
export function buildApiUrl(path: string): string {
  return `${ENV_CONFIG.api.baseUrl}${path}`;
}
