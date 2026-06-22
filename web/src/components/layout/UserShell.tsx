'use client'

import type { ReactNode } from 'react'
import { Footer, Navbar } from './Navbar'

/**
 * 用户端统一页面外壳。
 * 主入口 hash 切换和旧路径直达都复用这一层，避免导航栏、页脚和页面布局重复实现。
 */
export function UserShell({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      {children}
      <Footer />
    </div>
  )
}
