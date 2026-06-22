import { apiGet } from '@/lib/api'
import type { OrderDetail } from '@/types/order'

export interface OrderDetailResult {
  order: OrderDetail | null
  error?: string
}

const orderDetailPromises = new Map<string, Promise<OrderDetailResult>>()
const cachedOrderDetails = new Map<string, OrderDetail>()

/**
 * 读取订单详情并在前端会话内复用结果，降低支付链路页面切换时的重复请求。
 */
export function getOrderDetail(orderNo: string) {
  if (!orderDetailPromises.has(orderNo)) {
    orderDetailPromises.set(
      orderNo,
      apiGet<{ order: OrderDetail }>(`/api/order/detail/${orderNo}`).then((res) => {
        if (res.success && res.order) {
          cachedOrderDetails.set(orderNo, res.order)
          return { order: res.order }
        }
        return { order: null, error: res.error }
      }),
    )
  }
  return orderDetailPromises.get(orderNo)!
}

export function getCachedOrderDetail(orderNo?: string | null) {
  return orderNo ? cachedOrderDetails.get(orderNo) || null : null
}

export function updateCachedOrderDetail(order: OrderDetail) {
  cachedOrderDetails.set(order.order_no, order)
  orderDetailPromises.set(order.order_no, Promise.resolve({ order }))
}
