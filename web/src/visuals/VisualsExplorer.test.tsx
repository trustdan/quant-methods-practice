import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { VisualsExplorer } from './VisualsExplorer';
import { NormalVisual } from './NormalVisual';
import { UniformVisual } from './UniformVisual';
import { ExponentialVisual } from './ExponentialVisual';
import { DiscreteVisual } from './DiscreteVisual';
import { VennVisual } from './VennVisual';

describe('VisualsExplorer & Distribution Visuals', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders all category tabs in VisualsExplorer', () => {
    render(<VisualsExplorer />);

    expect(screen.getByText('Normal Distribution')).toBeDefined();
    expect(screen.getByText('Uniform Distribution')).toBeDefined();
    expect(screen.getByText('Exponential Distribution')).toBeDefined();
    expect(screen.getByText('Discrete PMF / CDF')).toBeDefined();
    expect(screen.getByText('Venn Diagram')).toBeDefined();
  });

  it('switches active distribution view when tabs are clicked', () => {
    render(<VisualsExplorer initialTab="normal" />);

    // Initially in Normal Distribution
    expect(screen.getByText(/Normal Distribution Explorer/i)).toBeDefined();

    // Switch to Uniform Distribution
    fireEvent.click(screen.getByText('Uniform Distribution'));
    expect(screen.getByText(/Uniform Distribution Explorer/i)).toBeDefined();

    // Switch to Exponential Distribution
    fireEvent.click(screen.getByText('Exponential Distribution'));
    expect(screen.getByText(/Exponential Distribution Explorer/i)).toBeDefined();

    // Switch to Discrete PMF / CDF
    fireEvent.click(screen.getByText('Discrete PMF / CDF'));
    expect(screen.getByText(/Discrete Distribution Explorer/i)).toBeDefined();

    // Switch to Venn Diagram
    fireEvent.click(screen.getByText('Venn Diagram'));
    expect(screen.getByText(/Venn Diagram Explorer/i)).toBeDefined();
  });

  it('NormalVisual toggles between PDF and CDF modes and updates cutoff', async () => {
    const mockData = {
      family: 'normal',
      mu: 0,
      sigma: 1,
      tail: 'lower',
      cutoff_x: 1.5,
      z_score: 1.5,
      shaded_probability: 0.9332,
      density_at_cutoff: 0.1295,
      excel_formula: '=NORM.S.DIST(1.5, TRUE)',
      summary: 'Normal N(0, 1): Lower tail P(X <= 1.5) = 0.9332 (Z = 1.50).',
      empirical_rule: { one_sigma: 0.6827, two_sigma: 0.9545, three_sigma: 0.9973 },
      points: [
        { x: -3, pdf: 0.004, cdf: 0.001 },
        { x: 0, pdf: 0.3989, cdf: 0.5 },
        { x: 1.5, pdf: 0.1295, cdf: 0.9332 },
        { x: 3, pdf: 0.004, cdf: 0.999 },
      ],
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockData,
    }) as unknown as typeof fetch;

    render(<NormalVisual initialMu={0} initialSigma={1} />);

    await waitFor(() => {
      expect(screen.getByText('93.32%')).toBeDefined();
    });

    expect(screen.getByText(/Z = \(X - μ\) \/ σ/i)).toBeDefined();
    expect(screen.getByText('=NORM.S.DIST(1.5, TRUE)')).toBeDefined();

    // Toggle to CDF view
    const cdfBtn = screen.getByRole('button', { name: /CDF \(Cumulative S-Curve\)/i });
    fireEvent.click(cdfBtn);
    expect(cdfBtn.getAttribute('aria-pressed')).toBe('true');
  });

  it('UniformVisual displays constant density and updates interval', async () => {
    const mockUniData = {
      family: 'uniform',
      a: 0,
      b: 20,
      tail: 'interval',
      cutoff_x1: 5,
      cutoff_x2: 12,
      height_density: 0.05,
      shaded_probability: 0.35,
      mean: 10,
      variance: 33.33,
      summary: 'Uniform U(0, 20): Interval P(5 <= X <= 12) = 0.35.',
      points: [],
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockUniData,
    }) as unknown as typeof fetch;

    render(<UniformVisual />);

    await waitFor(() => {
      expect(screen.getByText('35.00%')).toBeDefined();
    });

    expect(screen.getByText('0.0500')).toBeDefined();
    expect(screen.getByText(/Height = 1 \/ \(20 - 0\)/i)).toBeDefined();
  });

  it('ExponentialVisual demonstrates memoryless property', async () => {
    const mockExpData = {
      family: 'exponential',
      lambda: 1.0,
      mean: 1.0,
      tail: 'upper',
      cutoff_x: 1.0,
      shaded_probability: 0.3679,
      density_at_cutoff: 0.3679,
      excel_formula: '=1 - EXPON.DIST(1, 1, TRUE)',
      memoryless_note: 'Memoryless property applies.',
      summary: 'Exponential (lambda=1): Survival P(X >= 1) = 0.3679.',
      points: [],
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockExpData,
    }) as unknown as typeof fetch;

    render(<ExponentialVisual />);

    await waitFor(() => {
      expect(screen.getByText('36.79%')).toBeDefined();
    });

    expect(screen.getByText(/Interactive Memoryless Property Check/i)).toBeDefined();
    expect(screen.getByText('=1 - EXPON.DIST(1, 1, TRUE)')).toBeDefined();
  });

  it('DiscreteVisual supports Binomial and Poisson toggling', async () => {
    const mockDiscData = {
      family: 'binomial',
      parameters: { n: 10, p: 0.5 },
      mean: 5.0,
      variance: 2.5,
      event: 'equal',
      total_probability: 0.2461,
      excel_formula: '=BINOM.DIST(5, 10, 0.5, FALSE)',
      summary: 'Binomial(n=10, p=0.5): Event equal has prob 0.2461.',
      bars: [
        { k: 4, pmf: 0.205, cdf: 0.377, is_highlighted: false },
        { k: 5, pmf: 0.2461, cdf: 0.623, is_highlighted: true },
      ],
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockDiscData,
    }) as unknown as typeof fetch;

    render(<DiscreteVisual />);

    await waitFor(() => {
      expect(screen.getByText('24.61%')).toBeDefined();
    });

    expect(screen.getByText('=BINOM.DIST(5, 10, 0.5, FALSE)')).toBeDefined();
    expect(screen.getByText(/μ = 5.00/i)).toBeDefined();
  });

  it('VennVisual displays regions and handles region selection', async () => {
    const mockVennData = {
      p_a: 0.6,
      p_b: 0.5,
      p_intersection: 0.3,
      region: 'union',
      regions: { a_only: 0.3, b_only: 0.2, intersection: 0.3, union: 0.8, neither: 0.2 },
      shaded_probability: 0.8,
      formula: 'P(A ∪ B) = P(A) + P(B) - P(A ∩ B)',
      summary: 'Venn Diagram: Shaded region union has probability 0.8.',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockVennData,
    }) as unknown as typeof fetch;

    render(<VennVisual />);

    // Click union button
    const unionBtn = screen.getByRole('button', { name: /A ∪ B \(Union\)/i });
    fireEvent.click(unionBtn);

    await waitFor(() => {
      expect(screen.getByText('80.0%')).toBeDefined();
    });

    expect(screen.getByText('P(A ∪ B) = P(A) + P(B) - P(A ∩ B)')).toBeDefined();
    expect(screen.getByText('Independent')).toBeDefined();
  });
});
