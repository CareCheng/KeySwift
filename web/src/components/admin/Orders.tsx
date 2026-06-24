'use client'

import { useState, useEffect, useCallback } from 'react'
import { Button, Card, Input, Modal } from '@/components/ui'
import { apiGet } from '@/lib/api'
import { Order } from './types'

interface OrderTrace {
  items?: Array<{ id: number; product_ref: string; quantity: number; unit_price_cents: number; owner_plugin_id: string }>
  payment_attempts?: Array<{ id: number; attempt_no: string; payment_plugin_id: string; payment_channel: string; amount_cents: number; status: string; provider_transaction_id: string }>
  payment_callbacks?: Array<{ id: number; callback_id: string; provider_transaction_id: string; provider_status: string; amount_cents: number; verified: boolean; status: string; error_message: string; received_at: string }>
  deliveries?: Array<{ id: number; delivery_no: string; fulfillment_plugin_id: string; delivery_type: string; status: string; error_message: string; delivered_at?: string }>
  delivery_facts?: Array<{ id: number; fact_id: string; delivery_no: string; fulfillment_plugin_id: string; fact_type: string; status: string; error_message: string; occurred_at: string }>
  state_events?: Array<{ id: number; event_type: string; from_status: string; to_status: string; payment_status: string; delivery_status: string; owner_plugin_id: string; created_at: string }>
}

/**
 * 订单管理页面
 * 支持移动端响应式布局
 */
