import { defineConfig, devices } from '@playwright/test';

/**
 * Real-backend E2E config — starts docker-compose.test.yml (full stack
 * with JWT_DEV_PUBLIC_KEY_PEM) + Vite dev server on port 3000.
 *
 * Usage:
 *   pnpm exec playwright test --config=playwright.m2.5.real.config.ts
 */
export default defineConfig({
  testDir: '../specs',
  timeout: 30_000,
  retries: 0,
  reporter: [['list']],
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'retain-on-failure',
  },
  webServer: [
    {
      command:
        'docker compose -f ../../../deploy/docker-compose.test.yml up -d --wait --wait-timeout 60',
      port: 8080,
      reuseExistingServer: true,
      timeout: 90_000,
    },
    {
      command:
        'npx vite --host 0.0.0.0 --port 3000',
      env: {
        VITE_API_TARGET: 'http://localhost:8080',
      },
      port: 3000,
      reuseExistingServer: true,
      timeout: 15_000,
      cwd: '../../../src/services/frontend',
    },
  ],
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
});
