'use client'

import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { SettingsPage } from './Settings'
import { PaymentPage } from './Payment'
import { EmailPage } from './Email'
import { DatabasePage } from './Database'

/**
 * 系统配置页面标签配置
 */
const TABS = [
  { id: 'settings', label: '基础设置', icon: 'fa-cog' },
  { id: 'payment', label: '支付配置', icon: 'fa-credit-card' },
  { id: 'email', label: '邮箱配置', icon: 'fa-envelope' },
  { id: 'database', label: '数据库配置', icon: 'fa-database' },
]

/**
 * 系统配置组合页面
 * 合并：基础设置、支付配置、邮箱配置、数据库配置
 */
export function ConfigManagePage() {
  const [activeTab, setActiveTab] = useState('settings')

  // 渲染当前标签内容
  const renderContent = () => {
    switch (activeTab) {
      case 'settings':
        return <SettingsPage />
      case 'payment':
        return <PaymentPage />
      case 'email':
        return <EmailPage />
      case 'database':
        return <DatabasePage />
      default:
        return <SettingsPage />
    }
  }

  return (
    <div className="space-y-6">
      {/* 顶部标签切换 */}
      <div className="flex flex-wrap gap-2 border-b border-dark-700/50 pb-4">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              activeTab === tab.id
                ? 'bg-primary-500/20 text-primary-400'
                : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/50'
            }`}
          >
            <i className={`fas ${tab.icon} mr-2`} />
            {tab.label}
          </button>
        ))}
      </div>

      {/* 内容区域 */}
      <AnimatePresence mode="wait">
        <motion.div
          key={activeTab}
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
