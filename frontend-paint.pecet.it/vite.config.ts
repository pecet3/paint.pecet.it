import { defineConfig } from 'vite'
import react, { reactCompilerPreset } from '@vitejs/plugin-react'
import babel from '@rolldown/plugin-babel'
import tailwindcss from '@tailwindcss/vite'

const target = 'http://localhost:8080'
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
      '/api/ws': {
        target,
        ws: true,
        changeOrigin: true,
      },
      '/api/login': {
        target,
        changeOrigin: true,
      },
      '/api/ping':
      {
        target,
        changeOrigin: true
      }
    },
  },
})