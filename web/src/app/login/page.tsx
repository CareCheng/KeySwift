'use client'

import { UserAuthFlow } from '@/components/auth/UserAuthFlow'

/**
 * 用户登录页面入口。
 */
export default function LoginPage() {
  return <UserAuthFlow initialMode="login" />
}

