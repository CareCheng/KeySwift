'use client'

import { useState, useEffect } from 'react'
import { UserRouteLink } from '@/components/layout/UserRouteLink'
import { getProducts, prefetchProductDetail } from '@/lib/productData'
import { formatMoney } from '@/lib/utils'
import type { HomepageConfig } from '@/types/homepage'
import type { ProductSummary } from '@/types/product'

interface ProductsSectionProps {
  config: HomepageConfig
}

/**
 * 商品展示区块组件
 */
export function ProductsSection({ config }: ProductsSectionProps) {
  const [products, setProducts] = useState<ProductSummary[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!config.products_enabled) return

    const loadProducts = async () => {
      const list = await getProducts()
      setProducts(list.slice(0, config.products_count || 6))
      setLoading(false)
    }
    loadProducts()
  }, [config.products_enabled, config.products_count])

  if (!config.products_enabled) return null

  return (
    <section className="py-16 px-4" style={{ backgroundColor: 'var(--bg-secondary)' }}>
      <div className="max-w-6xl mx-auto">
        {config.products_title && (
          <h2
            className="text-3xl font-bold text-center mb-12"
            style={{ color: 'var(--text-primary)' }}
          >
            {config.products_title}
          </h2>
        )}

        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3].map((i) => (
              <div key={i} className="card p-6 animate-pulse">
                <div className="h-32 rounded-xl mb-4" style={{ backgroundColor: 'var(--bg-tertiary)' }} />
                <div className="h-5 rounded w-3/4 mb-2" style={{ backgroundColor: 'var(--bg-tertiary)' }} />
                <div className="h-4 rounded w-1/2" style={{ backgroundColor: 'var(--bg-tertiary)' }} />
              </div>
            ))}
          </div>
        ) : products.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-5xl mb-4">📦</div>
            <p style={{ color: 'var(--text-muted)' }}>暂无商品</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {products.map((product) => (
              <div
                key={product.id}
                onMouseEnter={() => prefetchProductDetail(product.id)}
              >
                <UserRouteLink view="product" params={{ id: product.id }} className="block">
                  <div className="card overflow-hidden hover:shadow-lg transition-all hover:-translate-y-1">
                    <div
                      className="h-32 flex items-center justify-center"
                      style={{
                        background: `linear-gradient(135deg, ${config.primary_color}20, ${config.secondary_color}20)`,
                      }}
                    >
                      {product.image_url ? (
                        <img
                          src={product.image_url}
                          alt={product.name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <span className="text-5xl">📦</span>
                      )}
                    </div>
                    <div className="p-4">
                      <h3
                        className="font-semibold mb-1 truncate"
                        style={{ color: 'var(--text-primary)' }}
                      >
                        {product.name}
                      </h3>
                      <p
                        className="text-sm mb-3 line-clamp-2"
                        style={{ color: 'var(--text-muted)' }}
                      >
                        {product.description || '暂无描述'}
                      </p>
                      <div className="flex items-center justify-between">
                        <span className="text-sm" style={{ color: 'var(--text-muted)' }}>
                          {product.duration}{product.duration_unit}
                        </span>
                        <span
                          className="text-lg font-bold"
                          style={{ color: config.primary_color }}
                        >
                          {formatMoney(product.price)}
                        </span>
                      </div>
                    </div>
                  </div>
                </UserRouteLink>
              </div>
            ))}
          </div>
        )}

        <div
          className="text-center mt-8"
        >
          <UserRouteLink
            view="products"
            className="inline-flex items-center gap-2 px-6 py-3 rounded-xl transition-all hover:gap-3"
            style={{
              color: config.primary_color,
              backgroundColor: `${config.primary_color}15`,
            }}
          >
            查看全部商品
            <i className="fas fa-arrow-right" />
          </UserRouteLink>
        </div>
      </div>
    </section>
  )
}
