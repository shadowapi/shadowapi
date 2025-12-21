import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve('./src'),
    },
  },
  define: {
    'process.env': {},
  },
  build: {
    sourcemap: true,
  },
  // SSR configuration
  ssr: {
    // Don't externalize Ant Design ecosystem packages in SSR build
    // Include @emotion for hash function used by cssinjs
    noExternal: [/antd/, /@ant-design\//, /rc-/, /@emotion\//],
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    watch: {
      usePolling: true,
    },
    allowedHosts: true,
  },
})
