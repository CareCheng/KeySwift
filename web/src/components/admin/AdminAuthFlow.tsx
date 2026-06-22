'use client'

import type { ReactNode } from 'react'
import { ChangeEvent, FormEvent, useCallback, useEffect, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import toast from 'react-hot-toast'
import { Button, Input, Switch } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import AdminDashboardApp from './AdminDashboardApp'

export type AdminAuthStep = 'login' | 'setup' | 'totp'
type AdminAuthFlowState = AdminAuthStep | 'dashboard'

const TRANSIENT_STEP_KEY = 'keysWift.adminAuthStep'

interface AdminAuthFlowProps {
  initialStep: AdminAuthStep
}

interface AuthPanelProps {
  basePath: string
  onNavigate: (step: AdminAuthFlowState) => void
}

interface PublicAuthConfig {
  admin_enable_login: boolean
  admin_enable_captcha: boolean
}

function getAdminBasePath() {
  if (typeof window === 'undefined') return ''

  const parts = window.location.pathname.replace(/\/+$/, '').split('/').filter(Boolean)
  const last = parts[parts.length - 1]
  if (last === 'login' || last === 'setup' || last === 'totp') {
    parts.pop()
  }

  return parts.length > 0 ? `/${parts.join('/')}` : '/'
}

function buildAdminPath(basePath: string, step?: AdminAuthStep) {
  const normalizedBase = basePath.replace(/\/+$/, '') || ''
  return step ? `${normalizedBase}/${step}/` : `${normalizedBase}/`
}

function isAdminAuthFlowState(value: string | null): value is AdminAuthFlowState {
  return value === 'login' || value === 'setup' || value === 'totp' || value === 'dashboard'
}

function takeTransientStep(defaultStep: AdminAuthStep): AdminAuthFlowState {
  if (typeof window === 'undefined') return defaultStep

  const value = window.sessionStorage.getItem(TRANSIENT_STEP_KEY)
  if (isAdminAuthFlowState(value)) {
    window.sessionStorage.removeItem(TRANSIENT_STEP_KEY)
    return value
  }

  return defaultStep
}

function saveTransientStep(step: AdminAuthFlowState) {
  if (typeof window === 'undefined') return
  window.sessionStorage.setItem(TRANSIENT_STEP_KEY, step)
  window.setTimeout(() => {
    if (window.sessionStorage.getItem(TRANSIENT_STEP_KEY) === step) {
      window.sessionStorage.removeItem(TRANSIENT_STEP_KEY)
    }
  }, 1000)
}

function AuthCard({
  children,
  panelKey,
}: {
  children: ReactNode
  panelKey: string
}) {
  return (
    <motion.div
      key={panelKey}
      initial={{ opacity: 0, y: 18, filter: 'blur(4px)' }}
      animate={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
      exit={{ opacity: 0, y: -12, filter: 'blur(3px)' }}
      transition={{ duration: 0.22, ease: [0.22, 1, 0.36, 1] }}
      className="w-full max-w-md"
    >
      {children}
    </motion.div>
  )
}

function AdminLoginPanel({ basePath, onNavigate }: AuthPanelProps) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [captchaCode, setCaptchaCode] = useState('')
  const [captchaId, setCaptchaId] = useState('')
  const [captchaImage, setCaptchaImage] = useState('')
  const [remember, setRemember] = useState(false)
  const [loading, setLoading] = useState(false)
  const [checking, setChecking] = useState(true)
  const [authConfig, setAuthConfig] = useState<PublicAuthConfig>({
    admin_enable_login: true,
    admin_enable_captcha: true,
  })

  const loadCaptcha = useCallback(async () => {
    const res = await apiGet<{ captcha_id: string; image: string }>('/api/captcha')
    if (res.success && res.captcha_id) {
      setCaptchaId(res.captcha_id)
      setCaptchaImage(res.image || '')
    }
  }, [])

  useEffect(() => {
    let cancelled = false

    const checkSetup = async () => {
      const currentBasePath = basePath || getAdminBasePath()
      try {
        const res = await apiGet<{ needs_setup: boolean }>(`${currentBasePath}/check-setup`)
        if (!cancelled && res.success && res.needs_setup) {
          onNavigate('setup')
          return
        }
      } catch {
        // 初始化检查失败时继续显示登录页，由登录接口返回明确错误。
      }

      try {
        const configRes = await apiGet<{ config: PublicAuthConfig }>('/api/auth/config')
        if (!cancelled && configRes.success && configRes.config) {
          setAuthConfig(configRes.config)
          if (configRes.config.admin_enable_captcha) {
            loadCaptcha()
          }
        }
      } catch {
        if (!cancelled) {
          loadCaptcha()
        }
      }

      if (!cancelled) {
        setChecking(false)
      }
    }

    checkSetup()
    return () => {
      cancelled = true
    }
  }, [basePath, loadCaptcha, onNavigate])

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!username || !password) {
      toast.error('请输入用户名和密码')
      return
    }
    if (authConfig.admin_enable_captcha && !captchaCode) {
      toast.error('请输入验证码')
      return
    }

    const currentBasePath = basePath || getAdminBasePath()

    setLoading(true)
    const res = await apiPost<{ require_totp?: boolean; needs_setup?: boolean }>(`${currentBasePath}/login`, {
      username,
      password,
      ...(authConfig.admin_enable_captcha ? { captcha_id: captchaId, captcha_code: captchaCode } : {}),
      remember,
    })
    setLoading(false)

    if (res.needs_setup) {
      onNavigate('setup')
    } else if (res.require_totp) {
      onNavigate('totp')
    } else if (res.success) {
      toast.success('登录成功')
      setTimeout(() => onNavigate('dashboard'), 500)
    } else {
      toast.error(res.error || '登录失败')
      if (authConfig.admin_enable_captcha) {
        loadCaptcha()
      }
      setCaptchaCode('')
    }
  }

  if (checking) {
    return (
      <AuthCard panelKey="admin-login-loading">
        <div className="text-center text-dark-400">加载中...</div>
      </AuthCard>
    )
  }

  return (
    <AuthCard panelKey="admin-login">
      <div className="glass-card p-8">
        <div className="text-center mb-8">
          <div className="text-4xl mb-4">🔐</div>
          <h1 className="text-2xl font-bold text-dark-100">管理员登录</h1>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <Input
            label="用户名"
            placeholder="请输入用户名"
            value={username}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setUsername(e.target.value)}
            autoComplete="username"
          />

          <Input
            label="密码"
            type="password"
            placeholder="请输入密码"
            value={password}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
            autoComplete="current-password"
          />

          {authConfig.admin_enable_captcha && (
            <div className="space-y-1.5">
              <label className="block text-sm font-medium text-dark-300">验证码</label>
              <div className="flex items-center gap-3">
                <Input
                  placeholder="请输入验证码"
                  value={captchaCode}
                  onChange={(e: ChangeEvent<HTMLInputElement>) => setCaptchaCode(e.target.value)}
                />
                {captchaImage && (
                  <img
                    src={captchaImage}
                    alt="验证码"
                    onClick={loadCaptcha}
                    className="h-12 rounded-lg cursor-pointer hover:opacity-80 transition-opacity shrink-0"
                  />
                )}
              </div>
            </div>
          )}

          <Switch
            checked={remember}
            onChange={(checked) => setRemember(checked)}
            label="记住我"
            size="sm"
          />

          <Button type="submit" className="w-full" loading={loading}>
            登录
          </Button>
        </form>
      </div>
    </AuthCard>
  )
}

