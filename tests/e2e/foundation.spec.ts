import { test, expect } from '@playwright/test';
import { spawn, ChildProcess } from 'child_process';
import path from 'path';
import fs from 'fs';

const BIN_PATH = path.resolve(__dirname, '../../bin/quant-practice.exe');
const PORT = 8995;

let serverProcess: ChildProcess | null = null;
let bootstrapUrl = '';

test.beforeAll(async () => {
  // Ensure bin exists
  if (!fs.existsSync(BIN_PATH)) {
    throw new Error(`Binary not found at ${BIN_PATH}. Run build first.`);
  }

  // Start Go loopback server
  serverProcess = spawn(BIN_PATH, ['--port', String(PORT), '--no-browser'], {
    stdio: ['ignore', 'pipe', 'pipe'],
  });

  // Extract bootstrap URL from server stdout/stderr
  await new Promise<void>((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error('Timed out waiting for loopback server startup'));
    }, 10000);

    const onData = (data: Buffer) => {
      const text = data.toString();
      const match = text.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/);
      if (match) {
        bootstrapUrl = match[0];
        clearTimeout(timeout);
        resolve();
      }
    };

    serverProcess?.stdout?.on('data', onData);
    serverProcess?.stderr?.on('data', onData);
  });
});

test.afterAll(async () => {
  if (serverProcess) {
    serverProcess.kill('SIGINT');
  }
});

test.describe('Foundation Verification (Offline Math & Keyboard Shell)', () => {
  test('renders math completely offline and supports full keyboard navigation', async ({ context, page }) => {
    // Strictly disable all external networking: abort any request not targeting the loopback server
    await context.route('**/*', (route) => {
      const url = route.request().url();
      if (!url.startsWith(`http://127.0.0.1:${PORT}`)) {
        console.warn(`Blocked unexpected external request to: ${url}`);
        route.abort();
      } else {
        route.continue();
      }
    });

    // Navigate to the bootstrap URL
    await page.goto(bootstrapUrl);

    // 1. Verify title and basic header presence
    await expect(page).toHaveTitle('Quant Methods Practice');
    await expect(page.locator('.app-logo')).toContainText('Quant Methods Practice');
    await expect(page.locator('.app-header')).toContainText('Loopback Server Active');

    // 2. Verify static demonstration notice is clearly presented
    await expect(page.getByText('Stage 01 Foundation: Mathematical Rendering & Keyboard Shell Demonstration (Ungraded)')).toBeVisible();

    // 3. Verify local MathJax has rendered math formulas into SVG containers
    const mathContainer = page.locator('mjx-container[jax="SVG"]').first();
    await expect(mathContainer).toBeVisible({ timeout: 10000 });
    const svgElement = mathContainer.locator('svg');
    await expect(svgElement).toBeVisible();

    // 4. Test keyboard navigation: initial stage is 1/7 (Target)
    await expect(page.locator('.card-header').first()).toContainText('Stage 1/7: Target');

    // Press 'l' to advance to Stage 2/7 (Model)
    await page.keyboard.press('l');
    await expect(page.locator('.card-header').first()).toContainText('Stage 2/7: Model');

    // Press '2' to select choice option 2
    await page.keyboard.press('2');
    await expect(page.getByText('Choice 2 selected')).toBeVisible();

    // Press 'Enter' to open feedback / derivation panel
    await page.keyboard.press('Enter');
    await expect(page.getByText('Explanation & Mathematical Derivation:')).toBeVisible();

    // Press 'h' to go back to Stage 1/7 (Target)
    await page.keyboard.press('h');
    await expect(page.locator('.card-header').first()).toContainText('Stage 1/7: Target');

    // Press 'F1' to open Help modal
    await page.keyboard.press('F1');
    const helpModal = page.locator('div[role="dialog"]');
    await expect(helpModal).toBeVisible();
    await expect(helpModal).toContainText('Keyboard Navigation Contract');

    // Press 'Escape' to dismiss Help modal
    await page.keyboard.press('Escape');
    await expect(helpModal).not.toBeVisible();

    // 5. Capture screenshot of verified foundation UI
    const screenshotDir = path.resolve(__dirname, 'screenshots');
    fs.mkdirSync(screenshotDir, { recursive: true });
    await page.screenshot({ path: path.join(screenshotDir, 'foundation_verified.png'), fullPage: true });
  });
});
