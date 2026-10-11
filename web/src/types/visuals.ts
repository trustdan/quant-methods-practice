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
