import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@/components': path.resolve(__dirname, './src/components'),
      '@/hooks': path.resolve(__dirname, './src/hooks'),
      '@/utils': path.resolve(__dirname, './src/utils'),
      '@/types': path.resolve(__dirname, './src/types'),
      '@/config': path.resolve(__dirname, './src/config'),
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
  server: {
    host: true,
    port: Number(process.env.PORT) || 3000,
    watch: {
      usePolling: true,
      interval: 1000,
    },
  },
})