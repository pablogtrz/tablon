import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'

const apiProxy = process.env.TABLON_PROXY_URL ?? 'http://127.0.0.1:8080'

export default defineConfig({
  plugins: [tailwindcss(), svelte()],
  resolve: {
    alias: {
      $lib: path.resolve('./src/lib'),
    },
  },
  server: {
    proxy: {
      '/api': apiProxy,
      '/health': apiProxy,
      '/ready': apiProxy,
    },
  },
})
