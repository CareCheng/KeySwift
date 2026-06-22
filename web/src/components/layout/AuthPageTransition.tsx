'use client'

import type { ReactNode } from 'react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { motion, useReducedMotion } from 'framer-motion'
import {
  type AuthMode,
  takeAuthTriggerState,
} from '@/lib/authTransition'

type AuthDirection = 'forward' | 'backward'

interface AuthPageTransitionProps {
  mode: AuthMode
  success?: boolean
  children: ReactNode
}

interface OriginTransform {
  x: number
  y: number
  scaleX: number
  scaleY: number
}

const modeOrder: Record<AuthMode, number> = {
  login: 0,
  register: 1,
}

const springTransition = {
  type: 'spring' as const,
  stiffness: 360,
  damping: 32,
  mass: 0.72,
}

function getDirection(from: AuthMode | null, to: AuthMode): AuthDirection {
  if (!from) return 'forward'
  return modeOrder[to] > modeOrder[from] ? 'forward' : 'backward'
}

function getPanelOrigin(container: HTMLDivElement | null, mode: AuthMode): OriginTransform | null {
  if (!container || typeof window === 'undefined') return null

  const triggerState = takeAuthTriggerState()
  if (!triggerState) return null

  const containerRect = container.getBoundingClientRect()
  const source = triggerState.rect

  return {
    x: source.left + source.width / 2 - (containerRect.left + containerRect.width / 2),
    y: source.top + source.height / 2 - (containerRect.top + containerRect.height / 2),
    scaleX: Math.max(source.width / containerRect.width, 0.12),
    scaleY: Math.max(source.height / containerRect.height, 0.06),
  }
}

/**
 * 用户认证页动效容器。
 * 支持从导航按钮展开、登录/注册横向切换，以及认证成功后向导航用户区域收缩。
 */
export function AuthPageTransition({ mode, success = false, children }: AuthPageTransitionProps) {
  const prefersReducedMotion = useReducedMotion()
  const containerRef = useRef<HTMLDivElement>(null)
  const previousModeRef = useRef<AuthMode | null>(null)
  const [origin, setOrigin] = useState<OriginTransform | null>(null)
  const [originReady, setOriginReady] = useState(true)
  const [direction, setDirection] = useState<AuthDirection>('forward')

  useEffect(() => {
    setDirection(getDirection(previousModeRef.current, mode))
    previousModeRef.current = mode
  }, [mode])

  useEffect(() => {
    if (prefersReducedMotion) {
      setOriginReady(true)
      return
    }

    const frame = window.requestAnimationFrame(() => {
      setOrigin(getPanelOrigin(containerRef.current, mode))
      setOriginReady(true)
    })
    return () => window.cancelAnimationFrame(frame)
  }, [mode, prefersReducedMotion])

  const panelVariants = useMemo(() => {
    const horizontalOffset = direction === 'forward' ? 34 : -34
    const exitOffset = direction === 'forward' ? -30 : 30

    return {
      initial: origin
        ? {
            opacity: 0,
            x: origin.x,
            y: origin.y,
            scaleX: origin.scaleX,
            scaleY: origin.scaleY,
            filter: 'blur(10px)',
            borderRadius: 999,
          }
        : {
            opacity: 0,
            x: horizontalOffset,
            y: 10,
            scale: 0.985,
            filter: 'blur(8px)',
            borderRadius: 28,
          },
      animate: {
        opacity: 1,
        x: 0,
        y: 0,
        scale: 1,
        scaleX: 1,
        scaleY: 1,
        filter: 'blur(0px)',
        borderRadius: 24,
      },
      exit: success
        ? {
            opacity: 0,
            x: 260,
            y: -260,
            scale: 0.18,
            scaleX: 0.18,
            scaleY: 0.12,
            filter: 'blur(10px)',
            borderRadius: 999,
          }
        : {
            opacity: 0,
            x: exitOffset,
            y: -8,
            scale: 0.985,
            filter: 'blur(6px)',
            borderRadius: 28,
          },
    }
  }, [direction, origin, success])

  if (prefersReducedMotion) {
    return <div className="relative w-full max-w-md">{children}</div>
  }

  return (
    <div ref={containerRef} className="relative w-full max-w-md [perspective:1200px]">
      <motion.div
        key={`${mode}-${success ? 'success' : 'active'}`}
        variants={panelVariants}
        initial="initial"
        animate={success ? 'exit' : originReady ? 'animate' : 'initial'}
        transition={springTransition}
        style={{
          originX: 0.5,
          originY: 0.5,
          transformOrigin: 'center center',
          willChange: 'transform, opacity, filter',
        }}
      >
        {children}
      </motion.div>
    </div>
  )
}
