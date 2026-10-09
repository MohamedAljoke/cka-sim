import { defineConfig } from 'vite'

export default defineConfig({
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      '/ws': { target: 'ws://127.0.0.1:7070', ws: true },
      '/api': 'http://127.0.0.1:7070',
    },
  },
})
