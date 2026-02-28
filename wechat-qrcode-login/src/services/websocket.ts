/**
 * WebSocket Login Service
 * Production-ready WebSocket client for WeChat QR code login flow
 */

import type { LoginStatus, LoginTokens } from "@/types/auth";
import { TIMING } from "@/constants/timing";
import {
  validateServerMessage,
  validateQrCodeData,
  validateTokenData,
  validateErrorData,
  validateClientMessage,
  type ServerMessage,
  type ClientMessage,
} from "@/types/websocket";
import { getCSRFToken, regenerateCSRFToken } from "@/utils/csrf";
import { ENV_CONFIG } from "@/config/env";

// Event callbacks
export interface WebSocketLoginCallbacks {
  onStatusChange: (status: LoginStatus) => void;
  onQrCodeReceived: (qrCodeUrl: string) => void;
  onTokenReceived: (tokens: LoginTokens) => void;
  onError: (message: string) => void;
}

/**
 * WebSocket Login Service Class
 * Manages the WebSocket connection and login flow
 */
export class WebSocketLoginService {
  private ws: WebSocket | null = null;
  private timeoutTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private callbacks: WebSocketLoginCallbacks;
  private currentStatus: LoginStatus = "connecting";
  private isManualClose = false;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 3;
  private projectId: string = '';

  constructor(callbacks: WebSocketLoginCallbacks) {
    this.callbacks = callbacks;
  }

  /**
   * Connect to WebSocket server
   * @param projectId WeChat project ID for login
   */
  connect(projectId: string): void {
    this.disconnect(); // Clean up any existing connection
    this.isManualClose = false;
    this.updateStatus("connecting");
    this.reconnectAttempts = 0;
    this.projectId = projectId;

    try {
      // ✅ 使用统一的环境配置（支持代理）
      const url = ENV_CONFIG.websocket.getFullUrl();
      console.log("[WS] Connecting to:", url);
      console.log("[WS] WebSocket host:", ENV_CONFIG.websocket.host);

      this.ws = new WebSocket(url);
      this.setupEventHandlers();
    } catch (error) {
      console.error("[WS] Connection error:", error);
      this.handleError("无法建立连接");
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    this.isManualClose = true;
    this.clearTimeout();
    this.stopHeartbeat();

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * Reconnect with attempt limit
   */
  private reconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.handleError("连接失败，请刷新页面重试");
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), TIMING.RECONNECT_MAX_DELAY);

