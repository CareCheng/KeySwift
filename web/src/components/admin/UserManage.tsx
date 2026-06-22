'use client'

import { Card } from '@/components/ui'
import { UsersPage } from './Users'

/**
 * 用户管理入口。
 * 只承载用户账号、安全状态和用户域操作，资金与后台权限已拆到独立入口。
 */
export function UserManagePage() {
  return (
    <div className="space-y-6">
      <Card>
        <div className="flex flex-col gap-2">
          <h2 className="text-xl font-semibold text-dark-100">用户管理</h2>
          <p className="text-sm text-dark-400">
            管理用户账号状态、安全能力和用户关联摘要；余额资金与后台角色权限请在对应独立入口处理。
          </p>
        </div>
      </Card>
      <UsersPage />
    </div>
  )
}
