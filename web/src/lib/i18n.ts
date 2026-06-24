/**
 * 国际化配置
 */

export const locales = ['zh', 'en'] as const
export type Locale = typeof locales[number]

export const defaultLocale: Locale = 'zh'

export const localeNames: Record<Locale, string> = {
  zh: '中文',
  en: 'English',
}

const zhTranslations = {
  common: {
    cancel: '取消',
    clickRefresh: '点击刷新',
    close: '关闭',
    confirm: '确认',
    copied: '已复制',
    fillComplete: '请填写完整信息',
    loading: '加载中...',
    next: '下一步',
    prev: '上一步',
    seconds: '秒',
    sending: '发送中...',
    verify: '验证',
  },
  auth: {
    backToHome: '返回首页',
    backToLogin: '返回登录',
    forgotPassword: '忘记密码？',
    forgotStep1: '请输入您的用户名',
    forgotStep2: '请验证您的身份',
    forgotStep3: '请设置您的新密码',
    forgotTitle: '找回密码',
    hasAccount: '已有账号？立即登录',
    infoExpired: '验证信息已过期',
    invalidRequest: '无效的验证请求',
    login: '登录',
    loginFailed: '登录失败',
    loginSubtitle: '欢迎回来，请登录您的账号',
    loginSuccess: '登录成功',
    loginTitle: '用户登录',
    noAccount: '没有账号？立即注册',
    register: '注册',
    registerFailed: '注册失败',
    registerNewAccount: '注册新账号',
    registerSubtitle: '创建您的账号，开始购买',
    registerSuccess: '注册成功，已自动登录',
    registerTitle: '用户注册',
    resetFailed: '重置失败',
    resetSuccess: '密码重置成功，即将跳转登录',
    userNotFound: '用户不存在或未绑定邮箱',
    verifyFailed: '验证码错误',
    verifySubtitle: '为保护您的账号安全，请完成二次验证',
    verifySuccess: '验证成功',
    verifyTitle: '二次验证',
  },
  footer: {
    allRightsReserved: '保留所有权利',
  },
  nav: {
    home: '首页',
    login: '登录',
    logout: '退出登录',
    products: '商品列表',
    register: '注册',
  },
  order: {
    copyKami: '复制卡密',
    kamiCode: '卡密',
    orderNo: '订单号',
  },
  product: {
    buyNow: '立即购买',
    clearSearch: '清除搜索',
    confirmPurchase: '确认购买',
    congratulations: '恭喜您，购买成功！',
    duration: '有效期',
    foundProducts: '找到 {count} 个商品',
    name: '商品名称',
    noDescription: '暂无描述',
    noMatchingProducts: '没有找到匹配的商品',
    noProducts: '暂无商品',
    orderCreated: '订单创建成功，正在跳转支付页面...',
    orderCreateFailed: '创建订单失败',
    outOfStock: '已售罄',
    price: '价格',
    productList: '商品列表',
    purchaseSuccess: '购买成功',
    searchPlaceholder: '搜索商品名称或描述...',
    sortDefault: '默认排序',
    sortPriceAsc: '价格从低到高',
    sortPriceDesc: '价格从高到低',
    stock: '库存',
    stockSufficient: '库存充足',
    tryOtherKeywords: '试试其他关键词？',
  },
  user: {
    captcha: '人机验证',
    captchaPlaceholder: '请输入验证码',
    codeCorrect: '验证码正确',
    codeIncorrect: '验证码错误或已过期',
    codeSendFailed: '发送失败',
    codeSendTo: '验证码将发送到',
    codeSent: '验证码已发送，请查收邮件',
    confirmNewPassword: '确认新密码',
    confirmNewPasswordPlaceholder: '再次输入新密码',
    confirmPassword: '确认密码',
    confirmPasswordPlaceholder: '再次输入密码',
    email: '邮箱',
    emailCode: '邮箱验证码',
    emailCodePlaceholder: '请输入{length}位验证码',
    emailFirst: '请先输入邮箱地址',
    emailInvalid: '请输入有效的邮箱地址',
    emailPlaceholder: '用于验证和找回密码',
    emailVerify: '邮箱验证码验证',
    enter6DigitCode: '请输入6位验证码',
    enter6DigitTotp: '请输入6位动态口令',
    newPassword: '新密码',
    newPasswordPlaceholder: '至少6位',
    openAuthApp: '请打开验证器 APP 查看当前验证码',
    password: '密码',
    passwordMinLength: '至少6位',
    passwordMismatch: '两次密码不一致',
    passwordPlaceholder: '请输入密码',
    passwordTooShort: '密码长度至少6位',
    phoneOptional: '手机号（可选）',
    phonePlaceholder: '选填',
    rememberMe: '记住我',
    resetPassword: '重置密码',
    sendCode: '发送验证码',
    totpCode: '动态口令',
    totpHint: '请输入您验证器 APP 中的动态口令',
    totpPlaceholder: '000000',
    totpVerify: '动态口令验证',
    username: '用户名',
    usernamePlaceholder: '3-20个字符',
    verifyLogin: '验证登录',
  },
}

