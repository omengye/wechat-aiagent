/**
 * OAuth WebSocket Service
 * Handles WebSocket connection for OAuth authorization flow
 */

import type { LoginStatus } from '@/types/auth';
import { TIMING } from '@/constants/timing';
import {
  validateServerMessage,
  validateQrCodeData,
  type ServerMessage,
} from '@/types/websocket';
import { ENV_CONFIG } from '@/config/env';

// Event callbacks
export interface OAuthWebSocketCallbacks {
  onStatusChange: (status: LoginStatus) => void;
  onQrCodeReceived: (qrCodeUrl: string) => void;
  onRedirectReceived: (redirectUrl: string) => void;
  onError: (message: string) => void;
}

/**
 * OAuth WebSocket Service Class
 * Manages the WebSocket connection for OAuth authorization
 */
export class OAuthWebSocketService {
  private ws: WebSocket | null = null;
  private callbacks: OAuthWebSocketCallbacks;
  private currentStatus: LoginStatus = 'connecting';
  private isManualClose = false;

  constructor(callbacks: OAuthWebSocketCallbacks) {
    this.callbacks = callbacks;
  }

  /**
   * Connect to WebSocket and send OAuth session ID
   */
  connect(projectId: string, sessionId: string): void {
    this.disconnect();
    this.isManualClose = false;
    this.updateStatus('connecting');

    try {
      // ✅ 使用统一的环境配置（支持代理）
      const url = ENV_CONFIG.websocket.getFullUrl();
      console.log('[OAuth WS] Connecting to:', url);
      console.log('[OAuth WS] WebSocket host:', ENV_CONFIG.websocket.host);
      console.log('[OAuth WS] WebSocket path:', ENV_CONFIG.websocket.path);

      this.ws = new WebSocket(url);
      this.setupEventHandlers(projectId, sessionId);
    } catch (error) {
      console.error('[OAuth WS] Connection error:', error);
      this.handleError('无法建立连接');
    }
  }

  /**
   * Disconnect from WebSocket
   */
  disconnect(): void {
    this.isManualClose = true;

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Setup WebSocket event handlers
   */
  private setupEventHandlers(projectId: string, sessionId: string): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      console.log('[OAuth WS] Connected');
      this.updateStatus('loading_qr');
      this.sendOAuthMessage(projectId, sessionId);
    };

    this.ws.onmessage = (event) => {
      this.handleMessage(event);
    };

    this.ws.onerror = (event) => {
      console.error('[OAuth WS] Error:', event);
    };

    this.ws.onclose = (event) => {
      console.log('[OAuth WS] Closed:', event.code, event.reason);

      if (!this.isManualClose && this.currentStatus !== 'success') {
        this.handleError('连接已断开');
      }
    };
  }

  /**
   * Send OAuth authorization message with session ID
   */
  private sendOAuthMessage(projectId: string, sessionId: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[OAuth WS] Cannot send message - WebSocket not ready');
      return;
    }

    // 验证参数
    if (!sessionId) {
      console.error('[OAuth WS] ⚠️ sessionId is empty!');
    }
    if (!projectId) {
      console.error('[OAuth WS] ⚠️ projectId is empty!');
    }

    const message = {
      type: 'qrcode',
      msg: projectId,
      oauth_session_id: sessionId, // ✅ 携带 OAuth session ID
    };

    // 详细日志
    console.log('[OAuth WS] 📤 Sending message:', JSON.stringify(message, null, 2));
    console.log('[OAuth WS] Session ID:', sessionId);
    console.log('[OAuth WS] Project ID:', projectId);

    this.ws.send(JSON.stringify(message));

    console.log('[OAuth WS] ✅ Message sent successfully');
  }

  /**
   * Handle incoming messages from server
   */
  private handleMessage(event: MessageEvent): void {
    try {
      const rawData = JSON.parse(event.data);
      const data = validateServerMessage(rawData);

      if (!data) {
        console.error('[OAuth WS] Invalid message structure');
        return;
      }

      console.log('[OAuth WS] Received:', data.type);

      switch (data.type) {
        case 'QRCODE':
          if (!validateQrCodeData(data.data)) {
            console.error('[OAuth WS] Invalid QR code data');
            this.handleError('接收到无效的二维码数据');
            return;
          }
          this.handleQrCode(data.data);
          break;

        case 'REDIRECT':
          // OAuth authorization complete, redirect to callback URL
          this.handleRedirect(data.data);
          break;

        case 'ERROR':
          this.handleError(data.data || '服务器错误');
          break;

        default:
          console.warn('[OAuth WS] Unknown message type:', data.type);
      }
    } catch (error) {
      console.error('[OAuth WS] Failed to parse message:', error);
      this.handleError('消息解析失败');
    }
  }

  /**
   * Handle QR code received
   */
  private handleQrCode(qrCodeUrl: string): void {
    this.updateStatus('waiting_for_scan');
    this.callbacks.onQrCodeReceived(qrCodeUrl);
  }

  /**
   * Handle redirect URL received (OAuth flow complete)
   */
  private handleRedirect(redirectUrl: string): void {
    console.log('[OAuth WS] Redirect URL received:', redirectUrl);
    this.updateStatus('success');
    this.disconnect();
    this.callbacks.onRedirectReceived(redirectUrl);
  }

  /**
   * Handle error
   */
  private handleError(message: string): void {
    this.updateStatus('error');
    this.callbacks.onError(message);
  }

  /**
   * Update status and notify callback
   */
  private updateStatus(status: LoginStatus): void {
    this.currentStatus = status;
    this.callbacks.onStatusChange(status);
  }

  /**
   * Get current status
   */
  getStatus(): LoginStatus {
    return this.currentStatus;
  }
}
