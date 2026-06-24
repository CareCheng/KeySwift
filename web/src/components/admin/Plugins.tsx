'use client'

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import toast from 'react-hot-toast'
import { Badge, Button, Card, ConfirmModal, Input, Modal, Switch } from '@/components/ui'
import { apiGet, apiPost, apiUpload } from '@/lib/api'
import { cn, formatDateTime } from '@/lib/utils'
import { PLUGIN_FRONTEND_CHANGED_EVENT } from '@/lib/pluginRegistry'

type BadgeVariant = 'success' | 'warning' | 'danger' | 'info' | 'default'
type DetailTab = 'overview' | 'runtime' | 'permissions' | 'bindings' | 'migrations' | 'database' | 'configs' | 'manifest'

const PLUGIN_CATEGORY_LABELS: Record<string, string> = {
  payment: '支付',
  fulfillment: '发卡交付',
  security: '安全',
  'human-verification': '人机验证',
  'customer-support': '客服支持',
  marketing: '营销',
  notification: '通知',
  analytics: '分析统计',
  tooling: '工具维护',
  'ui-theme': '主题外观',
}

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
  categories?: string[]
  tags?: string[]
  keywords?: string[]
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
  database_tables: number
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
    categories?: string[]
    tags?: string[]
    keywords?: string[]
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

interface PluginConfigValue {
  id: number
  plugin_id: string
  config_key: string
  value_json: string
  secret_json: string
  revision: number
  updated_by: string
  updated_at: string
}

interface PluginConfigRevision {
  id: number
  plugin_id: string
  config_key: string
  revision: number
  value_digest: string
  secret_json: string
  updated_by: string
  change_summary: string
  created_at: string
}

interface ConfigField {
  id?: string
  key?: string
  label?: string
  type?: string
  required?: boolean
  secret?: boolean
  description?: string
  default?: unknown
  enumOptions?: string[]
  options?: Array<string | { label?: string; value?: string | number | boolean }>
}

interface ConfigSection {
  id?: string
  title?: string
  description?: string
  scope?: string
  fields?: ConfigField[]
}

interface PermissionDefinition {
  id: number
  permission_code: string
  owner_type: string
  owner_plugin_id: string
  risk_level: string
  group_key: string
  name: string
  description: string
  default_grant_policy: string
  status: string
}

interface PluginRuntimeSession {
  id: number
  plugin_id: string
  version: string
  instance_id: string
  pid: number
  state: string
  started_at: string
  ready_at?: string
  stopped_at?: string
  last_heartbeat_at?: string
  fault_reason: string
}

interface PluginStateEvent {
  id: number
  plugin_id: string
  from_state: string
  to_state: string
  event_type: string
  reason: string
  operator_subject_id: string
  created_at: string
}

interface PluginFaultLog {
  id: number
  plugin_id: string
  instance_id: string
  fault_type: string
  fault_reason: string
  stack_trace: string
  created_at: string
}

interface PluginTrustRecord {
  id: number
  plugin_id: string
  version: string
  trust_level: string
  signature_status: string
  approved_by: string
  approved_at?: string
  risk_summary: string
}

interface PluginDatabaseDeclaration {
  id?: number
  plugin_id?: string
  plugin_version?: string
  namespace?: string
  storage_mode?: string
  table_count?: number
  status?: string
  extensions_json?: string
  updated_at?: string
}

interface PluginDatabaseTable {
  id: number
  table_key: string
  physical_table_name: string
  table_kind: string
  schema_version: string
  schema_checksum: string
  status: string
  sensitivity: string
  create_policy: string
  drop_policy: string
  backup_policy: string
  retention_policy: string
  description: string
  extensions_json: string
}

interface PluginDatabaseColumn {
  id: number
  table_id: number
  column_key: string
  column_name: string
  db_type: string
  logical_type: string
  nullable: boolean
  default_value_json: string
  primary_key: boolean
  auto_increment: boolean
  unique_key: boolean
  indexed: boolean
  encrypted: boolean
  secret: boolean
  reference_type: string
  reference_target: string
  description: string
  extensions_json: string
}

interface PluginDatabaseIndex {
  id: number
  table_id: number
  index_key: string
  index_name: string
  columns_json: string
  unique_index: boolean
  status: string
  extensions_json: string
}

interface PluginDatabaseRelation {
  id: number
  table_id: number
  relation_key: string
  local_column: string
  target_resource_type: string
  target_key: string
  relation_type: string
  required: boolean
  on_delete_policy: string
  extensions_json: string
}

interface PluginDatabaseOperation {
  id: number
  operation_id: string
  plugin_version: string
  table_key: string
  operation_type: string
  path: string
  requires_review: boolean
  status: string
  schema_checksum: string
  executed_by: string
  error_message: string
  extensions_json: string
  started_at?: string
  finished_at?: string
  created_at?: string
}

interface PluginDatabaseSnapshot {
  declaration?: PluginDatabaseDeclaration | null
  tables?: PluginDatabaseTable[]
  columns?: PluginDatabaseColumn[]
  indexes?: PluginDatabaseIndex[]
  relations?: PluginDatabaseRelation[]
  operations?: PluginDatabaseOperation[]
}

interface ConfigSchema {
  schemaVersion?: string
  schema_version?: string
  pluginId?: string
  plugin_id?: string
  configVersion?: string
  config_version?: string
  schema_json?: string
  sections?: ConfigSection[]
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
  configValues: PluginConfigValue[]
  configRevisions: PluginConfigRevision[]
  schemas: ConfigSchema[]
  permissions: PermissionDefinition[]
  runtimeSessions: PluginRuntimeSession[]
  runtimeEvents: PluginStateEvent[]
  faultLogs: PluginFaultLog[]
  trustRecords: PluginTrustRecord[]
  database: PluginDatabaseSnapshot
}

type UninstallToggleTarget = 'config' | 'database'

const DETAIL_TABS: Array<{ key: DetailTab; label: string }> = [
  { key: 'overview', label: '概览' },
  { key: 'runtime', label: '运行时' },
  { key: 'permissions', label: '权限' },
  { key: 'bindings', label: '绑定' },
  { key: 'migrations', label: '迁移' },
  { key: 'database', label: '数据库' },
  { key: 'configs', label: '配置' },
  { key: 'manifest', label: 'Manifest' },
]

