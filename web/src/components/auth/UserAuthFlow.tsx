'use client'

import { useCallback, useEffect, useMemo, useRef, useState, type FormEvent } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { AnimatePresence, motion, useReducedMotion, type Variants } from 'framer-motion'
import { Button, Input, Switch } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import {
  type AuthMode,
  type AuthTransitionRect,
  saveAuthSuccessTarget,
  takeAuthTriggerState,
} from '@/lib/authTransition'
import { useAppStore } from '@/lib/store'
import { cn, isValidEmail } from '@/lib/utils'
import { useI18n } from '@/hooks/useI18n'
import { updateCachedUserInfo } from '@/lib/userData'
import { buildUserRouteUrl } from '@/lib/userNavigation'
import type { UserInfo } from '@/types/user'

type AuthDirection = 1 | -1
type CodeStatus = 'idle' | 'checking' | 'valid' | 'invalid'

interface UserAuthFlowProps {
  initialMode: AuthMode
}

interface PublicAuthConfig {
  user_allow_register: boolean
  user_enable_captcha: boolean
  user_enable_2fa: boolean
  user_require_email_verification: boolean
  email_enabled: boolean
}

interface PanelTransform {
  x: number
  y: number
  scaleX: number
  scaleY: number
}

interface PanelMotionState {
  direction: AuthDirection
  entryTransform: PanelTransform | null
  successTransform: PanelTransform | null
}

interface LoginPanelProps {
  authConfig: PublicAuthConfig
  onAuthenticated: (mode: AuthMode) => void
  onSwitchMode: (mode: AuthMode) => void
}

interface RegisterPanelProps {
  authConfig: PublicAuthConfig
  onAuthenticated: (mode: AuthMode) => void
  onSwitchMode: (mode: AuthMode) => void
}

const modeOrder: Record<AuthMode, number> = {
  login: 0,
  register: 1,
}

// 切换登录/注册时的弹性过渡（已进入页面后使用）
const authSpring = {
  type: 'spring' as const,
  stiffness: 400,
  damping: 32,
  mass: 0.6,
}

// 从触发按钮缩放展开到完整面板的入口过渡
// 用 tween 而非 spring：大距离位移时 spring 启动偏慢，tween 更干脆流畅
const entryTransition = {
  duration: 0.42,
  ease: [0.16, 1, 0.3, 1] as const,
}

const panelVariants: Variants = {
  entryInitial: (state: PanelMotionState) => {
    if (!state.entryTransform) {
      return {
        opacity: 0,
        y: 18,
        scale: 0.985,
      }
    }

    return {
      opacity: 0,
      x: state.entryTransform.x,
      y: state.entryTransform.y,
      scaleX: state.entryTransform.scaleX,
      scaleY: state.entryTransform.scaleY,
    }
  },
  switchInitial: (state: PanelMotionState) => ({
    opacity: 0,
    x: state.direction > 0 ? 96 : -96,
    scale: 0.972,
    rotateY: state.direction > 0 ? -5 : 5,
  }),
  center: {
    opacity: 1,
    x: 0,
    y: 0,
    scale: 1,
    scaleX: 1,
    scaleY: 1,
    rotateY: 0,
  },
  switchExit: (state: PanelMotionState) => ({
    opacity: 0,
    x: state.direction > 0 ? -88 : 88,
    y: -4,
    scale: 0.976,
    rotateY: state.direction > 0 ? 5 : -5,
    transition: {
      duration: 0.24,
      ease: [0.22, 1, 0.36, 1],
    },
  }),
  success: (state: PanelMotionState) => {
    if (!state.successTransform) {
      return {
        opacity: 0,
        y: -28,
        scale: 0.82,
        transition: {
          duration: 0.34,
          ease: [0.7, 0, 0.2, 1],
        },
      }
    }

    return {
      opacity: 0,
      x: state.successTransform.x,
      y: state.successTransform.y,
      scaleX: state.successTransform.scaleX,
      scaleY: state.successTransform.scaleY,
      transition: {
        duration: 0.52,
        ease: [0.74, 0, 0.18, 1],
      },
    }
  },
}

