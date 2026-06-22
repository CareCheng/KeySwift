'use client'

import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Badge } from '@/components/ui'
import { apiGet } from '@/lib/api'
import { copyToClipboard, formatDateTime } from '@/lib/utils'

interface Kami {
  order_no: string
  product_name: string
  kami_code: string
  quantity: number
  status: number
  payment_time?: string
  created_at: string
}

export function KamisTab() {
  const [kamis, setKamis] = useState<Kami[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const loadKamis = async () => {
      setLoading(true)
      const res = await apiGet<{ kamis: Kami[] }>('/api/user/kamis')
      if (res.success && res.kamis) {
        setKamis(res.kamis)
      }
      setLoading(false)
    }
    loadKamis()
  }, [])

  const handleCopyKami = async (code: string) => {
    const success = await copyToClipboard(code)
    if (success) toast.success('已复制到剪贴板')
  }

  const handleExportKami = (kami: Kami) => {
    const kamiCodes = kami.kami_code.split('\n').filter((code) => code.trim())
    if (kamiCodes.length <= 1) return

    const csvContent = [
      '序号,卡密,商品名称,订单号',
      ...kamiCodes.map((code, index) => `${index + 1},"${code.trim()}","${kami.product_name}","${kami.order_no}"`),
    ].join('\n')

    const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `卡密_${kami.order_no}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    toast.success('卡密已导出')
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div>
      <Card title="我的卡密" icon={<i className="fas fa-key" />}>
        {kamis.length === 0 ? (
          <div className="text-center py-8 text-dark-400">暂无卡密</div>
        ) : (
          <div className="space-y-4">
            {kamis.map((kami) => {
              const kamiCodes = kami.kami_code.split('\n').filter((code) => code.trim())
              return (
                <div key={kami.order_no} className="p-4 rounded-xl border bg-dark-700/30 border-dark-600/50">
                  <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="font-medium text-dark-100">{kami.product_name}</span>
                        <Badge variant="success">已发放</Badge>
                        {kamiCodes.length > 1 && <Badge variant="info">{kamiCodes.length}个卡密</Badge>}
                      </div>
                      <div className="bg-dark-800/50 rounded-lg p-3 mb-3">
                        <div className="flex items-start justify-between gap-2">
                          <div className="flex-1 space-y-2 max-h-40 overflow-y-auto">
                            {kamiCodes.map((code, index) => (
                              <div key={index} className="flex items-center gap-2">
                                {kamiCodes.length > 1 && <span className="text-dark-500 text-sm w-6 flex-shrink-0">{index + 1}.</span>}
                                <span className="font-mono text-primary-400 break-all text-sm flex-1">{code.trim()}</span>
                              </div>
                            ))}
                          </div>
                          <div className="flex flex-col gap-1 flex-shrink-0">
                            <Button size="sm" variant="ghost" onClick={() => handleCopyKami(kami.kami_code)} title="复制全部">
                              <i className="fas fa-copy" />
                            </Button>
                            {kamiCodes.length > 1 && (
                              <Button size="sm" variant="ghost" onClick={() => handleExportKami(kami)} title="导出CSV">
                                <i className="fas fa-download" />
                              </Button>
                            )}
                          </div>
                        </div>
                      </div>
                      <div className="flex flex-wrap gap-4 text-sm text-dark-400">
                        <span>
                          <i className="fas fa-receipt mr-1" />
                          订单: {kami.order_no}
                        </span>
                        <span>
                          <i className="fas fa-calendar mr-1" />
                          发放时间: {formatDateTime(kami.payment_time || kami.created_at)}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </Card>
    </div>
  )
}
