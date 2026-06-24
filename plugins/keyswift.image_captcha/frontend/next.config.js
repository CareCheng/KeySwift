/** @type {import('next').NextConfig} */
// 插件前端静态导出配置：与主程序 Program/web 同栈同模式（output: export）。
// 产物为单页 widget，构建后由插件 build 脚本重命名为 frontend/widget.html，
// 供宿主 /api/plugins/keyswift.image_captcha/1.0.0/frontend/widget.html 直接 iframe 加载。
//
// assetPrefix 必须指向宿主 serve 该插件前端资源的完整前缀路径，
// 使产物内 _next 资源引用解析到 /api/plugins/<id>/<ver>/frontend/_next/...，
// 与后端 PluginFrontendAsset 的 serve 路径完全对应，避免 404。
// 版本或插件 id 变更时需同步此 assetPrefix。
const nextConfig = {
  output: 'export',
  trailingSlash: false,
  assetPrefix: '/api/plugins/keyswift.image_captcha/1.0.0/frontend',
  images: {
    unoptimized: true,
  },
}

module.exports = nextConfig
