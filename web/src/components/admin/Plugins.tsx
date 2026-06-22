'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import toast from 'react-hot-toast'
import { Badge, Button, Card, Input } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { cn, formatDateTime } from '@/lib/utils'
import { PLUGIN_FRONTEND_CHANGED_EVENT } from '@/lib/pluginRegistry'

type BadgeVariant = 'success' | 'warning' | 'danger' | 'info' | 'default'
type DetailTab = 'overview' | 'bindings' | 'migrations' | 'configs' | 'manifest'

interface PluginSummary {
  plugin_root?: string
  protocol_version?: string
  backend_protocol?: string
  frontend_protocol?: string
  plugins?: number
  frontend_pages?: number
  frontend_menus?: number
  themes?: number
}

interface PluginListItem {
  id: string
  version: string
  plugin_kind: string
  display_name: string
  description: string
  author: string
  runtime_state: string
  traffic_enabled: boolean
  enabled: boolean
  desired_state: string
  lifecycle_state: string
  verify_status: string
  health_status: string
  frontend_enabled: boolean
  theme_enabled: boolean
  permissions: number
  pages: number
  menus: number
  routes: number
  events: number
  jobs: number
  migrations: number
}

interface PluginRegistryRecord {
  plugin_id?: string
  install_id?: string
  current_version?: string
  install_root?: string
  source_type?: string
  enabled?: boolean
  autostart?: boolean
  desired_state?: string
  actual_state?: string
  lifecycle_state?: string
  trust_level?: string
  signature_status?: string
  current_manifest_hash?: string
  last_verified_at?: string
  last_verify_status?: string
  tamper_status?: string
  quarantine_reason?: string
  config_version?: number
  selected_os?: string
  selected_arch?: string
  health_status?: string
  last_start_at?: string
  last_ready_at?: string
  last_stop_at?: string
  last_fault_at?: string
  manifest_json?: string
  updated_at?: string
}

interface RuntimePlugin {
  sessionId?: string
  pluginId?: string
  installId?: string
  instanceId?: string
  pid?: number
  state?: string
  trafficEnabled?: boolean
  selectedProtocolVersion?: string
  controlTransport?: string
  dataTransport?: string
  channelEndpoint?: string
  trustLevel?: string
  integrityState?: string
  configVersion?: number
  startedAt?: string
  readyAt?: string
  lastHeartbeatAt?: string
  loadSummary?: string
  recentErrorCount?: number
  recentLatencyMs?: number
}

interface PluginManifest {
  manifestVersion?: string
  id?: string
  version?: string
  pluginKind?: string
  identity?: {
    name?: string
    displayName?: string
    description?: string
    author?: string
    homepage?: string
    license?: string
  }
  compatibility?: Record<string, unknown>
  package?: Record<string, unknown>
  integrity?: Record<string, unknown>
  lifecycle?: Record<string, unknown>
  dependencies?: Record<string, unknown>
  permissions?: Array<Record<string, unknown>>
  capabilities?: Record<string, unknown>
  backend?: {
    controlProtocol?: string
    dataProtocol?: string
    settingsRef?: string
    routes?: Array<Record<string, unknown>>
    events?: Array<Record<string, unknown>>
    jobs?: Array<Record<string, unknown>>
    migrations?: Array<Record<string, unknown>>
  }
  frontend?: {
    enabled?: boolean
    renderMode?: string
    mountAreas?: string[]
    pages?: Array<Record<string, unknown>>
    menus?: Array<Record<string, unknown>>
    forms?: Array<Record<string, unknown>>
    views?: Array<Record<string, unknown>>
  }
  ui?: Record<string, unknown>
  observability?: Record<string, unknown>
  operations?: Record<string, unknown>
  metadata?: Record<string, unknown>
}

interface PluginDetail {
  manifest?: PluginManifest
  runtime?: RuntimePlugin
  registry?: PluginRegistryRecord | null
}

interface PluginBinding {
  id: number
  binding_type: string
  binding_key: string
  target_scope: string
  mount_area: string
  route_or_view_id: string
  enabled: boolean
  order_hint: number
  permission_guard: string
  updated_at: string
}

