import { test, expect, Page } from '@playwright/test';
import { spawn, ChildProcess } from 'child_process';
import path from 'path';
import fs from 'fs';
import os from 'os';

const root = path.resolve(__dirname, '../..');
const binary = path.join(root, 'bin', process.platform === 'win32' ? 'quant-practice.exe' : 'quant-practice');
let server: ChildProcess;
let tempDir: string;
let url: string;
async function start(): Promise<string> {
  server = spawn(binary, ['--no-browser', '--data-dir', tempDir], { cwd: root, stdio: ['ignore', 'pipe', 'pipe'] });
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => reject(new Error('Worksheet server startup timed out')), 10000);
    let output = '';
    server.on('error', error => { clearTimeout(timeout); reject(error); });
    server.once('exit', code => { clearTimeout(timeout); reject(new Error(`Worksheet server exited (${code}): ${output}`)); });
    const onData = (data: Buffer) => {
      output += data.toString();
      const match = output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/);
      if (match) { clearTimeout(timeout); resolve(match[0]); }
    };
    server.stdout?.on('data', onData); server.stderr?.on('data', onData);
  });
}
async function stop() {
  if (server.exitCode !== null) return;
  await new Promise<void>(resolve => { server.once('exit', () => resolve()); server.kill('SIGINT'); });
}
test.beforeAll(async () => { tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'quant-worksheets-')); url = await start(); });
test.afterAll(async () => { await stop(); fs.rmSync(tempDir, { recursive: true, force: true }); });

const reference = JSON.parse(fs.readFileSync(path.join(root, 'curriculum/approved/binomial-fair-coin-exactly-two.json'), 'utf8'));
async function fillReference(page: Page, numeric: string) {
  const article = page.getByRole('article', { name: reference.title, exact: true });
  for (const stage of reference.stages) {
    if (stage.kind === 'numeric') await article.getByLabel(`Numeric answer (${stage.id.replaceAll('_', ' ')})`).fill(numeric);
    else {
      const option = stage.options.find((item: { id: string }) => item.id === stage.expected_answer.option_id);
      const field = article.locator('fieldset').filter({ has: page.locator('legend', { hasText: stage.id.replaceAll('_', ' ') }) });
      // Use the input value, independent of math-rendered accessible text.
      await field.locator(`input[value="${option.id}"]`).check();
    }
  }
}

test('full-form drafts, atomic validation, retry and solution review survive native restart offline', async ({ page, context }) => {
  await context.route('**/*', route => route.request().url().startsWith('http://127.0.0.1:') ? route.continue() : route.abort());
  await page.goto(url);
  await expect(page.getByText(/Loopback Server Active.*Auth/)).toBeVisible();
  await page.keyboard.press('Shift+J');
  await expect(page.getByRole('heading', { name: 'Full solutions & dataset cases' })).toBeVisible();
  await page.getByRole('button', { name: 'Create worksheet', exact: true }).click();
  const workspace = page.getByRole('region', { name: 'Worksheets', exact: true });
  const numeric = workspace.getByLabel('Numeric answer (calculate probability)');
  await fillReference(page, '1/0');
  await numeric.press('End'); await numeric.type('jkl');
  await expect(numeric).toHaveValue('1/0jkl'); // native editable keys preserved
  await numeric.fill('1/0');
  await page.getByRole('button', { name: 'Save worksheet draft', exact: true }).click();
  await expect(page.getByText('Draft saved on this device.')).toBeVisible();
  await stop(); url = await start(); await page.goto(url);
  await page.getByRole('button', { name: 'Worksheets', exact: true }).click();
  await expect(numeric).toHaveValue('1/0');
  await page.getByRole('button', { name: 'Submit full solution', exact: true }).click();
  await expect(workspace.getByRole('alert')).toBeVisible();
  expect((await (await page.request.get(new URL('/api/worksheets', url).href)).json())[0].items.every((item: { attempts: unknown[] | null }) => !item.attempts?.length)).toBeTruthy();
  await numeric.fill('0');
  await page.getByRole('button', { name: 'Submit full solution', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Submit retry', exact: true })).toBeEnabled();
  await expect(numeric).toBeEnabled();
  await numeric.fill('37.5%');
  await page.getByRole('button', { name: 'Submit retry', exact: true }).click();
  await expect(page.getByText('Worksheet completed. Your submission and review are saved.')).toBeVisible();
  await expect(numeric).toBeDisabled();
  await stop(); url = await start(); await page.goto(url);
  await page.getByRole('button', { name: 'Worksheets', exact: true }).click();
  await expect(numeric).toHaveValue('37.5%'); await expect(numeric).toBeDisabled();
  const downloadPromise = page.waitForEvent('download'); await page.getByRole('link', { name: 'Download worksheet', exact: true }).click();
  const exported = fs.readFileSync((await (await downloadPromise).path())!, 'utf8');
  expect(exported).toContain('full_solution_form'); expect(exported).toContain('hint'); expect(exported).toContain('37.5%');
  await page.setViewportSize({ width: 375, height: 812 });
  await expect(workspace.locator('mjx-container').first()).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 2)).toBeFalsy();
});

