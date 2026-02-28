/**
 * CSRF (Cross-Site Request Forgery) Protection Utilities
 * Generates and manages CSRF tokens for WebSocket communication
 */

const CSRF_TOKEN_KEY = '__csrf_token__';

/**
 * Generate a cryptographically secure CSRF token
 * @returns A random UUID v4 token
 */
export function generateCSRFToken(): string {
  return crypto.randomUUID();
}

/**
 * Get or generate CSRF token for current session
 * Token is stored in sessionStorage and persists for the browser session
 * @returns Current CSRF token
 */
export function getCSRFToken(): string {
  // Check if token exists in sessionStorage
  let token = sessionStorage.getItem(CSRF_TOKEN_KEY);

  if (!token) {
    // Generate new token if not exists
    token = generateCSRFToken();
    sessionStorage.setItem(CSRF_TOKEN_KEY, token);
  }

  return token;
}

/**
 * Validate CSRF token
 * @param token Token to validate
 * @returns true if token matches current session token
 */
export function validateCSRFToken(token: string): boolean {
  const currentToken = sessionStorage.getItem(CSRF_TOKEN_KEY);
  return currentToken !== null && currentToken === token;
}

/**
 * Clear CSRF token (useful on logout)
 */
export function clearCSRFToken(): void {
  sessionStorage.removeItem(CSRF_TOKEN_KEY);
}

/**
 * Regenerate CSRF token
 * Useful after successful authentication
 * @returns New CSRF token
 */
export function regenerateCSRFToken(): string {
  const newToken = generateCSRFToken();
  sessionStorage.setItem(CSRF_TOKEN_KEY, newToken);
  return newToken;
}
