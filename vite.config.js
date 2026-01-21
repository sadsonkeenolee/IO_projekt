import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/v1/api': {
        target: 'http://localhost:9997', 
        changeOrigin: true,
      },
      '/v1/auth': {
        target: 'http://localhost:9999',
        changeOrigin: true,
      },
      '/recommend': {
        target: 'http://localhost:8001',
        changeOrigin: true,
      },
    }
  }
})
