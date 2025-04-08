import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import pluginPurgeCss from "@myelophone/vite-plugin-purgecss";

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte(), pluginPurgeCss()],
  build: {
    chunkSizeWarningLimit: 1200,
    rollupOptions: {
      treeshake: 'smallest',
      output: {
        manualChunks: {
          echarts: ['echarts'],
          svelte: ['svelte'],
        }
      }
    }
  }
});
