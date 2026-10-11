import { defineConfig } from 'vite'

export default defineConfig({
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      '/ws': { target: 'ws://127.0.0.1:7070', ws: true },
      // Before /api: Vite sends a request to the first prefix that matches.
      '/api/auth': 'http://127.0.0.1:7272',
      '/api': 'http://127.0.0.1:7070',
    },
  },
})
