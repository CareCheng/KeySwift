'use client'

import { useState, useEffect } from 'react'
import { Button, Card } from '@/components/ui'
import { apiGet } from '@/lib/api'

/**
 * 统计卡片组件
 * 支持移动端响应式布局
 */
function StatCard({ icon, value, label }: { icon: string; value: string | number; label: string }) {
  return (
    <div className="bg-dark-800/50 rounded-xl p-4 sm:p-6 border border-dark-700/50">
      <div className="text-2xl sm:text-3xl mb-2">{icon}</div>
      <div className="text-xl sm:text-2xl font-bold text-dark-100">{value}</div>
      <div className="text-dark-500 text-xs sm:text-sm">{label}</div>
    </div>
  )
}

/**
 * 仪表盘页面
 * 支持移动端响应式布局
 */
export function DashboardPage() {
  const [data, setData] = useState<{
    db_connected: boolean
    stats: { total_orders: number; paid_orders: number; total_revenue: number; today_orders: number }
  } | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadDashboard()
  }, [])

  const loadDashboard = async () => {
    const res = await apiGet<typeof data>('/api/admin/dashboard')
    if (res.success) setData(res as typeof data)
    setLoading(false)
  }

  if (loading) {
    return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>
  }

  if (!data?.db_connected) {
    return (
      <Card>
        <div className="text-center py-8">
          <div className="text-4xl mb-4">⚠️</div>
          <h3 className="text-lg font-medium text-dark-100 mb-2">数据库未连接</h3>
          <p className="text-dark-400">请先前往数据库配置页面配置数据库连接</p>
        </div>
      </Card>
    )
  }

  const stats = data?.stats || { total_orders: 0, paid_orders: 0, total_revenue: 0, today_orders: 0 }

  return (
    <div className="space-y-4 sm:space-y-6">
      {/* 主要统计 - 移动端2列，桌面端4列 */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
        <StatCard icon="📦" value={stats.total_orders} label="总订单数" />
        <StatCard icon="✅" value={stats.paid_orders} label="已完成订单" />
        <StatCard icon="💰" value={`¥${stats.total_revenue.toFixed(2)}`} label="总收入" />
        <StatCard icon="📈" value={stats.today_orders} label="今日订单" />
      </div>

      {/* 快捷操作 */}
      <Card title="快捷操作">
        <div className="flex flex-wrap gap-2 sm:gap-3">
          <Button onClick={() => (window.location.hash = 'products')} className="flex-1 sm:flex-none">
            <i className="fas fa-box mr-2 hidden sm:inline" />管理商品
          </Button>
          <Button variant="secondary" onClick={() => (window.location.hash = 'orders')} className="flex-1 sm:flex-none">
            <i className="fas fa-list mr-2 hidden sm:inline" />查看订单
          </Button>
          <Button variant="secondary" onClick={() => (window.location.hash = 'config')} className="flex-1 sm:flex-none">
            <i className="fas fa-cog mr-2 hidden sm:inline" />系统配置
          </Button>
        </div>
      </Card>
    </div>
  )
}