function getEstimatedPanelRect(mode: AuthMode): AuthTransitionRect | null {
  if (typeof window === 'undefined') return null

  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  const panelWidth = Math.min(448, Math.max(288, viewportWidth - 32))
  const panelHeight = mode === 'register' ? Math.min(720, viewportHeight - 64) : 470

  return {
    left: (viewportWidth - panelWidth) / 2,
    top: Math.max(24, (viewportHeight - panelHeight) / 2),
    width: panelWidth,
    height: panelHeight,
  }
}

function getFallbackSuccessTarget(mode: AuthMode): AuthTransitionRect | null {
  if (typeof window === 'undefined') return null

  const width = mode === 'register' ? 118 : 82
  return {
    left: Math.max(16, window.innerWidth - width - 32),
    top: 18,
    width,
    height: 34,
  }
}

function getTransformBetweenRects(from: AuthTransitionRect, to: AuthTransitionRect): PanelTransform {
  return {
    x: to.left + to.width / 2 - (from.left + from.width / 2),
    y: to.top + to.height / 2 - (from.top + from.height / 2),
    scaleX: Math.max(to.width / from.width, 0.12),
    scaleY: Math.max(to.height / from.height, 0.06),
  }
}

function getEntryTransform(source: AuthTransitionRect | null, mode: AuthMode): PanelTransform | null {
  const panel = getEstimatedPanelRect(mode)
  if (!source || !panel) return null
  return getTransformBetweenRects(panel, source)
}

function getSuccessTransform(panel: HTMLDivElement | null, target: AuthTransitionRect | null): PanelTransform | null {
  if (!panel || !target) return null

  const rect = panel.getBoundingClientRect()
  return getTransformBetweenRects(
    {
      left: rect.left,
      top: rect.top,
      width: rect.width,
      height: rect.height,
    },
    target
  )
}

function getModeFromPathname(): AuthMode | null {
  if (typeof window === 'undefined') return null
  if (window.location.pathname.startsWith('/register')) return 'register'
  if (window.location.pathname.startsWith('/login')) return 'login'
  return null
}

function LoginPanel({ authConfig, onAuthenticated, onSwitchMode }: LoginPanelProps) {
  const router = useRouter()
  const { t } = useI18n()
  const { setUser, setIsLoggedIn } = useAppStore()
  const [loading, setLoading] = useState(false)
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImage, setCaptchaImage] = useState('')
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    captcha: '',
    remember: false,
  })

  const refreshCaptcha = useCallback(async () => {
    const res = await apiGet<{ captcha_id: string; image: string }>('/api/captcha')
    if (res.success) {
      setCaptchaId(res.captcha_id)
      setCaptchaImage(res.image)
    }
  }, [])

  useEffect(() => {
    if (authConfig.user_enable_captcha) {
      refreshCaptcha()
    }
  }, [authConfig.user_enable_captcha, refreshCaptcha])

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()
    if (!formData.username || !formData.password || (authConfig.user_enable_captcha && !formData.captcha)) {
      toast.error(t('common.fillComplete'))
      return
    }

    setLoading(true)
    const res = await apiPost<{ require_2fa?: boolean; verify_token?: string; user?: UserInfo }>('/api/user/login', {
      username: formData.username,
      password: formData.password,
      ...(authConfig.user_enable_captcha ? { captcha_id: captchaId, captcha_code: formData.captcha } : {}),
      remember: formData.remember,
    })
    setLoading(false)

    if (res.require_2fa && res.verify_token) {
      router.push(`/verify/?token=${res.verify_token}`)
      return
    }

    if (res.success) {
      toast.success(t('auth.loginSuccess'))
      if (res.user) {
        updateCachedUserInfo(res.user)
        setUser(res.user)
        setIsLoggedIn(true)
      }
      onAuthenticated('login')
      return
    }

    toast.error(res.error || t('auth.loginFailed'))
    if (authConfig.user_enable_captcha) {
      refreshCaptcha()
    }
  }

  return (
    <div className="card p-8 shadow-2xl shadow-primary-950/20">
      <div className="mb-8 text-center">
        <h1 className="mb-2 text-2xl font-bold text-dark-100">{t('auth.loginTitle')}</h1>
        <p className="text-dark-400">{t('auth.loginSubtitle')}</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        <Input
          label={t('user.username')}
          placeholder={t('user.usernamePlaceholder')}
          value={formData.username}
          onChange={(event) => setFormData({ ...formData, username: event.target.value })}
          icon={<i className="fas fa-user" />}
        />

        <Input
          label={t('user.password')}
          type="password"
          placeholder={t('user.passwordPlaceholder')}
          value={formData.password}
          onChange={(event) => setFormData({ ...formData, password: event.target.value })}
          icon={<i className="fas fa-lock" />}
        />

        {authConfig.user_enable_captcha && (
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">{t('user.captcha')}</label>
            <div className="flex items-center gap-3">
              <Input
                placeholder={t('user.captchaPlaceholder')}
                value={formData.captcha}
                onChange={(event) => setFormData({ ...formData, captcha: event.target.value })}
              />
              {captchaImage && (
                <img
                  src={captchaImage}
                  alt={t('user.captcha')}
                  className="h-12 shrink-0 cursor-pointer rounded-lg transition-opacity hover:opacity-80"
                  onClick={refreshCaptcha}
                  title={t('common.clickRefresh')}
                />
              )}
            </div>
          </div>
        )}

        <Switch
          checked={formData.remember}
          onChange={(checked) => setFormData({ ...formData, remember: checked })}
          label={t('user.rememberMe')}
          size="sm"
        />

        <Button type="submit" className="w-full" loading={loading}>
          {t('auth.login')}
        </Button>
      </form>

      <div className="mt-6 flex justify-between text-sm">
        {authConfig.user_allow_register ? (
          <button
            type="button"
            onClick={() => onSwitchMode('register')}
            className="text-primary-400 transition-colors hover:text-primary-300"
          >
            {t('auth.noAccount')}
          </button>
        ) : (
          <span className="text-dark-500">暂未开放注册</span>
        )}
        <Link href="/forgot/" className="text-dark-400 transition-colors hover:text-dark-300">
          {t('auth.forgotPassword')}
        </Link>
      </div>
    </div>
  )
}

