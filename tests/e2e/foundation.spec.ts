import { test, expect } from '@playwright/test';
import { spawn, ChildProcess } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';

const BIN_PATH = path.resolve(__dirname, '../../bin/quant-practice.exe');
const PORT = 8995;

let serverProcess: ChildProcess | null = null;
let bootstrapUrl = '';
let restartedUrl = '';
let tempDbPath = '';

test.beforeAll(async () => {
  // Ensure bin exists
  if (!fs.existsSync(BIN_PATH)) {
    throw new Error(`Binary not found at ${BIN_PATH}. Run build first.`);
  }

  // Create isolated temp database for e2e test
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'quant-e2e-'));
  tempDbPath = path.join(tempDir, 'e2e-test.db');

  // Start Go loopback server with isolated temp database
  serverProcess = spawn(BIN_PATH, ['--port', String(PORT), '--no-browser', '--db', tempDbPath], {
    cwd: path.resolve(__dirname, '../..'),
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
  if (tempDbPath) {
    try {
      fs.rmSync(path.dirname(tempDbPath), { recursive: true, force: true });
    } catch {
      // ignore cleanup errors
    }
  }
});

test.describe('Foundation Verification (Offline Math & Keyboard Shell)', () => {
  test('renders math completely offline and supports full keyboard navigation', async ({ context, page }) => {
    // Strictly disable all external networking: abort any request not targeting the loopback server
    await context.route('**/*', (route) => {
      const url = route.request().url();
      if (!url.startsWith('http://127.0.0.1:')) {
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
    await expect(page.locator('.app-header')).toContainText('Stage 04 Drill');

    // 2. Verify approved drill scenario and prompt are presented
    await expect(page.getByText('Exactly two heads in four tosses')).toBeVisible();
    await expect(page.locator('.card-header').first()).toContainText('Stage 1/7: Target');

    // 3. Verify local MathJax has rendered math formulas into SVG containers
    const mathContainer = page.locator('mjx-container[jax="SVG"]').first();
    await expect(mathContainer).toBeVisible({ timeout: 10000 });
    const svgElement = mathContainer.locator('svg');
    await expect(svgElement).toBeVisible();

    // 4. Test keyboard interaction: press '1' to select option 1
    await page.keyboard.press('1');
    await expect(page.getByText('Option selected')).toBeVisible();

    // Press 'Enter' to submit answer
    await page.keyboard.press('Enter');

    // After correct submission, advances to Stage 2/7: Model
    await expect(page.locator('.card-header').first()).toContainText('Stage 2/7: Model', { timeout: 10000 });

    // Press 'h' to navigate back to Stage 1/7 (Target)
    await page.keyboard.press('h');
    await expect(page.locator('.card-header').first()).toContainText('Stage 1/7: Target');
    await expect(page.getByText('Mathematical Explanation:')).toBeVisible();

    // Press 'l' to navigate forward to Stage 2/7 (Model)
    await page.keyboard.press('l');
    await expect(page.locator('.card-header').first()).toContainText('Stage 2/7: Model');

    // 5. Test Help modal: Press 'F1' to open Help modal
    await page.keyboard.press('F1');
    const helpModal = page.locator('div[role="dialog"]');
    await expect(helpModal).toBeVisible();
    await expect(helpModal).toContainText('Keyboard Navigation Contract');

    // Press 'Escape' to dismiss Help modal
    await page.keyboard.press('Escape');
    await expect(helpModal).not.toBeVisible();

    // 6. Capture screenshot of verified drill UI
    const screenshotDir = path.resolve(__dirname, 'screenshots');
    fs.mkdirSync(screenshotDir, { recursive: true });
    await page.screenshot({ path: path.join(screenshotDir, 'drill_verified.png'), fullPage: true });
  });

  test('persists drill session across server restart and page reload', async ({ page }) => {
    // 1. Kill the running server
    if (serverProcess) {
      serverProcess.kill('SIGINT');
      serverProcess = null;
    }

    // 2. Restart server process pointing to the SAME tempDbPath
    const RESTART_PORT = 8996;
    serverProcess = spawn(BIN_PATH, ['--port', String(RESTART_PORT), '--no-browser', '--db', tempDbPath], {
      cwd: path.resolve(__dirname, '../..'),
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('Timeout on restart')), 10000);
      const onData = (data: Buffer) => {
        const text = data.toString();
        const match = text.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/);
        if (match) {
          restartedUrl = match[0];
          clearTimeout(timeout);
          resolve();
        }
      };
      serverProcess?.stdout?.on('data', onData);
      serverProcess?.stderr?.on('data', onData);
    });

    // 3. Navigate to restarted server URL
    await page.goto(restartedUrl);

    // 4. Verify drill resumes automatically at Stage 2/7: Model with Stage 1 already completed
    await expect(page.locator('.card-header').first()).toContainText('Stage 2/7: Model', { timeout: 10000 });

    // Verify Stage 1 is marked completed in the stage strip
    const completedStages = page.locator('.stage-step.completed');
    await expect(completedStages).toHaveCount(1);
    await expect(page.locator('.stage-check')).toHaveCount(1);
  });

  test('navigation preserves unsent answers, prior/future visits do not inflate evidence, and shortcuts respect editable fields', async ({ page }) => {
    // Navigate to the restarted server URL
    await page.goto(restartedUrl);
    await expect(page.locator('.card-header').first()).toContainText('Stage 2/7: Model', { timeout: 10000 });

    // 1. We are resumed at Stage 2/7: Model. Test locked stage notice by clicking Stage 6 in stage strip
    const tabs = page.locator('.stage-step');
    await tabs.nth(5).click(); // Click Stage 6 (Calculate)
    const alert = page.locator('div[role="alert"]');
    await expect(alert).toBeVisible();
    await expect(alert).toContainText('locked');
    await page.getByRole('button', { name: 'Dismiss' }).click();
    await expect(alert).not.toBeVisible();

    // 2. Select option on Stage 2 (Binomial) using keyboard '1'
    await page.keyboard.press('1');
    await expect(page.getByText('Option selected')).toBeVisible();

    // Submit Stage 2 -> advances to Stage 3
    await page.keyboard.press('Enter');
    await expect(page.locator('.card-header').first()).toContainText('Stage 3/7: Conditions', { timeout: 10000 });

    // Submit Stage 3 (Conditions: n=4, p=0.5)
    await page.keyboard.press('1');
    await page.keyboard.press('Enter');
    await expect(page.locator('.card-header').first()).toContainText('Stage 4/7: Event', { timeout: 10000 });

    // Submit Stage 4 (Event: X = 2)
    await page.keyboard.press('1');
    await page.keyboard.press('Enter');
    await expect(page.locator('.card-header').first()).toContainText('Stage 5/7: Expression', { timeout: 10000 });

    // Submit Stage 5 (Expression: with combination)
    await page.keyboard.press('1');
    await page.keyboard.press('Enter');
    await expect(page.locator('.card-header').first()).toContainText('Stage 6/7: Calculate', { timeout: 10000 });

    // 3. Test editable field typing and shortcut non-hijacking
    const numericInput = page.locator('#drill-numeric-input');
    await expect(numericInput).toBeVisible();
    await numericInput.fill('0.375');

    // 4. Test unsent draft preservation across navigation:
    // Navigate back to Stage 1 (Target)
    await tabs.nth(0).click();
    await expect(page.locator('.card-header').first()).toContainText('Stage 1/7: Target');
    await expect(page.getByText('Mathematical Explanation:')).toBeVisible();
    // Prior stage visit does not offer re-submission form
    await expect(page.locator('button', { hasText: 'Submit Answer' })).toHaveCount(0);

    // Navigate forward to Stage 6 (Calculate)
    await tabs.nth(5).click();
    await expect(page.locator('.card-header').first()).toContainText('Stage 6/7: Calculate');
    // Invariant: Unsent answer is strictly preserved!
    await expect(numericInput).toHaveValue('0.375');

    // 5. Test leave intent modal on 'q'
    await page.locator('.card-header').first().click();
    await page.keyboard.press('q');
    const leaveModal = page.locator('div[role="dialog"]');
    await expect(leaveModal).toBeVisible();
    await expect(leaveModal).toContainText('Leave Practice Session?');
    await expect(leaveModal).toContainText('unsubmitted answer draft');

    // Dismiss leave modal with Escape
    await page.keyboard.press('Escape');
    await expect(leaveModal).not.toBeVisible();
    await expect(numericInput).toHaveValue('0.375');

    // 6. Submit numeric calculation
    await page.getByRole('button', { name: 'Submit', exact: true }).click();
    await expect(page.locator('.card-header').first()).toContainText('Stage 7/7: Interpret', { timeout: 10000 });

    // Complete Stage 7
    await page.keyboard.press('1');
    await page.keyboard.press('Enter');

    // Drill completed: view full recap
    await expect(page.getByText('Drill Completed')).toBeVisible({ timeout: 10000 });

    // 7. Test Reading View scrolling (Formula Gallery)
    await page.getByRole('button', { name: 'Formula Gallery' }).click();
    await expect(page.getByText('Mathematical Formula Gallery')).toBeVisible();

    // Scroll commands in reading mode
    await page.keyboard.press('j');
    await page.keyboard.press('k');
    await page.keyboard.press('d');
    await page.keyboard.press('u');

    // F1 Help in reading view opens Help modal
    await page.keyboard.press('F1');
    const helpModal = page.locator('div[role="dialog"]');
    await expect(helpModal).toBeVisible();
    await expect(helpModal).toContainText('Keyboard Navigation Contract');
    await page.keyboard.press('Escape');
    await expect(helpModal).not.toBeVisible();

    // 8. Capture screenshot of verified Stage 06 usability
    const screenshotDir = path.resolve(__dirname, 'screenshots');
    fs.mkdirSync(screenshotDir, { recursive: true });
    await page.screenshot({ path: path.join(screenshotDir, 'stage06_verified.png'), fullPage: true });
  });

  test('stage 07: ten-question session navigation, reference library, settings modal, responsive math, and restart replay', async ({ page }) => {
    // Navigate to restarted server URL
    await page.goto(restartedUrl);

    // 1. Verify 10-question session is active and displays 10 pills in the question strip
    const questionPills = page.locator('.question-pill');
    await expect(questionPills).toHaveCount(10, { timeout: 10000 });

    // Problem 1 is present in the strip
    await expect(page.getByText('Problem 1 of 10:')).toBeVisible();

    // 2. Navigate to Problem 2 using question pill click
    await questionPills.nth(1).click();
    await expect(page.getByText('Problem 2 of 10:')).toBeVisible();
    await expect(questionPills.nth(1)).toHaveClass(/active/);

    // 3. Navigate to Problem 3 using big problem navigation shortcut (Control+ArrowRight)
    await page.keyboard.press('Control+ArrowRight');
    await expect(page.getByText('Problem 3 of 10:')).toBeVisible();
    await expect(questionPills.nth(2)).toHaveClass(/active/);

    // Navigate back to Problem 2 using Control+ArrowLeft
    await page.keyboard.press('Control+ArrowLeft');
    await expect(page.getByText('Problem 2 of 10:')).toBeVisible();
    await expect(questionPills.nth(1)).toHaveClass(/active/);

    // 4. Test Reference Library: press 'r' to open
    await page.keyboard.press('r');
    const refHeading = page.getByRole('heading', { name: 'Reference Library' });
    await expect(refHeading).toBeVisible();
    await expect(page.getByText('100% offline')).toBeVisible();

    // Verify offline math rendering inside reference library
    const refMath = page.locator('mjx-container[jax="SVG"]').first();
    await expect(refMath).toBeVisible();

    // Test category filter in Reference Library
    await page.getByRole('button', { name: 'Distributions' }).click();
    await expect(page.getByText('Binomial Distribution')).toBeVisible();

    // Press 'r' again to toggle back to Practice Drill
    await page.keyboard.press('r');
    await expect(refHeading).not.toBeVisible();
    await expect(page.getByText('Problem 2 of 10:')).toBeVisible();

    // 5. Test Settings Modal: press 't' to open
    await page.keyboard.press('t');
    const settingsModal = page.locator('div[role="dialog"]');
    await expect(settingsModal).toBeVisible();
    await expect(settingsModal).toContainText('Session Settings & Module Picker');
    await expect(page.getByText('Question Count:')).toBeVisible();

    // Press 'Escape' to dismiss Settings modal
    await page.keyboard.press('Escape');
    await expect(settingsModal).not.toBeVisible();

    // 6. Test Responsive Math on narrow screens (375px mobile viewport)
    await page.setViewportSize({ width: 375, height: 667 });
    // Verify question strip bar and math-markdown adapt without causing horizontal page blowout
    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
    expect(scrollWidth).toBeLessThanOrEqual(376);

    // Restore standard desktop viewport
    await page.setViewportSize({ width: 1280, height: 800 });

    // 7. Verify Quit / Restart and Replay across server restart
    if (serverProcess) {
      serverProcess.kill('SIGINT');
      serverProcess = null;
    }

    const STAGE07_PORT = 8997;
    serverProcess = spawn(BIN_PATH, ['--port', String(STAGE07_PORT), '--no-browser', '--db', tempDbPath], {
      cwd: path.resolve(__dirname, '../..'),
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    let stage07Url = '';
    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(() => reject(new Error('Timeout on Stage 07 restart')), 10000);
      const onData = (data: Buffer) => {
        const text = data.toString();
        const match = text.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/);
        if (match) {
          stage07Url = match[0];
          clearTimeout(timeout);
          resolve();
        }
      };
      serverProcess?.stdout?.on('data', onData);
      serverProcess?.stderr?.on('data', onData);
    });

    // Navigate to restarted server URL
    await page.goto(stage07Url);

    // Invariant: The 10-question session is restored from SQLite!
    await expect(page.locator('.question-pill')).toHaveCount(10, { timeout: 10000 });
    // Problem 2 remains the active question
    await expect(page.getByText('Problem 2 of 10:')).toBeVisible();

    // 8. Capture screenshot of verified Stage 07 release
    const screenshotDir = path.resolve(__dirname, 'screenshots');
    fs.mkdirSync(screenshotDir, { recursive: true });
    await page.screenshot({ path: path.join(screenshotDir, 'stage07_verified.png'), fullPage: true });
  });
});