const enTranslations: typeof zhTranslations = {
  common: {
    cancel: 'Cancel',
    clickRefresh: 'Click to refresh',
    close: 'Close',
    confirm: 'Confirm',
    copied: 'Copied',
    fillComplete: 'Please fill in all required fields',
    loading: 'Loading...',
    next: 'Next',
    prev: 'Previous',
    seconds: 's',
    sending: 'Sending...',
    verify: 'Verify',
  },
  auth: {
    backToHome: 'Back to home',
    backToLogin: 'Back to login',
    forgotPassword: 'Forgot password?',
    forgotStep1: 'Please enter your username',
    forgotStep2: 'Please verify your identity',
    forgotStep3: 'Please set your new password',
    forgotTitle: 'Forgot Password',
    hasAccount: 'Already have an account? Login now',
    infoExpired: 'Verification info has expired',
    invalidRequest: 'Invalid verification request',
    login: 'Login',
    loginFailed: 'Login failed',
    loginSubtitle: 'Welcome back, please login to your account',
    loginSuccess: 'Login successful',
    loginTitle: 'User Login',
    noAccount: "Don't have an account? Register now",
    register: 'Register',
    registerFailed: 'Registration failed',
    registerNewAccount: 'Register new account',
    registerSubtitle: 'Create your account to start shopping',
    registerSuccess: 'Registration successful, signed in automatically',
    registerTitle: 'User Registration',
    resetFailed: 'Reset failed',
    resetSuccess: 'Password reset successful, redirecting to login',
    userNotFound: 'User not found or email not bound',
    verifyFailed: 'Invalid verification code',
    verifySubtitle: 'Please complete verification to protect your account',
    verifySuccess: 'Verification successful',
    verifyTitle: 'Two-Factor Authentication',
  },
  footer: {
    allRightsReserved: 'All rights reserved',
  },
  nav: {
    home: 'Home',
    login: 'Login',
    logout: 'Logout',
    products: 'Products',
    register: 'Register',
  },
  order: {
    copyKami: 'Copy key',
    kamiCode: 'License key',
    orderNo: 'Order No.',
  },
  product: {
    buyNow: 'Buy now',
    clearSearch: 'Clear search',
    confirmPurchase: 'Confirm purchase',
    congratulations: 'Congratulations! Purchase successful!',
    duration: 'Duration',
    foundProducts: 'Found {count} products',
    name: 'Product name',
    noDescription: 'No description',
    noMatchingProducts: 'No matching products found',
    noProducts: 'No products',
    orderCreated: 'Order created, redirecting to payment...',
    orderCreateFailed: 'Failed to create order',
    outOfStock: 'Out of stock',
    price: 'Price',
    productList: 'Products',
    purchaseSuccess: 'Purchase successful',
    searchPlaceholder: 'Search products...',
    sortDefault: 'Default',
    sortPriceAsc: 'Price: low to high',
    sortPriceDesc: 'Price: high to low',
    stock: 'Stock',
    stockSufficient: 'In stock',
    tryOtherKeywords: 'Try other keywords?',
  },
  user: {
    captcha: 'Captcha',
    captchaPlaceholder: 'Enter captcha',
    codeCorrect: 'Code is correct',
    codeIncorrect: 'Code is incorrect or expired',
    codeSendFailed: 'Failed to send',
    codeSendTo: 'Code will be sent to',
    codeSent: 'Code sent, please check your email',
    confirmNewPassword: 'Confirm new password',
    confirmNewPasswordPlaceholder: 'Enter new password again',
    confirmPassword: 'Confirm password',
    confirmPasswordPlaceholder: 'Enter password again',
    email: 'Email',
    emailCode: 'Email verification code',
    emailCodePlaceholder: 'Enter {length}-digit code',
    emailFirst: 'Please enter email address first',
    emailInvalid: 'Please enter a valid email address',
    emailPlaceholder: 'For verification and password recovery',
    emailVerify: 'Email verification',
    enter6DigitCode: 'Please enter 6-digit code',
    enter6DigitTotp: 'Please enter 6-digit TOTP code',
    newPassword: 'New password',
    newPasswordPlaceholder: 'At least 6 characters',
    openAuthApp: 'Open your authenticator app to view the code',
    password: 'Password',
    passwordMinLength: 'At least 6 characters',
    passwordMismatch: 'Passwords do not match',
    passwordPlaceholder: 'Enter your password',
    passwordTooShort: 'Password must be at least 6 characters',
    phoneOptional: 'Phone (optional)',
    phonePlaceholder: 'Optional',
    rememberMe: 'Remember me',
    resetPassword: 'Reset password',
    sendCode: 'Send code',
    totpCode: 'TOTP code',
    totpHint: 'Enter the code from your authenticator app',
    totpPlaceholder: '000000',
    totpVerify: 'TOTP verification',
    username: 'Username',
    usernamePlaceholder: '3-20 characters',
    verifyLogin: 'Verify and login',
  },
}

const translations: Record<Locale, typeof zhTranslations> = {
  zh: zhTranslations,
  en: enTranslations,
}

/**
 * 获取当前语言
 */
export function getLocale(): Locale {
  if (typeof window === 'undefined') return defaultLocale

  const stored = localStorage.getItem('locale')
  if (stored && locales.includes(stored as Locale)) {
    return stored as Locale
  }

  const browserLang = navigator.language.split('-')[0]
  if (locales.includes(browserLang as Locale)) {
    return browserLang as Locale
  }

  return defaultLocale
}

/**
 * 设置当前语言
 */
export function setLocale(locale: Locale): void {
  if (typeof window === 'undefined') return
  localStorage.setItem('locale', locale)
  window.location.reload()
}

/**
 * 获取翻译文本
 */
export function t(key: string, locale?: Locale): string {
  const currentLocale = locale || getLocale()
  const dict = translations[currentLocale]

  const keys = key.split('.')
  let value: unknown = dict

  for (const k of keys) {
    if (value && typeof value === 'object' && k in value) {
      value = (value as Record<string, unknown>)[k]
    } else {
      return key
    }
  }

  return typeof value === 'string' ? value : key
}

/**
 * 获取翻译字典
 */
export function getTranslations(locale?: Locale): typeof zhTranslations {
  return translations[locale || getLocale()]
}

const i18n = { t, getLocale, setLocale, getTranslations, locales, localeNames }

export default i18n
