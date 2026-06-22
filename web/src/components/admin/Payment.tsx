'use client'

import { useCallback, useEffect, useState } from 'react'
import { Card } from '@/components/ui'
import { apiGet } from '@/lib/api'
import { PaymentConfig } from './types'

/**
 * 支付配置页面。
 */
export function PaymentPage() {
  const [config, setConfig] = useState<PaymentConfig>({})
  const [loading, setLoading] = useState(true)

  const loadConfig = useCallback(async () => {
    const res = await apiGet<{ config: PaymentConfig }>('/api/admin/payment/config')
    if (res.success && res.config) {
      setConfig(res.config)
    }
    setLoading(false)
  }, [])

  useEffect(() => {
    loadConfig()
  }, [loadConfig])

  if (loading) {
    return (
      <div className="text-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  const balanceEnabled = config.balance?.enabled ?? true

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium text-dark-100">支付配置</h2>

      <Card>
        <div className="space-y-5">
          <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
            <div className="flex items-start gap-3">
              <i className="fas fa-wallet text-emerald-400 mt-1" />
              <div>
                <h3 className="text-emerald-300 font-medium">余额支付</h3>
                <p className="text-dark-300 text-sm mt-1">
                  用户可使用账户余额支付商品订单。
                </p>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="p-4 rounded-xl border border-dark-700 bg-dark-800/40">
              <p className="text-dark-500 text-sm">状态</p>
              <p className={balanceEnabled ? 'text-emerald-400 font-medium mt-1' : 'text-red-400 font-medium mt-1'}>
                {balanceEnabled ? '已启用' : '未启用'}
              </p>
            </div>
            <div className="p-4 rounded-xl border border-dark-700 bg-dark-800/40">
              <p className="text-dark-500 text-sm">配置类型</p>
              <p className="text-dark-100 font-medium mt-1">系统默认</p>
            </div>
            <div className="p-4 rounded-xl border border-dark-700 bg-dark-800/40">
              <p className="text-dark-500 text-sm">密钥配置</p>
              <p className="text-dark-100 font-medium mt-1">无需配置</p>
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}
