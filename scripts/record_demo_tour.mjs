// scripts/record_demo_tour.mjs
// Automated high-definition video walkthrough of Quant Methods Practice app using Microsoft Edge & Playwright.
// Includes realistic thinking pauses, wrong answer misconception feedback, correct answer progression,
// multi-problem navigation across all 6 curriculum modules (137-question bank),
// smooth scrolling down to reveal SVG statistical diagrams & KPI cards,
// and a full tour of all tabs (Mastery, Reference, Notes, Candidates, Worksheets, Exams, Visuals Lab, Arcade, Gallery).

import { spawn, execSync } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const root = path.resolve(__dirname, '..');

// Import Playwright Core from local node_modules
const { chromium } = await import('../web/node_modules/playwright-core/index.mjs');

const binary = path.join(root, 'bin', process.platform === 'win32' ? 'quant-practice.exe' : 'quant-practice');
const recordingsDir = path.join(root, 'recordings');
const finalMp4Path = path.join(recordingsDir, 'quant_methods_demo_tour.mp4');
const finalWebmPath = path.join(recordingsDir, 'quant_methods_demo_tour.webm');
const artifactDir = process.env.ARTIFACT_DIR || 'C:\\Users\\Dan\\.gemini\\antigravity-ide\\brain\\19fcc101-7ade-4743-ba60-156b8f97ae69';

if (!fs.existsSync(recordingsDir)) {
  fs.mkdirSync(recordingsDir, { recursive: true });
}

let serverProcess = null;
const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'quant-demo-record-'));

function startServer() {
  console.log('==> Starting Quant Methods Practice server...');
  return new Promise((resolve, reject) => {
    serverProcess = spawn(binary, ['--no-browser', '--data-dir', tempDir], {
      cwd: root,
      stdio: ['ignore', 'pipe', 'pipe']
    });

    const timeout = setTimeout(() => reject(new Error('Server startup timed out')), 15000);
    let output = '';

    const onData = (data) => {
      output += data.toString();
      const match = output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/);
      if (match) {
        clearTimeout(timeout);
        console.log(`==> Server ready at: ${match[0]}`);
        resolve(match[0]);
      }
    };

    serverProcess.stdout?.on('data', onData);
    serverProcess.stderr?.on('data', onData);
    serverProcess.on('error', (err) => { clearTimeout(timeout); reject(err); });
    serverProcess.on('exit', (code) => {
      clearTimeout(timeout);
      if (code !== 0) reject(new Error(`Server exited early with code ${code}`));
    });
  });
}

async function stopServer() {
  if (serverProcess && serverProcess.exitCode === null) {
    console.log('==> Stopping server...');
    await new Promise((resolve) => {
      serverProcess.once('exit', () => resolve());
      serverProcess.kill('SIGINT');
    });
  }
  try {
    fs.rmSync(tempDir, { recursive: true, force: true });
  } catch {}
}

// Smooth cinematic scrolling helper with ease-in-out interpolation
async function smoothScrollTo(page, targetY, durationMs = 600) {
  await page.evaluate(async ({ targetY, durationMs }) => {
    const startY = window.scrollY;
    const diff = targetY - startY;
    if (Math.abs(diff) < 2) return;
    const startTime = performance.now();

    return new Promise((resolve) => {
      function step(now) {
        const elapsed = now - startTime;
        const progress = Math.min(elapsed / durationMs, 1);
        const ease = progress < 0.5 ? 2 * progress * progress : 1 - Math.pow(-2 * progress + 2, 2) / 2;
        window.scrollTo(0, startY + diff * ease);
        if (progress < 1) {
          requestAnimationFrame(step);
        } else {
          resolve();
        }
      }
      requestAnimationFrame(step);
    });
  }, { targetY, durationMs });
  await page.waitForTimeout(100);
}

