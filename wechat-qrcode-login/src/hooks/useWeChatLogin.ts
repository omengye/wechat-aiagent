/**
 * React Hook for WeChat QR Code Login
 * Provides a clean interface for the WebSocket login flow
 */

import { useState, useCallback, useRef, useEffect } from "react";
import type { LoginStatus, LoginTokens } from "@/types/auth";
import { WebSocketLoginService } from "@/services/websocket";

export function useWeChatLogin(
  projectId: string,
  onSuccess: (tokens: LoginTokens) => void | Promise<void>
) {
  const [status, setStatus] = useState<LoginStatus>("connecting");
  const [qrCodeUrl, setQrCodeUrl] = useState<string>("");
  const [errorMsg, setErrorMsg] = useState<string>("");

  // Use ref to store the service instance
  const serviceRef = useRef<WebSocketLoginService | null>(null);
  const onSuccessRef = useRef(onSuccess);
  const isMountedRef = useRef(true);
  const projectIdRef = useRef(projectId);

  // Keep refs in sync
  useEffect(() => {
    onSuccessRef.current = onSuccess;
    projectIdRef.current = projectId;
  }, [onSuccess, projectId]);

  // Initialize service on mount
  useEffect(() => {
    isMountedRef.current = true;

    // Create service instance with callbacks
    const service = new WebSocketLoginService({
      onStatusChange: (newStatus) => {
        if (isMountedRef.current) {
          setStatus(newStatus);
        }
      },
      onQrCodeReceived: (url) => {
        if (isMountedRef.current) {
          setQrCodeUrl(url);
        }
      },
      onTokenReceived: (tokens) => {
        if (isMountedRef.current) {
          onSuccessRef.current(tokens);
        }
      },
      onError: (message) => {
        if (isMountedRef.current) {
          setErrorMsg(message);
        }
      },
    });

    serviceRef.current = service;

    // Auto-connect on mount with projectId
    service.connect(projectIdRef.current);

    // Cleanup on unmount
    return () => {
      isMountedRef.current = false;
      service.disconnect();
      serviceRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run on mount

  // Refresh/reconnect function
  const refresh = useCallback(() => {
    setErrorMsg("");
    setQrCodeUrl("");
    serviceRef.current?.connect(projectIdRef.current);
  }, []);

  // Manual disconnect function
  const disconnect = useCallback(() => {
    serviceRef.current?.disconnect();
  }, []);

  return {
    status,
    qrCodeUrl,
    errorMsg,
    refresh,
    disconnect,
  };
}
