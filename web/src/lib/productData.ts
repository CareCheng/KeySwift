import { apiGet } from '@/lib/api'
import type { ProductDetail, ProductImage, ProductSummary } from '@/types/product'

let productsPromise: Promise<ProductSummary[]> | null = null
const productDetailPromises = new Map<number, Promise<ProductDetail | null>>()
const productImagesPromises = new Map<number, Promise<ProductImage[]>>()
let cachedProducts: ProductSummary[] | null = null
const cachedProductDetails = new Map<number, ProductDetail>()
const cachedProductImages = new Map<number, ProductImage[]>()

/**
 * 读取商品列表并在前端会话内复用结果，减少用户端单页切换时的重复请求和骨架屏闪烁。
 */
export function getProducts() {
  if (!productsPromise) {
    productsPromise = apiGet<{ products: ProductSummary[] }>('/api/products').then((res) => (
      res.success && res.products ? res.products : []
    ))
    productsPromise.then((products) => {
      cachedProducts = products
    })
  }
  return productsPromise
}

export function getCachedProducts() {
  return cachedProducts
}

/**
 * 读取商品详情并缓存同一商品的请求。
 */
export function getProductDetail(productId: number) {
  if (!productDetailPromises.has(productId)) {
    productDetailPromises.set(
      productId,
      apiGet<{ product: ProductDetail }>(`/api/product/${productId}`).then((res) => (
        res.success && res.product ? res.product : null
      )).then((product) => {
        if (product) {
          cachedProductDetails.set(productId, product)
        }
        return product
      }),
    )
  }
  return productDetailPromises.get(productId)!
}

export function getCachedProductDetail(productId: number) {
  return cachedProductDetails.get(productId) || null
}

/**
 * 读取商品图片并缓存同一商品的请求；图片失败不影响商品详情展示。
 */
export function getProductImages(productId: number) {
  if (!productImagesPromises.has(productId)) {
    productImagesPromises.set(
      productId,
      apiGet<{ data: ProductImage[] }>(`/api/product/${productId}/images`)
        .then((res) => (res.success && Array.isArray(res.data) ? res.data : []))
        .catch(() => []),
    )
    productImagesPromises.get(productId)!.then((images) => {
      cachedProductImages.set(productId, images)
    })
  }
  return productImagesPromises.get(productId)!
}

export function getCachedProductImages(productId: number) {
  return cachedProductImages.get(productId) || []
}

export function prefetchProducts() {
  void getProducts()
}

export function prefetchProductDetail(productId: number) {
  void getProductDetail(productId)
  void getProductImages(productId)
}
