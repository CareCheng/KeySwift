'use client'

import { useState, useEffect, ChangeEvent } from 'react'
import toast from 'react-hot-toast'
import { Button, Card, Badge, Modal, Input } from '@/components/ui'
import { ConfirmModal } from '@/components/ui/ConfirmModal'
import Toggle from '@/components/common/Toggle'
import { apiGet, apiPost, apiPut, apiDelete } from '@/lib/api'
import { formatDateTime } from '@/lib/utils'
import { PermissionGuard } from '@/contexts/PermissionContext'

/**
 * 角色接口
 */
interface Role {
  id: number
  name: string
  description: string
  permissions: string[]
  is_system: boolean
  status: number
  created_at: string
}

/**
 * 管理员接口
 */
interface Admin {
  id: number
  username: string
  email: string
  nickname: string
  role_id: number
  role_name: string
  enable_2fa: boolean
  status: number
  last_login_at: string
  last_login_ip: string
  created_at: string
}

/**
 * 权限接口
 */
interface Permission {
  code: string
  name: string
  description?: string
  group: string
  plugin_id?: string
  risk_level?: string
}

/**
 * 权限模板接口
 */
interface PermissionTemplate {
  name: string
  description: string
  permissions: string[]
}

/**
 * 角色权限管理页面
 */
