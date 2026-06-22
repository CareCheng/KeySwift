'use client'

import { useCallback, useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Modal } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'

interface BalanceInfo {
  balance: number
  frozen: number
  total_in: number
  total_out: number
}

interface BalanceLog {
  id: number
  type: string
  amount: number
  before_balance: number
  after_balance: number
  order_no: string
  remark: string
  created_at: string
}

interface PayPasswordStatus {
  is_set: boolean
  is_locked: boolean
  lock_remaining_seconds: number
}

export function WalletTab() {
  const [balanceInfo, setBalanceInfo] = useState<BalanceInfo | null>(null)
  const [balanceLogs, setBalanceLogs] = useState<BalanceLog[]>([])
  const [payPasswordStatus, setPayPasswordStatus] = useState<PayPasswordStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [showSetPayPasswordModal, setShowSetPayPasswordModal] = useState(false)
  const [showUpdatePayPasswordModal, setShowUpdatePayPasswordModal] = useState(false)
  const [showResetPayPasswordModal, setShowResetPayPasswordModal] = useState(false)
  const [payPassword, setPayPassword] = useState('')
  const [confirmPayPassword, setConfirmPayPassword] = useState('')
  const [loginPassword, setLoginPassword] = useState('')
  const [oldPayPassword, setOldPayPassword] = useState('')
  const [emailCode, setEmailCode] = useState('')
  const [sendingCode, setSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)

  const loadBalance = useCallback(async () => {
    const res = await apiGet<{ data: BalanceInfo }>('/api/user/balance')
    if (res.success && res.data) {
      setBalanceInfo(res.data)
    }
  }, [])

  const loadBalanceLogs = useCallback(async () => {
    const res = await apiGet<{ data: BalanceLog[] }>('/api/user/balance/logs')
    if (res.success && res.data) {
      setBalanceLogs(res.data)
    }
  }, [])

  const loadPayPasswordStatus = useCallback(async () => {
    const res = await apiGet<{ data: PayPasswordStatus }>('/api/user/pay-password/status')
    if (res.success && res.data) {
      setPayPasswordStatus(res.data)
    }
  }, [])

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      await Promise.all([loadBalance(), loadBalanceLogs(), loadPayPasswordStatus()])
      setLoading(false)
    }
    loadData()
  }, [loadBalance, loadBalanceLogs, loadPayPasswordStatus])

  const resetPasswordForm = () => {
    setPayPassword('')
    setConfirmPayPassword('')
    setLoginPassword('')
    setOldPayPassword('')
    setEmailCode('')
  }

  const validatePayPassword = () => {
    if (!/^\d{6}$/.test(payPassword)) {
      toast.error('支付密码必须为6位纯数字')
      return false
    }
    if (payPassword !== confirmPayPassword) {
      toast.error('两次输入的密码不一致')
      return false
    }
    return true
  }

  const handleSetPayPassword = async () => {
    if (!validatePayPassword()) return
    if (!loginPassword) {
      toast.error('请输入登录密码')
      return
    }
    const res = await apiPost('/api/user/pay-password/set', {
      password: payPassword,
      login_password: loginPassword,
    })
    if (res.success) {
      toast.success('支付密码设置成功')
      setShowSetPayPasswordModal(false)
      resetPasswordForm()
      loadPayPasswordStatus()
    } else {
      toast.error(res.error || '设置失败')
    }
  }

  const handleUpdatePayPassword = async () => {
    if (!validatePayPassword()) return
    if (!oldPayPassword) {
      toast.error('请输入原支付密码')
      return
    }
    const res = await apiPost('/api/user/pay-password/update', {
      old_password: oldPayPassword,
      new_password: payPassword,
    })
    if (res.success) {
      toast.success('支付密码修改成功')
      setShowUpdatePayPasswordModal(false)
      resetPasswordForm()
      loadPayPasswordStatus()
    } else {
      toast.error(res.error || '修改失败')
    }
  }

  const handleSendResetCode = async () => {
    setSendingCode(true)
    const res = await apiPost('/api/user/pay-password/send-reset-code', {})
    setSendingCode(false)
    if (res.success) {
      toast.success('验证码已发送到您的邮箱')
      setCountdown(60)
      const timer = window.setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            window.clearInterval(timer)
            return 0
          }
          return prev - 1
        })
      }, 1000)
    } else {
      toast.error(res.error || '发送失败')
    }
  }

  const handleResetPayPassword = async () => {
    if (!validatePayPassword()) return
    if (!emailCode) {
      toast.error('请输入邮箱验证码')
      return
    }
    const res = await apiPost('/api/user/pay-password/reset', {
      new_password: payPassword,
      email_code: emailCode,
    })
    if (res.success) {
      toast.success('支付密码重置成功')
      setShowResetPayPasswordModal(false)
      resetPasswordForm()
      loadPayPasswordStatus()
    } else {
      toast.error(res.error || '重置失败')
    }
  }

  const getBalanceTypeText = (type: string) => {
    const types: Record<string, string> = {
      consume: '消费',
      refund: '退款',
      withdraw: '提现',
      freeze: '冻结',
      unfreeze: '解冻',
      adjust: '调整',
    }
    return types[type] || type
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {payPasswordStatus && !payPasswordStatus.is_set && (
        <div className="bg-amber-500/10 border border-amber-500/30 rounded-xl p-4 flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <i className="fas fa-exclamation-triangle text-amber-400" />
            <span className="text-amber-200">请先设置支付密码才能使用余额支付</span>
          </div>
          <Button size="sm" onClick={() => setShowSetPayPasswordModal(true)}>
            立即设置
          </Button>
        </div>
      )}

      {payPasswordStatus && payPasswordStatus.is_set && (
        <Card className="bg-dark-800/50">
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center">
                <i className="fas fa-shield-alt text-green-400" />
              </div>
              <div>
                <div className="text-dark-100 font-medium">支付密码</div>
                <div className="text-sm text-dark-400">
                  {payPasswordStatus.is_locked
                    ? `已锁定，${Math.ceil(payPasswordStatus.lock_remaining_seconds / 60)}分钟后解锁`
                    : '已设置'}
                </div>
              </div>
            </div>
            <div className="flex gap-2">
              <Button size="sm" variant="secondary" onClick={() => setShowUpdatePayPasswordModal(true)}>
                修改
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setShowResetPayPasswordModal(true)}>
                忘记密码
              </Button>
            </div>
          </div>
        </Card>
      )}

      {balanceInfo && (
        <div className="bg-gradient-to-r from-primary-600 to-primary-500 rounded-2xl p-6 text-white">
          <div className="text-white/80 mb-4">可用余额</div>
          <div className="text-4xl font-bold mb-4">¥{balanceInfo.balance.toFixed(2)}</div>
          <div className="grid grid-cols-3 gap-4 text-sm">
            <div>
              <div className="text-white/60">冻结金额</div>
              <div className="font-medium">¥{balanceInfo.frozen.toFixed(2)}</div>
            </div>
            <div>
              <div className="text-white/60">累计入账</div>
              <div className="font-medium">¥{balanceInfo.total_in.toFixed(2)}</div>
            </div>
            <div>
              <div className="text-white/60">累计支出</div>
              <div className="font-medium">¥{balanceInfo.total_out.toFixed(2)}</div>
            </div>
          </div>
        </div>
      )}

      <Card title="余额明细" icon={<i className="fas fa-list" />}>
        {balanceLogs.length === 0 ? (
          <div className="text-center py-8 text-dark-400">暂无记录</div>
        ) : (
          <div className="space-y-3">
            {balanceLogs.map((log) => (
              <div key={log.id} className="flex items-center justify-between p-3 bg-dark-700/30 rounded-lg">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="text-dark-100">{getBalanceTypeText(log.type)}</span>
                    {log.remark && <span className="text-dark-500 text-sm">- {log.remark}</span>}
                  </div>
                  <div className="text-sm text-dark-500">{formatDateTime(log.created_at)}</div>
                </div>
                <div className={`font-medium ${log.amount >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                  {log.amount >= 0 ? '+' : ''}
                  {log.amount.toFixed(2)}
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>

      <Modal isOpen={showSetPayPasswordModal} onClose={() => { setShowSetPayPasswordModal(false); resetPasswordForm() }} title="设置支付密码" size="sm">
        <div className="space-y-4">
          <Input label="支付密码" type="password" placeholder="请输入6位数字" maxLength={6} value={payPassword} onChange={(e) => setPayPassword(e.target.value.replace(/\D/g, ''))} />
          <Input label="确认支付密码" type="password" placeholder="请再次输入" maxLength={6} value={confirmPayPassword} onChange={(e) => setConfirmPayPassword(e.target.value.replace(/\D/g, ''))} />
          <Input label="登录密码" type="password" placeholder="请输入登录密码验证身份" value={loginPassword} onChange={(e) => setLoginPassword(e.target.value)} />
          <Button className="w-full" onClick={handleSetPayPassword}>确认设置</Button>
        </div>
      </Modal>

      <Modal isOpen={showUpdatePayPasswordModal} onClose={() => { setShowUpdatePayPasswordModal(false); resetPasswordForm() }} title="修改支付密码" size="sm">
        <div className="space-y-4">
          <Input label="原支付密码" type="password" placeholder="请输入原支付密码" maxLength={6} value={oldPayPassword} onChange={(e) => setOldPayPassword(e.target.value.replace(/\D/g, ''))} />
          <Input label="新支付密码" type="password" placeholder="请输入6位数字" maxLength={6} value={payPassword} onChange={(e) => setPayPassword(e.target.value.replace(/\D/g, ''))} />
          <Input label="确认新密码" type="password" placeholder="请再次输入" maxLength={6} value={confirmPayPassword} onChange={(e) => setConfirmPayPassword(e.target.value.replace(/\D/g, ''))} />
          <Button className="w-full" onClick={handleUpdatePayPassword}>确认修改</Button>
        </div>
      </Modal>

      <Modal isOpen={showResetPayPasswordModal} onClose={() => { setShowResetPayPasswordModal(false); resetPasswordForm() }} title="重置支付密码" size="sm">
        <div className="space-y-4">
          <div className="flex gap-2">
            <Input label="邮箱验证码" placeholder="请输入验证码" value={emailCode} onChange={(e) => setEmailCode(e.target.value)} className="flex-1" />
            <Button variant="secondary" className="mt-6" onClick={handleSendResetCode} disabled={sendingCode || countdown > 0}>
              {countdown > 0 ? `${countdown}s` : '发送验证码'}
            </Button>
          </div>
          <Input label="新支付密码" type="password" placeholder="请输入6位数字" maxLength={6} value={payPassword} onChange={(e) => setPayPassword(e.target.value.replace(/\D/g, ''))} />
          <Input label="确认新密码" type="password" placeholder="请再次输入" maxLength={6} value={confirmPayPassword} onChange={(e) => setConfirmPayPassword(e.target.value.replace(/\D/g, ''))} />
          <Button className="w-full" onClick={handleResetPayPassword}>确认重置</Button>
        </div>
      </Modal>
    </div>
  )
}
