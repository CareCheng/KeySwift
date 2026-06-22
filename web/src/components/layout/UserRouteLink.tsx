'use client'

import type { AnchorHTMLAttributes, MouseEvent, ReactNode } from 'react'
import {
  buildUserRouteHref,
  type UserRouteParams,
  type UserRouteView,
  useUserNavigation,
} from '@/lib/userNavigation'

interface UserRouteLinkProps
  extends Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'href' | 'onClick'> {
  view: UserRouteView
  params?: UserRouteParams
  children: ReactNode
  onClick?: (event: MouseEvent<HTMLAnchorElement>) => void
}

/**
 * 用户端内部路由链接。
 * 左键点击时复用用户端 hash 状态切换，避免触发 Next 页面级导航；同时保留链接地址用于复制和新标签打开。
 */
export function UserRouteLink({
  view,
  params,
  children,
  onClick,
  target,
  ...props
}: UserRouteLinkProps) {
  const navigateUser = useUserNavigation()
  const href = buildUserRouteHref(view, params)

  const handleClick = (event: MouseEvent<HTMLAnchorElement>) => {
    onClick?.(event)
    if (
      event.defaultPrevented ||
      target === '_blank' ||
      event.button !== 0 ||
      event.metaKey ||
      event.ctrlKey ||
      event.shiftKey ||
      event.altKey
    ) {
      return
    }

    event.preventDefault()
    navigateUser(view, params)
  }

  return (
    <a href={href} target={target} onClick={handleClick} {...props}>
      {children}
    </a>
  )
}
