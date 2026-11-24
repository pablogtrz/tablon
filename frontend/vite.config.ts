import { mkdirSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'

const apiProxy = process.env.TABLON_PROXY_URL ?? 'http://127.0.0.1:8448'

export default defineConfig({
  plugins: [
    tailwindcss(),
    svelte(),
    {
      // `go:embed all:frontend/dist` needs the directory to exist even when the
      // SPA has not been built. Vite empties dist/, so restore the tracked
      // placeholder after each build.
      name: 'tablon-dist-placeholder',
      closeBundle() {
        mkdirSync(path.resolve('dist'), { recursive: true })
        writeFileSync(path.resolve('dist/.gitkeep'), '')
      },
    },
  ],
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