const EMPTY_DETAIL: DetailBundle = {
  detail: null,
  bindings: [],
  migrations: [],
  configValues: [],
  configRevisions: [],
  schemas: [],
  permissions: [],
  runtimeSessions: [],
  runtimeEvents: [],
  faultLogs: [],
  trustRecords: [],
  database: {
    declaration: null,
    tables: [],
    columns: [],
    indexes: [],
    relations: [],
    operations: [],
  },
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
  const [categoryFilter, setCategoryFilter] = useState('all')
  const [loading, setLoading] = useState(true)
  const [detailLoading, setDetailLoading] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [installing, setInstalling] = useState(false)
  const [updatingID, setUpdatingID] = useState('')
  const [uninstallTargetID, setUninstallTargetID] = useState('')
  const [showUninstallModal, setShowUninstallModal] = useState(false)
  const [deleteConfig, setDeleteConfig] = useState(false)
  const [deleteDatabase, setDeleteDatabase] = useState(false)
  const [pendingToggle, setPendingToggle] = useState<UninstallToggleTarget | null>(null)
  const [showFinalUninstallConfirm, setShowFinalUninstallConfirm] = useState(false)
  const [uninstallLoading, setUninstallLoading] = useState(false)
  const packageInputRef = useRef<HTMLInputElement>(null)

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
    const [detailRes, bindingsRes, migrationsRes, databaseRes, schemasRes, runtimeRes, permissionRes, configValuesRes] = await Promise.all([
      apiGet<{ plugin: PluginDetail }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}`),
      apiGet<{ bindings: PluginBinding[] }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/bindings`),
      apiGet<{ migrations: PluginMigration[] }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/migrations`),
      apiGet<{ database: PluginDatabaseSnapshot }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/database`),
      apiGet<{ schemas: ConfigSchema[] }>('/api/admin/plugins/config-schemas'),
      apiGet<{ sessions: PluginRuntimeSession[]; events: PluginStateEvent[]; faults: PluginFaultLog[]; trust_records: PluginTrustRecord[] }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/runtime-records`),
      apiGet<{ permissions: PermissionDefinition[] }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/permission-definitions`),
      apiGet<{ values: PluginConfigValue[]; revisions: PluginConfigRevision[] }>(`/api/admin/plugin/${encodeURIComponent(pluginID)}/config-values`),
    ])

    if (!detailRes.success) {
      toast.error(detailRes.error || '插件详情加载失败')
      setDetailBundle(EMPTY_DETAIL)
      setDetailLoading(false)
      return
    }

    const schemas = schemasRes.success && schemasRes.schemas
      ? schemasRes.schemas.map(normalizeConfigSchema).filter((item) => item.pluginId === pluginID)
      : []

    setDetailBundle({
      detail: detailRes.plugin || null,
      bindings: bindingsRes.success ? bindingsRes.bindings || [] : [],
      migrations: migrationsRes.success ? migrationsRes.migrations || [] : [],
      configValues: configValuesRes.success ? configValuesRes.values || [] : [],
      configRevisions: configValuesRes.success ? configValuesRes.revisions || [] : [],
      schemas,
      permissions: permissionRes.success ? permissionRes.permissions || [] : [],
      runtimeSessions: runtimeRes.success ? runtimeRes.sessions || [] : [],
      runtimeEvents: runtimeRes.success ? runtimeRes.events || [] : [],
      faultLogs: runtimeRes.success ? runtimeRes.faults || [] : [],
      trustRecords: runtimeRes.success ? runtimeRes.trust_records || [] : [],
      database: databaseRes.success && databaseRes.database ? normalizeDatabaseSnapshot(databaseRes.database) : EMPTY_DETAIL.database,
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

  const hasPluginDatabase = Boolean(detailBundle.database?.declaration) || (detailBundle.database?.tables?.length || 0) > 0

  const categoryOptions = useMemo(() => {
    const categories = new Set<string>()
    plugins.forEach((plugin) => {
      normalizeList(plugin.categories).forEach((category) => categories.add(category))
    })
    return Array.from(categories).sort((left, right) => formatPluginCategory(left).localeCompare(formatPluginCategory(right), 'zh-CN'))
  }, [plugins])

  useEffect(() => {
    if (categoryFilter !== 'all' && !categoryOptions.includes(categoryFilter)) {
      setCategoryFilter('all')
    }
  }, [categoryFilter, categoryOptions])

  const filteredPlugins = useMemo(() => {
    const normalizedKeyword = keyword.trim().toLowerCase()
    return plugins.filter((plugin) => {
      const matchesStatus = statusFilter === 'all'
        || (statusFilter === 'enabled' && plugin.enabled)
        || (statusFilter === 'disabled' && !plugin.enabled)
      const categories = normalizeList(plugin.categories)
      const matchesCategory = categoryFilter === 'all' || categories.includes(categoryFilter)
      const text = [
        plugin.id,
        plugin.display_name,
        plugin.description,
        plugin.author,
        plugin.plugin_kind,
        ...categories,
        ...normalizeList(plugin.tags),
        ...normalizeList(plugin.keywords),
      ].join(' ').toLowerCase()
      return matchesStatus && matchesCategory && (!normalizedKeyword || text.includes(normalizedKeyword))
    })
  }, [categoryFilter, keyword, plugins, statusFilter])

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

  const handleInstallPackage = async (file?: File) => {
    if (!file) return
    if (!file.name.toLowerCase().endsWith('.ksplugin.zip')) {
      toast.error('请选择 .ksplugin.zip 插件安装包')
      return
    }
    const formData = new FormData()
    formData.append('package', file)
    setInstalling(true)
    const res = await apiUpload<{ results: DiscoveryResult[] }>('/api/admin/plugins/install', formData)
    setInstalling(false)
    if (res.success) {
      const failed = (res.results || []).filter((item) => item.errors && item.errors.length > 0)
      toast.success(failed.length > 0 ? `插件已安装，${failed.length} 个声明需要检查` : '插件已安装')
      notifyPluginFrontendChanged()
      await loadPlugins()
    } else {
      toast.error(res.error || '插件安装失败')
    }
    if (packageInputRef.current) {
      packageInputRef.current.value = ''
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

  const openUninstallModal = () => {
    if (!selectedPluginID) return
    setUninstallTargetID(selectedPluginID)
    setDeleteConfig(false)
    setDeleteDatabase(false)
    setPendingToggle(null)
    setShowFinalUninstallConfirm(false)
    setShowUninstallModal(true)
  }

  const closeUninstallModal = () => {
    if (uninstallLoading) {
      return
    }
    setShowUninstallModal(false)
    setUninstallTargetID('')
    setDeleteConfig(false)
    setDeleteDatabase(false)
    setPendingToggle(null)
    setShowFinalUninstallConfirm(false)
  }

  const handleToggleRequest = (target: UninstallToggleTarget, nextValue: boolean) => {
    if (!nextValue) {
      if (target === 'config') {
        setDeleteConfig(false)
      } else {
        setDeleteDatabase(false)
      }
      return
    }
    setPendingToggle(target)
  }

  const confirmToggleEnable = () => {
    if (!pendingToggle) return
    if (pendingToggle === 'config') {
      setDeleteConfig(true)
    } else {
      setDeleteDatabase(true)
    }
    setPendingToggle(null)
  }

  const requestFinalUninstall = () => {
    setShowFinalUninstallConfirm(true)
  }

  const executeUninstall = async () => {
    if (!uninstallTargetID) return
    setUninstallLoading(true)
    const res = await apiPost(`/api/admin/plugin/${encodeURIComponent(uninstallTargetID)}/uninstall`, {
      delete_config: deleteConfig,
      delete_database: deleteDatabase,
    })
    setUninstallLoading(false)
    setShowFinalUninstallConfirm(false)
    if (res.success) {
      toast.success(typeof res.message === 'string' ? res.message : '插件已卸载')
      notifyPluginFrontendChanged()
      const remainingPlugins = plugins.filter((item) => item.id !== uninstallTargetID)
      setShowUninstallModal(false)
      setUninstallTargetID('')
      setDeleteConfig(false)
      setDeleteDatabase(false)
      setPendingToggle(null)
      await loadPlugins()
      setSelectedPluginID((current) => {
        if (remainingPlugins.some((item) => item.id === current)) {
          return current
        }
        return remainingPlugins[0]?.id || ''
      })
    } else {
      toast.error(res.error || '卸载失败')
    }
  }

  const capabilityTotal = useMemo(() => {
    return plugins.reduce((total, plugin) => {
      return total + plugin.pages + plugin.menus + plugin.routes + plugin.events + plugin.jobs + (plugin.database_tables || 0)
    }, 0)
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
          <input
            ref={packageInputRef}
            type="file"
            accept=".ksplugin.zip,application/zip"
            className="hidden"
            onChange={(event) => void handleInstallPackage(event.target.files?.[0])}
          />
          <Button
            variant="success"
            loading={installing}
            disabled={loading || refreshing}
            onClick={() => packageInputRef.current?.click()}
          >
            <i className="fas fa-upload" />
            安装插件
          </Button>
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

      <Modal
        isOpen={showUninstallModal}
        onClose={closeUninstallModal}
        title="卸载插件"
        size="lg"
      >
        {selectedPlugin ? (
          <div className="space-y-5">
            <div className="rounded-xl border border-dark-700/70 bg-dark-900/40 p-4 text-sm text-dark-300">
              <div className="font-medium text-dark-100">{selectedPlugin.display_name || selectedPlugin.id}</div>
              <div className="mt-2 grid gap-2 md:grid-cols-2">
                <div>ID：{selectedPlugin.id}</div>
                <div>版本：{selectedPlugin.version || '-'}</div>
                <div>状态：{selectedPlugin.enabled ? '已启用' : '已停用'}</div>
                <div>数据库表：{detailBundle.database?.tables?.length || 0}</div>
              </div>
              <p className="mt-3 text-xs text-dark-500">
                仅删除插件本体时，不会删除配置数据和独立数据表；默认两个开关都关闭。
              </p>
            </div>

            <div className="space-y-4">
              <Switch
                checked={deleteConfig}
                onChange={(next) => handleToggleRequest('config', next)}
                label="删除配置数据"
                description="删除宿主数据库中的插件配置值与配置 schema。默认关闭，打开时需要确认。"
              />
              {hasPluginDatabase ? (
                <Switch
                  checked={deleteDatabase}
                  onChange={(next) => handleToggleRequest('database', next)}
                  label="删除独立数据表"
                  description="删除该插件声明并登记的独立业务表及其数据。默认关闭，打开时需要确认。"
                />
              ) : null}
            </div>

            <div className="flex justify-end gap-3 pt-2">
              <Button variant="secondary" onClick={closeUninstallModal} disabled={uninstallLoading}>
                取消
              </Button>
              <Button variant="danger" onClick={requestFinalUninstall}>
                确认卸载
              </Button>
            </div>
          </div>
        ) : null}
      </Modal>

      <ConfirmModal
        isOpen={pendingToggle !== null}
        onClose={() => setPendingToggle(null)}
        title={pendingToggle === 'config' ? '确认删除配置数据' : '确认删除独立数据表'}
        message={pendingToggle === 'config'
          ? '打开后，本次卸载会一并删除该插件的配置数据，删除后不可恢复。'
          : '打开后，本次卸载会一并删除该插件的独立数据表和所有数据，删除后不可恢复。'}
        confirmText="确认打开"
        variant="warning"
        onConfirm={confirmToggleEnable}
      />

      <ConfirmModal
        isOpen={showFinalUninstallConfirm}
        onClose={() => setShowFinalUninstallConfirm(false)}
        title="最终确认卸载"
        message={
          deleteConfig && deleteDatabase
            ? '本次卸载将删除插件、配置数据和独立数据表。'
            : deleteConfig
              ? '本次卸载将删除插件并清理配置数据。'
              : deleteDatabase
                ? '本次卸载将删除插件并清理独立数据表。'
                : '本次卸载仅删除插件。'
        }
        confirmText="确认卸载"
        variant="danger"
        onConfirm={executeUninstall}
        loading={uninstallLoading}
      />

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

          <div className={cn(
            'mb-4 grid gap-3',
            categoryOptions.length > 0
              ? 'lg:grid-cols-[minmax(0,1fr)_auto_auto]'
              : 'lg:grid-cols-[minmax(0,1fr)_auto]',
          )}>
            <Input
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
              placeholder="搜索插件名称、ID、作者、分类"
              icon={<i className="fas fa-search" />}
            />
            {categoryOptions.length > 0 && (
              <select
                value={categoryFilter}
                onChange={(event) => setCategoryFilter(event.target.value)}
                className="input h-12 min-w-36"
              >
                <option value="all">全部分类</option>
                {categoryOptions.map((category) => (
                  <option key={category} value={category}>{formatPluginCategory(category)}</option>
                ))}
              </select>
            )}
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
                  <PluginCategoryBadges categories={plugin.categories} className="mt-3" limit={3} />
                  <div className="mt-4 grid grid-cols-5 gap-2 text-center text-xs">
                    <MiniStat label="页面" value={plugin.pages} />
                    <MiniStat label="路由" value={plugin.routes} />
                    <MiniStat label="权限" value={plugin.permissions} />
                    <MiniStat label="迁移" value={plugin.migrations} />
                    <MiniStat label="数据表" value={plugin.database_tables || 0} />
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
                  <PluginCategoryBadges categories={selectedPlugin.categories} className="mt-2" limit={4} />
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
                <Button
                  variant="danger"
                  disabled={!selectedPlugin}
                  onClick={openUninstallModal}
                >
                  <i className="fas fa-trash" />
                  卸载
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
                onRefresh={() => loadPluginDetail(selectedPluginID)}
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
  onRefresh,
}: {
  activeTab: DetailTab
  plugin: PluginListItem
  detailBundle: DetailBundle
  onRefresh: () => void
}) {
  if (activeTab === 'runtime') {
    return (
      <RuntimePanel
        sessions={detailBundle.runtimeSessions}
        events={detailBundle.runtimeEvents}
        faults={detailBundle.faultLogs}
        trustRecords={detailBundle.trustRecords}
      />
    )
  }
  if (activeTab === 'permissions') {
    return <PermissionsPanel permissions={detailBundle.permissions} />
  }
  if (activeTab === 'bindings') {
    return <BindingsPanel bindings={detailBundle.bindings} />
  }
  if (activeTab === 'migrations') {
    return <MigrationsPanel migrations={detailBundle.migrations} />
  }
  if (activeTab === 'database') {
    return <DatabasePanel database={detailBundle.database} />
  }
  if (activeTab === 'configs') {
    return (
      <ConfigsPanel
        schemas={detailBundle.schemas}
        values={detailBundle.configValues}
        revisions={detailBundle.configRevisions}
        onSaved={onRefresh}
      />
    )
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
  const categories = normalizeList(manifest?.identity?.categories).length > 0
    ? normalizeList(manifest?.identity?.categories)
    : normalizeList(plugin.categories)
  const tags = normalizeList(manifest?.identity?.tags).length > 0
    ? normalizeList(manifest?.identity?.tags)
    : normalizeList(plugin.tags)
  const keywords = normalizeList(manifest?.identity?.keywords).length > 0
    ? normalizeList(manifest?.identity?.keywords)
    : normalizeList(plugin.keywords)

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
            ['分类', formatPluginCategoryList(categories)],
            ['标签', formatStringList(tags)],
            ['关键词', formatStringList(keywords)],
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

function DatabasePanel({ database }: { database: PluginDatabaseSnapshot }) {
  const snapshot = normalizeDatabaseSnapshot(database)
  const hasData = Boolean(snapshot.declaration)
    || (snapshot.tables || []).length > 0
    || (snapshot.columns || []).length > 0
    || (snapshot.indexes || []).length > 0
    || (snapshot.relations || []).length > 0
    || (snapshot.operations || []).length > 0

  if (!hasData) {
    return <EmptyState title="暂无数据库声明" description="当前插件没有登记 database 顶层声明或插件表结构。" />
  }

  return (
    <div className="space-y-5">
      {snapshot.declaration && (
        <div className="grid gap-4 lg:grid-cols-[1fr_auto]">
          <InfoSection
            title="数据库声明"
            items={[
              ['命名空间', snapshot.declaration.namespace || '-'],
              ['存储模式', snapshot.declaration.storage_mode || '-'],
              ['插件版本', snapshot.declaration.plugin_version || '-'],
              ['表数量', String(snapshot.declaration.table_count || 0)],
              ['状态', formatStateValue(snapshot.declaration.status || '-')],
              ['更新时间', formatDateTime(snapshot.declaration.updated_at)],
            ]}
          />
          <JsonBlock title="Database Extensions" value={parseJSON(snapshot.declaration.extensions_json || '')} />
        </div>
      )}

      <div className="grid grid-cols-2 gap-3 md:grid-cols-5">
        <CapabilityCard label="表" value={snapshot.tables?.length || 0} />
        <CapabilityCard label="字段" value={snapshot.columns?.length || 0} />
        <CapabilityCard label="索引" value={snapshot.indexes?.length || 0} />
        <CapabilityCard label="关系" value={snapshot.relations?.length || 0} />
        <CapabilityCard label="操作" value={snapshot.operations?.length || 0} />
      </div>

      {(snapshot.tables || []).length > 0 ? (
        <div className="space-y-4">
          <h4 className="font-medium text-dark-100">插件表结构</h4>
          {(snapshot.tables || []).map((table) => {
            const columns = (snapshot.columns || []).filter((column) => column.table_id === table.id)
            const indexes = (snapshot.indexes || []).filter((index) => index.table_id === table.id)
            const relations = (snapshot.relations || []).filter((relation) => relation.table_id === table.id)

            return (
              <div key={table.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <div className="font-medium text-dark-100">{table.table_key}</div>
                    <div className="mt-1 break-all text-xs text-dark-500">{table.physical_table_name}</div>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Badge variant="info">{table.table_kind || 'table'}</Badge>
                    <Badge variant={getMigrationVariant(table.status)}>{formatStateValue(table.status)}</Badge>
                    <Badge variant={table.sensitivity === 'secret' || table.sensitivity === 'sensitive' ? 'warning' : 'default'}>
                      {table.sensitivity || 'internal'}
                    </Badge>
                  </div>
                </div>

                {table.description && <p className="mt-3 text-sm text-dark-400">{table.description}</p>}

                <div className="mt-4 grid gap-3 lg:grid-cols-2">
                  <InfoSection
                    title="表策略"
                    items={[
                      ['Schema', table.schema_version || '-'],
                      ['校验值', table.schema_checksum || '-'],
                      ['创建策略', table.create_policy || '-'],
                      ['删除策略', table.drop_policy || '-'],
                      ['备份策略', table.backup_policy || '-'],
                      ['保留策略', table.retention_policy || '-'],
                    ]}
                  />
                  <JsonBlock title="Table Extensions" value={parseJSON(table.extensions_json)} />
                </div>

                <DatabaseColumns columns={columns} />
                <DatabaseIndexes indexes={indexes} />
                <DatabaseRelations relations={relations} />
              </div>
            )
          })}
        </div>
      ) : (
        <EmptyState title="暂无插件表" description="当前插件只登记了 database 顶层声明，尚未声明具体数据表。" />
      )}

      {(snapshot.operations || []).length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">结构操作</h4>
          {(snapshot.operations || []).map((operation) => (
            <div key={operation.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="font-medium text-dark-100">{operation.operation_id}</div>
                  <div className="mt-1 text-xs text-dark-500">
                    {operation.table_key || '-'} · {operation.operation_type || '-'} · {operation.path || '-'}
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Badge variant={getMigrationVariant(operation.status)}>{formatStateValue(operation.status)}</Badge>
                  <Badge variant={operation.requires_review ? 'warning' : 'default'}>
                    {operation.requires_review ? '需复核' : '自动'}
                  </Badge>
                </div>
              </div>
              <div className="mt-3 grid gap-2 text-sm md:grid-cols-2">
                <InfoLine label="插件版本" value={operation.plugin_version || '-'} />
                <InfoLine label="Schema 校验" value={operation.schema_checksum || '-'} />
                <InfoLine label="执行人" value={operation.executed_by || '-'} />
                <InfoLine label="创建时间" value={formatDateTime(operation.created_at)} />
              </div>
              {operation.error_message && (
                <pre className="mt-3 whitespace-pre-wrap break-words rounded-lg bg-red-500/10 p-3 text-xs text-red-300">
                  {operation.error_message}
                </pre>
              )}
            </div>
          ))}
        </div>
      )}

      {(snapshot.operations || []).length === 0 && (
        <EmptyState title="暂无结构操作" description="当前插件没有声明需要宿主显式执行的数据库结构操作。" />
      )}
    </div>
  )
}

function DatabaseColumns({ columns }: { columns: PluginDatabaseColumn[] }) {
  if (columns.length === 0) {
    return <div className="mt-4 rounded-lg bg-dark-900/40 p-3 text-sm text-dark-500">未声明字段。</div>
  }

  return (
    <div className="mt-4 space-y-2">
      <div className="text-sm font-medium text-dark-100">字段</div>
      {columns.map((column) => (
        <div key={column.id} className="rounded-lg bg-dark-900/40 p-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div className="break-all text-sm font-medium text-dark-200">
              {column.column_name}
              <span className="ml-2 text-xs font-normal text-dark-500">{column.db_type || '-'}</span>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {column.primary_key && <Badge variant="info">主键</Badge>}
              {column.unique_key && <Badge variant="warning">唯一</Badge>}
              {column.indexed && <Badge variant="default">索引</Badge>}
              {column.secret && <Badge variant="danger">敏感</Badge>}
              {column.encrypted && <Badge variant="warning">加密</Badge>}
              <Badge variant={column.nullable ? 'default' : 'success'}>{column.nullable ? '可空' : '必填'}</Badge>
            </div>
          </div>
          <div className="mt-2 grid gap-2 text-xs md:grid-cols-2">
            <InfoLine label="字段键" value={column.column_key || '-'} />
            <InfoLine label="逻辑类型" value={column.logical_type || '-'} />
            <InfoLine label="引用类型" value={column.reference_type || '-'} />
            <InfoLine label="引用目标" value={column.reference_target || '-'} />
          </div>
          {column.description && <p className="mt-2 text-xs text-dark-500">{column.description}</p>}
        </div>
      ))}
    </div>
  )
}

function DatabaseIndexes({ indexes }: { indexes: PluginDatabaseIndex[] }) {
  if (indexes.length === 0) {
    return <div className="mt-4 rounded-lg bg-dark-900/40 p-3 text-sm text-dark-500">未声明索引。</div>
  }

  return (
    <div className="mt-4 space-y-2">
      <div className="text-sm font-medium text-dark-100">索引</div>
      {indexes.map((index) => (
        <div key={index.id} className="rounded-lg bg-dark-900/40 p-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div className="break-all text-sm font-medium text-dark-200">{index.index_name || index.index_key}</div>
            <div className="flex flex-wrap gap-2">
              <Badge variant={index.unique_index ? 'warning' : 'default'}>{index.unique_index ? '唯一索引' : '普通索引'}</Badge>
              <Badge variant={getMigrationVariant(index.status)}>{formatStateValue(index.status)}</Badge>
            </div>
          </div>
          <div className="mt-2 text-xs text-dark-500">字段：{formatStringArray(parseJSON(index.columns_json))}</div>
        </div>
      ))}
    </div>
  )
}

function DatabaseRelations({ relations }: { relations: PluginDatabaseRelation[] }) {
  if (relations.length === 0) {
    return <div className="mt-4 rounded-lg bg-dark-900/40 p-3 text-sm text-dark-500">未声明关系。</div>
  }

  return (
    <div className="mt-4 space-y-2">
      <div className="text-sm font-medium text-dark-100">关系</div>
      {relations.map((relation) => (
        <div key={relation.id} className="rounded-lg bg-dark-900/40 p-3">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <div className="break-all text-sm font-medium text-dark-200">{relation.relation_key}</div>
            <Badge variant={relation.required ? 'warning' : 'default'}>{relation.required ? '必需' : '可选'}</Badge>
          </div>
          <div className="mt-2 grid gap-2 text-xs md:grid-cols-2">
            <InfoLine label="本地字段" value={relation.local_column || '-'} />
            <InfoLine label="目标资源" value={relation.target_resource_type || '-'} />
            <InfoLine label="目标键" value={relation.target_key || '-'} />
            <InfoLine label="删除策略" value={relation.on_delete_policy || '-'} />
          </div>
        </div>
      ))}
    </div>
  )
}

function RuntimePanel({
  sessions,
  events,
  faults,
  trustRecords,
}: {
  sessions: PluginRuntimeSession[]
  events: PluginStateEvent[]
  faults: PluginFaultLog[]
  trustRecords: PluginTrustRecord[]
}) {
  const hasData = sessions.length > 0 || events.length > 0 || faults.length > 0 || trustRecords.length > 0
  if (!hasData) {
    return <EmptyState title="暂无运行记录" description="当前插件还没有运行会话、状态事件或故障记录。" />
  }

  return (
    <div className="space-y-5">
      {sessions.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">运行会话</h4>
          {sessions.map((session) => (
            <div key={session.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="font-mono text-sm text-dark-200">{session.instance_id}</div>
                <StatusBadge value={session.state} type="runtime" />
              </div>
              <div className="mt-3 grid gap-2 text-sm md:grid-cols-2">
                <InfoLine label="进程 ID" value={session.pid ? String(session.pid) : '-'} />
                <InfoLine label="版本" value={session.version || '-'} />
                <InfoLine label="启动时间" value={formatDateTime(session.started_at)} />
                <InfoLine label="就绪时间" value={formatDateTime(session.ready_at)} />
                <InfoLine label="心跳时间" value={formatDateTime(session.last_heartbeat_at)} />
                <InfoLine label="停止原因" value={session.fault_reason || '-'} />
              </div>
            </div>
          ))}
        </div>
      )}

      {trustRecords.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">信任记录</h4>
          {trustRecords.map((record) => (
            <div key={record.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="font-medium text-dark-100">版本 {record.version}</div>
                <Badge variant={record.trust_level === 'local-approved' || record.trust_level === 'approved' ? 'success' : 'warning'}>
                  {formatStateValue(record.trust_level)}
                </Badge>
              </div>
              <div className="mt-3 grid gap-2 text-sm md:grid-cols-2">
                <InfoLine label="签名状态" value={record.signature_status || '-'} />
                <InfoLine label="批准人" value={record.approved_by || '-'} />
                <InfoLine label="批准时间" value={formatDateTime(record.approved_at)} />
                <InfoLine label="风险摘要" value={record.risk_summary || '-'} />
              </div>
            </div>
          ))}
        </div>
      )}

      {events.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">状态事件</h4>
          {events.map((event) => (
            <div key={event.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="text-sm font-medium text-dark-100">{event.event_type}</div>
                <span className="text-xs text-dark-500">{formatDateTime(event.created_at)}</span>
              </div>
              <div className="mt-2 text-sm text-dark-400">
                {formatStateValue(event.from_state || '-')} → {formatStateValue(event.to_state || '-')}
              </div>
              {event.reason && <div className="mt-2 text-sm text-dark-500">{event.reason}</div>}
            </div>
          ))}
        </div>
      )}

      {faults.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">故障日志</h4>
          {faults.map((fault) => (
            <div key={fault.id} className="rounded-xl border border-red-500/30 bg-red-500/10 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="font-medium text-red-200">{fault.fault_type || 'fault'}</div>
                <span className="text-xs text-red-200/70">{formatDateTime(fault.created_at)}</span>
              </div>
              <div className="mt-2 text-sm text-red-100">{fault.fault_reason || '-'}</div>
              {fault.stack_trace && <JsonBlock title="Stack" value={fault.stack_trace} />}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function PermissionsPanel({ permissions }: { permissions: PermissionDefinition[] }) {
  if (permissions.length === 0) {
    return <EmptyState title="暂无权限定义" description="当前插件没有登记权限定义。" />
  }

  return (
    <div className="space-y-3">
      {permissions.map((permission) => (
        <div key={permission.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <div className="font-medium text-dark-100">{permission.name || permission.permission_code}</div>
              <div className="mt-1 font-mono text-xs text-dark-500">{permission.permission_code}</div>
            </div>
            <Badge variant={permission.risk_level === 'high' ? 'warning' : 'info'}>{permission.risk_level || 'normal'}</Badge>
          </div>
          <p className="mt-3 text-sm text-dark-400">{permission.description || '未提供说明'}</p>
          <div className="mt-3 grid gap-2 text-sm md:grid-cols-2">
            <InfoLine label="分组" value={permission.group_key || '-'} />
            <InfoLine label="授权策略" value={permission.default_grant_policy || '-'} />
            <InfoLine label="状态" value={permission.status || '-'} />
            <InfoLine label="插件 ID" value={permission.owner_plugin_id || '-'} />
          </div>
        </div>
      ))}
    </div>
  )
}

function ConfigsPanel({
  schemas,
  values,
  revisions,
  onSaved,
}: {
  schemas: ConfigSchema[]
  values: PluginConfigValue[]
  revisions: PluginConfigRevision[]
  onSaved: () => void
}) {
  const hasData = schemas.length > 0 || values.length > 0 || revisions.length > 0
  const [drafts, setDrafts] = useState<Record<string, Record<string, unknown>>>({})
  const [secretDrafts, setSecretDrafts] = useState<Record<string, Record<string, string>>>({})
  const [savingKey, setSavingKey] = useState('')

  useEffect(() => {
    const nextDrafts: Record<string, Record<string, unknown>> = {}
    const nextSecretDrafts: Record<string, Record<string, string>> = {}
    schemas.forEach((schema) => {
      const configKey = schemaConfigKey(schema)
      const savedValue = savedConfigValue(values, configKey)
      const value = objectFromJSON(savedValue?.value_json)
      nextDrafts[configKey] = configDefaults(schema, value)
      nextSecretDrafts[configKey] = {}
    })
    setDrafts(nextDrafts)
    setSecretDrafts(nextSecretDrafts)
  }, [schemas, values])

  if (!hasData) {
    return <EmptyState title="暂无配置声明" description="当前插件没有已登记的配置 schema。" />
  }

  const updateDraft = (configKey: string, fieldKey: string, value: unknown) => {
    setDrafts((current) => ({
      ...current,
      [configKey]: {
        ...(current[configKey] || {}),
        [fieldKey]: value,
      },
    }))
  }

  const updateSecretDraft = (configKey: string, fieldKey: string, value: string) => {
    setSecretDrafts((current) => ({
      ...current,
      [configKey]: {
        ...(current[configKey] || {}),
        [fieldKey]: value,
      },
    }))
  }

  const saveSchemaConfig = async (schema: ConfigSchema) => {
    const pluginID = schema.pluginId || ''
    if (!pluginID) {
      toast.error('配置 schema 缺少插件 ID')
      return
    }
    const configKey = schemaConfigKey(schema)
    const savedValue = savedConfigValue(values, configKey)
    const draft = drafts[configKey] || {}
    const secretDraft = secretDrafts[configKey] || {}
    const missingField = firstMissingRequiredField(schema, draft, secretDraft, Boolean(savedValue?.secret_json))
    if (missingField) {
      toast.error(`请填写${missingField}`)
      return
    }

    const secretValues = collectSecretValues(schema, secretDraft)
    setSavingKey(configKey)
    const res = await apiPost(`/api/admin/plugin/${encodeURIComponent(pluginID)}/config-values`, {
      config_key: configKey,
      value: draft,
      secret_json: Object.keys(secretValues).length > 0 ? JSON.stringify(secretValues) : '',
      change_summary: '通过插件管理页更新配置',
    })
    setSavingKey('')
    if (!res.success) {
      toast.error(res.error || '保存插件配置失败')
      return
    }
    toast.success('插件配置已保存')
    onSaved()
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
              <div className="mt-4 space-y-4">
                {(schema.sections || []).map((section) => (
                  <div key={section.id || section.title} className="rounded-lg bg-dark-900/40 p-3">
                    <div className="font-medium text-dark-200">{section.title || section.id || '未命名分组'}</div>
                    {section.description && <p className="mt-1 text-sm text-dark-500">{section.description}</p>}
                    <div className="mt-3 grid gap-3 md:grid-cols-2">
                      {(section.fields || []).map((field) => (
                        <ConfigFieldInput
                          key={field.key || field.id}
                          field={field}
                          configKey={schemaConfigKey(schema)}
                          value={drafts[schemaConfigKey(schema)]?.[fieldKey(field)]}
                          secretValue={secretDrafts[schemaConfigKey(schema)]?.[fieldKey(field)] || ''}
                          hasSavedSecret={Boolean(savedConfigValue(values, schemaConfigKey(schema))?.secret_json)}
                          onChange={updateDraft}
                          onSecretChange={updateSecretDraft}
                        />
                      ))}
                    </div>
                  </div>
                ))}
                {(schema.sections || []).length === 0 && (
                  <div className="rounded-lg bg-dark-900/40 p-3 text-sm text-dark-500">未声明配置分组。</div>
                )}
              </div>
              <div className="mt-4 flex justify-end border-t border-dark-700/70 pt-4">
                <Button
                  size="sm"
                  variant="success"
                  loading={savingKey === schemaConfigKey(schema)}
                  onClick={() => void saveSchemaConfig(schema)}
                >
                  <i className="fas fa-save" />
                  保存配置
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {values.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">当前配置值</h4>
          {values.map((value) => (
            <div key={value.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <div className="font-medium text-dark-100">{value.config_key}</div>
                  <div className="mt-1 text-xs text-dark-500">修订：{value.revision} · 更新人：{value.updated_by || '-'}</div>
                </div>
                <Badge variant={value.secret_json ? 'warning' : 'info'}>{value.secret_json ? '敏感值已隐藏' : '普通配置'}</Badge>
              </div>
              <div className="mt-3">
                {value.secret_json ? (
                  <div className="rounded-lg bg-dark-900/40 p-3 text-sm text-dark-400">配置值包含敏感配置，后台不回显明文。</div>
                ) : (
                  <JsonBlock title="Value JSON" value={parseJSON(value.value_json)} />
                )}
              </div>
              <div className="mt-3 text-xs text-dark-500">更新时间：{formatDateTime(value.updated_at)}</div>
            </div>
          ))}
        </div>
      )}

      {revisions.length > 0 && (
        <div className="space-y-3">
          <h4 className="font-medium text-dark-100">配置修订记录</h4>
          {revisions.map((revision) => (
            <div key={revision.id} className="rounded-xl border border-dark-700/70 bg-dark-800/40 p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="font-medium text-dark-100">{revision.config_key} · #{revision.revision}</div>
                <span className="text-xs text-dark-500">{formatDateTime(revision.created_at)}</span>
              </div>
              <div className="mt-3 grid gap-2 text-sm md:grid-cols-2">
                <InfoLine label="值摘要" value={revision.value_digest || '-'} />
                <InfoLine label="敏感配置" value={revision.secret_json ? '已保存' : '-'} />
                <InfoLine label="更新人" value={revision.updated_by || '-'} />
                <InfoLine label="说明" value={revision.change_summary || '-'} />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function ConfigFieldInput({
  field,
  configKey,
  value,
  secretValue,
  hasSavedSecret,
  onChange,
  onSecretChange,
}: {
  field: ConfigField
  configKey: string
  value: unknown
  secretValue: string
  hasSavedSecret: boolean
  onChange: (configKey: string, fieldKey: string, value: unknown) => void
  onSecretChange: (configKey: string, fieldKey: string, value: string) => void
}) {
  const key = fieldKey(field)
  const label = field.label || key
  const type = (field.type || 'text').toLowerCase()
  const commonLabel = (
    <div className="mb-1 flex items-center gap-1.5 text-sm font-medium text-dark-200">
      <span>{label}</span>
      {field.required && <span className="text-red-300">*</span>}
      {(field.secret || type === 'secret') && <span className="rounded bg-amber-500/20 px-1.5 py-0.5 text-[11px] text-amber-200">密钥</span>}
    </div>
  )
  const help = field.description ? <p className="mt-1 text-xs text-dark-500">{field.description}</p> : null

  if (field.secret || type === 'secret') {
    return (
      <div>
        {commonLabel}
        <Input
          type="password"
          value={secretValue}
          placeholder={hasSavedSecret ? '已保存，留空保持不变' : '请输入密钥'}
          onChange={(event) => onSecretChange(configKey, key, event.target.value)}
        />
        {help}
      </div>
    )
  }

  if (type === 'select') {
    const options = fieldOptions(field)
    return (
      <div>
        {commonLabel}
        <select
          value={stringValue(value)}
          onChange={(event) => onChange(configKey, key, event.target.value)}
          className="input h-12 w-full"
        >
          {options.map((option) => (
            <option key={String(option.value)} value={String(option.value)}>
              {option.label}
            </option>
          ))}
        </select>
        {help}
      </div>
    )
  }

  if (type === 'boolean' || type === 'bool') {
    return (
      <div className="rounded-lg bg-dark-950/30 p-3">
        <Switch
          checked={Boolean(value)}
          onChange={(checked) => onChange(configKey, key, checked)}
          label={label}
          description={field.description || ''}
        />
      </div>
    )
  }

  if (type === 'number' || type === 'integer') {
    return (
      <div>
        {commonLabel}
        <Input
          type="number"
          value={numberInputValue(value)}
          onChange={(event) => onChange(configKey, key, event.target.value === '' ? '' : Number(event.target.value))}
        />
        {help}
      </div>
    )
  }

  return (
    <div>
      {commonLabel}
      <Input
        value={stringValue(value)}
        onChange={(event) => onChange(configKey, key, event.target.value)}
      />
      {help}
    </div>
  )
}

function normalizeConfigSchema(raw: ConfigSchema) {
  const parsed = raw.schema_json ? parseJSON(raw.schema_json) : raw
  const schema = parsed && typeof parsed === 'object' && !Array.isArray(parsed)
    ? parsed as ConfigSchema
    : raw
  return {
    ...schema,
    schemaVersion: schema.schemaVersion || schema.schema_version || raw.schemaVersion || raw.schema_version,
    pluginId: schema.pluginId || schema.plugin_id || raw.pluginId || raw.plugin_id,
    configVersion: schema.configVersion || schema.config_version || raw.configVersion || raw.config_version,
    sections: schema.sections || [],
    secretPolicies: schema.secretPolicies || [],
    validationRules: schema.validationRules || [],
    reloadPolicies: schema.reloadPolicies || [],
    permissionGuards: schema.permissionGuards || [],
  }
}

function schemaConfigKey(_schema: ConfigSchema) {
  return 'default'
}

function savedConfigValue(values: PluginConfigValue[], configKey: string) {
  return values.find((value) => value.config_key === configKey)
}

function configDefaults(schema: ConfigSchema, saved: Record<string, unknown>) {
  const result: Record<string, unknown> = {}
  ;(schema.sections || []).forEach((section) => {
    ;(section.fields || []).forEach((field) => {
      const key = fieldKey(field)
      if (!key || field.secret || (field.type || '').toLowerCase() === 'secret') return
      if (saved[key] !== undefined) {
        result[key] = saved[key]
      } else if (field.default !== undefined) {
        result[key] = field.default
      } else if ((field.type || '').toLowerCase() === 'boolean' || (field.type || '').toLowerCase() === 'bool') {
        result[key] = false
      } else {
        result[key] = ''
      }
    })
  })
  return result
}

function fieldKey(field: ConfigField) {
  return field.key || field.id || ''
}

function fieldOptions(field: ConfigField) {
  if (field.options && field.options.length > 0) {
    return field.options.map((option) => {
      if (typeof option === 'object') {
        const value = option.value ?? option.label ?? ''
        return { label: String(option.label ?? value), value }
      }
      return { label: String(option), value: option }
    })
  }
  return (field.enumOptions || []).map((option) => ({ label: option, value: option }))
}

function objectFromJSON(value?: string) {
  const parsed = parseJSON(value || '')
  return parsed && typeof parsed === 'object' && !Array.isArray(parsed)
    ? parsed as Record<string, unknown>
    : {}
}

function stringValue(value: unknown) {
  return value === undefined || value === null ? '' : String(value)
}

function numberInputValue(value: unknown) {
  if (value === undefined || value === null || value === '') return ''
  return Number(value)
}

function firstMissingRequiredField(
  schema: ConfigSchema,
  draft: Record<string, unknown>,
  secretDraft: Record<string, string>,
  hasSavedSecret: boolean
) {
  for (const section of schema.sections || []) {
    for (const field of section.fields || []) {
      if (!field.required) continue
      const key = fieldKey(field)
      const label = field.label || key
      if (field.secret || (field.type || '').toLowerCase() === 'secret') {
        if (!secretDraft[key]?.trim() && !hasSavedSecret) return label
        continue
      }
      const value = draft[key]
      if (value === undefined || value === null || String(value).trim() === '') return label
    }
  }
  return ''
}

function collectSecretValues(schema: ConfigSchema, secretDraft: Record<string, string>) {
  const result: Record<string, string> = {}
  ;(schema.sections || []).forEach((section) => {
    ;(section.fields || []).forEach((field) => {
      const type = (field.type || '').toLowerCase()
      if (!field.secret && type !== 'secret') return
      const key = fieldKey(field)
      const value = secretDraft[key]?.trim()
      if (value) result[key] = value
    })
  })
  return result
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

function PluginCategoryBadges({ categories, className, limit = 4 }: { categories?: string[]; className?: string; limit?: number }) {
  const values = normalizeList(categories)
  if (values.length === 0) {
    return (
      <div className={cn('flex flex-wrap gap-1.5', className)}>
        <Badge variant="default">未分类</Badge>
      </div>
    )
  }
  const visible = values.slice(0, limit)
  const hiddenCount = values.length - visible.length
  return (
    <div className={cn('flex flex-wrap gap-1.5', className)}>
      {visible.map((category) => (
        <Badge key={category} variant="info">{formatPluginCategory(category)}</Badge>
      ))}
      {hiddenCount > 0 && <Badge variant="default">+{hiddenCount}</Badge>}
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

function normalizeDatabaseSnapshot(value?: PluginDatabaseSnapshot | null): PluginDatabaseSnapshot {
  return {
    declaration: value?.declaration || null,
    tables: value?.tables || [],
    columns: value?.columns || [],
    indexes: value?.indexes || [],
    relations: value?.relations || [],
    operations: value?.operations || [],
  }
}

function normalizeList(value?: string[] | null) {
  const values: string[] = []
  const seen = new Set<string>()
  ;(value || []).forEach((item) => {
    const normalized = String(item || '').trim()
    if (!normalized || seen.has(normalized)) return
    seen.add(normalized)
    values.push(normalized)
  })
  return values
}

function formatStringList(value?: string[] | null) {
  const values = normalizeList(value)
  return values.length > 0 ? values.join(', ') : '-'
}

function formatPluginCategory(value: string) {
  return PLUGIN_CATEGORY_LABELS[value] || value
}

function formatPluginCategoryList(value?: string[] | null) {
  const values = normalizeList(value)
  return values.length > 0 ? values.map(formatPluginCategory).join(', ') : '未分类'
}

function formatStringArray(value: unknown) {
  if (Array.isArray(value)) {
    return value.map((item) => String(item)).join(', ') || '-'
  }
  if (typeof value === 'string') {
    return value || '-'
  }
  return '-'
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
