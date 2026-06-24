'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import { postToHost, type HumanVerificationPayload, type HumanVerificationScope, type WidgetInbound } from '@/lib/protocol'
import { applyThemeTokens, type ThemeTokenPack } from '@/lib/themeTokens'

/* Cloudflare Turnstile 全局 API（由 challenges.cloudflare.com 脚本注入） */
declare global {
  interface Window {
    turnstile?: {
      render: (
        container: HTMLElement,
        options: {
          sitekey: string
          theme?: 'light' | 'dark' | 'auto'
          size?: 'normal' | 'flexible' | 'compact'
          appearance?: 'always' | 'execute' | 'interaction-only'
          callback?: (token: string) => void
          'expired-callback'?: () => void
          'error-callback'?: () => void
        }
      ) => string
      remove?: (widgetId: string) => void
    }
  }
}

interface WidgetState {
  scope: HumanVerificationScope | ''
  providerId: string
  providerType: string
  publicConfig: Record<string, unknown>
  widgetId: string | null
  rendered: boolean
}

const TURNSTILE_SCRIPT_SRC = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'

function generateNonce(): string {
  if (typeof window !== 'undefined' && window.crypto && typeof window.crypto.randomUUID === 'function') {
    return window.crypto.randomUUID()
  }
  return `ts-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

/**
 * Cloudflare Turnstile 人机验证 React 组件。
 * 职责：注入 Turnstile 脚本 → 接收宿主 init/theme 消息 → 渲染 widget → 回传 token payload。
 */
export function TurnstileWidget() {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [status, setStatus] = useState('')
  const [statusError, setStatusError] = useState(false)

  const stateRef = useRef<WidgetState>({
    scope: '',
    providerId: '',
    providerType: 'cloudflare_turnstile',
    publicConfig: {},
    widgetId: null,
    rendered: false,
  })

  const resize = useCallback(() => {
    if (typeof document === 'undefined') return
    postToHost({ type: 'resize', height: Math.max(82, document.body.scrollHeight) })
  }, [])

  const setStatusText = useCallback(
    (text: string, isError = false) => {
      setStatus(text)
      setStatusError(isError)
      resize()
    },
    [resize]
  )

  const clearPayload = useCallback(() => {
    postToHost({ type: 'change', payload: null })
  }, [])

  // 应用宿主下发的主题 token 包，写入自身 :root；缺字段由 applyThemeTokens 用默认值兜底
  const applyTheme = useCallback((pack?: ThemeTokenPack) => {
    applyThemeTokens(pack)
  }, [])

  const renderWidget = useCallback(() => {
    const state = stateRef.current

    if (!window.turnstile || typeof window.turnstile.render !== 'function') {
      // 脚本尚未就绪，轮询等待
      window.setTimeout(renderWidget, 100)
      return
    }

    const siteKey = String(state.publicConfig.site_key || '').trim()
    if (!siteKey) {
      setStatusText('Turnstile 站点密钥未配置', true)
      postToHost({ type: 'error', error: 'Turnstile 站点密钥未配置' })
      return
    }

    if (state.widgetId !== null && typeof window.turnstile.remove === 'function') {
      window.turnstile.remove(state.widgetId)
    }
    if (containerRef.current) {
      containerRef.current.innerHTML = ''
    }
    clearPayload()

    if (!containerRef.current) return

    state.widgetId = window.turnstile.render(containerRef.current, {
      sitekey: siteKey,
      theme: (state.publicConfig.theme as 'light' | 'dark' | 'auto') || 'auto',
      size: (state.publicConfig.size as 'normal' | 'flexible' | 'compact') || 'normal',
      appearance: (state.publicConfig.appearance as 'always' | 'execute' | 'interaction-only') || 'always',
      callback: (token: string) => {
        setStatusText('', false)
        const payload: HumanVerificationPayload = {
          provider_id: state.providerId,
          provider_type: state.providerType,
          scope: state.scope as HumanVerificationScope,
          token,
          client_nonce: generateNonce(),
        }
        postToHost({ type: 'change', payload })
      },
      'expired-callback': () => {
        setStatusText('验证已过期，请重新完成人机验证')
        clearPayload()
      },
      'error-callback': () => {
        setStatusText('Turnstile 组件加载失败', true)
        clearPayload()
        postToHost({ type: 'error', error: 'Turnstile 组件加载失败' })
      },
    })
    state.rendered = true
    resize()
  }, [clearPayload, resize, setStatusText])

  // 注入 Turnstile 脚本（仅一次）
  useEffect(() => {
    if (typeof document === 'undefined') return
    const existing = document.querySelector(`script[src="${TURNSTILE_SCRIPT_SRC}"]`)
    if (existing) return
    const script = document.createElement('script')
    script.src = TURNSTILE_SCRIPT_SRC
    script.async = true
    script.defer = true
    document.head.appendChild(script)
  }, [])

  useEffect(() => {
    const handleMessage = (event: MessageEvent<WidgetInbound>) => {
      if (event.origin !== window.location.origin) return
      const data = event.data
      if (!data || data.source !== 'keyswift-human-verification-host') return

      if (data.type === 'theme') {
        applyTheme(data.theme_tokens)
        return
      }

      if (data.type !== 'init') return
      stateRef.current = {
        ...stateRef.current,
        scope: (data.scope || '') as HumanVerificationScope | '',
        providerId: data.provider_id || '',
        providerType: data.provider_type || 'cloudflare_turnstile',
        publicConfig: data.public_config || {},
      }
      applyTheme(data.theme_tokens)
      renderWidget()
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [applyTheme, renderWidget])

  useEffect(() => {
    postToHost({ type: 'ready' })
    resize()
  }, [resize])

  useEffect(() => {
    resize()
  }, [status, resize])

  return (
    <div className="w-full">
      <div ref={containerRef} className="min-h-[65px]" />
      <div className={`min-h-[20px] text-xs leading-5 ${statusError ? 'text-red-300' : 'text-[var(--text-muted)]'}`}>{status}</div>
    </div>
  )
}
