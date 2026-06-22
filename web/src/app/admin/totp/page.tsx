'use client'

import AdminAuthFlow from '@/components/admin/AdminAuthFlow'

/**
 * 管理员两步验证入口页。
 */
export default function AdminTOTPPage() {
  return <AdminAuthFlow initialStep="totp" />
}
