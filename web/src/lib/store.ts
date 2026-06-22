import { create } from 'zustand'
import type { TwoFAStatus, UserInfo } from '@/types/user'

/**
 * 应用状态存储
 */
interface AppState {
  // 用户信息
  user: UserInfo | null
  setUser: (user: UserInfo | null) => void

  // 2FA 状态
  twoFAStatus: TwoFAStatus | null
  setTwoFAStatus: (status: TwoFAStatus | null) => void

  // 登录状态
  isLoggedIn: boolean
  setIsLoggedIn: (value: boolean) => void

  // 加载状态
  isLoading: boolean
  setIsLoading: (value: boolean) => void
}

export const useAppStore = create<AppState>((set) => ({
  // 用户信息
  user: null,
  setUser: (user) => set({ user, isLoggedIn: !!user }),

  // 2FA 状态
  twoFAStatus: null,
  setTwoFAStatus: (twoFAStatus) => set({ twoFAStatus }),

  // 登录状态
  isLoggedIn: false,
  setIsLoggedIn: (isLoggedIn) => set({ isLoggedIn }),

  // 加载状态
  isLoading: false,
  setIsLoading: (isLoading) => set({ isLoading }),
}))
