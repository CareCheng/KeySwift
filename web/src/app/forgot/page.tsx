'use client'

import { useState, useEffect } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { AuthPageTransition } from '@/components/layout/AuthPageTransition'
import { Button, Input } from '@/components/ui'
import { apiPost } from '@/lib/api'
import { useI18n } from '@/hooks/useI18n'

/**
 * 找回密码页面
 */
export default function ForgotPasswordPage() {
  const router = useRouter()
  const { t } = useI18n()
  const [step, setStep] = useState(1)
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)

  // 表单数据
  const [username, setUsername] = useState('')
  const [userEmail, setUserEmail] = useState('')
  const [maskedEmail, setMaskedEmail] = useState('')
  const [has2FA, setHas2FA] = useState(false)
  const [resetToken, setResetToken] = useState('')
  const [emailCode, setEmailCode] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')

  // 倒计时
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown])

  // 步骤1: 查找用户
  const handleStep1 = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username) {
      toast.error(t('user.username'))
      return
    }

    setLoading(true)
    const res = await apiPost<{ email: string; masked_email: string; has_2fa: boolean }>(
      '/api/user/forgot/check',
      { username }
    )
    setLoading(false)

    if (res.success) {
      setUserEmail(res.email)
      setMaskedEmail(res.masked_email)
      setHas2FA(res.has_2fa)
      setStep(2)
    } else {
      toast.error(res.error || t('auth.userNotFound'))
    }
  }

  // 发送邮箱验证码
  const sendEmailCode = async () => {
    setSendingCode(true)
    const res = await apiPost('/api/user/email/send_code', {
      email: userEmail,
      code_type: 'reset_password',
    })
    setSendingCode(false)

    if (res.success) {
      toast.success(t('user.codeSent'))
      setCountdown(60)
    } else {
      toast.error(res.error || t('user.codeSendFailed'))
    }
  }

  // 步骤2: 验证身份
  const handleStep2 = async (e: React.FormEvent) => {
    e.preventDefault()

    if (has2FA) {
      if (!totpCode || totpCode.length !== 6) {
        toast.error(t('user.enter6DigitTotp'))
        return
      }
    } else {
      if (!emailCode) {
        toast.error(t('user.captchaPlaceholder'))
        return
      }
    }

    setLoading(true)
    const res = await apiPost<{ reset_token: string }>('/api/user/forgot/verify', {
      username,
      email_code: has2FA ? undefined : emailCode,
      totp_code: has2FA ? totpCode : undefined,
    })
    setLoading(false)

    if (res.success) {
      setResetToken(res.reset_token)
      toast.success(t('auth.verifySuccess'))
      setStep(3)
    } else {
      toast.error(res.error || t('auth.verifyFailed'))
    }
  }

  // 步骤3: 重置密码
  const handleStep3 = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!newPassword || newPassword.length < 6) {
      toast.error(t('user.passwordTooShort'))
      return
    }

    if (newPassword !== confirmPassword) {
      toast.error(t('user.passwordMismatch'))
      return
    }

    setLoading(true)
    const res = await apiPost('/api/user/forgot/reset', {
      username,
      reset_token: resetToken,
      new_password: newPassword,
    })
    setLoading(false)

    if (res.success) {
      toast.success(t('auth.resetSuccess'))
      setTimeout(() => router.push('/login/'), 1500)
    } else {
      toast.error(res.error || t('auth.resetFailed'))
    }
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
            <h1 className="text-2xl font-bold text-dark-100 mb-2">{t('auth.forgotTitle')}</h1>
            <p className="text-dark-400">
              {step === 1 && t('auth.forgotStep1')}
              {step === 2 && t('auth.forgotStep2')}
              {step === 3 && t('auth.forgotStep3')}
            </p>
          </div>

          {/* 步骤指示器 */}
          <div className="flex justify-center gap-3 mb-8">
            {[1, 2, 3].map((s) => (
              <div
                key={s}
                className={`step-dot ${s === step ? 'active' : ''} ${s < step ? 'done' : 'bg-dark-700 text-dark-500'}`}
              >
                {s < step ? <i className="fas fa-check text-xs" /> : s}
              </div>
            ))}
          </div>

          {/* 步骤1: 输入用户名 */}
          {step === 1 && (
            <form onSubmit={handleStep1} className="space-y-5">
              <Input
                label={t('user.username')}
                placeholder={t('user.usernamePlaceholder')}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                icon={<i className="fas fa-user" />}
              />
              <Button type="submit" className="w-full" loading={loading}>
                {t('common.next')}
              </Button>
            </form>
          )}

          {/* 步骤2: 验证身份 */}
          {step === 2 && (
            <form onSubmit={handleStep2} className="space-y-5">
              {has2FA ? (
                <div className="space-y-1.5">
                  <label className="block text-sm font-medium text-dark-300">{t('user.totpCode')}</label>
                  <input
                    type="text"
                    maxLength={6}
                    placeholder={t('user.totpPlaceholder')}
                    value={totpCode}
                    onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ''))}
                    className="verify-code-input"
                  />
                  <p className="text-dark-500 text-sm">{t('user.totpHint')}</p>
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
                  <p className="text-dark-500 text-sm">{t('user.codeSendTo')} {maskedEmail}</p>
                </div>
              )}

              <div className="flex flex-col sm:flex-row gap-3">
                <Button type="button" variant="secondary" className="flex-1 sm:flex-none" onClick={() => setStep(1)}>
                  {t('common.prev')}
                </Button>
                <Button type="submit" className="flex-1" loading={loading}>
                  {t('common.verify')}
                </Button>
              </div>
            </form>
          )}

          {/* 步骤3: 设置新密码 */}
          {step === 3 && (
            <form onSubmit={handleStep3} className="space-y-5">
              <Input
                label={t('user.newPassword')}
                type="password"
                placeholder={t('user.newPasswordPlaceholder')}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                icon={<i className="fas fa-lock" />}
              />
              <Input
                label={t('user.confirmNewPassword')}
                type="password"
                placeholder={t('user.confirmNewPasswordPlaceholder')}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                icon={<i className="fas fa-lock" />}
              />
              <Button type="submit" className="w-full" loading={loading}>
                {t('user.resetPassword')}
              </Button>
            </form>
          )}

          <div className="mt-6 flex justify-between text-sm">
            <Link href="/login/" className="text-primary-400 hover:text-primary-300 transition-colors">
              {t('auth.backToLogin')}
            </Link>
            <Link href="/register/" className="text-dark-400 hover:text-dark-300 transition-colors">
              {t('auth.registerNewAccount')}
            </Link>
          </div>
        </div>
      </AuthPageTransition>
    </div>
  )
}
