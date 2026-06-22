'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Input, Modal } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { DBConfig } from './types'

export function DatabasePage() {
  const [config, setConfig] = useState<DBConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState({
    type: 'sqlite', host: 'localhost', port: '3306', user: '', password: '', database: 'user_data.db'
  })
  // 密钥重置相关状态
  const [showResetModal, setShowResetModal] = useState(false)
  const [resetKeyLength, setResetKeyLength] = useState('256')
  const [resetConfirmText, setResetConfirmText] = useState('')
  const [resetLoading, setResetLoading] = useState(false)

  const loadConfig = useCallback(async () => {
    const res = await apiGet<{ config: DBConfig }>('/api/admin/db/config')
    if (res.success && res.config) {
      setConfig(res.config)
      setForm({
        type: res.config.type || 'sqlite', host: res.config.host || 'localhost',
        port: String(res.config.port || 3306), user: res.config.user || '',
        password: '', database: res.config.database || 'user_data.db'
      })
    }
    setLoading(false)
  }, [])

  useEffect(() => { loadConfig() }, [loadConfig])

  const handleTest = async () => {
    toast.loading('正在测试连接...')
    const data = {
      type: form.type, host: form.host, port: parseInt(form.port) || 3306,
      user: form.user, password: form.password, database: form.database
    }
    const res = await apiPost('/api/admin/db/test', data)
    toast.dismiss()
    if (res.success) toast.success('连接成功！')
    else toast.error(res.error || '连接失败')
  }

  const handleSave = async () => {
    const data: Record<string, unknown> = {
      type: form.type, host: form.host, port: parseInt(form.port) || 3306,
      user: form.user, database: form.database
    }
    if (form.password) data.password = form.password
    const res = await apiPost('/api/admin/db/config', data)
    if (res.success) { toast.success('配置已保存，请重启程序生效'); loadConfig() }
    else toast.error(res.error || '保存失败')
  }

  const copyKey = () => {
    if (config?.encryption_key) {
      navigator.clipboard.writeText(config.encryption_key)
      toast.success('密钥已复制到剪贴板')
    }
  }

  // 重置加密密钥
  const handleResetKey = async () => {
    if (resetConfirmText !== '我确认重置密钥并了解数据将永久丢失') {
      toast.error('请输入正确的确认文字')
      return
    }

    setResetLoading(true)
    const res = await apiPost<{ encryption_key: string; key_length: number }>('/api/admin/db/reset-key', {
      key_length: parseInt(resetKeyLength),
      confirm: 'RESET_KEY',
      confirm_text: resetConfirmText
    })
    setResetLoading(false)

    if (res.success) {
      toast.success('密钥已重置')
      setShowResetModal(false)
      setResetConfirmText('')
      loadConfig()
    } else {
      toast.error(res.error || '重置失败')
    }
  }


  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium text-dark-100">数据库配置</h2>
      <Card>
        <div className={`p-4 rounded-lg mb-4 ${config?.connected ? 'bg-green-500/10 border border-green-500/20' : 'bg-yellow-500/10 border border-yellow-500/20'}`}>
          <p className={config?.connected ? 'text-green-400' : 'text-yellow-400'}>
            {config?.connected ? '✅ 数据库已连接' : '⚠️ 数据库未连接'}
          </p>
        </div>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">数据库类型</label>
            <select className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100" value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}>
              <option value="sqlite">SQLite</option>
              <option value="mysql">MySQL</option>
              <option value="postgres">PostgreSQL</option>
            </select>
          </div>
          {form.type !== 'sqlite' && (
            <>
              <div className="grid grid-cols-2 gap-4">
                <Input label="主机" value={form.host} onChange={(e) => setForm({ ...form, host: e.target.value })} />
                <Input label="端口" type="number" value={form.port} onChange={(e) => setForm({ ...form, port: e.target.value })} />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <Input label="用户名" value={form.user} onChange={(e) => setForm({ ...form, user: e.target.value })} />
                <Input label="密码" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} placeholder="留空保持不变" />
              </div>
            </>
          )}
          <Input label="数据库名/文件路径" value={form.database} onChange={(e) => setForm({ ...form, database: e.target.value })} />
          <div className="flex gap-2">
            <Button variant="secondary" onClick={handleTest}>测试连接</Button>
            <Button onClick={handleSave}>保存配置</Button>
          </div>
        </div>
      </Card>

      <Card title="🔐 数据加密密钥">
        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg mb-4">
          <p className="text-blue-400 text-sm">此密钥用于加密数据库中的敏感数据。密钥在首次启动时自动生成，请妥善保管以便更换环境或恢复数据。</p>
        </div>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">当前密钥长度</label>
            <input type="text" value={`${config?.key_length || 256} 位`} readOnly className="w-full px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-400 cursor-not-allowed" />
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">加密密钥（Base64编码）</label>
            <div className="flex gap-2">
              <input type="text" value={config?.encryption_key || '未生成'} readOnly className="flex-1 px-3 py-2 bg-dark-700/50 border border-dark-600 rounded-lg text-dark-400 font-mono text-sm cursor-not-allowed" />
              <Button variant="secondary" onClick={copyKey}>复制</Button>
            </div>
            <p className="text-dark-500 text-xs mt-1">更换运行环境或恢复数据时需要使用相同的密钥才能解密数据</p>
          </div>
          <div className="pt-4 border-t border-dark-700">
            <Button variant="danger" onClick={() => setShowResetModal(true)}>
              <i className="fas fa-exclamation-triangle mr-2" />重置密钥
            </Button>
            <p className="text-dark-500 text-xs mt-2">⚠️ 重置密钥后，之前加密的数据将无法解密，请谨慎操作</p>
          </div>
        </div>
      </Card>

      {/* 重置密钥确认弹窗 */}
      <Modal isOpen={showResetModal} onClose={() => { setShowResetModal(false); setResetConfirmText('') }} title="⚠️ 重置加密密钥">
        <div className="space-y-4">
          <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
            <p className="text-red-400 text-sm font-medium mb-2">危险操作警告</p>
            <ul className="text-red-400/80 text-sm space-y-1 list-disc list-inside">
              <li>重置密钥后，所有使用旧密钥加密的数据将<strong>永久无法解密</strong></li>
              <li>数据库连接密码等敏感配置将丢失</li>
              <li>需要重新配置数据库连接信息</li>
              <li>此操作<strong>不可撤销</strong></li>
            </ul>
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">新密钥长度</label>
            <select
              value={resetKeyLength}
              onChange={(e) => setResetKeyLength(e.target.value)}
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100"
            >
              <option value="128">128 位</option>
              <option value="192">192 位</option>
              <option value="256">256 位（推荐）</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-1">
              请输入确认文字：<span className="text-red-400">我确认重置密钥并了解数据将永久丢失</span>
            </label>
            <input
              type="text"
              value={resetConfirmText}
              onChange={(e) => setResetConfirmText(e.target.value)}
              placeholder="请输入上方红色文字"
              className="w-full px-3 py-2 bg-dark-700 border border-dark-600 rounded-lg text-dark-100"
            />
          </div>
          <div className="flex gap-2 justify-end pt-2">
            <Button variant="secondary" onClick={() => { setShowResetModal(false); setResetConfirmText('') }}>取消</Button>
            <Button
              variant="danger"
              onClick={handleResetKey}
              disabled={resetLoading || resetConfirmText !== '我确认重置密钥并了解数据将永久丢失'}
            >
              {resetLoading ? '重置中...' : '确认重置'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