    console.log(`[WS] Reconnecting... Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`);
    setTimeout(() => {
      if (!this.isManualClose) {
        this.connect(this.projectId);
      }
    }, delay);
  }

  /**
   * Setup WebSocket event handlers
   */
  private setupEventHandlers(): void {
    if (!this.ws) return;

    this.ws.onopen = () => {
      console.log("[WS] Connected");
      this.updateStatus("loading_qr");
      this.sendInitialMessage();
      this.startTimeout();
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      this.handleMessage(event);
    };

    this.ws.onerror = (event) => {
      console.error("[WS] Error:", event);
      // Error is followed by close, handle there
    };

    this.ws.onclose = (event) => {
      console.log("[WS] Closed:", event.code, event.reason);

      if (this.isManualClose) {
        return; // Expected close, no action needed
      }

      // Handle unexpected close based on current status
      if (this.currentStatus !== "success" && this.currentStatus !== "expired") {
        // Try to reconnect for transient errors
        if (event.code === 1006 || event.code === 1000) {
          this.reconnect();
        } else {
          this.handleError("连接已断开");
        }
      }
    };
  }

  /**
   * Send initial message to request QR code
   * Includes CSRF token for protection
   */
  private sendInitialMessage(): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return;
    }

    const message: ClientMessage = {
      type: "qrcode",
      msg: this.projectId,
      csrf_token: getCSRFToken(), // Include CSRF token
    };

    // Validate message before sending
    const validated = validateClientMessage(message);
    if (!validated) {
      console.error("[WS] Failed to validate client message");
      this.handleError("无法发送请求");
      return;
    }

    this.ws.send(JSON.stringify(validated));
    console.log("[WS] Sent initial message with CSRF token");
  }

  /**
   * Handle incoming messages from server
   * With strict validation using Zod schemas
   */
  private handleMessage(event: MessageEvent): void {
    try {
      if (event.data == "pong") {
        console.log("received pong from server")
        return;
      }

      // Parse JSON
      const rawData = JSON.parse(event.data);

      // Validate message structure
      const data = validateServerMessage(rawData);
      if (!data) {
        console.error("[WS] Invalid message structure, ignoring");
        return;
      }

      console.log("[WS] Received:", data.type);

      switch (data.type) {
        case "QRCODE":
          // Validate QR code data format
          if (!validateQrCodeData(data.data)) {
            console.error("[WS] Invalid QR code data, ignoring");
            this.handleError("接收到无效的二维码数据");
            return;
          }
          this.handleQrCode(data.data);
          break;

        case "SCANNED":
          this.handleScanned();
          break;

        case "TOKEN":
          // Validate token data format
          if (!validateTokenData(data.data)) {
            console.error("[WS] Invalid token data, ignoring");
            this.handleError("接收到无效的登录凭证");
            return;
          }
          this.handleToken(data.data);
          break;

        case "EXPIRED":
          this.handleExpired();
          break;

        case "ERROR":
          // Validate error data
          if (!validateErrorData(data.data)) {
            console.error("[WS] Invalid error data, using default");
            this.handleError("服务器错误");
            return;
          }
          this.handleError(data.data);
          break;

        case "REDIRECT":
          // OAuth redirect - not used in normal login flow
          console.warn("[WS] Received REDIRECT message in normal login flow, ignoring");
          break;

        default:
          // TypeScript exhaustiveness check
          const _exhaustive: never = data.type;
          console.warn("[WS] Unknown message type:", _exhaustive);
      }
    } catch (error) {
      console.error("[WS] Failed to parse message:", error);
      this.handleError("消息解析失败");
    }
  }

  /**
   * Handle QR code received from server
   */
  private handleQrCode(qrCodeUrl: string): void {
    this.clearTimeout(); // Reset timeout on QR code received
    this.updateStatus("waiting_for_scan");
    this.callbacks.onQrCodeReceived(qrCodeUrl);
    this.startTimeout(); // Restart timeout
  }

  /**
   * Handle QR code scanned by user
   */
  private handleScanned(): void {
    this.clearTimeout();
    this.updateStatus("scanned");
    // Start a shorter timeout for confirmation
    this.timeoutTimer = setTimeout(() => {
      this.handleExpired();
    }, TIMING.CONFIRMATION_TIMEOUT);
  }

  /**
   * Handle token received (login success)
   */
  private handleToken(tokenData: string): void {
    try {
      const tokens = JSON.parse(tokenData) as LoginTokens;
      console.log("[WS] Login successful");

      // Regenerate CSRF token after successful authentication
      regenerateCSRFToken();
      console.log("[WS] CSRF token regenerated");

      this.clearTimeout();
      this.disconnect(); // Clean close
      this.updateStatus("success");
      this.callbacks.onTokenReceived(tokens);
    } catch (error) {
      console.error("[WS] Failed to parse token:", error);
      this.handleError("无效的令牌响应");
    }
  }

  /**
   * Handle QR code expired
   */
  private handleExpired(): void {
    this.clearTimeout();
    this.disconnect();
    this.updateStatus("expired");
  }

  /**
   * Handle error
   */
  private handleError(message: string): void {
    this.clearTimeout();
    this.updateStatus("error");
    this.callbacks.onError(message);
  }

  /**
   * Start timeout timer
   */
  private startTimeout(): void {
    this.clearTimeout();
    this.timeoutTimer = setTimeout(() => {
      console.warn("[WS] Timeout");
      this.handleExpired();
    }, ENV_CONFIG.websocket.timeout);
  }

  /**
   * Clear timeout timer
   */
  private clearTimeout(): void {
    if (this.timeoutTimer) {
      clearTimeout(this.timeoutTimer);
      this.timeoutTimer = null;
    }
  }

  /**
   * Update status and notify callback
   */
  private updateStatus(status: LoginStatus): void {
    this.currentStatus = status;
    this.callbacks.onStatusChange(status);
  }

  /**
   * Get current connection status
   */
  getStatus(): LoginStatus {
    return this.currentStatus;
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Start heartbeat to keep connection alive
   */
  private startHeartbeat(): void {
    this.stopHeartbeat(); // Clear any existing heartbeat

    this.heartbeatTimer = setInterval(() => {
      if (this.isConnected()) {
        try {
          // Send ping message to keep connection alive with CSRF token
          const pingMessage: ClientMessage = {
            type: "ping",
            msg: "heartbeat",
            csrf_token: getCSRFToken(), // Include CSRF token
          };

          // Validate before sending
          const validated = validateClientMessage(pingMessage);
          if (validated) {
            this.ws?.send(JSON.stringify(validated));
            console.log("[WS] Heartbeat sent");
          }
        } catch (error) {
          console.error("[WS] Failed to send heartbeat:", error);
        }
      }
    }, TIMING.HEARTBEAT_INTERVAL);
  }

  /**
   * Stop heartbeat timer
   */
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
}
