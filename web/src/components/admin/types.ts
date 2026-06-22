/**
 * 管理后台类型定义
 */

// 商品
export interface Product {
  id: number
  name: string
  description: string
  detail: string           // 详细介绍（Markdown/HTML）
  specs: string            // 规格参数（JSON格式）
  features: string         // 特性/卖点列表（JSON格式）
  tags: string             // 商品标签（逗号分隔）
  category_name: string
  price: number
  stock: number
  duration: number
  duration_unit: string
  status: number
  image_url: string
  product_type: number  // 1: 手动卡密
  created_at: string
}

// 手动卡密
export interface ManualKami {
  id: number
  product_id: number
  kami_code: string
  status: number  // 0: 可用, 1: 已售出, 2: 已禁用
  order_id: number
  order_no: string
  sold_at: string
  created_at: string
}

// 卡密统计
export interface KamiStats {
  total: number
  available: number
  sold: number
  disabled: number
}

// 分类
export interface Category {
  id: number
  name: string
  icon: string
  sort_order: number
  status: number
}

// 订单
export interface Order {
  id: number
  order_no: string
  username: string
  product_name: string
  quantity: number
  price: number
  status: number
  created_at: string
  paid_at: string
  card_info: string
}

// 用户
export interface User {
  id: number
  username: string
  email: string
  phone: string
  email_verified: boolean
  enable_2fa: boolean
  pay_password_set: boolean
  status: number
  last_login_at: string
  last_login_ip: string
  order_count: number
  paid_order_count: number
  available_balance: number
  created_at: string
}

// 日志（文件存储版本，使用AES-256-GCM加密）
export interface Log {
  id: number
  user_type: string    // user, admin, security
  user_id: number
  username: string
  action: string
  target: string
  target_id: string
  detail: string
  ip: string
  user_agent: string
  created_at: string
}

// 支付配置
export interface PaymentConfig {
  balance?: { enabled: boolean; builtin: boolean; name: string }
}

// 邮箱配置
export interface EmailConfig {
  enabled: boolean
  smtp_host: string
  smtp_port: number
  smtp_user: string
  has_password: boolean
  from_name: string
  from_email: string
  encryption: string  // 加密方式：none/ssl/starttls
  code_length: number
}

// 数据库配置
export interface DBConfig {
  connected: boolean
  type: string
  host: string
  port: number
  user: string
  database: string
  key_length: number
  encryption_key: string
}

// 系统设置
export interface Settings {
  system_title: string
  admin_suffix: string
  server_port: number
  enable_login: boolean
  enable_captcha: boolean
  admin_username: string
  enable_2fa: boolean
  totp_secret: string
  enable_session_timeout: boolean
  session_timeout: number
  user_allow_register: boolean
  user_enable_captcha: boolean
  user_enable_2fa: boolean
  user_require_email_verification: boolean
  user_enable_session_timeout: boolean
  user_session_timeout: number
}

export { HOST_ADMIN_PAGES as PAGE_CONFIG } from '@/lib/pluginRegistry'