function AdminSetupPanel({ basePath, onNavigate }: AuthPanelProps) {
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()

    if (!password || !confirmPassword) {
      toast.error('请输入密码')
      return
    }

    if (password.length < 6) {
      toast.error('密码长度至少6位')
      return
    }

    if (password !== confirmPassword) {
      toast.error('两次输入的密码不一致')
      return
    }

    const currentBasePath = basePath || getAdminBasePath()

    setLoading(true)
    const res = await apiPost(`${currentBasePath}/setup`, {
      password,
      confirm_password: confirmPassword,
    })
    setLoading(false)

    if (res.success) {
      toast.success('密码设置成功')
      setTimeout(() => onNavigate('login'), 500)
    } else {
      toast.error(res.error || '设置失败')
    }
  }

  return (
    <AuthCard panelKey="admin-setup">
      <div className="glass-card p-8">
        <div className="text-center mb-8">
          <div className="text-4xl mb-4">🔧</div>
          <h1 className="text-2xl font-bold text-dark-100">初始化设置</h1>
          <p className="text-dark-400 mt-2">首次使用，请设置管理员密码</p>
        </div>

        <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-4 mb-6">
          <div className="flex items-start gap-3">
            <span className="text-yellow-500 text-xl">⚠️</span>
            <div className="text-sm text-yellow-200">
              <p className="font-medium mb-1">安全提示</p>
              <p className="text-yellow-300/80">
                请设置一个强密码，建议包含字母、数字和特殊字符。
                此密码将用于管理后台登录。
              </p>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <Input
            label="管理员密码"
            type="password"
            placeholder="请输入密码（至少6位）"
            value={password}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)}
            autoComplete="new-password"
          />

          <Input
            label="确认密码"
            type="password"
            placeholder="请再次输入密码"
            value={confirmPassword}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setConfirmPassword(e.target.value)}
            autoComplete="new-password"
          />

          <div className="text-sm text-dark-400 space-y-1">
            <p>• 默认用户名：<span className="text-dark-200 font-mono">admin</span></p>
            <p>• 密码设置后可在系统设置中修改</p>
          </div>

          <Button type="submit" className="w-full" loading={loading}>
            完成设置
          </Button>
        </form>
      </div>
    </AuthCard>
  )
}

