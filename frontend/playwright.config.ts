import { defineConfig, devices } from '@playwright/test'

const apiURL = 'http://127.0.0.1:18080'
const appURL = 'http://127.0.0.1:5174'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  workers: process.env.CI ? 2 : 4,
  timeout: 30_000,
  expect: { timeout: 5_000 },
  retries: process.env.CI ? 2 : 0,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: appURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: [
    {
      command: 'go run ./cmd/server',
      cwd: '..',
      url: `${apiURL}/health`,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      env: {
        TABLON_PORT: '18080',
        TABLON_DATA_DIR: './tmp/e2e-data',
        TABLON_API_TOKEN: 'e2e-token',
        TABLON_LOG_LEVEL: 'error',
      },
    },
    {
      command: 'pnpm dev --host 127.0.0.1 --port 5174 --strictPort',
      url: appURL,
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      env: {
        TABLON_PROXY_URL: apiURL,
      },
    },
  ],
})
