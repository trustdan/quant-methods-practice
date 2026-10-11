import { test, expect } from '@playwright/test';
import { spawn, ChildProcess } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';

const root = path.resolve(__dirname, '../..');
const binary = path.join(root, 'bin', process.platform === 'win32' ? 'quant-practice.exe' : 'quant-practice');
let server: ChildProcess;
let tempDir: string;
let url: string;

async function startServer(): Promise<string> {
  server = spawn(binary, ['--no-browser', '--data-dir', tempDir], { cwd: root, stdio: ['ignore', 'pipe', 'pipe'] });
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => reject(new Error('Arcade server startup timed out')), 10000);
    let output = '';
    server.on('error', (error) => { clearTimeout(timeout); reject(error); });
    server.once('exit', (code) => { clearTimeout(timeout); reject(new Error(`Arcade server exited (${code}): ${output}`)); });
    const onData = (data: Buffer) => {
      output += data.toString();
      const match = output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/);
      if (match) { clearTimeout(timeout); resolve(match[0]); }
    };
    server.stdout?.on('data', onData);
    server.stderr?.on('data', onData);
  });
}

async function stopServer() {
  if (server.exitCode !== null) return;
  await new Promise<void>((resolve) => {
    server.once('exit', () => resolve());
    server.kill('SIGINT');
  });
}

test.beforeAll(async () => {
  tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'quant-arcade-'));
  url = await startServer();
});

test.afterAll(async () => {
  await stopServer();
  fs.rmSync(tempDir, { recursive: true, force: true });
});

test('quant flight arcade, attract mode, manual takeover, pause, skip, and mobile responsiveness', async ({ page, context }) => {
  // Offline loopback enforcement
  await context.route('**/*', (route) => {
    if (route.request().url().startsWith('http://127.0.0.1:')) {
      route.continue();
    } else {
      route.abort();
    }
  });

  page.on('pageerror', (err) => {
    console.error('BROWSER ERROR:', err.message, err.stack);
  });

  await page.goto(url);
  await page.waitForLoadState('networkidle');

  // Verify practice drill is initially loaded
  const practiceTab = page.getByRole('button', { name: 'Practice Drill' });
  await expect(practiceTab).toBeVisible();

  // 1. Open Arcade via Tab Button
  const arcadeTab = page.locator('#tab-arcade');
  await expect(arcadeTab).toBeVisible();
  await arcadeTab.click();

  // 2. Verify Arcade Header and Attract Mode
  const arcadeHeading = page.getByRole('heading', { name: /Quant Flight/i });
  await expect(arcadeHeading).toBeVisible();

  const attractBadge = page.getByText(/⚡ Attract Mode/i);
  await expect(attractBadge).toBeVisible();

  const canvas = page.locator('#arcade-canvas');
  await expect(canvas).toBeVisible();

  // 3. Engage Manual Flight with player input (Space)
  await page.keyboard.press('Space');
  await page.waitForTimeout(100);

  const manualBadge = page.getByText(/🎮 Manual Flight/i);
  await expect(manualBadge).toBeVisible();

  // 4. Test Pause Toggle Button
  const pauseBtn = page.locator('#btn-pause-arcade');
  await expect(pauseBtn).toBeVisible();
  await pauseBtn.click();
  await expect(page.getByRole('button', { name: /Resume/i })).toBeVisible();

  // Press P to Resume
  await page.keyboard.press('p');
  await expect(page.getByRole('button', { name: /Pause/i })).toBeVisible();

  // 5. Test Reduced Motion Toggle
  const motionToggle = page.locator('#arcade-reduced-motion-toggle');
  await expect(motionToggle).toBeVisible();
  await motionToggle.check();
  await expect(motionToggle).toBeChecked();

  // 6. Test 375px mobile viewport responsiveness
  await page.setViewportSize({ width: 375, height: 667 });
  await page.waitForTimeout(300);
  const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
  expect(scrollWidth).toBeLessThanOrEqual(376);

  // Restore desktop viewport
  await page.setViewportSize({ width: 1280, height: 800 });
  await page.evaluate(() => window.scrollTo(0, 0));
  await page.waitForTimeout(200);

  // 7. Capture verified screenshot for audit
  const screenshotDir = path.join(__dirname, 'screenshots');
  if (!fs.existsSync(screenshotDir)) {
    fs.mkdirSync(screenshotDir, { recursive: true });
  }
  const screenshotPath = path.join(screenshotDir, 'stage17_verified.png');
  await page.screenshot({ path: screenshotPath, fullPage: true });
  expect(fs.existsSync(screenshotPath)).toBe(true);

  // 8. Test Skip to Practice button
  const skipBtn = page.locator('#btn-skip-arcade');
  await expect(skipBtn).toBeVisible();
  await skipBtn.click();

  // Verify immediate return to practice drill
  await expect(arcadeHeading).not.toBeVisible();
  await expect(practiceTab).toHaveClass(/btn-primary/);

  // 9. Re-enter Arcade and verify Escape returns to practice
  await arcadeTab.click();
  await expect(arcadeHeading).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(arcadeHeading).not.toBeVisible();
  await expect(practiceTab).toHaveClass(/btn-primary/);
});
