'use client'

import type { ReactNode } from 'react'
import { DynamicHomepage } from '@/components/homepage'
import { useUserRouteState } from '@/lib/userNavigation'
import { UserShell } from '@/components/layout/UserShell'
import { UserPageTransition } from '@/components/layout/UserPageTransition'
import { ProductsView } from './products/page'
import { ProductDetailView } from './product/page'
import { PaymentView } from './payment/page'
import { PaymentResultView } from './payment/result/page'
import { OrderDetailView } from './order/detail/page'
import { UserCenterView } from './user/page'

/**
 * 用户端主入口。
 * 常规页面切换使用 hash 状态，复用后台管理页“单页面内切换内容”的交互模式。
 */
export default function HomePage() {
  const route = useUserRouteState()
  const routeKey = `${route.view}:${route.params.toString()}`
  let content: ReactNode

  switch (route.view) {
    case 'products':
      content = <ProductsView />
      break
    case 'product':
      content = <ProductDetailView productId={route.params.get('id')} />
      break
    case 'payment':
      content = <PaymentView orderNo={route.params.get('order_no')} />
      break
    case 'payment-result':
      content = (
        <PaymentResultView
          orderNo={route.params.get('order_no')}
          status={route.params.get('status')}
        />
      )
      break
    case 'order-detail':
      content = <OrderDetailView orderNo={route.params.get('order_no')} />
      break
    case 'user':
      content = <UserCenterView />
      break
    default:
      content = <DynamicHomepage />
      break
  }

  return (
    <UserShell>
      <UserPageTransition pageKey={routeKey}>
        {content}
      </UserPageTransition>
    </UserShell>
  )
}
