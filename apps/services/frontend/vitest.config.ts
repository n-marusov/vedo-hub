import { fileURLToPath } from 'node:url'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

/**
 * Vitest configuration for the VEDO Core frontend.
 *
 * The test environment is set to jsdom so Vue component tests using
 * @vue/test-utils can mount components that touch the DOM. The configuration
 * mirrors the Vite setup used for builds so import resolution stays
 * consistent between development, production, and test runs.
 *
 * @see https://vitest.dev/config/
 */
export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.test.ts', 'src/**/*.spec.ts'],
    setupFiles: ['vitest.setup.ts']
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})
