import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { Worksheets } from './Worksheets';
import type { Worksheet } from './types';

const record: Worksheet = {
  id: 'worksheet_test', revision: 1, mode: 'full_solution', status: 'draft', evidence: 'Supported practice', updated_at: '2026-10-09T12:00:00Z',
  questions: [{ id: 'question', title: 'Saved question', template_id: 'test', template_version: 1, seed: 42, scenario_markdown: 'A probability problem', assumptions: ['Explicit model assumptions'] }],
  items: [{ key: 'question:value', question_id: 'question', stage_id: 'value', kind: 'numeric', prompt_markdown: 'Calculate the probability.', status: 'active', options: [], numeric_policy: { version: 1, allowed_forms: ['decimal', 'fraction', 'percent'], absolute_tolerance: 1e-6, relative_tolerance: 1e-6 }, attempts: [] }],
};
const response = (body: unknown, status = 200) => ({ ok: status < 400, status, json: async () => body, text: async () => String(body) }) as Response;
afterEach(() => vi.unstubAllGlobals());
describe('Worksheets recovery', () => {
  it('preserves raw input across navigation and replays the same command after a lost response', async () => {
    const commands: string[] = [];
    vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
      if (url.endsWith('/commands')) {
        commands.push(String(init?.body));
        if (commands.length === 1) throw new Error('Connection lost');
        return response({ ...record, revision: 2, items: [{ ...record.items[0], draft_answer: { kind: 'numeric', numeric_raw: '3/8' } }] });
      }
      return response(url === '/api/bank' ? [{ id: 'test', title: 'Saved question' }] : [record]);
    }));
    const view = render(<Worksheets active launchMode="full_solution" />);
    const input = await screen.findByLabelText('Numeric answer (value)');
    await waitFor(() => expect(input).toBeEnabled());
    fireEvent.change(input, { target: { value: '3/8' } });
    view.rerender(<Worksheets active={false} launchMode="full_solution" />);
    view.rerender(<Worksheets active launchMode="full_solution" />);
    await waitFor(() => expect(input).toBeEnabled());
    expect(input).toHaveValue('3/8');
    fireEvent.click(screen.getByRole('button', { name: 'Save worksheet draft' }));
    await screen.findByRole('alert');
    expect(input).toBeDisabled();
    fireEvent.click(screen.getByRole('button', { name: 'Retry pending save or submission' }));
    await screen.findByText('Draft saved on this device.');
    expect(commands).toHaveLength(2); expect(commands[0]).toBe(commands[1]);
    expect(JSON.parse(commands[1]).type).toBe('save_draft');
    expect(input).toHaveValue('3/8');
  });
  it('keeps invalid answers editable and ignores an aborted creation response', async () => {
    let finish: (response: Response) => void = () => {};
    let signal: AbortSignal | undefined;
    vi.stubGlobal('fetch', vi.fn((url: string, init?: RequestInit) => {
      if (url.endsWith('/commands')) return Promise.resolve(response('Enter a valid number', 400));
      if (init?.method === 'POST') { signal = init.signal as AbortSignal; return new Promise<Response>(resolve => { finish = resolve; }); }
      return Promise.resolve(response(url === '/api/bank' ? [{ id: 'test', title: 'Saved question' }] : [record]));
    }));
    const view = render(<Worksheets active launchMode="full_solution" />);
    const input = await screen.findByLabelText('Numeric answer (value)');
    await waitFor(() => expect(input).toBeEnabled());
    fireEvent.change(input, { target: { value: '1/0' } });
    fireEvent.click(screen.getByRole('button', { name: 'Submit full solution' }));
    await screen.findByRole('alert');
    expect(input).toHaveValue('1/0'); expect(input).toBeEnabled();
    fireEvent.click(screen.getByRole('button', { name: 'Discard unsaved edits' }));
    fireEvent.click(screen.getByRole('button', { name: 'Create worksheet' }));
    view.rerender(<Worksheets active={false} launchMode="full_solution" />);
    expect(signal?.aborted).toBe(true);
    await act(async () => finish(response({ ...record, id: 'stale', questions: [{ ...record.questions[0], title: 'Stale completion' }] })));
    await act(async () => { view.rerender(<Worksheets active launchMode="full_solution" />); });
    await waitFor(() => expect(screen.getByRole('button', { name: 'Create worksheet' })).toBeEnabled());
    expect(screen.queryByRole('heading', { name: 'Stale completion' })).toBeNull();
  });
});
