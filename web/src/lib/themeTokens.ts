/**
 * 主题 token 字典 —— 宿主与插件 iframe 前端之间的标准主题契约。
 *
 * 设计目标：让任意主题插件改配色时，挂载在 iframe 里的插件前端（如人机验证 widget）
 * 能无需预知主题名、无需各自维护配色，只消费这组固定字段即可跟随宿主主题。
 *
 * 契约规则（详见《插件开发手册》"主题 token 字典"章节）：
 * 1. 字段名固定，主题插件只能改值，不得增删字段名；新增字段属于协议变更，需同步所有插件前端。
 * 2. 宿主从当前生效的 CSS 变量读取这组值，经 postMessage 下发给插件 iframe。
 * 3. 插件前端把收到的值写入自身 :root，缺字段时用内置默认值兜底，避免白屏。
 */

/** 标准 token 字段名清单（协议字段，勿随意增删） */
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

/** 主题 token 包：字段名 → 值 */
export type ThemeTokenPack = Partial<Record<ThemeTokenKey, string>>

/** 插件前端内置默认值（深色），用于宿主未下发或字段缺失时兜底 */
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
 * 从 document 的 CSS 变量读取当前主题 token 包。
 * 在 SSR 或无 document 环境返回空对象，由调用方用默认值兜底。
 */
export function readThemeTokensFromDocument(): ThemeTokenPack {
  if (typeof window === 'undefined' || typeof document === 'undefined') return {}
  const root = document.documentElement
  const computed = window.getComputedStyle(root)
  const pack: ThemeTokenPack = {}
  for (const key of THEME_TOKEN_KEYS) {
    const value = computed.getPropertyValue(`--${key}`).trim()
    if (value) {
      pack[key] = value
    }
  }
  return pack
}
