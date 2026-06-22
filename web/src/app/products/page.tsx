'use client'

import { useState, useEffect, useMemo } from 'react'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { Button, Modal } from '@/components/ui'
import { UserShell } from '@/components/layout/UserShell'
import { UserRouteLink } from '@/components/layout/UserRouteLink'
import { apiPost } from '@/lib/api'
import { getCachedProducts, getProducts, prefetchProductDetail } from '@/lib/productData'
import { formatMoney } from '@/lib/utils'
import { useUserNavigation } from '@/lib/userNavigation'
import { useI18n } from '@/hooks/useI18n'
import type { ProductSummary } from '@/types/product'

/**
 * 商品列表视图。
 * 由用户端主入口和旧路径直达共同复用。
 */
export function ProductsView() {
  const { t } = useI18n()
  const router = useRouter()
  const navigateUser = useUserNavigation()
  const [products, setProducts] = useState<ProductSummary[]>(() => getCachedProducts() || [])
  const [loading, setLoading] = useState(() => !getCachedProducts())
  const [selectedProduct, setSelectedProduct] = useState<ProductSummary | null>(null)
  const [showPurchaseModal, setShowPurchaseModal] = useState(false)
  const [purchasing, setPurchasing] = useState(false)
  
  // 搜索相关状态
  const [searchQuery, setSearchQuery] = useState('')
  const [sortBy, setSortBy] = useState<'default' | 'price_asc' | 'price_desc'>('default')
  


  // 加载商品列表
  useEffect(() => {
    const loadProducts = async () => {
      const list = await getProducts()
      setProducts(list)
      setLoading(false)
    }
    loadProducts()
  }, [])

  // 过滤和排序商品
  const filteredProducts = useMemo(() => {
    let result = [...products]
    
    // 搜索过滤
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase()
      result = result.filter(p => 
        p.name.toLowerCase().includes(query) || 
        (p.description && p.description.toLowerCase().includes(query))
      )
    }
    
    // 排序
    if (sortBy === 'price_asc') {
      result.sort((a, b) => a.price - b.price)
    } else if (sortBy === 'price_desc') {
      result.sort((a, b) => b.price - a.price)
    }
    
    return result
  }, [products, searchQuery, sortBy])

  // 选择商品
  const handleSelectProduct = (product: ProductSummary) => {
    setSelectedProduct(product)
    setShowPurchaseModal(true)
  }

  // 确认购买
  const handlePurchase = async () => {
    if (!selectedProduct) return

    setPurchasing(true)
    const res = await apiPost<{ order_no: string }>('/api/order/create', {
      product_id: selectedProduct.id,
    })
    setPurchasing(false)

    if (res.success && res.order_no) {
      setShowPurchaseModal(false)
      toast.success(t('product.orderCreated'))
      navigateUser('payment', { order_no: res.order_no })
    } else {
      if (res.error === '请先登录') {
        router.push('/login/')
      } else {
        toast.error(res.error || t('product.orderCreateFailed'))
      }
    }
  }

  return (
    <>
      <main className="flex-1 py-8 px-4">
        <div className="max-w-6xl mx-auto">
          <h1
            className="text-3xl font-bold text-dark-100 mb-8 text-center"
          >
            {t('product.productList')}
          </h1>

          {/* 搜索和筛选栏 */}
          <div
            className="mb-6 flex flex-col sm:flex-row gap-4"
          >
            <div className="flex-1 relative">
              <i className="fas fa-search absolute left-4 top-1/2 -translate-y-1/2 text-dark-500" />
              <input
                type="text"
                placeholder={t('product.searchPlaceholder')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-11 pr-4 py-3 bg-dark-800/50 border border-dark-700/50 rounded-xl text-dark-100 placeholder-dark-500 focus:outline-none focus:border-primary-500/50 transition-colors"
              />
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery('')}
                  className="absolute right-4 top-1/2 -translate-y-1/2 text-dark-500 hover:text-dark-300"
                >
                  <i className="fas fa-times" />
                </button>
              )}
            </div>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as typeof sortBy)}
              className="px-4 py-3 bg-dark-800/50 border border-dark-700/50 rounded-xl text-dark-100 focus:outline-none focus:border-primary-500/50 transition-colors"
            >
              <option value="default">{t('product.sortDefault')}</option>
              <option value="price_asc">{t('product.sortPriceAsc')}</option>
              <option value="price_desc">{t('product.sortPriceDesc')}</option>
            </select>
          </div>

          {/* 搜索结果提示 */}
          {searchQuery && (
            <div
              className="mb-4 text-dark-400 text-sm"
            >
              {t('product.foundProducts').replace('{count}', String(filteredProducts.length))}
              {filteredProducts.length === 0 && (
                <span className="ml-2">- {t('product.tryOtherKeywords')}</span>
              )}
            </div>
          )}

          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {[1, 2, 3].map((i) => (
                <div key={i} className="card p-6 animate-pulse">
                  <div className="h-40 bg-dark-700/50 rounded-xl mb-4" />
                  <div className="h-6 bg-dark-700/50 rounded w-3/4 mb-2" />
                  <div className="h-4 bg-dark-700/50 rounded w-1/2" />
                </div>
              ))}
            </div>
          ) : filteredProducts.length === 0 ? (
            <div className="text-center py-20">
              <div className="text-6xl mb-4">{searchQuery ? '🔍' : '📦'}</div>
              <p className="text-dark-400">
                {searchQuery ? t('product.noMatchingProducts') : t('product.noProducts')}
              </p>
              {searchQuery && (
                <Button variant="secondary" className="mt-4" onClick={() => setSearchQuery('')}>
                  {t('product.clearSearch')}
                </Button>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {filteredProducts.map((product) => (
                <div
                  key={product.id}
                  className="product-card cursor-pointer"
                  onMouseEnter={() => prefetchProductDetail(product.id)}
                >
                  <UserRouteLink view="product" params={{ id: product.id }} className="block">
                    <div className="h-40 bg-gradient-to-br from-primary-500/20 to-purple-500/20 flex items-center justify-center">
                      {product.image_url ? (
                        <img src={product.image_url} alt={product.name} className="w-full h-full object-cover" />
                      ) : (
                        <span className="text-6xl">📦</span>
                      )}
                    </div>
                    <div className="p-5">
                      <h3 className="text-lg font-semibold text-dark-100 mb-2">{product.name}</h3>
                      <p className="text-dark-400 text-sm mb-4 line-clamp-2">
                        {product.description || t('product.noDescription')}
                      </p>
                      <div className="flex items-center justify-between">
                        <span className="text-dark-500 text-sm">{product.duration}{product.duration_unit}</span>
                        <span className="text-xl font-bold text-primary-400">{formatMoney(product.price)}</span>
                      </div>
                      <div className="mt-3 text-sm">
                        {product.stock === undefined || product.stock === -1 ? (
                          <span className="text-emerald-400">{t('product.stockSufficient')}</span>
                        ) : product.stock > 0 ? (
                          <span className="text-amber-400">{t('product.stock')}: {product.stock}</span>
                        ) : (
                          <span className="text-red-400">{t('product.outOfStock')}</span>
                        )}
                      </div>
                    </div>
                  </UserRouteLink>
                  <div className="px-5 pb-5">
                    <Button
                      variant="primary"
                      size="sm"
                      className="w-full"
                      onClick={(e) => { e.preventDefault(); e.stopPropagation(); handleSelectProduct(product) }}
                      disabled={product.stock === 0}
                    >
                      {t('product.buyNow')}
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </main>

      {/* 购买确认弹窗 */}
      <Modal isOpen={showPurchaseModal} onClose={() => setShowPurchaseModal(false)} title={t('product.confirmPurchase')} size="sm">
        {selectedProduct && (
          <div className="space-y-4">
            <div className="bg-dark-700/30 rounded-xl p-4 space-y-2">
              <div className="flex justify-between">
                <span className="text-dark-400">{t('product.name')}</span>
                <span className="text-dark-100">{selectedProduct.name}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-dark-400">{t('product.duration')}</span>
                <span className="text-dark-100">{selectedProduct.duration}{selectedProduct.duration_unit}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-dark-400">{t('product.price')}</span>
                <span className="text-primary-400 font-bold">{formatMoney(selectedProduct.price)}</span>
              </div>
            </div>
            <div className="flex flex-col sm:flex-row gap-3">
              <Button variant="secondary" className="flex-1" onClick={() => setShowPurchaseModal(false)}>{t('common.cancel')}</Button>
              <Button variant="primary" className="flex-1" onClick={handlePurchase} loading={purchasing}>{t('common.confirm')}</Button>
            </div>
          </div>
        )}
      </Modal>

    </>
  )
}

/**
 * 商品列表旧路径直达入口。
 */
export default function ProductsPage() {
  return (
    <UserShell>
      <ProductsView />
    </UserShell>
  )
}
