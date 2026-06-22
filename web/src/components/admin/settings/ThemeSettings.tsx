'use client'

import toast from 'react-hot-toast'
import { Card } from '@/components/ui'
import { useTheme, Theme } from '@/lib/theme'

/**
 * 主题设置子页面
 * 职责：切换深色/浅色主题，设置自动保存到 localStorage 并应用到所有页面。
 */
export function ThemeSettings() {
  const { theme, setTheme } = useTheme()

  const handleThemeChange = (newTheme: Theme) => {
    setTheme(newTheme)
    toast.success(`已切换到${newTheme === 'dark' ? '深色' : '浅色'}主题`)
  }

  return (
    <div className="space-y-4">
      <Card title="主题设置">
        <div className="space-y-4">
          <p className="text-sm" style={{ color: 'var(--text-muted)' }}>选择您喜欢的界面主题风格，设置将自动保存并应用于所有页面</p>
          <div className="grid grid-cols-2 gap-4">
            <button
              onClick={() => handleThemeChange('dark')}
              className={`p-4 rounded-xl border-2 transition-all duration-200 ${
                theme === 'dark'
                  ? 'border-primary-500 bg-primary-500/10'
                  : 'border-dark-600 hover:border-dark-500'
              }`}
            >
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-lg bg-dark-800 border border-dark-600 flex items-center justify-center">
                  <i className="fas fa-moon text-primary-400" />
                </div>
                <div className="text-left">
                  <div className="font-medium" style={{ color: 'var(--text-primary)' }}>深色主题</div>
                  <div className="text-xs" style={{ color: 'var(--text-muted)' }}>护眼暗色风格</div>
                </div>
              </div>
              <div className="flex gap-1">
                <div className="w-6 h-6 rounded bg-slate-900 border border-dark-600" />
                <div className="w-6 h-6 rounded bg-slate-800 border border-dark-600" />
                <div className="w-6 h-6 rounded bg-purple-900 border border-dark-600" />
                <div className="w-6 h-6 rounded bg-primary-500 border border-dark-600" />
              </div>
            </button>
            <button
              onClick={() => handleThemeChange('light')}
              className={`p-4 rounded-xl border-2 transition-all duration-200 ${
                theme === 'light'
                  ? 'border-primary-500 bg-primary-500/10'
                  : 'border-dark-600 hover:border-dark-500'
              }`}
            >
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-lg bg-white border border-gray-200 flex items-center justify-center">
                  <i className="fas fa-sun text-amber-500" />
                </div>
                <div className="text-left">
                  <div className="font-medium" style={{ color: 'var(--text-primary)' }}>浅色主题</div>
                  <div className="text-xs" style={{ color: 'var(--text-muted)' }}>明亮清爽风格</div>
                </div>
              </div>
              <div className="flex gap-1">
                <div className="w-6 h-6 rounded bg-gray-50 border border-gray-200" />
                <div className="w-6 h-6 rounded bg-white border border-gray-200" />
                <div className="w-6 h-6 rounded bg-indigo-50 border border-gray-200" />
                <div className="w-6 h-6 rounded bg-primary-500 border border-gray-200" />
              </div>
            </button>
          </div>
        </div>
      </Card>
    </div>
  )
}
