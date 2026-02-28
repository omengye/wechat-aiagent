/**
 * Timing Constants
 * Centralized timing configuration to avoid magic numbers
 */

export const TIMING = {
  // WebSocket related
  HEARTBEAT_INTERVAL: 30 * 1000, // 30 seconds - ping interval
  RECONNECT_MAX_DELAY: 10 * 1000, // 10 seconds - max reconnect delay

  // QR Code related
  QR_CODE_TIMEOUT: 5 * 60 * 1000, // 5 minutes - QR code expiration
  CONFIRMATION_TIMEOUT: 60 * 1000, // 1 minute - confirmation timeout after scan

  // UI related
  REDIRECT_DELAY: 1500, // 1.5 seconds - delay before redirect after success
  ANIMATION_DURATION: 300, // 300ms - standard animation duration

  // Countdown
  COUNTDOWN_UPDATE_INTERVAL: 1000, // 1 second - countdown update frequency
} as const;

export type TimingKey = keyof typeof TIMING;
