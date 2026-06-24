import { apiGet } from './api'

export type PluginRenderMode = 'host-rendered' | 'bundle' | 'iframe'
export type PluginMountArea = 'admin' | 'user' | 'settings' | string

export interface AdminPageDefinition {
  id: string
  title: string
  icon: string
  permissions?: string[]
  source: 'host' | 'plugin'
  pluginId?: string
  renderMode?: PluginRenderMode
  path?: string
  viewId?: string
  order?: number
}

export interface PluginPageDeclaration {
  id: string
  area: PluginMountArea
  path: string
  title: string
  viewId?: string
  renderMode?: PluginRenderMode
  permissionKeys?: string[]
  allowDirectAccess?: boolean
  visible?: boolean
}

export interface PluginMenuDeclaration {
  id: string
  targetPageId: string
  title: string
  icon?: string
  defaultGroup?: string
  order?: number
  visible?: boolean
  allowWorkspacePin?: boolean
  permissionKeys?: string[]
}

export interface PluginThemeContribution {
  enabled: boolean
  uiKind?: string
  themeScope?: string
  tokenExtensions?: Record<string, string>
  componentOverridesRef?: string
  layoutSkinRef?: string
  iconPackRef?: string
  activationPolicy?: string
}

export interface PluginFrontendContribution {
  protocolVersion: string
  pages: PluginPageDeclaration[]
  menus: PluginMenuDeclaration[]
  forms: unknown[]
  views: unknown[]
  themes: PluginThemeContribution[]
}

export interface PluginFrontendResponse {
  frontend: PluginFrontendContribution
}

export const PLUGIN_FRONTEND_CHANGED_EVENT = 'plugins:frontend-changed'

export const HOST_ADMIN_PAGES: Record<string, AdminPageDefinition> = {
  dashboard: { id: 'dashboard', title: '仪表盘', icon: '📊', permissions: ['dashboard:view'], source: 'host', order: 10 },
  products: { id: 'products', title: '商品管理', icon: '📦', permissions: ['product:view'], source: 'host', order: 20 },
  categories: { id: 'categories', title: '分类管理', icon: '📁', permissions: ['category:view'], source: 'host', order: 30 },
  orders: { id: 'orders', title: '订单管理', icon: '📋', permissions: ['order:view'], source: 'host', order: 40 },
  users: { id: 'users', title: '用户管理', icon: '👥', permissions: ['user:view'], source: 'host', order: 50 },
  balance: { id: 'balance', title: '余额管理', icon: '💳', permissions: ['balance:view'], source: 'host', order: 60 },
  access: { id: 'access', title: '权限与管理员', icon: '🛡️', permissions: ['role:view', 'admin:view'], source: 'host', order: 70 },
  config: { id: 'config', title: '系统配置', icon: '⚙️', permissions: ['settings:view', 'settings:security', 'settings:payment', 'settings:email', 'settings:database'], source: 'host', order: 80 },
  plugins: { id: 'plugins', title: '插件管理', icon: '🧩', permissions: ['plugin:view'], source: 'host', order: 90 },
}

export function buildAdminPageRegistry(contribution?: PluginFrontendContribution | null): Record<string, AdminPageDefinition> {
  const registry: Record<string, AdminPageDefinition> = { ...HOST_ADMIN_PAGES }

  if (!contribution?.pages?.length) {
    return registry
  }

  for (const page of contribution.pages) {
    if (page.area !== 'admin' || page.visible === false) continue

    const pageId = `plugin:${page.id}`
    const menu = contribution.menus?.find((item) => item.targetPageId === page.id)
    registry[pageId] = {
      id: pageId,
      title: menu?.title || page.title || page.id,
      icon: menu?.icon || '🧩',
      permissions: page.permissionKeys || menu?.permissionKeys || [],
      source: 'plugin',
      pluginId: inferPluginId(page.id),
      renderMode: page.renderMode || 'host-rendered',
      path: page.path,
      viewId: page.viewId,
      order: menu?.order ?? 1000,
    }
  }

  return Object.fromEntries(
    Object.entries(registry).sort(([, a], [, b]) => (a.order ?? 0) - (b.order ?? 0)),
  )
}

export async function loadPluginFrontendContribution() {
  return apiGet<PluginFrontendResponse>('/api/admin/plugins/frontend')
}

function inferPluginId(pageId: string) {
  const [pluginId] = pageId.split(':')
  return pluginId || pageId
}
