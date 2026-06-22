import type { Metadata } from 'next'
import { Toaster } from 'react-hot-toast'
import { RouteTransitionProvider } from '@/components/layout/RouteTransitionProvider'
import { ThemeProvider } from '@/lib/theme'
// 引入本地 Font Awesome 样式
import '@fortawesome/fontawesome-free/css/all.min.css'
import './globals.css'

export const metadata: Metadata = {
  title: '卡密购买系统',
  description: '安全、便捷的卡密购买平台',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `
              (function() {
                try {
                  var theme = localStorage.getItem('app-theme') || 'dark';
                  document.documentElement.setAttribute('data-theme', theme);
                } catch (e) {}
              })();
            `,
          }}
        />
      </head>
      <body className="min-h-screen transition-colors duration-300">
        <ThemeProvider>
          <RouteTransitionProvider>{children}</RouteTransitionProvider>
          <Toaster
            position="top-center"
            toastOptions={{
              duration: 3000,
              style: {
                background: 'var(--toast-bg)',
                color: 'var(--toast-text)',
                border: '1px solid var(--toast-border)',
              },
              success: {
                iconTheme: { primary: '#10b981', secondary: 'var(--toast-text)' },
              },
              error: {
                iconTheme: { primary: '#ef4444', secondary: 'var(--toast-text)' },
              },
            }}
          />
        </ThemeProvider>
      </body>
    </html>
  )
}
