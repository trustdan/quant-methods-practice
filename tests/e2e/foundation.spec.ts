import { test, expect } from '@playwright/test';
import { spawn, ChildProcess } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';

const BIN_PATH = path.resolve(__dirname, process.platform === 'win32' ? '../../bin/quant-practice.exe' : '../../bin/quant-practice');
const PORT = 8995;

let serverProcess: ChildProcess | null = null;
let bootstrapUrl = '';
let restartedUrl = '';
let stage07Url = '';
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
  serverProcess = spawn(BIN_PATH, ['--port', String(PORT), '--no-browser', '--data-dir', path.dirname(tempDbPath), '--db', tempDbPath], {
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
    serverProcess = spawn(BIN_PATH, ['--port', String(RESTART_PORT), '--no-browser', '--data-dir', path.dirname(tempDbPath), '--db', tempDbPath], {
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
    await numericInput.blur();
    await page.locator('.app-logo').click();
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
    serverProcess = spawn(BIN_PATH, ['--port', String(STAGE07_PORT), '--no-browser', '--data-dir', path.dirname(tempDbPath), '--db', tempDbPath], {
      cwd: path.resolve(__dirname, '../..'),
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    stage07Url = '';
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

  test('Stage 08: renders concept mastery view, scaffold levels, and keyboard shortcuts', async ({ page }) => {
    // Navigate to active server URL
    await page.goto(stage07Url || restartedUrl || bootstrapUrl);

    // 1. Verify practice drill is visible with scaffold badge
    await expect(page.locator('.question-strip')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#scaffold-indicator')).toBeVisible();

    // 2. Open Mastery View via keyboard shortcut 's'
    await page.keyboard.press('s');

    // 3. Verify Mastery Dashboard elements
    const masteryView = page.locator('#mastery-view');
    await expect(masteryView).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#mastery-stats')).toBeVisible();
    await expect(page.getByText('Concept Evidence & Transfer')).toBeVisible();
    await expect(page.getByText(/Policy v1/)).toBeVisible();
    await expect(page.getByText('Overall Retention Score')).toBeVisible();

    // 4. Verify concept cards container or empty state is rendered
    await expect(page.locator('#concept-cards-container, .empty-state').first()).toBeVisible();

    // 5. Test filter pills
    const learningPill = page.locator('#filter-learning');
    await learningPill.click();
    await expect(learningPill).toHaveClass(/active/);

    const allPill = page.locator('#filter-all');
    await allPill.click();
    await expect(allPill).toHaveClass(/active/);

    // 6. Test search filter
    const searchInput = page.locator('#mastery-search-input');
    await searchInput.fill('random');
    const cardCount = await page.locator('.concept-card').count();
    if (cardCount > 0) {
      await expect(page.locator('.concept-card').first()).toBeVisible();
    }

    // Clear search filter
    await searchInput.fill('');

    // 7. Test closing mastery view with Escape
    await page.keyboard.press('Escape');
    await expect(masteryView).not.toBeVisible();
    await expect(page.locator('.question-strip')).toBeVisible();

    // 8. Re-open via nav tab click and close via button
    const masteryTab = page.locator('#tab-mastery');
    await masteryTab.click();
    await expect(masteryView).toBeVisible();

    const closeBtn = page.locator('#btn-close-mastery');
    await closeBtn.click();
    await expect(masteryView).not.toBeVisible();

    // 9. Capture screenshot of verified Stage 08 release
    const screenshotDir = path.resolve(__dirname, 'screenshots');
    fs.mkdirSync(screenshotDir, { recursive: true });
    await page.screenshot({ path: path.join(screenshotDir, 'stage08_verified.png'), fullPage: true });
  });

  test('Stage 09: read-only AI tutor streaming, cancellation, leave-intent protection, and note library with export', async ({ page }) => {
    // Navigate to active server URL
    await page.goto(stage07Url || restartedUrl || bootstrapUrl);

    // 1. Practice drill is visible; open AI Tutor panel via 'n' keyboard shortcut
    await expect(page.locator('.question-strip')).toBeVisible({ timeout: 10000 });
    await page.keyboard.press('n');

    const tutorModal = page.locator('#tutor-panel');
    await expect(tutorModal).toBeVisible({ timeout: 10000 });

    // 2. Verify strict advisory banner and offline indicator
    await expect(page.locator('#tutor-advisory-banner')).toBeVisible();
    await expect(page.getByText('Advisory AI explanation:')).toBeVisible();
    await expect(page.getByText('Offline Reviewed')).toBeVisible();

    // 3. Request a Causal Hint via #btn-tutor-hint
    const hintBtn = page.locator('#btn-tutor-hint');
    await hintBtn.click();

    // Verify streamed response appears in tutor-response-area
    const responseArea = page.locator('#tutor-response-area');
    await expect(responseArea).toContainText('Causal Hint', { timeout: 10000 });

    // Math formula SVG rendering verification in tutor area
    const tutorMathSvg = responseArea.locator('mjx-container[jax="SVG"]');
    await expect(tutorMathSvg.first()).toBeVisible({ timeout: 10000 });

    // 4. Test dirty explanation leave-intent protection
    // Try to close the tutor while an unsaved explanation is displayed
    const closeBtn = page.locator('#btn-tutor-close');
    await closeBtn.click();

    const leaveModal = page.locator('#tutor-leave-modal');
    await expect(leaveModal).toBeVisible();
    await expect(leaveModal).toContainText('Save Explanation Before Leaving?');

    // Test "Stay" action (Esc or #btn-tutor-leave-stay)
    await page.locator('#btn-tutor-leave-stay').click();
    await expect(leaveModal).not.toBeVisible();
    await expect(tutorModal).toBeVisible();

    // 5. Save explanation to personal notes library
    const saveBtn = page.locator('#btn-tutor-save');
    await saveBtn.click();
    await expect(page.getByText('Saved to Personal Notes Library')).toBeVisible({ timeout: 10000 });

    // Close tutor cleanly now that note is saved (should not trigger leave-intent modal)
    await closeBtn.click();
    await expect(tutorModal).not.toBeVisible();

    // 6. Test cancellation during streaming
    // Re-open tutor
    await page.keyboard.press('n');
    await expect(tutorModal).toBeVisible();

    // Request Step-by-Step Solution
    const explainBtn = page.locator('#btn-tutor-explain');
    await explainBtn.click();

    // While or immediately after starting, test Stop/cancel action
    const cancelBtn = page.locator('#btn-tutor-cancel');
    if (await cancelBtn.isVisible()) {
      await cancelBtn.click();
    }

    // Close tutor (discard if leave modal shows)
    await closeBtn.click();
    if (await leaveModal.isVisible()) {
      await page.locator('#btn-tutor-leave-discard').click();
    }
    await expect(tutorModal).not.toBeVisible();

    // 7. Open Saved Notes Library via #tab-notes
    const notesTab = page.locator('#tab-notes');
    await notesTab.click();

    const noteLibraryView = page.locator('#note-library-view');
    await expect(noteLibraryView).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Saved Explanations & Notes Library')).toBeVisible();

    // 8. Verify the saved note appears in the note library list
    const noteItems = page.locator('.note-list-item');
    await expect(noteItems.first()).toBeVisible({ timeout: 10000 });

    // Check detail pane title and rendered math
    await expect(page.locator('#selected-note-title')).toBeVisible();
    const detailMath = page.locator('#note-detail-pane mjx-container[jax="SVG"]');
    await expect(detailMath.first()).toBeVisible();

    // 9. Test search filter in notes
    const notesSearchInput = page.locator('#notes-search-input');
    await notesSearchInput.fill('defective');
    await expect(noteItems.first()).toBeVisible();
    await notesSearchInput.fill('');

    // 10. Test note export action
    const exportBtn = page.locator('#btn-note-export');
    await expect(exportBtn).toBeVisible();

    // Verify clicking export triggers download or runs without error
    const downloadPromise = page.waitForEvent('download', { timeout: 3000 }).catch(() => null);
    await exportBtn.click();
    const download = await downloadPromise;
    if (download) {
      expect(download.suggestedFilename()).toMatch(/\.md$/);
    }

    // 11. Return to practice drill
    const closeNotesBtn = page.locator('#btn-notes-close');
    await closeNotesBtn.click();
    await expect(noteLibraryView).not.toBeVisible();
    await expect(page.locator('.question-strip')).toBeVisible();

    // 12. Capture screenshot of verified Stage 09 release
    const screenshotDir = path.resolve(__dirname, 'screenshots');
    fs.mkdirSync(screenshotDir, { recursive: true });
    await page.screenshot({ path: path.join(screenshotDir, 'stage09_verified.png'), fullPage: true });
  });

  test('verifies Stage 10 provider settings, vault credential masking, and AI tutor provider badge', async ({ context, page }) => {
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

    await page.goto(stage07Url || restartedUrl || bootstrapUrl);
    await expect(page).toHaveTitle('Quant Methods Practice');

    // 1. Open Settings modal via 't' key shortcut
    await page.keyboard.press('t');
    const settingsModal = page.locator('.modal-backdrop');
    await expect(settingsModal).toBeVisible();

    // 2. Verify Tab Navigation: Session Settings vs AI Providers
    const sessionTab = page.locator('#tab-session-settings');
    const providerTab = page.locator('#tab-provider-settings');
    await expect(sessionTab).toBeVisible();
    await expect(providerTab).toBeVisible();

    // 3. Switch to AI Providers & Credentials Tab
    await providerTab.click();
    await expect(page.locator('.provider-settings-container')).toBeVisible({ timeout: 5000 });

    // Verify provider route selector tabs
    await expect(page.getByRole('button', { name: /Offline Reviewed/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /Google Gemini/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /Anthropic Claude/ })).toBeVisible();
    await expect(page.getByRole('button', { name: /OpenAI/ })).toBeVisible();

    // Verify Offline Reviewed is active by default
    await expect(page.getByText('ACTIVE').first()).toBeVisible();

    // 4. Select Google Gemini tab
    await page.getByRole('button', { name: 'Google Gemini' }).click();
    await expect(page.getByText('Google Gemini API Key (GEMINI_API_KEY)')).toBeVisible();

    // 5. Store an API key into backend vault
    const keyInput = page.locator('input[type="password"]');
    await keyInput.fill('AIzaSyE2ETestSecretKey9988');

    const saveKeyBtn = page.getByRole('button', { name: 'Save Key' });
    await saveKeyBtn.click();

    // Verify success confirmation and masked key display
    await expect(page.getByText('API key stored securely in backend vault.')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/AIzaSy\.\.\.9988/)).toBeVisible();

    // 6. Set Google Gemini as Active Provider
    const setActiveBtn = page.getByRole('button', { name: 'Set as Active' });
    await setActiveBtn.click();
    await expect(page.getByText('Currently Active')).toBeVisible({ timeout: 5000 });

    // 7. Close Settings modal via Escape
    await page.keyboard.press('Escape');
    await expect(settingsModal).not.toBeVisible();

    // 8. Open AI Tutor panel (press 'n')
    await page.keyboard.press('n');
    const tutorModal = page.locator('#tutor-panel');
    await expect(tutorModal).toBeVisible({ timeout: 5000 });

    // 9. Verify header badge reflects active Google Gemini provider
    const providerBadge = page.locator('#tutor-provider-badge');
    await expect(providerBadge).toBeVisible();
    await expect(providerBadge).toContainText('Google Gemini');

    // 10. Close AI Tutor panel
    const closeTutorBtn = page.locator('#btn-tutor-close');
    await closeTutorBtn.click();
    await expect(tutorModal).not.toBeVisible();

    // 11. Reopen Settings and revert to Offline Reviewed (ensuring offline default)
    await page.keyboard.press('t');
    await expect(settingsModal).toBeVisible();
    await providerTab.click();
    await page.getByRole('button', { name: /Offline Reviewed/ }).click();
    const setOfflineActive = page.getByRole('button', { name: 'Set as Active' });
    if (await setOfflineActive.isVisible()) {
      await setOfflineActive.click();
    }
    await page.keyboard.press('Escape');
    await expect(settingsModal).not.toBeVisible();

    // 12. Capture screenshot of verified Stage 10 release
    const screenshotDir = path.resolve(__dirname, 'screenshots');
    fs.mkdirSync(screenshotDir, { recursive: true });
    await page.screenshot({ path: path.join(screenshotDir, 'stage10_verified.png'), fullPage: true });
  });

  test('Stage 11: ChatGPT plan route shows plan billing, starts PKCE sign-in on loopback, and cancels cleanly', async ({ context, page }) => {
    const externalRequests: string[] = [];
    await context.route('**/*', (route) => {
      const url = route.request().url();
      if (!url.startsWith('http://127.0.0.1:')) {
        externalRequests.push(url);
        route.abort();
      } else {
        route.continue();
      }
    });
    // Record the sign-in URL instead of navigating a real popup to OpenAI.
    await context.addInitScript(() => {
      (window as unknown as { __openedUrls: string[] }).__openedUrls = [];
      window.open = (url?: string | URL) => {
        (window as unknown as { __openedUrls: string[] }).__openedUrls.push(String(url));
        return null;
      };
    });

    await page.goto(stage07Url || restartedUrl || bootstrapUrl);
    await expect(page).toHaveTitle('Quant Methods Practice');

    await page.keyboard.press('t');
    const settingsModal = page.locator('.modal-backdrop');
    await expect(settingsModal).toBeVisible();
    await page.locator('#tab-provider-settings').click();

    // Plan and API-key routes are separate tabs with distinct billing labels.
    await page.getByRole('button', { name: 'OpenAI API' }).click();
    await expect(page.locator('#provider-billing-label')).toHaveText('Billing: API usage billing');
    await page.getByRole('button', { name: 'ChatGPT Plan' }).click();
    await expect(page.locator('#provider-billing-label')).toHaveText('Billing: ChatGPT plan usage');
    await expect(page.getByText('Not signed in.')).toBeVisible();
    await expect(page.locator('#chatgpt-plan-panel input[type="password"]')).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Set as Active' })).toBeDisabled();

    await page.getByRole('button', { name: 'Sign in with ChatGPT' }).click();
    await expect(page.getByText(/Waiting for you to finish signing in/)).toBeVisible({ timeout: 5000 });

    const opened = await page.evaluate(() => (window as unknown as { __openedUrls: string[] }).__openedUrls);
    expect(opened).toHaveLength(1);
    const authorize = new URL(opened[0]);
    expect(authorize.origin + authorize.pathname).toBe('https://auth.openai.com/api/accounts/authorize');
    expect(authorize.searchParams.get('client_id')).toBe('dynamic_agent_client');
    expect(authorize.searchParams.get('code_challenge_method')).toBe('S256');
    expect(authorize.searchParams.get('redirect_uri')).toMatch(/^http:\/\/127\.0\.0\.1:\d+\/auth\/callback$/);

    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('button', { name: 'Sign in with ChatGPT' })).toBeVisible();

    // The route never became active and nothing left the machine.
    await expect(page.getByText('ACTIVE').first()).toBeVisible();
    expect(externalRequests).toEqual([]);

    const screenshotDir = path.resolve(__dirname, 'screenshots');
    fs.mkdirSync(screenshotDir, { recursive: true });
    await page.screenshot({ path: path.join(screenshotDir, 'stage11_verified.png'), fullPage: true });

    await page.keyboard.press('Escape');
    await expect(settingsModal).not.toBeVisible();
  });

  test('renders provider markdown: multi-line display math and headings are typeset', async ({ context, page }) => {
    await context.route('**/*', (route) => {
      const url = route.request().url();
      if (!url.startsWith('http://127.0.0.1:')) {
        route.abort();
      } else {
        route.continue();
      }
    });

    // Replay a live provider answer shape (multi-line $$ block, ### heading) through the real panel.
    const text = [
      '### Heads counter',
      'HHTT has two heads, so $X = 2$.',
      'So we define:',
      '',
      '$$',
      'X=\\text{the number of heads in the four tosses}.',
      '$$',
      '',
      '#### Your turn',
      '1. If the tosses are HTTT, what value would $X$ have?',
    ].join('\n');
    await page.route('**/api/tutor/requests/*/events', (route) =>
      route.fulfill({
        status: 200,
        headers: { 'Content-Type': 'text/event-stream' },
        body: `data: ${JSON.stringify({ type: 'text_delta', delta: text })}\n\n` +
          `data: ${JSON.stringify({ type: 'complete', text })}\n\n`,
      }),
    );

    await page.goto(stage07Url || restartedUrl || bootstrapUrl);
    await expect(page.locator('.question-strip')).toBeVisible({ timeout: 10000 });
    await page.keyboard.press('n');
    await expect(page.locator('#tutor-panel')).toBeVisible({ timeout: 10000 });
    await page.locator('#btn-tutor-hint').click();

    const responseArea = page.locator('#tutor-response-area');
    await expect(responseArea.getByRole('heading', { name: 'Heads counter' })).toBeVisible({ timeout: 10000 });
    await expect(responseArea.getByRole('heading', { name: 'Your turn' })).toBeVisible();
    await expect(responseArea.locator('.display-math mjx-container[jax="SVG"][display="true"]')).toHaveCount(1, { timeout: 10000 });
    await expect(responseArea.locator('ol > li')).toHaveCount(1);
    await expect(responseArea).not.toContainText('$$');
    await expect(responseArea).not.toContainText('###');
    await expect(responseArea).not.toContainText('\\text{');
  });
});
