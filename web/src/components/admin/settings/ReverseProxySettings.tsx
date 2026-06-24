'use client'

import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { apiGet, apiPost } from '@/lib/api'
import { Button, Card, Input, Switch } from '@/components/ui'
import { ReverseProxyConfig, ReverseProxyDiagnostics } from '../types'

const defaultConfig: ReverseProxyConfig = {
  public_base_url: '',
  reverse_proxy_enabled: false,
  trusted_proxies: [],
  client_ip_header: 'X-Forwarded-For',
  real_ip_header: 'X-Real-IP',
  proto_header: 'X-Forwarded-Proto',
  host_header: 'X-Forwarded-Host',
  port_header: 'X-Forwarded-Port',
  cookie_secure_mode: 'auto',
  cookie_domain: '',
  app_base_path: '/',
  cors_enabled: false,
  cors_allow_origins: [],
  cors_allow_credentials: true,
  hsts_enabled: false,
}

type ReverseProxyResponse = {
  config: ReverseProxyConfig
}

type DiagnosticsResponse = {
  diagnostics: ReverseProxyDiagnostics
}

/**
 * 访问与代理设置页面
 * 职责：维护反向代理、公网访问地址、可信代理、Cookie/CORS/HSTS 和当前请求诊断。
 */
export function ReverseProxySettings() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [config, setConfig] = useState<ReverseProxyConfig>(defaultConfig)
  const [trustedProxiesText, setTrustedProxiesText] = useState('')
  const [corsOriginsText, setCorsOriginsText] = useState('')
  const [diagnostics, setDiagnostics] = useState<ReverseProxyDiagnostics | null>(null)

  const loadConfig = async () => {
    setLoading(true)
    const res = await apiGet<ReverseProxyResponse>('/api/admin/settings/reverse-proxy')
    if (res.success && res.config) {
      setConfig({ ...defaultConfig, ...res.config })
      setTrustedProxiesText((res.config.trusted_proxies || []).join('\n'))
      setCorsOriginsText((res.config.cors_allow_origins || []).join('\n'))
    } else {
      toast.error(res.error || '获取反向代理配置失败')
    }
    setLoading(false)
  }

  const loadDiagnostics = async () => {
    const res = await apiGet<DiagnosticsResponse>('/api/admin/settings/reverse-proxy/diagnostics')
    if (res.success && res.diagnostics) {
      setDiagnostics(res.diagnostics)
    } else {
      toast.error(res.error || '获取代理诊断失败')
    }
  }

  useEffect(() => {
    loadConfig()
    loadDiagnostics()
  }, [])

  const updateConfig = <K extends keyof ReverseProxyConfig>(key: K, value: ReverseProxyConfig[K]) => {
    setConfig((current) => ({ ...current, [key]: value }))
  }

  const handleSave = async () => {
    setSaving(true)
    const payload: ReverseProxyConfig = {
      ...config,
      trusted_proxies: splitLines(trustedProxiesText),
      cors_allow_origins: splitLines(corsOriginsText),
      app_base_path: '/',
    }
    const res = await apiPost<ReverseProxyResponse & { need_restart?: boolean; message?: string }>(
      '/api/admin/settings/reverse-proxy',
      payload as unknown as Record<string, unknown>
    )
    setSaving(false)

    if (res.success && res.config) {
      setConfig({ ...defaultConfig, ...res.config })
      setTrustedProxiesText((res.config.trusted_proxies || []).join('\n'))
      setCorsOriginsText((res.config.cors_allow_origins || []).join('\n'))
      await loadDiagnostics()
      toast.success(res.need_restart ? '配置已保存，建议重启程序同步框架级可信代理' : '配置已保存')
    } else {
      toast.error(res.error || '保存反向代理配置失败')
    }
  }

  if (loading) {
    return (
      <Card title="访问与代理">
        <div className="py-8 text-center text-dark-400">
          <i className="fas fa-spinner fa-spin mr-2" />
          正在加载配置...
        </div>
      </Card>
    )
  }

  return (
    <div className="space-y-4">
      <Card title="反向代理模式">
        <div className="space-y-4">
          <Switch
            checked={config.reverse_proxy_enabled}
            onChange={(checked) => updateConfig('reverse_proxy_enabled', checked)}
            label="启用反向代理可信头解析"
            description="默认关闭。仅当程序部署在可信 Nginx、Caddy、网关或负载均衡之后时开启。"
          />

          <Input
            label="公网访问地址"
            placeholder="https://example.com"
            value={config.public_base_url}
            onChange={(e) => updateConfig('public_base_url', e.target.value)}
          />
          <p className="text-xs text-dark-500">用于支付结果页、外部链接和 HTTPS 识别。当前版本只支持根路径，不填写子路径。</p>

          <div>
            <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>可信代理 IP / CIDR</label>
            <textarea
              className="input min-h-28 w-full py-3"
              value={trustedProxiesText}
              onChange={(e) => setTrustedProxiesText(e.target.value)}
              placeholder={'127.0.0.1\n::1\n172.16.0.0/12'}
            />
            <p className="text-xs text-dark-500 mt-1">每行一个 IP 或 CIDR。只有这些来源发来的代理头会被信任。</p>
          </div>
        </div>
      </Card>

      <Card title="代理请求头">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Input label="真实 IP 多级头" value={config.client_ip_header} onChange={(e) => updateConfig('client_ip_header', e.target.value)} />
          <Input label="真实 IP 单级头" value={config.real_ip_header} onChange={(e) => updateConfig('real_ip_header', e.target.value)} />
          <Input label="外部协议头" value={config.proto_header} onChange={(e) => updateConfig('proto_header', e.target.value)} />
          <Input label="外部 Host 头" value={config.host_header} onChange={(e) => updateConfig('host_header', e.target.value)} />
          <Input label="外部端口头" value={config.port_header} onChange={(e) => updateConfig('port_header', e.target.value)} />
          <Input label="Cookie 域名" placeholder=".example.com" value={config.cookie_domain} onChange={(e) => updateConfig('cookie_domain', e.target.value)} />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
          <label className="space-y-1.5">
            <span className="block text-sm font-medium" style={{ color: 'var(--text-secondary)' }}>Cookie Secure 策略</span>
            <select
              className="input h-12 w-full"
              value={config.cookie_secure_mode}
              onChange={(e) => updateConfig('cookie_secure_mode', e.target.value as ReverseProxyConfig['cookie_secure_mode'])}
            >
              <option value="auto">自动：根据外部 HTTPS 判断</option>
              <option value="always">始终启用 Secure</option>
              <option value="never">永不启用 Secure</option>
            </select>
          </label>

          <Switch
            checked={config.hsts_enabled}
            onChange={(checked) => updateConfig('hsts_enabled', checked)}
            label="启用 HSTS"
            description="仅外部 HTTPS 请求返回。首次部署证书未稳定前建议保持关闭。"
          />

          <Input label="应用挂载路径" value="/" disabled />
        </div>
      </Card>

      <Card title="CORS 高级配置">
        <div className="space-y-4">
          <Switch
            checked={config.cors_enabled}
            onChange={(checked) => updateConfig('cors_enabled', checked)}
            label="启用 CORS"
            description="同域根路径部署不需要开启。前后端分域时才需要配置。"
          />
          <Switch
            checked={config.cors_allow_credentials}
            onChange={(checked) => updateConfig('cors_allow_credentials', checked)}
            label="允许携带 Cookie"
            description="开启后不允许 CORS 来源使用 *。"
          />
          <div>
            <label className="block text-sm font-medium mb-2" style={{ color: 'var(--text-secondary)' }}>允许的 CORS 来源</label>
            <textarea
              className="input min-h-24 w-full py-3"
              value={corsOriginsText}
              onChange={(e) => setCorsOriginsText(e.target.value)}
              placeholder={'https://web.example.com\nhttps://admin.example.com'}
              disabled={!config.cors_enabled}
            />
            <p className="text-xs text-dark-500 mt-1">每行一个来源，只填写协议、域名和端口，不填写路径。</p>
          </div>
        </div>
      </Card>

      <Card title="当前请求诊断">
        <div className="space-y-3">
          <div className="flex justify-between items-center">
            <p className="text-sm text-dark-400">用于确认后台当前请求是否来自可信代理，以及程序识别到的真实 IP 和外部访问地址。</p>
            <Button variant="secondary" onClick={loadDiagnostics}>刷新诊断</Button>
          </div>
          {diagnostics ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
              <InfoItem label="真实客户端 IP" value={diagnostics.client_ip.client_ip || '-'} />
              <InfoItem label="直接连接 IP" value={diagnostics.client_ip.remote_ip || '-'} />
              <InfoItem label="IP 来源" value={diagnostics.client_ip.source || '-'} />
              <InfoItem label="可信代理" value={diagnostics.client_ip.trusted_proxy ? '是' : '否'} />
              <InfoItem label="外部协议" value={diagnostics.external_access.scheme || '-'} />
              <InfoItem label="外部 Host" value={diagnostics.external_access.host || '-'} />
              <InfoItem label="外部 BaseURL" value={diagnostics.external_access.base_url || '-'} />
              <InfoItem label="应用路径" value={diagnostics.external_access.path_prefix || '/'} />
            </div>
          ) : (
            <p className="text-sm text-dark-500">暂无诊断数据</p>
          )}
        </div>
      </Card>

      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={saving}>
          {saving ? <><i className="fas fa-spinner fa-spin mr-2" />保存中...</> : '保存访问与代理设置'}
        </Button>
      </div>
    </div>
  )
}

function splitLines(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-dark-700/60 bg-dark-800/40 p-3">
      <div className="text-xs text-dark-500">{label}</div>
      <div className="mt-1 break-all" style={{ color: 'var(--text-secondary)' }}>{value}</div>
    </div>
  )
}