function RegisterPanel({ authConfig, onAuthenticated, onSwitchMode }: RegisterPanelProps) {
  const { t } = useI18n()
  const { setUser, setIsLoggedIn } = useAppStore()
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImage, setCaptchaImage] = useState('')
  const [codeLength, setCodeLength] = useState(6)
  const [codeStatus, setCodeStatus] = useState<CodeStatus>('idle')
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    emailCode: '',
    phone: '',
    password: '',
    confirmPassword: '',
    captcha: '',
  })

  const refreshCaptcha = useCallback(async () => {
    const res = await apiGet<{ captcha_id: string; image: string }>('/api/captcha')
    if (res.success) {
      setCaptchaId(res.captcha_id)
      setCaptchaImage(res.image)
    }
  }, [])

  const loadCodeLength = useCallback(async () => {
    const res = await apiGet<{ code_length: number }>('/api/user/email/code_length')
    if (res.success && res.code_length) {
      setCodeLength(res.code_length)
    }
  }, [])

  useEffect(() => {
    if (authConfig.user_enable_captcha) {
      refreshCaptcha()
    }
    if (authConfig.user_require_email_verification) {
      loadCodeLength()
    }
  }, [authConfig.user_enable_captcha, authConfig.user_require_email_verification, loadCodeLength, refreshCaptcha])

  useEffect(() => {
    if (countdown <= 0) return

    const timer = window.setTimeout(() => setCountdown((value) => value - 1), 1000)
    return () => window.clearTimeout(timer)
  }, [countdown])

  const verifyCodeRealtime = useCallback(
    async (code: string) => {
      if (!formData.email || code.length !== codeLength) {
        setCodeStatus('idle')
        return
      }

      setCodeStatus('checking')
      const res = await apiPost<{ valid: boolean }>('/api/user/email/verify_only', {
        email: formData.email,
        code,
        code_type: 'register',
      })

      setCodeStatus(res.success && res.valid ? 'valid' : 'invalid')
    },
    [codeLength, formData.email]
  )

  useEffect(() => {
    if (!authConfig.user_require_email_verification) {
      setCodeStatus('idle')
      return
    }
    if (formData.emailCode.length === codeLength && formData.email) {
      verifyCodeRealtime(formData.emailCode)
      return
    }

    if (formData.emailCode.length < codeLength) {
      setCodeStatus('idle')
    }
  }, [authConfig.user_require_email_verification, codeLength, formData.email, formData.emailCode, verifyCodeRealtime])

  const sendEmailCode = async () => {
    if (!formData.email) {
      toast.error(t('user.emailFirst'))
      return
    }
    if (!isValidEmail(formData.email)) {
      toast.error(t('user.emailInvalid'))
      return
    }

    setSendingCode(true)
    const res = await apiPost('/api/user/email/send_code', {
      email: formData.email,
      code_type: 'register',
    })
    setSendingCode(false)

    if (res.success) {
      toast.success(t('user.codeSent'))
      setCountdown(60)
      setCodeStatus('idle')
      setFormData((prev) => ({ ...prev, emailCode: '' }))
      return
    }

    toast.error(res.error || t('user.codeSendFailed'))
  }

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault()

    if (!authConfig.user_allow_register) {
      toast.error('当前暂未开放注册')
      return
    }
    if (!formData.username || !formData.email || !formData.password || (authConfig.user_require_email_verification && !formData.emailCode)) {
      toast.error(t('common.fillComplete'))
      return
    }
    if (authConfig.user_enable_captcha && !formData.captcha) {
      toast.error(t('common.fillComplete'))
      return
    }
    if (formData.password !== formData.confirmPassword) {
      toast.error(t('user.passwordMismatch'))
      return
    }
    if (formData.password.length < 6) {
      toast.error(t('user.passwordTooShort'))
      return
    }
    if (authConfig.user_require_email_verification && codeStatus === 'invalid') {
      toast.error(t('user.codeIncorrect'))
      return
    }

    setLoading(true)
    const res = await apiPost<{ user?: UserInfo }>('/api/user/register', {
      username: formData.username,
      email: formData.email,
      ...(authConfig.user_require_email_verification ? { email_code: formData.emailCode } : {}),
      phone: formData.phone,
      password: formData.password,
      confirm_password: formData.confirmPassword,
      ...(authConfig.user_enable_captcha ? { captcha_id: captchaId, captcha_code: formData.captcha } : {}),
    })
    setLoading(false)

    if (res.success && res.user) {
      toast.success(t('auth.registerSuccess'))
      updateCachedUserInfo(res.user)
      setUser(res.user)
      setIsLoggedIn(true)
      onAuthenticated('register')
      return
    }

    toast.error(res.error || t('auth.registerFailed'))
    if (authConfig.user_enable_captcha) {
      refreshCaptcha()
    }
  }

  const getCodeStatusIcon = () => {
    switch (codeStatus) {
      case 'checking':
        return <i className="fas fa-spinner fa-spin text-primary-400" />
      case 'valid':
        return <i className="fas fa-check-circle text-green-400" />
      case 'invalid':
        return <i className="fas fa-times-circle text-red-400" />
      default:
        return null
    }
  }

  return (
    <div className="card p-8 shadow-2xl shadow-primary-950/20">
      <div className="mb-8 text-center">
        <h1 className="mb-2 text-2xl font-bold text-dark-100">{t('auth.registerTitle')}</h1>
        <p className="text-dark-400">{t('auth.registerSubtitle')}</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label={t('user.username')}
          placeholder={t('user.usernamePlaceholder')}
          value={formData.username}
          onChange={(event) => setFormData({ ...formData, username: event.target.value })}
          icon={<i className="fas fa-user" />}
        />

        <Input
          label={t('user.email')}
          type="email"
          placeholder={t('user.emailPlaceholder')}
          value={formData.email}
          onChange={(event) => setFormData({ ...formData, email: event.target.value })}
          icon={<i className="fas fa-envelope" />}
        />

        {authConfig.user_require_email_verification && (
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">{t('user.emailCode')}</label>
            <div className="flex items-center gap-3">
              <div className="relative flex-1">
                <Input
                  placeholder={t('user.emailCodePlaceholder').replace('{length}', String(codeLength))}
                  value={formData.emailCode}
                  onChange={(event) => {
                    const value = event.target.value.replace(/\D/g, '').slice(0, codeLength)
                    setFormData({ ...formData, emailCode: value })
                  }}
                  maxLength={codeLength}
                  className={cn(
                    codeStatus === 'valid' && 'border-green-500/50 focus:border-green-500',
                    codeStatus === 'invalid' && 'border-red-500/50 focus:border-red-500'
                  )}
                />
                <div className="absolute right-3 top-1/2 -translate-y-1/2">{getCodeStatusIcon()}</div>
              </div>
              <Button type="button" variant="secondary" onClick={sendEmailCode} disabled={countdown > 0 || sendingCode}>
                {countdown > 0
                  ? `${countdown}${t('common.seconds')}`
                  : sendingCode
                    ? t('common.sending')
                    : t('user.sendCode')}
              </Button>
            </div>
            {codeStatus === 'valid' && (
              <p className="mt-1 text-xs text-green-400">
                <i className="fas fa-check mr-1" />
                {t('user.codeCorrect')}
              </p>
            )}
            {codeStatus === 'invalid' && (
              <p className="mt-1 text-xs text-red-400">
                <i className="fas fa-times mr-1" />
                {t('user.codeIncorrect')}
              </p>
            )}
          </div>
        )}

        <Input
          label={t('user.phoneOptional')}
          type="tel"
          placeholder={t('user.phonePlaceholder')}
          value={formData.phone}
          onChange={(event) => setFormData({ ...formData, phone: event.target.value })}
          icon={<i className="fas fa-phone" />}
        />

        <Input
          label={t('user.password')}
          type="password"
          placeholder={t('user.passwordMinLength')}
          value={formData.password}
          onChange={(event) => setFormData({ ...formData, password: event.target.value })}
          icon={<i className="fas fa-lock" />}
        />

        <Input
          label={t('user.confirmPassword')}
          type="password"
          placeholder={t('user.confirmPasswordPlaceholder')}
          value={formData.confirmPassword}
          onChange={(event) => setFormData({ ...formData, confirmPassword: event.target.value })}
          icon={<i className="fas fa-lock" />}
        />

        {authConfig.user_enable_captcha && (
          <div className="space-y-1.5">
            <label className="block text-sm font-medium text-dark-300">{t('user.captcha')}</label>
            <div className="flex items-center gap-3">
              <Input
                placeholder={t('user.captchaPlaceholder')}
                value={formData.captcha}
                onChange={(event) => setFormData({ ...formData, captcha: event.target.value })}
              />
              {captchaImage && (
                <img
                  src={captchaImage}
                  alt={t('user.captcha')}
                  className="h-12 shrink-0 cursor-pointer rounded-lg transition-opacity hover:opacity-80"
                  onClick={refreshCaptcha}
                  title={t('common.clickRefresh')}
                />
              )}
            </div>
          </div>
        )}

        <Button type="submit" className="w-full" loading={loading}>
          {t('auth.register')}
        </Button>
      </form>

      <div className="mt-6 flex justify-between text-sm">
        <button
          type="button"
          onClick={() => onSwitchMode('login')}
          className="text-primary-400 transition-colors hover:text-primary-300"
        >
          {t('auth.hasAccount')}
        </button>
        <Link href={buildUserRouteUrl('home')} className="text-dark-400 transition-colors hover:text-dark-300">
          {t('auth.backToHome')}
        </Link>
      </div>
    </div>
  )
}