test('CSV case requires semantic review, saves data and distinguishes observed and model probabilities', async ({ page, context }) => {
  // The previous test may have consumed the startup token; a new launch gives this browser its own session.
  await stop(); url = await start();
  await context.route('**/*', route => route.request().url().startsWith('http://127.0.0.1:') ? route.continue() : route.abort());
  await page.goto(url); await expect(page.getByText(/Loopback Server Active.*Auth/)).toBeVisible(); await page.keyboard.press('Shift+F');
  await page.getByLabel('CSV data', { exact: true }).fill('experiment_id,successes\na,=2\n');
  await page.getByRole('button', { name: 'Preview CSV', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('integer from 0 to 4');
  await page.getByLabel('Choose CSV file', { exact: true }).setInputFiles({ name: 'counts.csv', mimeType: 'text/csv', buffer: Buffer.from('experiment_id,successes\na,2\nb,1\nc,4\nd,2\ne,0\n') });
  await expect(page.getByLabel('CSV data', { exact: true })).toHaveValue('experiment_id,successes\na,2\nb,1\nc,4\nd,2\ne,0\n');
  await page.getByRole('button', { name: 'Preview CSV', exact: true }).click();
  await expect(page.getByText('5 experiments accepted. Showing the first 5 rows.')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Review case wording, conditions and keys' })).toBeVisible();
  await page.getByLabel('Dataset reviewer', { exact: true }).fill('E2E test fixture');
  await page.getByLabel('Data source and row meaning', { exact: true }).fill('Synthetic experiment counts; each row records heads in four tosses.');
  const create = page.getByRole('button', { name: 'Create worksheet', exact: true });
  await expect(create).toBeDisabled(); await page.getByRole('checkbox').check(); await create.click();
  await expect(page.getByText(/Saved status: draft. Supported reviewed-case/)).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Review case wording, conditions and keys' })).not.toBeVisible();
  const observed = page.getByRole('article', { name: 'Recorded experiments: observed relative frequency', exact: true });
  const keys = { method: 'option_2', assumptions: 'option_3', interpretation: 'option_1' };
  for (const [stage, option] of Object.entries(keys)) await observed.locator('fieldset').filter({ has: page.locator('legend', { hasText: stage }) }).locator(`input[value="${option}"]`).check();
  await observed.getByLabel('Numeric answer (observed value)').fill('2/5');
  await fillReference(page, '3/8');
  await page.getByRole('button', { name: 'Submit full solution', exact: true }).click();
  await expect(page.getByText('Worksheet completed. Your submission and review are saved.')).toBeVisible();
  const snapshot = (await (await page.request.get(new URL('/api/worksheets', url).href)).json())[0];
  expect(snapshot.dataset.rows).toHaveLength(5);
  expect(snapshot.items.every((item: { attempts: { is_correct: boolean; assistance: string[] }[] }) => item.attempts[0].is_correct && item.attempts[0].assistance.includes('reference'))).toBeTruthy();
  expect((await (await page.request.get(new URL('/api/bank', url).href)).json()).length).toBe(10);
  await stop(); url = await start(); await page.goto(url);
  await page.getByRole('button', { name: 'Worksheets', exact: true }).click();
  await expect(observed.getByLabel('Numeric answer (observed value)')).toHaveValue('2/5');
  await expect(page.getByText(/Saved status: completed. Supported reviewed-case/)).toBeVisible();
});
