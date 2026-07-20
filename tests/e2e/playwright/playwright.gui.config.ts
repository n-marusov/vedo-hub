import { defineConfig, devices } from '@playwright/test';

/**
 * GUI test configuration — page rendering, user stories, a11y, keyboard nav.
 * Runs AFTER API tests in CI gates. Stops on first failure for fast feedback.
 *
 * Usage:
 *   pnpm exec playwright test --config=playwright.gui.config.ts
 */
export default defineConfig({
  testDir: './tests/gui',
  timeout: 30_000,
  retries: 0,           // no retries — first failure stops the run
  maxFailures: 1,        // stop entire run after 1 failure
  reporter: [['list'], ['html'], ['json', { outputFile: 'test-results/gui-results.json' }]],
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
