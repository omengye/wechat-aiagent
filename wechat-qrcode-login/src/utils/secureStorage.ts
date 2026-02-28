/**
 * Secure Storage Utility
 * Enhanced security using Web Crypto API with AES-GCM encryption
 *
 * Security improvements:
 * - Uses AES-GCM (256-bit) for strong encryption
 * - Generates unique session key per browser session
 * - Uses cryptographically secure random IVs
 * - Key stored in memory only (non-extractable)
 * - No hardcoded keys
 */

const STORAGE_KEY_PREFIX = '__secure__';

/**
 * Session-specific encryption key manager
 * Key is generated once per session and stored in memory
 */
class SecureKeyManager {
  private key: CryptoKey | null = null;
  private keyPromise: Promise<CryptoKey> | null = null;

  /**
   * Get or generate the session encryption key
   */
  async getKey(): Promise<CryptoKey> {
    // If key already exists, return it
    if (this.key) {
      return this.key;
    }

    // If key is being generated, wait for it
    if (this.keyPromise) {
      return this.keyPromise;
    }

    // Generate new key
    this.keyPromise = this.generateKey();
    this.key = await this.keyPromise;
    this.keyPromise = null;

    return this.key;
  }

  /**
   * Generate a new AES-GCM 256-bit encryption key
   */
  private async generateKey(): Promise<CryptoKey> {
    return await crypto.subtle.generateKey(
      {
        name: 'AES-GCM',
        length: 256, // 256-bit key
      },
      false, // Not extractable - key cannot be exported
      ['encrypt', 'decrypt']
    );
  }

  /**
   * Clear the key (useful for logout)
   */
  clearKey(): void {
    this.key = null;
    this.keyPromise = null;
  }
}

// Singleton instance
const keyManager = new SecureKeyManager();

/**
 * Encrypt text using AES-GCM with random IV
 */
async function encrypt(text: string): Promise<string> {
  try {
    const key = await keyManager.getKey();
    const encoder = new TextEncoder();
    const data = encoder.encode(text);

    // Generate random 12-byte IV (96 bits - recommended for AES-GCM)
    const iv = crypto.getRandomValues(new Uint8Array(12));

    // Encrypt data
    const encrypted = await crypto.subtle.encrypt(
      {
        name: 'AES-GCM',
        iv: iv,
      },
      key,
      data
    );

    // Combine IV + encrypted data
    const combined = new Uint8Array(iv.length + encrypted.byteLength);
    combined.set(iv, 0);
    combined.set(new Uint8Array(encrypted), iv.length);

    // Convert to base64 for storage
    return arrayBufferToBase64(combined);
  } catch (error) {
    console.error('[SecureStorage] Encryption failed:', error);
    throw new Error('Failed to encrypt data');
  }
}

/**
 * Decrypt text using AES-GCM
 */
async function decrypt(encryptedBase64: string): Promise<string> {
  try {
    const key = await keyManager.getKey();

    // Decode from base64
    const combined = base64ToArrayBuffer(encryptedBase64);

    // Extract IV (first 12 bytes) and ciphertext (rest)
    const iv = combined.slice(0, 12);
    const ciphertext = combined.slice(12);

    // Decrypt data
    const decrypted = await crypto.subtle.decrypt(
      {
        name: 'AES-GCM',
        iv: iv,
      },
      key,
      ciphertext
    );

    // Convert back to string
    const decoder = new TextDecoder();
    return decoder.decode(decrypted);
  } catch (error) {
    console.error('[SecureStorage] Decryption failed:', error);
    throw new Error('Failed to decrypt data');
  }
}

/**
 * Convert ArrayBuffer to Base64 string
 */
function arrayBufferToBase64(buffer: Uint8Array): string {
  let binary = '';
  const bytes = new Uint8Array(buffer);
  const len = bytes.byteLength;
  for (let i = 0; i < len; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

/**
 * Convert Base64 string to Uint8Array
 */
function base64ToArrayBuffer(base64: string): Uint8Array {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

/**
 * Secure Storage API with AES-GCM encryption
 * All methods are async due to Web Crypto API
 */
export const secureStorage = {
  /**
   * Store encrypted item in sessionStorage
   * @param key - Storage key
   * @param value - Value to encrypt and store
   */
  async setItem(key: string, value: string): Promise<void> {
    try {
      const encrypted = await encrypt(value);
      sessionStorage.setItem(`${STORAGE_KEY_PREFIX}${key}`, encrypted);
    } catch (error) {
      console.error('[SecureStorage] Failed to set item:', error);
      throw error;
    }
  },

  /**
   * Retrieve and decrypt item from sessionStorage
   * @param key - Storage key
   * @returns Decrypted value or null if not found/invalid
   */
  async getItem(key: string): Promise<string | null> {
    try {
      const encrypted = sessionStorage.getItem(`${STORAGE_KEY_PREFIX}${key}`);
      if (!encrypted) return null;

      return await decrypt(encrypted);
    } catch (error) {
      console.error('[SecureStorage] Failed to get item:', error);
      // Return null on decryption failure (e.g., corrupted data)
      return null;
    }
  },

  /**
   * Remove item from sessionStorage
   * @param key - Storage key
   */
  removeItem(key: string): void {
    try {
      sessionStorage.removeItem(`${STORAGE_KEY_PREFIX}${key}`);
    } catch (error) {
      console.error('[SecureStorage] Failed to remove item:', error);
    }
  },

  /**
   * Clear all secure storage items and encryption key
   */
  clear(): void {
    try {
      const keys = Object.keys(sessionStorage);
      keys.forEach((key) => {
        if (key.startsWith(STORAGE_KEY_PREFIX)) {
          sessionStorage.removeItem(key);
        }
      });

      // Clear the encryption key from memory
      keyManager.clearKey();
    } catch (error) {
      console.error('[SecureStorage] Failed to clear storage:', error);
    }
  },

  /**
   * Check if a key exists in storage
   * @param key - Storage key
   * @returns true if key exists
   */
  hasItem(key: string): boolean {
    return sessionStorage.getItem(`${STORAGE_KEY_PREFIX}${key}`) !== null;
  },
};

/**
 * Token validation utilities
 */

/**
 * Check if a JWT token is expired
 * @param token - JWT token string
 * @returns true if expired or invalid
 */
export function isTokenExpired(token: string): boolean {
  try {
    // JWT tokens have 3 parts: header.payload.signature
    const parts = token.split('.');
    if (parts.length !== 3) return true;

    // Decode payload (second part)
    const payload = JSON.parse(atob(parts[1]));

    // Check expiration (exp is in seconds, Date.now() is in milliseconds)
    if (payload.exp) {
      return payload.exp * 1000 < Date.now();
    }

    // If no exp field, consider it valid
    return false;
  } catch {
    // If parsing fails, consider token invalid
    return true;
  }
}

/**
 * Extract user info from JWT token
 * @param token - JWT token string
 * @returns Decoded payload or null if invalid
 */
export function decodeToken(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    return JSON.parse(atob(parts[1]));
  } catch {
    return null;
  }
}

/**
 * Validate token format (basic check)
 * @param token - Token string
 * @returns true if token appears valid
 */
export function validateTokenFormat(token: string): boolean {
  if (!token || typeof token !== 'string') return false;
  if (token.length < 10) return false;

  // For JWT, check if it has 3 parts
  const parts = token.split('.');
  return parts.length === 3;
}
