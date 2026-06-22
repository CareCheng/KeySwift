'use client'

import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Badge, Button, Card, Input, Modal } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { formatDateTime, formatMoney } from '@/lib/utils'
import { PermissionGuard } from '@/contexts/PermissionContext'

interface UserBalance {
  id: number
  user_id: number
  username?: string
  email?: string
  balance: number
  frozen: number
  total_in: number
  total_out: number
  updated_at: string
}

interface BalanceLog {
  id: number
  user_id: number
  type: string
  amount: number
  before_balance: number
  after_balance: number
  order_no: string
  remark: string
  operator_type: string
  created_at: string
}

interface BalanceStats {
  total_balance: number
  total_frozen: number
  total_in: number
  total_out: number
  user_count: number
}

type BalanceTab = 'balances' | 'logs'

const pageSize = 20

/**
 * 余额管理页面
 */
export function BalancePage() {
  const [activeTab, setActiveTab] = useState<BalanceTab>('balances')
  const [balances, setBalances] = useState<UserBalance[]>([])
  const [logs, setLogs] = useState<BalanceLog[]>([])
  const [stats, setStats] = useState<BalanceStats | null>(null)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [keyword, setKeyword] = useState('')
  const [logType, setLogType] = useState('')
  const [selectedUser, setSelectedUser] = useState<UserBalance | null>(null)
  const [showAdjustModal, setShowAdjustModal] = useState(false)
  const [adjustForm, setAdjustForm] = useState({
    amount: '',
    remark: '',
  })

  useEffect(() => {
    loadStats()
  }, [])

  useEffect(() => {
    loadData()
  }, [activeTab, page, logType])

  const loadStats = async () => {
    const res = await apiGet<{ data: BalanceStats }>('/api/admin/balance/stats')
    if (res.success && res.data) {
      setStats(res.data)
    }
  }

  const loadData = async () => {
    setLoading(true)
    if (activeTab === 'balances') {
      await loadBalances()
    } else {
      await loadLogs()
    }
    setLoading(false)
  }

  const loadBalances = async () => {
    const params = new URLSearchParams({
      page: String(page),
      page_size: String(pageSize),
    })
    if (keyword.trim()) {
      params.set('keyword', keyword.trim())
    }

    const res = await apiGet<{ data: UserBalance[]; total: number }>(`/api/admin/balances?${params.toString()}`)
    if (res.success) {
      setBalances(res.data || [])
      setTotal(res.total || 0)
    } else {
      toast.error(res.error || '获取余额列表失败')
    }
  }

  const loadLogs = async () => {
    const params = new URLSearchParams({
      page: String(page),
      page_size: String(pageSize),
    })
    if (logType) {
      params.set('type', logType)
    }

    const res = await apiGet<{ data: BalanceLog[]; total: number }>(`/api/admin/balance/logs?${params.toString()}`)
    if (res.success) {
      setLogs(res.data || [])
      setTotal(res.total || 0)
    } else {
      toast.error(res.error || '获取余额日志失败')
    }
  }

  const searchBalances = () => {
    setPage(1)
    if (activeTab === 'balances') {
      loadData()
    }
  }

  const openAdjustModal = (user: UserBalance) => {
    setSelectedUser(user)
    setAdjustForm({ amount: '', remark: '' })
    setShowAdjustModal(true)
  }

  const submitAdjust = async () => {
    if (!selectedUser) return

    const amount = Number(adjustForm.amount)
    if (!Number.isFinite(amount) || amount === 0) {
      toast.error('请输入非零调整金额')
      return
    }

    setSaving(true)
    const res = await apiPost('/api/admin/balance/adjust', {
      user_id: selectedUser.user_id,
      amount,
      remark: adjustForm.remark.trim() || '管理员调整余额',
    })
    setSaving(false)

    if (res.success) {
      toast.success('余额调整成功')
      setShowAdjustModal(false)
      await loadStats()
      await loadData()
    } else {
      toast.error(res.error || '余额调整失败')
    }
  }

  const getLogTypeBadge = (type: string) => {
    const typeMap: Record<string, { label: string; variant: 'success' | 'warning' | 'danger' | 'info' | 'default' }> = {
      consume: { label: '消费', variant: 'danger' },
      refund: { label: '退款', variant: 'success' },
      freeze: { label: '冻结', variant: 'warning' },
      unfreeze: { label: '解冻', variant: 'info' },
      adjust: { label: '调整', variant: 'default' },
    }

    const item = typeMap[type] || { label: type || '未知', variant: 'default' as const }
    return <Badge variant={item.variant}>{item.label}</Badge>
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-dark-100">余额管理</h2>
        <p className="text-dark-400 mt-1">查看用户余额、余额流水并执行管理员余额调整</p>
      </div>

      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-5 gap-4">
          <Card>
            <p className="text-sm text-dark-400">余额用户</p>
            <p className="text-2xl font-bold text-dark-100 mt-2">{stats.user_count}</p>
          </Card>
          <Card>
            <p className="text-sm text-dark-400">可用余额</p>
            <p className="text-2xl font-bold text-emerald-400 mt-2">{formatMoney(stats.total_balance || 0)}</p>
          </Card>
          <Card>
            <p className="text-sm text-dark-400">冻结余额</p>
            <p className="text-2xl font-bold text-amber-400 mt-2">{formatMoney(stats.total_frozen || 0)}</p>
          </Card>
          <Card>
            <p className="text-sm text-dark-400">累计入账</p>
            <p className="text-2xl font-bold text-blue-400 mt-2">{formatMoney(stats.total_in || 0)}</p>
          </Card>
          <Card>
            <p className="text-sm text-dark-400">累计支出</p>
            <p className="text-2xl font-bold text-red-400 mt-2">{formatMoney(stats.total_out || 0)}</p>
          </Card>
        </div>
      )}

      <div className="flex gap-2 border-b border-dark-700">
        <button
          onClick={() => {
            setActiveTab('balances')
            setPage(1)
          }}
          className={`px-4 py-3 font-medium transition-colors ${activeTab === 'balances' ? 'text-primary-400 border-b-2 border-primary-400' : 'text-dark-400 hover:text-dark-200'}`}
        >
          用户余额
        </button>
        <button
          onClick={() => {
            setActiveTab('logs')
            setPage(1)
          }}
          className={`px-4 py-3 font-medium transition-colors ${activeTab === 'logs' ? 'text-primary-400 border-b-2 border-primary-400' : 'text-dark-400 hover:text-dark-200'}`}
        >
          余额流水
        </button>
      </div>

      {activeTab === 'balances' && (
        <Card>
          <div className="flex flex-col md:flex-row gap-3 mb-6">
            <Input
              placeholder="搜索用户名或邮箱"
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') searchBalances()
              }}
            />
            <Button onClick={searchBalances}>
              <i className="fas fa-search" />
              搜索
            </Button>
            <Button variant="secondary" onClick={loadData}>
              <i className="fas fa-rotate" />
              刷新
            </Button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-dark-700 text-dark-400">
                  <th className="pb-3 font-medium">用户</th>
                  <th className="pb-3 font-medium">可用余额</th>
                  <th className="pb-3 font-medium">冻结余额</th>
                  <th className="pb-3 font-medium">累计入账</th>
                  <th className="pb-3 font-medium">累计支出</th>
                  <th className="pb-3 font-medium">更新时间</th>
                  <th className="pb-3 font-medium text-right">操作</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={7} className="py-10 text-center text-dark-400">
                      <i className="fas fa-spinner fa-spin mr-2" />
                      加载中...
                    </td>
                  </tr>
                ) : balances.length === 0 ? (
                  <tr>
                    <td colSpan={7} className="py-10 text-center text-dark-400">暂无余额记录</td>
                  </tr>
                ) : (
                  balances.map((item) => (
                    <tr key={item.id} className="border-b border-dark-700/50 text-dark-200">
                      <td className="py-4">
                        <div>
                          <p className="font-medium">{item.username || `用户 ${item.user_id}`}</p>
                          {item.email && <p className="text-sm text-dark-500">{item.email}</p>}
                        </div>
                      </td>
                      <td className="py-4 text-emerald-400 font-medium">{formatMoney(item.balance || 0)}</td>
                      <td className="py-4 text-amber-400">{formatMoney(item.frozen || 0)}</td>
                      <td className="py-4">{formatMoney(item.total_in || 0)}</td>
                      <td className="py-4">{formatMoney(item.total_out || 0)}</td>
                      <td className="py-4 text-dark-400">{item.updated_at ? formatDateTime(item.updated_at) : '-'}</td>
                      <td className="py-4 text-right">
                        <PermissionGuard permission="balance:adjust">
                          <Button size="sm" onClick={() => openAdjustModal(item)}>
                            调整
                          </Button>
                        </PermissionGuard>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {activeTab === 'logs' && (
        <Card>
          <div className="flex flex-col md:flex-row gap-3 mb-6">
            <select
              className="input h-12 md:w-48"
              value={logType}
              onChange={(event) => {
                setLogType(event.target.value)
                setPage(1)
              }}
            >
              <option value="">全部类型</option>
              <option value="consume">消费</option>
              <option value="refund">退款</option>
              <option value="freeze">冻结</option>
              <option value="unfreeze">解冻</option>
              <option value="adjust">调整</option>
            </select>
            <Button variant="secondary" onClick={loadData}>
              <i className="fas fa-rotate" />
              刷新
            </Button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left">
              <thead>
                <tr className="border-b border-dark-700 text-dark-400">
                  <th className="pb-3 font-medium">用户ID</th>
                  <th className="pb-3 font-medium">类型</th>
                  <th className="pb-3 font-medium">变动金额</th>
                  <th className="pb-3 font-medium">变动前</th>
                  <th className="pb-3 font-medium">变动后</th>
                  <th className="pb-3 font-medium">订单号</th>
                  <th className="pb-3 font-medium">备注</th>
                  <th className="pb-3 font-medium">时间</th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan={8} className="py-10 text-center text-dark-400">
                      <i className="fas fa-spinner fa-spin mr-2" />
                      加载中...
                    </td>
                  </tr>
                ) : logs.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="py-10 text-center text-dark-400">暂无余额流水</td>
                  </tr>
                ) : (
                  logs.map((item) => (
                    <tr key={item.id} className="border-b border-dark-700/50 text-dark-200">
                      <td className="py-4">{item.user_id}</td>
                      <td className="py-4">{getLogTypeBadge(item.type)}</td>
                      <td className={`py-4 font-medium ${item.amount >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                        {item.amount >= 0 ? '+' : ''}{formatMoney(item.amount || 0)}
                      </td>
                      <td className="py-4">{formatMoney(item.before_balance || 0)}</td>
                      <td className="py-4">{formatMoney(item.after_balance || 0)}</td>
                      <td className="py-4 text-dark-400 font-mono text-sm">{item.order_no || '-'}</td>
                      <td className="py-4 text-dark-400 max-w-xs truncate" title={item.remark}>{item.remark || '-'}</td>
                      <td className="py-4 text-dark-400">{formatDateTime(item.created_at)}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      <div className="flex items-center justify-between">
        <span className="text-sm text-dark-400">共 {total} 条，第 {page} / {totalPages} 页</span>
        <div className="flex gap-2">
          <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => setPage(page - 1)}>
            上一页
          </Button>
          <Button variant="secondary" size="sm" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
            下一页
          </Button>
        </div>
      </div>

      <Modal isOpen={showAdjustModal} onClose={() => setShowAdjustModal(false)} title="调整余额" size="sm">
        {selectedUser && (
          <div className="space-y-4">
            <div className="rounded-lg bg-dark-800/50 p-4 space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-dark-400">用户</span>
                <span className="text-dark-100">{selectedUser.username || selectedUser.user_id}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-dark-400">当前余额</span>
                <span className="text-emerald-400">{formatMoney(selectedUser.balance || 0)}</span>
              </div>
            </div>

            <Input
              label="调整金额"
              type="number"
              step="0.01"
              placeholder="正数增加，负数扣减"
              value={adjustForm.amount}
              onChange={(event) => setAdjustForm({ ...adjustForm, amount: event.target.value })}
            />
            <Input
              label="备注"
              placeholder="填写调整原因"
              value={adjustForm.remark}
              onChange={(event) => setAdjustForm({ ...adjustForm, remark: event.target.value })}
            />
            <div className="flex gap-3">
              <Button variant="secondary" className="flex-1" onClick={() => setShowAdjustModal(false)}>
                取消
              </Button>
              <Button className="flex-1" onClick={submitAdjust} loading={saving}>
                确认调整
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}
