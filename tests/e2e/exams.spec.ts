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
    const timeout = setTimeout(() => reject(new Error('Exam server startup timed out')), 10000);
    let output = '';
    server.on('error', error => { clearTimeout(timeout); reject(error); });
    server.once('exit', code => { clearTimeout(timeout); reject(new Error(`Exam server exited (${code}): ${output}`)); });
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
  await new Promise<void>(resolve => {
    server.once('exit', () => resolve());
    server.kill('SIGINT');
  });
}

test.beforeAll(async () => {
  tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'quant-exams-'));
  url = await startServer();
});

test.afterAll(async () => {
  await stopServer();
  fs.rmSync(tempDir, { recursive: true, force: true });
});

test('exam mode setup, strict aid withholding, persistence across restart, and post-submission evaluation', async ({ page, context }) => {
  // Offline loopback enforcement
  await context.route('**/*', route => {
    if (route.request().url().startsWith('http://127.0.0.1:')) {
      route.continue();
    } else {
      route.abort();
    }
  });

  page.on('pageerror', err => {
    console.error('BROWSER ERROR:', err.message, err.stack);
  });
  page.on('console', msg => {
    if (msg.type() === 'error') console.error('BROWSER CONSOLE ERROR:', msg.text());
  });

  await page.goto(url);
  await expect(page.getByText(/Loopback Server Active.*Auth/)).toBeVisible();

  // 1. Navigation via tab button
  await page.locator('#tab-exams').click();
  await expect(page.getByRole('heading', { name: 'Assessments & Exam Mode' })).toBeVisible();

  // 2. Setup assessment with 1 question, 30 minutes
  const titleInput = page.locator('#exam-title-input');
  await expect(titleInput).toBeVisible();
  await titleInput.fill('Midterm Mastery Assessment');

  await page.getByRole('button', { name: '1 Question', exact: true }).click();
  await page.getByRole('button', { name: '30 Minutes', exact: true }).click();

  // Start assessment
  await page.getByRole('button', { name: 'Begin Assessment' }).click();

  // 3. Verify Active Exam and Strict Aid Withholding
  await expect(page.getByText('Midterm Mastery Assessment')).toBeVisible();
  await expect(page.getByText('Question 1 of 1')).toBeVisible();
  await expect(page.getByLabel('Time Remaining')).toBeVisible();

  // Withholding checks: hints, solutions, explanations, and expected keys must NOT be present
  const examWorkspace = page.locator('.exams-container');
  await expect(examWorkspace.getByText('Expected Answer:')).not.toBeVisible();
  await expect(examWorkspace.getByText(/Detailed Question Review/i)).not.toBeVisible();
  await expect(examWorkspace.getByText(/Assessment Evaluation/i)).not.toBeVisible();
  await expect(examWorkspace.getByText(/Offline Hint/i)).not.toBeVisible();
  await expect(examWorkspace.getByText(/Correct!/i)).not.toBeVisible();

  // 4. Fill in answers
  // Answer choice stage (first radio found in stage 1)
  const firstRadio = page.locator('input[type="radio"]').first();
  if (await firstRadio.isVisible()) {
    await firstRadio.click();
    await expect(firstRadio).toBeChecked();
  }

  // Answer numeric stage if present
  const numericInput = page.locator('input[data-stage-id]').first();
  if (await numericInput.isVisible()) {
    await numericInput.fill('0.375');
    // Save via Save button
    const saveBtn = page.getByRole('button', { name: 'Save', exact: true });
    if (await saveBtn.isVisible()) {
      await saveBtn.click();
    }
  }

  // 5. Native Server Restart Survival (Wall-clock deadline & saved answers)
  await stopServer();
  url = await startServer();
  await page.goto(url);
  await expect(page.getByText(/Loopback Server Active.*Auth/)).toBeVisible();

  // Navigate back to Exams tab
  await page.locator('#tab-exams').click();

  // Verify the exam is restored and active
  await expect(page.getByText('Midterm Mastery Assessment')).toBeVisible();
  await expect(page.getByText('Question 1 of 1')).toBeVisible();
  await expect(page.getByLabel('Time Remaining')).toBeVisible();

  // Verify radio remains checked after restart
  if (await firstRadio.isVisible()) {
    await expect(firstRadio).toBeChecked();
  }

  // 6. Submit Exam with Confirmation Modal
  await page.getByRole('button', { name: 'Submit Exam' }).click();
  await expect(page.getByText('Finish & Submit Assessment?')).toBeVisible();
  await page.getByRole('button', { name: 'Yes, Submit Assessment' }).click();

  // 7. Verify Post-Completion Score Report
  await expect(page.getByText('Assessment Evaluation')).toBeVisible();
  await expect(page.getByText('Detailed Question Review')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Download Markdown Report' })).toBeVisible();

  // Post-completion unlocked derivations and solutions
  await expect(page.getByText('Expected Answer:').first()).toBeVisible();

  // 8. Verify Assessment in History
  await page.getByRole('button', { name: 'Take Another Assessment' }).click();
  await page.getByRole('button', { name: /History/ }).click();
  await expect(page.getByText('Midterm Mastery Assessment')).toBeVisible();
  await expect(page.getByText(/completed/i)).toBeVisible();

  // 9. Responsive layout test at 375px mobile viewport
  await page.setViewportSize({ width: 375, height: 667 });
  await expect(page.getByRole('heading', { name: 'Assessments & Exam Mode' })).toBeVisible();
});
