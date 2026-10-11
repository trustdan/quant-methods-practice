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
    const timeout = setTimeout(() => reject(new Error('Visuals server startup timed out')), 10000);
    let output = '';
    server.on('error', (error) => { clearTimeout(timeout); reject(error); });
    server.once('exit', (code) => { clearTimeout(timeout); reject(new Error(`Visuals server exited (${code}): ${output}`)); });
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
  tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'quant-visuals-'));
  url = await startServer();
});

test.afterAll(async () => {
  await stopServer();
  fs.rmSync(tempDir, { recursive: true, force: true });
});

test('distribution visuals lab, continuous PDF/CDF evaluation, interactive Venn diagram, and responsive graphics', async ({ page, context }) => {
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

  // 1. Navigate to Visuals tab via header button or 'v' shortcut
  const visualsTabBtn = page.locator('#tab-visuals');
  await expect(visualsTabBtn).toBeVisible();
  await visualsTabBtn.click();

  // Verify Visuals Explorer is rendered
  const labHeading = page.getByRole('heading', { name: /Mathematical Visuals & Distribution Lab/i });
  await expect(labHeading).toBeVisible();

  // 2. Normal Distribution Explorer
  const normalTitle = page.getByRole('heading', { name: /Normal Distribution Explorer/i });
  await expect(normalTitle).toBeVisible();

  // Verify Normal SVG chart elements
  const normalSvg = page.locator('svg[role="img"]').first();
  await expect(normalSvg).toBeVisible();

  // Check initial probability card and formula
  const normalProb = page.locator('#normal-shaded-prob-value');
  await expect(normalProb).toBeVisible();
  const normalFormula = page.locator('#normal-excel-formula');
  await expect(normalFormula).toContainText('NORM');

  // Switch tail to Upper Tail
  const tailSelect = page.locator('#norm-tail-select');
  await tailSelect.selectOption('upper');
  await page.waitForTimeout(300);

  // Toggle to CDF view
  const cdfBtn = page.locator('#btn-toggle-normal-cdf');
  await cdfBtn.click();
  await expect(cdfBtn).toHaveAttribute('aria-pressed', 'true');

  // 3. Switch to Uniform Distribution tab
  const uniformTab = page.locator('#tab-visual-uniform');
  await uniformTab.click();
  const uniformTitle = page.getByRole('heading', { name: /Uniform Distribution Explorer/i });
  await expect(uniformTitle).toBeVisible();

  // Verify constant height density is shown
  const densityValue = page.locator('#uniform-density-value');
  await expect(densityValue).toContainText('0.05');

  // 4. Switch to Exponential Distribution tab
  const expTab = page.locator('#tab-visual-exponential');
  await expTab.click();
  const expTitle = page.getByRole('heading', { name: /Exponential Distribution Explorer/i });
  await expect(expTitle).toBeVisible();

  // Verify Memoryless Property demonstrator
  await expect(page.getByText(/Interactive Memoryless Property Check/i)).toBeVisible();
  await expect(page.locator('#exp-excel-formula')).toContainText('EXPON.DIST');

  // 5. Switch to Discrete PMF / CDF tab
  const discreteTab = page.locator('#tab-visual-discrete');
  await discreteTab.click();
  const discreteTitle = page.getByRole('heading', { name: /Discrete Distribution Explorer/i });
  await expect(discreteTitle).toBeVisible();

  // Verify PMF bars and moments
  await expect(page.locator('#discrete-total-prob-value')).toBeVisible();
  await expect(page.locator('#discrete-excel-formula')).toContainText('BINOM.DIST');

  // Toggle to Poisson
  const familySelect = page.locator('#disc-family-select');
  await familySelect.selectOption('poisson');
  await page.waitForTimeout(300);
  await expect(page.locator('#discrete-excel-formula')).toContainText('POISSON.DIST');

  // 6. Switch to Venn Diagram tab
  const vennTab = page.locator('#tab-visual-venn');
  await vennTab.click();
  const vennTitle = page.getByRole('heading', { name: /Venn Diagram Explorer/i });
  await expect(vennTitle).toBeVisible();

  // Click Union button
  const unionBtn = page.locator('#btn-venn-union');
  await unionBtn.click();
  await expect(unionBtn).toHaveAttribute('aria-pressed', 'true');
  await expect(page.locator('#venn-formula')).toContainText('cup');

  // 7. Responsive check on narrow viewport (375px mobile)
  await page.setViewportSize({ width: 375, height: 667 });
  await page.waitForTimeout(300);
  const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
  expect(scrollWidth).toBeLessThanOrEqual(376);

  // Restore desktop viewport
  await page.setViewportSize({ width: 1280, height: 800 });

  // 8. Capture verified screenshot for audit
  const screenshotDir = path.join(__dirname, 'screenshots');
  if (!fs.existsSync(screenshotDir)) {
    fs.mkdirSync(screenshotDir, { recursive: true });
  }
  const screenshotPath = path.join(screenshotDir, 'stage15_verified.png');
  await page.screenshot({ path: screenshotPath, fullPage: true });
  expect(fs.existsSync(screenshotPath)).toBe(true);

  // 9. Press Escape to return to practice drill
  await page.keyboard.press('Escape');
  await expect(labHeading).not.toBeVisible();
  const practiceBtn = page.getByRole('button', { name: 'Practice Drill' });
  await expect(practiceBtn).toBeVisible();
  await expect(practiceBtn).toHaveClass(/btn-primary/);
});
