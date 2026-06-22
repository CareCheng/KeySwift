'use client'

import { useState, useEffect, useCallback } from 'react'
import { Button, Card, Input, Badge } from '@/components/ui'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost } from '@/lib/api'
import toast from 'react-hot-toast'

/**
 * 日志条目接口
 */
interface LogEntry {
  id: number
  user_type: string
  user_id: number
  username: string
  action: string
  category: string
  target: string
  target_id: string
  detail: string
  ip: string
  user_agent: string
  created_at: string
}

/**
 * 日志配置接口
 */
interface LogConfig {
  enable_user_log: boolean
  enable_admin_log: boolean
}

/**
 * 操作日志页面组件
 * 支持按日期查询加密存储的日志文件
 */
export function LogsPage() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [total, setTotal] = useState(0)
  const [availableDates, setAvailableDates] = useState<string[]>([])
  const [selectedDate, setSelectedDate] = useState<string>('')
  const [userTypeFilter, setUserTypeFilter] = useState<string>('')
  const [categoryFilter, setCategoryFilter] = useState<string>('')
  
  // 日志配置
  const [logConfig, setLogConfig] = useState<LogConfig>({ enable_user_log: false, enable_admin_log: false })
  const [configLoading, setConfigLoading] = useState(false)
  const [showConfig, setShowConfig] = useState(false)

  // 加载日志配置
  const loadLogConfig = useCallback(async () => {
    const res = await apiGet<{ config: LogConfig }>('/api/admin/logs/config')
    if (res.success && res.config) {
      setLogConfig(res.config)
    }
  }, [])

  // 加载可用的日志日期列表
  const loadAvailableDates = useCallback(async () => {
    const res = await apiGet<{ dates: string[] }>('/api/admin/logs/dates')
    if (res.success && res.dates) {
      setAvailableDates(res.dates)
      // 默认选择最新日期（今天）
      if (res.dates.length > 0 && !selectedDate) {
        setSelectedDate(res.dates[0])
      }
    }
  }, [selectedDate])

  // 加载日志数据
  const loadLogs = useCallback(async () => {
    setLoading(true)
    let url = `/api/admin/logs?page=${page}&page_size=50`
    if (selectedDate) {
      url += `&date=${selectedDate}`
    }
    if (userTypeFilter) {
      url += `&user_type=${userTypeFilter}`
    }
    if (categoryFilter) {
      url += `&category=${categoryFilter}`
    }
    
    const res = await apiGet<{ logs: LogEntry[]; total_pages: number; total: number }>(url)
    if (res.success) {
      setLogs(res.logs || [])
      setTotalPages(res.total_pages || 1)
      setTotal(res.total || 0)
    }
    setLoading(false)
  }, [page, selectedDate, userTypeFilter, categoryFilter])

  // 初始化加载
  useEffect(() => {
    loadAvailableDates()
    loadLogConfig()
  }, [loadAvailableDates, loadLogConfig])

  // 当日期或过滤条件变化时重新加载日志
  useEffect(() => {
    if (selectedDate) {
      loadLogs()
    }
  }, [loadLogs, selectedDate])

  // 日期变化时重置页码
  const handleDateChange = (date: string) => {
    setSelectedDate(date)
    setPage(1)
  }

  // 用户类型过滤变化
  const handleUserTypeChange = (type: string) => {
    setUserTypeFilter(type)
    setPage(1)
  }

  // 分类过滤变化
  const handleCategoryChange = (category: string) => {
    setCategoryFilter(category)
    setPage(1)
  }

  // 保存日志配置
  const saveLogConfig = async () => {
    setConfigLoading(true)
    const res = await apiPost('/api/admin/logs/config', {
      enable_user_log: logConfig.enable_user_log,
      enable_admin_log: logConfig.enable_admin_log,
    })
    setConfigLoading(false)
    if (res.success) {
      toast.success('日志配置已保存')
      setShowConfig(false)
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  // 获取用户类型显示文本
  const getUserTypeLabel = (type: string) => {
    switch (type) {
      case 'admin': return '管理员'
      case 'user': return '用户'
      case 'security': return '安全事件'
      default: return type
    }
  }

  // 获取用户类型样式
  const getUserTypeStyle = (type: string) => {
    switch (type) {
      case 'admin': return 'bg-yellow-500/20 text-yellow-400'
      case 'user': return 'bg-blue-500/20 text-blue-400'
      case 'security': return 'bg-red-500/20 text-red-400'
      default: return 'bg-gray-500/20 text-gray-400'
    }
  }

  // 获取分类显示文本
  const getCategoryLabel = (category: string) => {
    const labels: Record<string, string> = {
      'auth': '认证',
      'product': '商品',
      'order': '订单',
      'user': '用户',
      'system': '系统',
      'payment': '支付',
      'security': '安全',
    }
    return labels[category] || category
  }

  // 获取分类样式
  const getCategoryStyle = (category: string) => {
    const styles: Record<string, string> = {
      'auth': 'bg-purple-500/20 text-purple-400',
      'product': 'bg-green-500/20 text-green-400',
      'order': 'bg-blue-500/20 text-blue-400',
      'user': 'bg-cyan-500/20 text-cyan-400',
      'system': 'bg-orange-500/20 text-orange-400',
      'payment': 'bg-emerald-500/20 text-emerald-400',
      'security': 'bg-red-500/20 text-red-400',
    }
    return styles[category] || 'bg-gray-500/20 text-gray-400'
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium text-dark-100">操作日志</h2>
        <div className="flex items-center gap-3">
          <div className="text-sm text-dark-400">
            日志使用 AES-256-GCM 加密存储
          </div>
          <Button size="sm" variant="secondary" onClick={() => setShowConfig(!showConfig)}>
            <i className="fas fa-cog mr-1" />
            配置
          </Button>
        </div>
      </div>

      {/* 日志配置面板 */}
      {showConfig && (
        <Card className="p-4 border-primary-500/30">
          <h3 className="text-dark-100 font-medium mb-4">日志开关配置</h3>
          <div className="flex flex-wrap gap-6 items-center">
            <Toggle
              checked={logConfig.enable_user_log}
              onChange={(checked) => setLogConfig({ ...logConfig, enable_user_log: checked })}
              label="启用用户端日志"
            />
            <span className="text-dark-500 text-xs -ml-4">（登录、注册、购买等）</span>
            <Toggle
              checked={logConfig.enable_admin_log}
              onChange={(checked) => setLogConfig({ ...logConfig, enable_admin_log: checked })}
              label="启用管理端日志"
            />
            <span className="text-dark-500 text-xs -ml-4">（商品管理、订单操作等）</span>
            <Button size="sm" onClick={saveLogConfig} loading={configLoading}>
              保存配置
            </Button>
          </div>
          <p className="text-dark-500 text-xs mt-3">
            <i className="fas fa-info-circle mr-1" />
            安全事件日志始终记录，不受开关控制
          </p>
        </Card>
      )}

      {/* 筛选条件 */}
      <Card className="p-4">
        <div className="flex flex-wrap gap-4 items-center">
          {/* 日期选择 */}
          <div className="flex items-center gap-2">
            <label className="text-dark-400 text-sm">日期：</label>
            {availableDates.length > 0 ? (
              <select
                value={selectedDate}
                onChange={(e) => handleDateChange(e.target.value)}
                className="bg-dark-700 border border-dark-600 rounded px-3 py-1.5 text-dark-100 text-sm focus:outline-none focus:border-primary-500"
              >
                {availableDates.map((date) => (
                  <option key={date} value={date}>
                    {date}
                  </option>
                ))}
              </select>
            ) : (
              <Input
                type="date"
                value={selectedDate}
                onChange={(e) => handleDateChange(e.target.value)}
                className="w-40"
              />
            )}
          </div>

          {/* 用户类型过滤 */}
          <div className="flex items-center gap-2">
            <label className="text-dark-400 text-sm">来源：</label>
            <select
              value={userTypeFilter}
              onChange={(e) => handleUserTypeChange(e.target.value)}
              className="bg-dark-700 border border-dark-600 rounded px-3 py-1.5 text-dark-100 text-sm focus:outline-none focus:border-primary-500"
            >
              <option value="">全部</option>
              <option value="admin">管理员</option>
              <option value="user">用户</option>
              <option value="security">安全事件</option>
            </select>
          </div>

          {/* 分类过滤 */}
          <div className="flex items-center gap-2">
            <label className="text-dark-400 text-sm">分类：</label>
            <select
              value={categoryFilter}
              onChange={(e) => handleCategoryChange(e.target.value)}
              className="bg-dark-700 border border-dark-600 rounded px-3 py-1.5 text-dark-100 text-sm focus:outline-none focus:border-primary-500"
            >
              <option value="">全部</option>
              <option value="auth">认证</option>
              <option value="product">商品</option>
              <option value="order">订单</option>
              <option value="user">用户</option>
              <option value="system">系统</option>
              <option value="payment">支付</option>
              <option value="security">安全</option>
            </select>
          </div>

          {/* 刷新按钮 */}
          <Button size="sm" variant="secondary" onClick={loadLogs}>
            <i className="fas fa-sync-alt mr-1" />
            刷新
          </Button>

          {/* 统计信息 */}
          <div className="ml-auto text-sm text-dark-400">
            共 {total} 条记录
          </div>
        </div>
      </Card>

      {/* 日志列表 */}
      <Card>
        {loading ? (
          <div className="text-center py-12">
            <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
          </div>
        ) : logs.length === 0 ? (
          <div className="text-center py-12 text-dark-500">
            {selectedDate ? `${selectedDate} 暂无日志记录` : '请选择日期查看日志'}
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-dark-700">
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">时间</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">来源</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">分类</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">用户</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">操作</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">目标</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">详情</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">IP</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => (
                    <tr key={log.id} className="border-b border-dark-700/50 hover:bg-dark-700/30">
                      <td className="py-3 px-4 text-dark-300 text-sm whitespace-nowrap">
                        {log.created_at}
                      </td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded text-xs ${getUserTypeStyle(log.user_type)}`}>
                          {getUserTypeLabel(log.user_type)}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        {log.category && (
                          <span className={`px-2 py-1 rounded text-xs ${getCategoryStyle(log.category)}`}>
                            {getCategoryLabel(log.category)}
                          </span>
                        )}
                      </td>
                      <td className="py-3 px-4 text-dark-100">{log.username}</td>
                      <td className="py-3 px-4 text-dark-300">{log.action}</td>
                      <td className="py-3 px-4 text-dark-300">
                        {log.target}
                        {log.target_id ? ` #${log.target_id}` : ''}
                      </td>
                      <td className="py-3 px-4 text-dark-400 text-sm max-w-xs truncate" title={log.detail}>
                        {log.detail || '-'}
                      </td>
                      <td className="py-3 px-4 text-dark-300 font-mono text-sm">{log.ip}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* 分页 */}
            {totalPages > 1 && (
              <div className="flex justify-center gap-2 mt-4 pb-4">
                <Button size="sm" variant="secondary" disabled={page <= 1} onClick={() => setPage(1)}>
                  首页
                </Button>
                <Button size="sm" variant="secondary" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                  上一页
                </Button>
                <span className="px-4 py-2 text-dark-400">
                  第 {page} / {totalPages} 页
                </span>
                <Button size="sm" variant="secondary" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
                  下一页
                </Button>
                <Button size="sm" variant="secondary" disabled={page >= totalPages} onClick={() => setPage(totalPages)}>
                  末页
                </Button>
              </div>
            )}
          </>
        )}
      </Card>
    </div>
  )
}
