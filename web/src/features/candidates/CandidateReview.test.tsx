import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CandidateReview } from './CandidateReview';

const record = {
  id: 'candidate_test', revision: 1, status: 'pending', created_at: '2026-10-09T12:00:00Z',
  proposal: { family_id: 'binomial_pmf', n: 4, p: .5, k: 2, title: 'Draft', scenario_markdown: 'Independent trials', success_label: 'success' },
  source: { mode: 'local', route: 'offline', model: 'local-binomial-v1', seed: 0 },
  template: { title: 'Draft', version: 1, scenario_markdown: 'Independent trials', assumptions: ['Independent trials'], stages: [], setting_group: 'binomial_fixed_independent_exact_count' }, reviews: [],
};
const response = (body: unknown, ok = true) => ({ ok, json: async () => body, text: async () => String(body) }) as Response;
afterEach(() => vi.unstubAllGlobals());
describe('Candidate review boundaries', () => {
  it('requires semantic confirmation and preserves review text on a conflict', async () => {
    const fetcher = vi.fn(async (url: string, init?: RequestInit) => {
      if (url.endsWith('/review')) return response('Candidate changed; reload', false);
      if (url === '/api/providers') return response({ active_route: 'offline' });
      if (init?.method === 'POST') return response(record);
      return response([record]);
    });
    vi.stubGlobal('fetch', fetcher);
    render(<CandidateReview active />);
    fireEvent.click(await screen.findByRole('button', { name: 'Draft — pending' }));
    fireEvent.change(screen.getByLabelText('Reviewer name'), { target: { value: 'Human' } });
    fireEvent.change(screen.getByLabelText('Review notes'), { target: { value: 'Checked every stage' } });
    const approve = screen.getByRole('button', { name: 'Approve for practice' });
    expect(approve).toBeDisabled();
    expect(fetcher.mock.calls.some(([url]) => url.endsWith('/review'))).toBe(false);
    fireEvent.click(screen.getByRole('checkbox'));
    fireEvent.click(approve);
    await screen.findByRole('alert');
    expect(screen.getByLabelText('Review notes')).toHaveValue('Checked every stage');
    expect(screen.getByLabelText('Reviewer name')).toHaveValue('Human');
    expect(fetcher.mock.calls.filter(([url]) => url.endsWith('/review'))).toHaveLength(1);
  });
  it('aborts a generation on navigation and ignores a stale completion', async () => {
    let finish: (r: Response) => void = () => {};
    let signal: AbortSignal | undefined;
    vi.stubGlobal('fetch', vi.fn((url: string, init?: RequestInit) => {
      if (init?.method === 'POST') { signal = init.signal as AbortSignal; return new Promise<Response>(resolve => { finish = resolve; }); }
      return Promise.resolve(response(url === '/api/providers' ? { active_route: 'offline' } : []));
    }));
    const view = render(<CandidateReview active />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'Generate local variation' })).toBeEnabled());
    fireEvent.click(screen.getByRole('button', { name: 'Generate local variation' }));
    view.rerender(<CandidateReview active={false} />);
    expect(signal?.aborted).toBe(true);
    await act(async () => finish(response(record)));
    await act(async () => { view.rerender(<CandidateReview active />); });
    expect(screen.queryByRole('heading', { name: 'Preview: Draft' })).toBeNull();
  });
});
