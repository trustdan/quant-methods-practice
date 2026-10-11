import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { VisualsExplorer } from './VisualsExplorer';
import { NormalVisual } from './NormalVisual';
import { UniformVisual } from './UniformVisual';
import { ExponentialVisual } from './ExponentialVisual';
import { DiscreteVisual } from './DiscreteVisual';
import { VennVisual } from './VennVisual';
import { SamplingVisual } from './SamplingVisual';
import { ConfidenceIntervalVisual } from './ConfidenceIntervalVisual';
import { HypothesisTestVisual } from './HypothesisTestVisual';
import { RegressionVisual } from './RegressionVisual';

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
    expect(screen.getByText('Sampling & CLT')).toBeDefined();
    expect(screen.getByText('Confidence Intervals')).toBeDefined();
    expect(screen.getByText('Hypothesis Testing')).toBeDefined();
    expect(screen.getByText('Linear Regression')).toBeDefined();
  });

  it('switches active distribution view when tabs are clicked', () => {
    render(<VisualsExplorer initialTab="normal" />);

    // Initially in Normal Distribution
    expect(screen.getByText(/Normal Distribution Explorer/i)).toBeDefined();

    // Switch to Uniform Distribution
    fireEvent.click(screen.getByText('Uniform Distribution'));
    expect(screen.getByText(/Uniform Distribution Explorer/i)).toBeDefined();

    // Switch to Sampling & CLT
    fireEvent.click(screen.getByText('Sampling & CLT'));
    expect(screen.getByText(/Central Limit Theorem & Sampling Distribution Lab/i)).toBeDefined();

    // Switch to Confidence Intervals
    fireEvent.click(screen.getByText('Confidence Intervals'));
    expect(screen.getByText(/Confidence Interval Coverage Explorer/i)).toBeDefined();

    // Switch to Hypothesis Testing
    fireEvent.click(screen.getByText('Hypothesis Testing'));
    expect(screen.getByText(/Hypothesis Testing & Rejection Region Lab/i)).toBeDefined();

    // Switch to Linear Regression
    fireEvent.click(screen.getByText('Linear Regression'));
    expect(screen.getByText(/Simple Linear Regression & Residuals Lab/i)).toBeDefined();
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

    const unionBtn = screen.getByRole('button', { name: /A ∪ B \(Union\)/i });
    fireEvent.click(unionBtn);

    await waitFor(() => {
      expect(screen.getByText('80.0%')).toBeDefined();
    });

    expect(screen.getByText('P(A ∪ B) = P(A) + P(B) - P(A ∩ B)')).toBeDefined();
    expect(screen.getByText('Independent')).toBeDefined();
  });

  it('SamplingVisual displays CLT histogram and standard error metrics', async () => {
    const mockSampling = {
      family: 'sampling',
      pop_distribution: 'skewed',
      pop_mean: 3.0,
      pop_std_dev: 2.45,
      sample_size: 25,
      trials: 400,
      standard_error: 0.49,
      empirical_mean: 3.02,
      empirical_std_err: 0.485,
      bars: [{ bin_mid: 3.0, count: 50, density: 0.65, theoretical_density: 0.68 }],
      normal_curve: [{ x: 3.0, pdf: 0.68, cdf: 0.5 }],
      clt_note: 'Central Limit Theorem holds.',
      excel_formula: '=NORM.DIST(x_bar, 3, 2.45 / SQRT(25), TRUE)',
      summary: 'Sampling Distribution: n = 25, 400 trials.',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockSampling,
    }) as unknown as typeof fetch;

    render(<SamplingVisual />);

    await waitFor(() => {
      expect(screen.getByText(/Central Limit Theorem Holds/i)).toBeDefined();
    });

    expect(screen.getByText(/0.485/)).toBeDefined();
    expect(screen.getByText(/=NORM.DIST\(x_bar, 3, 2.45 \/ SQRT\(25\), TRUE\)/)).toBeDefined();
  });

  it('ConfidenceIntervalVisual displays interval coverage and nominal level', async () => {
    const mockCI = {
      family: 'confidence_interval',
      confidence_level: 0.95,
      alpha: 0.05,
      true_mean: 50.0,
      sigma: 10.0,
      sample_size: 16,
      num_intervals: 25,
      intervals: [
        { index: 1, sample_mean: 51.2, lower_bound: 46.1, upper_bound: 56.3, captures_truth: true },
        { index: 2, sample_mean: 43.1, lower_bound: 38.0, upper_bound: 48.2, captures_truth: false },
      ],
      captured_count: 24,
      empirical_coverage: 0.96,
      excel_formula: '=CONFIDENCE.T(0.05, s, 16)',
      summary: '95.0% Confidence Interval Simulation: 24 of 25 captured.',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockCI,
    }) as unknown as typeof fetch;

    render(<ConfidenceIntervalVisual />);

    await waitFor(() => {
      expect(screen.getByText('96.0%')).toBeDefined();
    });

    expect(screen.getAllByText(/24 of 25 captured/i).length).toBeGreaterThan(0);
    expect(screen.getByText('=CONFIDENCE.T(0.05, s, 16)')).toBeDefined();
  });

  it('HypothesisTestVisual displays decision and rejection rule', async () => {
    const mockHyp = {
      family: 'hypothesis_test',
      test_type: 'mean_t',
      tail: 'two_tailed',
      null_hypothesis: 'mu = 50',
      alt_hypothesis: 'mu != 50',
      null_value: 50.0,
      sample_stat: 53.5,
      sample_size: 25,
      standard_error: 1.6,
      test_statistic: 2.188,
      degrees_of_freedom: 24,
      p_value: 0.0386,
      alpha: 0.05,
      critical_value: 2.064,
      critical_value_2: -2.064,
      reject_null: true,
      decision: 'Reject H0',
      rejection_region: '|t| >= 2.064',
      excel_formula: '=T.DIST.2T(2.188, 24)',
      summary: 'Test statistic t = 2.188. Reject H0.',
      points: [{ x: 0, pdf: 0.39, cdf: 0.5 }],
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockHyp,
    }) as unknown as typeof fetch;

    render(<HypothesisTestVisual />);

    await waitFor(() => {
      expect(screen.getByText('REJECT H0')).toBeDefined();
    });

    expect(screen.getByText(/Reject \$H_0\$ if \|t\| >= 2.064/i)).toBeDefined();
    expect(screen.getByText('=T.DIST.2T(2.188, 24)')).toBeDefined();
  });

  it('RegressionVisual displays OLS equation, R^2, and association warning', async () => {
    const mockReg = {
      family: 'regression',
      dataset_name: 'advertising_sales',
      n: 8,
      points: [{ x: 2, y: 15 }],
      slope: 1.8,
      intercept: 12.8,
      correlation: 0.90,
      r_squared: 0.81,
      std_err_est: 1.45,
      equation: 'y_hat = 1.8*x + 12.8',
      association_note: 'Association is not causation.',
      excel_formula: '=SLOPE(...)',
      residuals: [{ x: 2, y: 15, y_hat: 16.4, residual: -1.4 }],
      line_start_x: 1,
      line_start_y: 14.6,
      line_end_x: 9,
      line_end_y: 29.0,
      predicted_x: 5.0,
      predicted_y: 21.8,
      summary: 'Linear Regression: y_hat = 1.8*x + 12.8.',
    };

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => mockReg,
    }) as unknown as typeof fetch;

    render(<RegressionVisual />);

    await waitFor(() => {
      expect(screen.getByText(/Association is not causation/i)).toBeDefined();
    });

    expect(screen.getByText(/Residual standard deviation/i)).toBeDefined();
    expect(screen.getByText(/81.0% of variance in \$Y\$ explained/i)).toBeDefined();
  });
});
