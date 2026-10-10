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
async function start(): Promise<string> {
  server = spawn(binary, ['--no-browser', '--data-dir', tempDir], { cwd: root, stdio: ['ignore', 'pipe', 'pipe'] });
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => reject(new Error('Candidate server startup timed out')), 10000);
    let output = '';
    server.on('error', error => { clearTimeout(timeout); reject(error); });
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
test.beforeAll(async () => { tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'quant-candidates-')); url = await start(); });
test.afterAll(async () => { await stop(); fs.rmSync(tempDir, { recursive: true, force: true }); });

test('candidate preview, explicit review, replay, retirement and export stay offline', async ({ page, context }) => {
  await context.route('**/*', route => route.request().url().startsWith('http://127.0.0.1:') ? route.continue() : route.abort());
  await page.goto(url);
  await page.getByRole('button', { name: 'Question candidates', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Request AI wording', exact: true })).toBeDisabled();
  await page.getByLabel('Variation seed').fill('1');
  await page.getByRole('button', { name: 'Generate local variation', exact: true }).click();
  const preview = page.getByRole('article', { name: 'Candidate preview' });
  await expect(preview.getByRole('heading', { name: 'Preview: Exactly 2 successes in 8 trials' })).toBeVisible();
  await expect(preview.getByText('0.31146240234375', { exact: true })).toBeVisible();
  await expect(preview.locator('mjx-container').first()).toBeVisible();
  const draft = (await (await page.request.get(new URL('/api/candidates', url).href)).json())[0];
  const bankURL = new URL('/api/bank', url).href;
  expect((await (await page.request.get(bankURL)).json()).length).toBe(10);
  const approve = page.getByRole('button', { name: 'Approve for practice', exact: true });
  await page.getByLabel('Reviewer name', { exact: true }).fill('E2E fixture reviewer');
  await page.getByLabel('Review notes', { exact: true }).fill('Reviewed the scenario, parameters, all stages, hints and answer. Isolated test fixture.');
  await expect(approve).toBeDisabled();
  await page.getByRole('checkbox').check();
  // Editable fields must not trigger practice shortcuts; candidate reading keys must not submit.
  await page.getByLabel('Review notes', { exact: true }).press('End');
  await page.getByLabel('Review notes', { exact: true }).type(' jkl');
  await approve.click();
  await expect(page.getByText('Approved for future practice sessions. Existing sessions keep their saved questions.')).toBeVisible();
  expect((await (await page.request.get(bankURL)).json()).length).toBe(11);
  const sessionResponse = await page.request.post(new URL('/api/practice/sessions', url).href, { data: { template_id: draft.id, seed: 42 } });
  expect(sessionResponse.ok()).toBeTruthy();
  const session = await sessionResponse.json();
  expect(session.stages.every((stage: { expected_answer?: unknown }) => !stage.expected_answer)).toBeTruthy();
  await page.getByLabel('Review notes', { exact: true }).fill('Retire test fixture, preserve its existing session.');
  await page.getByRole('button', { name: 'Retire from future practice', exact: true }).click();
  await expect(page.getByText('Retired from future practice. Past sessions are preserved.')).toBeVisible();
  expect((await (await page.request.get(bankURL)).json()).length).toBe(10);
  const downloadPromise = page.waitForEvent('download');
  await page.getByRole('button', { name: 'Export approved bank', exact: true }).click();
  const download = await downloadPromise;
  const exportPath = await download.path();
  const exported = JSON.parse(fs.readFileSync(exportPath!, 'utf8'));
  expect(exported).toHaveLength(10); expect(exported.some((item: { id: string }) => item.id === draft.id)).toBeFalsy();
  await stop(); url = await start();
  await page.goto(url);
  await page.getByRole('button', { name: 'Question candidates', exact: true }).click();
  await page.getByRole('button', { name: 'Exactly 2 successes in 8 trials — retired', exact: true }).click();
  await expect(preview.getByText('retired', { exact: true })).toBeVisible();
  const resumed = await (await page.request.get(new URL(`/api/practice/sessions/${session.id}`, url).href)).json();
  expect(resumed.scenario_markdown).toBe(session.scenario_markdown);
  expect(resumed.template_id).toBe(draft.id);
  await page.setViewportSize({ width: 375, height: 812 });
  await expect(preview.locator('mjx-container').first()).toBeVisible();
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 2);
  expect(overflow).toBeFalsy();
});
