'use client'

import { useEffect, Suspense, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import toast from 'react-hot-toast'
import { Button, Card } from '@/components/ui'
import { UserShell } from '@/components/layout/UserShell'
import { getCachedOrderDetail, getOrderDetail } from '@/lib/orderData'
import { copyToClipboard } from '@/lib/utils'
import { useUserNavigation } from '@/lib/userNavigation'

interface PaymentResult {
  order_no: string
  kami_code: string
  product_name?: string
  quantity?: number
}

interface PaymentResultViewProps {
  orderNo?: string | null
  status?: string | null
}

/**
 * 支付结果视图。
 * 支持独立路由和用户端主入口 hash 切换复用。
 */
export function PaymentResultView({ orderNo, status }: PaymentResultViewProps) {
  const navigateUser = useUserNavigation()

  const cachedOrder = getCachedOrderDetail(orderNo)
  const [loading, setLoading] = useState(() => !(cachedOrder && status === 'success'))
  const [result, setResult] = useState<PaymentResult | null>(() => (
    cachedOrder && status === 'success'
      ? {
          order_no: cachedOrder.order_no,
          kami_code: cachedOrder.kami_code,
          product_name: cachedOrder.product_name,
          quantity: cachedOrder.quantity,
        }
      : null
  ))
  const [error, setError] = useState('')

  useEffect(() => {
    const loadOrderResult = async () => {
      if (!orderNo || status !== 'success') {
        setError('无效的支付信息')
        setLoading(false)
        return
      }

      const res = await getOrderDetail(orderNo)

      if (res.order) {
        setResult({
          order_no: res.order.order_no,
          kami_code: res.order.kami_code,
          product_name: res.order.product_name,
          quantity: res.order.quantity,
        })
      } else {
        setError(res.error || '获取订单信息失败')
      }
      setLoading(false)
    }

    loadOrderResult()
  }, [orderNo, status])

  const handleCopyKami = async () => {
    if (!result?.kami_code) return
    const success = await copyToClipboard(result.kami_code)
    if (success) {
      toast.success('卡密已复制到剪贴板')
    }
  }

  const handleExportKami = () => {
    if (!result?.kami_code) return

    const kamiCodes = result.kami_code.split('\n').filter((code) => code.trim())
    if (kamiCodes.length <= 1) return

    const csvContent = [
      '序号,卡密',
      ...kamiCodes.map((code, index) => `${index + 1},"${code.trim()}"`),
    ].join('\n')

    const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `卡密_${result.order_no}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    toast.success('卡密已导出')
  }

  const getKamiCount = () => {
    if (!result?.kami_code) return 0
    return result.kami_code.split('\n').filter((code) => code.trim()).length
  }

  if (loading) {
    return (
      <main className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <i className="fas fa-spinner fa-spin text-5xl text-primary-400 mb-4" />
          <p className="text-dark-300 text-lg">正在确认支付结果...</p>
          <p className="text-dark-500 text-sm mt-2">请稍候，不要关闭此页面</p>
        </div>
      </main>
    )
  }

  if (error) {
    return (
      <main className="flex-1 py-8 px-4">
        <div className="max-w-lg mx-auto">
          <div>
            <Card>
              <div className="text-center py-8">
                <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-red-500/20 flex items-center justify-center">
                  <i className="fas fa-times text-4xl text-red-400" />
                </div>
                <h2 className="text-2xl font-bold text-dark-100 mb-2">支付失败</h2>
                <p className="text-dark-400 mb-6">{error}</p>
                <div className="flex flex-col sm:flex-row gap-4 justify-center">
                  <Button variant="secondary" onClick={() => navigateUser('user')}>
                    查看订单
                  </Button>
                  <Button variant="primary" onClick={() => navigateUser('products')}>
                    继续购买
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </main>
    )
  }

  const kamiCount = getKamiCount()

  return (
    <>
      <main className="flex-1 py-8 px-4">
        <div className="max-w-lg mx-auto">
          <div>
            <Card>
              <div className="text-center py-6">
                <div className="w-20 h-20 mx-auto mb-6 rounded-full bg-emerald-500/20 flex items-center justify-center">
                  <i className="fas fa-check text-4xl text-emerald-400" />
                </div>
                <h2 className="text-2xl font-bold text-dark-100 mb-2">支付成功</h2>
                <p className="text-dark-400 mb-6">感谢您的购买！</p>

                <div className="bg-dark-700/30 rounded-xl p-4 mb-6 text-left">
                  <div className="space-y-3">
                    <div>
                      <span className="text-dark-500 text-sm">订单号</span>
                      <p className="text-dark-100 font-mono">{result?.order_no}</p>
                    </div>
                    <div>
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-dark-500 text-sm">卡密 {kamiCount > 1 && `(${kamiCount}个)`}</span>
                        {kamiCount > 1 && (
                          <Button size="sm" variant="ghost" onClick={handleExportKami} className="text-xs">
                            <i className="fas fa-download mr-1" />
                            导出CSV
                          </Button>
                        )}
                      </div>
                      <div className="mt-1 bg-dark-800/50 rounded-lg p-3">
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex-1 space-y-2 max-h-60 overflow-y-auto">
                            {result?.kami_code?.split('\n').filter((code) => code.trim()).map((code, index) => (
                              <div key={index} className="flex items-center gap-2">
                                {kamiCount > 1 && (
                                  <span className="text-dark-500 text-sm w-6 flex-shrink-0">{index + 1}.</span>
                                )}
                                <p className="font-mono text-primary-400 break-all text-lg flex-1">
                                  {code.trim()}
                                </p>
                              </div>
                            ))}
                          </div>
                          <Button size="sm" variant="ghost" onClick={handleCopyKami} className="flex-shrink-0">
                            <i className="fas fa-copy" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="bg-blue-500/10 border border-blue-500/30 rounded-xl p-4 mb-6 text-left">
                  <div className="flex items-start gap-3">
                    <i className="fas fa-info-circle text-blue-400 mt-0.5" />
                    <div className="text-sm">
                      <p className="text-blue-300 font-medium">请妥善保管您的卡密</p>
                      <p className="text-blue-400/70 mt-1">
                        卡密是您使用服务的凭证，请勿泄露给他人。您可以在用户中心随时查看。
                      </p>
                    </div>
                  </div>
                </div>

                <div className="flex flex-col sm:flex-row gap-4">
                  <Button variant="secondary" className="flex-1" onClick={() => navigateUser('user')}>
                    <i className="fas fa-user mr-2" />
                    用户中心
                  </Button>
                  <Button variant="primary" className="flex-1" onClick={handleCopyKami}>
                    <i className="fas fa-copy mr-2" />
                    复制卡密
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </main>
    </>
  )
}

function PaymentResultContent() {
  const searchParams = useSearchParams()
  return (
    <PaymentResultView
      orderNo={searchParams.get('order_no')}
      status={searchParams.get('status')}
    />
  )
}

export default function PaymentResultPage() {
  return (
    <UserShell>
      <Suspense
        fallback={
          <main className="flex-1 flex items-center justify-center">
            <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
          </main>
        }
      >
        <PaymentResultContent />
      </Suspense>
    </UserShell>
  )
}
