import { defineConfig } from 'vite'
import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'
import tailwindcss from '@tailwindcss/vite'

const target = 'http://localhost:8080/api'
// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    babel({ presets: [reactCompilerPreset()] }),
    tailwindcss(),
  ],
  base: "/",
  server: {
    proxy: {
      '/ws': {
        target,
        ws: true,
        changeOrigin: true,
      },
      '/login': {
        target,
        changeOrigin: true,
      },
      '/hello': {
        target,
        changeOrigin: true,
      },
      '/test': {
        target,
        changeOrigin: true,
      },
      '/ping':
      {
        target,
        changeOrigin: true
      }
    },
  },
})