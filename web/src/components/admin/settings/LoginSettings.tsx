'use client'

import { useState } from 'react'
import { Button, Card, Input, Modal, Switch } from '@/components/ui'
import toast from 'react-hot-toast'
import { apiGet, apiPost } from '@/lib/api'
import { SettingsState } from './useSettingsState'

type SecurityForm = SettingsState['securityForm']

const timeoutPresets = [30, 60, 120, 480, 1440]

function clampTimeout(value: number, fallback: number) {
  if (!Number.isFinite(value) || value <= 0) return fallback
  if (value < 5) return 5
  if (value > 1440) return 1440
  return value
}

function formatPreset(value: number) {
  return value >= 60 ? `${value / 60}小时` : `${value}分钟`
}

function SettingRow({
  children,
  action,
}: {
  children: React.ReactNode
  action?: React.ReactNode
}) {
  return (
    <div className="flex flex-col gap-3 rounded-lg bg-dark-900/30 p-3 md:flex-row md:items-center md:justify-between">
      <div className="min-w-0">{children}</div>
      {action && <div className="flex shrink-0 items-center gap-2">{action}</div>}
    </div>
  )
}

function TimeoutEditor({
  value,
  onChange,
}: {
  value: number
  onChange: (value: number) => void
}) {
  return (
    <div className="rounded-lg bg-dark-900/30 p-3">
      <div className="mb-2 flex items-center gap-3">
        <i className="fas fa-hourglass-half text-dark-400" />
        <div>
          <div className="text-sm font-medium text-dark-200">超时时间</div>
          <p className="mt-0.5 text-xs text-dark-500">登录后超过此时间将要求重新登录，范围 5-1440 分钟</p>
        </div>
      </div>
      <div className="mt-2 flex flex-col gap-3 xl:flex-row xl:items-center">
        <div className="flex items-center gap-3">
          <Input
            type="number"
            min={5}
            max={1440}
            value={value}
            onChange={(event) => onChange(clampTimeout(parseInt(event.target.value), 60))}
            className="w-32"
          />
          <span className="text-sm text-dark-400">分钟</span>
        </div>
        <div className="flex flex-wrap gap-1.5">
          {timeoutPresets.map((preset) => (
            <button
              key={preset}
              type="button"
              onClick={() => onChange(preset)}
              className={`rounded px-2 py-1 text-xs transition-colors ${
                value === preset
                  ? 'bg-primary-500 text-white'
                  : 'bg-dark-800 text-dark-400 hover:bg-dark-700 hover:text-dark-200'
              }`}
            >
              {formatPreset(preset)}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

/**
 * 登录设置子页面。
 * 职责：统一管理后台登录验证与用户侧登录注册策略。
 */
export function LoginSettings({ state }: { state: SettingsState }) {
  const { securityForm, setSecurityForm, saveSecurity, basicForm } = state
  const [showAdminAccountModal, setShowAdminAccountModal] = useState(false)
  const [showAdminTotpModal, setShowAdminTotpModal] = useState(false)
  const [showUserAdvancedModal, setShowUserAdvancedModal] = useState(false)
  const [modalUsername, setModalUsername] = useState('')
  const [modalPassword, setModalPassword] = useState('')
  const [totpTestCode, setTotpTestCode] = useState('')
  const [totpTestResult, setTotpTestResult] = useState<boolean | null>(null)
  const [emailEnabled, setEmailEnabled] = useState(false)

  const updateForm = (patch: Partial<SecurityForm>) => {
    setSecurityForm((prev) => ({ ...prev, ...patch }))
  }

  const handleAdminLoginToggle = (checked: boolean) => {
    if (checked) {
      updateForm({ enable_login: true, enable_captcha: true, enable_session_timeout: true })
      return
    }
    updateForm({
      enable_login: false,
      enable_captcha: false,
      enable_2fa: false,
      enable_session_timeout: false,
    })
  }

  const handleAdminTotpToggle = (checked: boolean) => {
    updateForm({ enable_2fa: checked })
    if (checked && !securityForm.totp_secret) {
      setShowAdminTotpModal(true)
    }
  }

  const handleUserRegisterToggle = (checked: boolean) => {
    if (checked) {
      updateForm({ user_allow_register: true, user_enable_captcha: true, user_enable_session_timeout: true })
      return
    }
    updateForm({
      user_allow_register: false,
      user_require_email_verification: false,
    })
  }

  const handleSave = async () => {
    if (await saveSecurity()) {
      toast.success('登录设置已保存')
      return
    }
    toast.error('保存失败')
  }

  const openAdminAccountModal = () => {
    setModalUsername(securityForm.admin_username)
    setModalPassword('')
    setShowAdminAccountModal(true)
  }

  const saveAdminAccountModal = () => {
    if (!modalUsername.trim()) {
      toast.error('请输入管理员用户名')
      return
    }
    updateForm({
      admin_username: modalUsername.trim(),
      admin_password: modalPassword,
    })
    setShowAdminAccountModal(false)
    toast.success('账户配置已写入待保存表单')
  }

  const generateTotp = async () => {
    const res = await apiPost<{ secret: string }>('/api/admin/2fa/generate', {})
    if (res.success && res.secret) {
      updateForm({ totp_secret: res.secret })
      setTotpTestCode('')
      setTotpTestResult(null)
      toast.success('新密钥已生成，请扫描二维码并验证')
      return
    }
    toast.error(res.error || '生成失败')
  }

  const openUserAdvancedModal = async () => {
    const res = await apiGet<{ config: { enabled: boolean } }>('/api/admin/email/config')
    const enabled = Boolean(res.success && res.config?.enabled)
    setEmailEnabled(enabled)
    if (!enabled && securityForm.user_require_email_verification) {
      updateForm({ user_require_email_verification: false })
    }
    setShowUserAdvancedModal(true)
  }

  const testTotp = async () => {
    if (!totpTestCode || totpTestCode.length !== 6) {
      toast.error('请输入6位验证码')
      return
    }
    const res = await apiPost('/api/admin/2fa/verify', { code: totpTestCode, secret: securityForm.totp_secret })
    setTotpTestResult(res.success)
  }

  const getTotpQrUrl = () => {
    const title = encodeURIComponent(basicForm.system_title || '卡密购买系统')
    const user = encodeURIComponent(securityForm.admin_username || 'admin')
    const uri = `otpauth://totp/${title}:${user}?secret=${securityForm.totp_secret}&issuer=${title}`
    return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(uri)}`
  }

  return (
    <div className="space-y-6">
      <Card title="登录设置">
        <div className="grid gap-6 xl:grid-cols-2">
          <div className="space-y-3">
            <div>
              <h3 className="text-base font-semibold text-dark-100">后台管理登录</h3>
              <p className="mt-1 text-sm text-dark-500">控制管理后台入口的登录验证、图形验证码、二步验证和会话有效期。</p>
            </div>

            <SettingRow
              action={
                securityForm.enable_login ? (
                  <Button variant="primary" size="sm" onClick={openAdminAccountModal}>
                    <i className="fas fa-user-cog" /> 账户配置
                  </Button>
                ) : null
              }
            >
              <Switch
                checked={securityForm.enable_login}
                onChange={handleAdminLoginToggle}
                label="启用登录验证"
                description="关闭后可直接访问管理后台，不建议在公网环境关闭"
              />
            </SettingRow>

            {securityForm.enable_login ? (
              <div className="ml-4 space-y-3 border-l-2 border-primary-500/30 pl-4">
                <SettingRow>
                  <Switch
                    checked={securityForm.enable_captcha}
                    onChange={(checked) => updateForm({ enable_captcha: checked })}
                    label="登录图形验证码"
                    description="登录时必须输入图形验证码，降低暴力破解风险"
                  />
                </SettingRow>

                <SettingRow
                  action={
                    securityForm.enable_2fa ? (
                      <Button variant="primary" size="sm" onClick={() => setShowAdminTotpModal(true)}>
                        <i className="fas fa-shield-alt" /> 配置
                      </Button>
                    ) : null
                  }
                >
                  <Switch
                    checked={securityForm.enable_2fa}
                    onChange={handleAdminTotpToggle}
                    label="二步验证 (TOTP)"
                    description="登录后需要输入验证器动态口令"
                  />
                </SettingRow>

                <SettingRow>
                  <Switch
                    checked={securityForm.enable_session_timeout}
                    onChange={(checked) => updateForm({ enable_session_timeout: checked })}
                    label="会话超时"
                    description="关闭后后台登录会话保持长期有效"
                  />
                </SettingRow>

                {securityForm.enable_session_timeout && (
                  <TimeoutEditor
                    value={securityForm.session_timeout}
                    onChange={(value) => updateForm({ session_timeout: value })}
                  />
                )}
              </div>
            ) : (
              <div className="rounded-lg border border-amber-500/20 bg-amber-500/10 p-3 text-sm text-amber-300">
                <i className="fas fa-exclamation-triangle mr-2" />
                登录验证已关闭，管理后台入口不会要求账号密码。
              </div>
            )}
          </div>

          <div className="space-y-3">
            <div>
              <h3 className="text-base font-semibold text-dark-100">用户侧登录与注册</h3>
              <p className="mt-1 text-sm text-dark-500">主页面只保留注册、图形验证码和 TOTP 开关，更多策略通过高级配置维护。</p>
            </div>

            <SettingRow
              action={
                <Button variant="primary" size="sm" onClick={openUserAdvancedModal}>
                  <i className="fas fa-sliders-h" /> 高级
                </Button>
              }
            >
              <Switch
                checked={securityForm.user_allow_register}
                onChange={handleUserRegisterToggle}
                label="允许注册"
                description="关闭后仅已有用户可以登录"
              />
            </SettingRow>

            <SettingRow>
              <Switch
                checked={securityForm.user_enable_captcha}
                onChange={(checked) => updateForm({ user_enable_captcha: checked })}
                label="图形验证码"
                description="用户登录和注册时要求输入图形验证码"
              />
            </SettingRow>

            <SettingRow>
              <Switch
                checked={securityForm.user_enable_2fa}
                onChange={(checked) => updateForm({ user_enable_2fa: checked })}
                label="TOTP / 二步验证"
                description="允许用户启用动态口令或邮箱二步验证"
              />
            </SettingRow>

            {!securityForm.user_allow_register && (
              <div className="rounded-lg border border-amber-500/20 bg-amber-500/10 p-3 text-sm text-amber-300">
                <i className="fas fa-info-circle mr-2" />
                新用户注册入口将隐藏，注册接口也会拒绝新注册请求。
              </div>
            )}
          </div>
        </div>
      </Card>

      <Card>
        <div className="flex justify-end">
          <Button variant="success" onClick={handleSave}>
            <i className="fas fa-save" /> 保存登录设置
          </Button>
        </div>
      </Card>

      <Modal
        isOpen={showAdminAccountModal}
        onClose={() => setShowAdminAccountModal(false)}
        title="后台账户配置"
        size="md"
      >
        <div className="space-y-4">
          <Input
            label="管理员用户名"
            value={modalUsername}
            onChange={(event) => setModalUsername(event.target.value)}
            icon={<i className="fas fa-user" />}
          />
          <Input
            label="管理员密码"
            type="password"
            value={modalPassword}
            onChange={(event) => setModalPassword(event.target.value)}
            placeholder="留空保持不变"
            icon={<i className="fas fa-lock" />}
          />
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={() => setShowAdminAccountModal(false)}>取消</Button>
            <Button variant="success" onClick={saveAdminAccountModal}>
              <i className="fas fa-save" /> 写入表单
            </Button>
          </div>
        </div>
      </Modal>

      <Modal
        isOpen={showAdminTotpModal}
        onClose={() => setShowAdminTotpModal(false)}
        title="配置后台二步验证"
        size="md"
      >
        <div className="space-y-4">
          {securityForm.totp_secret ? (
            <>
              <div className="rounded-lg bg-dark-900/30 p-3">
                <p className="mb-2 text-sm font-medium text-dark-200">
                  <i className="fas fa-qrcode mr-1.5 text-primary-400" />
                  步骤 1：使用身份验证器扫描二维码
                </p>
                <div className="inline-block rounded-lg bg-white p-2">
                  <img src={getTotpQrUrl()} alt="TOTP QR Code" className="h-48 w-48" />
                </div>
              </div>

              <div className="rounded-lg bg-dark-900/30 p-3">
                <p className="mb-2 text-sm font-medium text-dark-200">
                  <i className="fas fa-key mr-1.5 text-primary-400" />
                  步骤 2：或手动输入密钥
                </p>
                <code className="block rounded bg-dark-800 px-3 py-2 font-mono text-sm text-primary-300">
                  {securityForm.totp_secret}
                </code>
              </div>

              <div className="rounded-lg bg-dark-900/30 p-3">
                <p className="mb-2 text-sm font-medium text-dark-200">
                  <i className="fas fa-check-circle mr-1.5 text-primary-400" />
                  步骤 3：输入验证码测试
                </p>
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
                  <Input
                    value={totpTestCode}
                    onChange={(event) => {
                      setTotpTestCode(event.target.value.replace(/\D/g, '').slice(0, 6))
                      setTotpTestResult(null)
                    }}
                    placeholder="输入6位验证码"
                    maxLength={6}
                    className="w-40"
                  />
                  <Button variant="secondary" onClick={testTotp} disabled={totpTestCode.length !== 6}>验证</Button>
                  {totpTestResult !== null && (
                    <span className={totpTestResult ? 'text-green-400' : 'text-red-400'}>
                      {totpTestResult ? '验证通过' : '验证失败'}
                    </span>
                  )}
                </div>
              </div>
            </>
          ) : (
            <div className="py-6 text-center">
              <p className="mb-4 text-dark-400">尚未生成 TOTP 密钥</p>
            </div>
          )}

          <div className="flex justify-between gap-3 border-t border-dark-700/50 pt-4">
            <Button variant="secondary" onClick={generateTotp}>
              <i className="fas fa-sync-alt" /> {securityForm.totp_secret ? '重新生成密钥' : '生成密钥'}
            </Button>
            <Button variant="success" onClick={() => setShowAdminTotpModal(false)}>
              完成
            </Button>
          </div>
        </div>
      </Modal>

      <Modal
        isOpen={showUserAdvancedModal}
        onClose={() => setShowUserAdvancedModal(false)}
        title="用户侧高级配置"
        size="lg"
      >
        <div className="space-y-4">
          <SettingRow>
            <Switch
              checked={securityForm.user_require_email_verification}
              onChange={(checked) => updateForm({ user_require_email_verification: checked })}
              disabled={!securityForm.user_allow_register || !emailEnabled}
              label="注册需要验证邮箱"
              description={emailEnabled ? '开启后注册时必须输入邮箱验证码' : '邮箱服务未启用，暂不能开启注册邮箱验证'}
            />
          </SettingRow>

          <SettingRow>
            <Switch
              checked={securityForm.user_enable_session_timeout}
              onChange={(checked) => updateForm({ user_enable_session_timeout: checked })}
              label="用户会话超时"
              description="关闭后用户登录会话保持长期有效"
            />
          </SettingRow>

          {securityForm.user_enable_session_timeout && (
            <TimeoutEditor
              value={securityForm.user_session_timeout}
              onChange={(value) => updateForm({ user_session_timeout: value })}
            />
          )}

          <div className="rounded-lg bg-dark-900/30 p-3 text-sm text-dark-400">
            <div className="mb-1 font-medium text-dark-200">当前策略摘要</div>
            <p>
              注册：{securityForm.user_allow_register ? '开放' : '关闭'}；
              图形验证码：{securityForm.user_enable_captcha ? '启用' : '关闭'}；
              TOTP：{securityForm.user_enable_2fa ? '允许用户启用' : '禁止用户启用'}。
            </p>
          </div>

          <div className="flex justify-end gap-3 border-t border-dark-700/50 pt-4">
            <Button variant="secondary" onClick={() => setShowUserAdvancedModal(false)}>关闭</Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
