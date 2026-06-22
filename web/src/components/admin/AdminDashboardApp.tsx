'use client'

import { useState, useEffect, useCallback, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui'
import { apiGet, apiPost } from '@/lib/api'
import { PermissionProvider } from '@/contexts/PermissionContext'
import { DashboardPage } from './Dashboard'
import { CategoriesPage } from './Categories'
import { OrdersPage } from './Orders'
import { ConfigManagePage } from './ConfigManage'
import { ProductsPage } from './products'
import { UserManagePage } from './UserManage'
import { BalancePage } from './Balance'
import { RolesPage } from './Roles'
import { PluginsPage } from './Plugins'
import {
  AdminPageDefinition,
  HOST_ADMIN_PAGES,
  PLUGIN_FRONTEND_CHANGED_EVENT,
  PluginFrontendContribution,
  buildAdminPageRegistry,
  loadPluginFrontendContribution,
} from '@/lib/pluginRegistry'

/**
 * 管理员信息接口
 */
interface AdminInfo {
  username: string
  role: string
  permissions: string[]
  is_super_admin: boolean
}

interface SidebarContentProps {
  sidebarCollapsed: boolean
  accessiblePages: string[]
  pageRegistry: Record<string, AdminPageDefinition>
  currentPage: string
  onPageChange: (page: string) => void
  onToggleCollapse: () => void
}

interface AdminDashboardAppProps {
  onAuthRequired?: (basePath: string) => void
  onLogout?: (basePath: string) => void
}

function SidebarContent({
  sidebarCollapsed,
  accessiblePages,
  pageRegistry,
  currentPage,
  onPageChange,
  onToggleCollapse,
}: SidebarContentProps) {
  return (
    <>
      <div className={`p-4 lg:p-6 border-b border-dark-700 flex items-center ${sidebarCollapsed ? 'justify-center' : 'justify-between'}`}>
        {!sidebarCollapsed && (
          <h2 className="text-lg lg:text-xl font-bold text-dark-100 truncate">管理后台</h2>
        )}
        <button
          onClick={onToggleCollapse}
          className="hidden lg:flex items-center justify-center w-8 h-8 rounded-lg hover:bg-dark-700/50 text-dark-400 hover:text-dark-200 transition-colors"
          title={sidebarCollapsed ? '展开侧边栏' : '折叠侧边栏'}
        >
          <i className={`fas fa-chevron-${sidebarCollapsed ? 'right' : 'left'} text-sm`} />
        </button>
      </div>
      <nav className="flex-1 py-2 lg:py-4 overflow-y-auto">
        {Object.entries(pageRegistry)
          .filter(([key]) => accessiblePages.includes(key))
          .map(([key, config]) => (
            <button
              key={key}
              onClick={() => onPageChange(key)}
              className={`w-full px-4 lg:px-6 py-2.5 lg:py-3 text-left flex items-center gap-3 transition-colors ${
                currentPage === key
                  ? 'bg-primary-500/20 text-primary-400 border-r-2 border-primary-400'
                  : 'text-dark-400 hover:bg-dark-700/50 hover:text-dark-200'
              } ${sidebarCollapsed ? 'justify-center px-2' : ''}`}
              title={sidebarCollapsed ? config.title : undefined}
            >
              <span className="text-lg">{config.icon}</span>
              {!sidebarCollapsed && <span className="truncate">{config.title}</span>}
            </button>
          ))}
      </nav>
    </>
  )
}

/**
 * 获取当前管理后台路径前缀
 */
function getAdminBasePath() {
  const path = window.location.pathname
  const cleanPath = path.replace(/\/$/, '').split('#')[0]
  return cleanPath || '/'
}

/**
 * 检查是否有页面访问权限
 */
function canAccessPage(pageKey: string, permissions: string[], isSuperAdmin: boolean): boolean {
  if (isSuperAdmin) return true
  
  const pageConfig = HOST_ADMIN_PAGES[pageKey]
  if (!pageConfig?.permissions || pageConfig.permissions.length === 0) {
    return true // 没有定义权限要求的页面默认可访问
  }
  
  // 只要有任意一个权限就可以访问
  return pageConfig.permissions.some(p => permissions.includes(p))
}

function canAccessPageByConfig(
  pageKey: string,
  config: AdminPageDefinition,
  permissions: string[],
  isSuperAdmin: boolean,
) {
  if (isSuperAdmin) return true
  if (config.source === 'host') {
    return canAccessPage(pageKey, permissions, isSuperAdmin)
  }
  if (!config.permissions || config.permissions.length === 0) {
    return true
  }
  return config.permissions.some(p => permissions.includes(p))
}

/**
 * 管理后台主界面。
 * 支持移动端响应式布局和权限控制，可被后台入口页和认证流复用。
 */
export default function AdminDashboardApp({
  onAuthRequired,
  onLogout,
}: AdminDashboardAppProps = {}) {
  return (
    <PermissionProvider>
      <AdminDashboardContent onAuthRequired={onAuthRequired} onLogout={onLogout} />
    </PermissionProvider>
  )
}

function AdminDashboardContent({
  onAuthRequired,
  onLogout,
}: AdminDashboardAppProps = {}) {
  const [currentPage, setCurrentPage] = useState('dashboard')
  const [adminInfo, setAdminInfo] = useState<AdminInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [basePath, setBasePath] = useState('')
  const [pluginFrontend, setPluginFrontend] = useState<PluginFrontendContribution | null>(null)
  // 移动端侧边栏状态
  const [sidebarOpen, setSidebarOpen] = useState(false)
  // 侧边栏折叠状态（桌面端）
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)

  const pageRegistry = useMemo(() => {
    return buildAdminPageRegistry(pluginFrontend)
  }, [pluginFrontend])

  const loadPluginFrontend = useCallback(async () => {
    const res = await loadPluginFrontendContribution()
    if (res.success && res.frontend) {
      setPluginFrontend(res.frontend)
    }
  }, [])

  // 计算可访问的页面列表
  const accessiblePages = useMemo(() => {
    if (!adminInfo) return []
    const permissions = adminInfo.permissions || []
    const isSuperAdmin = adminInfo.is_super_admin || false
    
    return Object.entries(pageRegistry).filter(([key, config]) =>
      canAccessPageByConfig(key, config, permissions, isSuperAdmin)
    ).map(([key]) => key)
  }, [adminInfo, pageRegistry])

  useEffect(() => {
    loadPluginFrontend()
    window.addEventListener(PLUGIN_FRONTEND_CHANGED_EVENT, loadPluginFrontend)
    return () => window.removeEventListener(PLUGIN_FRONTEND_CHANGED_EVENT, loadPluginFrontend)
  }, [loadPluginFrontend])

  useEffect(() => {
    const adminBase = getAdminBasePath()
    setBasePath(adminBase)
    
    const checkAuth = async () => {
      const res = await apiGet<{ admin: AdminInfo }>('/api/admin/info')
      if (res.success && res.admin) {
        setAdminInfo(res.admin)
      } else {
        if (onAuthRequired) {
          onAuthRequired(adminBase)
        } else {
          window.location.href = `${adminBase}/login/`
        }
      }
      setLoading(false)
    }
    checkAuth()

  }, [])

  useEffect(() => {
    const handleHashChange = () => {
      const hash = window.location.hash.slice(1)
      if (hash && pageRegistry[hash]) {
        setCurrentPage(hash)
      }
    }
    handleHashChange()
    window.addEventListener('hashchange', handleHashChange)
    return () => window.removeEventListener('hashchange', handleHashChange)
  }, [pageRegistry])

  // 确保当前页面是有权限访问的，否则跳转到第一个有权限的页面
  useEffect(() => {
    if (!loading && adminInfo && accessiblePages.length > 0) {
      if (!accessiblePages.includes(currentPage)) {
        const firstAccessiblePage = accessiblePages[0]
        setCurrentPage(firstAccessiblePage)
        window.location.hash = firstAccessiblePage
      }
    }
  }, [loading, adminInfo, accessiblePages, currentPage])

  // 监听窗口大小变化，自动关闭移动端侧边栏
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= 1024) {
        setSidebarOpen(false)
      }
    }
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  const handlePageChange = useCallback((page: string) => {
    setCurrentPage(page)
    window.location.hash = page
    // 移动端切换页面后自动关闭侧边栏
    setSidebarOpen(false)
  }, [])

  const handleLogout = async () => {
    await apiPost('/api/admin/logout', {})
    if (onLogout) {
      onLogout(basePath)
    } else {
      window.location.href = `${basePath}/login/`
    }
  }

  // 切换移动端侧边栏
  const toggleSidebar = useCallback(() => {
    setSidebarOpen(prev => !prev)
  }, [])

  // 切换桌面端侧边栏折叠
  const toggleSidebarCollapse = useCallback(() => {
    setSidebarCollapsed(prev => !prev)
  }, [])

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-dark-900">
        <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
      </div>
    )
  }

  // 渲染当前页面内容
  const renderPageContent = () => {
    switch (currentPage) {
      case 'dashboard': return <DashboardPage />
      case 'products': return <ProductsPage />
      case 'categories': return <CategoriesPage />
      case 'orders': return <OrdersPage />
      case 'users': return <UserManagePage />
      case 'balance': return <BalancePage />
      case 'access': return <RolesPage />
      case 'config': return <ConfigManagePage />
      case 'plugins': return <PluginsPage />
      default:
        if (currentPage.startsWith('plugin:')) {
          return <PluginPageHost page={pageRegistry[currentPage]} />
        }
        return <DashboardPage />
    }
  }

  return (
    <div className="min-h-screen flex bg-dark-900">
      {/* 移动端遮罩层 */}
      <AnimatePresence>
        {sidebarOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-0 bg-black/60 z-40 lg:hidden"
            onClick={() => setSidebarOpen(false)}
          />
        )}
      </AnimatePresence>

      {/* 移动端侧边栏 */}
      <AnimatePresence>
        {sidebarOpen && (
          <motion.aside
            initial={{ x: -280 }}
            animate={{ x: 0 }}
            exit={{ x: -280 }}
            transition={{ type: 'spring', damping: 25, stiffness: 300 }}
            className="fixed left-0 top-0 bottom-0 w-[280px] bg-dark-800 border-r border-dark-700 flex flex-col z-50 lg:hidden"
          >
            {/* 移动端关闭按钮 */}
            <button
              onClick={() => setSidebarOpen(false)}
              className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center rounded-lg hover:bg-dark-700/50 text-dark-400 hover:text-dark-200 transition-colors"
            >
              <i className="fas fa-times" />
            </button>
            <SidebarContent
              sidebarCollapsed={sidebarCollapsed}
              accessiblePages={accessiblePages}
              pageRegistry={pageRegistry}
              currentPage={currentPage}
              onPageChange={handlePageChange}
              onToggleCollapse={toggleSidebarCollapse}
            />
          </motion.aside>
        )}
      </AnimatePresence>

      {/* 桌面端侧边栏 */}
      <aside 
        className={`hidden lg:flex flex-col bg-dark-800 border-r border-dark-700 transition-all duration-300 ${
          sidebarCollapsed ? 'w-16' : 'w-64'
        }`}
      >
        <SidebarContent
          sidebarCollapsed={sidebarCollapsed}
          accessiblePages={accessiblePages}
          pageRegistry={pageRegistry}
          currentPage={currentPage}
          onPageChange={handlePageChange}
          onToggleCollapse={toggleSidebarCollapse}
        />
      </aside>

      {/* 主内容区 */}
      <main className="flex-1 flex flex-col min-w-0">
        <header className="h-14 lg:h-16 bg-dark-800 border-b border-dark-700 flex items-center justify-between px-4 lg:px-6 sticky top-0 z-30">
          <div className="flex items-center gap-3">
            {/* 移动端菜单按钮 */}
            <button
              onClick={toggleSidebar}
              className="lg:hidden w-10 h-10 flex items-center justify-center rounded-lg hover:bg-dark-700/50 text-dark-400 hover:text-dark-200 transition-colors"
            >
              <i className="fas fa-bars text-lg" />
            </button>
            <h1 className="text-base lg:text-lg font-medium text-dark-100 truncate">
              {pageRegistry[currentPage]?.title || '管理后台'}
            </h1>
          </div>
          <div className="flex items-center gap-2 lg:gap-4">
            <span className="text-dark-400 text-sm hidden sm:inline">{adminInfo?.username}</span>
            <Button size="sm" variant="ghost" onClick={handleLogout} className="text-sm">
              <span className="hidden sm:inline">退出</span>
              <i className="fas fa-sign-out-alt sm:hidden" />
            </Button>
          </div>
        </header>

        <div className="flex-1 p-3 sm:p-4 lg:p-6 overflow-y-auto">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentPage}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              transition={{ duration: 0.2 }}
            >
              {renderPageContent()}
            </motion.div>
          </AnimatePresence>
        </div>
      </main>
    </div>
  )
}

function PluginPageHost({ page }: { page?: AdminPageDefinition }) {
  if (!page) {
    return <DashboardPage />
  }

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-lg font-medium" style={{ color: 'var(--text-primary)' }}>{page.title}</h2>
        <p className="text-sm mt-1" style={{ color: 'var(--text-muted)' }}>
          当前页面来自已启用插件，请在插件管理中查看入口信息和运行状态。
        </p>
      </div>
      <div className="rounded-xl border p-6" style={{ borderColor: 'var(--border-color)', backgroundColor: 'var(--card-bg)' }}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <InfoItem label="插件" value={page.pluginId || '-'} />
          <InfoItem label="加载方式" value={page.renderMode || 'host-rendered'} />
          <InfoItem label="页面路径" value={page.path || '-'} />
          <InfoItem label="视图标识" value={page.viewId || '-'} />
        </div>
      </div>
    </div>
  )
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg px-3 py-2" style={{ backgroundColor: 'var(--bg-tertiary)' }}>
      <div className="text-xs" style={{ color: 'var(--text-muted)' }}>{label}</div>
      <div className="font-mono text-sm mt-1 break-all" style={{ color: 'var(--text-primary)' }}>{value}</div>
    </div>
  )
}
