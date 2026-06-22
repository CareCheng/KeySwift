'use client'

import type { ReactNode } from 'react'
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'

interface UserPageTransitionProps {
  pageKey: string
  children: ReactNode
}

/**
 * 用户端页面过渡容器。
 * 只作用于主内容区域，保留导航栏和页脚稳定，避免页面切换时产生整页闪烁。
 */
export function UserPageTransition({ pageKey, children }: UserPageTransitionProps) {
  const shouldReduceMotion = useReducedMotion()

  if (shouldReduceMotion) {
    return <>{children}</>
  }

  return (
    <AnimatePresence mode="wait" initial={false}>
      <motion.div
        key={pageKey}
        initial={{ opacity: 0, y: 12, filter: 'blur(4px)' }}
        animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
        exit={{ opacity: 0, y: -8, filter: 'blur(3px)' }}
        transition={{
          duration: 0.24,
          ease: [0.22, 1, 0.36, 1],
        }}
        className="flex-1 flex flex-col will-change-transform"
      >
        {children}
      </motion.div>
    </AnimatePresence>
  )
}
