import { Button } from "@/components/ui/button";
import { Link } from "wouter";
import { ArrowRight, Lock, Smartphone, Zap, Shield, QrCode } from "lucide-react";
import { useTheme } from "@/contexts/ThemeContext";

export default function Landing() {
  const { resolvedTheme } = useTheme();

  return (
    <div className="min-h-screen w-full bg-gradient-to-br from-blue-50 via-white to-indigo-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 transition-colors duration-500">
      {/* Hero Section */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="pt-20 pb-16 text-center lg:pt-32">
          {/* Logo/Title */}
          <div className="flex justify-center mb-8">
            <div className="h-16 w-16 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-2xl flex items-center justify-center shadow-lg">
              <QrCode className="h-9 w-9 text-white" />
            </div>
          </div>

          <h1 className="text-5xl md:text-6xl font-bold text-gray-900 dark:text-white mb-6 animate-in fade-in slide-in-from-bottom-4 duration-1000">
            微信扫码登录
          </h1>

          <p className="text-xl md:text-2xl text-gray-600 dark:text-gray-300 mb-12 max-w-3xl mx-auto animate-in fade-in slide-in-from-bottom-5 duration-1000 delay-100">
            简单、安全、快速的微信授权登录方案
          </p>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center animate-in fade-in slide-in-from-bottom-6 duration-1000 delay-200">
            <Link href="/login">
              <Button size="lg" className="gap-2 text-lg px-8 py-6 shadow-lg hover:shadow-xl transition-all">
                开始登录
                <ArrowRight className="h-5 w-5" />
              </Button>
            </Link>

            <Link href="/oauth/authorize">
              <Button size="lg" variant="outline" className="gap-2 text-lg px-8 py-6">
                OAuth 授权
                <Lock className="h-5 w-5" />
              </Button>
            </Link>
          </div>
        </div>

        {/* Features Section */}
        <div className="py-16 grid grid-cols-1 md:grid-cols-3 gap-8 max-w-5xl mx-auto">
          {/* Feature 1 */}
          <div className="bg-white dark:bg-gray-800 p-8 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 animate-in fade-in slide-in-from-bottom-4 duration-1000 delay-300">
            <div className="h-12 w-12 bg-blue-100 dark:bg-blue-900/30 rounded-xl flex items-center justify-center mb-4">
              <Smartphone className="h-6 w-6 text-blue-600 dark:text-blue-400" />
            </div>
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">
              便捷扫码
            </h3>
            <p className="text-gray-600 dark:text-gray-400">
              使用微信扫一扫，无需输入账号密码，一键完成登录授权
            </p>
          </div>

          {/* Feature 2 */}
          <div className="bg-white dark:bg-gray-800 p-8 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 animate-in fade-in slide-in-from-bottom-4 duration-1000 delay-400">
            <div className="h-12 w-12 bg-green-100 dark:bg-green-900/30 rounded-xl flex items-center justify-center mb-4">
              <Shield className="h-6 w-6 text-green-600 dark:text-green-400" />
            </div>
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">
              安全可靠
            </h3>
            <p className="text-gray-600 dark:text-gray-400">
              基于 OAuth 2.0 标准协议，保护用户隐私和数据安全
            </p>
          </div>

          {/* Feature 3 */}
          <div className="bg-white dark:bg-gray-800 p-8 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 animate-in fade-in slide-in-from-bottom-4 duration-1000 delay-500">
            <div className="h-12 w-12 bg-purple-100 dark:bg-purple-900/30 rounded-xl flex items-center justify-center mb-4">
              <Zap className="h-6 w-6 text-purple-600 dark:text-purple-400" />
            </div>
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">
              快速接入
            </h3>
            <p className="text-gray-600 dark:text-gray-400">
              简单配置即可集成到您的应用，支持多种场景和平台
            </p>
          </div>
        </div>

        {/* Tech Stack / Info Section */}
        <div className="py-16 text-center">
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            技术栈
          </p>
          <div className="flex flex-wrap justify-center gap-4 text-sm text-gray-600 dark:text-gray-400">
            <span className="px-4 py-2 bg-white dark:bg-gray-800 rounded-full shadow">React 19</span>
            <span className="px-4 py-2 bg-white dark:bg-gray-800 rounded-full shadow">TypeScript</span>
            <span className="px-4 py-2 bg-white dark:bg-gray-800 rounded-full shadow">Tailwind CSS</span>
            <span className="px-4 py-2 bg-white dark:bg-gray-800 rounded-full shadow">OAuth 2.0</span>
            <span className="px-4 py-2 bg-white dark:bg-gray-800 rounded-full shadow">WebSocket</span>
          </div>
        </div>

        {/* Footer */}
        <div className="py-8 border-t border-gray-200 dark:border-gray-700 text-center">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            © 2026 微信扫码登录系统
          </p>
        </div>
      </div>
    </div>
  );
}
