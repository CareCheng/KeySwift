/**
 * 主题 token 字典 —— 插件 iframe 前端侧的契约定义。
 * 与宿主 Program/web/src/lib/themeTokens.ts 保持字段一致。
 *
 * 插件前端消费规则：收到宿主 init/theme 消息中的 theme_tokens 后，
 * 调用 applyThemeTokens 写入自身 :root；缺字段用 DEFAULT_THEME_TOKENS 兜底。
 * 字段名固定，详见《插件开发手册》"主题 token 字典"章节。
 */

export const THEME_TOKEN_KEYS = [
  'bg-input',
  'border-color',
  'text-primary',
  'text-secondary',
  'text-muted',
  'text-placeholder',
  'primary-500',
  'primary-ring',
] as const

export type ThemeTokenKey = (typeof THEME_TOKEN_KEYS)[number]

export type ThemeTokenPack = Partial<Record<ThemeTokenKey, string>>

/** 内置默认值（深色），宿主未下发或字段缺失时兜底，避免白屏 */
export const DEFAULT_THEME_TOKENS: Record<ThemeTokenKey, string> = {
  'bg-input': 'rgba(30, 41, 59, 0.5)',
  'border-color': '#334155',
  'text-primary': '#f1f5f9',
  'text-secondary': '#cbd5e1',
  'text-muted': '#64748b',
  'text-placeholder': '#475569',
  'primary-500': '#667eea',
  'primary-ring': 'rgba(102, 126, 234, 0.2)',
}

/**
 * 将主题 token 包写入 document :root，缺字段用默认值补齐。
 * 在 SSR 或无 document 环境下安全跳过。
 */
export function applyThemeTokens(pack?: ThemeTokenPack): void {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  for (const key of THEME_TOKEN_KEYS) {
    const value = pack?.[key] || DEFAULT_THEME_TOKENS[key]
    root.style.setProperty(`--${key}`, value)
  }
}
