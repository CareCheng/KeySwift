'use client'

import { UserAuthFlow } from '@/components/auth/UserAuthFlow'

/**
 * 用户注册页面入口。
 */
export default function RegisterPage() {
  return <UserAuthFlow initialMode="register" />
}

