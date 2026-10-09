import { defineConfig } from 'vite'

export default defineConfig({
  build: {
    // go:embed can't reach outside apps/cka-sim, so the page is built straight into that module.
    outDir: '../cka-sim/internal/web/static/dist',
    emptyOutDir: true,
  },
})
