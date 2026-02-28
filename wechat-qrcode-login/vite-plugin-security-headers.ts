/**
 * Vite plugin to add security headers
 * Applies security headers to all responses in dev and preview mode
 */
import type { Plugin } from 'vite';

export function securityHeadersPlugin(): Plugin {
  // 从 Node.js 环境变量读取 WebSocket CSP URL
  const wsCspUrl = process.env.VITE_WS_CSP_URL;

  return {
    name: 'security-headers',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        // Content Security Policy
        res.setHeader(
          'Content-Security-Policy',
          [
            "default-src 'self'",
            "script-src 'self' 'unsafe-inline' 'unsafe-eval'", // unsafe-eval needed for dev HMR
            "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
            "font-src 'self' https://fonts.gstatic.com",
            "img-src 'self' data: https:",
            `connect-src 'self' ${wsCspUrl}`,
            "frame-ancestors 'none'",
            "base-uri 'self'",
            "form-action 'self'",
          ].join('; ')
        );

        // Additional security headers
        res.setHeader('X-Content-Type-Options', 'nosniff');
        res.setHeader('X-Frame-Options', 'DENY');
        res.setHeader('X-XSS-Protection', '1; mode=block');
        res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
        res.setHeader('Permissions-Policy', 'geolocation=(), microphone=(), camera=()');

        // Strict Transport Security (only for HTTPS)
        if (req.socket && 'encrypted' in req.socket) {
          res.setHeader('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');
        }

        next();
      });
    },
    configurePreviewServer(server) {
      server.middlewares.use((req, res, next) => {
        // Same headers for preview server
        res.setHeader(
          'Content-Security-Policy',
          [
            "default-src 'self'",
            "script-src 'self'", // No unsafe-inline in production
            "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
            "font-src 'self' https://fonts.gstatic.com",
            "img-src 'self' data: https:",
            `connect-src 'self' ${wsCspUrl}`,
            "frame-ancestors 'none'",
            "base-uri 'self'",
            "form-action 'self'",
          ].join('; ')
        );

        res.setHeader('X-Content-Type-Options', 'nosniff');
        res.setHeader('X-Frame-Options', 'DENY');
        res.setHeader('X-XSS-Protection', '1; mode=block');
        res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
        res.setHeader('Permissions-Policy', 'geolocation=(), microphone=(), camera=()');

        if (req.socket && 'encrypted' in req.socket) {
          res.setHeader('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');
        }

        next();
      });
    },
  };
}
