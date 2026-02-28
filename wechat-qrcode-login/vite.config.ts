import path from "path";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { securityHeadersPlugin } from "./vite-plugin-security-headers";

export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: [
          // Inject data-source attribute for AI agent source location
          "./scripts/babel-plugin-jsx-source-location.cjs",
        ],
      },
    }),
    tailwindcss(),
    securityHeadersPlugin(), // Add security headers
  ],
  resolve: { alias: { "@": path.resolve(__dirname, "./src") } },
  base: "./",
  build: { outDir: "dist", emptyOutDir: true },

  // 开发服务器配置
  server: {
    port: 5173,
    host: 'localhost',

    // 代理配置 - 解决跨域问题
    proxy: {
      // 代理所有 /api 开头的 HTTP 请求到后端
      '/api': {
        target: 'http://127.0.0.1:8443',
        changeOrigin: false,  // ✅ 保持原始 Origin header
        secure: false,
        // WebSocket 支持
        ws: true,  // ✅ 启用 WebSocket 代理
        // 手动设置 Origin header
        configure: (proxy, _options) => {
          proxy.on('proxyReq', (proxyReq, _req, _res) => {
            // 强制设置 Origin 为 https://agent.sprwhisp.cc
            proxyReq.setHeader('Origin', 'http://127.0.0.1:8443');
            console.log('[Vite Proxy] Setting Origin:', 'http://127.0.0.1:8443');
          });
        },
      },
    },
  },
});
