'use client'

import { Suspense, useCallback, useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import toast from 'react-hot-toast'
import { Button, Card, Input } from '@/components/ui'
import { UserShell } from '@/components/layout/UserShell'
import { apiGet, apiPost } from '@/lib/api'
import { getCachedOrderDetail, getOrderDetail, updateCachedOrderDetail } from '@/lib/orderData'
import { formatDateTime, formatMoney } from '@/lib/utils'
import { useUserNavigation } from '@/lib/userNavigation'
import type { OrderDetail } from '@/types/order'

interface PaymentMethods {
  balance?: { enabled: boolean; builtin?: boolean; name?: string }
}

interface PaymentViewProps {
  orderNo?: string | null
}

/**
 * 支付内容视图。
 * 支持独立路由和用户端主入口 hash 切换复用。
 */
export function PaymentView({ orderNo }: PaymentViewProps) {
  const router = useRouter()
  const navigateUser = useUserNavigation()

  const [order, setOrder] = useState<OrderDetail | null>(() => getCachedOrderDetail(orderNo))
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethods | null>(null)
  const [loading, setLoading] = useState(true)
  const [paying, setPaying] = useState(false)
  const [selectedMethod, setSelectedMethod] = useState('balance')
  const [countdown, setCountdown] = useState(0)
  const [userBalance, setUserBalance] = useState(0)
  const [payPasswordSet, setPayPasswordSet] = useState(false)
  const [balancePayPassword, setBalancePayPassword] = useState('')
  const [showPayPasswordInput, setShowPayPasswordInput] = useState(false)

  const loadOrder = useCallback(async () => {
    if (!orderNo) {
      toast.error('订单号无效')
      navigateUser('products')
      return false
    }

    const orderRes = await getOrderDetail(orderNo)
    if (!orderRes.order) {
      if (orderRes.error === '请先登录') {
        router.push('/login')
      } else {
        toast.error(orderRes.error || '订单不存在')
        navigateUser('products')
      }
      return false
    }

    if (orderRes.order.status !== 0) {
      toast.error('订单状态异常，无法支付')
      navigateUser('user')
      return false
    }

    setOrder(orderRes.order)
    const createdAt = new Date(orderRes.order.created_at).getTime()
    const expireAt = createdAt + 30 * 60 * 1000
    const remaining = Math.floor((expireAt - Date.now()) / 1000)
    setCountdown(remaining > 0 ? remaining : 0)
    return true
  }, [orderNo, router, navigateUser])

  useEffect(() => {
    const loadData = async () => {
      const success = await loadOrder()
      if (!success) return

      const [paymentRes, balanceRes, payPwdRes] = await Promise.all([
        apiGet<{ methods: PaymentMethods }>('/api/payment/methods'),
        apiGet<{ data: { balance: number } }>('/api/user/balance'),
        apiGet<{ data: { is_set: boolean } }>('/api/user/pay-password/status'),
      ])

      if (paymentRes.success && paymentRes.methods) {
        setPaymentMethods(paymentRes.methods)
      }
      if (balanceRes.success && balanceRes.data) {
        setUserBalance(balanceRes.data.balance || 0)
      }
      if (payPwdRes.success && payPwdRes.data) {
        setPayPasswordSet(payPwdRes.data.is_set)
      }
      setLoading(false)
    }

    loadData()
  }, [loadOrder])

  useEffect(() => {
    if (countdown <= 0) return

    const timer = window.setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          window.clearInterval(timer)
          toast.error('订单已过期，请重新下单')
          navigateUser('products')
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => window.clearInterval(timer)
  }, [countdown, navigateUser])

  const formatCountdown = useCallback(() => {
    const minutes = Math.floor(countdown / 60)
    const seconds = countdown % 60
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }, [countdown])

  const handleBalancePay = async () => {
    if (!order) {
      toast.error('订单数据异常')
      return
    }
    if (userBalance < order.price) {
      toast.error('余额不足')
      return
    }
    if (!payPasswordSet) {
      toast.error('请先在用户中心设置支付密码')
      return
    }
    if (!/^\d{6}$/.test(balancePayPassword)) {
      setShowPayPasswordInput(true)
      toast.error('请输入6位支付密码')
      return
    }

    setPaying(true)
    const res = await apiPost('/api/order/pay/balance', {
      order_no: order.order_no,
      pay_password: balancePayPassword,
    })
    setPaying(false)

    if (res.success) {
      toast.success('支付成功')
      updateCachedOrderDetail({
        ...order,
        status: 2,
        payment_method: 'balance',
        payment_time: new Date().toISOString(),
      })
      navigateUser('payment-result', { order_no: order.order_no, status: 'success' })
    } else {
      toast.error(res.error || '余额支付失败')
      setBalancePayPassword('')
    }
  }

  const handleCancel = async () => {
    if (!order) return
    const res = await apiPost('/api/order/cancel', { order_no: order.order_no })
    if (res.success) {
      toast.success('订单已取消')
      updateCachedOrderDetail({ ...order, status: 3 })
      navigateUser('products')
    } else {
      toast.error(res.error || '取消订单失败')
    }
  }

  const balanceEnabled = paymentMethods?.balance?.enabled !== false
  const canUseBalance = balanceEnabled && payPasswordSet && !!order && userBalance >= order.price

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark-900">
        <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
      </div>
    )
  }

  return (
    <>
      <main className="flex-1 py-8 px-4">
        <div className="max-w-2xl mx-auto">
          <div className="space-y-6">
            {order && (
              <Card title="订单信息" icon={<i className="fas fa-receipt" />}>
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">订单号</span>
                    <span className="text-dark-100 font-mono text-sm">{order.order_no}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">商品名称</span>
                    <span className="text-dark-100">{order.product_name}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">购买数量</span>
                    <span className="text-dark-100">{order.quantity || 1}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">有效期</span>
                    <span className="text-dark-100">
                      {order.duration}
                      {order.duration_unit}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-dark-400">创建时间</span>
                    <span className="text-dark-100">{formatDateTime(order.created_at)}</span>
                  </div>
                  <div className="border-t border-dark-700 pt-3 flex justify-between items-center">
                    <span className="text-dark-400 text-lg">应付金额</span>
                    <span className="text-primary-400 text-2xl font-bold">{formatMoney(order.price)}</span>
                  </div>
                </div>
              </Card>
            )}

            <Card title="选择支付方式" icon={<i className="fas fa-credit-card" />}>
              <button
                onClick={() => {
                  setSelectedMethod('balance')
                  setShowPayPasswordInput(true)
                }}
                className={`w-full p-4 rounded-xl border-2 transition-all text-left ${
                  selectedMethod === 'balance'
                    ? 'border-primary-500 bg-primary-500/10'
                    : 'border-dark-600 bg-dark-700/30 hover:border-dark-500'
                }`}
              >
                <div className="flex items-center gap-3">
                  <span className="text-3xl">{'\u{1F4B0}'}</span>
                  <div>
                    <h4 className="text-dark-100 font-medium">余额支付</h4>
                    <p className="text-dark-500 text-sm">
                      {payPasswordSet ? `可用余额 ¥${userBalance.toFixed(2)}` : '请先设置支付密码'}
                    </p>
                  </div>
                  {selectedMethod === 'balance' && <i className="fas fa-check-circle text-primary-400 ml-auto" />}
                </div>
              </button>

              {!canUseBalance && (
                <div className="mt-4 rounded-xl border border-amber-500/20 bg-amber-500/10 p-4 text-amber-200 text-sm">
                  {!payPasswordSet ? '请先在用户中心设置支付密码。' : !balanceEnabled ? '余额支付当前不可用。' : '余额不足。'}
                </div>
              )}

              {showPayPasswordInput && (
                <div className="mt-4 p-4 bg-dark-700/30 rounded-xl border border-dark-600">
                  <label className="block text-sm text-dark-300 mb-2">
                    <i className="fas fa-lock mr-2" />
                    请输入支付密码
                  </label>
                  <Input
                    type="password"
                    placeholder="请输入6位支付密码"
                    maxLength={6}
                    value={balancePayPassword}
                    onChange={(e) => setBalancePayPassword(e.target.value.replace(/\D/g, ''))}
                    className="text-center text-lg tracking-widest"
                  />
                </div>
              )}
            </Card>

            <div className="flex flex-col sm:flex-row gap-4">
              <Button variant="secondary" className="flex-1" onClick={handleCancel} disabled={paying}>
                取消订单
              </Button>
              <Button
                variant="primary"
                className="flex-1"
                onClick={handleBalancePay}
                loading={paying}
                disabled={!canUseBalance}
              >
                <i className="fas fa-lock mr-2" />
                立即支付
              </Button>
            </div>

            <div className={`text-center text-sm flex items-center justify-center gap-2 ${countdown < 300 ? 'text-red-400' : 'text-dark-500'}`}>
              <i className="fas fa-clock" />
              <span>支付剩余时间：</span>
              <span className="font-mono font-bold text-lg">{formatCountdown()}</span>
            </div>
          </div>
        </div>
      </main>
    </>
  )
}

function PaymentPageContent() {
  const searchParams = useSearchParams()
  return <PaymentView orderNo={searchParams.get('order_no')} />
}

export default function PaymentPage() {
  return (
    <UserShell>
      <Suspense fallback={<main className="flex-1 flex items-center justify-center"><i className="fas fa-spinner fa-spin text-4xl text-primary-400" /></main>}>
        <PaymentPageContent />
      </Suspense>
    </UserShell>
  )
}
