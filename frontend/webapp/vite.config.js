import path from "path"
import { defineConfig } from "vite"
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig(({ command }) => ({
  plugins: [ 
    tailwindcss()
  ],
  server: {
    proxy: command === 'serve' ? {
      '/hack': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/hack/, '')
      }
    } : undefined
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
}))
