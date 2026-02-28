/**
 * WebSocket message schemas and validation
 * Uses Zod for runtime type validation
 */
import { z } from "zod";

/**
 * Server message types
 */
export const ServerMessageTypeSchema = z.enum([
  "QRCODE",
  "TOKEN",
  "SCANNED",
  "EXPIRED",
  "ERROR",
  "REDIRECT", // OAuth redirect message
]);

export type ServerMessageType = z.infer<typeof ServerMessageTypeSchema>;

/**
 * Base server message schema
 */
export const ServerMessageSchema = z.object({
  type: ServerMessageTypeSchema,
  data: z.string(),
});

export type ServerMessage = z.infer<typeof ServerMessageSchema>;

/**
 * QR Code data schema
 * Must be a data URL with image/png format
 */
export const QrCodeDataSchema = z
  .string()
  .min(1, "QR code data cannot be empty")
  .refine(
    (val) => val.startsWith("https://"),
    "QR code must be a valid data URL with image format",
  )
  .refine(
    (val) => val.length < 1024, // 1KB max
    "QR code data too large (max 1KB)",
  );

/**
 * Token data schema
 * Must be valid JSON containing access_token and refresh_token
 */
export const TokenDataSchema = z.string().refine(
  (val) => {
    try {
      const parsed = JSON.parse(val);
      return (
        typeof parsed === "object" &&
        parsed !== null &&
        typeof parsed.access_token === "string" &&
        typeof parsed.refresh_token === "string" &&
        parsed.access_token.length >= 10 &&
        parsed.refresh_token.length >= 10
      );
    } catch {
      return false;
    }
  },
  {
    message:
      "Token data must be valid JSON with access_token and refresh_token fields",
  },
);

/**
 * Error message data schema
 */
export const ErrorDataSchema = z
  .string()
  .min(1, "Error message cannot be empty")
  .max(1000, "Error message too long");

/**
 * Client message schema
 */
export const ClientMessageSchema = z.object({
  type: z.string().min(1).max(50),
  msg: z.string().min(1).max(1000),
  csrf_token: z.string().optional(), // CSRF protection token
  oauth_session_id: z.string().optional(), // OAuth session ID for OAuth flow
});

export type ClientMessage = z.infer<typeof ClientMessageSchema>;

/**
 * Validate and parse server message
 */
export function validateServerMessage(data: unknown): ServerMessage | null {
  try {
    const parsed = ServerMessageSchema.safeParse(data);
    if (!parsed.success) {
      console.error("[WS Validation] Invalid message structure:", parsed.error);
      return null;
    }
    return parsed.data;
  } catch (error) {
    console.error("[WS Validation] Failed to validate message:", error);
    return null;
  }
}

/**
 * Validate QR code data
 */
export function validateQrCodeData(data: string): boolean {
  const result = QrCodeDataSchema.safeParse(data);
  if (!result.success) {
    console.error("[WS Validation] Invalid QR code data:", result.error);
    return false;
  }
  return true;
}

/**
 * Validate token data
 */
export function validateTokenData(data: string): boolean {
  const result = TokenDataSchema.safeParse(data);
  if (!result.success) {
    console.error("[WS Validation] Invalid token data:", result.error);
    return false;
  }
  return true;
}

/**
 * Validate error data
 */
export function validateErrorData(data: string): boolean {
  const result = ErrorDataSchema.safeParse(data);
  if (!result.success) {
    console.error("[WS Validation] Invalid error data:", result.error);
    return false;
  }
  return true;
}

/**
 * Validate client message before sending
 */
export function validateClientMessage(message: unknown): ClientMessage | null {
  const result = ClientMessageSchema.safeParse(message);
  if (!result.success) {
    console.error("[WS Validation] Invalid client message:", result.error);
    return null;
  }
  return result.data;
}
