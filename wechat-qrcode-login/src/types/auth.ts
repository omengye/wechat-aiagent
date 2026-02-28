export type LoginStatus = 
  | 'connecting'        // Connecting to WebSocket
  | 'loading_qr'        // Connected, waiting for QR code
  | 'waiting_for_scan'  // QR displayed, waiting for user to scan
  | 'scanned'           // User scanned, waiting for confirmation on phone
  | 'success'           // Login success, token received
  | 'expired'           // QR code expired (timeout)
  | 'error';            // Connection error

export interface LoginTokens {
  access_token: string;
  refresh_token: string;
}
