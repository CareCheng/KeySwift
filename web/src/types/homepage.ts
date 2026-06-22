/**
 * 首页核心展示配置类型
 */

export interface FeatureItem {
  icon: string
  title: string
  description: string
}

export interface StatItem {
  value: string
  label: string
  icon: string
}

export interface HomepageConfig {
  primary_color: string
  secondary_color: string

  hero_enabled: boolean
  hero_title: string
  hero_subtitle: string
  hero_button_text: string
  hero_button_link: string
  hero_background: 'gradient' | 'image' | 'solid'
  hero_bg_image: string
  hero_bg_color: string

  features_enabled: boolean
  features_title: string
  features: FeatureItem[]

  products_enabled: boolean
  products_title: string
  products_count: number

  stats_enabled: boolean
  stats: StatItem[]

  cta_enabled: boolean
  cta_title: string
  cta_subtitle: string
  cta_button_text: string
  cta_button_link: string
}

export const defaultHomepageConfig: HomepageConfig = {
  primary_color: '#6366f1',
  secondary_color: '#8b5cf6',

  hero_enabled: true,
  hero_title: '欢迎使用卡密购买系统',
  hero_subtitle: '安全、便捷的卡密购买平台',
  hero_button_text: '浏览商品',
  hero_button_link: '#/products',
  hero_background: 'gradient',
  hero_bg_image: '',
  hero_bg_color: '',

  features_enabled: true,
  features_title: '核心能力',
  features: [
    { icon: 'fa-shield-halved', title: '安全下单', description: '账户、订单和支付流程保持清晰可控' },
    { icon: 'fa-wallet', title: '余额支付', description: '主程序默认保留余额支付能力' },
    { icon: 'fa-key', title: '自动发卡', description: '支付成功后自动交付可用卡密' },
  ],

  products_enabled: true,
  products_title: '商品列表',
  products_count: 6,

  stats_enabled: true,
  stats: [
    { value: '核心', label: '卡密销售', icon: 'fa-box' },
    { value: '默认', label: '余额支付', icon: 'fa-wallet' },
    { value: '自动', label: '订单发卡', icon: 'fa-key' },
    { value: '清晰', label: '用户中心', icon: 'fa-user' },
  ],

  cta_enabled: true,
  cta_title: '开始购买卡密',
  cta_subtitle: '注册账号后即可创建订单并使用余额完成支付',
  cta_button_text: '立即注册',
  cta_button_link: '/register/',
}
