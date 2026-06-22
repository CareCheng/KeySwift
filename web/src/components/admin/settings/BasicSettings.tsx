'use client'

import { Button, Card, Input } from '@/components/ui'
import toast from 'react-hot-toast'
import { SettingsState } from './useSettingsState'

/**
 * 基本设置子页面
 * 职责：系统标题、管理后台入口后缀、服务器端口。
 * 数据状态由父容器统一管理，本组件只负责渲染与触发保存。
 */
export function BasicSettings({ state }: { state: SettingsState }) {
  const { basicForm, setBasicForm, saveBasic } = state

  const handleSave = async () => {
    const suffix = basicForm.admin_suffix.trim()
    if (suffix && !/^[a-zA-Z0-9_-]+$/.test(suffix)) {
      toast.error('管理后台入口后缀只能包含字母、数字、下划线和横线')
      return
    }
    const port = parseInt(basicForm.server_port) || 8080
    if (port < 1 || port > 65535) {
      toast.error('端口号必须在 1-65535 之间')
      return
    }
    if (await saveBasic()) {
      toast.success('设置已保存')
    } else {
      toast.error('保存失败')
    }
  }

  return (
    <Card title="基本设置">
      <div className="space-y-4">
        <Input label="系统标题" value={basicForm.system_title} onChange={(e) => setBasicForm({ ...basicForm, system_title: e.target.value })} />
        <div className="grid grid-cols-2 gap-4">
          <div>
            <Input label="管理后台入口后缀" value={basicForm.admin_suffix} onChange={(e) => setBasicForm({ ...basicForm, admin_suffix: e.target.value })} placeholder="manage" />
            <p className="text-dark-500 text-xs mt-1">管理后台访问地址: {typeof window !== 'undefined' ? window.location.origin : ''}/{basicForm.admin_suffix || 'manage'}</p>
          </div>
          <div>
            <Input label="服务器端口" type="number" value={basicForm.server_port} onChange={(e) => setBasicForm({ ...basicForm, server_port: e.target.value })} />
            <p className="text-dark-500 text-xs mt-1">修改端口后需要重启程序生效</p>
          </div>
        </div>
        <div className="flex justify-end">
          <Button onClick={handleSave}>保存设置</Button>
        </div>
      </div>
    </Card>
  )
}
