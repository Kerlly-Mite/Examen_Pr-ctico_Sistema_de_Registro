import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'

// En desarrollo y en `vite preview`, las llamadas a /api se reenvían al backend en Go.
const proxy = { '/api': { target: 'http://localhost:3001', changeOrigin: true } }

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      // injectManifest: el service worker lo escribimos nosotros (src/sw.ts) y Workbox solo le
      // inyecta la lista de archivos del build para el precaché del "app shell".
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'sw.ts',
      // El registro lo hace PwaStatus (virtual:pwa-register/react) y solo en producción.
      injectRegister: false,
      registerType: 'prompt',
      // El service worker NO corre con `npm run dev`: se prueba con `npm run build` + `npm run preview`.
      devOptions: { enabled: false },
      includeAssets: ['favicon.svg', 'icons/apple-touch-icon.png', 'offline.html'],
      injectManifest: {
        globPatterns: ['**/*.{js,css,html,svg,png,webmanifest}'],
        maximumFileSizeToCacheInBytes: 3 * 1024 * 1024,
      },
      manifest: {
        id: '/',
        name: 'Postulación Directorio BIESS',
        short_name: 'Postulación BIESS',
        description: 'Registro y postulación al Concurso de Méritos y Oposición del Directorio del BIESS.',
        lang: 'es',
        start_url: '/',
        scope: '/',
        display: 'standalone',
        orientation: 'any',
        theme_color: '#17356b',
        background_color: '#f8fafc',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
          { src: '/icons/icon-maskable-512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
    }),
  ],
  server: {
    port: 5180,
    strictPort: true,
    proxy,
  },
  // `npm run preview` sirve el build de producción (con service worker) en el mismo puerto 5180.
  preview: {
    port: 5180,
    strictPort: true,
    proxy,
  },
})
