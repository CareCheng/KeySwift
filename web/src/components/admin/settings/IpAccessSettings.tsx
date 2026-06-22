'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Switch } from '@/components/ui'
import { apiGet, apiPost, apiDelete } from '@/lib/api'

// 黑名单条目类型
interface BlacklistEntry {
  ip: string
  expires_at: string
  remaining: number
}

// 白名单配置类型
interface WhitelistConfig {
  enabled: boolean
  whitelist: string[]
}

/**
 * IP 访问管理子页面
 * 职责：合并 IP 黑名单管理与 IP 白名单管理。
 * 黑名单：登录失败自动封禁的 IP 列表，支持解封与清空。
 * 白名单：可选启用，启用后仅白名单内 IP 可访问后台。
 */
export function IpAccessSettings() {
  const [blacklist, setBlacklist] = useState<BlacklistEntry[]>([])
  const [blacklistLoading, setBlacklistLoading] = useState(false)
  const [whitelistEnabled, setWhitelistEnabled] = useState(false)
  const [whitelist, setWhitelist] = useState<string[]>([])
  const [whitelistLoading, setWhitelistLoading] = useState(false)
  const [newWhitelistIP, setNewWhitelistIP] = useState('')

  const loadBlacklist = useCallback(async () => {
    setBlacklistLoading(true)
    const res = await apiGet<{ blacklist: BlacklistEntry[] }>('/api/admin/blacklist')
    if (res.success && res.blacklist) {
      setBlacklist(res.blacklist)
    }
    setBlacklistLoading(false)
  }, [])

  const loadWhitelist = useCallback(async () => {
    setWhitelistLoading(true)
    const res = await apiGet<WhitelistConfig>('/api/admin/whitelist')
    if (res.success) {
      setWhitelistEnabled(res.enabled || false)
      setWhitelist(res.whitelist || [])
    }
    setWhitelistLoading(false)
  }, [])

  useEffect(() => { loadBlacklist(); loadWhitelist() }, [loadBlacklist, loadWhitelist])

  // 从黑名单移除IP
  const handleRemoveFromBlacklist = async (ip: string) => {
    if (!confirm(`确定要将 ${ip} 从黑名单中移除吗？`)) return
    const res = await apiDelete(`/api/admin/blacklist/${encodeURIComponent(ip)}`)
    if (res.success) {
      toast.success('已从黑名单中移除')
      loadBlacklist()
    } else {
      toast.error(res.error || '移除失败')
    }
  }

  // 清空黑名单
  const handleClearBlacklist = async () => {
    if (!confirm('确定要清空所有黑名单吗？此操作不可撤销。')) return
    const res = await apiDelete('/api/admin/blacklist')
    if (res.success) {
      toast.success('已清空黑名单')
      loadBlacklist()
    } else {
      toast.error(res.error || '清空失败')
    }
  }

  // 保存白名单配置
  const handleSaveWhitelist = async (enabled: boolean, list: string[]) => {
    const res = await apiPost('/api/admin/whitelist', { enabled, whitelist: list })
    if (res.success) {
      toast.success('白名单配置已保存')
      setWhitelistEnabled(enabled)
      setWhitelist(list)
    } else {
      toast.error(res.error || '保存失败')
    }
  }

  // 切换白名单开关
  const handleToggleWhitelist = async (enabled: boolean) => {
    await handleSaveWhitelist(enabled, whitelist)
  }

  // 添加IP到白名单
  const handleAddWhitelistIP = async () => {
    const ip = newWhitelistIP.trim()
    if (!ip) {
      toast.error('请输入IP地址')
      return
    }
    // 简单的IP格式验证
    const ipRegex = /^(\d{1,3}\.){3}\d{1,3}$/
    if (!ipRegex.test(ip)) {
      toast.error('请输入有效的IP地址格式')
      return
    }
    if (whitelist.includes(ip)) {
      toast.error('该IP已在白名单中')
      return
    }
    const newList = [...whitelist, ip]
    await handleSaveWhitelist(whitelistEnabled, newList)
    setNewWhitelistIP('')
  }

  // 从白名单移除IP
  const handleRemoveWhitelistIP = async (ip: string) => {
    if (!confirm(`确定要将 ${ip} 从白名单中移除吗？`)) return
    const newList = whitelist.filter(item => item !== ip)
    await handleSaveWhitelist(whitelistEnabled, newList)
  }

  // 清空白名单
  const handleClearWhitelist = async () => {
    if (!confirm('确定要清空所有白名单吗？')) return
    await handleSaveWhitelist(whitelistEnabled, [])
  }

  // 格式化剩余时间
  const formatRemaining = (seconds: number) => {
    if (seconds < 60) return `${seconds}秒`
    if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`
    return `${Math.floor(seconds / 3600)}时${Math.floor((seconds % 3600) / 60)}分`
  }

  return (
    <div className="space-y-4">
      <Card title="IP黑名单管理">
        <div className="space-y-4">
          <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
            连续登录失败10次的IP会被自动加入临时黑名单30分钟。您可以在此查看和管理被封禁的IP。
          </p>

          <div className="flex justify-between items-center">
            <div className="flex items-center gap-2">
              <Button variant="secondary" onClick={loadBlacklist} disabled={blacklistLoading}>
                {blacklistLoading ? <i className="fas fa-spinner fa-spin mr-2" /> : <i className="fas fa-sync-alt mr-2" />}
                刷新
              </Button>
              <span className="text-sm" style={{ color: 'var(--text-muted)' }}>
                共 {blacklist.length} 个IP被封禁
              </span>
            </div>
            {blacklist.length > 0 && (
              <Button variant="danger" onClick={handleClearBlacklist}>
                <i className="fas fa-trash-alt mr-2" />清空全部
              </Button>
            )}
          </div>

          {blacklist.length === 0 ? (
            <div className="text-center py-8" style={{ color: 'var(--text-muted)' }}>
              <i className="fas fa-shield-alt text-4xl mb-3 opacity-50" />
              <p>暂无被封禁的IP</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b" style={{ borderColor: 'var(--border-color)' }}>
                    <th className="text-left py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>IP地址</th>
                    <th className="text-left py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>过期时间</th>
                    <th className="text-left py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>剩余时间</th>
                    <th className="text-right py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>操作</th>
                  </tr>
                </thead>
                <tbody>
                  {blacklist.map((entry) => (
                    <tr key={entry.ip} className="border-b" style={{ borderColor: 'var(--border-color)' }}>
                      <td className="py-3 px-4">
                        <code className="px-2 py-1 rounded text-sm" style={{ backgroundColor: 'var(--bg-tertiary)' }}>
                          {entry.ip}
                        </code>
                      </td>
                      <td className="py-3 px-4 text-sm" style={{ color: 'var(--text-secondary)' }}>
                        {entry.expires_at}
                      </td>
                      <td className="py-3 px-4">
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-500/20 text-red-400">
                          <i className="fas fa-clock mr-1" />{formatRemaining(entry.remaining)}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-right">
                        <Button variant="secondary" size="sm" onClick={() => handleRemoveFromBlacklist(entry.ip)}>
                          <i className="fas fa-unlock mr-1" />解封
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </Card>

      <Card title="IP白名单管理">
        <div className="space-y-4">
          <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
            启用后，只有白名单中的IP才能访问管理后台。请确保将您当前的IP添加到白名单中，否则可能无法访问。
          </p>

          <Switch
            checked={whitelistEnabled}
            onChange={handleToggleWhitelist}
            label="启用IP白名单"
            description="开启后只有白名单中的IP可以访问管理后台"
          />

          {whitelistEnabled && (
            <div className="p-4 bg-dark-700/30 rounded-lg space-y-4">
              <div className="flex gap-2">
                <Input
                  placeholder="输入IP地址，如 192.168.1.1"
                  value={newWhitelistIP}
                  onChange={(e) => setNewWhitelistIP(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleAddWhitelistIP()}
                />
                <Button onClick={handleAddWhitelistIP} disabled={whitelistLoading}>
                  <i className="fas fa-plus mr-2" />添加
                </Button>
              </div>

              <div className="flex justify-between items-center">
                <div className="flex items-center gap-2">
                  <Button variant="secondary" onClick={loadWhitelist} disabled={whitelistLoading}>
                    {whitelistLoading ? <i className="fas fa-spinner fa-spin mr-2" /> : <i className="fas fa-sync-alt mr-2" />}
                    刷新
                  </Button>
                  <span className="text-sm" style={{ color: 'var(--text-muted)' }}>
                    共 {whitelist.length} 个IP
                  </span>
                </div>
                {whitelist.length > 0 && (
                  <Button variant="danger" onClick={handleClearWhitelist}>
                    <i className="fas fa-trash-alt mr-2" />清空全部
                  </Button>
                )}
              </div>

              {whitelist.length === 0 ? (
                <div className="text-center py-8" style={{ color: 'var(--text-muted)' }}>
                  <i className="fas fa-list text-4xl mb-3 opacity-50" />
                  <p>白名单为空，请添加允许访问的IP</p>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b" style={{ borderColor: 'var(--border-color)' }}>
                        <th className="text-left py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>IP地址</th>
                        <th className="text-right py-3 px-4 text-sm font-medium" style={{ color: 'var(--text-muted)' }}>操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {whitelist.map((ip) => (
                        <tr key={ip} className="border-b" style={{ borderColor: 'var(--border-color)' }}>
                          <td className="py-3 px-4">
                            <code className="px-2 py-1 rounded text-sm" style={{ backgroundColor: 'var(--bg-tertiary)' }}>
                              {ip}
                            </code>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <Button variant="secondary" size="sm" onClick={() => handleRemoveWhitelistIP(ip)}>
                              <i className="fas fa-trash-alt mr-1" />移除
                            </Button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}
        </div>
      </Card>
    </div>
  )
}
