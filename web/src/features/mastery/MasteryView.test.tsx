import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MasteryView } from './MasteryView';
import { MasterySummary } from '../../types/practice';

const mockMasteryData: MasterySummary = {
  policy_version: 1,
  total_mastered: 1,
  total_transferring: 1,
  total_learning: 1,
  total_new: 0,
  overall_score: 0.75,
  generated_at: '2026-10-05T18:00:00Z',
  concepts: [
    {
      concept_id: 'binomial_pmf',
      policy_version: 1,
      status: 'mastered',
      scaffold_level: 'faded',
      base_score: 0.85,
      retention_factor: 0.98,
      decayed_score: 0.833,
      independent_successes: 3,
      independent_errors: 0,
      assisted_count: 0,
      total_evidence_count: 3,
      half_life_days: 3.0,
      setting_groups_seen: ['coins', 'defects'],
      delayed_transfer_achieved: true,
      recent_error: false,
      priority_score: 0.2,
    },
    {
      concept_id: 'poisson_pmf',
      policy_version: 1,
      status: 'transferring',
      scaffold_level: 'intermediate',
      base_score: 0.65,
      retention_factor: 0.95,
      decayed_score: 0.6175,
      independent_successes: 2,
      independent_errors: 1,
      assisted_count: 1,
      total_evidence_count: 4,
      half_life_days: 3.0,
      setting_groups_seen: ['arrivals'],
      delayed_transfer_achieved: false,
      recent_error: false,
      priority_score: 0.5,
    },
    {
      concept_id: 'bayes_theorem',
      policy_version: 1,
      status: 'learning',
      scaffold_level: 'full',
      base_score: 0.40,
      retention_factor: 1.0,
      decayed_score: 0.40,
      independent_successes: 0,
      independent_errors: 2,
      assisted_count: 2,
      total_evidence_count: 4,
      half_life_days: 3.0,
      setting_groups_seen: ['medical'],
      delayed_transfer_achieved: false,
      recent_error: true,
      priority_score: 0.9,
    },
  ],
};

describe('MasteryView', () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    global.fetch = vi.fn().mockImplementation((url: string) => {
      if (url === '/api/mastery') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve(mockMasteryData),
        });
      }
      return Promise.reject(new Error(`Unhandled URL: ${url}`));
    });
  });

  afterEach(() => {
    global.fetch = originalFetch;
    vi.clearAllMocks();
  });

  it('renders summary statistics and concept cards from API', async () => {
    const handleClose = vi.fn();
    render(<MasteryView onClose={handleClose} />);

    // Shows loading initially
    expect(screen.getByText(/Computing read-time decay and transfer records/)).toBeInTheDocument();

    // Await API response resolution
    await waitFor(() => {
      expect(screen.getByText('Concept Evidence & Transfer')).toBeInTheDocument();
    });

    // Check header and policy badge
    expect(screen.getByText('Policy v1')).toBeInTheDocument();

    // Check overall stats
    expect(screen.getByText('75.0%')).toBeInTheDocument();
    expect(screen.getByText('Mastered (Faded)')).toBeInTheDocument();

    // Check concepts rendered
    expect(screen.getByText('Binomial Pmf')).toBeInTheDocument();
    expect(screen.getByText('Poisson Pmf')).toBeInTheDocument();
    expect(screen.getByText('Bayes Theorem')).toBeInTheDocument();

    // Close button triggers callback
    const closeBtn = screen.getByRole('button', { name: /Return to practice/i });
    fireEvent.click(closeBtn);
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('filters concepts by status pill', async () => {
    render(<MasteryView onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('Binomial Pmf')).toBeInTheDocument();
    });

    // Filter to "Mastered"
    const masteredFilter = screen.getByRole('button', { name: 'Mastered' });
    fireEvent.click(masteredFilter);

    expect(screen.getByText('Binomial Pmf')).toBeInTheDocument();
    expect(screen.queryByText('Poisson Pmf')).not.toBeInTheDocument();
    expect(screen.queryByText('Bayes Theorem')).not.toBeInTheDocument();

    // Filter to "Transferring"
    const transferringFilter = screen.getByRole('button', { name: 'Transferring' });
    fireEvent.click(transferringFilter);

    expect(screen.queryByText('Binomial Pmf')).not.toBeInTheDocument();
    expect(screen.getByText('Poisson Pmf')).toBeInTheDocument();
    expect(screen.queryByText('Bayes Theorem')).not.toBeInTheDocument();

    // Filter back to "All"
    const allFilter = screen.getByRole('button', { name: 'All' });
    fireEvent.click(allFilter);

    expect(screen.getByText('Binomial Pmf')).toBeInTheDocument();
    expect(screen.getByText('Poisson Pmf')).toBeInTheDocument();
    expect(screen.getByText('Bayes Theorem')).toBeInTheDocument();
  });

  it('filters concepts by search term', async () => {
    render(<MasteryView onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText('Binomial Pmf')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search concepts...');
    fireEvent.change(searchInput, { target: { value: 'poisson' } });

    expect(screen.getByText('Poisson Pmf')).toBeInTheDocument();
    expect(screen.queryByText('Binomial Pmf')).not.toBeInTheDocument();
    expect(screen.queryByText('Bayes Theorem')).not.toBeInTheDocument();
  });

  it('displays error message when API call fails', async () => {
    global.fetch = vi.fn().mockImplementation(() =>
      Promise.resolve({
        ok: false,
        status: 500,
      })
    );

    render(<MasteryView onClose={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText(/Server returned 500/)).toBeInTheDocument();
    });
  });
});
