'use client'

import AdminDashboardApp from '@/components/admin/AdminDashboardApp'

/**
 * 管理后台静态导出入口。
 * 实际后台路径由 Go 服务按自定义后缀映射到该页面。
 */
export default function AdminPage() {
  return <AdminDashboardApp />
}