interface PluginMigration {
  id: number
  migration_id: string
  version: string
  direction: string
  path: string
  checksum: string
  status: string
  executed_at?: string
  error_message?: string
  updated_at: string
}

interface PluginConfigRecord {
  id: number
  config_key: string
  config_version: number
  schema_json: string
  value_json: string
  encrypted_fields: string
  enabled: boolean
  updated_by: string
  updated_at: string
}

interface ConfigSchema {
  schemaVersion?: string
  pluginId?: string
  configVersion?: string
  sections?: Array<{
    id?: string
    title?: string
    description?: string
    scope?: string
    fields?: Array<{
      id?: string
      key?: string
      label?: string
      type?: string
      required?: boolean
      secret?: boolean
      description?: string
    }>
  }>
  defaultsRef?: string
  secretPolicies?: string[]
  validationRules?: string[]
  reloadPolicies?: string[]
  permissionGuards?: string[]
}

interface DiscoveryResult {
  pluginId?: string
  version?: string
  manifestPath?: string
  installRoot?: string
  errors?: string[]
}

interface DetailBundle {
  detail: PluginDetail | null
  bindings: PluginBinding[]
  migrations: PluginMigration[]
  configs: PluginConfigRecord[]
  schemas: ConfigSchema[]
}

const DETAIL_TABS: Array<{ key: DetailTab; label: string }> = [
  { key: 'overview', label: '概览' },
  { key: 'bindings', label: '绑定' },
  { key: 'migrations', label: '迁移' },
  { key: 'configs', label: '配置' },
  { key: 'manifest', label: 'Manifest' },
]

const EMPTY_DETAIL: DetailBundle = {
  detail: null,
  bindings: [],
  migrations: [],
  configs: [],
  schemas: [],
}

/**
 * 插件管理页负责展示宿主已发现的插件声明、治理状态和扩展绑定信息。
 */
