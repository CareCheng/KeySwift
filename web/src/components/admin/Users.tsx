'use client'

import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { Button, Card } from '@/components/ui'
import { apiGet, apiPut } from '@/lib/api'
import { User } from './types'
import { PermissionGuard } from '@/contexts/PermissionContext'

/**
 * 用户管理页面
 * 支持移动端响应式布局
 */
export function UsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  const loadUsers = useCallback(async () => {
    const res = await apiGet<{ users: User[]; total_pages: number }>(`/api/admin/users?page=${page}&page_size=20`)
    if (res.success) {
      setUsers(res.users || [])
      setTotalPages(res.total_pages || 1)
    }
    setLoading(false)
  }, [page])

  useEffect(() => { loadUsers() }, [loadUsers])

  const toggleUserStatus = async (user: User) => {
    const newStatus = user.status === 1 ? 0 : 1
    const res = await apiPut(`/api/admin/user/${user.id}/status`, { status: newStatus })
    if (res.success) {
      toast.success(newStatus === 1 ? '用户已启用' : '用户已禁用')
      loadUsers()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  const securityBadges = (user: User) => [
    { label: '邮箱', active: user.email_verified },
    { label: 'TOTP', active: user.enable_2fa },
    { label: '支付密码', active: user.pay_password_set },
  ]

  if (loading) return <div className="text-center py-12"><i className="fas fa-spinner fa-spin text-2xl text-primary-400" /></div>

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium text-dark-100">用户列表</h2>
      <Card>
        {users.length === 0 ? (
          <div className="text-center py-12 text-dark-500">暂无用户</div>
        ) : (
          <>
            {/* 桌面端表格视图 */}
            <div className="hidden lg:block overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-dark-700">
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">ID</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">用户名</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">邮箱</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">手机</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">安全状态</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">订单/余额</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">状态</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">注册时间</th>
                    <th className="text-left py-3 px-4 text-dark-400 font-medium">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((user) => (
                    <tr key={user.id} className="border-b border-dark-700/50 hover:bg-dark-700/30">
                      <td className="py-3 px-4 text-dark-300">{user.id}</td>
                      <td className="py-3 px-4 text-dark-100">{user.username}</td>
                      <td className="py-3 px-4 text-dark-300">{user.email || '-'}</td>
                      <td className="py-3 px-4 text-dark-300">{user.phone || '-'}</td>
                      <td className="py-3 px-4">
                        <div className="flex flex-wrap gap-1">
                          {securityBadges(user).map((badge) => (
                            <span
                              key={badge.label}
                              className={`px-2 py-1 rounded text-xs ${badge.active ? 'bg-blue-500/20 text-blue-300' : 'bg-dark-700 text-dark-500'}`}
                            >
                              {badge.label}{badge.active ? '已启用' : '未启用'}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td className="py-3 px-4 text-dark-300 text-sm">
                        <div>订单 {user.order_count || 0} / 已付 {user.paid_order_count || 0}</div>
                        <div className="text-emerald-400">余额 ¥{(user.available_balance || 0).toFixed(2)}</div>
                      </td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded text-xs ${user.status === 1 ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                          {user.status === 1 ? '正常' : '禁用'}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-dark-300 text-sm">{user.created_at}</td>
                      <td className="py-3 px-4">
                        <PermissionGuard permission="user:edit">
                          <Button size="sm" variant="ghost" onClick={() => toggleUserStatus(user)}>
                            {user.status === 1 ? '禁用' : '启用'}
                          </Button>
                        </PermissionGuard>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* 移动端卡片视图 */}
            <div className="lg:hidden space-y-3">
              {users.map((user) => (
                <div key={user.id} className="bg-dark-700/30 rounded-lg p-4 space-y-2">
                  {/* 用户名和状态 */}
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div className="w-8 h-8 rounded-full bg-primary-500/20 flex items-center justify-center">
                        <span className="text-primary-400 text-sm font-medium">
                          {user.username.charAt(0).toUpperCase()}
                        </span>
                      </div>
                      <span className="text-dark-100 font-medium">{user.username}</span>
                    </div>
                    <span className={`px-2 py-1 rounded text-xs ${user.status === 1 ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                      {user.status === 1 ? '正常' : '禁用'}
                    </span>
                  </div>
                  {/* 联系方式 */}
                  <div className="text-sm space-y-1">
                    {user.email && (
                      <div className="flex items-center gap-2 text-dark-400">
                        <i className="fas fa-envelope w-4 text-center" />
                        <span className="text-dark-300 truncate">{user.email}</span>
                      </div>
                    )}
                    {user.phone && (
                      <div className="flex items-center gap-2 text-dark-400">
                        <i className="fas fa-phone w-4 text-center" />
                        <span className="text-dark-300">{user.phone}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2 text-dark-400">
                      <i className="fas fa-clock w-4 text-center" />
                      <span className="text-dark-500 text-xs">{user.created_at}</span>
                    </div>
                    <div className="flex flex-wrap gap-1 pt-1">
                      {securityBadges(user).map((badge) => (
                        <span
                          key={badge.label}
                          className={`px-2 py-1 rounded text-xs ${badge.active ? 'bg-blue-500/20 text-blue-300' : 'bg-dark-700 text-dark-500'}`}
                        >
                          {badge.label}{badge.active ? '已启用' : '未启用'}
                        </span>
                      ))}
                    </div>
                    <div className="grid grid-cols-2 gap-2 pt-1 text-xs">
                      <div className="rounded bg-dark-800/60 p-2">
                        <div className="text-dark-500">订单</div>
                        <div className="text-dark-200">{user.order_count || 0} 笔，已付 {user.paid_order_count || 0}</div>
                      </div>
                      <div className="rounded bg-dark-800/60 p-2">
                        <div className="text-dark-500">余额</div>
                        <div className="text-emerald-400">¥{(user.available_balance || 0).toFixed(2)}</div>
                      </div>
                    </div>
                  </div>
                  {/* 操作按钮 */}
                  <div className="pt-2 border-t border-dark-600/50">
                    <PermissionGuard permission="user:edit">
                      <Button
                        size="sm"
                        variant={user.status === 1 ? 'ghost' : 'secondary'}
                        onClick={() => toggleUserStatus(user)}
                        className="w-full"
                      >
                        <i className={`fas fa-${user.status === 1 ? 'ban' : 'check'} mr-2`} />
                        {user.status === 1 ? '禁用用户' : '启用用户'}
                      </Button>
                    </PermissionGuard>
                  </div>
                </div>
              ))}
            </div>

            {/* 分页 - 移动端简化 */}
            {totalPages > 1 && (
              <div className="flex flex-wrap justify-center gap-2 mt-4">
                <Button size="sm" variant="secondary" disabled={page <= 1} onClick={() => setPage(1)} className="hidden sm:inline-flex">首页</Button>
                <Button size="sm" variant="secondary" disabled={page <= 1} onClick={() => setPage(page - 1)}>
                  <i className="fas fa-chevron-left sm:mr-1" /><span className="hidden sm:inline">上一页</span>
                </Button>
                <span className="px-3 py-2 text-dark-400 text-sm">{page} / {totalPages}</span>
                <Button size="sm" variant="secondary" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
                  <span className="hidden sm:inline">下一页</span><i className="fas fa-chevron-right sm:ml-1" />
                </Button>
                <Button size="sm" variant="secondary" disabled={page >= totalPages} onClick={() => setPage(totalPages)} className="hidden sm:inline-flex">末页</Button>
              </div>
            )}
          </>
        )}
      </Card>
    </div>
  )
}
