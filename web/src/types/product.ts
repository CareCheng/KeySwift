export interface ProductSummary {
  id: number
  name: string
  description: string
  price: number
  duration: number
  duration_unit: string
  stock?: number
  image_url: string
}

export interface ProductDetail extends ProductSummary {
  stock: number
  detail: string
  specs: string
  features: string
  tags: string
  category_id: number
  category_name?: string
}

export interface ProductImage {
  id: number
  product_id: number
  image_url: string
  sort_order: number
  is_primary: boolean
}
