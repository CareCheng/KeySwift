'use client'

import Link from 'next/link'
import type { HomepageConfig } from '@/types/homepage'

interface CTASectionProps {
  config: HomepageConfig
}

/**
 * CTA 区块组件
 */
export function CTASection({ config }: CTASectionProps) {
  if (!config.cta_enabled) return null

  return (
    <section className="py-20 px-4">
      <div className="max-w-4xl mx-auto">
        <div
          className="relative rounded-3xl p-12 text-center overflow-hidden"
          style={{
            background: `linear-gradient(135deg, ${config.primary_color}, ${config.secondary_color})`,
          }}
        >
          {/* 装饰元素 */}
          <div className="absolute top-0 right-0 w-64 h-64 rounded-full opacity-20 -translate-y-1/2 translate-x-1/2"
            style={{ backgroundColor: 'white' }}
          />
          <div className="absolute bottom-0 left-0 w-48 h-48 rounded-full opacity-20 translate-y-1/2 -translate-x-1/2"
            style={{ backgroundColor: 'white' }}
          />

          <div className="relative z-10">
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
              {config.cta_title}
            </h2>
            {config.cta_subtitle && (
              <p className="text-lg text-white/80 mb-8">
                {config.cta_subtitle}
              </p>
            )}
            <Link
              href={config.cta_button_link}
              className="inline-flex items-center gap-2 px-8 py-4 bg-white rounded-xl font-semibold transition-all hover:shadow-lg hover:-translate-y-1"
              style={{ color: config.primary_color }}
            >
              <i className="fas fa-rocket" />
              {config.cta_button_text}
            </Link>
          </div>
        </div>
      </div>
    </section>
  )
}
