'use client'

import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import { apiGet } from '@/lib/api'

/**
 * 管理员信息接口
 */
interface AdminInfo {
  username: string
  role: string
  permissions: string[]
  is_super_admin: boolean
}

/**
 * 权限上下文接口
 */
interface PermissionContextType {
  // 管理员信息
  adminInfo: AdminInfo | null
  // 是否正在加载
  loading: boolean
  // 是否是超级管理员
  isSuperAdmin: boolean
  // 权限列表
  permissions: string[]
  // 检查是否有指定权限
  hasPermission: (permission: string) => boolean
  // 检查是否有任意一个权限
  hasAnyPermission: (permissions: string[]) => boolean
  // 检查是否有所有权限
  hasAllPermissions: (permissions: string[]) => boolean
  // 刷新权限信息
  refreshPermissions: () => Promise<void>
}

// 创建上下文
const PermissionContext = createContext<PermissionContextType | undefined>(undefined)

/**
 * 权限提供者组件
 */
export function PermissionProvider({ children }: { children: ReactNode }) {
  const [adminInfo, setAdminInfo] = useState<AdminInfo | null>(null)
  const [loading, setLoading] = useState(true)

  // 加载权限信息
  const loadPermissions = useCallback(async () => {
    setLoading(true)
    try {
      const res = await apiGet<{ admin: AdminInfo }>('/api/admin/info')
      if (res.success && res.admin) {
        setAdminInfo(res.admin)
      }
    } catch (error) {
      console.error('加载权限信息失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  // 初始化加载
  useEffect(() => {
    loadPermissions()
  }, [loadPermissions])

  // 是否是超级管理员
  const isSuperAdmin = adminInfo?.is_super_admin || false

  // 权限列表
  const permissions = adminInfo?.permissions || []

  // 检查是否有指定权限
  const hasPermission = useCallback((permission: string): boolean => {
    if (isSuperAdmin) return true
    return permissions.includes(permission)
  }, [isSuperAdmin, permissions])

  // 检查是否有任意一个权限
  const hasAnyPermission = useCallback((perms: string[]): boolean => {
    if (isSuperAdmin) return true
    return perms.some(p => permissions.includes(p))
  }, [isSuperAdmin, permissions])

  // 检查是否有所有权限
  const hasAllPermissions = useCallback((perms: string[]): boolean => {
    if (isSuperAdmin) return true
    return perms.every(p => permissions.includes(p))
  }, [isSuperAdmin, permissions])

  // 刷新权限信息
  const refreshPermissions = useCallback(async () => {
    await loadPermissions()
  }, [loadPermissions])

  return (
    <PermissionContext.Provider
      value={{
        adminInfo,
        loading,
        isSuperAdmin,
        permissions,
        hasPermission,
        hasAnyPermission,
        hasAllPermissions,
        refreshPermissions,
      }}
    >
      {children}
    </PermissionContext.Provider>
  )
}

/**
 * 使用权限上下文的 Hook
 */
export function usePermission() {
  const context = useContext(PermissionContext)
  if (context === undefined) {
    throw new Error('usePermission 必须在 PermissionProvider 内部使用')
  }
  return context
}

/**
 * 权限检查组件 - 用于按钮级权限控制
 * 如果没有权限，则不渲染子组件
 */
interface PermissionGuardProps {
  // 需要的权限（满足任意一个即可）
  permissions?: string[]
  // 需要的权限（必须满足所有）
  allPermissions?: string[]
  // 单个权限
  permission?: string
  // 没有权限时显示的内容（可选）
  fallback?: ReactNode
  // 子组件
  children: ReactNode
}

export function PermissionGuard({
  permissions,
  allPermissions,
  permission,
  fallback = null,
  children,
}: PermissionGuardProps) {
  const { hasPermission, hasAnyPermission, hasAllPermissions, loading } = usePermission()

  // 加载中不显示
  if (loading) return null

  // 检查权限
  let hasAccess = true

  if (permission) {
    hasAccess = hasPermission(permission)
  } else if (permissions && permissions.length > 0) {
    hasAccess = hasAnyPermission(permissions)
  } else if (allPermissions && allPermissions.length > 0) {
    hasAccess = hasAllPermissions(allPermissions)
  }

  // 没有权限显示 fallback
  if (!hasAccess) {
    return <>{fallback}</>
  }

  return <>{children}</>
}

/**
 * 页面权限映射 - 定义每个页面需要的权限
 */
export const PAGE_PERMISSIONS: Record<string, string[]> = {
  dashboard: ['dashboard:view'],
  products: ['product:view'],
  categories: ['category:view'],
  orders: ['order:view'],
  users: ['user:view'],
  balance: ['balance:view'],
  access: ['role:view', 'admin:view'],
  config: ['settings:view', 'settings:security', 'settings:payment', 'settings:email', 'settings:database'],
  plugins: ['plugin:view'],
}

/**
 * 检查用户是否可以访问指定页面
 */
export function canAccessPage(pageKey: string, permissions: string[], isSuperAdmin: boolean): boolean {
  if (isSuperAdmin) return true
  
  const requiredPermissions = PAGE_PERMISSIONS[pageKey]
  if (!requiredPermissions || requiredPermissions.length === 0) {
    return true // 没有定义权限要求的页面默认可访问
  }
  
  // 只要有任意一个权限就可以访问
  return requiredPermissions.some(p => permissions.includes(p))
}
