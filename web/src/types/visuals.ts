export interface CurvePoint {
  x: number;
  pdf: number;
  cdf: number;
}

export interface NormalVisualResponse {
  family: string;
  mu: number;
  sigma: number;
  tail: 'lower' | 'upper' | 'interval' | 'two_tailed';
  cutoff_x?: number;
  cutoff_x1?: number;
  cutoff_x2?: number;
  z_score?: number;
  z_score_1?: number;
  z_score_2?: number;
  shaded_probability: number;
  density_at_cutoff?: number;
  excel_formula: string;
  summary: string;
  empirical_rule: {
    one_sigma: number;
    two_sigma: number;
    three_sigma: number;
  };
  points: CurvePoint[];
}

export interface UniformVisualResponse {
  family: string;
  a: number;
  b: number;
  tail: 'lower' | 'upper' | 'interval';
  cutoff_x?: number;
  cutoff_x1?: number;
  cutoff_x2?: number;
  height_density: number;
  shaded_probability: number;
  mean: number;
  variance: number;
  summary: string;
  points: CurvePoint[];
}

export interface ExponentialVisualResponse {
  family: string;
  lambda: number;
  mean: number;
  tail: 'lower' | 'upper' | 'interval';
  cutoff_x?: number;
  cutoff_x1?: number;
  cutoff_x2?: number;
  shaded_probability: number;
  density_at_cutoff?: number;
  excel_formula: string;
  memoryless_note: string;
  summary: string;
  points: CurvePoint[];
}

export interface DiscreteBar {
  k: number;
  pmf: number;
  cdf: number;
  is_highlighted: boolean;
}

export interface DiscreteVisualResponse {
  family: string;
  parameters: Record<string, unknown>;
  mean: number;
  variance: number;
  event: string;
  total_probability: number;
  excel_formula: string;
  summary: string;
  bars: DiscreteBar[];
}

export interface VennVisualResponse {
  p_a: number;
  p_b: number;
  p_intersection: number;
  region: string;
  regions: Record<string, number>;
  shaded_probability: number;
  formula: string;
  summary: string;
}

export interface SamplingBar {
  bin_mid: number;
  count: number;
  density: number;
  theoretical_density: number;
}

export interface SamplingVisualResponse {
  family: string;
  pop_distribution: string;
  pop_mean: number;
  pop_std_dev: number;
  sample_size: number;
  trials: number;
  standard_error: number;
  empirical_mean: number;
  empirical_std_err: number;
  bars: SamplingBar[];
  normal_curve: CurvePoint[];
  clt_note: string;
  excel_formula: string;
  summary: string;
}

export interface SimulatedInterval {
  index: number;
  sample_mean: number;
  lower_bound: number;
  upper_bound: number;
  captures_truth: boolean;
}

export interface ConfidenceIntervalVisualResponse {
  family: string;
  confidence_level: number;
  alpha: number;
  true_mean: number;
  sigma: number;
  sample_size: number;
  num_intervals: number;
  intervals: SimulatedInterval[];
  captured_count: number;
  empirical_coverage: number;
  excel_formula: string;
  summary: string;
}

export interface HypothesisVisualResponse {
  family: string;
  test_type: string;
  tail: string;
  null_hypothesis: string;
  alt_hypothesis: string;
  null_value: number;
  sample_stat: number;
  sample_size: number;
  standard_error: number;
  test_statistic: number;
  degrees_of_freedom?: number;
  p_value: number;
  alpha: number;
  critical_value: number;
  critical_value_2?: number;
  reject_null: boolean;
  decision: string;
  rejection_region: string;
  excel_formula: string;
  summary: string;
  points: CurvePoint[];
}

export interface DataPoint {
  x: number;
  y: number;
}

export interface ResidualPoint {
  x: number;
  y: number;
  y_hat: number;
  residual: number;
}

export interface RegressionVisualResponse {
  family: string;
  dataset_name: string;
  n: number;
  points: DataPoint[];
  slope: number;
  intercept: number;
  correlation: number;
  r_squared: number;
  std_err_est: number;
  equation: string;
  association_note: string;
  excel_formula: string;
  residuals: ResidualPoint[];
  line_start_x: number;
  line_start_y: number;
  line_end_x: number;
  line_end_y: number;
  predicted_x?: number;
  predicted_y?: number;
  summary: string;
}
