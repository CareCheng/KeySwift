export interface OrderDetail {
  id: number
  order_no: string
  user_id: number
  username: string
  product_id: number
  product_name: string
  quantity: number
  price: number
  original_price: number
  status: number
  is_test: boolean
  kami_code: string
  payment_method: string
  payment_no: string
  payment_time: string
  created_at: string
  updated_at: string
  duration: number
  duration_unit: string
}
