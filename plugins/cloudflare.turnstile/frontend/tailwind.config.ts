import type { Config } from 'tailwindcss'

// 与主程序 Program/web/tailwind.config.ts 保持一致的设计 token。
// 版本跟随主程序：主程序升级 Tailwind 大版本或调整 token 时，插件前端须同步更新。
const config: Config = {
  content: [
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0f1ff',
          100: '#e4e6ff',
          200: '#cdd0ff',
          300: '#a5abff',
          400: '#7c7dff',
          500: '#667eea',
          600: '#5046e5',
          700: '#4338ca',
          800: '#3730a3',
          900: '#312e81',
        },
        dark: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
          950: '#020617',
        },
      },
    },
  },
  plugins: [],
}
export default config
