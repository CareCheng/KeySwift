'use client'

import { useCallback, useEffect, useRef, useState } from 'react'
import { postToHost, type HumanVerificationPayload, type HumanVerificationScope, type WidgetInbound } from '@/lib/protocol'
import { applyThemeTokens, type ThemeTokenPack } from '@/lib/themeTokens'

interface ChallengeResponse {
  challenge_id: string
  image: string
}

interface WidgetState {
  scope: HumanVerificationScope | ''
  providerId: string
  providerType: string
  endpoint: string
  challengeId: string
  loading: boolean
}

/**
 * 图片人机验证 React 组件。
 * 职责：接收宿主 init/theme 消息 → 拉取图片挑战 → 回传答案 payload。
 * 样式与登录页用户名/密码输入框一致（h-12 / rounded-xl / 紫色焦点），主题由宿主注入。
 */
export function ImageCaptchaWidget() {
  const [answer, setAnswer] = useState('')
  const [status, setStatus] = useState('')
  const [statusError, setStatusError] = useState(false)
  const [imageSrc, setImageSrc] = useState('')
  const [imageState, setImageState] = useState<'loading' | 'loaded' | 'error'>('loading')
  const [height, setHeight] = useState(96)

  const stateRef = useRef<WidgetState>({
    scope: '',
    providerId: '',
    providerType: 'image_captcha',
    endpoint: '/api/human-verification/challenge',
    challengeId: '',
    loading: false,
  })

  const resize = useCallback(() => {
    const h = typeof document !== 'undefined' ? Math.max(58, document.body.scrollHeight) : 96
    setHeight(h)
    postToHost({ type: 'resize', height: h })
  }, [])

  const clearPayload = useCallback(() => {
    postToHost({ type: 'change', payload: null })
  }, [])

  const emitAnswer = useCallback((value: string) => {
    const trimmed = value.trim()
    const state = stateRef.current
    if (!trimmed || !state.challengeId) {
      clearPayload()
      return
    }
    const payload: HumanVerificationPayload = {
      provider_id: state.providerId,
      provider_type: state.providerType,
      scope: state.scope as HumanVerificationScope,
      challenge_id: state.challengeId,
      answer: trimmed,
    }
    postToHost({ type: 'change', payload })
  }, [clearPayload])

  const loadChallenge = useCallback(async () => {
    const state = stateRef.current
    if (!state.scope || !state.providerId || state.loading) return
    state.loading = true
    setImageState('loading')
    setStatus('')
    setStatusError(false)
    setAnswer('')
    state.challengeId = ''
    clearPayload()

    try {
      const response = await fetch(state.endpoint, {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ scope: state.scope, provider_id: state.providerId }),
      })
      const result = (await response.json()) as { success?: boolean; challenge?: ChallengeResponse; error?: string }
      if (!result.success || !result.challenge) {
        throw new Error(result.error || '加载失败')
      }
      state.challengeId = result.challenge.challenge_id || ''
      if (!state.challengeId || !result.challenge.image) {
        throw new Error('挑战内容不完整')
      }
      setImageSrc(result.challenge.image)
      setImageState('loaded')
    } catch (error) {
      setImageState('error')
      const message = error instanceof Error ? error.message : '加载失败'
      setStatus(message)
      setStatusError(true)
      postToHost({ type: 'error', error: message })
    } finally {
      state.loading = false
      resize()
    }
  }, [clearPayload, resize])

  // 应用宿主下发的主题 token 包，写入自身 :root；缺字段由 applyThemeTokens 用默认值兜底
  const applyTheme = useCallback((pack?: ThemeTokenPack) => {
    applyThemeTokens(pack)
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
      const prev = stateRef.current
      const scopeChanged = prev.scope !== (data.scope || '') || prev.providerId !== (data.provider_id || '')
      stateRef.current = {
        ...prev,
        scope: (data.scope || '') as HumanVerificationScope | '',
        providerId: data.provider_id || '',
        providerType: data.provider_type || 'image_captcha',
        endpoint: data.challenge_endpoint || '/api/human-verification/challenge',
      }
      applyTheme(data.theme_tokens)
      // 仅首次初始化或 provider 变更时拉取挑战，主题切换不刷新验证码
      if (scopeChanged || !prev.challengeId) {
        loadChallenge()
      }
    }

    window.addEventListener('message', handleMessage)
    return () => window.removeEventListener('message', handleMessage)
  }, [applyTheme, loadChallenge])

  useEffect(() => {
    postToHost({ type: 'ready' })
    resize()
  }, [resize])

  // 内容变化时通知宿主调整 iframe 高度
  useEffect(() => {
    resize()
  }, [status, imageState, height, resize])

  const handleAnswerChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value
    setAnswer(value)
    emitAnswer(value)
  }

  return (
    <div className="w-full">
      {/*
        结构对齐参考程序 Source/KeySwift-old 登录页验证码：
        输入框（含左侧图标，与用户名/密码框同款 .captcha-input）与验证码图作为并列 flex 项，
        gap-3 间隔，验证码图独立无外层容器边框，避免与输入框边框叠加产生"双层边框"。
      */}
      <div className="flex items-center gap-3">
        <div className="relative min-w-0 flex-1">
          <span
            className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]"
            aria-hidden="true"
          >
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 3l8 4v5c0 5-3.5 8-8 9-4.5-1-8-4-8-9V7l8-4z" />
              <path d="M9 12l2 2 4-4" />
            </svg>
          </span>
          <input
            className="captcha-input captcha-input--with-icon"
            autoComplete="off"
            inputMode="numeric"
            placeholder="请输入人机验证答案"
            value={answer}
            onChange={handleAnswerChange}
          />
        </div>
        {imageState === 'loaded' && imageSrc ? (
          <img
            src={imageSrc}
            alt="人机验证"
            className="captcha-image"
            onClick={loadChallenge}
            title="点击刷新验证码"
          />
        ) : (
          <button
            type="button"
            onClick={loadChallenge}
            className="captcha-image-placeholder"
            title="点击刷新验证码"
          >
            {imageState === 'error' ? '重试' : '加载中'}
          </button>
        )}
      </div>
      {status && (
        <p className={`mt-1.5 text-xs leading-5 ${statusError ? 'text-red-300' : 'text-[var(--text-muted)]'}`}>{status}</p>
      )}
    </div>
  )
}
