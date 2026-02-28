import { useEffect, useState } from "react";
import { useLocation } from "wouter";
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ShieldCheck, LogOut, User, Key, AlertTriangle, Eye, EyeOff, Copy } from "lucide-react";
import { toast } from "sonner";
import { secureStorage, isTokenExpired } from "@/utils/secureStorage";
import { clearCSRFToken } from "@/utils/csrf";

/**
 * Mask a token by showing only first 6 and last 4 characters
 */
function maskToken(token: string): string {
  if (!token || token.length < 10) return '***';
  return `${token.slice(0, 6)}...${token.slice(-4)}`;
}

/**
 * Copy text to clipboard
 */
async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch (error) {
    console.error('Failed to copy:', error);
    return false;
  }
}

export default function Home() {
  const [, setLocation] = useLocation();
  const [tokens, setTokens] = useState<{access: string, refresh: string, isExpired: boolean} | null>(null);
  const [loading, setLoading] = useState(true);
  const [showAccessToken, setShowAccessToken] = useState(false);
  const [showRefreshToken, setShowRefreshToken] = useState(false);

  useEffect(() => {
    // Check for tokens using secure storage (async)
    const checkAuth = async () => {
      try {
        const access = await secureStorage.getItem('access_token');
        const refresh = await secureStorage.getItem('refresh_token');

        if (!access || !refresh) {
          toast.error("未登录或会话已过期");
          setLocation('/login');
          return;
        }

        // Check if token is expired
        const expired = isTokenExpired(access);
        if (expired) {
          toast.warning("登录已过期，请重新登录");
        }

        setTokens({ access, refresh, isExpired: expired });
      } catch (error) {
        console.error('[Home] Failed to load tokens:', error);
        toast.error("加载失败，请重新登录");
        setLocation('/login');
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, [setLocation]);

  const handleLogout = () => {
    // Clear all security data
    secureStorage.clear(); // Clear all secure storage and encryption key
    clearCSRFToken(); // Clear CSRF token
    toast.success("已退出登录");
    setLocation('/login');
  };

  if (loading || !tokens) return null; // Or loading spinner

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-gray-50 dark:bg-background p-4 transition-colors duration-300">
      <div className="w-full max-w-2xl animate-in fade-in slide-in-from-bottom-8 duration-700">
        <Card className="shadow-2xl border-none bg-white/90 dark:bg-zinc-900/80 backdrop-blur-md transition-colors duration-300">
          <CardHeader className="flex flex-row items-center space-x-4 pb-2">
            <div className="h-12 w-12 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center">
              <User className="h-6 w-6 text-primary" />
            </div>
            <div>
              <CardTitle className="text-xl text-gray-900 dark:text-gray-100">欢迎回来</CardTitle>
              <CardDescription className="text-gray-500 dark:text-gray-400">您已成功通过微信扫码登录</CardDescription>
            </div>
          </CardHeader>
          
          <CardContent className="pt-6 space-y-6">
            {/* Security status */}
            <div className={`p-4 rounded-lg border ${
              tokens?.isExpired
                ? 'bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800'
                : 'bg-gray-50 dark:bg-zinc-800/50 border-gray-100 dark:border-zinc-700'
            }`}>
              <div className={`flex items-center space-x-2 mb-2 ${
                tokens?.isExpired ? 'text-orange-600 dark:text-orange-400' : 'text-primary'
              }`}>
                {tokens?.isExpired ? (
                  <AlertTriangle className="h-4 w-4" />
                ) : (
                  <ShieldCheck className="h-4 w-4" />
                )}
                <span className="text-sm font-medium">
                  {tokens?.isExpired ? 'Token 已过期' : '安全验证通过'}
                </span>
              </div>
              <p className="text-sm text-gray-500 dark:text-gray-400">
                {tokens?.isExpired
                  ? '您的登录凭证已过期，部分功能可能受限。建议重新登录。'
                  : '您的会话已建立，Token 已加密存储在 sessionStorage。'}
              </p>
            </div>

            <div className="space-y-4">
              {/* Access Token */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Access Token</label>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setShowAccessToken(!showAccessToken)}
                      className="h-6 px-2 text-xs"
                      title={showAccessToken ? "隐藏" : "显示"}
                    >
                      {showAccessToken ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={async () => {
                        const success = await copyToClipboard(tokens.access);
                        toast.success(success ? "已复制到剪贴板" : "复制失败");
                      }}
                      className="h-6 px-2 text-xs"
                      title="复制"
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
                <div className="flex items-center space-x-2 bg-slate-100 dark:bg-zinc-800 p-3 rounded-md font-mono text-xs text-slate-600 dark:text-slate-300 break-all transition-colors duration-300">
                  <Key className="h-3 w-3 shrink-0" />
                  <span className={showAccessToken ? '' : 'select-none'}>
                    {showAccessToken ? tokens.access : maskToken(tokens.access)}
                  </span>
                </div>
              </div>

              {/* Refresh Token */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Refresh Token</label>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setShowRefreshToken(!showRefreshToken)}
                      className="h-6 px-2 text-xs"
                      title={showRefreshToken ? "隐藏" : "显示"}
                    >
                      {showRefreshToken ? <EyeOff className="h-3 w-3" /> : <Eye className="h-3 w-3" />}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={async () => {
                        const success = await copyToClipboard(tokens.refresh);
                        toast.success(success ? "已复制到剪贴板" : "复制失败");
                      }}
                      className="h-6 px-2 text-xs"
                      title="复制"
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
                <div className="flex items-center space-x-2 bg-slate-100 dark:bg-zinc-800 p-3 rounded-md font-mono text-xs text-slate-600 dark:text-slate-300 break-all transition-colors duration-300">
                  <Key className="h-3 w-3 shrink-0" />
                  <span className={showRefreshToken ? '' : 'select-none'}>
                    {showRefreshToken ? tokens.refresh : maskToken(tokens.refresh)}
                  </span>
                </div>
              </div>
            </div>
          </CardContent>

          <CardFooter className="bg-gray-50/50 dark:bg-zinc-900/50 p-6 flex justify-end rounded-b-lg">
             <Button variant="destructive" onClick={handleLogout} className="gap-2">
               <LogOut className="h-4 w-4" />
               退出登录
             </Button>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