export function PluginsPage() {
  const [summary, setSummary] = useState<PluginSummary>({})
  const [plugins, setPlugins] = useState<PluginListItem[]>([])
  const [selectedPluginID, setSelectedPluginID] = useState<string>('')
  const [detailBundle, setDetailBundle] = useState<DetailBundle>(EMPTY_DETAIL)
  const [activeTab, setActiveTab] = useState<DetailTab>('overview')
  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all')
  const [loading, setLoading] = useState(true)
  const [detailLoading, setDetailLoading] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [updatingID, setUpdatingID] = useState('')

  const loadPlugins = useCallback(async () => {
    setLoading(true)
    const [summaryRes, listRes] = await Promise.all([
      apiGet<{ summary: PluginSummary }>('/api/admin/plugins/summary'),
      apiGet<{ plugins: PluginListItem[] }>('/api/admin/plugins'),
    ])

    if (summaryRes.success && summaryRes.summary) {
      setSummary(summaryRes.summary)
    }
    if (listRes.success && listRes.plugins) {
      setPlugins(listRes.plugins)
      const firstID = listRes.plugins[0]?.id || ''
      setSelectedPluginID((current) => {
        const stillExists = current && listRes.plugins.some((item) => item.id === current)
        return stillExists ? current : firstID
      })
    } else if (!listRes.success) {
      toast.error(listRes.error || '插件列表加载失败')
    }
    setLoading(false)
  }, [])

  const loadPluginDetail = useCallback(async (pluginID: string) => {
    if (!pluginID) {
      setDetailBundle(EMPTY_DETAIL)
      return
    }

    setDetailLoading(true)
    const [detailRes, bindingsRes, migrationsRes, configsRes, schemasRes] = await Promise.all([
      apiGet<{ plugin: PluginDetail }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}`),
      apiGet<{ bindings: PluginBinding[] }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/bindings`),
      apiGet<{ migrations: PluginMigration[] }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/migrations`),
      apiGet<{ configs: PluginConfigRecord[] }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/configs`),
      apiGet<{ schemas: ConfigSchema[] }>('/api/admin/plugins/config-schemas'),
    ])

    if (!detailRes.success) {
      toast.error(detailRes.error || '插件详情加载失败')
      setDetailBundle(EMPTY_DETAIL)
      setDetailLoading(false)
      return
    }

    const schemas = schemasRes.success && schemasRes.schemas
      ? schemasRes.schemas.filter((item) => item.pluginId === pluginID)
      : []

    setDetailBundle({
      detail: detailRes.plugin || null,
      bindings: bindingsRes.success ? bindingsRes.bindings || [] : [],
      migrations: migrationsRes.success ? migrationsRes.migrations || [] : [],
      configs: configsRes.success ? configsRes.configs || [] : [],
      schemas,
    })
    setDetailLoading(false)
  }, [])

  useEffect(() => {
    loadPlugins()
  }, [loadPlugins])

  useEffect(() => {
    loadPluginDetail(selectedPluginID)
  }, [loadPluginDetail, selectedPluginID])

  const selectedPlugin = useMemo(() => {
    return plugins.find((plugin) => plugin.id === selectedPluginID) || null
  }, [plugins, selectedPluginID])

  const filteredPlugins = useMemo(() => {
    const normalizedKeyword = keyword.trim().toLowerCase()
    return plugins.filter((plugin) => {
      const matchesStatus = statusFilter === 'all'
        || (statusFilter === 'enabled' && plugin.enabled)
        || (statusFilter === 'disabled' && !plugin.enabled)
      const text = [
        plugin.id,
        plugin.display_name,
        plugin.description,
        plugin.author,
        plugin.plugin_kind,
      ].join(' ').toLowerCase()
      return matchesStatus && (!normalizedKeyword || text.includes(normalizedKeyword))
    })
  }, [keyword, plugins, statusFilter])

  const handleRefresh = async () => {
    setRefreshing(true)
    const res = await apiPost<{ results: DiscoveryResult[] }>('/api/admin/plugins/refresh', {})
    setRefreshing(false)
    if (res.success) {
      const failed = (res.results || []).filter((item) => item.errors && item.errors.length > 0)
      toast.success(failed.length > 0 ? `已刷新，${failed.length} 个插件存在声明错误` : '插件目录已刷新')
      notifyPluginFrontendChanged()
      await loadPlugins()
    } else {
      toast.error(res.error || '刷新失败')
    }
  }

  const handleTogglePlugin = async (plugin: PluginListItem) => {
    setUpdatingID(plugin.id)
    const url = plugin.enabled
      ? `/api/admin/plugin/${encodeURIComponent(plugin.id)}/disable`
      : `/api/admin/plugin/${encodeURIComponent(plugin.id)}/enable`
    const res = await apiPost(url, {})
    setUpdatingID('')
    if (res.success) {
      toast.success(plugin.enabled ? '插件已停用' : '插件已启用')
      notifyPluginFrontendChanged()
      await loadPlugins()
      await loadPluginDetail(plugin.id)
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  const capabilityTotal = useMemo(() => {
    return plugins.reduce((total, plugin) => total + plugin.pages + plugin.menus + plugin.routes + plugin.events + plugin.jobs, 0)
  }, [plugins])

  return (
    <div className="space-y-5">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h2 className="text-xl font-semibold text-dark-100">插件管理</h2>
          <p className="mt-1 text-sm text-dark-400">
            管理已安装插件的声明、能力、运行状态和宿主绑定关系。
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Button variant="secondary" onClick={loadPlugins} disabled={loading || refreshing}>
            <i className={cn('fas fa-rotate-right', loading && 'fa-spin')} />
            重新加载
          </Button>
          <Button onClick={handleRefresh} loading={refreshing}>
            <i className="fas fa-folder-open" />
            扫描目录
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="已发现插件" value={summary.plugins ?? plugins.length} icon="fas fa-puzzle-piece" />
        <MetricCard label="前端页面" value={summary.frontend_pages ?? 0} icon="fas fa-window-maximize" />
        <MetricCard label="前端菜单" value={summary.frontend_menus ?? 0} icon="fas fa-list" />
        <MetricCard label="能力声明" value={capabilityTotal} icon="fas fa-diagram-project" />
      </div>

      <div className="grid gap-5 xl:grid-cols-[minmax(0,0.95fr)_minmax(0,1.45fr)]">
        <Card className="min-h-[620px]">
          <div className="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h3 className="text-lg font-semibold text-dark-100">插件列表</h3>
              <p className="text-sm text-dark-400">选择插件查看详情和治理数据。</p>
            </div>
            <Badge variant="info">{filteredPlugins.length} / {plugins.length}</Badge>
          </div>

          <div className="mb-4 grid gap-3 lg:grid-cols-[1fr_auto]">
            <Input
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
              placeholder="搜索插件名称、ID、作者"
              icon={<i className="fas fa-search" />}
            />
            <select
              value={statusFilter}
              onChange={(event) => setStatusFilter(event.target.value as 'all' | 'enabled' | 'disabled')}
              className="input h-12 min-w-32"
            >
              <option value="all">全部状态</option>
              <option value="enabled">已启用</option>
              <option value="disabled">已停用</option>
            </select>
          </div>

          {loading ? (
            <LoadingState label="正在加载插件列表" />
          ) : filteredPlugins.length === 0 ? (
            <EmptyState title="暂无插件" description="当前没有匹配的插件声明。" />
          ) : (
            <div className="space-y-3">
              {filteredPlugins.map((plugin) => (
                <button
                  key={plugin.id}
                  type="button"
                  onClick={() => {
                    setSelectedPluginID(plugin.id)
                    setActiveTab('overview')
                  }}
                  className={cn(
                    'w-full rounded-2xl border p-4 text-left transition-all duration-200',
                    selectedPluginID === plugin.id
                      ? 'border-primary-500/60 bg-primary-500/10 shadow-lg shadow-primary-500/10'
                      : 'border-dark-700/70 bg-dark-800/40 hover:border-dark-600 hover:bg-dark-800/70',
                  )}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="truncate font-semibold text-dark-100">{plugin.display_name || plugin.id}</span>
                        <Badge variant={plugin.enabled ? 'success' : 'default'}>
                          {plugin.enabled ? '已启用' : '已停用'}
                        </Badge>
                      </div>
                      <div className="mt-1 truncate text-xs text-dark-500">{plugin.id}</div>
                    </div>
                    <StatusBadge value={plugin.health_status || plugin.runtime_state} type="runtime" />
                  </div>
                  <p className="mt-3 line-clamp-2 text-sm text-dark-400">
                    {plugin.description || '未提供描述'}
                  </p>
                  <div className="mt-4 grid grid-cols-4 gap-2 text-center text-xs">
                    <MiniStat label="页面" value={plugin.pages} />
                    <MiniStat label="路由" value={plugin.routes} />
                    <MiniStat label="权限" value={plugin.permissions} />
                    <MiniStat label="迁移" value={plugin.migrations} />
                  </div>
                </button>
              ))}
            </div>
          )}
        </Card>

        <Card className="min-h-[620px]">
          {!selectedPlugin ? (
            <EmptyState title="未选择插件" description="请从左侧列表选择一个插件。" />
          ) : (
            <div className="space-y-5">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="truncate text-xl font-semibold text-dark-100">
                      {selectedPlugin.display_name || selectedPlugin.id}
                    </h3>
                    <Badge variant="info">{selectedPlugin.plugin_kind || 'plugin'}</Badge>
                    <Badge variant={selectedPlugin.enabled ? 'success' : 'default'}>
                      {selectedPlugin.enabled ? '已启用' : '已停用'}
                    </Badge>
                  </div>
                  <p className="mt-1 text-sm text-dark-400">{selectedPlugin.description || '未提供描述'}</p>
                  <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-dark-500">
                    <span>ID：{selectedPlugin.id}</span>
                    <span>版本：{selectedPlugin.version || '-'}</span>
                    <span>作者：{selectedPlugin.author || '-'}</span>
                  </div>
                </div>
                <Button
                  variant={selectedPlugin.enabled ? 'warning' : 'success'}
                  loading={updatingID === selectedPlugin.id}
                  onClick={() => handleTogglePlugin(selectedPlugin)}
                >
                  <i className={selectedPlugin.enabled ? 'fas fa-power-off' : 'fas fa-play'} />
                  {selectedPlugin.enabled ? '停用' : '启用'}
                </Button>
              </div>

              <div className="flex flex-wrap gap-2 border-b border-dark-700/70 pb-3">
                {DETAIL_TABS.map((tab) => (
                  <button
                    key={tab.key}
                    type="button"
                    onClick={() => setActiveTab(tab.key)}
                    className={cn(
                      'rounded-lg px-3 py-2 text-sm transition-colors',
                      activeTab === tab.key
                        ? 'bg-primary-500/20 text-primary-300'
                        : 'text-dark-400 hover:bg-dark-700/60 hover:text-dark-200',
                    )}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>

              {detailLoading ? (
                <LoadingState label="正在加载插件详情" />
              ) : (
                <PluginDetailPanel
                  activeTab={activeTab}
                  plugin={selectedPlugin}
                  detailBundle={detailBundle}
                />
              )}
            </div>
          )}
        </Card>
      </div>
    </div>
  )
}

function PluginDetailPanel({
  activeTab,
  plugin,
  detailBundle,
}: {
  activeTab: DetailTab
  plugin: PluginListItem
  detailBundle: DetailBundle
}) {
  if (activeTab === 'bindings') {
    return <BindingsPanel bindings={detailBundle.bindings} />
  }
  if (activeTab === 'migrations') {
    return <MigrationsPanel migrations={detailBundle.migrations} />
  }
  if (activeTab === 'configs') {
    return <ConfigsPanel configs={detailBundle.configs} schemas={detailBundle.schemas} />
  }
  if (activeTab === 'manifest') {
    return <JsonPanel value={detailBundle.detail?.manifest || {}} />
  }
  return <OverviewPanel plugin={plugin} detail={detailBundle.detail} />
}

function OverviewPanel({ plugin, detail }: { plugin: PluginListItem; detail: PluginDetail | null }) {
  const manifest = detail?.manifest
  const registry = detail?.registry
  const runtime = detail?.runtime

  return (
    <div className="space-y-5">
      <div className="grid gap-3 md:grid-cols-3">
        <StateCard label="启用状态" value={plugin.enabled ? '已启用' : '已停用'} variant={plugin.enabled ? 'success' : 'default'} />
        <StateCard label="运行状态" value={registry?.health_status || runtime?.state || plugin.runtime_state || '-'} variant={getRuntimeVariant(registry?.health_status || runtime?.state || plugin.runtime_state)} />
        <StateCard label="校验状态" value={registry?.last_verify_status || plugin.verify_status || '-'} variant={getVerifyVariant(registry?.last_verify_status || plugin.verify_status)} />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <InfoSection
          title="声明信息"
          items={[
            ['插件 ID', manifest?.id || plugin.id],
            ['显示名称', manifest?.identity?.displayName || manifest?.identity?.name || plugin.display_name],
            ['版本', manifest?.version || plugin.version],
            ['类型', manifest?.pluginKind || plugin.plugin_kind],
            ['作者', manifest?.identity?.author || plugin.author || '-'],
            ['许可证', manifest?.identity?.license || '-'],
            ['主页', manifest?.identity?.homepage || '-'],
          ]}
        />
        <InfoSection
          title="治理信息"
          items={[
            ['安装 ID', registry?.install_id || runtime?.installId || '-'],
            ['安装目录', registry?.install_root || '-'],
            ['来源类型', registry?.source_type || '-'],
            ['期望状态', registry?.desired_state || plugin.desired_state || '-'],
            ['生命周期', registry?.lifecycle_state || plugin.lifecycle_state || '-'],
            ['信任级别', registry?.trust_level || runtime?.trustLevel || '-'],
            ['签名状态', registry?.signature_status || '-'],
          ]}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <InfoSection
          title="运行信息"
          items={[
            ['进程 ID', runtime?.pid ? String(runtime.pid) : '-'],
            ['控制协议', runtime?.controlTransport || manifest?.backend?.controlProtocol || '-'],
            ['数据协议', runtime?.dataTransport || manifest?.backend?.dataProtocol || '-'],
            ['协议版本', runtime?.selectedProtocolVersion || '-'],
            ['流量状态', runtime?.trafficEnabled ? '已接入' : '未接入'],
            ['最近延迟', runtime?.recentLatencyMs ? `${runtime.recentLatencyMs} ms` : '-'],
            ['错误计数', runtime?.recentErrorCount !== undefined ? String(runtime.recentErrorCount) : '-'],
          ]}
        />
        <InfoSection
          title="时间信息"
          items={[
            ['最近校验', formatDateTime(registry?.last_verified_at)],
            ['最近启动', formatDateTime(registry?.last_start_at || runtime?.startedAt)],
            ['最近就绪', formatDateTime(registry?.last_ready_at || runtime?.readyAt)],
            ['最近停止', formatDateTime(registry?.last_stop_at)],
            ['最近故障', formatDateTime(registry?.last_fault_at)],
            ['心跳时间', formatDateTime(runtime?.lastHeartbeatAt)],
            ['更新时间', formatDateTime(registry?.updated_at)],
          ]}
        />
      </div>

      <div>
        <h4 className="mb-3 font-medium text-dark-100">能力统计</h4>
        <div className="grid grid-cols-2 gap-3 md:grid-cols-4 xl:grid-cols-6">
          <CapabilityCard label="页面" value={plugin.pages} />
          <CapabilityCard label="菜单" value={plugin.menus} />
          <CapabilityCard label="路由" value={plugin.routes} />
          <CapabilityCard label="事件" value={plugin.events} />
          <CapabilityCard label="任务" value={plugin.jobs} />
          <CapabilityCard label="权限" value={plugin.permissions} />
        </div>
      </div>

      {registry?.quarantine_reason && (
        <div className="rounded-xl border border-red-500/30 bg-red-500/10 p-4">
          <div className="mb-2 font-medium text-red-300">隔离原因</div>
          <pre className="whitespace-pre-wrap break-words text-sm text-red-200">{registry.quarantine_reason}</pre>
        </div>
      )}
    </div>
  )
}

function BindingsPanel({ bindings }: { bindings: PluginBinding[] }) {
  if (bindings.length === 0) {
    return <EmptyState title="暂无绑定" description="当前插件没有已登记的宿主绑定关系。" />
  }

  return (
    <div className="overflow-hidden rounded-xl border border-dark-700/70">
      <div className="grid grid-cols-[1.1fr_1.2fr_0.8fr_0.9fr_0.8fr] gap-3 bg-dark-800/70 px-4 py-3 text-xs font-medium text-dark-400">
        <span>类型</span>
        <span>键</span>
        <span>挂载区</span>
        <span>目标</span>
        <span>状态</span>
      </div>
      {bindings.map((binding) => (
        <div key={binding.id} className="grid grid-cols-[1.1fr_1.2fr_0.8fr_0.9fr_0.8fr] gap-3 border-t border-dark-700/70 px-4 py-3 text-sm">
          <span className="text-dark-200">{binding.binding_type}</span>
          <span className="break-all text-dark-300">{binding.binding_key}</span>
          <span className="text-dark-400">{binding.mount_area || binding.target_scope || '-'}</span>
          <span className="break-all text-dark-400">{binding.route_or_view_id || '-'}</span>
          <span>
            <Badge variant={binding.enabled ? 'success' : 'default'}>{binding.enabled ? '启用' : '停用'}</Badge>
          </span>
        </div>
      ))}
    </div>
  )
}

function MigrationsPanel({ migrations }: { migrations: PluginMigration[] }) {
  if (migrations.length === 0) {
    return <EmptyState title="暂无迁移" description="当前插件没有声明数据库迁移。" />
  }

  return (
    <div className="space-y-3">
      {migrations.map((migration) => (
        <div key={migration.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <div className="font-medium text-dark-100">{migration.migration_id}</div>
              <div className="mt-1 text-xs text-dark-500">
                {migration.version || '-'} · {migration.direction || '-'} · {migration.path || '-'}
              </div>
            </div>
            <StatusBadge value={migration.status || 'declared'} type="migration" />
          </div>
          <div className="mt-3 grid gap-2 text-sm md:grid-cols-2">
            <InfoLine label="校验值" value={migration.checksum || '-'} />
            <InfoLine label="执行时间" value={formatDateTime(migration.executed_at)} />
          </div>
          {migration.error_message && (
            <pre className="mt-3 whitespace-pre-wrap break-words rounded-lg bg-red-500/10 p-3 text-xs text-red-300">
              {migration.error_message}
            </pre>
          )}
        </div>
      ))}
    </div>
  )
}

function ConfigsPanel({
  configs,
  schemas,
}: {
  configs: PluginConfigRecord[]
  schemas: ConfigSchema[]
}) {
  const hasData = configs.length > 0 || schemas.length > 0
  if (!hasData) {
    return <EmptyState title="暂无配置声明" description="当前插件没有已登记的配置记录或配置 schema。" />
  }

  return (
    <div className="space-y-5">
      {schemas.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">配置 Schema</h4>
          {schemas.map((schema, index) => (
            <div key={`${schema.pluginId || 'schema'}-${index}`} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div className="font-medium text-dark-100">版本 {schema.configVersion || '-'}</div>
                  <div className="mt-1 text-xs text-dark-500">Schema：{schema.schemaVersion || '-'}</div>
                </div>
                <Badge variant="info">{schema.sections?.length || 0} 个分组</Badge>
              </div>
              <div className="mt-4 space-y-3">
                {(schema.sections || []).map((section) => (
                  <div key={section.id || section.title} className="rounded-lg bg-dark-900/40 p-3">
                    <div className="font-medium text-dark-200">{section.title || section.id || '未命名分组'}</div>
                    {section.description && <p className="mt-1 text-sm text-dark-500">{section.description}</p>}
                    <div className="mt-3 flex flex-wrap gap-2">
                      {(section.fields || []).map((field) => (
                        <span key={field.key || field.id} className="rounded-full bg-dark-700/70 px-3 py-1 text-xs text-dark-300">
                          {field.label || field.key || field.id}
                          {field.required ? ' *' : ''}
                          {field.secret ? ' · 密钥' : ''}
                        </span>
                      ))}
                    </div>
                  </div>
                ))}
                {(schema.sections || []).length === 0 && (
                  <div className="rounded-lg bg-dark-900/40 p-3 text-sm text-dark-500">未声明配置分组。</div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {configs.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">配置记录</h4>
          {configs.map((config) => (
            <div key={config.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div className="font-medium text-dark-100">{config.config_key}</div>
                  <div className="mt-1 text-xs text-dark-500">版本：{config.config_version} · 更新人：{config.updated_by || '-'}</div>
                </div>
                <Badge variant={config.enabled ? 'success' : 'default'}>{config.enabled ? '启用' : '停用'}</Badge>
              </div>
              <div className="mt-3 grid gap-3 lg:grid-cols-2">
                <JsonBlock title="Schema JSON" value={parseJSON(config.schema_json)} />
                <JsonBlock title="Value JSON" value={parseJSON(config.value_json)} />
              </div>
              <div className="mt-3 text-xs text-dark-500">更新时间：{formatDateTime(config.updated_at)}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function MetricCard({ label, value, icon }: { label: string; value: number; icon: string }) {
  return (
    <Card className="p-5">
      <div className="flex items-center justify-between">
        <div>
          <div className="text-sm text-dark-400">{label}</div>
          <div className="mt-2 text-2xl font-semibold text-dark-100">{value}</div>
        </div>
        <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary-500/15 text-primary-300">
          <i className={icon} />
        </div>
      </div>
    </Card>
  )
}

function MiniStat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg bg-dark-900/40 px-2 py-2">
      <div className="font-semibold text-dark-100">{value}</div>
      <div className="mt-0.5 text-dark-500">{label}</div>
    </div>
  )
}

function StateCard({ label, value, variant }: { label: string; value: string; variant: BadgeVariant }) {
  return (
    <div className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
      <div className="text-sm text-dark-400">{label}</div>
      <div className="mt-3">
        <Badge variant={variant}>{formatStateValue(value)}</Badge>
      </div>
    </div>
  )
}

function CapabilityCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4 text-center">
      <div className="text-xl font-semibold text-dark-100">{value}</div>
      <div className="mt-1 text-xs text-dark-500">{label}</div>
    </div>
  )
}

function InfoSection({ title, items }: { title: string; items: Array<[string, string]> }) {
  return (
    <div className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
      <h4 className="mb-3 font-medium text-dark-100">{title}</h4>
      <div className="space-y-2">
        {items.map(([label, value]) => (
          <InfoLine key={label} label={label} value={value} />
        ))}
      </div>
    </div>
  )
}

function InfoLine({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid grid-cols-[6rem_1fr] gap-3 text-sm">
      <span className="text-dark-500">{label}</span>
      <span className="break-all text-dark-300">{value || '-'}</span>
    </div>
  )
}

function StatusBadge({ value, type }: { value: string; type: 'runtime' | 'migration' }) {
  const variant = type === 'runtime' ? getRuntimeVariant(value) : getMigrationVariant(value)
  return <Badge variant={variant}>{formatStateValue(value)}</Badge>
}

function LoadingState({ label }: { label: string }) {
  return (
    <div className="flex min-h-64 flex-col items-center justify-center gap-3 text-dark-400">
      <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      <span>{label}</span>
    </div>
  )
}

function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex min-h-64 flex-col items-center justify-center rounded-xl border border-dashed border-dark-700/70 bg-dark-800/30 p-8 text-center">
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-dark-700/70 text-dark-400">
        <i className="fas fa-puzzle-piece" />
      </div>
      <h4 className="mt-4 font-medium text-dark-100">{title}</h4>
      <p className="mt-2 text-sm text-dark-500">{description}</p>
    </div>
  )
}

function JsonPanel({ value }: { value: unknown }) {
  return <JsonBlock title="原始声明" value={value} tall />
}

function JsonBlock({ title, value, tall = false }: { title: string; value: unknown; tall?: boolean }) {
  return (
    <div className="rounded-xl border border-dark-700/70 bg-dark-950/60">
      <div className="border-b border-dark-700/70 px-4 py-3 text-sm font-medium text-dark-200">{title}</div>
      <pre className={cn('overflow-auto p-4 text-xs leading-5 text-dark-300', tall ? 'max-h-[520px]' : 'max-h-64')}>
        {safeStringify(value)}
      </pre>
    </div>
  )
}

function getRuntimeVariant(value?: string): BadgeVariant {
  switch ((value || '').toLowerCase()) {
    case 'running':
    case 'ready':
    case 'enabled':
      return 'success'
    case 'starting':
    case 'stopping':
    case 'degraded':
      return 'warning'
    case 'faulted':
    case 'failed':
    case 'quarantined':
      return 'danger'
    case 'stopped':
    case 'approved-disabled':
    case 'discovered':
      return 'default'
    default:
      return 'info'
  }
}

function getVerifyVariant(value?: string): BadgeVariant {
  switch ((value || '').toLowerCase()) {
    case 'passed':
      return 'success'
    case 'failed':
      return 'danger'
    case 'pending':
    case 'unchecked':
      return 'warning'
    default:
      return 'default'
  }
}

function getMigrationVariant(value?: string): BadgeVariant {
  switch ((value || '').toLowerCase()) {
    case 'applied':
    case 'completed':
      return 'success'
    case 'failed':
      return 'danger'
    case 'running':
    case 'pending':
      return 'warning'
    default:
      return 'default'
  }
}

function formatStateValue(value?: string) {
  const labels: Record<string, string> = {
    enabled: '已启用',
    'approved-disabled': '已停用',
    disabled: '已停用',
    discovered: '已发现',
    stopped: '未运行',
    running: '运行中',
    ready: '就绪',
    starting: '启动中',
    stopping: '停止中',
    faulted: '故障',
    failed: '失败',
    passed: '通过',
    pending: '待处理',
    declared: '已声明',
    applied: '已应用',
    completed: '已完成',
    quarantined: '已隔离',
    unchecked: '未校验',
  }
  if (!value) return '-'
  return labels[value.toLowerCase()] || value
}

function parseJSON(value: string) {
  if (!value) return {}
  try {
    return JSON.parse(value) as unknown
  } catch {
    return value
  }
}

function safeStringify(value: unknown) {
  try {
    return JSON.stringify(value ?? {}, null, 2)
  } catch {
    return String(value)
  }
}

function notifyPluginFrontendChanged() {
  window.dispatchEvent(new Event(PLUGIN_FRONTEND_CHANGED_EVENT))
}
