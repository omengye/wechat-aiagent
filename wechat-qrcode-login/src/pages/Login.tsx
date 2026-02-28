import { useLocation } from "wouter";
import { useEffect, useRef, useState } from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Loader2,
  RefreshCcw,
  ScanLine,
  Smartphone,
  CheckCircle2,
  AlertCircle,
  Clock,
} from "lucide-react";
import bgImageLight from "@/assets/login-bg.jpg";
import bgImageDark from "@/assets/login-bg-dark.jpg";
import { useWeChatLogin } from "@/hooks/useWeChatLogin";
import { toast } from "sonner";
import type { LoginTokens } from "@/types/auth";
import { secureStorage } from "@/utils/secureStorage";
import { TIMING } from "@/constants/timing";
import { useTheme } from "@/contexts/ThemeContext";
import { ENV_CONFIG } from "@/config/env";

export default function Login() {
  const [, setLocation] = useLocation();
  const redirectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { resolvedTheme } = useTheme();

  // QR code image state
  const [qrImageLoaded, setQrImageLoaded] = useState(false);
  const [qrImageError, setQrImageError] = useState(false);

  // Countdown timer state
  const [timeRemaining, setTimeRemaining] = useState(TIMING.QR_CODE_TIMEOUT / 1000); // Convert to seconds

  // Cleanup redirect timer on unmount
  useEffect(() => {
    return () => {
      if (redirectTimerRef.current) {
        clearTimeout(redirectTimerRef.current);
      }
    };
  }, []);

  // Callback when login succeeds
  const handleLoginSuccess = async (tokens: LoginTokens) => {
    try {
      // Use secure storage with AES-GCM encryption
      await secureStorage.setItem("access_token", tokens.access_token);
      await secureStorage.setItem("refresh_token", tokens.refresh_token);

      toast.success("登录成功");

      // Slight delay for visual feedback before redirect
      redirectTimerRef.current = setTimeout(() => {
        setLocation("/home");
      }, TIMING.REDIRECT_DELAY);
    } catch (error) {
      console.error("[Login] Failed to save tokens:", error);
      toast.error("保存登录信息失败，请重试");
    }
  };

  // Use the hook with login project ID
  const { status, qrCodeUrl, errorMsg, refresh } =
    useWeChatLogin(ENV_CONFIG.wechat.loginProjectId, handleLoginSuccess);

  // Reset QR image state when QR code changes
  useEffect(() => {
    if (qrCodeUrl) {
      setQrImageLoaded(false);
      setQrImageError(false);
      setTimeRemaining(TIMING.QR_CODE_TIMEOUT / 1000);
    }
  }, [qrCodeUrl]);

  // Countdown timer effect
  useEffect(() => {
    if (status === "waiting_for_scan" || status === "scanned") {
      const timer = setInterval(() => {
        setTimeRemaining((prev) => {
          if (prev <= 1) {
            clearInterval(timer);
            return 0;
          }
          return prev - 1;
        });
      }, TIMING.COUNTDOWN_UPDATE_INTERVAL);

      return () => clearInterval(timer);
    }
  }, [status]);

  // Format time remaining as MM:SS
  const formatTime = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  // Get background image based on theme
  const bgImage = resolvedTheme === "dark" ? bgImageDark : bgImageLight;

  return (
    <div className="min-h-screen w-full flex items-center justify-center bg-gray-50 dark:bg-background relative overflow-hidden transition-colors duration-500">
      {/* Optimized Background Image - Only load one based on theme */}
      <div
        className="absolute inset-0 z-0 opacity-30 dark:opacity-60 transition-opacity duration-500"
        style={{
          backgroundImage: `url(${bgImage})`,
          backgroundSize: "cover",
          backgroundPosition: "center",
        }}
      />

      <div className="z-10 w-full max-w-md px-4 animate-in fade-in zoom-in duration-500">
        <Card className="w-full shadow-2xl border-none bg-white/90 dark:bg-zinc-900/80 backdrop-blur-md transition-colors duration-300">
          <CardHeader className="text-center pb-2">
            <CardTitle className="text-2xl font-normal text-gray-800 dark:text-gray-100">
              微信登录
            </CardTitle>
            <CardDescription className="text-gray-500 dark:text-gray-400">
              请使用微信扫一扫，关注公众号后自动登录
            </CardDescription>
          </CardHeader>

          <CardContent className="flex flex-col items-center pt-6 pb-10 min-h-[360px] justify-center">
            {/* Status: Connecting / Loading */}
            {(status === "connecting" ||
              (status === "loading_qr" && !qrCodeUrl)) && (
              <div className="flex flex-col items-center justify-center h-[200px] w-[200px] bg-gray-50 dark:bg-zinc-800/50 rounded-lg">
                <Loader2 className="h-10 w-10 text-primary animate-spin mb-4" />
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {status === "connecting"
                    ? "正在连接..."
                    : "正在获取二维码..."}
                </p>
              </div>
            )}

            {/* Status: QR Display (Waiting for scan) */}
            {status === "waiting_for_scan" && qrCodeUrl && (
              <div className="relative group animate-in fade-in duration-500">
                <div className="relative h-[200px] w-[200px] bg-white p-2 border border-gray-100 dark:border-zinc-700 rounded-lg shadow-sm">
                  {/* Loading spinner while image loads */}
                  {!qrImageLoaded && !qrImageError && (
                    <div className="absolute inset-0 flex items-center justify-center">
                      <Loader2 className="h-8 w-8 text-primary animate-spin" />
                    </div>
                  )}

                  {/* Error state */}
                  {qrImageError && (
                    <div className="absolute inset-0 flex flex-col items-center justify-center text-gray-400">
                      <AlertCircle className="h-8 w-8 mb-2" />
                      <span className="text-xs">图片加载失败</span>
                    </div>
                  )}

                  {/* QR Code Image */}
                  <img
                    src={qrCodeUrl}
                    alt="Login QR Code"
                    className={`w-full h-full object-contain transition-opacity duration-300 ${
                      qrImageLoaded ? "opacity-100" : "opacity-0"
                    }`}
                    loading="lazy"
                    onLoad={() => setQrImageLoaded(true)}
                    onError={() => setQrImageError(true)}
                  />

                  {/* Hover overlay hint */}
                  <div className="absolute inset-0 bg-black/5 dark:bg-white/5 opacity-0 group-hover:opacity-100 transition-opacity rounded-lg flex items-center justify-center pointer-events-none"></div>
                </div>

                {/* Scan prompt with countdown */}
                <div className="mt-6 space-y-2">
                  <div className="flex items-center justify-center space-x-2 text-gray-500 dark:text-gray-400 text-sm bg-gray-100 dark:bg-zinc-800 py-2 px-4 rounded-full">
                    <ScanLine className="h-4 w-4" />
                    <span>打开微信扫一扫</span>
                  </div>

                  {/* Countdown timer */}
                  <div className="flex items-center justify-center space-x-1 text-xs text-gray-400 dark:text-gray-500">
                    <Clock className="h-3 w-3" />
                    <span>{formatTime(timeRemaining)}</span>
                  </div>
                </div>
              </div>
            )}

            {/* Status: Scanned */}
            {status === "scanned" && (
              <div className="flex flex-col items-center justify-center h-[200px] w-[200px] animate-in zoom-in duration-300">
                <div className="h-20 w-20 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mb-6 animate-pulse">
                  <Smartphone className="h-10 w-10 text-primary" />
                </div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-1">
                  扫描成功
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 text-center">
                  请在手机上点击确认登录
                </p>
              </div>
            )}

            {/* Status: Success */}
            {status === "success" && (
              <div className="flex flex-col items-center justify-center h-[200px] w-[200px] animate-in zoom-in duration-300">
                <div className="h-20 w-20 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mb-6">
                  <CheckCircle2 className="h-10 w-10 text-primary" />
                </div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-1">
                  登录成功
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 text-center">
                  正在跳转...
                </p>
              </div>
            )}

            {/* Status: Expired or Error */}
            {(status === "expired" || status === "error") && (
              <div className="flex flex-col items-center justify-center h-[200px] w-[200px] relative">
                {/* Blurred QR background for effect */}
                <div className="absolute inset-0 opacity-10 blur-sm pointer-events-none">
                  {qrCodeUrl && (
                    <img src={qrCodeUrl} className="w-full h-full" />
                  )}
                </div>

                <div className="z-10 flex flex-col items-center bg-white/90 dark:bg-zinc-800/90 p-6 rounded-xl shadow-sm backdrop-blur-sm animate-in fade-in slide-in-from-bottom-4">
                  <AlertCircle className="h-10 w-10 text-orange-500 mb-2" />
                  <p className="text-sm font-medium text-gray-800 dark:text-gray-200 mb-4 text-center">
                    {status === "expired"
                      ? "二维码已过期"
                      : errorMsg || "连接失败"}
                  </p>
                  <Button onClick={refresh} variant="default" className="gap-2">
                    <RefreshCcw className="h-4 w-4" />
                    刷新二维码
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
