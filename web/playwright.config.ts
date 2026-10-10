import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: '../tests/e2e',
  timeout: 30000,
  fullyParallel: false,
  workers: 1,
  use: {
    channel: process.platform === 'win32' ? 'msedge' : undefined,
    headless: true,
  },
});
