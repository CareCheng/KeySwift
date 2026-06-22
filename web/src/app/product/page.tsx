'use client'

import { useState, useEffect, Suspense } from 'react'
import { useSearchParams, useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Button, Modal } from '@/components/ui'
import { UserShell } from '@/components/layout/UserShell'
import { UserRouteLink } from '@/components/layout/UserRouteLink'
import { apiPost } from '@/lib/api'
import {
  getCachedProductDetail,
  getCachedProductImages,
  getProductDetail,
  getProductImages,
} from '@/lib/productData'
import { formatMoney } from '@/lib/utils'
import { useUserNavigation } from '@/lib/userNavigation'
import type { ProductDetail, ProductImage } from '@/types/product'

/**
 * 商品接口
 */
/**
 * 规格参数项接口
 */
interface SpecItem {
  key: string
  value: string
}

interface ProductDetailViewProps {
  productId?: string | null
}

/**
 * 商品详情内容组件。
 * 支持独立路由和用户端主入口 hash 切换复用。
 */
export function ProductDetailView({ productId }: ProductDetailViewProps) {
  const router = useRouter()
  const navigateUser = useUserNavigation()
  const numericProductId = Number(productId)
  const validProductId = Number.isFinite(numericProductId) && numericProductId > 0

  const [product, setProduct] = useState<ProductDetail | null>(() => (
    validProductId ? getCachedProductDetail(numericProductId) : null
  ))
  const [images, setImages] = useState<ProductImage[]>(() => (
    validProductId ? getCachedProductImages(numericProductId) : []
  ))
  const [loading, setLoading] = useState(() => !validProductId || !getCachedProductDetail(numericProductId))
  const [currentImageIndex, setCurrentImageIndex] = useState(0)
  const [showPurchaseModal, setShowPurchaseModal] = useState(false)
  const [purchasing, setPurchasing] = useState(false)
  const [quantity, setQuantity] = useState(1)

  // 加载商品详情
  useEffect(() => {
    const loadProduct = async () => {
      if (!productId) {
        navigateUser('products')
        return
      }

      setLoading(true)
      try {
        if (!validProductId) {
          toast.error('商品不存在')
          navigateUser('products')
          return
        }

        // 先加载商品基本信息（必须成功）
        const productDetail = await getProductDetail(numericProductId)

        if (!productDetail) {
          toast.error('商品不存在')
          navigateUser('products')
          return
        }

        setProduct(productDetail)

        // 并行加载商品图片，图片失败不影响主信息展示。
        const productImages = await getProductImages(numericProductId)
        setImages(productImages)
      } catch (err) {
        console.error('加载商品信息失败:', err)
        toast.error('加载商品信息失败')
      }
      setLoading(false)
    }

    loadProduct()
  }, [numericProductId, validProductId, navigateUser])

  // 获取所有图片
  const allImages = product ? [
    { id: 0, image_url: product.image_url, is_primary: true },
    ...images.filter(img => img.image_url !== product.image_url)
  ].filter(img => img.image_url) : []

  // 确认购买
  const handlePurchase = async () => {
    if (!product) return
    setPurchasing(true)
    const res = await apiPost<{ order_no: string }>('/api/order/create', {
      product_id: product.id,
      quantity: quantity,
    })
    setPurchasing(false)

    if (res.success && res.order_no) {
      setShowPurchaseModal(false)
      toast.success('订单创建成功，正在跳转支付页面...')
      navigateUser('payment', { order_no: res.order_no })
    } else {
      if (res.error === '请先登录') {
        router.push('/login/')
      } else {
        toast.error(res.error || '创建订单失败')
      }
    }
  }

  if (loading) {
    return (
      <main className="flex-1 py-8 px-4">
        <div className="max-w-6xl mx-auto">
          <div className="animate-pulse">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
              <div className="h-96 bg-dark-700/50 rounded-xl" />
              <div className="space-y-4">
                <div className="h-8 bg-dark-700/50 rounded w-3/4" />
                <div className="h-4 bg-dark-700/50 rounded w-1/2" />
                <div className="h-20 bg-dark-700/50 rounded" />
              </div>
            </div>
          </div>
        </div>
      </main>
    )
  }

  if (!product) return null

  return (
    <>
      <main className="flex-1 py-8 px-4">
        <div className="max-w-6xl mx-auto">
          {/* 面包屑 */}
          <div className="mb-6 flex items-center gap-2 text-sm text-dark-400">
            <UserRouteLink view="products" className="hover:text-primary-400 transition-colors">商品列表</UserRouteLink>
            <i className="fas fa-chevron-right text-xs" />
            <span className="text-dark-200">{product.name}</span>
          </div>

          {/* 商品信息 */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-12">
            {/* 图片区域 */}
            <div className="space-y-4">
              <div className="aspect-square bg-dark-800/50 rounded-2xl overflow-hidden border border-dark-700/50">
                {allImages.length > 0 && allImages[currentImageIndex]?.image_url ? (
                  <img src={allImages[currentImageIndex].image_url} alt={product.name} className="w-full h-full object-cover" />
                ) : (
                  <div className="w-full h-full flex items-center justify-center"><span className="text-8xl">📦</span></div>
                )}
              </div>
              {allImages.length > 1 && (
                <div className="flex gap-2 overflow-x-auto pb-2">
                  {allImages.map((img, index) => (
                    <button key={img.id} onClick={() => setCurrentImageIndex(index)}
                      className={`flex-shrink-0 w-20 h-20 rounded-lg overflow-hidden border-2 transition-colors ${index === currentImageIndex ? 'border-primary-500' : 'border-dark-700/50 hover:border-dark-600'}`}>
                      <img src={img.image_url} alt="" className="w-full h-full object-cover" />
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* 商品详情 */}
            <div className="space-y-6">
              <div>
                <h1 className="text-3xl font-bold text-dark-100 mb-2">{product.name}</h1>
                {product.category_name && <span className="inline-block px-3 py-1 bg-primary-500/20 text-primary-400 text-sm rounded-full">{product.category_name}</span>}
              </div>

              <div className="flex items-baseline gap-2">
                <span className="text-4xl font-bold text-primary-400">{formatMoney(product.price)}</span>
                <span className="text-dark-500">/ {product.duration}{product.duration_unit}</span>
              </div>

              <div className="bg-dark-800/30 rounded-xl p-4">
                <h3 className="text-dark-300 text-sm mb-2">商品描述</h3>
                <p className="text-dark-200 whitespace-pre-wrap">{product.description || '暂无描述'}</p>
              </div>

              {/* 商品标签 */}
              {product.tags && (
                <div className="flex flex-wrap gap-2">
                  {product.tags.split(',').filter(t => t.trim()).map((tag, i) => (
                    <span key={i} className="px-3 py-1 bg-primary-500/20 text-primary-400 text-sm rounded-full">
                      {tag.trim()}
                    </span>
                  ))}
                </div>
              )}

              <div className="flex items-center gap-4">
                <span className="text-dark-400">库存状态：</span>
                {product.stock === -1 ? (
                  <span className="text-emerald-400"><i className="fas fa-check-circle mr-1" />库存充足</span>
                ) : product.stock > 0 ? (
                  <span className="text-amber-400"><i className="fas fa-exclamation-circle mr-1" />剩余 {product.stock} 件</span>
                ) : (
                  <span className="text-red-400"><i className="fas fa-times-circle mr-1" />已售罄</span>
                )}
              </div>

              {product.stock !== 0 && (
                <div className="flex items-center gap-4">
                  <span className="text-dark-400">购买数量：</span>
                  <div className="flex items-center gap-2">
                    <button onClick={() => setQuantity(Math.max(1, quantity - 1))} className="w-10 h-10 rounded-lg bg-dark-700/50 text-dark-300 hover:bg-dark-700 transition-colors flex items-center justify-center">
                      <i className="fas fa-minus" />
                    </button>
                    <input type="number" value={quantity} onChange={(e) => setQuantity(Math.max(1, parseInt(e.target.value) || 1))}
                      className="w-16 h-10 text-center bg-dark-800/50 border border-dark-700/50 rounded-lg text-dark-100" />
                    <button onClick={() => setQuantity(quantity + 1)} className="w-10 h-10 rounded-lg bg-dark-700/50 text-dark-300 hover:bg-dark-700 transition-colors flex items-center justify-center">
                      <i className="fas fa-plus" />
                    </button>
                  </div>
                </div>
              )}

              <div className="flex flex-col sm:flex-row gap-3 pt-4">
                <Button variant="primary" size="lg" className="w-full" onClick={() => setShowPurchaseModal(true)} disabled={product.stock === 0}>
                  <i className="fas fa-shopping-bag mr-2" />立即购买
                </Button>
              </div>
            </div>
          </div>

          {/* 商品详情（Markdown） */}
          {product.detail && (
            <div className="card p-6">
              <h2 className="text-xl font-bold text-dark-100 mb-6">
                <i className="fas fa-file-alt mr-2 text-primary-400" />商品详情
              </h2>
              <div className="prose prose-invert prose-sm max-w-none">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {product.detail}
                </ReactMarkdown>
              </div>
            </div>
          )}

          {/* 规格参数 */}
          {product.specs && (() => {
            try {
              const specs: SpecItem[] = JSON.parse(product.specs)
              if (specs.length > 0) {
                return (
                  <div className="card p-6">
                    <h2 className="text-xl font-bold text-dark-100 mb-6">
                      <i className="fas fa-list-ul mr-2 text-primary-400" />规格参数
                    </h2>
                    <div className="overflow-x-auto">
                      <table className="w-full">
                        <tbody>
                          {specs.map((spec, index) => (
                            <tr key={index} className={index % 2 === 0 ? 'bg-dark-700/30' : ''}>
                              <td className="py-3 px-4 text-dark-400 font-medium w-1/3">{spec.key}</td>
                              <td className="py-3 px-4 text-dark-200">{spec.value}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )
              }
            } catch { /* 解析失败忽略 */ }
            return null
          })()}

          {/* 特性/卖点 */}
          {product.features && (() => {
            try {
              const features: string[] = JSON.parse(product.features)
              if (features.length > 0) {
                return (
                  <div className="card p-6">
                    <h2 className="text-xl font-bold text-dark-100 mb-6">
                      <i className="fas fa-star mr-2 text-primary-400" />产品特性
                    </h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {features.map((feature, index) => (
                        <div key={index} className="flex items-start gap-3 p-3 bg-dark-700/30 rounded-lg">
                          <i className="fas fa-check-circle text-emerald-400 mt-0.5" />
                          <span className="text-dark-200">{feature}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )
              }
            } catch { /* 解析失败忽略 */ }
            return null
          })()}

        </div>
      </main>

      {/* 购买确认弹窗 */}
      <Modal isOpen={showPurchaseModal} onClose={() => setShowPurchaseModal(false)} title="确认购买" size="sm">
        <div className="space-y-4">
          <div className="bg-dark-700/30 rounded-xl p-4 space-y-2">
            <div className="flex justify-between"><span className="text-dark-400">商品名称</span><span className="text-dark-100">{product.name}</span></div>
            <div className="flex justify-between"><span className="text-dark-400">有效期</span><span className="text-dark-100">{product.duration}{product.duration_unit}</span></div>
            <div className="flex justify-between"><span className="text-dark-400">数量</span><span className="text-dark-100">{quantity}</span></div>
            <div className="flex justify-between border-t border-dark-600/50 pt-2 mt-2">
              <span className="text-dark-400">总价</span>
              <span className="text-primary-400 font-bold text-lg">{formatMoney(product.price * quantity)}</span>
            </div>
          </div>
          <div className="flex gap-3">
            <Button variant="secondary" className="flex-1" onClick={() => setShowPurchaseModal(false)}>取消</Button>
            <Button variant="primary" className="flex-1" onClick={handlePurchase} loading={purchasing}>确认购买</Button>
          </div>
        </div>
      </Modal>

    </>
  )
}

/**
 * 独立商品详情路由内容。
 */
function ProductDetailContent() {
  const searchParams = useSearchParams()
  return <ProductDetailView productId={searchParams.get('id')} />
}

/**
 * 商品详情页面
 */
export default function ProductDetailPage() {
  return (
    <UserShell>
      <Suspense fallback={
        <main className="flex-1 py-8 px-4">
          <div className="max-w-6xl mx-auto">
            <div className="animate-pulse">
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <div className="h-96 bg-dark-700/50 rounded-xl" />
                <div className="space-y-4">
                  <div className="h-8 bg-dark-700/50 rounded w-3/4" />
                  <div className="h-4 bg-dark-700/50 rounded w-1/2" />
                </div>
              </div>
            </div>
          </div>
        </main>
      }>
        <ProductDetailContent />
      </Suspense>
    </UserShell>
  )
}
