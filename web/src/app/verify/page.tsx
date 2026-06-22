'use client'

import { useState, useEffect, Suspense } from 'react'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import toast from 'react-hot-toast'
import { AuthPageTransition } from '@/components/layout/AuthPageTransition'
import { Button, Input } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { useI18n } from '@/hooks/useI18n'
import { buildUserRouteUrl } from '@/lib/userNavigation'

/**
 * 二次验证信息接口
 */
interface VerifyInfo {
  username: string
  email: string
  masked_email: string
  has_totp: boolean
  prefer_email: boolean
}

/**
 * 二次验证页面内容
 */
function VerifyContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const token = searchParams.get('token') || ''
  const { t } = useI18n()

  const [loading, setLoading] = useState(true)
  const [verifying, setVerifying] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [verifyInfo, setVerifyInfo] = useState<VerifyInfo | null>(null)
  const [totpCode, setTotpCode] = useState('')
  const [emailCode, setEmailCode] = useState('')

  // 加载验证信息
  useEffect(() => {
    if (!token) {
      toast.error(t('auth.invalidRequest'))
      setTimeout(() => router.push('/login/'), 2000)
      return
    }

    const loadVerifyInfo = async () => {
      const res = await apiGet<{ username: string; email: string; masked_email: string; has_totp: boolean; prefer_email: boolean }>(`/api/user/2fa/info?token=${token}`)
      if (res.success) {
        setVerifyInfo({
          username: res.username,
          email: res.email,
          masked_email: res.masked_email,
          has_totp: res.has_totp,
          prefer_email: res.prefer_email,
        })
      } else {
        toast.error(res.error || t('auth.infoExpired'))
        setTimeout(() => router.push('/login/'), 2000)
      }
      setLoading(false)
    }
    loadVerifyInfo()
  }, [token, t, router])

  // 倒计时
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 判断使用哪种验证方式
  const useTOTP = verifyInfo?.has_totp && !verifyInfo?.prefer_email

  // 发送邮箱验证码
  const sendEmailCode = async () => {
    if (!verifyInfo?.email) return

    setSendingCode(true)
    const res = await apiPost('/api/user/email/send_code', {
      email: verifyInfo.email,
      code_type: 'login',
    })
    setSendingCode(false)

    if (res.success) {
      toast.success(t('user.codeSent'))
      setCountdown(60)
    } else {
      toast.error(res.error || t('user.codeSendFailed'))
    }
  }

  // 提交验证
  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault()

    if (useTOTP) {
      if (!totpCode || totpCode.length !== 6) {
        toast.error(t('user.enter6DigitCode'))
        return
      }
    } else {
      if (!emailCode) {
        toast.error(t('user.captchaPlaceholder'))
        return
      }
    }

    setVerifying(true)
    const res = await apiPost('/api/user/2fa/verify_login', {
      token,
      totp_code: useTOTP ? totpCode : undefined,
      email_code: !useTOTP ? emailCode : undefined,
    })
    setVerifying(false)

    if (res.success) {
      toast.success(t('auth.verifySuccess'))
      setTimeout(() => router.push(buildUserRouteUrl('products')), 1000)
    } else {
      toast.error(res.error || t('auth.verifyFailed'))
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400 mb-4" />
          <p className="text-dark-400">{t('common.loading')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      {/* 背景装饰 */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-80 h-80 bg-primary-500/20 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-purple-500/20 rounded-full blur-3xl" />
      </div>

      <AuthPageTransition mode="login">
        <div className="card p-8">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-bold text-dark-100 mb-2">{t('auth.verifyTitle')}</h1>
            <p className="text-dark-400">{t('auth.verifySubtitle')}</p>
          </div>

          {/* 用户信息 */}
          {verifyInfo && (
            <div className="bg-dark-700/30 rounded-xl p-4 mb-6 text-center">
              <p className="text-dark-100 font-medium">{verifyInfo.username}</p>
              <p className="text-dark-500 text-sm">{verifyInfo.masked_email}</p>
            </div>
          )}

          {/* 验证方式 */}
          <div className="bg-primary-500/10 rounded-xl p-4 mb-6 text-center">
            <div className="text-3xl mb-2">{useTOTP ? '🔐' : '📧'}</div>
            <p className="text-dark-200 font-medium">
              {useTOTP ? t('user.totpVerify') : t('user.emailVerify')}
            </p>
          </div>

          <form onSubmit={handleVerify} className="space-y-5">
            {useTOTP ? (
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-dark-300">{t('user.totpCode')}</label>
                <input
                  type="text"
                  maxLength={6}
                  placeholder={t('user.totpPlaceholder')}
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ''))}
                  className="verify-code-input"
                  autoFocus
                />
                <p className="text-dark-500 text-sm">{t('user.openAuthApp')}</p>
              </div>
            ) : (
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-dark-300">{t('user.emailCode')}</label>
                <div className="flex items-center gap-3">
                  <Input
                    placeholder={t('user.captchaPlaceholder')}
                    value={emailCode}
                    onChange={(e) => setEmailCode(e.target.value)}
                  />
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={sendEmailCode}
                    disabled={countdown > 0 || sendingCode}
                  >
                    {countdown > 0 ? `${countdown}${t('common.seconds')}` : sendingCode ? t('common.sending') : t('user.sendCode')}
                  </Button>
                </div>
                <p className="text-dark-500 text-sm">
                  {t('user.codeSendTo')} {verifyInfo?.masked_email}
                </p>
              </div>
            )}

            <Button type="submit" className="w-full" loading={verifying}>
              {t('user.verifyLogin')}
            </Button>
          </form>

          <div className="mt-6 text-center">
            <Link href="/login/" className="text-dark-400 hover:text-dark-300 transition-colors text-sm">
              {t('auth.backToLogin')}
            </Link>
          </div>
        </div>
      </AuthPageTransition>
    </div>
  )
}

/**
 * 二次验证页面
 */
export default function VerifyPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center">
          <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
        </div>
      }
    >
      <VerifyContent />
    </Suspense>
  )
}
