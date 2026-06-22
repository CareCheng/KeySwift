import { apiGet } from '@/lib/api'
import type { UserInfo } from '@/types/user'

let userInfoPromise: Promise<UserInfo | null> | null = null
let cachedUserInfo: UserInfo | null = null

/**
 * 读取用户信息并在前端会话内复用，避免导航栏和用户中心重复请求同一接口。
 */
export function getUserInfo() {
  if (!userInfoPromise) {
    userInfoPromise = apiGet<{ user: UserInfo }>('/api/user/info').then((res) => {
      cachedUserInfo = res.success && res.user ? res.user : null
      return cachedUserInfo
    })
  }
  return userInfoPromise
}

export function getCachedUserInfo() {
  return cachedUserInfo
}

export function updateCachedUserInfo(user: UserInfo | null) {
  cachedUserInfo = user
  userInfoPromise = Promise.resolve(user)
}