export function RolesPage() {
  const [activeTab, setActiveTab] = useState<'roles' | 'admins'>('roles')
  const [roles, setRoles] = useState<Role[]>([])
  const [admins, setAdmins] = useState<Admin[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [templates, setTemplates] = useState<PermissionTemplate[]>([])
  const [loading, setLoading] = useState(true)
  const [showRoleModal, setShowRoleModal] = useState(false)
  const [showAdminModal, setShowAdminModal] = useState(false)
  const [editingRole, setEditingRole] = useState<Role | null>(null)
  const [editingAdmin, setEditingAdmin] = useState<Admin | null>(null)

  // 角色表单
  const [roleForm, setRoleForm] = useState({
    name: '',
    description: '',
    permissions: [] as string[],
  })

  // 管理员表单
  const [adminForm, setAdminForm] = useState({
    username: '',
    password: '',
    email: '',
    nickname: '',
    role_id: 0,
    status: 1,
  })

  // 删除确认弹窗状态
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteTarget, setDeleteTarget] = useState<{ type: 'role' | 'admin'; id: number; name: string } | null>(null)
  const [deleteLoading, setDeleteLoading] = useState(false)

  // 加载数据
  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setLoading(true)
    await Promise.all([loadRoles(), loadAdmins(), loadPermissions()])
    setLoading(false)
  }

  const loadRoles = async () => {
    const res = await apiGet<{ data: Role[] }>('/api/admin/roles')
    if (res.success && res.data) {
      setRoles(res.data)
    }
  }

  const loadAdmins = async () => {
    const res = await apiGet<{ data: Admin[] }>('/api/admin/admins')
    if (res.success && res.data) {
      setAdmins(res.data)
    }
  }

  const loadPermissions = async () => {
    const res = await apiGet<{
      permissions: Permission[]
      plugin_permissions?: Permission[]
      templates: PermissionTemplate[]
    }>('/api/admin/permissions')
    if (res.success && res.permissions) {
      setPermissions(res.permissions)
    }
    if (res.success && res.templates) {
      setTemplates(res.templates)
    }
  }

  // 应用权限模板
  const applyTemplate = (templateName: string) => {
    const template = templates.find(t => t.name === templateName)
    if (template) {
      setRoleForm(prev => ({
        ...prev,
        permissions: [...template.permissions],
      }))
      toast.success(`已应用「${template.description.split(' - ')[0]}」模板`)
    }
  }

  // 打开角色编辑弹窗
  const openRoleModal = (role?: Role) => {
    if (role) {
      setEditingRole(role)
      setRoleForm({
        name: role.name,
        description: role.description,
        permissions: role.permissions || [],
      })
    } else {
      setEditingRole(null)
      setRoleForm({ name: '', description: '', permissions: [] })
    }
    setShowRoleModal(true)
  }

  // 保存角色
  const handleSaveRole = async () => {
    if (!roleForm.name) {
      toast.error('请填写角色名称')
      return
    }
    const data = {
      name: roleForm.name,
      description: roleForm.description,
      permissions: roleForm.permissions,
    }
    let res
    if (editingRole) {
      res = await apiPut(`/api/admin/role/${editingRole.id}`, data)
    } else {
      res = await apiPost('/api/admin/role', data)
    }
    if (res.success) {
      toast.success(editingRole ? '角色已更新' : '角色已创建')
      setShowRoleModal(false)
      loadRoles()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 打开删除确认弹窗
  const openDeleteConfirm = (type: 'role' | 'admin', id: number, name: string) => {
    setDeleteTarget({ type, id, name })
    setShowDeleteConfirm(true)
  }

  // 执行删除
  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleteLoading(true)
    const url = deleteTarget.type === 'role'
      ? `/api/admin/role/${deleteTarget.id}`
      : `/api/admin/admin/${deleteTarget.id}`
    const res = await apiDelete(url)
    setDeleteLoading(false)
    if (res.success) {
      toast.success(deleteTarget.type === 'role' ? '角色已删除' : '管理员已删除')
      setShowDeleteConfirm(false)
      setDeleteTarget(null)
      if (deleteTarget.type === 'role') {
        loadRoles()
      } else {
        loadAdmins()
      }
    } else {
      toast.error(res.error || '删除失败')
    }
  }

  // 删除角色（打开确认弹窗）
  const handleDeleteRole = (id: number) => {
    const role = roles.find(r => r.id === id)
    openDeleteConfirm('role', id, role?.name || '')
  }

  // 删除管理员（打开确认弹窗）
  const handleDeleteAdmin = (id: number) => {
    const admin = admins.find(a => a.id === id)
    openDeleteConfirm('admin', id, admin?.username || '')
  }

  // 打开管理员编辑弹窗
  const openAdminModal = (admin?: Admin) => {
    if (admin) {
      setEditingAdmin(admin)
      setAdminForm({
        username: admin.username,
        password: '',
        email: admin.email,
        nickname: admin.nickname,
        role_id: admin.role_id,
        status: admin.status,
      })
    } else {
      setEditingAdmin(null)
      setAdminForm({ username: '', password: '', email: '', nickname: '', role_id: roles[0]?.id || 0, status: 1 })
    }
    setShowAdminModal(true)
  }

  // 保存管理员
  const handleSaveAdmin = async () => {
    if (!adminForm.username) {
      toast.error('请填写用户名')
      return
    }
    if (!editingAdmin && !adminForm.password) {
      toast.error('请填写密码')
      return
    }
    let res
    if (editingAdmin) {
      res = await apiPut(`/api/admin/admin/${editingAdmin.id}`, adminForm)
    } else {
      res = await apiPost('/api/admin/admin', adminForm)
    }
    if (res.success) {
      toast.success(editingAdmin ? '管理员已更新' : '管理员已创建')
      setShowAdminModal(false)
      loadAdmins()
    } else {
      toast.error(res.error || '操作失败')
    }
  }

  // 切换权限
  const togglePermission = (code: string) => {
    setRoleForm(prev => ({
      ...prev,
      permissions: prev.permissions.includes(code)
        ? prev.permissions.filter(p => p !== code)
        : [...prev.permissions, code],
    }))
  }

  // 按组分类权限
  const groupedPermissions = permissions.reduce((acc, perm) => {
    if (!acc[perm.group]) acc[perm.group] = []
    acc[perm.group].push(perm)
    return acc
  }, {} as Record<string, Permission[]>)

  const hostPermissionGroups = Object.entries(groupedPermissions).filter(([group]) => !group.startsWith('插件权限'))
  const pluginPermissionGroups = Object.entries(groupedPermissions).filter(([group]) => group.startsWith('插件权限'))

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <i className="fas fa-spinner fa-spin text-2xl text-primary-400" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <Card>
        <div className="space-y-2">
          <h2 className="text-xl font-semibold text-dark-100">权限与管理员</h2>
          <p className="text-sm text-dark-400">
            管理后台角色、管理员账号和当前插件声明的权限点；插件授权在此分配，插件启停仍在插件管理中处理。
          </p>
        </div>
      </Card>

      {/* 标签切换 */}
      <div className="flex gap-2 border-b border-dark-700/50 pb-4">
        <button
          onClick={() => setActiveTab('roles')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'roles'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-user-tag mr-2" />
          角色管理
        </button>
        <button
          onClick={() => setActiveTab('admins')}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            activeTab === 'admins'
              ? 'bg-primary-500/20 text-primary-400'
              : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
          }`}
        >
          <i className="fas fa-user-shield mr-2" />
          管理员列表
        </button>
      </div>

      {/* 角色管理 */}
      {activeTab === 'roles' && (
        <Card
          title="角色列表"
          icon={<i className="fas fa-user-tag" />}
          action={
            <PermissionGuard permission="role:create">
              <Button size="sm" onClick={() => openRoleModal()}>
                <i className="fas fa-plus mr-1" />
                添加角色
              </Button>
            </PermissionGuard>
          }
        >
          <div className="space-y-3">
            {roles.map((role) => (
              <div key={role.id} className="p-4 bg-dark-700/30 rounded-xl border border-dark-600/50">
                <div className="flex items-start justify-between">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-dark-100">{role.name}</span>
                      {role.is_system && <Badge variant="info">系统角色</Badge>}
                      <Badge variant={role.status === 1 ? 'success' : 'danger'}>
                        {role.status === 1 ? '启用' : '禁用'}
                      </Badge>
                    </div>
                    <div className="text-sm text-dark-400 mt-1">{role.description}</div>
                    <div className="text-sm text-dark-500 mt-2">
                      权限数: {role.permissions?.length || 0}
                    </div>
                  </div>
                  {!role.is_system && (
                    <div className="flex gap-2">
                      <PermissionGuard permission="role:edit">
                        <Button size="sm" variant="ghost" onClick={() => openRoleModal(role)}>
                          <i className="fas fa-edit" />
                        </Button>
                      </PermissionGuard>
                      <PermissionGuard permission="role:delete">
                        <Button size="sm" variant="ghost" onClick={() => handleDeleteRole(role.id)}>
                          <i className="fas fa-trash text-red-400" />
                        </Button>
                      </PermissionGuard>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* 管理员列表 */}
      {activeTab === 'admins' && (
        <Card
          title="管理员列表"
          icon={<i className="fas fa-user-shield" />}
          action={
            <PermissionGuard permission="admin:create">
              <Button size="sm" onClick={() => openAdminModal()}>
                <i className="fas fa-plus mr-1" />
                添加管理员
              </Button>
            </PermissionGuard>
          }
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-dark-400 text-sm border-b border-dark-700">
                  <th className="pb-3 font-medium">用户名</th>
                  <th className="pb-3 font-medium">昵称</th>
                  <th className="pb-3 font-medium">角色</th>
                  <th className="pb-3 font-medium">2FA</th>
                  <th className="pb-3 font-medium">状态</th>
                  <th className="pb-3 font-medium">最后登录</th>
                  <th className="pb-3 font-medium">操作</th>
                </tr>
              </thead>
              <tbody className="text-dark-200">
                {admins.map((admin) => (
                  <tr key={admin.id} className="border-b border-dark-700/50">
                    <td className="py-3">{admin.username}</td>
                    <td className="py-3">{admin.nickname || '-'}</td>
                    <td className="py-3">{admin.role_name}</td>
                    <td className="py-3">
                      <Badge variant={admin.enable_2fa ? 'success' : 'default'}>
                        {admin.enable_2fa ? '已开启' : '未开启'}
                      </Badge>
                    </td>
                    <td className="py-3">
                      <Badge variant={admin.status === 1 ? 'success' : 'danger'}>
                        {admin.status === 1 ? '正常' : '禁用'}
                      </Badge>
                    </td>
                    <td className="py-3 text-sm text-dark-400">
                      {admin.last_login_at ? formatDateTime(admin.last_login_at) : '-'}
                    </td>
                    <td className="py-3">
                      <div className="flex gap-2">
                        <PermissionGuard permission="admin:edit">
                          <Button size="sm" variant="ghost" onClick={() => openAdminModal(admin)}>
                            <i className="fas fa-edit" />
                          </Button>
                        </PermissionGuard>
                        <PermissionGuard permission="admin:delete">
                          <Button size="sm" variant="ghost" onClick={() => handleDeleteAdmin(admin.id)}>
                            <i className="fas fa-trash text-red-400" />
                          </Button>
                        </PermissionGuard>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* 角色编辑弹窗 */}
      <Modal
        isOpen={showRoleModal}
        onClose={() => setShowRoleModal(false)}
        title={editingRole ? '编辑角色' : '添加角色'}
        size="lg"
      >
        <div className="space-y-4">
          <Input
            label="角色名称"
            placeholder="请输入角色名称"
            value={roleForm.name}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setRoleForm({ ...roleForm, name: e.target.value })}
          />
          <Input
            label="角色描述"
            placeholder="请输入角色描述"
            value={roleForm.description}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setRoleForm({ ...roleForm, description: e.target.value })}
          />
          
          {/* 权限模板选择 */}
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">快速选择权限模板</label>
            <div className="flex flex-wrap gap-2">
              {templates.map((template) => (
                <button
                  key={template.name}
                  onClick={() => applyTemplate(template.name)}
                  className="px-3 py-1.5 text-sm rounded-lg bg-dark-700/50 hover:bg-dark-600/50 text-dark-300 hover:text-dark-100 transition-colors border border-dark-600"
                  title={template.description}
                >
                  {template.description.split(' - ')[0]}
                </button>
              ))}
              <button
                onClick={() => setRoleForm(prev => ({ ...prev, permissions: [] }))}
                className="px-3 py-1.5 text-sm rounded-lg bg-red-500/20 hover:bg-red-500/30 text-red-400 transition-colors border border-red-500/30"
              >
                清空权限
              </button>
            </div>
            <p className="text-xs text-dark-500 mt-1">
              已选择 {roleForm.permissions.length} 个权限
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">宿主权限</label>
            <div className="max-h-64 overflow-y-auto space-y-4 p-3 bg-dark-700/30 rounded-lg">
              {hostPermissionGroups.map(([group, perms]) => (
                <div key={group}>
                  <div className="text-sm font-medium text-dark-400 mb-2">{group}</div>
                  <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
                    {perms.map((perm) => (
                      <label key={perm.code} className="flex items-center gap-2 cursor-pointer">
                        <Toggle
                          size="sm"
                          checked={roleForm.permissions.includes(perm.code)}
                          onChange={() => togglePermission(perm.code)}
                        />
                        <span className="text-sm text-dark-300" title={perm.description}>{perm.name}</span>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">插件权限</label>
            <div className="max-h-64 overflow-y-auto space-y-4 p-3 bg-dark-700/30 rounded-lg">
              {pluginPermissionGroups.length === 0 ? (
                <div className="text-sm text-dark-500">当前没有插件声明权限</div>
              ) : (
                pluginPermissionGroups.map(([group, perms]) => (
                  <div key={group}>
                    <div className="text-sm font-medium text-dark-400 mb-2">{group}</div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                      {perms.map((perm) => (
                        <label key={perm.code} className="flex items-start gap-2 cursor-pointer rounded-lg bg-dark-800/50 p-2">
                          <Toggle
                            size="sm"
                            checked={roleForm.permissions.includes(perm.code)}
                            onChange={() => togglePermission(perm.code)}
                          />
                          <span>
                            <span className="block text-sm text-dark-200">{perm.name}</span>
                            <span className="block text-xs text-dark-500">{perm.code}</span>
                          </span>
                        </label>
                      ))}
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
          <Button className="w-full" onClick={handleSaveRole}>
            保存
          </Button>
        </div>
      </Modal>

      {/* 管理员编辑弹窗 */}
      <Modal
        isOpen={showAdminModal}
        onClose={() => setShowAdminModal(false)}
        title={editingAdmin ? '编辑管理员' : '添加管理员'}
        size="md"
      >
        <div className="space-y-4">
          <Input
            label="用户名"
            placeholder="请输入用户名"
            value={adminForm.username}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setAdminForm({ ...adminForm, username: e.target.value })}
            disabled={!!editingAdmin}
          />
          <Input
            label={editingAdmin ? '新密码（留空不修改）' : '密码'}
            type="password"
            placeholder={editingAdmin ? '留空不修改' : '请输入密码'}
            value={adminForm.password}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setAdminForm({ ...adminForm, password: e.target.value })}
          />
          <Input
            label="邮箱"
            type="email"
            placeholder="请输入邮箱"
            value={adminForm.email}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setAdminForm({ ...adminForm, email: e.target.value })}
          />
          <Input
            label="昵称"
            placeholder="请输入昵称"
            value={adminForm.nickname}
            onChange={(e: ChangeEvent<HTMLInputElement>) => setAdminForm({ ...adminForm, nickname: e.target.value })}
          />
          <div>
            <label className="block text-sm font-medium text-dark-300 mb-2">角色</label>
            <select
              className="w-full px-4 py-3 bg-dark-700/50 border border-dark-600 rounded-xl text-dark-100 focus:outline-none focus:border-primary-500"
              value={adminForm.role_id}
              onChange={(e) => setAdminForm({ ...adminForm, role_id: parseInt(e.target.value) })}
            >
              {roles.map((role) => (
                <option key={role.id} value={role.id}>{role.name}</option>
              ))}
            </select>
          </div>
          {editingAdmin && (
            <div>
              <label className="block text-sm font-medium text-dark-300 mb-2">状态</label>
              <select
                className="w-full px-4 py-3 bg-dark-700/50 border border-dark-600 rounded-xl text-dark-100 focus:outline-none focus:border-primary-500"
                value={adminForm.status}
                onChange={(e) => setAdminForm({ ...adminForm, status: parseInt(e.target.value) })}
              >
                <option value={1}>正常</option>
                <option value={0}>禁用</option>
              </select>
            </div>
          )}
          <Button className="w-full" onClick={handleSaveAdmin}>
            保存
          </Button>
        </div>
      </Modal>

      {/* 删除确认弹窗 */}
      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => { setShowDeleteConfirm(false); setDeleteTarget(null) }}
        title={deleteTarget?.type === 'role' ? '删除角色' : '删除管理员'}
        message={deleteTarget?.type === 'role'
          ? `确定要删除角色 "${deleteTarget?.name}" 吗？`
          : `确定要删除管理员 "${deleteTarget?.name}" 吗？`}
        confirmText="删除"
        variant="danger"
        onConfirm={handleDelete}
        loading={deleteLoading}
      />
    </div>
  )
}
