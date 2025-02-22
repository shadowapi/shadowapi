import path from 'path'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve('./src'),
    },
  },
  build: {
    sourcemap: true,
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    allowedHosts: ['localtest.me', 'localtest.me:5173', 'localtest.me:3000'],
    strictPort: true,
    cors: true,
  },
})
