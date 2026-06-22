'use client'

import AdminAuthFlow from '@/components/admin/AdminAuthFlow'

/**
 * 管理员登录入口页。
 */
export default function AdminLoginPage() {
  return <AdminAuthFlow initialStep="login" />
}
