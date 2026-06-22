'use client'

import { HeroSection } from './HeroSection'
import { FeaturesSection } from './FeaturesSection'
import { ProductsSection } from './ProductsSection'
import { StatsSection } from './StatsSection'
import { CTASection } from './CTASection'
import { defaultHomepageConfig } from '@/types/homepage'

/**
 * 首页组件
 * 仅展示卡密销售核心内容。
 */
export function DynamicHomepage() {
  const config = defaultHomepageConfig

  return (
    <main className="flex-1">
      <HeroSection config={config} />
      <FeaturesSection config={config} />
      <ProductsSection config={config} />
      <StatsSection config={config} />
      <CTASection config={config} />
    </main>
  )
}

export { HeroSection } from './HeroSection'
export { FeaturesSection } from './FeaturesSection'
export { ProductsSection } from './ProductsSection'
export { StatsSection } from './StatsSection'
export { CTASection } from './CTASection'
