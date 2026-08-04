import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'

// https://vitejs.dev/config/
export default defineConfig({
  // Served at the domain root (user page / apex domain), so assets and the
  // router base ('import.meta.env.BASE_URL') resolve from '/'. Change to
  // '/<repo>/' only if this ever moves to a GitHub Pages project page.
  base: '/',
  plugins: [
    vue(),
    vueJsx(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})
