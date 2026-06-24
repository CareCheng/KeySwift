/**
 * API 请求封装
 * 统一处理请求、响应、错误和 CSRF Token
 */

// API 配置
const apiConfig = {
  timeout: 30000,
  csrfToken: null as string | null,
}

/**
 * 获取 Cookie 值
 */
function getCookie(name: string): string | null {
  if (typeof document === 'undefined') return null
  const value = `; ${document.cookie}`
  const parts = value.split(`; ${name}=`)
  if (parts.length === 2) return parts.pop()?.split(';').shift() || null
  return null
}

/**
 * 获取 CSRF Token
 */
async function getCSRFToken(): Promise<string | null> {
  // 优先从 Cookie 获取
  const cookieToken = getCookie('csrf_token')
  if (cookieToken) {
    apiConfig.csrfToken = cookieToken
    return cookieToken
  }

  // 从服务器获取
  try {
    const res = await fetch('/api/csrf-token')
    const data = await res.json()
    if (data.success && data.token) {
      apiConfig.csrfToken = data.token
      return data.token
    }
  } catch (err) {
    console.warn('获取 CSRF Token 失败:', err)
  }
  return null
}

/**
 * 通用 API 请求
 */
export async function apiRequest<T = Record<string, unknown>>(
  url: string,
  options: {
    method?: string
    body?: Record<string, unknown>
    headers?: Record<string, string>
  } = {}
): Promise<T & { success: boolean; error?: string }> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...options.headers,
  }

  const mergedOptions: RequestInit = {
    method: options.method || 'GET',
    headers,
  }

  // 对于非 GET 请求，添加 CSRF Token
  if (options.method && options.method !== 'GET') {
    let csrfToken = apiConfig.csrfToken || getCookie('csrf_token')
    if (!csrfToken) {
      csrfToken = await getCSRFToken()
    }
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken
    }
    // 处理请求体
    if (options.body) {
      mergedOptions.body = JSON.stringify(options.body)
    }
  }

  mergedOptions.headers = headers

  // 创建 AbortController 用于超时控制
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), apiConfig.timeout)
  mergedOptions.signal = controller.signal

  try {
    const res = await fetch(url, mergedOptions)
    clearTimeout(timeoutId)

    const data = await res.json()

    // 检查 CSRF 错误，自动刷新令牌重试
    if (res.status === 403 && data.error?.includes('CSRF')) {
      apiConfig.csrfToken = null
      await getCSRFToken()
      return apiRequest(url, options)
    }

    return data
  } catch (err) {
    clearTimeout(timeoutId)

    if (err instanceof Error && err.name === 'AbortError') {
      return { success: false, error: '请求超时' } as T & { success: boolean; error: string }
    }

    return { success: false, error: '网络错误' } as T & { success: boolean; error: string }
  }
}

/**
 * GET 请求
 */
export function apiGet<T = Record<string, unknown>>(url: string) {
  return apiRequest<T>(url, { method: 'GET' })
}

/**
 * POST 请求
 */
export function apiPost<T = Record<string, unknown>>(
  url: string,
  body?: Record<string, unknown>
) {
  return apiRequest<T>(url, { method: 'POST', body })
}

/**
 * 表单上传请求
 */
export async function apiUpload<T = Record<string, unknown>>(
  url: string,
  formData: FormData
): Promise<T & { success: boolean; error?: string }> {
  const headers: Record<string, string> = {}
  let csrfToken = apiConfig.csrfToken || getCookie('csrf_token')
  if (!csrfToken) {
    csrfToken = await getCSRFToken()
  }
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken
  }

  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), apiConfig.timeout)

  try {
    const res = await fetch(url, {
      method: 'POST',
      headers,
      body: formData,
      signal: controller.signal,
    })
    clearTimeout(timeoutId)
    const data = await res.json()
    if (res.status === 403 && data.error?.includes('CSRF')) {
      apiConfig.csrfToken = null
      await getCSRFToken()
      return apiUpload(url, formData)
    }
    return data
  } catch (err) {
    clearTimeout(timeoutId)
    if (err instanceof Error && err.name === 'AbortError') {
      return { success: false, error: '请求超时' } as T & { success: boolean; error: string }
    }
    return { success: false, error: '网络错误' } as T & { success: boolean; error: string }
  }
}

/**
 * PUT 请求
 */
export function apiPut<T = Record<string, unknown>>(
  url: string,
  body?: Record<string, unknown>
) {
  return apiRequest<T>(url, { method: 'PUT', body })
}

/**
 * DELETE 请求
 */
export function apiDelete<T = Record<string, unknown>>(
  url: string,
  body?: Record<string, unknown>
) {
  return apiRequest<T>(url, { method: 'DELETE', body })
}

export type ApiResponse<T> = T & { success: boolean; error?: string }