// Automatically scroll down to bring the SVG diagram and KPI metrics into prime viewport view
async function scrollToDiagram(page) {
  const svg = page.locator('.visuals-card svg, .visual-card svg, svg[role="img"], svg').first();
  if (await svg.isVisible()) {
    const box = await svg.boundingBox();
    if (box) {
      const currentScrollY = await page.evaluate(() => window.scrollY);
      // Place SVG top around 120px below viewport top, keeping SVG curve and KPI cards fully visible
      const targetY = Math.max(0, currentScrollY + box.y - 120);
      await smoothScrollTo(page, targetY, 600);
      return;
    }
  }
  await smoothScrollTo(page, 240, 600);
}

async function runDemo() {
  let bootstrapUrl = '';
  try {
    bootstrapUrl = await startServer();
  } catch (err) {
    console.error('Failed to start server:', err);
    process.exit(1);
  }

  console.log('==> Launching Microsoft Edge with video capture (1280x720 60fps)...');
  const browser = await chromium.launch({
    channel: 'msedge',
    headless: true
  });

  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    recordVideo: {
      dir: recordingsDir,
      size: { width: 1280, height: 720 }
    }
  });

  const page = await context.newPage();

  try {
    // =========================================================================
    // ACT 1: Practice Drill - Multi-Stage Problem Solving
    // =========================================================================
    console.log('==> [Act 1] Loading Practice Drill session...');
    await page.goto(bootstrapUrl);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2500); // Allow MathJax rendering

    // 1. Initial contemplation time: User reads the scenario and prompt
    console.log('    [Step 1] Contemplating Scenario: "Exactly two heads in four tosses"...');
    await page.waitForTimeout(3000); // User reads problem description and reflects

    // 2. Stage 1: Make a mistake intentionally (Misconception: confusing parameter with random variable)
    console.log('    [Step 2] Stage 1: Intentionally selecting misconception option (Option 2: chance on one toss)...');
    const optionCards = page.locator('button.option-btn');
    await optionCards.nth(1).click();
    await page.waitForTimeout(1500); // Deliberate before submitting

    console.log('    [Step 3] Submitting wrong answer to trigger misconception feedback...');
    const submitBtn = page.locator('button.btn-primary:has-text("Submit Answer")');
    if (await submitBtn.isVisible()) {
      await submitBtn.click();
    } else {
      await page.keyboard.press('Enter');
    }

    // 3. Pause for user to review misconception hint and retry state
    console.log('    [Step 4] Observing Causal Hint & "1 Retry Remaining" badge...');
    await page.waitForTimeout(3500); // Viewer reads misconception explanation

    // 4. Correct the answer: Pick Option 1 (Count of heads in four tosses)
    console.log('    [Step 5] Selecting correct option (Option 1: count of heads)...');
    await optionCards.nth(0).click();
    await page.waitForTimeout(1500);

    console.log('    [Step 6] Submitting correct answer...');
    if (await submitBtn.isVisible()) {
      await submitBtn.click();
    } else {
      await page.keyboard.press('Enter');
    }

    // 5. Pause to read mathematical explanation and solved status
    console.log('    [Step 7] Reading mathematical derivation and "Solved on Retry" feedback...');
    await smoothScrollTo(page, 100, 500);
    await page.waitForTimeout(3000);

    // 6. Advance to Stage 2
    console.log('    [Step 8] Advancing to Stage 2...');
    const nextStageBtn = page.locator('button.btn-primary:has-text("Next Stage")');
    if (await nextStageBtn.isVisible()) {
      await nextStageBtn.click();
    } else {
      await page.keyboard.press('Enter');
    }
    await smoothScrollTo(page, 0, 400);
    await page.waitForTimeout(2000);

    // 7. Stage 2: Contemplate distribution model, explore on-demand Offline Hint
    console.log('    [Step 9] Stage 2: Contemplating distribution model...');
    await page.waitForTimeout(2000);

    console.log('    [Step 10] Requesting on-demand Offline Hint...');
    const offlineHintBtn = page.locator('button:has-text("Offline Hint")');
    if (await offlineHintBtn.isVisible()) {
      await offlineHintBtn.click();
      await page.waitForTimeout(2500); // Viewer reads on-demand hint
    }

    console.log('    [Step 11] Selecting Binomial model and submitting...');
    await optionCards.nth(0).click(); // Binomial
    await page.waitForTimeout(1500);
    if (await submitBtn.isVisible()) {
      await submitBtn.click();
    } else {
      await page.keyboard.press('Enter');
    }
    await page.waitForTimeout(2500); // View explanation

    // 8. Advance to Stage 3 (Combinatorial Expression)
    console.log('    [Step 12] Advancing to Stage 3 (Combinatorial Expression)...');
    if (await nextStageBtn.isVisible()) {
      await nextStageBtn.click();
    } else {
      await page.keyboard.press('Enter');
    }
    await smoothScrollTo(page, 0, 400);
    await page.waitForTimeout(2000);

    console.log('    [Step 13] Selecting combinatorial expression and submitting...');
    const stage3Options = page.locator('button.option-btn');
    if (await stage3Options.count() > 0) {
      await stage3Options.nth(0).click();
      await page.waitForTimeout(1500);
      if (await submitBtn.isVisible()) {
        await submitBtn.click();
      } else {
        await page.keyboard.press('Enter');
      }
      await page.waitForTimeout(2500);
    }

    // =========================================================================
    // ACT 2: Full Question Bank (137 Questions) & Module Filtering
    // =========================================================================
    console.log('==> [Act 2] Showcasing Full Question Bank & Module Filtering...');

    // Open Settings Modal to view all 6 modules
    console.log('    Opening Session Settings to demonstrate Module 1 - 6 selection...');
    const settingsBtn = page.locator('button:has-text("Settings")');
    if (await settingsBtn.isVisible()) {
      await settingsBtn.click();
      await page.waitForTimeout(3000); // Viewer inspects all 6 module checkboxes & question count options

      // Close Settings Modal
      console.log('    Closing settings modal...');
      await page.keyboard.press('Escape');
      await page.waitForTimeout(1000);
    }

    // Start a fresh randomized drill using the 🎲 New Drill button
    console.log('    Starting fresh randomized drill session from active 137-question bank...');
    const newDrillBtn = page.locator('#btn-header-new-drill');
    if (await newDrillBtn.isVisible()) {
      await newDrillBtn.click();
      await page.waitForTimeout(3000); // Viewer sees newly randomized question scenario
    }

    // Contemplate and answer new randomized question Stage 1
    console.log('    Answering fresh question from bank...');
    const freshOptions = page.locator('button.option-btn');
    if (await freshOptions.count() > 0) {
      await freshOptions.nth(0).click();
      await page.waitForTimeout(1500);
      const freshSubmit = page.locator('button.btn-primary:has-text("Submit Answer")');
      if (await freshSubmit.isVisible()) {
        await freshSubmit.click();
        await page.waitForTimeout(2500);
      }
    }

    // Navigate to Question 2 in this randomized session
    console.log('    Navigating to next question in the drill...');
    const nextProbBtn = page.locator('button:has-text("Next Problem")');
    if (await nextProbBtn.isVisible() && await nextProbBtn.isEnabled()) {
      await nextProbBtn.click();
    } else {
      const problemPill = page.locator('.question-pill').nth(1);
      if (await problemPill.isVisible()) {
        await problemPill.click();
      }
    }
    await page.waitForTimeout(2500); // Contemplate new problem scenario

    // Answer Problem 2 Stage 1
    console.log('    Answering Problem 2...');
    const prob2Options = page.locator('button.option-btn');
    if (await prob2Options.count() > 0) {
      await prob2Options.nth(0).click();
      await page.waitForTimeout(1500);
      const prob2Submit = page.locator('button.btn-primary:has-text("Submit Answer")');
      if (await prob2Submit.isVisible()) {
        await prob2Submit.click();
        await page.waitForTimeout(2500);
      }
    }

    // =========================================================================
    // ACT 3: Comprehensive Tour of All Tabs
    // =========================================================================

    // Tab 1: Concept Mastery & Transfer
    console.log('==> [Act 3.1] Touring Concept Mastery & Transfer Tab...');
    const masteryTab = page.locator('#tab-mastery');
    if (await masteryTab.isVisible()) {
      await masteryTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(3000); // View mastery profile, score, and badges

      // Filter / search concepts
      const masterySearch = page.locator('input[placeholder*="Search"]');
      if (await masterySearch.isVisible()) {
        await masterySearch.fill('regression');
        await page.waitForTimeout(2000);
        await masterySearch.fill('');
        await page.waitForTimeout(1000);
      }
    }

    // Tab 2: Reference Library
    console.log('==> [Act 3.2] Touring Reference Library Tab...');
    const refTab = page.locator('#tab-reference');
    if (await refTab.isVisible()) {
      await refTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(2500); // View reference entries and axioms

      const refSearch = page.locator('input[placeholder*="Search"]');
      if (await refSearch.isVisible()) {
        await refSearch.fill('Hypothesis');
        await page.waitForTimeout(2000);
        await refSearch.fill('');
        await page.waitForTimeout(1000);
      }
    }

    // Tab 3: Saved Notes & Explanations
    console.log('==> [Act 3.3] Touring Saved Notes & Explanations Tab...');
    const notesTab = page.locator('#tab-notes');
    if (await notesTab.isVisible()) {
      await notesTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(2500); // View SQLite saved note repository
    }

    // Tab 4: Question Candidates Review
    console.log('==> [Act 3.4] Touring Question Candidates Review Tab...');
    const candidatesTab = page.locator('#tab-candidates');
    if (await candidatesTab.isVisible()) {
      await candidatesTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(2500); // View candidates, generator, and review controls
    }

    // Tab 5: Worksheets & Dataset Case Explorer
    console.log('==> [Act 3.5] Touring Worksheets Tab...');
    const worksheetsTab = page.locator('#tab-worksheets');
    if (await worksheetsTab.isVisible()) {
      await worksheetsTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(2000);
      console.log('    Scrolling down to inspect worksheet solution form and dataset case...');
      await smoothScrollTo(page, 220, 600);
      await page.waitForTimeout(2500);
      await smoothScrollTo(page, 0, 400);
    }

    // Tab 6: Exams Mode Assessment
    console.log('==> [Act 3.6] Touring Exams Mode Tab...');
    const examsTab = page.locator('#tab-exams');
    if (await examsTab.isVisible()) {
      await examsTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(3000); // View exam presets, module dropdown, timer, strict aids policy
    }

    // Tab 7: Mathematical Visuals Lab (With full scrolling down to diagrams)
    console.log('==> [Act 3.7] Touring Mathematical Visuals Lab Tab...');
    const visualsTab = page.locator('#tab-visuals');
    if (await visualsTab.isVisible()) {
      await visualsTab.click();
      await page.waitForTimeout(800);

      // Normal Distribution explorer
      console.log('    [Visuals] Configuring Normal Distribution parameters (Upper Tail)...');
      const tailSelect = page.locator('#norm-tail-select');
      if (await tailSelect.isVisible()) {
        await tailSelect.selectOption('upper');
        await page.waitForTimeout(600);
      }

      // CRITICAL: Scroll down smoothly when getting to this page so the full bell curve, shaded tail & KPI cards are in view
      console.log('    [Visuals] Scrolling down to view full Normal Distribution curve and spreadsheet formulas...');
      await scrollToDiagram(page);
      await page.waitForTimeout(3500); // Let viewer inspect the full SVG bell curve, shaded tail & Excel formula

      // Uniform tab
      console.log('    [Visuals] Switching to Uniform Distribution & scrolling to diagram...');
      await smoothScrollTo(page, 60, 400); // Reveal sub-tab bar
      const uniformTab = page.locator('#tab-visual-uniform');
      if (await uniformTab.isVisible()) {
        await uniformTab.click();
        await scrollToDiagram(page);
        await page.waitForTimeout(2500); // View uniform density rectangle and area
      }

      // Exponential tab
      console.log('    [Visuals] Switching to Exponential Distribution & scrolling to diagram...');
      await smoothScrollTo(page, 60, 400);
      const expTab = page.locator('#tab-visual-exponential');
      if (await expTab.isVisible()) {
        await expTab.click();
        await scrollToDiagram(page);
        await page.waitForTimeout(2500); // View decaying exponential density & memoryless property
      }

      // Discrete tab
      console.log('    [Visuals] Switching to Discrete PMF/CDF explorer & scrolling to diagram...');
      await smoothScrollTo(page, 60, 400);
      const discreteTab = page.locator('#tab-visual-discrete');
      if (await discreteTab.isVisible()) {
        await discreteTab.click();
        // Switch to Poisson
        const familySelect = page.locator('#disc-family-select');
        if (await familySelect.isVisible()) {
          await familySelect.selectOption('poisson');
          await page.waitForTimeout(500);
        }
        await scrollToDiagram(page);
        await page.waitForTimeout(2500); // View discrete PMF/CDF bars
      }

      // Sampling / CLT tab
      console.log('    [Visuals] Exploring Central Limit Theorem Sampling Simulator & scrolling to diagram...');
      await smoothScrollTo(page, 60, 400);
      const samplingTab = page.locator('#tab-visual-sampling');
      if (await samplingTab.isVisible()) {
        await samplingTab.click();
        await scrollToDiagram(page);
        await page.waitForTimeout(2500); // View CLT sampling distribution simulation
      }

      // Confidence Intervals tab
      console.log('    [Visuals] Exploring Confidence Intervals & scrolling to diagram...');
      await smoothScrollTo(page, 60, 400);
      const ciTab = page.locator('#tab-visual-ci');
      if (await ciTab.isVisible()) {
        await ciTab.click();
        await scrollToDiagram(page);
        await page.waitForTimeout(2500); // View confidence intervals coverage rate
      }

      // Hypothesis Testing tab
      console.log('    [Visuals] Exploring Hypothesis Testing & scrolling to diagram...');
      await smoothScrollTo(page, 60, 400);
      const htTab = page.locator('#tab-visual-hypothesis');
      if (await htTab.isVisible()) {
        await htTab.click();
        await scrollToDiagram(page);
        await page.waitForTimeout(2500); // View hypothesis rejection region
      }

      // Regression tab
      console.log('    [Visuals] Exploring Bivariate Linear Regression & scrolling to diagram...');
      await smoothScrollTo(page, 60, 400);
      const regTab = page.locator('#tab-visual-regression');
      if (await regTab.isVisible()) {
        await regTab.click();
        await scrollToDiagram(page);
        await page.waitForTimeout(2500); // View scatter plot, OLS line & residual bars
      }

      // Reset scroll position before exiting tab
      await smoothScrollTo(page, 0, 400);
    }

    // Tab 8: Space Arcade (Quant Flight)
    console.log('==> [Act 3.8] Touring Quant Flight Space Arcade Tab...');
    const arcadeTab = page.locator('#tab-arcade');
    if (await arcadeTab.isVisible()) {
      await arcadeTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(2500); // Watch retro vector autopilot attract mode

      console.log('    Engaging manual flight takeover...');
      await page.keyboard.press('Space');
      await page.waitForTimeout(600);

      // Maneuver ship and blast statistical hazards
      for (let i = 0; i < 6; i++) {
        await page.keyboard.press('ArrowUp');
        await page.keyboard.press('Space'); // Fire laser
        await page.waitForTimeout(200);
        await page.keyboard.press('ArrowRight');
        await page.keyboard.press('f'); // Wing cannon
        await page.waitForTimeout(250);
      }
      await page.waitForTimeout(1500);

      const skipBtn = page.locator('#btn-skip-arcade');
      if (await skipBtn.isVisible()) {
        console.log('    Exiting arcade back to drills...');
        await skipBtn.click();
        await page.waitForTimeout(1500);
      }
    }

    // Tab 9: Formula Gallery
    console.log('==> [Act 3.9] Touring Mathematical Formula Gallery Tab...');
    const galleryTab = page.locator('#tab-gallery');
    if (await galleryTab.isVisible()) {
      await galleryTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(2000);
      console.log('    Scrolling through mathematical formulas in gallery...');
      await smoothScrollTo(page, 240, 700);
      await page.waitForTimeout(2500);
      await smoothScrollTo(page, 0, 400);
    }

    // Tab 10: Smooth Return to Practice Drill
    console.log('==> [Act 3.10] Returning to active Practice Drill...');
    const drillTab = page.locator('#tab-drill');
    if (await drillTab.isVisible()) {
      await drillTab.click();
      await smoothScrollTo(page, 0, 300);
      await page.waitForTimeout(2500); // View drill state intact and ready
    }

    console.log('==> Walkthrough sequence completed successfully.');
  } catch (err) {
    console.error('Error during demo walkthrough:', err);
  } finally {
    const video = page.video();
    await page.close();
    await context.close();
    await browser.close();
    await stopServer();

    if (video) {
      const generatedPath = await video.path();
      console.log(`==> Raw video saved at: ${generatedPath}`);
      if (fs.existsSync(generatedPath)) {
        if (fs.existsSync(finalWebmPath)) {
          fs.unlinkSync(finalWebmPath);
        }
        fs.renameSync(generatedPath, finalWebmPath);
        console.log(`==> WebM demo recording saved: ${finalWebmPath}`);

        // Transcode to MP4 (H.264 / 720p 60fps) using ffmpeg for universal playback
        console.log('==> Transcoding to MP4 (H.264 / 720p)...');
        try {
          execSync(`ffmpeg -y -i "${finalWebmPath}" -c:v libx264 -preset fast -crf 22 -pix_fmt yuv420p -movflags +faststart "${finalMp4Path}"`, { stdio: 'inherit' });
          const mp4Stats = fs.statSync(finalMp4Path);
          console.log(`==> MP4 demo recording finalized: ${finalMp4Path} (${(mp4Stats.size / (1024 * 1024)).toFixed(2)} MB)`);
        } catch (err) {
          console.warn('ffmpeg transcode failed, copying webm as mp4 fallback:', err.message);
          fs.copyFileSync(finalWebmPath, finalMp4Path);
        }

        // Copy to artifact directory if available
        const currentArtifactDir = 'C:\\Users\\Dan\\.gemini\\antigravity-ide\\brain\\99bdccb8-3294-4c8c-b2a9-589e73714a54';
        const targetDirs = [artifactDir, currentArtifactDir].filter((d) => d && fs.existsSync(d));
        for (const dir of targetDirs) {
          if (fs.existsSync(finalMp4Path)) {
            fs.copyFileSync(finalMp4Path, path.join(dir, 'quant_methods_demo_tour.mp4'));
          }
          if (fs.existsSync(finalWebmPath)) {
            fs.copyFileSync(finalWebmPath, path.join(dir, 'quant_methods_demo_tour.webm'));
          }
          console.log(`==> Copied video assets to artifact directory: ${dir}`);
        }
      }
    }
  }
}

runDemo().catch((err) => {
  console.error('Fatal demo runner error:', err);
  process.exit(1);
});