/**
 * 用户端登录注册统一认证流。
 * 登录/注册互切在同一 React 树内完成，避免跨页面卸载导致过渡动画不可见。
 */
export function UserAuthFlow({ initialMode }: UserAuthFlowProps) {
  const router = useRouter()
  const prefersReducedMotion = useReducedMotion()
  const panelRef = useRef<HTMLDivElement>(null)
  const redirectTimerRef = useRef<number | null>(null)
  const [entryState] = useState(() => (typeof window === 'undefined' ? null : takeAuthTriggerState()))
  const [mode, setMode] = useState<AuthMode>(initialMode)
  const [hasEntered, setHasEntered] = useState(false)
  const [success, setSuccess] = useState(false)
  const [direction, setDirection] = useState<AuthDirection>(1)
  const [successTransform, setSuccessTransform] = useState<PanelTransform | null>(null)
  const [willChange, setWillChange] = useState<'transform, opacity' | 'auto'>('transform, opacity')
  const [authConfig, setAuthConfig] = useState<PublicAuthConfig>({
    user_allow_register: true,
    user_enable_captcha: true,
    user_enable_2fa: true,
    user_require_email_verification: false,
    email_enabled: false,
  })

  const entryTransform = useMemo(
    () => (entryState ? getEntryTransform(entryState.rect, initialMode) : null),
    [entryState, initialMode]
  )

  useEffect(() => {
    let cancelled = false
    const loadAuthConfig = async () => {
      const res = await apiGet<{ config: PublicAuthConfig }>('/api/auth/config')
      if (!cancelled && res.success && res.config) {
        setAuthConfig(res.config)
        if (!res.config.user_allow_register && mode === 'register') {
          setMode('login')
          window.history.replaceState({ authMode: 'login' }, '', '/login/')
        }
      }
    }

    loadAuthConfig()
    return () => {
      cancelled = true
    }
  }, [mode])

  useEffect(() => {
    return () => {
      if (redirectTimerRef.current) {
        window.clearTimeout(redirectTimerRef.current)
      }
    }
  }, [])

  useEffect(() => {
    const handlePopState = () => {
      const nextMode = getModeFromPathname()
      if (!nextMode || nextMode === mode) return
      setDirection(modeOrder[nextMode] > modeOrder[mode] ? 1 : -1)
      setHasEntered(true)
      setMode(nextMode)
    }

    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [mode])

  const switchMode = (nextMode: AuthMode) => {
    if (nextMode === mode || success) return
    if (nextMode === 'register' && !authConfig.user_allow_register) {
      toast.error('当前暂未开放注册')
      return
    }

    setDirection(modeOrder[nextMode] > modeOrder[mode] ? 1 : -1)
    setHasEntered(true)
    setMode(nextMode)
    setWillChange('transform, opacity')

    const nextUrl = nextMode === 'login' ? '/login/' : '/register/'
    window.history.pushState({ authMode: nextMode }, '', nextUrl)
  }

  const handleAuthenticated = (authMode: AuthMode) => {
    const target = entryState?.rect ?? getFallbackSuccessTarget(authMode)
    setSuccessTransform(getSuccessTransform(panelRef.current, target))
    if (target) {
      saveAuthSuccessTarget(authMode, target)
    }
    setSuccess(true)
    setWillChange('transform, opacity')
    redirectTimerRef.current = window.setTimeout(() => {
      router.push(buildUserRouteUrl('products'))
    }, prefersReducedMotion ? 80 : 560)
  }

  const motionState: PanelMotionState = {
    direction,
    entryTransform,
    successTransform,
  }

  return (
    <motion.div
      initial={prefersReducedMotion ? false : { opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
      className="relative flex min-h-screen items-center justify-center overflow-hidden p-4 py-12"
    >
      <div
        className="absolute inset-0"
        style={{
          backgroundImage: [
            'radial-gradient(circle 320px at calc(100% + 160px) -160px, rgba(102,126,234,0.20), transparent 70%)',
            'radial-gradient(circle 320px at -160px calc(100% + 160px), rgba(168,85,247,0.20), transparent 70%)',
            'radial-gradient(circle 448px at 50% 50%, rgba(6,182,212,0.05), transparent 70%)',
          ].join(', '),
          backgroundRepeat: 'no-repeat',
        }}
      />

      <div className="relative w-full max-w-md [perspective:1200px]">
        <AnimatePresence mode="popLayout" custom={motionState}>
          <motion.div
            ref={panelRef}
            key={mode}
            custom={motionState}
            variants={panelVariants}
            initial={prefersReducedMotion ? false : hasEntered ? 'switchInitial' : 'entryInitial'}
            animate={prefersReducedMotion ? undefined : success ? 'success' : 'center'}
            exit={prefersReducedMotion ? undefined : 'switchExit'}
            transition={hasEntered ? authSpring : entryTransition}
            style={{
              transformOrigin: 'center center',
              willChange,
            }}
            onAnimationComplete={(definition) => {
              if (definition === 'center') {
                setHasEntered(true)
                setWillChange('auto')
              }
            }}
          >
            <motion.div
              initial={prefersReducedMotion ? false : { opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.26, delay: 0.1, ease: [0.22, 1, 0.36, 1] }}
            >
              {mode === 'login' ? (
                <LoginPanel authConfig={authConfig} onAuthenticated={handleAuthenticated} onSwitchMode={switchMode} />
              ) : (
                <RegisterPanel authConfig={authConfig} onAuthenticated={handleAuthenticated} onSwitchMode={switchMode} />
              )}
            </motion.div>
          </motion.div>
        </AnimatePresence>
      </div>
    </motion.div>
  )
}
