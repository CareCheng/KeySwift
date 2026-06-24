/**
 * 宿主 ↔ 插件 iframe 标准消息协议类型定义。
 * 与宿主 HumanVerificationWidget.tsx 保持一致，插件前端只收发、不硬编码宿主渲染逻辑。
 */

import type { ThemeTokenPack } from './themeTokens'

export type HumanVerificationScope = 'admin_login' | 'user_login' | 'user_register'

export interface WidgetOutbound {
  source: 'keyswift-human-verification-plugin'
  type: 'ready' | 'change' | 'error' | 'resize'
  payload?: HumanVerificationPayload | null
  error?: string
  height?: number
}

export interface WidgetInbound {
  source: 'keyswift-human-verification-host'
  type: 'init' | 'theme'
  scope?: HumanVerificationScope
  provider_id?: string
  provider_type?: string
  public_config?: Record<string, unknown>
  challenge_endpoint?: string
  reset_signal?: number
  theme_tokens?: ThemeTokenPack
}

export interface HumanVerificationPayload {
  provider_id: string
  provider_type: string
  scope: HumanVerificationScope
  token?: string
  challenge_id?: string
  answer?: string
  client_nonce?: string
  metadata?: Record<string, unknown>
}

/** 向父窗口推送消息，限定同源。source 由本函数自动补齐，调用方无需提供。 */
export function postToHost(message: Omit<WidgetOutbound, 'source'>): void {
  if (typeof window === 'undefined') return
  parent.postMessage({ source: 'keyswift-human-verification-plugin', ...message }, window.location.origin)
}
