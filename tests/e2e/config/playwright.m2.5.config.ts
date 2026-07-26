import { defineConfig, devices } from '@playwright/test';

// @ctx: M2.5 E2E test config — starts Vite dev server + stub API server
export default defineConfig({
  testDir: '../specs/m2.5',
  timeout: 30_000,
  retries: 0,
  reporter: [['list']],
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'retain-on-failure',
  },
  webServer: [
    {
      command: 'node ../scripts/stub-server.mjs',
      port: 3001,
      reuseExistingServer: false,
      timeout: 10_000,
    },
    {
      command: 'npx vite --host 0.0.0.0 --port 3000',
      port: 3000,
      reuseExistingServer: false,
      timeout: 15_000,
      cwd: '../../../apps/services/frontend',
    },
  ],
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
});
