'use client'

import type { ReactNode } from 'react'
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'
import { usePathname } from 'next/navigation'

const standaloneTransitionPaths = [
  '/login',
  '/register',
  '/forgot',
  '/verify',
  '/admin/login',
  '/admin/setup',
  '/admin/totp',
]

function shouldUseStandaloneTransition(pathname: string) {
  const normalizedPath = pathname.replace(/\/+$/, '') || '/'
  return standaloneTransitionPaths.includes(normalizedPath)
}

/**
 * 页面级路由过渡容器。
 * 只跟随 pathname 变化，用户端 hash 内部切页继续由 UserPageTransition 负责。
 * 认证页拥有独立过渡容器，不能再叠加父级淡入，否则会吞掉入口展开动画。
 */
export function RouteTransitionProvider({ children }: { children: ReactNode }) {
  const pathname = usePathname()
  const prefersReducedMotion = useReducedMotion()

  if (shouldUseStandaloneTransition(pathname)) {
    return <>{children}</>
  }

  return (
    <AnimatePresence mode="wait" initial={false}>
      <motion.div
        key={pathname}
        initial={prefersReducedMotion ? { opacity: 1 } : { opacity: 0, y: 8 }}
        animate={prefersReducedMotion ? { opacity: 1 } : { opacity: 1, y: 0 }}
        exit={prefersReducedMotion ? { opacity: 1 } : { opacity: 0, y: -6 }}
        transition={{ duration: 0.18, ease: [0.22, 1, 0.36, 1] }}
        className="min-h-screen"
      >
        {children}
      </motion.div>
    </AnimatePresence>
  )
}