function AdminTotpPanel({ basePath, onNavigate }: AuthPanelProps) {
  const [code, setCode] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    if (!code || code.length !== 6) {
      toast.error('请输入6位验证码')
      return
    }

    const currentBasePath = basePath || getAdminBasePath()

    setLoading(true)
    const res = await apiPost(`${currentBasePath}/totp`, { code })
    setLoading(false)

    if (res.success) {
      toast.success('验证成功')
      setTimeout(() => onNavigate('dashboard'), 500)
    } else {
      toast.error(res.error || '验证码错误')
      setCode('')
    }
  }

  return (
    <AuthCard panelKey="admin-totp">
      <div className="glass-card p-8">
        <div className="text-center mb-8">
          <div className="text-4xl mb-4">🔐</div>
          <h1 className="text-2xl font-bold text-dark-100">两步验证</h1>
          <p className="text-dark-400 mt-2">请输入验证器APP中的动态口令</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <input
              type="text"
              maxLength={6}
              placeholder="000000"
              value={code}
              onChange={(e: ChangeEvent<HTMLInputElement>) => setCode(e.target.value.replace(/\D/g, ''))}
              className="verify-code-input"
              autoFocus
            />
          </div>

          <Button type="submit" className="w-full" loading={loading}>
            验证
          </Button>

          <div className="text-center">
            <button
              type="button"
              onClick={() => onNavigate('login')}
              className="text-sm text-dark-400 hover:text-primary-400 transition-colors"
            >
              返回登录
            </button>
          </div>
        </form>
      </div>
    </AuthCard>
  )
}

export default function AdminAuthFlow({ initialStep }: AdminAuthFlowProps) {
  const [step, setStep] = useState<AdminAuthFlowState>(() => takeTransientStep(initialStep))
  const [basePath, setBasePath] = useState('')

  useEffect(() => {
    setBasePath(getAdminBasePath())
  }, [])

  const navigate = useCallback((nextStep: AdminAuthFlowState) => {
    const currentBasePath = getAdminBasePath()
    saveTransientStep(nextStep)
    setBasePath(currentBasePath)
    setStep(nextStep)

    const nextPath = nextStep === 'dashboard'
      ? buildAdminPath(currentBasePath)
      : buildAdminPath(currentBasePath, nextStep)
    window.history.pushState(null, '', nextPath)
  }, [])

  if (step === 'dashboard') {
    return (
      <AnimatePresence mode="wait" initial={false}>
        <motion.div
          key="admin-dashboard"
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -8 }}
          transition={{ duration: 0.2 }}
        >
          <AdminDashboardApp
            onAuthRequired={() => navigate('login')}
            onLogout={() => navigate('login')}
          />
        </motion.div>
      </AnimatePresence>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-dark-900 via-dark-800 to-dark-900 p-4">
      <AnimatePresence mode="wait" initial={false}>
        {step === 'login' && <AdminLoginPanel basePath={basePath} onNavigate={navigate} />}
        {step === 'setup' && <AdminSetupPanel basePath={basePath} onNavigate={navigate} />}
        {step === 'totp' && <AdminTotpPanel basePath={basePath} onNavigate={navigate} />}
      </AnimatePresence>
    </div>
  )
}
