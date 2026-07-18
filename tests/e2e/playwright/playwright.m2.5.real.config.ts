import { defineConfig, devices } from '@playwright/test';

// @ctx: M2.5 E2E real-backend config — starts docker-compose.test.yml (5 services)
// + Vite dev server proxied to real api-gateway-test (port 8081)
export default defineConfig({
  testDir: './tests/m2.5',
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
        'docker compose -f deploy/docker-compose.test.yml up -d --wait --wait-timeout 60',
      port: 8081,
      reuseExistingServer: true,
      timeout: 90_000,
    },
    {
      command:
        'npx vite --host 0.0.0.0 --port 3000',
      env: {
        VITE_API_TARGET: 'http://localhost:8081',
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
