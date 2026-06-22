'use client'

import { useCallback, useEffect, useMemo, useState } from 'react'
import { usePathname, useRouter } from 'next/navigation'

export type UserRouteView =
  | 'home'
  | 'products'
  | 'product'
  | 'payment'
  | 'payment-result'
  | 'order-detail'
  | 'user'

export interface UserRouteState {
  view: UserRouteView
  params: URLSearchParams
}

export type UserRouteParams = Record<string, string | number | boolean | null | undefined>

const VIEW_TO_HASH_PATH: Record<UserRouteView, string> = {
  home: '/home',
  products: '/products',
  product: '/product',
  payment: '/payment',
  'payment-result': '/payment/result',
  'order-detail': '/order/detail',
  user: '/user',
}

const HASH_PATH_TO_VIEW = new Map(
  Object.entries(VIEW_TO_HASH_PATH).map(([view, path]) => [path, view as UserRouteView]),
)

const USER_PATH_TO_VIEW = new Map<string, UserRouteView>([
  ['/', 'home'],
  ['/products', 'products'],
  ['/product', 'product'],
  ['/payment', 'payment'],
  ['/payment/result', 'payment-result'],
  ['/order/detail', 'order-detail'],
  ['/user', 'user'],
])

function normalizePath(pathname: string | null | undefined) {
  if (!pathname) return '/'
  return pathname.replace(/\/+$/, '') || '/'
}

function buildQuery(params?: Record<string, string | number | boolean | null | undefined>) {
  const query = new URLSearchParams()
  Object.entries(params || {}).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      query.set(key, String(value))
    }
  })
  const queryString = query.toString()
  return queryString ? `?${queryString}` : ''
}

export function buildUserRouteHash(
  view: UserRouteView,
  params?: UserRouteParams,
) {
  return `#${VIEW_TO_HASH_PATH[view]}${buildQuery(params)}`
}

export function buildUserRouteHref(
  view: UserRouteView,
  params?: UserRouteParams,
) {
  return `/${buildUserRouteHash(view, params)}`
}

export function buildUserRouteUrl(
  view: UserRouteView,
  params?: UserRouteParams,
) {
  return buildUserRouteHref(view, params)
}

export function parseUserRouteHash(hash: string | null | undefined): UserRouteState {
  const rawHash = hash?.startsWith('#') ? hash.slice(1) : hash || ''
  const [rawPath, rawQuery = ''] = rawHash.split('?')
  const path = normalizePath(rawPath || '/home')

  return {
    view: HASH_PATH_TO_VIEW.get(path) || 'home',
    params: new URLSearchParams(rawQuery),
  }
}

export function getUserRouteViewFromPath(pathname: string | null | undefined) {
  return USER_PATH_TO_VIEW.get(normalizePath(pathname)) || null
}

export function useUserRouteState() {
  const pathname = usePathname()
  const [route, setRoute] = useState<UserRouteState>(() => {
    if (typeof window === 'undefined') {
      return { view: 'home', params: new URLSearchParams() }
    }
    return parseUserRouteHash(window.location.hash)
  })

  useEffect(() => {
    const syncRoute = () => {
      setRoute(parseUserRouteHash(window.location.hash))
    }

    syncRoute()
    window.addEventListener('hashchange', syncRoute)
    return () => window.removeEventListener('hashchange', syncRoute)
  }, [pathname])

  return route
}

export function useCurrentUserView() {
  const pathname = usePathname()
  const route = useUserRouteState()

  return useMemo(() => {
    return getUserRouteViewFromPath(pathname) || route.view
  }, [pathname, route.view])
}

export function useUserNavigation() {
  const router = useRouter()
  const pathname = usePathname()

  return useCallback(
    (view: UserRouteView, params?: UserRouteParams) => {
      const hash = buildUserRouteHash(view, params)
      if (normalizePath(pathname) === '/') {
        window.location.hash = hash
      } else {
        router.push(`/${hash}`)
      }
    },
    [pathname, router],
  )
}
