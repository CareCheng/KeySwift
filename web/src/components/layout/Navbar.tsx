'use client'

import { useState, useEffect, useRef } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { motion } from 'framer-motion'
import { cn } from '@/lib/utils'
import { apiPost } from '@/lib/api'
import { useAppStore } from '@/lib/store'
import { useTheme } from '@/lib/theme'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { useI18n } from '@/hooks/useI18n'
import { captureAuthTrigger, takeAuthSuccessTarget } from '@/lib/authTransition'
import { getCachedUserInfo, getUserInfo, updateCachedUserInfo } from '@/lib/userData'
import { useCurrentUserView } from '@/lib/userNavigation'
import { UserRouteLink } from './UserRouteLink'

/**
 * 导航栏组件
 */
export function Navbar() {
  const router = useRouter()
  const currentView = useCurrentUserView()
  const { user, setUser, isLoggedIn, setIsLoggedIn } = useAppStore()
  const { theme, toggleTheme } = useTheme()
  const { t, locale } = useI18n()
  const [loading, setLoading] = useState(() => !getCachedUserInfo() && !user)
  const [successPulse, setSuccessPulse] = useState(false)
  const loginLinkRef = useRef<HTMLAnchorElement>(null)
  const registerLinkRef = useRef<HTMLAnchorElement>(null)

  // 加载用户信息
  useEffect(() => {
    const loadUser = async () => {
      const userInfo = await getUserInfo()
      if (userInfo) {
        setUser(userInfo)
        setIsLoggedIn(true)
      } else {
        setUser(null)
        setIsLoggedIn(false)
      }
      setLoading(false)
    }
    loadUser()
  }, [setUser, setIsLoggedIn])

  useEffect(() => {
    if (!isLoggedIn || !user) return

    const successTarget = takeAuthSuccessTarget()
    if (!successTarget) return

    setSuccessPulse(true)
    const timer = window.setTimeout(() => setSuccessPulse(false), 900)
    return () => window.clearTimeout(timer)
  }, [isLoggedIn, user])

  // 退出登录
  const handleLogout = async () => {
    await apiPost('/api/user/logout')
    updateCachedUserInfo(null)
    setUser(null)
    setIsLoggedIn(false)
    router.push('/#/')
  }

  const navLinks = [
    { view: 'home' as const, label: t('nav.home'), icon: 'fa-home' },
    { view: 'products' as const, label: t('nav.products'), icon: 'fa-box' },
  ]

  return (
    <nav className="navbar">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <UserRouteLink view="home" className="flex items-center gap-2">
            <span className="text-2xl">🔐</span>
            <span className="text-lg font-bold" style={{ color: 'var(--text-primary)' }}>
              {locale === 'en' ? 'License Store' : '卡密购买系统'}
            </span>
          </UserRouteLink>

          {/* 导航链接 */}
          <div className="flex items-center gap-4 sm:gap-6">
            {navLinks.map((link) => (
              <UserRouteLink
                key={link.view}
                view={link.view}
                className={cn(
                  'text-sm font-medium transition-colors duration-200 flex items-center gap-1.5',
                  currentView === link.view
                    ? 'text-primary-400'
                    : 'text-dark-400 hover:text-dark-200'
                )}
              >
                <i className={`fas ${link.icon} text-base`} />
                <span className="hidden sm:inline">{link.label}</span>
              </UserRouteLink>
            ))}

            {/* 语言切换 */}
            <LanguageSwitcher />

            {/* 主题切换按钮 */}
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg transition-all duration-200 hover:bg-primary-500/10 group"
              title={theme === 'dark' ? (locale === 'en' ? 'Switch to Light Theme' : '切换到浅色主题') : (locale === 'en' ? 'Switch to Dark Theme' : '切换到深色主题')}
            >
              {theme === 'dark' ? (
                <i className="fas fa-sun text-lg text-amber-400 group-hover:text-amber-300 transition-colors" />
              ) : (
                <i className="fas fa-moon text-lg text-indigo-400 group-hover:text-indigo-300 transition-colors" />
              )}
            </button>

            {/* 用户菜单 */}
            {loading ? (
              <div className="w-20 h-8 rounded-lg animate-pulse" style={{ background: 'var(--bg-tertiary)' }} />
            ) : isLoggedIn && user ? (
              <motion.div
                initial={successPulse ? { opacity: 0, scale: 0.72, filter: 'blur(6px)' } : false}
                animate={{ opacity: 1, scale: 1, filter: 'blur(0px)' }}
                transition={{ type: 'spring', stiffness: 420, damping: 30, mass: 0.7 }}
                className="flex items-center gap-4"
              >
                <UserRouteLink
                  view="user"
                  className="flex items-center gap-2 text-sm hover:text-primary-400 transition-colors"
                  style={{ color: 'var(--text-secondary)' }}
                >
                  <i className="fas fa-user-circle text-lg" />
                  <span className="hidden sm:inline">{user.username}</span>
                </UserRouteLink>
                <button
                  onClick={handleLogout}
                  className="flex items-center gap-1.5 text-sm hover:text-red-400 transition-colors"
                  style={{ color: 'var(--text-muted)' }}
                >
                  <i className="fas fa-right-from-bracket" />
                  <span className="hidden sm:inline">{t('nav.logout')}</span>
                </button>
              </motion.div>
            ) : (
              <div className="flex items-center gap-3">
                <Link
                  ref={loginLinkRef}
                  href="/login/"
                  onClick={() => captureAuthTrigger(loginLinkRef.current, 'login')}
                  className="flex items-center gap-1.5 text-sm transition-colors hover:text-primary-400"
                  style={{ color: 'var(--text-muted)' }}
                >
                  <i className="fas fa-right-to-bracket" />
                  <span className="hidden sm:inline">{t('nav.login')}</span>
                </Link>
                <Link
                  ref={registerLinkRef}
                  href="/register/"
                  onClick={() => captureAuthTrigger(registerLinkRef.current, 'register')}
                  className="btn btn-primary btn-sm"
                >
                  <i className="fas fa-user-plus mr-1" />
                  <span className="hidden sm:inline">{t('nav.register')}</span>
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>
    </nav>
  )
}

/**
 * 页脚组件
 */
export function Footer() {
  const { t, locale } = useI18n()
  
  return (
    <footer className="py-8 mt-auto" style={{ borderTop: '1px solid var(--border-light)' }}>
      <div className="max-w-7xl mx-auto px-4 text-center text-sm" style={{ color: 'var(--text-muted)' }}>
        <p>
          &copy; {new Date().getFullYear()} {locale === 'en' ? 'License Store' : '卡密购买系统'}. {t('footer.allRightsReserved')}.
        </p>
      </div>
    </footer>
  )
}
