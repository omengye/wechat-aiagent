/**
 * OAuth Authorization Page
 * Displays OAuth authorization consent screen with QR code
 */

import { useEffect, useState } from 'react';
import { useLocation } from 'wouter';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Loader2,
  RefreshCcw,
  ScanLine,
  CheckCircle2,
  AlertCircle,
  Shield,
} from 'lucide-react';
import { requestOAuthAuthorization, generateState, getDefaultOAuthParams } from '@/services/oauth';
import { OAuthWebSocketService } from '@/services/oauth-websocket';
import { validateEnvConfig } from '@/config/env';
import type { LoginStatus } from '@/types/auth';
import type { OAuthAuthorizeResponse } from '@/types/oauth';
import { toast } from 'sonner';

export default function OAuthAuthorize() {
  const [, setLocation] = useLocation();
  const [status, setStatus] = useState<LoginStatus>('connecting');
  const [qrCodeUrl, setQrCodeUrl] = useState<string>('');
  const [errorMsg, setErrorMsg] = useState<string>('');
  const [authData, setAuthData] = useState<OAuthAuthorizeResponse | null>(null);
  const [qrImageLoaded, setQrImageLoaded] = useState(false);

  useEffect(() => {
    let wsService: OAuthWebSocketService | null = null;

    async function initOAuthFlow() {
      try {
        // 验证环境配置
        validateEnvConfig();

        setStatus('connecting');

        // 1. 获取默认 OAuth 参数
        const defaultParams = getDefaultOAuthParams();

        // 2. 调用 /api/oauth/authorize 获取 session_id
        const response = await requestOAuthAuthorization({
          client_id: defaultParams.client_id!,
          redirect_url: defaultParams.redirect_url!,
          response_type: defaultParams.response_type!,
          scope: defaultParams.scope!,
          state: generateState(),
        });

        console.log('[OAuth] Authorization response:', {
          session_id: response.session_id,
          client_name: response.client_name,
          project_id: response.project_id,
          scopes: response.scope_names,
        });

        setAuthData(response);

        // 3. 创建 WebSocket 服务
        wsService = new OAuthWebSocketService({
          onStatusChange: (newStatus) => {
            console.log('[OAuth] Status changed:', newStatus);
            setStatus(newStatus);
          },
          onQrCodeReceived: (url) => {
            console.log('[OAuth] QR Code URL received:', url);
            setQrCodeUrl(url);
            setQrImageLoaded(false);
          },
          onRedirectReceived: (redirectUrl) => {
            console.log('[OAuth] Redirect URL received:', redirectUrl);
            toast.success('授权成功，正在跳转...');

            // 延迟跳转，让用户看到成功提示
            setTimeout(() => {
              // 解析 redirect URL，提取 hash 路由部分
              try {
                const url = new URL(redirectUrl);
                console.log('[OAuth] Parsed URL:', {
                  origin: url.origin,
                  hash: url.hash,
                  search: url.search,
                  currentOrigin: window.location.origin,
                });

                // 检查是否是本地 URL 并且包含 hash
                if (url.origin === window.location.origin && url.hash) {
                  // 使用 hash 路由，不刷新页面
                  console.log('[OAuth] Using hash navigation:', url.hash);
                  window.location.hash = url.hash;
                } else {
                  // 外部 URL 或没有 hash，使用完整跳转
                  console.log('[OAuth] Using full navigation:', redirectUrl);
                  window.location.href = redirectUrl;
                }
              } catch (error) {
                // URL 解析失败，尝试直接跳转
                console.error('[OAuth] Failed to parse redirect URL:', error);
                window.location.href = redirectUrl;
              }
            }, 1000);
          },
          onError: (message) => {
            console.error('[OAuth] Error:', message);
            setErrorMsg(message);
            toast.error(message);
          },
        });

        // 4. 连接 WebSocket 并发送带有 oauth_session_id 的消息
        wsService.connect(response.project_id, response.session_id);

      } catch (error) {
        console.error('[OAuth] Initialization failed:', error);
        const message = error instanceof Error ? error.message : '初始化失败';
        setErrorMsg(message);
        setStatus('error');
        toast.error(message);
      }
    }

    initOAuthFlow();

    // 清理函数
    return () => {
      if (wsService) {
        wsService.disconnect();
      }
    };
  }, []);

  // 刷新/重试
  const handleRefresh = () => {
    window.location.reload();
  };

  // 取消授权
  const handleCancel = () => {
    toast.info('已取消授权');
    // 可以跳转到其他页面或关闭窗口
    setLocation('/');
  };

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <div className="w-full max-w-md px-4">
        <Card className="w-full shadow-2xl border-none bg-white/95 dark:bg-zinc-900/95 backdrop-blur-md">
          <CardHeader className="text-center pb-4">
            <div className="flex items-center justify-center mb-2">
              <Shield className="h-8 w-8 text-blue-500 mr-2" />
              <CardTitle className="text-2xl font-bold">OAuth 授权</CardTitle>
            </div>
            {authData && (
              <CardDescription className="text-base">
                <span className="font-semibold text-blue-600 dark:text-blue-400">
                  {authData.client_name}
                </span>
                {' '}请求访问你的账号
              </CardDescription>
            )}
          </CardHeader>

          <CardContent className="space-y-6 pb-8">
            {/* 权限列表 */}
            {authData && status !== 'error' && (
              <div className="bg-gray-50 dark:bg-zinc-800/50 p-4 rounded-lg">
                <h3 className="text-sm font-semibold mb-3 text-gray-700 dark:text-gray-300">
                  该应用将获得以下权限：
                </h3>
                <ul className="space-y-2">
                  {authData.scope_names.map((scope, idx) => (
                    <li key={idx} className="flex items-start text-sm text-gray-600 dark:text-gray-400">
                      <span className="text-green-500 mr-2 mt-0.5">✓</span>
                      <span>{scope}</span>
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {/* 加载状态 */}
            {(status === 'connecting' || status === 'loading_qr') && (
              <div className="flex flex-col items-center justify-center py-12">
                <Loader2 className="h-12 w-12 text-blue-500 animate-spin mb-4" />
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {status === 'connecting' ? '正在连接服务器...' : '正在获取二维码...'}
                </p>
              </div>
            )}

            {/* 二维码展示 */}
            {status === 'waiting_for_scan' && qrCodeUrl && (
              <div className="space-y-4">
                <div className="relative bg-white p-4 border-2 border-gray-200 dark:border-zinc-700 rounded-xl shadow-sm">
                  {/* 加载中 */}
                  {!qrImageLoaded && (
                    <div className="absolute inset-0 flex items-center justify-center">
                      <Loader2 className="h-8 w-8 text-blue-500 animate-spin" />
                    </div>
                  )}

                  {/* 二维码图片 */}
                  <img
                    src={qrCodeUrl}
                    alt="OAuth 授权二维码"
                    className={`w-full h-auto transition-opacity duration-300 ${
                      qrImageLoaded ? 'opacity-100' : 'opacity-0'
                    }`}
                    onLoad={() => setQrImageLoaded(true)}
                    onError={() => {
                      setQrImageLoaded(true);
                      toast.error('二维码加载失败');
                    }}
                  />
                </div>

                {/* 扫码提示 */}
                <div className="flex items-center justify-center space-x-2 text-gray-600 dark:text-gray-400 bg-blue-50 dark:bg-blue-900/20 py-3 px-4 rounded-lg">
                  <ScanLine className="h-5 w-5" />
                  <span className="text-sm font-medium">请使用微信扫描二维码</span>
                </div>

                {/* 授权说明 */}
                <p className="text-xs text-center text-gray-400 dark:text-gray-500">
                  扫码成功即表示同意授权。授权后可在账号设置中撤销。
                </p>
              </div>
            )}

            {/* 成功状态 */}
            {status === 'success' && (
              <div className="flex flex-col items-center justify-center py-12">
                <div className="h-16 w-16 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mb-4">
                  <CheckCircle2 className="h-10 w-10 text-green-600 dark:text-green-400" />
                </div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
                  授权成功！
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  正在跳转到应用...
                </p>
              </div>
            )}

            {/* 错误状态 */}
            {status === 'error' && (
              <div className="flex flex-col items-center justify-center py-12">
                <div className="h-16 w-16 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mb-4">
                  <AlertCircle className="h-10 w-10 text-red-600 dark:text-red-400" />
                </div>
                <h3 className="text-lg font-semibold text-red-600 dark:text-red-400 mb-2">
                  {errorMsg || '连接失败'}
                </h3>
                <div className="flex space-x-3 mt-4">
                  <Button onClick={handleRefresh} variant="default" className="gap-2">
                    <RefreshCcw className="h-4 w-4" />
                    重试
                  </Button>
                  <Button onClick={handleCancel} variant="outline">
                    取消
                  </Button>
                </div>
              </div>
            )}

            {/* 二维码过期 */}
            {status === 'expired' && (
              <div className="flex flex-col items-center justify-center py-12">
                <AlertCircle className="h-12 w-12 text-orange-500 mb-4" />
                <p className="text-sm font-medium text-gray-800 dark:text-gray-200 mb-4">
                  二维码已过期
                </p>
                <Button onClick={handleRefresh} variant="default" className="gap-2">
                  <RefreshCcw className="h-4 w-4" />
                  刷新二维码
                </Button>
              </div>
            )}

            {/* 取消按钮（在等待扫码时显示） */}
            {status === 'waiting_for_scan' && (
              <div className="flex justify-center">
                <Button onClick={handleCancel} variant="ghost" size="sm">
                  取消授权
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        {/* 调试信息（仅开发环境） */}
        {import.meta.env.DEV && authData && (
          <div className="mt-4 p-4 bg-gray-800 text-gray-100 rounded-lg text-xs font-mono">
            <div>Session ID: {authData.session_id}</div>
            <div>Project ID: {authData.project_id}</div>
            <div>Redirect URL: {authData.redirect_url}</div>
          </div>
        )}
      </div>
    </div>
  );
}
