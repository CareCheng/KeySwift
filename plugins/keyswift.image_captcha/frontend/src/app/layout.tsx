import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: '图片人机验证',
  description: 'KeySwift 图片人机验证插件前端组件',
}

export const dynamic = 'force-static'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN" data-theme="dark">
      <body>{children}</body>
    </html>
  )
}
