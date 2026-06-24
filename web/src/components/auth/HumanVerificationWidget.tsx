'use client'

import { useEffect, useMemo, useRef, useState } from 'react'
import { useTheme } from '@/lib/theme'
import { readThemeTokensFromDocument, type ThemeTokenPack } from '@/lib/themeTokens'

export type HumanVerificationScope = 'admin_login' | 'user_login' | 'user_register'

export interface HumanVerificationPayload {
  provider_id: string
  provider_type: string
  scope: HumanVerificationScope
  token?: string
  challenge_id?: string
  answer?: string
  client_nonce?: string
  metadata?: Record<string, unknown>
}

export interface PublicHumanVerificationConfig {
  enabled: boolean
  scope: HumanVerificationScope
  provider_id?: string
  provider_type?: string
  display_name?: string
  render_mode?: string
  frontend_url?: string
  frontend_height?: number
  challenge_endpoint?: string
  public_config?: Record<string, unknown>
  error?: string
}

interface HumanVerificationWidgetProps {
  scope: HumanVerificationScope
  config?: PublicHumanVerificationConfig
  onChange: (payload: HumanVerificationPayload | null) => void
  resetSignal?: number
}

interface WidgetMessage {
  source?: string
  type?: string
  payload?: HumanVerificationPayload | null
  error?: string
  height?: number
}

// 推送给插件 iframe 的主题消息：下发标准主题 token 包，插件前端写入自身 :root
interface ThemeMessage {
  source: 'keyswift-human-verification-host'
  type: 'theme'
  theme_tokens: ThemeTokenPack
}

const messageSource = 'keyswift-human-verification-plugin'

/**
 * 人机验证宿主挂载组件。
 * 职责：按插件 manifest 提供的 iframe 入口挂载插件前端，并接收标准消息协议。
 */
export function HumanVerificationWidget({
  scope,
  config,
  onChange,
  resetSignal = 0,
}: HumanVerificationWidgetProps) {
  const iframeRef = useRef<HTMLIFrameElement | null>(null)
  const [error, setError] = useState('')
  const [height, setHeight] = useState(96)
  const { theme } = useTheme()
  // 当前主题 token 包：从 document CSS 变量读取，主题切换时重读
  const [themeTokens, setThemeTokens] = useState<ThemeTokenPack>({})

  const enabled = Boolean(config?.enabled)
  const frontendURL = config?.frontend_url || ''
  const providerID = config?.provider_id || ''
  const providerType = config?.provider_type || ''

  // 主题变化后，等 DOM 应用新 CSS 变量，再读取 token 包
  useEffect(() => {
    setThemeTokens(readThemeTokensFromDocument())
  }, [theme])

  const initPayload = useMemo(
    () => ({
      source: 'keyswift-human-verification-host',
      type: 'init',
      scope,
      provider_id: providerID,
      provider_type: providerType,
      public_config: config?.public_config || {},
      challenge_endpoint: config?.challenge_endpoint || '/api/human-verification/challenge',
      reset_signal: resetSignal,
      theme_tokens: themeTokens,
    }),
    [config?.challenge_endpoint, config?.public_config, providerID, providerType, resetSignal, scope, themeTokens],
  )

  // 主题 token 变化时仅推送配色，不重置挑战，避免验证码被刷新打断用户输入
  useEffect(() => {
    if (!enabled || !frontendURL) return
    const payload: ThemeMessage = {
      source: 'keyswift-human-verification-host',
      type: 'theme',
      theme_tokens: themeTokens,
    }
    iframeRef.current?.contentWindow?.postMessage(payload, window.location.origin)
  }, [enabled, frontendURL, themeTokens])

  useEffect(() => {
    if (!enabled) {
      setError('')
      onChange(null)
      return
    }
    if (config?.error) {
      setError(config.error)
      onChange(null)
      return
    }
    if (!providerID || !providerType) {
      setError('当前人机验证插件声明不完整')
      onChange(null)
      return
    }
    if (!frontendURL) {
      setError('当前人机验证插件未声明前端入口')
      onChange(null)
      return
    }
    setError('')
    setHeight(config?.frontend_height || 96)
    onChange(null)
  }, [config?.error, config?.frontend_height, enabled, frontendURL, onChange, providerID, providerType])

  useEffect(() => {
    const handleMessage = (event: MessageEvent<WidgetMessage>) => {
      if (event.origin !== window.location.origin) return
      const data = event.data || {}
      if (data.source !== messageSource) return

      switch (data.type) {
        case 'ready':
          iframeRef.current?.contentWindow?.postMessage(initPayload, window.location.origin)
          break
        case 'change':
          if (data.payload) {
            onChange({
              ...data.payload,
              provider_id: providerID,
              provider_type: providerType,
              scope,
            })
          } else {
            onChange(null)
          }
          break
        case 'error':
          setError(data.error || '人机验证组件异常')
          onChange(null)
          break
        case 'resize':
          if (typeof data.height === 'number' && data.height >= 48 && data.height <= 420) {
            setHeight(data.height)
          }
          break
      }
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [initPayload, onChange, providerID, providerType, scope])

  useEffect(() => {
    iframeRef.current?.contentWindow?.postMessage(initPayload, window.location.origin)
  }, [initPayload])

  if (!enabled) return null

  if (config?.error || error) {
    return (
      <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-sm text-red-300">
        <i className="fas fa-exclamation-triangle mr-2" />
        {config?.error || error}
      </div>
    )
  }

  return (
    <div className="space-y-1.5">
      <label className="block text-sm font-medium" style={{ color: 'var(--text-secondary)' }}>
        {config?.display_name || '人机验证'}
      </label>
      <iframe
        key={`${providerID}:${resetSignal}`}
        ref={iframeRef}
        src={frontendURL}
        title={config?.display_name || '人机验证'}
        className="w-full rounded-xl border-0 bg-transparent"
        style={{ height }}
        sandbox="allow-scripts allow-same-origin allow-forms"
      />
    </div>
  )
}
