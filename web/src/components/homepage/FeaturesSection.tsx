'use client'

import type { HomepageConfig } from '@/types/homepage'

interface FeaturesSectionProps {
  config: HomepageConfig
}

/**
 * 特性区块组件
 */
export function FeaturesSection({ config }: FeaturesSectionProps) {
  if (!config.features_enabled || !config.features?.length) return null

  // 判断图标是否是 Font Awesome 图标
  const renderIcon = (icon: string) => {
    if (icon.startsWith('fa-')) {
      return <i className={`fas ${icon} text-3xl`} style={{ color: config.primary_color }} />
    }
    return <span className="text-4xl">{icon}</span>
  }

  return (
    <section className="py-16 px-4">
      <div className="max-w-6xl mx-auto">
        {config.features_title && (
          <h2
            className="text-3xl font-bold text-center mb-12"
            style={{ color: 'var(--text-primary)' }}
          >
            {config.features_title}
          </h2>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {config.features.map((feature, index) => (
            <div
              key={index}
              className="card p-6 text-center hover:shadow-lg transition-shadow"
            >
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl flex items-center justify-center"
                style={{ backgroundColor: `${config.primary_color}15` }}
              >
                {renderIcon(feature.icon)}
              </div>
              <h3 className="text-lg font-semibold mb-2" style={{ color: 'var(--text-primary)' }}>
                {feature.title}
              </h3>
              <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