export function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [search, setSearch] = useState('')
  const [showDetailModal, setShowDetailModal] = useState(false)
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)
  const [selectedTrace, setSelectedTrace] = useState<OrderTrace | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)

  const loadOrders = useCallback(async () => {
    setLoading(true)
    const params = new URLSearchParams({ page: String(page), page_size: '20' })
    if (search) params.append('search', search)
    const res = await apiGet<{ orders: Order[]; total_pages: number }>(`/api/admin/orders?${params}`)
    if (res.success) {
      setOrders(res.orders || [])
      setTotalPages(res.total_pages || 1)
    }
    setLoading(false)
  }, [page, search])

  useEffect(() => { loadOrders() }, [loadOrders])

  const getStatusInfo = (status: number) => {
    const map: Record<number, { text: string; class: string }> = {
      0: { text: '待支付', class: 'bg-yellow-500/20 text-yellow-400' },
      1: { text: '已支付', class: 'bg-blue-500/20 text-blue-400' },
      2: { text: '已完成', class: 'bg-green-500/20 text-green-400' },
      3: { text: '已取消', class: 'bg-red-500/20 text-red-400' },
    }
    return map[status] || { text: '未知', class: 'bg-gray-500/20 text-gray-400' }
  }

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setPage(1)
    loadOrders()
  }

  const viewDetail = async (order: Order) => {
    setSelectedOrder(order)
    setSelectedTrace(null)
    setShowDetailModal(true)
    setDetailLoading(true)
    const res = await apiGet<{ order: Order; trace?: OrderTrace }>(`/api/admin/order/${order.id}`)
    if (res.success) {
      setSelectedOrder(res.order || order)
      setSelectedTrace(res.trace || null)
    }
    setDetailLoading(false)
  }

  if (loading && orders.length === 0) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <div className="space-y-4">
      {/* 标题和搜索栏 - 移动端垂直布局 */}
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-3">
        <h2 className="text-lg font-medium text-dark-100">订单列表</h2>
        <form onSubmit={handleSearch} className="flex gap-2">
          <Input placeholder="搜索订单号/用户名" value={search} onChange={(e) => setSearch(e.target.value)} className="flex-1 sm:w-48" />
          <Button type="submit" size="sm">搜索</Button>
        </form>
      </div>
      <Card>
        {orders.length === 0 ? (
          <div className="text-center py-12 text-dark-500">暂无订单</div>
        ) : (
          <>
            {/* 桌面端表格视图 */}
            <div className="hidden lg:block overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-dark-700">
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">订单号</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">用户</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">商品</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">数量</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">金额</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">状态</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">时间</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {orders.map((order) => {
                    const status = getStatusInfo(order.status)
                    return (
                      <tr key={order.id} className="border-b border-dark-700/50 hover:bg-dark-700/30">
                        <td className="py-3 px-4 text-dark-100 font-mono text-sm">
                          {order.order_no}
                        </td>
                        <td className="py-3 px-4 text-dark-300">{order.username}</td>
                        <td className="py-3 px-4 text-dark-300">{order.product_name}</td>
                        <td className="py-3 px-4 text-dark-300">{order.quantity || 1}</td>
                        <td className="py-3 px-4 text-dark-300">¥{order.price.toFixed(2)}</td>
                        <td className="py-3 px-4">
                          <span className={`px-2 py-1 rounded text-xs ${status.class}`}>{status.text}</span>
                        </td>
                        <td className="py-3 px-4 text-dark-300 text-sm">{order.created_at}</td>
                        <td className="py-3 px-4">
                          <Button size="sm" variant="ghost" onClick={() => viewDetail(order)}>详情</Button>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>

            {/* 移动端卡片视图 */}
            <div className="lg:hidden space-y-3">
              {orders.map((order) => {
                const status = getStatusInfo(order.status)
                return (
                  <div key={order.id} className="bg-dark-700/30 rounded-lg p-4 space-y-3">
                    {/* 订单号和状态 */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="text-dark-100 font-mono text-sm truncate max-w-[180px]">{order.order_no}</span>
                      </div>
                      <span className={`px-2 py-1 rounded text-xs ${status.class}`}>{status.text}</span>
                    </div>
                    {/* 商品和金额 */}
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-dark-300 truncate max-w-[60%]">
                        {order.product_name}
                        {(order.quantity || 1) > 1 && <span className="text-dark-500 ml-1">x{order.quantity}</span>}
                      </span>
                      <span className="text-primary-400 font-medium">¥{order.price.toFixed(2)}</span>
                    </div>
                    {/* 用户和时间 */}
                    <div className="flex items-center justify-between text-xs text-dark-500">
                      <span>{order.username}</span>
                      <span>{order.created_at}</span>
                    </div>
                    {/* 操作按钮 */}
                    <div className="pt-2 border-t border-dark-600/50">
                      <Button size="sm" variant="ghost" onClick={() => viewDetail(order)} className="w-full">
                        <i className="fas fa-eye mr-2" />查看详情
                      </Button>
                    </div>
                  </div>
                )
              })}
            </div>

            {/* 分页 - 移动端简化 */}
            {totalPages > 1 && (
              <div className="flex flex-wrap justify-center gap-2 mt-4">
                <Button size="sm" variant="secondary" disabled={page <= 1} onClick={() => setPage(1)} className="hidden sm:inline-flex">首页</Button>
                <Button size="sm" variant="secondary" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                  <i className="fas fa-chevron-left sm:mr-1" /><span className="hidden sm:inline">上一页</span>
                </Button>
                <span className="px-3 py-2 text-dark-400 text-sm">{page} / {totalPages}</span>
                <Button size="sm" variant="secondary" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
                  <span className="hidden sm:inline">下一页</span><i className="fas fa-chevron-right sm:ml-1" />
                </Button>
                <Button size="sm" variant="secondary" disabled={page >= totalPages} onClick={() => setPage(totalPages)} className="hidden sm:inline-flex">末页</Button>
              </div>
            )}
          </>
        )}
      </Card>

      {/* 订单详情弹窗 */}
      <Modal isOpen={showDetailModal} onClose={() => setShowDetailModal(false)} title="订单详情">
        {selectedOrder && (
          <div className="space-y-4">
            {detailLoading && (
              <div className="rounded-lg bg-dark-700/40 p-3 text-sm text-dark-400">
                <i className="fas fa-spinner fa-spin mr-2" />正在加载订单追踪信息
              </div>
            )}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
              <div className="flex flex-col sm:flex-row sm:items-center gap-1">
                <span className="text-dark-500">订单号：</span>
                <span className="text-dark-100 font-mono break-all">{selectedOrder.order_no}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center gap-1">
                <span className="text-dark-500">用户：</span>
                <span className="text-dark-100">{selectedOrder.username}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center gap-1">
                <span className="text-dark-500">商品：</span>
                <span className="text-dark-100">{selectedOrder.product_name}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center gap-1">
                <span className="text-dark-500">数量：</span>
                <span className="text-dark-100">{selectedOrder.quantity || 1}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center gap-1">
                <span className="text-dark-500">金额：</span>
                <span className="text-dark-100">¥{selectedOrder.price.toFixed(2)}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center gap-1">
                <span className="text-dark-500">状态：</span>
                <span className={`px-2 py-1 rounded text-xs inline-block ${getStatusInfo(selectedOrder.status).class}`}>{getStatusInfo(selectedOrder.status).text}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center gap-1">
                <span className="text-dark-500">创建时间：</span>
                <span className="text-dark-100">{selectedOrder.created_at}</span>
              </div>
              <div className="flex flex-col sm:flex-row sm:items-center gap-1">
                <span className="text-dark-500">支付时间：</span>
                <span className="text-dark-100">{selectedOrder.paid_at || '-'}</span>
              </div>
            </div>
            {selectedOrder.card_info && (
              <div>
                <span className="text-dark-500 text-sm">卡密信息：</span>
                <pre className="mt-2 p-3 bg-dark-700 rounded text-dark-100 text-sm whitespace-pre-wrap break-all">{selectedOrder.card_info}</pre>
              </div>
            )}
            {selectedTrace && (
              <OrderTracePanel trace={selectedTrace} />
            )}
            <div className="flex justify-end pt-4">
              <Button variant="secondary" onClick={() => setShowDetailModal(false)}>关闭</Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  )
}

function OrderTracePanel({ trace }: { trace: OrderTrace }) {
  const hasTrace = Boolean(
    (trace.items && trace.items.length > 0)
    || (trace.payment_attempts && trace.payment_attempts.length > 0)
    || (trace.payment_callbacks && trace.payment_callbacks.length > 0)
    || (trace.deliveries && trace.deliveries.length > 0)
    || (trace.delivery_facts && trace.delivery_facts.length > 0)
    || (trace.state_events && trace.state_events.length > 0),
  )

  if (!hasTrace) {
    return (
      <div className="rounded-lg border border-dark-700 bg-dark-800/40 p-4 text-sm text-dark-500">
        暂无订单追踪记录。
      </div>
    )
  }

  return (
    <div className="space-y-4 border-t border-dark-700 pt-4">
      <h3 className="text-sm font-medium text-dark-100">订单追踪</h3>

      {trace.items && trace.items.length > 0 && (
        <TraceSection title="订单明细">
          {trace.items.map((item) => (
            <TraceCard key={item.id}>
              <TraceLine label="商品引用" value={item.product_ref || '-'} />
              <TraceLine label="数量" value={String(item.quantity || 0)} />
              <TraceLine label="单价" value={`¥${((item.unit_price_cents || 0) / 100).toFixed(2)}`} />
              <TraceLine label="归属插件" value={item.owner_plugin_id || '宿主'} />
            </TraceCard>
          ))}
        </TraceSection>
      )}

      {trace.payment_attempts && trace.payment_attempts.length > 0 && (
        <TraceSection title="支付尝试">
          {trace.payment_attempts.map((attempt) => (
            <TraceCard key={attempt.id}>
              <TraceLine label="尝试号" value={attempt.attempt_no || '-'} />
              <TraceLine label="支付方" value={attempt.payment_plugin_id || '-'} />
              <TraceLine label="渠道" value={attempt.payment_channel || '-'} />
              <TraceLine label="金额" value={`¥${((attempt.amount_cents || 0) / 100).toFixed(2)}`} />
              <TraceLine label="状态" value={attempt.status || '-'} />
              <TraceLine label="交易号" value={attempt.provider_transaction_id || '-'} />
            </TraceCard>
          ))}
        </TraceSection>
      )}

      {trace.payment_callbacks && trace.payment_callbacks.length > 0 && (
        <TraceSection title="支付回调">
          {trace.payment_callbacks.map((callback) => (
            <TraceCard key={callback.id}>
              <TraceLine label="回调号" value={callback.callback_id || '-'} />
              <TraceLine label="交易号" value={callback.provider_transaction_id || '-'} />
              <TraceLine label="渠道状态" value={callback.provider_status || '-'} />
              <TraceLine label="金额" value={`¥${((callback.amount_cents || 0) / 100).toFixed(2)}`} />
              <TraceLine label="验证" value={callback.verified ? '已验证' : '未验证'} />
              <TraceLine label="状态" value={callback.status || '-'} />
              <TraceLine label="错误" value={callback.error_message || '-'} />
            </TraceCard>
          ))}
        </TraceSection>
      )}

      {trace.deliveries && trace.deliveries.length > 0 && (
        <TraceSection title="交付任务">
          {trace.deliveries.map((delivery) => (
            <TraceCard key={delivery.id}>
              <TraceLine label="交付号" value={delivery.delivery_no || '-'} />
              <TraceLine label="交付方" value={delivery.fulfillment_plugin_id || '-'} />
              <TraceLine label="类型" value={delivery.delivery_type || '-'} />
              <TraceLine label="状态" value={delivery.status || '-'} />
              <TraceLine label="完成时间" value={delivery.delivered_at || '-'} />
              <TraceLine label="错误" value={delivery.error_message || '-'} />
            </TraceCard>
          ))}
        </TraceSection>
      )}

      {trace.delivery_facts && trace.delivery_facts.length > 0 && (
        <TraceSection title="交付事实">
          {trace.delivery_facts.map((fact) => (
            <TraceCard key={fact.id}>
              <TraceLine label="事实号" value={fact.fact_id || '-'} />
              <TraceLine label="交付号" value={fact.delivery_no || '-'} />
              <TraceLine label="交付方" value={fact.fulfillment_plugin_id || '-'} />
              <TraceLine label="事实类型" value={fact.fact_type || '-'} />
              <TraceLine label="状态" value={fact.status || '-'} />
              <TraceLine label="错误" value={fact.error_message || '-'} />
            </TraceCard>
          ))}
        </TraceSection>
      )}

      {trace.state_events && trace.state_events.length > 0 && (
        <TraceSection title="状态事件">
          {trace.state_events.map((event) => (
            <TraceCard key={event.id}>
              <TraceLine label="事件" value={event.event_type || '-'} />
              <TraceLine label="状态流转" value={`${event.from_status || '-'} → ${event.to_status || '-'}`} />
              <TraceLine label="支付状态" value={event.payment_status || '-'} />
              <TraceLine label="交付状态" value={event.delivery_status || '-'} />
              <TraceLine label="归属插件" value={event.owner_plugin_id || '宿主'} />
              <TraceLine label="时间" value={event.created_at || '-'} />
            </TraceCard>
          ))}
        </TraceSection>
      )}
    </div>
  )
}

function TraceSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="space-y-2">
      <div className="text-xs font-medium text-dark-400">{title}</div>
      <div className="space-y-2">{children}</div>
    </div>
  )
}

function TraceCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid gap-2 rounded-lg border border-dark-700 bg-dark-800/40 p-3 text-xs sm:grid-cols-2">
      {children}
    </div>
  )
}

function TraceLine({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid grid-cols-[4.5rem_1fr] gap-2">
      <span className="text-dark-500">{label}</span>
      <span className="break-all text-dark-300">{value || '-'}</span>
    </div>
  )
}
