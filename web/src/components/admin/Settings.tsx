'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { ThemeSettings } from './settings/ThemeSettings'
import { BasicSettings } from './settings/BasicSettings'
import { LoginSettings } from './settings/LoginSettings'
import { IpAccessSettings } from './settings/IpAccessSettings'
import { ReverseProxySettings } from './settings/ReverseProxySettings'
import { useSettingsState } from './settings/useSettingsState'

/**
 * 基础设置三级标签配置
 * 主题设置 / 基本设置 / 登录设置（含两步验证 TOTP）/ IP 访问管理（黑名单+白名单）。
 * 说明：原「安全设置」页的 TOTP 绑定能力已并入「登录设置」统一管理。
 */
const SUB_TABS = [
  { id: 'theme', label: '主题设置', icon: 'fa-palette' },
  { id: 'basic', label: '基本设置', icon: 'fa-cog' },
  { id: 'login', label: '登录设置', icon: 'fa-sign-in-alt' },
  { id: 'proxy', label: '访问与代理', icon: 'fa-globe' },
  { id: 'ip', label: 'IP 访问管理', icon: 'fa-network-wired' },
] as const

type SubTabId = typeof SUB_TABS[number]['id']

/**
 * 系统设置页面（基础设置的二级入口）
 * 负责三级标签切换，并统一持有基本/登录/安全设置共享的状态实例。
 */
export function SettingsPage() {
  const [activeSub, setActiveSub] = useState<SubTabId>('theme')
  // 基本/登录/安全三个子页面共享同一份配置状态，避免后端全量覆盖式保存时互相清零
  const settingsState = useSettingsState()

  const renderContent = () => {
    switch (activeSub) {
      case 'theme':
        return <ThemeSettings />
      case 'basic':
        return <BasicSettings state={settingsState} />
      case 'login':
        return <LoginSettings state={settingsState} />
      case 'proxy':
        return <ReverseProxySettings />
      case 'ip':
        return <IpAccessSettings />
      default:
        return <ThemeSettings />
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium" style={{ color: 'var(--text-primary)' }}>系统设置</h2>

      {/* 三级标签切换 */}
      <div className="flex flex-wrap gap-2 border-b border-dark-700/50 pb-4">
        {SUB_TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveSub(tab.id)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeSub === tab.id
                ? 'bg-primary-500/20 text-primary-400'
                : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
            }`}
          >
            <i className={`fas ${tab.icon} mr-2`} />
            {tab.label}
          </button>
        ))}
      </div>

      <AnimatePresence mode="wait">
        <motion.div
          key={activeSub}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          transition={{ duration: 0.15 }}
        >
          {renderContent()}
        </motion.div>
      </AnimatePresence>
    </div>
  )
}
