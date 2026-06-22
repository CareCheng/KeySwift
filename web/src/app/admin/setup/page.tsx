'use client'

import AdminAuthFlow from '@/components/admin/AdminAuthFlow'

/**
 * 管理员初始化入口页。
 */
export default function AdminSetupPage() {
  return <AdminAuthFlow initialStep="setup" />
}
