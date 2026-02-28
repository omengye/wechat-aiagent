/**
 * OAuth Callback Page
 * Handles authorization code and exchanges it for access token
 */

import { useEffect, useState, useRef } from 'react';
import { useLocation } from 'wouter';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
  Loader2,
  CheckCircle2,
  AlertCircle,
  Copy,
  Check,
  Eye,
  EyeOff,
} from 'lucide-react';
import { exchangeCodeForToken } from '@/services/oauth';
import { toast } from 'sonner';

export default function OAuthCallback() {
  const [, setLocation] = useLocation();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [tokens, setTokens] = useState<{
    access_token: string;
    refresh_token: string;
    expires_in: number;
  } | null>(null);

  // Token 可见性控制
  const [showAccessToken, setShowAccessToken] = useState(false);
  const [showRefreshToken, setShowRefreshToken] = useState(false);

  // 复制状态
  const [copiedAccess, setCopiedAccess] = useState(false);
  const [copiedRefresh, setCopiedRefresh] = useState(false);

  // 防止重复请求（React StrictMode 会导致 useEffect 执行两次）
  const hasRequestedRef = useRef(false);

  useEffect(() => {
    // 如果已经请求过，直接返回（防止 React StrictMode 双重渲染导致的重复请求）
    if (hasRequestedRef.current) {
      console.log('[OAuth Callback] ⚠️ Request already sent, skipping duplicate (React StrictMode)');
      return;
    }

    async function handleCallback() {
      try {
        // 标记为已请求（必须在异步函数开始前设置，防止并发问题）
        hasRequestedRef.current = true;
        console.log('[OAuth Callback] 🚀 Starting token exchange...');

        // 从 URL 获取参数（支持 hash 路由）
        // Hash 路由格式: http://localhost:5173/#/oauth/callback?code=xxx&state=yyy
        const hash = window.location.hash;
        const queryString = hash.includes('?') ? hash.split('?')[1] : '';
        const urlParams = new URLSearchParams(queryString);
        const code = urlParams.get('code');
        const state = urlParams.get('state');
        const errorParam = urlParams.get('error');
        const errorDescription = urlParams.get('error_description');

        console.log('[OAuth Callback] URL params:', {
          hash,
          queryString,
          code: code ? `${code.substring(0, 10)}...` : null,
          state,
          error: errorParam,
        });

        // 检查是否有错误
        if (errorParam) {
          throw new Error(errorDescription || `OAuth Error: ${errorParam}`);
        }

        // 检查是否有授权码
        if (!code) {
          throw new Error('缺少授权码（code）参数');
        }

        // 交换授权码获取 token
        const tokenData = await exchangeCodeForToken(code, state || undefined);

        console.log('[OAuth Callback] Tokens received successfully');

        setTokens(tokenData);
        toast.success('授权成功！已获取访问令牌');
      } catch (err) {
        console.error('[OAuth Callback] Error:', err);
        const errorMessage = err instanceof Error ? err.message : '未知错误';
        setError(errorMessage);
        toast.error(`授权失败: ${errorMessage}`);
      } finally {
        setLoading(false);
      }
    }

    handleCallback();
  }, []);

  // 复制到剪贴板
  const copyToClipboard = async (text: string, type: 'access' | 'refresh') => {
    try {
      await navigator.clipboard.writeText(text);
      if (type === 'access') {
        setCopiedAccess(true);
        setTimeout(() => setCopiedAccess(false), 2000);
      } else {
        setCopiedRefresh(true);
        setTimeout(() => setCopiedRefresh(false), 2000);
      }
      toast.success('已复制到剪贴板');
    } catch (err) {
      toast.error('复制失败');
    }
  };

  // 格式化过期时间
  const formatExpiresIn = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) {
      return `${hours} 小时 ${minutes} 分钟`;
    }
    return `${minutes} 分钟`;
  };

  // 遮罩 token（只显示前后几位）
  const maskToken = (token: string) => {
    if (token.length <= 20) return token;
    return `${token.substring(0, 10)}...${token.substring(token.length - 10)}`;
  };

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800 p-4">
      <Card className="w-full max-w-2xl shadow-2xl border-none bg-white/95 dark:bg-zinc-900/95 backdrop-blur-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold">OAuth 授权回调</CardTitle>
        </CardHeader>

        <CardContent className="space-y-6">
          {/* 加载状态 */}
          {loading && (
            <div className="flex flex-col items-center justify-center py-12">
              <Loader2 className="h-12 w-12 text-blue-500 animate-spin mb-4" />
              <p className="text-gray-600 dark:text-gray-400">
                正在处理授权码...
              </p>
            </div>
          )}

          {/* 成功状态 */}
          {!loading && !error && tokens && (
            <div className="space-y-6">
              {/* 成功提示 */}
              <div className="flex flex-col items-center py-6">
                <div className="h-16 w-16 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mb-4">
                  <CheckCircle2 className="h-10 w-10 text-green-600 dark:text-green-400" />
                </div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
                  授权成功！
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  已成功获取访问令牌
                </p>
              </div>

              {/* Access Token */}
              <div className="bg-gray-50 dark:bg-zinc-800/50 p-4 rounded-lg space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                    Access Token
                  </h4>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setShowAccessToken(!showAccessToken)}
                      className="h-8 px-2"
                    >
                      {showAccessToken ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(tokens.access_token, 'access')}
                      className="h-8 px-2"
                    >
                      {copiedAccess ? (
                        <Check className="h-4 w-4 text-green-500" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
                <div className="bg-white dark:bg-zinc-900 p-3 rounded border border-gray-200 dark:border-zinc-700">
                  <code className="text-xs font-mono break-all text-gray-800 dark:text-gray-200">
                    {showAccessToken ? tokens.access_token : maskToken(tokens.access_token)}
                  </code>
                </div>
                <div className="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                  <span>有效期: {formatExpiresIn(tokens.expires_in)}</span>
                  <span className="text-green-600 dark:text-green-400">
                    ✓ 已验证
                  </span>
                </div>
              </div>

              {/* Refresh Token */}
              <div className="bg-gray-50 dark:bg-zinc-800/50 p-4 rounded-lg space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                    Refresh Token
                  </h4>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setShowRefreshToken(!showRefreshToken)}
                      className="h-8 px-2"
                    >
                      {showRefreshToken ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(tokens.refresh_token, 'refresh')}
                      className="h-8 px-2"
                    >
                      {copiedRefresh ? (
                        <Check className="h-4 w-4 text-green-500" />
                      ) : (
                        <Copy className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                </div>
                <div className="bg-white dark:bg-zinc-900 p-3 rounded border border-gray-200 dark:border-zinc-700">
                  <code className="text-xs font-mono break-all text-gray-800 dark:text-gray-200">
                    {showRefreshToken ? tokens.refresh_token : maskToken(tokens.refresh_token)}
                  </code>
                </div>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  <span>用于刷新过期的 Access Token</span>
                </div>
              </div>

              {/* 提示信息 */}
              <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg">
                <p className="text-sm text-blue-800 dark:text-blue-300">
                  ℹ️ 请妥善保管这些令牌，不要泄露给他人。Access Token 用于访问受保护的资源，Refresh Token 用于在 Access Token 过期后获取新的令牌。
                </p>
              </div>

              {/* 操作按钮 */}
              <div className="flex justify-center gap-3">
                <Button
                  onClick={() => setLocation('/')}
                  variant="outline"
                >
                  返回首页
                </Button>
              </div>
            </div>
          )}

          {/* 错误状态 */}
          {!loading && error && (
            <div className="space-y-6">
              <div className="flex flex-col items-center py-12">
                <div className="h-16 w-16 bg-red-100 dark:bg-red-900/30 rounded-full flex items-center justify-center mb-4">
                  <AlertCircle className="h-10 w-10 text-red-600 dark:text-red-400" />
                </div>
                <h3 className="text-lg font-semibold text-red-600 dark:text-red-400 mb-2">
                  授权失败
                </h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 text-center max-w-md">
                  {error}
                </p>
              </div>

              {/* 错误详情 */}
              <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-lg">
                <p className="text-sm text-red-800 dark:text-red-300">
                  {error}
                </p>
              </div>

              {/* 操作按钮 */}
              <div className="flex justify-center gap-3">
                <Button
                  onClick={() => window.location.reload()}
                  variant="default"
                >
                  重试
                </Button>
                <Button
                  onClick={() => setLocation('/')}
                  variant="outline"
                >
                  返回首页
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
