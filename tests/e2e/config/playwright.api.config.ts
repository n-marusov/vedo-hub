import { defineConfig, devices } from '@playwright/test';

/**
 * API test configuration — fast REST/GraphQL integration tests.
 * Runs first in CI gates, provides quick feedback before slower GUI tests.
 *
 * Usage:
 *   pnpm exec playwright test --config=playwright.api.config.ts
 */
export default defineConfig({
  globalSetup: '../scripts/global-setup',
  testDir: '../specs/api',
  timeout: 30_000,
  retries: 2,
  reporter: [['list'], ['html'], ['json', { outputFile: '../test-results/api-results.json' }]],
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'retain-on-failure',
    video: 'retain-on-failure',
    screenshot: 'only-on-failure'
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
  webServer: {
    command: 'docker compose -f ../../../deploy/docker-compose.test.yml up -d --wait --wait-timeout 120',
    url: 'http://localhost:3000/health',
    reuseExistingServer: true,
    timeout: 600_000
  }
});
