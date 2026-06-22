'use client'

import { UserRouteLink } from '@/components/layout/UserRouteLink'
import type { HomepageConfig } from '@/types/homepage'

interface HeroSectionProps {
  config: HomepageConfig
}

/**
 * Hero 区块组件
 */
export function HeroSection({ config }: HeroSectionProps) {
  if (!config.hero_enabled) return null

  // 根据背景类型生成样式
  const getBackgroundStyle = () => {
    switch (config.hero_background) {
      case 'image':
        return {
          backgroundImage: `url(${config.hero_bg_image})`,
          backgroundSize: 'cover',
          backgroundPosition: 'center',
        }
      case 'solid':
        return { backgroundColor: config.hero_bg_color || config.primary_color }
      default:
        return {}
    }
  }

  return (
    <section className="relative py-20 px-4 overflow-hidden" style={getBackgroundStyle()}>
      {/* 渐变背景装饰 */}
      {config.hero_background === 'gradient' && (
        <div className="absolute inset-0 overflow-hidden">
          <div
            className="absolute -top-40 -right-40 w-80 h-80 rounded-full blur-3xl"
            style={{ backgroundColor: `${config.primary_color}33` }}
          />
          <div
            className="absolute -bottom-40 -left-40 w-80 h-80 rounded-full blur-3xl"
            style={{ backgroundColor: `${config.secondary_color}33` }}
          />
        </div>
      )}

      {/* 图片背景遮罩 */}
      {config.hero_background === 'image' && (
        <div className="absolute inset-0 bg-black/50" />
      )}

      <div className="relative max-w-4xl mx-auto text-center">
        <h1
          className="text-4xl md:text-5xl font-bold mb-6"
          style={{ color: 'var(--text-primary)' }}
        >
          {config.hero_title.includes('卡密') ? (
            <>
              {config.hero_title.split('卡密')[0]}
              <span
                className="text-transparent bg-clip-text bg-gradient-to-r"
                style={{
                  backgroundImage: `linear-gradient(to right, ${config.primary_color}, ${config.secondary_color})`,
                }}
              >
                卡密{config.hero_title.split('卡密')[1]}
              </span>
            </>
          ) : (
            <span
              className="text-transparent bg-clip-text bg-gradient-to-r"
              style={{
                backgroundImage: `linear-gradient(to right, ${config.primary_color}, ${config.secondary_color})`,
              }}
            >
              {config.hero_title}
            </span>
          )}
        </h1>

        <p
          className="text-xl mb-8"
          style={{ color: 'var(--text-muted)' }}
        >
          {config.hero_subtitle}
        </p>

        <div>
          <UserRouteLink
            view="products"
            className="btn btn-primary btn-lg inline-flex items-center gap-2"
            style={{
              background: `linear-gradient(135deg, ${config.primary_color}, ${config.secondary_color})`,
            }}
          >
            <i className="fas fa-box-open" />
            {config.hero_button_text}
          </UserRouteLink>
        </div>
      </div>
    </section>
  )
}
