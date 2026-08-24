import { defineConfig, devices } from '@playwright/test';

// Tests always target the shared environment's endpoint. When a dev has deployed
// a service into their own namespace with Divert, OKTETO_SHARED_NAMESPACE points
// at the shared namespace while OKTETO_NAMESPACE is their personal one; the
// baggage header below tells the shared ingress which requests to divert back
// to that personal namespace. When running fully in the shared namespace (no
// divert), OKTETO_SHARED_NAMESPACE is unset and both values are the same.
const SHARED_NAMESPACE = process.env.OKTETO_SHARED_NAMESPACE || process.env.OKTETO_NAMESPACE;
const BASE_URL = `https://movies-${SHARED_NAMESPACE}.${process.env.OKTETO_DOMAIN}`;

/**
 * @see https://playwright.dev/docs/test-configuration
 */
const config = defineConfig({
  testDir: './tests',
  /* Run tests in files in parallel */
  fullyParallel: true,
  /* Fail the build on CI if you accidentally left test.only in the source code. */
  forbidOnly: !!process.env.CI,
  /* Retry on CI only */
  retries: process.env.CI ? 0 : 0,
  /* Opt out of parallel tests on CI. */
  workers: process.env.CI ? 1 : undefined,
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: [['list'], ['html', { open: 'never' }]],
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL: BASE_URL,

    /* Route requests to this namespace's diverted services on the shared environment. */
    extraHTTPHeaders: {
      baggage: `okteto-divert=${process.env.OKTETO_NAMESPACE}`,
    },

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});

export default config;