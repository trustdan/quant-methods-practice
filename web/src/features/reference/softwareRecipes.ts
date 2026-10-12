/**
 * Canonical Software Recipe Catalog (Phase 1)
 *
 * Strongly typed registry of Microsoft Excel functions/formulae and R / RStudio code
 * covering all 6 curriculum modules and 22 learning objectives.
 *
 * Maps cleanly to existing question templates without requiring schema migrations.
 */

export interface SoftwareArgumentDesc {
  name: string;
  description: string;
}

export interface SoftwareSpec {
  functionName: string;
  genericSyntax: string;
  argumentDescriptions: SoftwareArgumentDesc[];
  notes?: string;
  commonPitfalls?: string;
  exampleFormula?: string;
}

export interface SoftwareRecipe {
  id: string;
  familyId?: string;
  moduleId: string;
  title: string;
  statisticalTask: string;
  relevantTemplateIds?: string[];
  applicableKeywords?: string[];
  excel: SoftwareSpec;
  r: SoftwareSpec;
}

export const SOFTWARE_RECIPES: SoftwareRecipe[] = [
  // ==========================================
  // MODULE 1: DESCRIPTIVE MEASURES & PROBABILITY
  // ==========================================
  {
    id: 'm1_center_measures',
    familyId: 'descriptive_stats',
    moduleId: 'module_1',
    title: 'Measures of Center: Mean, Median, and Mode',
    statisticalTask: 'Calculate central tendency benchmarks for numerical data.',
    relevantTemplateIds: ['desc_mean_median_outlier'],
    applicableKeywords: ['mean', 'median', 'mode', 'average', 'center'],
    excel: {
      functionName: 'AVERAGE, MEDIAN, MODE.SNGL',
      genericSyntax: '=AVERAGE(range) | =MEDIAN(range) | =MODE.SNGL(range)',
      argumentDescriptions: [
        { name: 'range', description: 'Continuous cell range of numerical observations (e.g., A2:A50).' },
      ],
      notes: 'Median is robust to extreme outliers and skewed distributions; Mean is sensitive to every value.',
      commonPitfalls: 'Using AVERAGE on data with extreme outliers instead of MEDIAN, which distorts the typical value.',
      exampleFormula: '=MEDIAN(B2:B8)',
    },
    r: {
      functionName: 'mean, median',
      genericSyntax: 'mean(x, na.rm = TRUE) | median(x, na.rm = TRUE)',
      argumentDescriptions: [
        { name: 'x', description: 'Numeric vector of observations.' },
        { name: 'na.rm', description: 'Logical flag whether to strip missing values before computing (default FALSE).' },
      ],
      notes: 'Base R does not have a single mode() function for statistical mode; use names(which.max(table(x))) or the DescTools package.',
      commonPitfalls: 'Calling mode(x) in R returns the storage mode (e.g., "numeric"), not the statistical mode.',
      exampleFormula: 'median(sales_data, na.rm = TRUE)',
    },
  },
  {
    id: 'm1_dispersion_sample',
    familyId: 'descriptive_stats',
    moduleId: 'module_1',
    title: 'Sample Variance & Standard Deviation (Bessel Correction)',
    statisticalTask: 'Compute sample variability with n - 1 degrees of freedom.',
    relevantTemplateIds: ['desc_sample_sd_delivery_times'],
    applicableKeywords: ['variance', 'standard deviation', 'stdev', 'dispersion', 'bessel'],
    excel: {
      functionName: 'VAR.S, STDEV.S',
      genericSyntax: '=VAR.S(range) | =STDEV.S(range)',
      argumentDescriptions: [
        { name: 'range', description: 'Cell range of sample observations.' },
      ],
      notes: 'Uses divisor (n - 1) to provide an unbiased estimator of population variance.',
      commonPitfalls: 'Using VAR.P or STDEV.P on sample data, which divides by n and underestimates population dispersion.',
      exampleFormula: '=STDEV.S(A2:A7)',
    },
    r: {
      functionName: 'var, sd',
      genericSyntax: 'var(x, na.rm = TRUE) | sd(x, na.rm = TRUE)',
      argumentDescriptions: [
        { name: 'x', description: 'Numeric vector of sample observations.' },
        { name: 'na.rm', description: 'Logical flag to omit missing NA values.' },
      ],
      notes: 'Base R functions var() and sd() automatically apply Bessel correction (n - 1).',
      commonPitfalls: 'Expecting sd() to divide by n for population data; R strictly computes sample sd.',
      exampleFormula: 'sd(delivery_times)',
    },
  },
  {
    id: 'm1_coef_variation',
    familyId: 'descriptive_stats',
    moduleId: 'module_1',
    title: 'Coefficient of Variation (Relative Risk & Dispersion)',
    statisticalTask: 'Compare relative dispersion across variables with different scales or means.',
    relevantTemplateIds: ['desc_cv_compare_stocks'],
    applicableKeywords: ['coefficient of variation', 'relative dispersion', 'relative risk', 'cv'],
    excel: {
      functionName: 'CV (Formula)',
      genericSyntax: '=(STDEV.S(range) / AVERAGE(range)) * 100',
      argumentDescriptions: [
        { name: 'range', description: 'Cell range of numerical observations (with strictly positive mean).' },
      ],
      notes: 'Expressed as a percentage (or decimal ratio) indicating risk per unit of expected return.',
      commonPitfalls: 'Comparing raw standard deviations between assets with widely different price levels or means.',
      exampleFormula: '=(STDEV.S(B2:B20) / AVERAGE(B2:B20)) * 100',
    },
    r: {
      functionName: 'sd(x) / mean(x)',
      genericSyntax: '(sd(x) / mean(x)) * 100',
      argumentDescriptions: [
        { name: 'x', description: 'Numeric vector with non-zero mean.' },
      ],
      notes: 'CV is dimensionless, allowing direct comparisons between different currencies, units, or scales.',
      commonPitfalls: 'Calculating CV on interval data where zero is arbitrary (e.g. Celsius or Fahrenheit temperatures).',
      exampleFormula: '(sd(stock_returns) / mean(stock_returns)) * 100',
    },
  },
  {
    id: 'm1_zscore_standardize',
    familyId: 'descriptive_stats',
    moduleId: 'module_1',
    title: 'Standardization and Z-Scores',
    statisticalTask: 'Standardize observations into standard deviation units from the mean.',
    relevantTemplateIds: ['desc_zscore_empirical_rule'],
    applicableKeywords: ['z-score', 'standardize', 'empirical rule', 'standard deviation units'],
    excel: {
      functionName: 'STANDARDIZE',
      genericSyntax: '=STANDARDIZE(x, mean, standard_dev)',
      argumentDescriptions: [
        { name: 'x', description: 'The individual observation value to standardize.' },
        { name: 'mean', description: 'Arithmetic mean of the distribution (or AVERAGE(range)).' },
        { name: 'standard_dev', description: 'Standard deviation (or STDEV.S(range)).' },
      ],
      notes: 'A z-score measures how many standard deviations an observation falls above (+) or below (-) the mean.',
      commonPitfalls: 'Forgetting parentheses when typing =(x - mean) / sd manually, yielding x - (mean / sd).',
      exampleFormula: '=STANDARDIZE(85, 70, 10)',
    },
    r: {
      functionName: 'scale',
      genericSyntax: '(x - mean(x)) / sd(x) | as.vector(scale(x))',
      argumentDescriptions: [
        { name: 'x', description: 'Individual value or vector of observations to standardize.' },
      ],
      notes: 'scale() returns a 1-column matrix by default; wrap in as.vector() for a standard numeric vector.',
      commonPitfalls: 'Using scale() on a single scalar number without providing population or sample parameters.',
      exampleFormula: '(sales - mean(sales)) / sd(sales)',
    },
  },
  {
    id: 'm1_weighted_mean',
    familyId: 'descriptive_stats',
    moduleId: 'module_1',
    title: 'Weighted Mean',
    statisticalTask: 'Calculate weighted average where observations contribute according to weights.',
    relevantTemplateIds: ['desc_weighted_mean_price'],
    applicableKeywords: ['weighted mean', 'weighted average', 'weights', 'sumproduct'],
    excel: {
      functionName: 'SUMPRODUCT / SUM',
      genericSyntax: '=SUMPRODUCT(values_range, weights_range) / SUM(weights_range)',
      argumentDescriptions: [
        { name: 'values_range', description: 'Range containing numerical values (e.g., unit prices).' },
        { name: 'weights_range', description: 'Range containing corresponding weights (e.g., quantities sold).' },
      ],
      notes: 'Multiplies corresponding elements and divides by the total weight sum.',
      commonPitfalls: 'Unequal range dimensions between values_range and weights_range causes a #VALUE! error.',
      exampleFormula: '=SUMPRODUCT(B2:B5, C2:C5) / SUM(C2:C5)',
    },
    r: {
      functionName: 'weighted.mean',
      genericSyntax: 'weighted.mean(x, w, na.rm = TRUE)',
      argumentDescriptions: [
        { name: 'x', description: 'Numeric vector of values.' },
        { name: 'w', description: 'Numeric vector of corresponding weights.' },
      ],
      notes: 'Length of x and w must match exactly.',
      commonPitfalls: 'Passing unnormalized percentages or negative weights where not applicable.',
      exampleFormula: 'weighted.mean(prices, quantities)',
    },
  },
  {
    id: 'm1_quartiles_iqr',
    familyId: 'descriptive_stats',
    moduleId: 'module_1',
    title: 'Quartiles, Interquartile Range (IQR) & Outlier Fences',
    statisticalTask: 'Compute Q1, Q3, IQR and fence limits [Q1 - 1.5*IQR, Q3 + 1.5*IQR].',
    relevantTemplateIds: ['desc_quartiles_iqr_fences'],
    applicableKeywords: ['quartile', 'iqr', 'fences', 'outlier', 'tukey'],
    excel: {
      functionName: 'QUARTILE.INC',
      genericSyntax: '=QUARTILE.INC(array, quart)',
      argumentDescriptions: [
        { name: 'array', description: 'Cell range of numerical observations.' },
        { name: 'quart', description: 'Quartile index: 1 = Q1 (25th percentile), 2 = Median, 3 = Q3 (75th percentile).' },
      ],
      notes: 'QUARTILE.INC uses standard inclusive interpolation, matching standard course textbook conventions.',
      commonPitfalls: 'Using QUARTILE.EXC which excludes 0% and 100% and can yield #NUM! errors on small datasets.',
      exampleFormula: '=QUARTILE.INC(A2:A20, 3) - QUARTILE.INC(A2:A20, 1)',
    },
    r: {
      functionName: 'quantile, IQR',
      genericSyntax: 'quantile(x, probs = c(0.25, 0.75), type = 7) | IQR(x)',
      argumentDescriptions: [
        { name: 'x', description: 'Numeric vector of data.' },
        { name: 'probs', description: 'Numeric vector of probabilities in [0, 1].' },
        { name: 'type', description: 'Quantile algorithm (type = 7 is R default, matching Excel QUARTILE.INC).' },
      ],
      notes: 'type = 7 provides the standard linear interpolation of order statistics.',
      commonPitfalls: 'Different statistical packages use different quantile types (1 through 9); type 7 matches Excel.',
      exampleFormula: 'IQR(data_vector)',
    },
  },
  {
    id: 'm1_prob_complement',
    familyId: 'set_probability',
    moduleId: 'module_1',
    title: 'Complement Rule (At Least One Strategy)',
    statisticalTask: 'Compute P(A^c) = 1 - P(A), especially for at-least-one independent outcomes.',
    relevantTemplateIds: ['prob_at_least_one_server_failure', 'set_probability_complement_rule'],
    applicableKeywords: ['complement', 'at least one', 'none', 'independent failure'],
    excel: {
      functionName: '1 - P(None)',
      genericSyntax: '=1 - (1 - p)^n | =1 - P_A',
      argumentDescriptions: [
        { name: 'p', description: 'Probability of individual component event or failure.' },
        { name: 'n', description: 'Number of independent components or trials.' },
      ],
      notes: 'P(at least one failure) = 1 - P(all survive) = 1 - (1 - p)^n.',
      commonPitfalls: 'Multiplying n * p linearly (e.g. 3 * 0.05 = 0.15) which ignores joint simultaneous events.',
      exampleFormula: '=1 - (1 - 0.05)^3',
    },
    r: {
      functionName: '1 - (1 - p)^n',
      genericSyntax: '1 - (1 - p)^n',
      argumentDescriptions: [
        { name: 'p', description: 'Probability of individual component failure.' },
        { name: 'n', description: 'Number of independent components.' },
      ],
      notes: 'Standard complement formulation for reliability engineering.',
      commonPitfalls: 'Assuming independence when components share a common point of failure (e.g. shared power grid).',
      exampleFormula: '1 - (1 - 0.05)^3',
    },
  },
  {
    id: 'm1_prob_union_addition',
    familyId: 'set_probability',
    moduleId: 'module_1',
    title: 'Addition Rule for the Union of Events',
    statisticalTask: 'Calculate P(A or B) = P(A) + P(B) - P(A and B).',
    relevantTemplateIds: ['set_probability_union_rule'],
    applicableKeywords: ['union', 'addition rule', 'disjoint', 'mutually exclusive'],
    excel: {
      functionName: 'P(A) + P(B) - P(A and B)',
      genericSyntax: '=p_A + p_B - p_AB',
      argumentDescriptions: [
        { name: 'p_A', description: 'Marginal probability of event A.' },
        { name: 'p_B', description: 'Marginal probability of event B.' },
        { name: 'p_AB', description: 'Joint probability of the intersection P(A and B).' },
      ],
      notes: 'If events are mutually exclusive (disjoint), P(A and B) = 0 and formula simplifies to P(A) + P(B).',
      commonPitfalls: 'Forgetting to subtract the joint overlap P(A and B), which double-counts outcomes and may exceed 1.',
      exampleFormula: '=0.40 + 0.35 - 0.15',
    },
    r: {
      functionName: 'p_A + p_B - p_AB',
      genericSyntax: 'p_A + p_B - p_AB',
      argumentDescriptions: [
        { name: 'p_A', description: 'Probability of event A.' },
        { name: 'p_B', description: 'Probability of event B.' },
        { name: 'p_AB', description: 'Joint probability of A and B.' },
      ],
      notes: 'Fréchet bounds guarantee max(p_A, p_B) <= P(A or B) <= min(1, p_A + p_B).',
      commonPitfalls: 'Assuming events are disjoint without explicit verification.',
      exampleFormula: '0.40 + 0.35 - 0.15',
    },
  },
  {
    id: 'm1_prob_conditional_mult',
    familyId: 'set_probability',
    moduleId: 'module_1',
    title: 'Conditional Probability & Multiplication Rule',
    statisticalTask: 'Calculate P(A|B) = P(A and B) / P(B) and test independence.',
    relevantTemplateIds: ['set_probability_conditional', 'prob_multiplication_independent_suppliers', 'prob_independence_check'],
    applicableKeywords: ['conditional', 'multiplication rule', 'independence', 'joint probability'],
    excel: {
      functionName: 'P(A and B) / P(B)',
      genericSyntax: '=p_AB / p_B | =p_A * p_B',
      argumentDescriptions: [
        { name: 'p_AB', description: 'Joint intersection probability P(A and B).' },
        { name: 'p_B', description: 'Given conditioning event probability P(B) > 0.' },
      ],
      notes: 'Events A and B are independent if and only if P(A and B) = P(A) * P(B) or P(A|B) = P(A).',
      commonPitfalls: 'Confusing P(A|B) with P(B|A) (Confusion of the Inverse / Prosecutor Fallacy).',
      exampleFormula: '=0.12 / 0.30',
    },
    r: {
      functionName: 'p_AB / p_B',
      genericSyntax: 'p_AB / p_B',
      argumentDescriptions: [
        { name: 'p_AB', description: 'Joint probability of A and B.' },
        { name: 'p_B', description: 'Conditioning probability of B.' },
      ],
      notes: 'Under independence, conditional probability equals the unconditional marginal probability.',
      commonPitfalls: 'Confusing mutual exclusivity (disjoint, P(AB) = 0) with independence (P(AB) = P(A)P(B)).',
      exampleFormula: '0.12 / 0.30',
    },
  },
  {
    id: 'm1_prob_bayes_total',
    familyId: 'set_probability',
    moduleId: 'module_1',
    title: 'Law of Total Probability and Bayes’ Theorem',
    statisticalTask: 'Update prior probabilities to posterior probabilities given observed evidence.',
    relevantTemplateIds: ['prob_two_plant_defect_tree'],
    applicableKeywords: ['bayes', 'total probability', 'prior', 'posterior', 'tree'],
    excel: {
      functionName: 'Bayes Formula',
      genericSyntax: '=(p_B_given_A * p_A) / (p_B_given_A * p_A + p_B_given_notA * p_notA)',
      argumentDescriptions: [
        { name: 'p_B_given_A', description: 'Likelihood of evidence B given hypothesis A.' },
        { name: 'p_A', description: 'Prior probability of hypothesis A.' },
      ],
      notes: 'Denominator is the total probability of event B across all partition states.',
      commonPitfalls: 'Forgetting to weight the conditional probabilities by the prior base rates in the denominator.',
      exampleFormula: '=(0.02 * 0.60) / (0.02 * 0.60 + 0.05 * 0.40)',
    },
    r: {
      functionName: 'Bayes Theorem',
      genericSyntax: '(p_B_given_A * p_A) / sum(p_B_given_states * p_states)',
      argumentDescriptions: [
        { name: 'p_B_given_A', description: 'Conditional probability of evidence given target state.' },
        { name: 'p_A', description: 'Prior probability of target state.' },
        { name: 'p_B_given_states', description: 'Vector of conditional probabilities across all partition states.' },
        { name: 'p_states', description: 'Vector of prior probabilities summing to 1.' },
      ],
      notes: 'Vectorized in R: (p_B_given_states[1] * p_states[1]) / sum(p_B_given_states * p_states).',
      commonPitfalls: 'Base Rate Fallacy: ignoring rare prior probabilities when interpreting diagnostic tests.',
      exampleFormula: '(0.05 * 0.40) / sum(c(0.02, 0.05) * c(0.60, 0.40))',
    },
  },
  {
    id: 'm1_contingency_independence',
    familyId: 'concept_check',
    moduleId: 'module_1',
    title: 'Contingency Tables & Statistical Dependence',
    statisticalTask: 'Analyze joint frequencies and evaluate marginal vs conditional probabilities.',
    relevantTemplateIds: ['prob_contingency_age_purchase', 'prob_disjoint_vs_independent', 'concept_cross_sectional_vs_time_series', 'concept_data_types_business', 'concept_descriptive_vs_inferential', 'concept_parameter_vs_statistic', 'concept_sampling_bias_online_poll'],
    applicableKeywords: ['contingency table', 'joint probability', 'marginal probability', 'cross-sectional'],
    excel: {
      functionName: 'COUNTIFS, PivotTable',
      genericSyntax: '=COUNTIFS(col1_range, criteria1, col2_range, criteria2) / total_count',
      argumentDescriptions: [
        { name: 'col1_range', description: 'Column range for row classification.' },
        { name: 'criteria1', description: 'Specific category value for row.' },
      ],
      notes: 'Excel PivotTables quickly generate 2-way contingency tables with "% of Row Total" or "% of Grand Total".',
      commonPitfalls: 'Confusing cell percentage (% of Grand Total = joint) with row percentage (% of Row = conditional).',
      exampleFormula: '=60 / 500',
    },
    r: {
      functionName: 'table, prop.table',
      genericSyntax: 'table(row_var, col_var) | prop.table(table(row_var, col_var))',
      argumentDescriptions: [
        { name: 'row_var', description: 'Factor or vector defining table rows.' },
        { name: 'col_var', description: 'Factor or vector defining table columns.' },
      ],
      notes: 'prop.table(tbl, 1) gives row conditionals; prop.table(tbl, 2) gives column conditionals; prop.table(tbl) gives joint distribution.',
      commonPitfalls: 'Omitting margin parameter in prop.table() defaults to total proportion instead of conditional.',
      exampleFormula: 'prop.table(table(age_group, purchased), margin = 1)',
    },
  },

  // ==========================================
  // MODULE 2: DISCRETE RANDOM VARIABLES
  // ==========================================
  {
    id: 'm2_discrete_rv_moments',
    familyId: 'discrete_rv',
    moduleId: 'module_2',
    title: 'Discrete Random Variable Expected Value, Variance & SD',
    statisticalTask: 'Compute E[X] = SUM(x * P(x)) and Var(X) = E[X^2] - (E[X])^2.',
    relevantTemplateIds: ['drv_expected_cars_sold', 'drv_variance_cars_sold', 'drv_valid_pmf_missing_value', 'drv_compare_projects_risk', 'drv_insurance_break_even', 'drv_cumulative_from_table'],
    applicableKeywords: ['discrete random variable', 'expected value', 'second moment', 'variance', 'sumproduct'],
    excel: {
      functionName: 'SUMPRODUCT',
      genericSyntax: '=SUMPRODUCT(x_range, p_range)',
      argumentDescriptions: [
        { name: 'x_range', description: 'Cell range of discrete outcome values x.' },
        { name: 'p_range', description: 'Cell range of corresponding probabilities P(X = x).' },
      ],
      notes: 'Second moment E[X^2] is =SUMPRODUCT(x_range^2, p_range). Variance is E[X^2] - (E[X])^2. Standard deviation is SQRT(Var).',
      commonPitfalls: 'Forgetting to subtract (E[X])^2 from E[X^2], which results in second moment rather than variance.',
      exampleFormula: '=SUMPRODUCT(A2:A6, B2:B6)',
    },
    r: {
      functionName: 'sum(x * p)',
      genericSyntax: 'E_X <- sum(x * p); Var_X <- sum(x^2 * p) - E_X^2; SD_X <- sqrt(Var_X)',
      argumentDescriptions: [
        { name: 'x', description: 'Numeric vector of discrete outcomes.' },
        { name: 'p', description: 'Numeric vector of probabilities (must sum to 1).' },
      ],
      notes: 'Can also use weighted.mean(x, p) for E[X].',
      commonPitfalls: 'Using sample variance var(x) instead of population distribution variance sum(x^2 * p) - E[X]^2.',
      exampleFormula: 'sum(x * p)',
    },
  },
  {
    id: 'm2_linear_comb_affine',
    familyId: 'linear_combination',
    moduleId: 'module_2',
    title: 'Affine Transformations: Scaling & Fixed Constants',
    statisticalTask: 'Compute E[aX + b] = aE[X] + b and Var(aX + b) = a^2 * Var(X).',
    relevantTemplateIds: ['affine_profit_from_units', 'affine_currency_conversion', 'lincomb_weekly_demand_sum_vs_scale'],
    applicableKeywords: ['affine', 'scaling', 'shift', 'linear transformation', 'fixed cost'],
    excel: {
      functionName: 'E[aX + b] & Var(aX + b)',
      genericSyntax: '=a * E_X + b | =a^2 * Var_X | =ABS(a) * SD_X',
      argumentDescriptions: [
        { name: 'a', description: 'Multiplicative slope / conversion rate constant.' },
        { name: 'b', description: 'Additive fixed shift constant (contributes 0 variance).' },
      ],
      notes: 'Adding a fixed constant b shifts the mean by b, but adds ZERO to variance because dispersion is unaffected.',
      commonPitfalls: 'Adding b to the variance or squaring b. Constants have zero dispersion: Var(b) = 0.',
      exampleFormula: '=50 * 30 - 200',
    },
    r: {
      functionName: 'Linear Transformation Rules',
      genericSyntax: 'E_Y <- a * E_X + b; Var_Y <- a^2 * Var_X; SD_Y <- abs(a) * SD_X',
      argumentDescriptions: [
        { name: 'a', description: 'Scaling multiplier constant.' },
        { name: 'b', description: 'Location shift constant.' },
      ],
      notes: 'SD scales linearly (|a| * SD), while Variance scales quadratically (a^2 * Var).',
      commonPitfalls: 'Confusing 5 independent random days (Var = 5 * Var(X)) with 5 * one single day (Var = 25 * Var(X)).',
      exampleFormula: '50 * 30 - 200',
    },
  },
  {
    id: 'm2_linear_comb_sums',
    familyId: 'linear_combination',
    moduleId: 'module_2',
    title: 'Sum and Difference of Independent Random Variables',
    statisticalTask: 'Calculate E[X +/- Y] = E[X] +/- E[Y] and Var(X +/- Y) = Var(X) + Var(Y).',
    relevantTemplateIds: ['lincomb_total_two_product_lines', 'lincomb_difference_independent', 'linear_comb_sum_variance_independent'],
    applicableKeywords: ['independent sums', 'difference of variables', 'variances always add'],
    excel: {
      functionName: 'Var(X +/- Y) = Var(X) + Var(Y)',
      genericSyntax: 'Mean Diff: =E_X - E_Y | Variance: =Var_X + Var_Y | SD: =SQRT(Var_X + Var_Y)',
      argumentDescriptions: [
        { name: 'Var_X', description: 'Variance of independent variable X.' },
        { name: 'Var_Y', description: 'Variance of independent variable Y.' },
      ],
      notes: 'For independent variables, VARIANCES ALWAYS ADD, even when taking the difference (X - Y)!',
      commonPitfalls: 'Subtracting variances for difference Var(X - Y) = Var(X) - Var(Y). Variance cannot decrease when subtracting noisy variables!',
      exampleFormula: '=SQRT(160 + 90)',
    },
    r: {
      functionName: 'Independent Sums / Differences',
      genericSyntax: 'E_Diff <- E_X - E_Y; Var_Diff <- Var_X + Var_Y; SD_Diff <- sqrt(Var_X + Var_Y)',
      argumentDescriptions: [
        { name: 'Var_X', description: 'Variance of random variable X.' },
        { name: 'Var_Y', description: 'Variance of random variable Y.' },
      ],
      notes: 'Adding standard deviations directly is invalid (SD(X + Y) != SD(X) + SD(Y)).',
      commonPitfalls: 'Attempting to calculate SD by adding SD(X) + SD(Y) instead of sqrt(Var(X) + Var(Y)).',
      exampleFormula: 'sqrt(var_A + var_B)',
    },
  },
  {
    id: 'm2_portfolio_covariance',
    familyId: 'linear_combination',
    moduleId: 'module_2',
    title: 'Correlated Variables & Portfolio Variance',
    statisticalTask: 'Calculate Var(aX + bY) = a^2*Var(X) + b^2*Var(Y) + 2*a*b*Cov(X, Y).',
    relevantTemplateIds: ['lincomb_portfolio_correlated'],
    applicableKeywords: ['portfolio', 'covariance', 'correlation', 'diversification'],
    excel: {
      functionName: 'Portfolio Variance',
      genericSyntax: '=w1^2 * Var1 + w2^2 * Var2 + 2 * w1 * w2 * Cov12',
      argumentDescriptions: [
        { name: 'w1, w2', description: 'Portfolio weights assigned to assets 1 and 2.' },
        { name: 'Var1, Var2', description: 'Variances of asset 1 and asset 2.' },
        { name: 'Cov12', description: 'Covariance between assets (= Corr * SD1 * SD2).' },
      ],
      notes: 'If correlation is negative, the covariance term reduces portfolio variance below either asset (diversification).',
      commonPitfalls: 'Omitting the 2*w1*w2*Cov term assumes independence, which misstates portfolio risk.',
      exampleFormula: '=0.5^2 * 100 + 0.5^2 * 64 + 2 * 0.5 * 0.5 * (-20)',
    },
    r: {
      functionName: 'Portfolio Variance Formula',
      genericSyntax: 'w1^2 * var1 + w2^2 * var2 + 2 * w1 * w2 * cov12',
      argumentDescriptions: [
        { name: 'w1, w2', description: 'Asset weights.' },
        { name: 'cov12', description: 'Covariance between assets.' },
      ],
      notes: 'Can also be written in matrix notation: t(w) %*% Sigma %*% w.',
      commonPitfalls: 'Plugging in standard deviations instead of variances into the squared weight terms.',
      exampleFormula: '0.6^2 * 144 + 0.4^2 * 81 + 2 * 0.6 * 0.4 * 30',
    },
  },
  {
    id: 'm2_binom_exact_pmf',
    familyId: 'binomial_pmf',
    moduleId: 'module_2',
    title: 'Binomial Distribution: Exact Point Probability (PMF)',
    statisticalTask: 'Calculate P(X = k) for n independent Bernoulli trials with constant probability p.',
    relevantTemplateIds: ['binomial_fair_coin_exactly_two', 'binomial_service_calls_zero'],
    applicableKeywords: ['binomial', 'exact', 'pmf', 'trials', 'coin toss'],
    excel: {
      functionName: 'BINOM.DIST (FALSE)',
      genericSyntax: '=BINOM.DIST(number_s, trials, prob_s, FALSE)',
      argumentDescriptions: [
        { name: 'number_s', description: 'Target number of successes k.' },
        { name: 'trials', description: 'Total number of independent trials n.' },
        { name: 'prob_s', description: 'Constant probability of success on each trial p.' },
        { name: 'FALSE', description: 'Logical flag: FALSE computes exact probability mass P(X = k).' },
      ],
      notes: 'Returns (n choose k) * p^k * (1 - p)^(n - k).',
      commonPitfalls: 'Passing TRUE instead of FALSE, which returns cumulative sum P(X <= k) instead of point P(X = k).',
      exampleFormula: '=BINOM.DIST(2, 4, 0.5, FALSE)',
    },
    r: {
      functionName: 'dbinom',
      genericSyntax: 'dbinom(x, size, prob)',
      argumentDescriptions: [
        { name: 'x', description: 'Target number of successes k.' },
        { name: 'size', description: 'Total number of independent trials n.' },
        { name: 'prob', description: 'Success probability per trial p.' },
      ],
      notes: 'd stands for density/mass function (PMF).',
      commonPitfalls: 'Using pbinom() instead of dbinom() when an exact point probability is requested.',
      exampleFormula: 'dbinom(2, size = 4, prob = 0.5)',
    },
  },
  {
    id: 'm2_binom_cumulative_cdf',
    familyId: 'binomial_pmf',
    moduleId: 'module_2',
    title: 'Binomial Distribution: Cumulative and Tail Probabilities',
    statisticalTask: 'Compute P(X <= k) or tail probabilities P(X >= k) = 1 - P(X <= k - 1).',
    relevantTemplateIds: ['binomial_at_least_three_defects', 'binomial_defective_at_most_one'],
    applicableKeywords: ['binomial', 'cumulative', 'at least', 'at most', 'tail', 'cdf'],
    excel: {
      functionName: 'BINOM.DIST (TRUE)',
      genericSyntax: 'At most k: =BINOM.DIST(k, n, p, TRUE) | At least k: =1 - BINOM.DIST(k - 1, n, p, TRUE)',
      argumentDescriptions: [
        { name: 'k', description: 'Cutoff number of successes.' },
        { name: 'n', description: 'Number of independent trials.' },
        { name: 'p', description: 'Success probability per trial.' },
        { name: 'TRUE', description: 'Logical flag: TRUE returns cumulative distribution function P(X <= k).' },
      ],
      notes: 'For discrete distributions, "at least k" requires subtracting k - 1 so that k remains included.',
      commonPitfalls: 'Off-by-one boundary error: 1 - BINOM.DIST(k, n, p, TRUE) erroneously computes P(X >= k + 1).',
      exampleFormula: '=1 - BINOM.DIST(2, 12, 0.15, TRUE)',
    },
    r: {
      functionName: 'pbinom',
      genericSyntax: 'At most k: pbinom(k, size, prob) | At least k: pbinom(k - 1, size, prob, lower.tail = FALSE)',
      argumentDescriptions: [
        { name: 'q', description: 'Quantile cutoff value.' },
        { name: 'size', description: 'Total trials n.' },
        { name: 'prob', description: 'Success probability p.' },
        { name: 'lower.tail', description: 'If TRUE (default), P(X <= q). If FALSE, P(X > q).' },
      ],
      notes: 'p stands for probability distribution function (CDF). Setting lower.tail = FALSE computes P(X > q).',
      commonPitfalls: 'For P(X >= k) with lower.tail = FALSE, pass k - 1 as q, because P(X > k - 1) = P(X >= k).',
      exampleFormula: 'pbinom(2, size = 12, prob = 0.15, lower.tail = FALSE)',
    },
  },
  {
    id: 'm2_binom_parameters',
    familyId: 'binomial_pmf',
    moduleId: 'module_2',
    title: 'Binomial Distribution Parameters: Mean, Variance & SD',
    statisticalTask: 'Calculate mu = n*p, sigma^2 = n*p*(1 - p), and sigma = sqrt(n*p*(1 - p)).',
    relevantTemplateIds: ['binomial_mean_sd_shipments', 'binomial_condition_violated_small_lot', 'choose_binomial_or_poisson'],
    applicableKeywords: ['binomial parameters', 'binomial mean', 'binomial standard deviation', 'hypergeometric check'],
    excel: {
      functionName: 'SQRT(n * p * (1 - p))',
      genericSyntax: 'Mean: =n * p | Variance: =n * p * (1 - p) | SD: =SQRT(n * p * (1 - p))',
      argumentDescriptions: [
        { name: 'n', description: 'Number of independent trials.' },
        { name: 'p', description: 'Constant probability of success.' },
      ],
      notes: 'Check 5% rule: If sampling without replacement from population N, n/N <= 0.05 is required for binomial.',
      commonPitfalls: 'Confusing binomial standard deviation with Poisson standard deviation (sqrt(lambda)).',
      exampleFormula: '=SQRT(200 * 0.02 * (1 - 0.02))',
    },
    r: {
      functionName: 'n * p & sqrt(n * p * (1 - p))',
      genericSyntax: 'mu <- n * p; sigma <- sqrt(n * p * (1 - p))',
      argumentDescriptions: [
        { name: 'n', description: 'Trial count.' },
        { name: 'p', description: 'Success rate.' },
      ],
      notes: 'Binomial variance is maximized when p = 0.5 (Var = 0.25 * n).',
      commonPitfalls: 'Applying binomial formulas when trials are dependent (e.g. small lot sampling without replacement).',
      exampleFormula: 'sqrt(200 * 0.02 * 0.98)',
    },
  },
  {
    id: 'm2_poisson_exact_pmf',
    familyId: 'poisson_pmf',
    moduleId: 'module_2',
    title: 'Poisson Distribution: Exact Point Probability',
    statisticalTask: 'Compute P(X = k) = (lambda^k * e^(-lambda)) / k! for rare arrival events.',
    relevantTemplateIds: ['poisson_rare_event_one', 'poisson_server_failures_zero', 'poisson_call_center_arrivals'],
    applicableKeywords: ['poisson', 'exact', 'rate', 'lambda', 'arrivals'],
    excel: {
      functionName: 'POISSON.DIST (FALSE)',
      genericSyntax: '=POISSON.DIST(x, mean, FALSE)',
      argumentDescriptions: [
        { name: 'x', description: 'Number of occurrences or arrivals k.' },
        { name: 'mean', description: 'Expected average arrival rate lambda over the interval.' },
        { name: 'FALSE', description: 'Logical flag: FALSE computes exact probability P(X = x).' },
      ],
      notes: 'For Poisson distributions, Mean = lambda and Variance = lambda.',
      commonPitfalls: 'Specifying TRUE when an exact occurrence probability P(X = k) is required.',
      exampleFormula: '=POISSON.DIST(1, 1.8, FALSE)',
    },
    r: {
      functionName: 'dpois',
      genericSyntax: 'dpois(x, lambda)',
      argumentDescriptions: [
        { name: 'x', description: 'Number of events k.' },
        { name: 'lambda', description: 'Mean arrival rate parameter.' },
      ],
      notes: 'd stands for discrete density/mass.',
      commonPitfalls: 'Passing the rate per hour when the question asks about a 15-minute interval without rescaling.',
      exampleFormula: 'dpois(1, lambda = 1.8)',
    },
  },
  {
    id: 'm2_poisson_cumulative_cdf',
    familyId: 'poisson_pmf',
    moduleId: 'module_2',
    title: 'Poisson Distribution: Cumulative & Rate Rescaling',
    statisticalTask: 'Rescale arrival rates and compute cumulative tail probabilities P(X <= k) or P(X >= k).',
    relevantTemplateIds: ['poisson_rate_rescaling_quarter_hour'],
    applicableKeywords: ['poisson', 'cumulative', 'rate rescaling', 'at most', 'at least'],
    excel: {
      functionName: 'POISSON.DIST (TRUE)',
      genericSyntax: 'At most: =POISSON.DIST(k, lambda_scaled, TRUE) | At least: =1 - POISSON.DIST(k - 1, lambda_scaled, TRUE)',
      argumentDescriptions: [
        { name: 'k', description: 'Event count threshold.' },
        { name: 'lambda_scaled', description: 'Rescaled mean rate = lambda_base * (target_time / base_time).' },
        { name: 'TRUE', description: 'Logical flag: TRUE returns cumulative probability P(X <= k).' },
      ],
      notes: 'Poisson rate parameter lambda scales proportionally with the length of the continuous interval.',
      commonPitfalls: 'Forgetting to rescale lambda (e.g. using 6 calls/hr instead of 1.5 calls per 15 min).',
      exampleFormula: '=POISSON.DIST(1, 6.0 * (15 / 60), TRUE)',
    },
    r: {
      functionName: 'ppois',
      genericSyntax: 'ppois(q, lambda = lambda_base * (target_t / base_t))',
      argumentDescriptions: [
        { name: 'q', description: 'Quantile cutoff.' },
        { name: 'lambda', description: 'Rescaled interval arrival rate.' },
      ],
      notes: 'To compute P(X >= k), use ppois(k - 1, lambda, lower.tail = FALSE).',
      commonPitfalls: 'Subtracting from k instead of k - 1 when computing upper tail.',
      exampleFormula: 'ppois(1, lambda = 1.5)',
    },
  },

  // ==========================================
  // MODULE 3: CONTINUOUS DISTRIBUTIONS
  // ==========================================
  {
    id: 'm3_normal_standard_z',
    familyId: 'normal_dist',
    moduleId: 'module_3',
    title: 'Standard Normal Distribution Z ~ N(0, 1)',
    statisticalTask: 'Compute cumulative probability P(Z <= z), tail P(Z > z), and inverse critical cutoff z*.',
    relevantTemplateIds: ['normal_standard_upper_tail', 'normal_compare_standardized_scores'],
    applicableKeywords: ['standard normal', 'z distribution', 'z critical', 'norm.s.dist', 'norm.s.inv'],
    excel: {
      functionName: 'NORM.S.DIST, NORM.S.INV',
      genericSyntax: 'P(Z <= z): =NORM.S.DIST(z, TRUE) | Cutoff z*: =NORM.S.INV(prob)',
      argumentDescriptions: [
        { name: 'z', description: 'Standardized z-score observation value.' },
        { name: 'TRUE', description: 'Cumulative probability flag (always TRUE for normal area).' },
        { name: 'prob', description: 'Left-tail cumulative probability area in [0, 1].' },
      ],
      notes: 'Upper tail is =1 - NORM.S.DIST(z, TRUE). Symmetric two-tailed critical value is =NORM.S.INV(1 - alpha/2).',
      commonPitfalls: 'Passing FALSE for the cumulative argument in NORM.S.DIST returns the bell curve height, not a probability!',
      exampleFormula: '=1 - NORM.S.DIST(1.645, TRUE)',
    },
    r: {
      functionName: 'pnorm, qnorm (Standard)',
      genericSyntax: 'P(Z <= z): pnorm(z) | Tail P(Z > z): pnorm(z, lower.tail = FALSE) | Cutoff: qnorm(p)',
      argumentDescriptions: [
        { name: 'z', description: 'Standard normal quantile.' },
        { name: 'p', description: 'Cumulative probability area.' },
      ],
      notes: 'pnorm defaults to mean = 0, sd = 1. qnorm is the inverse quantile function.',
      commonPitfalls: 'Specifying percentages as integers (e.g. 95 instead of 0.95) into qnorm.',
      exampleFormula: '1 - pnorm(1.645)',
    },
  },
  {
    id: 'm3_normal_general_x',
    familyId: 'normal_dist',
    moduleId: 'module_3',
    title: 'General Normal Distribution X ~ N(mu, sigma)',
    statisticalTask: 'Calculate cumulative area P(X <= x) and inverse quantile cutoffs without manual z-scores.',
    relevantTemplateIds: ['normal_inventory_service_level', 'normal_inverse_bonus_cutoff', 'normal_warranty_lower_percentile', 'normal_empirical_vs_exact'],
    applicableKeywords: ['general normal', 'reorder point', 'percentile cutoff', 'norm.dist', 'norm.inv'],
    excel: {
      functionName: 'NORM.DIST, NORM.INV',
      genericSyntax: 'P(X <= x): =NORM.DIST(x, mean, sd, TRUE) | Cutoff x: =NORM.INV(prob, mean, sd)',
      argumentDescriptions: [
        { name: 'x', description: 'Observation value in original measurement units.' },
        { name: 'mean', description: 'Population distribution mean mu.' },
        { name: 'sd', description: 'Population distribution standard deviation sigma.' },
        { name: 'prob', description: 'Target cumulative probability percentile (e.g. 0.95).' },
      ],
      notes: 'Reorder Point under safety stock: =NORM.INV(service_level, mean_demand, sd_demand).',
      commonPitfalls: 'Supplying variance instead of standard deviation into sd argument.',
      exampleFormula: '=NORM.INV(0.95, 500, 50)',
    },
    r: {
      functionName: 'pnorm, qnorm (General)',
      genericSyntax: 'P(X <= x): pnorm(q, mean, sd) | Cutoff x: qnorm(p, mean, sd)',
      argumentDescriptions: [
        { name: 'q', description: 'Observation quantile.' },
        { name: 'mean', description: 'Distribution mean mu.' },
        { name: 'sd', description: 'Distribution standard deviation sigma.' },
        { name: 'p', description: 'Cumulative probability.' },
      ],
      notes: 'qnorm(0.95, mean = 500, sd = 50) directly gives the 95th percentile.',
      commonPitfalls: 'Passing variance instead of standard deviation.',
      exampleFormula: 'qnorm(0.95, mean = 500, sd = 50)',
    },
  },
  {
    id: 'm3_normal_interval',
    familyId: 'normal_dist',
    moduleId: 'module_3',
    title: 'Normal Interval Probability P(a <= X <= b)',
    statisticalTask: 'Compute probability that a normally distributed variable falls between two bounds.',
    relevantTemplateIds: ['normal_between_package_weight'],
    applicableKeywords: ['normal interval', 'between bounds', 'middle area'],
    excel: {
      functionName: 'NORM.DIST Interval',
      genericSyntax: '=NORM.DIST(b, mean, sd, TRUE) - NORM.DIST(a, mean, sd, TRUE)',
      argumentDescriptions: [
        { name: 'b', description: 'Upper interval boundary.' },
        { name: 'a', description: 'Lower interval boundary.' },
        { name: 'mean', description: 'Distribution mean mu.' },
        { name: 'sd', description: 'Distribution standard deviation sigma.' },
      ],
      notes: 'For continuous distributions, P(a <= X <= b) = P(X <= b) - P(X <= a) regardless of strict or non-strict inequalities.',
      commonPitfalls: 'Subtracting upper from lower bound instead of lower from upper (yielding negative probability).',
      exampleFormula: '=NORM.DIST(10.5, 10, 0.5, TRUE) - NORM.DIST(9.5, 10, 0.5, TRUE)',
    },
    r: {
      functionName: 'pnorm Interval',
      genericSyntax: 'pnorm(b, mean, sd) - pnorm(a, mean, sd) | diff(pnorm(c(a, b), mean, sd))',
      argumentDescriptions: [
        { name: 'a, b', description: 'Lower and upper bounds.' },
        { name: 'mean, sd', description: 'Distribution parameters.' },
      ],
      notes: 'diff(pnorm(c(a, b), mean, sd)) evaluates both bounds and subtracts cleanly.',
      commonPitfalls: 'Forgetting that continuous point probabilities P(X = a) = 0.',
      exampleFormula: 'pnorm(10.5, 10, 0.5) - pnorm(9.5, 10, 0.5)',
    },
  },
  {
    id: 'm3_uniform_interval',
    familyId: 'uniform_dist',
    moduleId: 'module_3',
    title: 'Continuous Uniform Distribution U(a, b)',
    statisticalTask: 'Calculate constant density height f(x) = 1/(b - a) and interval probability (d - c)/(b - a).',
    relevantTemplateIds: ['uniform_commute_wait_time', 'contvdisc_area_is_probability'],
    applicableKeywords: ['uniform', 'continuous uniform', 'rectangle', 'density height', 'wait time'],
    excel: {
      functionName: 'Uniform Interval Formula',
      genericSyntax: 'Height: =1 / (b - a) | Area P(c <= X <= d): =(d - c) / (b - a)',
      argumentDescriptions: [
        { name: 'a, b', description: 'Minimum and maximum domain limits of the uniform distribution.' },
        { name: 'c, d', description: 'Sub-interval bounds within [a, b].' },
      ],
      notes: 'Mean is (a + b)/2; Variance is (b - a)^2 / 12.',
      commonPitfalls: 'Dividing by the interval width (d - c) rather than total domain range (b - a).',
      exampleFormula: '=(12 - 5) / (20 - 0)',
    },
    r: {
      functionName: 'punif, qunif',
      genericSyntax: 'P(c <= X <= d): punif(d, min, max) - punif(c, min, max) | Cutoff: qunif(p, min, max)',
      argumentDescriptions: [
        { name: 'min, max', description: 'Lower (a) and upper (b) bounds of uniform distribution.' },
        { name: 'p', description: 'Cumulative probability percentile.' },
      ],
      notes: 'dunif(x, min, max) gives constant density 1 / (max - min).',
      commonPitfalls: 'Querying values outside [min, max] where punif clamps to 0 or 1.',
      exampleFormula: 'punif(12, 0, 20) - punif(5, 0, 20)',
    },
  },
  {
    id: 'm3_exponential_cdf',
    familyId: 'exponential_dist',
    moduleId: 'module_3',
    title: 'Exponential Distribution Cumulative & Memoryless Property',
    statisticalTask: 'Calculate P(X <= x) = 1 - e^(-lambda * x) and upper tail P(X > x) = e^(-lambda * x).',
    relevantTemplateIds: ['exponential_arrival_within_two', 'exponential_between_two_times', 'exponential_memoryless_wait', 'exponential_percentile_service', 'exponential_poisson_link', 'exponential_rate_vs_mean_param', 'exponential_service_lifetime'],
    applicableKeywords: ['exponential', 'memoryless', 'service time', 'arrival time', 'expon.dist'],
    excel: {
      functionName: 'EXPON.DIST',
      genericSyntax: 'P(X <= x): =EXPON.DIST(x, lambda, TRUE) | Upper Tail: =EXP(-lambda * x)',
      argumentDescriptions: [
        { name: 'x', description: 'Elapsed time or distance threshold.' },
        { name: 'lambda', description: 'Rate parameter lambda (events per unit time).' },
        { name: 'TRUE', description: 'Cumulative probability flag.' },
      ],
      notes: 'If given mean time mu (e.g. 5 min), rate lambda = 1 / mu = 0.20 per min. Memoryless: P(X > s + t | X > s) = P(X > t).',
      commonPitfalls: 'Plugging the mean time mu directly into the lambda parameter instead of lambda = 1 / mu.',
      exampleFormula: '=EXPON.DIST(2, 0.5, TRUE)',
    },
    r: {
      functionName: 'pexp, qexp',
      genericSyntax: 'P(X <= x): pexp(q, rate = lambda) | Tail P(X > x): pexp(q, rate = lambda, lower.tail = FALSE)',
      argumentDescriptions: [
        { name: 'q', description: 'Time cutoff.' },
        { name: 'rate', description: 'Arrival rate parameter lambda = 1 / mean.' },
      ],
      notes: 'Quantile cutoff (percentile) is qexp(p, rate = lambda).',
      commonPitfalls: 'Passing mean time instead of rate parameter into rate argument.',
      exampleFormula: 'pexp(2, rate = 0.5)',
    },
  },
  {
    id: 'm3_exponential_tail_quantile',
    familyId: 'exponential_dist',
    moduleId: 'module_3',
    title: 'Exponential Distribution: Upper Tail & Quantile Percentile Cutoff',
    statisticalTask: 'Compute survival tail P(X > x) = e^(-lambda*x) and inverse percentile cutoff x_p = -ln(1 - p) / lambda.',
    relevantTemplateIds: ['exponential_percentile_service'],
    applicableKeywords: ['exponential percentile', 'exponential quantile', 'sla cutoff', 'upper tail', 'qexp'],
    excel: {
      functionName: 'EXP Upper Tail & Quantile Formula',
      genericSyntax: 'Upper Tail P(X > x): =EXP(-lambda * x) | Percentile Cutoff: =-LN(1 - p) / lambda',
      argumentDescriptions: [
        { name: 'lambda', description: 'Rate parameter lambda = 1 / mean.' },
        { name: 'x', description: 'Elapsed time threshold.' },
        { name: 'p', description: 'Target cumulative percentile probability (e.g. 0.90 for 90th percentile SLA).' },
      ],
      notes: 'Excel lacks a built-in EXPON.INV function, so solving the CDF 1 - e^(-lambda*x) = p yields x = -LN(1 - p) / lambda.',
      commonPitfalls: 'Forgetting the negative sign in -LN(1 - p), or forgetting that LN of a fraction is negative.',
      exampleFormula: '=-LN(1 - 0.90) / 0.5',
    },
    r: {
      functionName: 'qexp, pexp(lower.tail = FALSE)',
      genericSyntax: 'Tail P(X > x): pexp(x, rate = lambda, lower.tail = FALSE) | Cutoff: qexp(p, rate = lambda)',
      argumentDescriptions: [
        { name: 'p', description: 'Percentile probability.' },
        { name: 'rate', description: 'Rate parameter lambda = 1 / mean.' },
      ],
      notes: 'qexp(0.90, rate = lambda) directly calculates the 90th percentile resolution cutoff.',
      commonPitfalls: 'Passing mean duration directly instead of rate parameter lambda = 1 / mean into qexp().',
      exampleFormula: 'qexp(0.90, rate = 0.5)',
    },
  },
  {
    id: 'm3_concept_continuous_vs_discrete',
    familyId: 'concept_check',
    moduleId: 'module_3',
    title: 'Continuous vs. Discrete Probability Axioms',
    statisticalTask: 'Identify properties: point probability P(X = c) = 0, density vs mass, integration.',
    relevantTemplateIds: ['contvdisc_classify_variables', 'contvdisc_match_model', 'contvdisc_point_probability_zero', 'contvdisc_strict_vs_nonstrict'],
    applicableKeywords: ['continuous vs discrete', 'point probability zero', 'strict inequality', 'density curve'],
    excel: {
      functionName: 'Continuous Properties',
      genericSyntax: 'For continuous X: P(X = c) = 0 | P(X <= c) = P(X < c)',
      argumentDescriptions: [
        { name: 'c', description: 'Any individual numerical point value.' },
      ],
      notes: 'In continuous distributions, area under a single point is zero; probabilities exist only over intervals.',
      commonPitfalls: 'Treating continuous distributions like discrete ones by expecting P(X = c) > 0.',
      exampleFormula: '=0',
    },
    r: {
      functionName: 'Continuous Concepts',
      genericSyntax: 'integrate(f, lower = a, upper = b)$value',
      argumentDescriptions: [
        { name: 'f', description: 'Probability density function (PDF).' },
      ],
      notes: 'Strict vs non-strict inequalities are identical in continuous distributions: P(X < c) == P(X <= c).',
      commonPitfalls: 'Attempting to sum point probabilities for continuous variables.',
      exampleFormula: 'integrate(dnorm, -1.96, 1.96)$value',
    },
  },

  // ==========================================
  // MODULE 4: SAMPLING DISTRIBUTIONS & CIS
  // ==========================================
  {
    id: 'm4_clt_sampling_dist_mean',
    familyId: 'sampling_distribution',
    moduleId: 'module_4',
    title: 'Central Limit Theorem: Sampling Distribution of the Mean',
    statisticalTask: 'Calculate standard error SE = sigma / sqrt(n) and sample mean probabilities.',
    relevantTemplateIds: ['clt_mean_order_value', 'clt_individual_vs_mean', 'clt_finite_population_correction', 'clt_skewed_service_times', 'clt_small_n_not_justified'],
    applicableKeywords: ['central limit theorem', 'clt', 'sampling distribution', 'standard error', 'sample mean'],
    excel: {
      functionName: 'NORM.DIST with Standard Error',
      genericSyntax: '=NORM.DIST(x_bar, mu, sigma / SQRT(n), TRUE)',
      argumentDescriptions: [
        { name: 'x_bar', description: 'Observed sample mean cutoff.' },
        { name: 'mu', description: 'Population mean.' },
        { name: 'sigma', description: 'Population standard deviation.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 'Standard error shrinks by factor of sqrt(n). If sampling from finite population N, multiply by FPC = SQRT((N - n)/(N - 1)).',
      commonPitfalls: 'Using population sigma instead of standard error sigma / SQRT(n) when finding probability for a sample mean.',
      exampleFormula: '=NORM.DIST(102, 100, 10 / SQRT(36), TRUE)',
    },
    r: {
      functionName: 'pnorm with SE',
      genericSyntax: 'pnorm(x_bar, mean = mu, sd = sigma / sqrt(n))',
      argumentDescriptions: [
        { name: 'x_bar', description: 'Sample mean value.' },
        { name: 'mu', description: 'Population mean.' },
        { name: 'sigma', description: 'Population standard deviation.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 'By CLT, sampling distribution approaches normal as n >= 30 regardless of parent population shape.',
      commonPitfalls: 'Applying CLT to individual observations rather than sample averages.',
      exampleFormula: 'pnorm(102, mean = 100, sd = 10 / sqrt(36))',
    },
  },
  {
    id: 'm4_clt_sampling_dist_prop',
    familyId: 'sampling_distribution',
    moduleId: 'module_4',
    title: 'Sampling Distribution of the Proportion',
    statisticalTask: 'Calculate SE = sqrt(p * (1 - p) / n) and sample proportion probabilities.',
    relevantTemplateIds: ['clt_proportion_coupon_users', 'se_proportion_depends_on_p'],
    applicableKeywords: ['proportion sampling distribution', 'sample proportion', 'se of proportion', 'p_hat'],
    excel: {
      functionName: 'NORM.DIST for Proportion',
      genericSyntax: '=NORM.DIST(p_hat, p, SQRT(p * (1 - p) / n), TRUE)',
      argumentDescriptions: [
        { name: 'p_hat', description: 'Observed sample proportion cutoff.' },
        { name: 'p', description: 'True population proportion parameter.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 'Requires normal approximation conditions: n*p >= 10 and n*(1 - p) >= 10.',
      commonPitfalls: 'Forgetting the square root around n in the standard error denominator.',
      exampleFormula: '=NORM.DIST(0.35, 0.30, SQRT(0.30 * 0.70 / 200), TRUE)',
    },
    r: {
      functionName: 'pnorm for Proportion',
      genericSyntax: 'pnorm(p_hat, mean = p, sd = sqrt(p * (1 - p) / n))',
      argumentDescriptions: [
        { name: 'p_hat', description: 'Sample proportion cutoff.' },
        { name: 'p', description: 'Population proportion.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 'Standard error of proportion is maximized when p = 0.50.',
      commonPitfalls: 'Using sample p_hat in the standard error when the true population parameter p is known.',
      exampleFormula: 'pnorm(0.35, mean = 0.30, sd = sqrt(0.30 * 0.70 / 200))',
    },
  },
  {
    id: 'm4_ci_mean_z',
    familyId: 'confidence_interval',
    moduleId: 'module_4',
    title: 'Confidence Interval for Mean: Z-Interval (Known Sigma)',
    statisticalTask: 'Calculate margin of error E = z* * sigma / sqrt(n) and interval [x_bar - E, x_bar + E].',
    relevantTemplateIds: ['ci_mean_z_known_sigma', 'ci_level_vs_width', 'ci_plausible_claim', 'ci_interpretation_statement'],
    applicableKeywords: ['confidence interval', 'z interval', 'known sigma', 'confidence.norm', 'margin of error'],
    excel: {
      functionName: 'CONFIDENCE.NORM',
      genericSyntax: 'Margin of Error: =CONFIDENCE.NORM(alpha, standard_dev, size)',
      argumentDescriptions: [
        { name: 'alpha', description: 'Significance level = 1 - confidence_level (e.g. 0.05 for 95% CI).' },
        { name: 'standard_dev', description: 'Known population standard deviation sigma.' },
        { name: 'size', description: 'Sample size n.' },
      ],
      notes: 'Interval is [x_bar - CONFIDENCE.NORM(...), x_bar + CONFIDENCE.NORM(...)]. Critical value is =NORM.S.INV(1 - alpha/2).',
      commonPitfalls: 'Entering confidence level (e.g. 0.95) into alpha instead of 1 - 0.95 = 0.05.',
      exampleFormula: '=CONFIDENCE.NORM(0.05, 12, 49)',
    },
    r: {
      functionName: 'qnorm Margin of Error',
      genericSyntax: 'E <- qnorm(1 - alpha / 2) * (sigma / sqrt(n)); CI <- c(x_bar - E, x_bar + E)',
      argumentDescriptions: [
        { name: 'alpha', description: 'Significance level (e.g. 0.05).' },
        { name: 'sigma', description: 'Known population standard deviation.' },
        { name: 'n', description: 'Sample size.' },
        { name: 'x_bar', description: 'Sample mean.' },
      ],
      notes: 'For 95% confidence, qnorm(0.975) = 1.95996 (~1.960).',
      commonPitfalls: 'Interpreting the interval as "a 95% probability that mu is inside"; mu is a fixed constant, not random.',
      exampleFormula: 'qnorm(0.975) * (12 / sqrt(49))',
    },
  },
  {
    id: 'm4_ci_mean_t',
    familyId: 'confidence_interval',
    moduleId: 'module_4',
    title: 'Confidence Interval for Mean: t-Interval (Unknown Sigma)',
    statisticalTask: 'Calculate margin of error E = t* * s / sqrt(n) with df = n - 1 degrees of freedom.',
    relevantTemplateIds: ['ci_mean_t_small_sample_90', 'ci_mean_sample_estimate'],
    applicableKeywords: ['t interval', 'unknown sigma', 'student t', 'confidence.t', 't.inv.2t'],
    excel: {
      functionName: 'CONFIDENCE.T, T.INV.2T',
      genericSyntax: 'Margin: =CONFIDENCE.T(alpha, s, n) | Critical t*: =T.INV.2T(alpha, n - 1)',
      argumentDescriptions: [
        { name: 'alpha', description: 'Significance level = 1 - confidence_level (e.g. 0.10 for 90% CI).' },
        { name: 's', description: 'Sample standard deviation.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 'CONFIDENCE.T takes sample size n; T.INV.2T takes degrees of freedom df = n - 1.',
      commonPitfalls: 'Passing n instead of n - 1 to T.INV.2T degrees of freedom.',
      exampleFormula: '=CONFIDENCE.T(0.10, 5.1, 10)',
    },
    r: {
      functionName: 'qt, t.test',
      genericSyntax: 'E <- qt(1 - alpha / 2, df = n - 1) * (s / sqrt(n)) | t.test(x, conf.level = 1 - alpha)$conf.int',
      argumentDescriptions: [
        { name: 'alpha', description: 'Significance level.' },
        { name: 's', description: 'Sample standard deviation.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 't.test(data_vector, conf.level = 0.95)$conf.int computes the entire t-interval directly.',
      commonPitfalls: 'Using qnorm instead of qt when population sigma is unknown, which yields an artificially narrow interval.',
      exampleFormula: 'qt(0.95, df = 9) * (5.1 / sqrt(10))',
    },
  },
  {
    id: 'm4_ci_proportion',
    familyId: 'confidence_interval',
    moduleId: 'module_4',
    title: 'Confidence Interval for Population Proportion',
    statisticalTask: 'Calculate margin of error E = z* * sqrt(p_hat * (1 - p_hat) / n).',
    relevantTemplateIds: ['ci_proportion_recommend'],
    applicableKeywords: ['proportion confidence interval', 'sample proportion', 'margin of error proportion'],
    excel: {
      functionName: 'Proportion CI Formula',
      genericSyntax: '=NORM.S.INV(1 - alpha / 2) * SQRT(p_hat * (1 - p_hat) / n)',
      argumentDescriptions: [
        { name: 'alpha', description: 'Significance level (e.g. 0.05 for 95% confidence).' },
        { name: 'p_hat', description: 'Observed sample proportion = x / n.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 'Interval bounds are [p_hat - E, p_hat + E].',
      commonPitfalls: 'Using p = 0.50 in the CI standard error instead of the observed sample proportion p_hat.',
      exampleFormula: '=NORM.S.INV(0.975) * SQRT(0.33 * 0.67 / 400)',
    },
    r: {
      functionName: 'prop.test',
      genericSyntax: 'prop.test(x, n, conf.level = 1 - alpha, correct = FALSE)$conf.int',
      argumentDescriptions: [
        { name: 'x', description: 'Number of successes in sample.' },
        { name: 'n', description: 'Total sample size.' },
        { name: 'correct', description: 'Continuity correction (FALSE for standard textbook Wald formula).' },
      ],
      notes: 'Manual formula: E <- qnorm(1 - alpha / 2) * sqrt(p_hat * (1 - p_hat) / n).',
      commonPitfalls: 'Omitting correct = FALSE in prop.test produces Wilson interval with continuity correction.',
      exampleFormula: 'prop.test(132, 400, conf.level = 0.95, correct = FALSE)$conf.int',
    },
  },
  {
    id: 'm4_sample_size_mean',
    familyId: 'sample_size',
    moduleId: 'module_4',
    title: 'Sample Size Determination for Mean Margin of Error',
    statisticalTask: 'Solve for minimal sample size n = (z* * sigma / E)^2 with strict ceiling rounding.',
    relevantTemplateIds: ['n_mean_margin_five', 'n_halve_margin_quadruple', 'n_round_up_misconception', 'se_n_for_target_se', 'se_quadruple_n_halves', 'se_survey_design_compare', 'se_what_n_changes'],
    applicableKeywords: ['sample size', 'sample size for mean', 'roundup', 'ceiling', 'margin of error'],
    excel: {
      functionName: 'ROUNDUP with Z^2 Formula',
      genericSyntax: '=ROUNDUP((NORM.S.INV(1 - alpha / 2) * sigma / E)^2, 0)',
      argumentDescriptions: [
        { name: 'alpha', description: 'Significance level (e.g. 0.05 for 95% confidence).' },
        { name: 'sigma', description: 'Preliminary planning standard deviation.' },
        { name: 'E', description: 'Desired maximum margin of error.' },
      ],
      notes: 'ALWAYS round UP to the next whole integer using ROUNDUP(..., 0) so margin of error does not exceed E.',
      commonPitfalls: 'Rounding down or standard rounding (e.g. 96.04 -> 96), which results in actual margin exceeding target tolerance.',
      exampleFormula: '=ROUNDUP((NORM.S.INV(0.975) * 25 / 5)^2, 0)',
    },
    r: {
      functionName: 'ceiling',
      genericSyntax: 'ceiling((qnorm(1 - alpha / 2) * sigma / E)^2)',
      argumentDescriptions: [
        { name: 'alpha', description: 'Significance level.' },
        { name: 'sigma', description: 'Standard deviation.' },
        { name: 'E', description: 'Desired margin of error.' },
      ],
      notes: 'Inverse-square law: Halving margin of error (E / 2) requires 4x sample size.',
      commonPitfalls: 'Using round() instead of ceiling().',
      exampleFormula: 'ceiling((qnorm(0.975) * 25 / 5)^2)',
    },
  },
  {
    id: 'm4_sample_size_prop',
    familyId: 'sample_size',
    moduleId: 'module_4',
    title: 'Sample Size Determination for Proportion',
    statisticalTask: 'Calculate n = (z*^2 * p* * (1 - p*)) / E^2 using pilot estimate or conservative p* = 0.50.',
    relevantTemplateIds: ['n_proportion_conservative', 'n_proportion_pilot'],
    applicableKeywords: ['sample size proportion', 'conservative planning', 'pilot proportion', 'margin of error'],
    excel: {
      functionName: 'ROUNDUP for Proportion Sample Size',
      genericSyntax: '=ROUNDUP((NORM.S.INV(1 - alpha / 2)^2 * p * (1 - p)) / E^2, 0)',
      argumentDescriptions: [
        { name: 'alpha', description: 'Significance level.' },
        { name: 'p', description: 'Prior estimate p* (or 0.50 for conservative planning maximizing p*(1 - p) = 0.25).' },
        { name: 'E', description: 'Target margin of error (e.g. 0.03 for +/- 3 percentage points).' },
      ],
      notes: 'If no prior estimate is available, using p* = 0.50 guarantees the sample size is sufficient.',
      commonPitfalls: 'Entering percentage margin as integer (e.g. 3 instead of 0.03) into E.',
      exampleFormula: '=ROUNDUP((NORM.S.INV(0.975)^2 * 0.20 * 0.80) / (0.03^2), 0)',
    },
    r: {
      functionName: 'ceiling for Proportion Sample Size',
      genericSyntax: 'ceiling((qnorm(1 - alpha / 2)^2 * p * (1 - p)) / E^2)',
      argumentDescriptions: [
        { name: 'p', description: 'Prior pilot estimate p* (or 0.50).' },
        { name: 'E', description: 'Desired margin of error in decimal form.' },
      ],
      notes: 'Conservative planning with p* = 0.50 yields n = ceiling(1.96^2 * 0.25 / E^2).',
      commonPitfalls: 'Forgetting to square the margin of error E in the denominator.',
      exampleFormula: 'ceiling((qnorm(0.975)^2 * 0.20 * 0.80) / (0.03^2))',
    },
  },

  // ==========================================
  // MODULE 5: HYPOTHESIS TESTING
  // ==========================================
  {
    id: 'm5_hyp_test_mean_z',
    familyId: 'hypothesis_test',
    moduleId: 'module_5',
    title: 'One-Sample Z-Test for Population Mean (Known Sigma)',
    statisticalTask: 'Compute standardized test statistic Z = (x_bar - mu_0) / (sigma / sqrt(n)) and p-value.',
    relevantTemplateIds: ['hyp_z_mean_two_tailed'],
    applicableKeywords: ['z-test', 'one sample z test', 'p-value', 'known sigma', 'hypothesis test'],
    excel: {
      functionName: 'Z Test Statistic & NORM.S.DIST',
      genericSyntax: 'Z: =(x_bar - mu_0) / (sigma / SQRT(n)) | Two-tail p: =2 * (1 - NORM.S.DIST(ABS(z), TRUE))',
      argumentDescriptions: [
        { name: 'x_bar', description: 'Observed sample mean.' },
        { name: 'mu_0', description: 'Hypothesized population mean under null hypothesis H0.' },
        { name: 'sigma', description: 'Known population standard deviation.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 'Left-tailed p-value is =NORM.S.DIST(z, TRUE). Right-tailed is =1 - NORM.S.DIST(z, TRUE).',
      commonPitfalls: 'Forgetting to multiply by 2 for a two-tailed test, or taking 1 - p when z is in the left tail.',
      exampleFormula: '=2 * (1 - NORM.S.DIST(2.40, TRUE))',
    },
    r: {
      functionName: 'Z-Test p-value',
      genericSyntax: 'z <- (x_bar - mu_0) / (sigma / sqrt(n)); p_val <- 2 * pnorm(-abs(z))',
      argumentDescriptions: [
        { name: 'z', description: 'Standardized z test statistic.' },
      ],
      notes: 'pnorm(-abs(z)) computes the lower tail area; multiplying by 2 yields the two-tailed p-value.',
      commonPitfalls: 'Using sample standard deviation s in a z-test instead of t-test.',
      exampleFormula: '2 * pnorm(-abs(2.40))',
    },
  },
  {
    id: 'm5_hyp_test_mean_t',
    familyId: 'hypothesis_test',
    moduleId: 'module_5',
    title: 'One-Sample t-Test for Population Mean (Unknown Sigma)',
    statisticalTask: 'Calculate test statistic t = (x_bar - mu_0) / (s / sqrt(n)) with df = n - 1 and p-value.',
    relevantTemplateIds: ['hyp_t_mean_right_tailed', 'hypothesis_mean_one_sample'],
    applicableKeywords: ['t-test', 'one sample t test', 'p-value', 'degrees of freedom', 't.dist'],
    excel: {
      functionName: 'T.DIST.2T, T.DIST.RT, T.DIST',
      genericSyntax: 't: =(x_bar - mu_0) / (s / SQRT(n)) | Two-tail: =T.DIST.2T(ABS(t), df) | Right: =T.DIST.RT(t, df) | Left: =T.DIST(t, df, TRUE)',
      argumentDescriptions: [
        { name: 't', description: 'Test statistic value.' },
        { name: 'df', description: 'Degrees of freedom n - 1.' },
      ],
      notes: 'T.DIST.2T strictly requires positive |t| argument. T.DIST.RT computes right-tail area for any t.',
      commonPitfalls: 'Passing a negative t to T.DIST.2T yields a #NUM! error; wrap with ABS(t).',
      exampleFormula: '=T.DIST.RT(2.236, 19)',
    },
    r: {
      functionName: 't.test, pt',
      genericSyntax: 't.test(x, mu = mu_0, alternative = "two.sided") | pt(t, df, lower.tail = FALSE)',
      argumentDescriptions: [
        { name: 'x', description: 'Numeric vector of sample data.' },
        { name: 'mu', description: 'Hypothesized mean mu_0.' },
        { name: 'alternative', description: '"two.sided", "greater" (right-tailed), or "less" (left-tailed).' },
      ],
      notes: 't.test() automatically returns t statistic, degrees of freedom, p-value, and confidence interval.',
      commonPitfalls: 'Specifying alternative = "greater" when the alternative is left-tailed.',
      exampleFormula: 'pt(2.236, df = 19, lower.tail = FALSE)',
    },
  },
  {
    id: 'm5_hyp_test_prop_z',
    familyId: 'hypothesis_test',
    moduleId: 'module_5',
    title: 'One-Sample Z-Test for Population Proportion',
    statisticalTask: 'Calculate Z = (p_hat - p_0) / sqrt(p_0 * (1 - p_0) / n) and directional p-value.',
    relevantTemplateIds: ['hyp_proportion_left_tailed', 'hyp_formulate_proportion'],
    applicableKeywords: ['proportion test', 'one sample proportion', 'z test proportion', 'p-value'],
    excel: {
      functionName: 'NORM.S.DIST for Proportion Test',
      genericSyntax: 'Z: =(p_hat - p_0) / SQRT(p_0 * (1 - p_0) / n) | Left-tail p: =NORM.S.DIST(z, TRUE)',
      argumentDescriptions: [
        { name: 'p_hat', description: 'Sample proportion x / n.' },
        { name: 'p_0', description: 'Hypothesized proportion parameter under null H0.' },
        { name: 'n', description: 'Sample size.' },
      ],
      notes: 'CRUCIAL: The standard error under H0 MUST use the hypothesized value p_0, NOT the sample proportion p_hat!',
      commonPitfalls: 'Using p_hat in the denominator standard error under the null hypothesis.',
      exampleFormula: '=NORM.S.DIST((0.14 - 0.20) / SQRT(0.20 * 0.80 / 200), TRUE)',
    },
    r: {
      functionName: 'prop.test (One Sample)',
      genericSyntax: 'prop.test(x, n, p = p_0, alternative = "less", correct = FALSE)',
      argumentDescriptions: [
        { name: 'x', description: 'Count of successes.' },
        { name: 'n', description: 'Sample size.' },
        { name: 'p', description: 'Null hypothesis proportion p_0.' },
      ],
      notes: 'Manual z calculation: z <- (p_hat - p_0) / sqrt(p_0 * (1 - p_0) / n); pnorm(z).',
      commonPitfalls: 'Forgetting correct = FALSE produces a continuity-corrected test statistic that differs slightly from textbook formula.',
      exampleFormula: 'pnorm((0.14 - 0.20) / sqrt(0.20 * 0.80 / 200))',
    },
  },
  {
    id: 'm5_twosample_independent_t',
    familyId: 'two_sample_test',
    moduleId: 'module_5',
    title: 'Two-Sample Independent Pooled t-Test (Equal Variance)',
    statisticalTask: 'Compare independent group means with pooled degrees of freedom df = n1 + n2 - 2.',
    relevantTemplateIds: ['twosample_independent_means_layouts', 'twosample_read_excel_output', 'twosample_choose_design'],
    applicableKeywords: ['two sample t test', 'independent samples', 'pooled variance', 't.test', 'layout'],
    excel: {
      functionName: 'T.TEST (Type 2)',
      genericSyntax: '=T.TEST(array1, array2, tails, 2)',
      argumentDescriptions: [
        { name: 'array1', description: 'Cell range for sample 1.' },
        { name: 'array2', description: 'Cell range for sample 2.' },
        { name: 'tails', description: '1 for one-tailed test, 2 for two-tailed test.' },
        { name: '2', description: 'Type: 2 specifies two-sample equal variance (pooled) t-test.' },
      ],
      notes: 'Excel Data Analysis Toolpak provides "t-Test: Two-Sample Assuming Equal Variances" with full output table.',
      commonPitfalls: 'Specifying Type 1 (paired) or Type 3 (Welch unequal variance) when a pooled independent test is stipulated.',
      exampleFormula: '=T.TEST(A2:A31, B2:B31, 2, 2)',
    },
    r: {
      functionName: 't.test (Independent Pooled)',
      genericSyntax: 't.test(x, y, var.equal = TRUE, alternative = "two.sided")',
      argumentDescriptions: [
        { name: 'x, y', description: 'Vectors of observations for group 1 and group 2.' },
        { name: 'var.equal', description: 'Must set to TRUE for pooled equal variance test (default is FALSE, Welch test).' },
      ],
      notes: 'Pooled variance: s_p^2 = ((n1 - 1)*s1^2 + (n2 - 1)*s2^2) / (n1 + n2 - 2).',
      commonPitfalls: 'Omitting var.equal = TRUE defaults to Welch-Satterthwaite approximation in R.',
      exampleFormula: 't.test(group1, group2, var.equal = TRUE)',
    },
  },
  {
    id: 'm5_twosample_paired_t',
    familyId: 'two_sample_test',
    moduleId: 'module_5',
    title: 'Paired Samples t-Test (Dependent Observations)',
    statisticalTask: 'Analyze before/after or matched pairs by testing mean difference d_bar with df = n - 1.',
    relevantTemplateIds: ['twosample_paired_training'],
    applicableKeywords: ['paired t test', 'dependent samples', 'before after', 'repeated measures'],
    excel: {
      functionName: 'T.TEST (Type 1)',
      genericSyntax: '=T.TEST(array1, array2, tails, 1)',
      argumentDescriptions: [
        { name: 'array1', description: 'Baseline / before measurements.' },
        { name: 'array2', description: 'Post-intervention / after measurements on the same subjects.' },
        { name: 'tails', description: '1 for one-tailed directional test, 2 for two-tailed test.' },
        { name: '1', description: 'Type: 1 specifies paired t-test.' },
      ],
      notes: 'Equivalent to computing d = After - Before in a new column and running a one-sample t-test on d vs 0.',
      commonPitfalls: 'Running an independent two-sample test (Type 2) on paired data, which discards subject-matching variance reduction.',
      exampleFormula: '=T.TEST(B2:B26, C2:C26, 2, 1)',
    },
    r: {
      functionName: 't.test (Paired)',
      genericSyntax: 't.test(after, before, paired = TRUE, alternative = "two.sided")',
      argumentDescriptions: [
        { name: 'after, before', description: 'Paired vectors of equal length.' },
        { name: 'paired', description: 'Set to TRUE for matched pairs analysis.' },
      ],
      notes: 'Degrees of freedom is n_pairs - 1.',
      commonPitfalls: 'Mismatched vector lengths or scrambled row orders that break pair alignment.',
      exampleFormula: 't.test(post_training, pre_training, paired = TRUE)',
    },
  },
  {
    id: 'm5_twosample_prop_z',
    familyId: 'two_sample_test',
    moduleId: 'module_5',
    title: 'Two-Sample Pooled Proportion Z-Test',
    statisticalTask: 'Compare two binomial proportions using pooled proportion p_bar = (x1 + x2) / (n1 + n2).',
    relevantTemplateIds: ['twosample_two_proportion_ads'],
    applicableKeywords: ['two proportion test', 'pooled proportion', 'a/b test', 'conversion test'],
    excel: {
      functionName: 'Two-Proportion Z Formula',
      genericSyntax: 'p_pool: =(x1 + x2)/(n1 + n2) | Z: =(p1 - p2)/SQRT(p_pool*(1 - p_pool)*(1/n1 + 1/n2))',
      argumentDescriptions: [
        { name: 'x1, x2', description: 'Success counts in groups 1 and 2.' },
        { name: 'n1, n2', description: 'Sample sizes for groups 1 and 2.' },
      ],
      notes: 'p-value is =2 * (1 - NORM.S.DIST(ABS(z), TRUE)).',
      commonPitfalls: 'Using separate sample proportions in the standard error rather than pooling under H0: p1 = p2.',
      exampleFormula: '=2 * (1 - NORM.S.DIST(ABS(z), TRUE))',
    },
    r: {
      functionName: 'prop.test (Two Sample)',
      genericSyntax: 'prop.test(c(x1, x2), c(n1, n2), alternative = "two.sided", correct = FALSE)',
      argumentDescriptions: [
        { name: 'c(x1, x2)', description: 'Vector of success counts.' },
        { name: 'c(n1, n2)', description: 'Vector of group sample sizes.' },
      ],
      notes: 'Returns 2-sample test for equality of proportions with chi-squared statistic (where chi^2 = z^2).',
      commonPitfalls: 'Flipping order of groups, which reverses the sign of the estimated difference.',
      exampleFormula: 'prop.test(c(45, 30), c(500, 500), correct = FALSE)',
    },
  },
  {
    id: 'm5_excel_tail_function_guide',
    familyId: 'concept_check',
    moduleId: 'module_5',
    title: 'Selecting Excel t-Distribution Functions',
    statisticalTask: 'Distinguish T.DIST, T.DIST.RT, T.DIST.2T, T.INV, and T.INV.2T.',
    relevantTemplateIds: ['hyp_excel_tail_function', 'hyp_alpha_changes_decision', 'hyp_formulate_right_tailed_sales', 'hyp_formulate_two_tailed_calibration', 'hyp_formulate_underfill', 'hyp_pvalue_interpretation', 'hyp_steps_and_conclusion', 'errors_alpha_beta_tradeoff', 'errors_define_in_context', 'errors_identify_after_the_fact', 'errors_larger_n_reduces_beta', 'errors_which_is_worse'],
    applicableKeywords: ['excel t functions', 't.dist.rt', 't.dist.2t', 'tail direction', 'critical value'],
    excel: {
      functionName: 'T.DIST Selector',
      genericSyntax: 'Right: =T.DIST.RT(t, df) | Left: =T.DIST(t, df, TRUE) | Two-tailed: =T.DIST.2T(ABS(t), df)',
      argumentDescriptions: [
        { name: 't', description: 'Computed test statistic.' },
        { name: 'df', description: 'Degrees of freedom.' },
      ],
      notes: 'Critical values: Left cutoff is =T.INV(alpha, df); Two-tailed positive cutoff is =T.INV.2T(alpha, df).',
      commonPitfalls: 'Using T.DIST.2T when t is negative (causes #NUM!) or using T.DIST.RT for a left-tailed test.',
      exampleFormula: '=T.DIST.2T(ABS(2.15), 24)',
    },
    r: {
      functionName: 'pt, qt Tail Selector',
      genericSyntax: 'Right: pt(t, df, lower.tail = FALSE) | Left: pt(t, df) | Two-tail: 2 * pt(-abs(t), df)',
      argumentDescriptions: [
        { name: 't', description: 't test statistic.' },
        { name: 'df', description: 'Degrees of freedom.' },
      ],
      notes: 'In R, pt(..., lower.tail = FALSE) is numerically more accurate than 1 - pt(...) for extreme tails.',
      commonPitfalls: 'Forgetting lower.tail = FALSE when calculating upper right-tail probability.',
      exampleFormula: '2 * pt(-abs(2.15), df = 24)',
    },
  },

  // ==========================================
  // MODULE 6: REGRESSION ANALYSIS & INFERENCE
  // ==========================================
  {
    id: 'm6_reg_bivariate_fit',
    familyId: 'linear_regression',
    moduleId: 'module_6',
    title: 'Simple Linear Regression: OLS Slope, Intercept & Prediction',
    statisticalTask: 'Fit y_hat = b0 + b1*x, calculate R^2, and forecast values within the observed range.',
    relevantTemplateIds: ['reg_fit_raw_points_price_units', 'regression_sales_advertising', 'reg_residual_observed_point', 'reg_interpret_slope_intercept', 'reg_extrapolation_warning'],
    applicableKeywords: ['simple linear regression', 'slope', 'intercept', 'ols', 'rsq', 'forecast.linear'],
    excel: {
      functionName: 'SLOPE, INTERCEPT, RSQ, FORECAST.LINEAR',
      genericSyntax: '=SLOPE(y_range, x_range) | =INTERCEPT(y_range, x_range) | =FORECAST.LINEAR(x_val, y_range, x_range)',
      argumentDescriptions: [
        { name: 'y_range', description: 'Known dependent response variable data range.' },
        { name: 'x_range', description: 'Known independent explanatory predictor data range.' },
        { name: 'x_val', description: 'Target x value for point forecast prediction.' },
      ],
      notes: 'In Excel regression functions, the known_y\'s range ALWAYS comes FIRST: =SLOPE(known_y\'s, known_x\'s).',
      commonPitfalls: 'Reversing argument order: passing x_range before y_range inverts the regression slope!',
      exampleFormula: '=FORECAST.LINEAR(17.0, B2:B7, A2:A7)',
    },
    r: {
      functionName: 'lm, predict',
      genericSyntax: 'model <- lm(y ~ x, data = df); summary(model); predict(model, newdata = data.frame(x = 17))',
      argumentDescriptions: [
        { name: 'y ~ x', description: 'Formula interface: response y on left of tilde, predictor x on right.' },
        { name: 'data', description: 'Data frame containing variables.' },
      ],
      notes: 'b0 is coef(model)[1] ("(Intercept)"); b1 is coef(model)[2].',
      commonPitfalls: 'Extrapolating predictions far outside the range of observed predictor data.',
      exampleFormula: 'model <- lm(units_sold ~ price, data = sales_data)',
    },
  },
  {
    id: 'm6_reg_residuals_se',
    familyId: 'linear_regression',
    moduleId: 'module_6',
    title: 'Residuals & Standard Error of the Estimate (s_e)',
    statisticalTask: 'Compute residual e = y - y_hat and standard error of estimate s_e = sqrt(SSE / (n - 2)).',
    relevantTemplateIds: ['regout_se_from_sse', 'reg_residual_observed_point'],
    applicableKeywords: ['residuals', 'standard error of estimate', 's_e', 'steyx', 'sse'],
    excel: {
      functionName: 'STEYX, Residual Formula',
      genericSyntax: 'Residual: =y_obs - y_hat | s_e: =STEYX(known_y\'s, known_x\'s) | =SQRT(SSE / (n - 2))',
      argumentDescriptions: [
        { name: 'known_y\'s, known_x\'s', description: 'Data ranges for y and x.' },
        { name: 'SSE', description: 'Sum of Squared Errors (Residual Sum of Squares from ANOVA table).' },
        { name: 'n - 2', description: 'Residual degrees of freedom for simple linear regression.' },
      ],
      notes: 's_e represents the typical distance that observed data points fall from the regression benchmark line.',
      commonPitfalls: 'Subtracting observed from predicted (y_hat - y) instead of observed minus predicted (y - y_hat).',
      exampleFormula: '=SQRT(288.0 / 18)',
    },
    r: {
      functionName: 'residuals, summary(model)$sigma',
      genericSyntax: 'resids <- residuals(model); s_e <- summary(model)$sigma',
      argumentDescriptions: [
        { name: 'model', description: 'Fitted lm() model object.' },
      ],
      notes: 'summary(model)$sigma directly extracts the standard error of the estimate (residual standard error).',
      commonPitfalls: 'Confusing standard error of the estimate s_e with the standard error of the slope s_b1.',
      exampleFormula: 'summary(model)$sigma',
    },
  },
  {
    id: 'm6_reg_anova_f',
    familyId: 'regression_output',
    moduleId: 'module_6',
    title: 'Regression ANOVA Table & Model Significance F',
    statisticalTask: 'Interpret ANOVA partitions SST = SSR + SSE, R^2 = SSR / SST, F = MSR / MSE, and Significance F.',
    relevantTemplateIds: ['regout_significance_f', 'regout_r2_from_anova', 'regout_r_from_r2_sign', 'regout_t_squared_equals_f'],
    applicableKeywords: ['regression anova', 'significance f', 'f-statistic', 'r-squared', 'multiple r'],
    excel: {
      functionName: 'F.DIST.RT (Significance F)',
      genericSyntax: '=F.DIST.RT(F_stat, df_regression, df_residual)',
      argumentDescriptions: [
        { name: 'F_stat', description: 'Overall F test statistic = MSR / MSE.' },
        { name: 'df_regression', description: 'Numerator degrees of freedom = k (number of predictors, 1 for simple regression).' },
        { name: 'df_residual', description: 'Denominator degrees of freedom = n - k - 1.' },
      ],
      notes: 'Significance F is the model p-value. For simple regression (k = 1), F = t^2 exactly, and p-values match.',
      commonPitfalls: 'Inverting numerator and denominator degrees of freedom in F.DIST.RT.',
      exampleFormula: '=F.DIST.RT(16.00, 1, 24)',
    },
    r: {
      functionName: 'anova, summary(model)',
      genericSyntax: 'anova(model); pf(f_stat, df1, df2, lower.tail = FALSE)',
      argumentDescriptions: [
        { name: 'model', description: 'Fitted lm() regression model.' },
      ],
      notes: 'anova(model) outputs the ANOVA table with Sum Sq, Mean Sq, F value, and Pr(>F).',
      commonPitfalls: 'R^2 = r^2 is strictly positive; correlation r inherits the negative sign of the slope if b1 < 0.',
      exampleFormula: 'pf(16.00, 1, 24, lower.tail = FALSE)',
    },
  },
  {
    id: 'm6_reg_slope_inference_t',
    familyId: 'regression_output',
    moduleId: 'module_6',
    title: 'Slope Hypothesis Testing and p-Value',
    statisticalTask: 'Test H0: beta_1 = 0 vs H1: beta_1 != 0 using t = b1 / s_b1 and df = n - 2.',
    relevantTemplateIds: ['regout_slope_t_pvalue', 'reg_predict_from_excel_coefficients'],
    applicableKeywords: ['slope t-statistic', 'slope p-value', 'regression coefficients table', 't.dist.2t'],
    excel: {
      functionName: 'T.DIST.2T for Slope',
      genericSyntax: 't: =b1 / s_b1 | p-value: =T.DIST.2T(ABS(t), df_residual)',
      argumentDescriptions: [
        { name: 'b1', description: 'Estimated slope coefficient.' },
        { name: 's_b1', description: 'Standard error of the slope coefficient.' },
        { name: 'df_residual', description: 'Residual degrees of freedom n - 2.' },
      ],
      notes: 'If p-value < alpha (e.g. 0.05), reject H0: a statistically significant linear relationship exists.',
      commonPitfalls: 'Testing against intercept degrees of freedom or using sample size n instead of n - 2.',
      exampleFormula: '=T.DIST.2T(ABS(2.40 / 0.80), 23)',
    },
    r: {
      functionName: 'summary(model)$coefficients',
      genericSyntax: 'summary(model)$coefficients["x", ]',
      argumentDescriptions: [
        { name: 'model', description: 'Fitted linear model.' },
      ],
      notes: 'Returns table with Estimate, Std. Error, t value, and Pr(>|t|).',
      commonPitfalls: 'Confusing statistical significance with practical economic importance or causation.',
      exampleFormula: 'summary(model)$coefficients',
    },
  },
  {
    id: 'm6_reg_slope_ci',
    familyId: 'regression_output',
    moduleId: 'module_6',
    title: 'Slope Confidence Interval',
    statisticalTask: 'Compute [b1 - t* * s_b1, b1 + t* * s_b1] with df = n - 2.',
    relevantTemplateIds: ['regout_compute_slope_ci', 'regout_slope_ci_includes_zero', 'regout_slope_ci_interpretation'],
    applicableKeywords: ['slope confidence interval', 't.inv.2t', 'duality with hypothesis test', 'margin of error slope'],
    excel: {
      functionName: 'Slope CI Formula',
      genericSyntax: 'Lower: =b1 - T.INV.2T(alpha, df) * s_b1 | Upper: =b1 + T.INV.2T(alpha, df) * s_b1',
      argumentDescriptions: [
        { name: 'b1', description: 'Slope estimate.' },
        { name: 'alpha', description: 'Significance level (e.g. 0.05 for 95% CI).' },
        { name: 'df', description: 'Residual degrees of freedom n - 2.' },
        { name: 's_b1', description: 'Standard error of the slope.' },
      ],
      notes: 'By testing duality: If the 95% CI contains 0, we fail to reject H0: beta_1 = 0 at alpha = 0.05.',
      commonPitfalls: 'Forgetting that confidence interval containing zero implies lack of statistical significance.',
      exampleFormula: '=120.0 - T.INV.2T(0.05, 25) * 15.0',
    },
    r: {
      functionName: 'confint',
      genericSyntax: 'confint(model, parm = "x", level = 1 - alpha)',
      argumentDescriptions: [
        { name: 'model', description: 'Fitted lm() model.' },
        { name: 'parm', description: 'Predictor variable name.' },
        { name: 'level', description: 'Confidence level (default 0.95).' },
      ],
      notes: 'confint(model) prints confidence intervals for both intercept and all slope predictors.',
      commonPitfalls: 'Supplying significance level alpha (0.05) instead of confidence level (0.95) to level argument.',
      exampleFormula: 'confint(model, level = 0.95)',
    },
  },
  {
    id: 'm6_mreg_equation_prediction',
    familyId: 'regression_output',
    moduleId: 'module_6',
    title: 'Multiple Linear Regression Equation & Point Forecast',
    statisticalTask: 'Evaluate y_hat = b0 + b1*x1 + b2*x2 + ... and interpret partial slopes under ceteris paribus.',
    relevantTemplateIds: ['mreg_write_equation_predict', 'mreg_interpret_coefficient_ceteris_paribus'],
    applicableKeywords: ['multiple regression', 'partial slope', 'ceteris paribus', 'forecast', 'linest'],
    excel: {
      functionName: 'LINEST, SUMPRODUCT Forecast',
      genericSyntax: 'Array Formula: {=LINEST(y_range, x_matrix, TRUE, TRUE)} | Forecast: =b0 + SUMPRODUCT(slopes, x_new)',
      argumentDescriptions: [
        { name: 'y_range', description: 'Column range of response variable.' },
        { name: 'x_matrix', description: 'Contiguous multi-column range of predictor variables.' },
        { name: 'slopes', description: 'Range of estimated partial slopes b1, b2, ...' },
        { name: 'x_new', description: 'Range of new predictor inputs.' },
      ],
      notes: 'Each partial slope b_j estimates the change in y per 1-unit increase in x_j, strictly holding all other predictors constant (ceteris paribus).',
      commonPitfalls: 'In LINEST, slope coefficients are output in REVERSE order (b_k, ..., b_1, b_0).',
      exampleFormula: '=50.0 + 4.2*10 - 2.5*20 + 0.8*25',
    },
    r: {
      functionName: 'lm (Multiple Regression), predict',
      genericSyntax: 'model <- lm(y ~ x1 + x2 + x3, data = df); predict(model, newdata = data.frame(x1 = 10, x2 = 20, x3 = 25))',
      argumentDescriptions: [
        { name: 'y ~ x1 + x2 + x3', description: 'Model formula specifying multiple additive predictors.' },
        { name: 'newdata', description: 'Data frame with column names matching model predictors.' },
      ],
      notes: 'summary(model) displays full multiple regression output with individual t-tests and overall F-test.',
      commonPitfalls: 'Omitting variable names or column structure in newdata produces recycling warnings.',
      exampleFormula: 'predict(model, newdata = data.frame(x1 = 10, x2 = 20, x3 = 25))',
    },
  },
  {
    id: 'm6_mreg_adj_r2_f',
    familyId: 'regression_output',
    moduleId: 'module_6',
    title: 'Model Comparison: Adjusted R^2 & Overall F vs Individual t-Tests',
    statisticalTask: 'Penalize complexity using Adjusted R^2 = 1 - (1 - R^2)*(n - 1)/(n - k - 1) and assess overall fit.',
    relevantTemplateIds: ['mreg_adjusted_r2_weak_predictor', 'mreg_overall_f_vs_individual_t', 'mreg_which_predictors_significant', 'regout_compare_models_fit'],
    applicableKeywords: ['adjusted r-squared', 'overall f vs individual t', 'weak predictor', 'model comparison', 'parsimony'],
    excel: {
      functionName: 'Adjusted R^2 Formula',
      genericSyntax: '=1 - ((1 - R2) * (n - 1) / (n - k - 1))',
      argumentDescriptions: [
        { name: 'R2', description: 'Standard coefficient of determination (RSQ).' },
        { name: 'n', description: 'Sample size.' },
        { name: 'k', description: 'Number of explanatory predictors.' },
      ],
      notes: 'Adding a weak or useless predictor always increases raw R^2, but DECREASES Adjusted R^2 due to degrees of freedom penalty.',
      commonPitfalls: 'Comparing raw R^2 between models with different numbers of predictors; always use Adjusted R^2.',
      exampleFormula: '=1 - ((1 - 0.75) * (29) / (26))',
    },
    r: {
      functionName: 'summary(model)$adj.r.squared, AIC',
      genericSyntax: 'summary(model)$adj.r.squared | AIC(model1, model2)',
      argumentDescriptions: [
        { name: 'model', description: 'Fitted multiple regression model.' },
      ],
      notes: 'AIC and BIC provide formal information-criterion penalties for predictor selection.',
      commonPitfalls: 'Assuming a significant overall F-test implies every individual predictor must be statistically significant.',
      exampleFormula: 'summary(model)$adj.r.squared',
    },
  },
  {
    id: 'm6_mreg_dummy_variables',
    familyId: 'regression_output',
    moduleId: 'module_6',
    title: 'Dummy / Indicator Variables in Multiple Regression',
    statisticalTask: 'Incorporate categorical predictors coded 0/1 and interpret slope relative to the baseline category.',
    relevantTemplateIds: ['mreg_dummy_variable'],
    applicableKeywords: ['dummy variable', 'indicator variable', 'categorical predictor', 'reference category', 'baseline'],
    excel: {
      functionName: 'Binary Dummy Coding (0/1)',
      genericSyntax: 'y_hat = b0 + b1*x + b2*D_urban (where D = 1 if Urban, 0 otherwise)',
      argumentDescriptions: [
        { name: 'D', description: 'Indicator variable taking value 1 for target category, 0 for reference baseline.' },
      ],
      notes: 'For a categorical factor with m levels, include exactly m - 1 dummy variables to prevent the Dummy Variable Trap (perfect multicollinearity).',
      commonPitfalls: 'Including m dummies along with an intercept, which causes exact multicollinearity (#NUM! in LINEST).',
      exampleFormula: '=25.0 + 3.5*10 + 8.0*1',
    },
    r: {
      functionName: 'as.factor in lm',
      genericSyntax: 'lm(y ~ x + factor(region), data = df)',
      argumentDescriptions: [
        { name: 'factor(region)', description: 'Categorical factor automatically expanded into m - 1 dummies.' },
      ],
      notes: 'R automatically creates m - 1 contrast dummies and assigns the first alphabetical level as reference.',
      commonPitfalls: 'Treating nominal coded numbers (e.g. 1, 2, 3) as continuous numbers instead of wrapping in factor().',
      exampleFormula: 'lm(price ~ sqft + factor(urban), data = real_estate)',
    },
  },
  {
    id: 'm6_reg_diagnostics_problems',
    familyId: 'concept_check',
    moduleId: 'module_6',
    title: 'Regression Diagnostics: Residual Analysis & Violation Remedies',
    statisticalTask: 'Diagnose curvature (nonlinearity), fan shapes (heteroscedasticity), autocorrelation, multicollinearity (VIF), and outliers.',
    relevantTemplateIds: ['regprob_nonlinearity', 'regprob_heteroscedasticity', 'regprob_autocorrelation', 'regprob_multicollinearity', 'regprob_outlier_influence', 'regprob_omitted_variable_causation'],
    applicableKeywords: ['residual plot', 'heteroscedasticity', 'multicollinearity', 'vif', 'autocorrelation', 'outlier influence'],
    excel: {
      functionName: 'Data Analysis Toolpak Residuals & VIF',
      genericSyntax: 'Residual Plot: Insert Scatter(y_hat, residuals) | VIF = 1 / (1 - Rj^2)',
      argumentDescriptions: [
        { name: 'residuals', description: 'Regression residuals column from Toolpak output.' },
        { name: 'Rj^2', description: 'R^2 from regressing predictor x_j on all other predictors.' },
      ],
      notes: 'Heteroscedasticity shows fan/cone dispersion (remedy: log transform or robust SEs); Curvature shows U/inverted-U pattern (remedy: quadratic x^2); VIF > 5 or 10 signals multicollinearity.',
      commonPitfalls: 'Plotting residuals against raw row index rather than fitted values y_hat or predictor x.',
      exampleFormula: '=1 / (1 - 0.85)',
    },
    r: {
      functionName: 'plot(model), car::vif, lmtest',
      genericSyntax: 'par(mfrow = c(2, 2)); plot(model) | car::vif(model) | lmtest::dwtest(model)',
      argumentDescriptions: [
        { name: 'model', description: 'Fitted lm() model.' },
      ],
      notes: 'plot(model) generates: 1) Residuals vs Fitted, 2) Normal Q-Q, 3) Scale-Location, 4) Residuals vs Leverage (Cook\'s distance).',
      commonPitfalls: 'Dropping significant predictors without assessing multicollinearity or omitted variable bias.',
      exampleFormula: 'car::vif(multiple_model)',
    },
  },
];

/**
 * Deterministic mapping table from every approved template ID to its primary Software Recipe.
 * Guarantees 100% of the 137 approved templates resolve immediately with pedagogical precision.
 */
export const TEMPLATE_TO_RECIPE_MAP: Record<string, string> = {
  // Module 1 (20 templates)
  'concept_cross_sectional_vs_time_series': 'm1_contingency_independence',
  'concept_data_types_business': 'm1_contingency_independence',
  'concept_descriptive_vs_inferential': 'm1_contingency_independence',
  'concept_parameter_vs_statistic': 'm1_contingency_independence',
  'concept_sampling_bias_online_poll': 'm1_contingency_independence',
  'prob_contingency_age_purchase': 'm1_contingency_independence',
  'prob_disjoint_vs_independent': 'm1_contingency_independence',
  'desc_cv_compare_stocks': 'm1_coef_variation',
  'desc_mean_median_outlier': 'm1_center_measures',
  'desc_quartiles_iqr_fences': 'm1_quartiles_iqr',
  'desc_sample_sd_delivery_times': 'm1_dispersion_sample',
  'desc_weighted_mean_price': 'm1_weighted_mean',
  'desc_zscore_empirical_rule': 'm1_zscore_standardize',
  'prob_at_least_one_server_failure': 'm1_prob_complement',
  'prob_independence_check': 'm1_prob_conditional_mult',
  'prob_multiplication_independent_suppliers': 'm1_prob_conditional_mult',
  'prob_two_plant_defect_tree': 'm1_prob_bayes_total',
  'set_probability_complement_rule': 'm1_prob_complement',
  'set_probability_conditional': 'm1_prob_conditional_mult',
  'set_probability_union_rule': 'm1_prob_union_addition',

  // Module 2 (24 templates)
  'binomial_at_least_three_defects': 'm2_binom_cumulative_cdf',
  'binomial_defective_at_most_one': 'm2_binom_cumulative_cdf',
  'binomial_fair_coin_exactly_two': 'm2_binom_exact_pmf',
  'binomial_mean_sd_shipments': 'm2_binom_parameters',
  'binomial_service_calls_zero': 'm2_binom_exact_pmf',
  'binomial_condition_violated_small_lot': 'm2_binom_parameters',
  'choose_binomial_or_poisson': 'm2_binom_parameters',
  'drv_compare_projects_risk': 'm2_discrete_rv_moments',
  'drv_cumulative_from_table': 'm2_discrete_rv_moments',
  'drv_expected_cars_sold': 'm2_discrete_rv_moments',
  'drv_insurance_break_even': 'm2_discrete_rv_moments',
  'drv_valid_pmf_missing_value': 'm2_discrete_rv_moments',
  'drv_variance_cars_sold': 'm2_discrete_rv_moments',
  'affine_currency_conversion': 'm2_linear_comb_affine',
  'affine_profit_from_units': 'm2_linear_comb_affine',
  'lincomb_difference_independent': 'm2_linear_comb_sums',
  'lincomb_portfolio_correlated': 'm2_portfolio_covariance',
  'lincomb_total_two_product_lines': 'm2_linear_comb_sums',
  'lincomb_weekly_demand_sum_vs_scale': 'm2_linear_comb_affine',
  'linear_comb_sum_variance_independent': 'm2_linear_comb_sums',
  'poisson_call_center_arrivals': 'm2_poisson_exact_pmf',
  'poisson_rare_event_one': 'm2_poisson_exact_pmf',
  'poisson_rate_rescaling_quarter_hour': 'm2_poisson_cumulative_cdf',
  'poisson_server_failures_zero': 'm2_poisson_exact_pmf',

  // Module 3 (20 templates)
  'contvdisc_classify_variables': 'm3_concept_continuous_vs_discrete',
  'contvdisc_match_model': 'm3_concept_continuous_vs_discrete',
  'contvdisc_point_probability_zero': 'm3_concept_continuous_vs_discrete',
  'contvdisc_strict_vs_nonstrict': 'm3_concept_continuous_vs_discrete',
  'exponential_arrival_within_two': 'm3_exponential_cdf',
  'exponential_between_two_times': 'm3_exponential_cdf',
  'exponential_memoryless_wait': 'm3_exponential_cdf',
  'exponential_percentile_service': 'm3_exponential_tail_quantile',
  'exponential_poisson_link': 'm3_exponential_cdf',
  'exponential_rate_vs_mean_param': 'm3_exponential_cdf',
  'exponential_service_lifetime': 'm3_exponential_cdf',
  'normal_between_package_weight': 'm3_normal_interval',
  'normal_compare_standardized_scores': 'm3_normal_standard_z',
  'normal_empirical_vs_exact': 'm3_normal_general_x',
  'normal_inventory_service_level': 'm3_normal_general_x',
  'normal_inverse_bonus_cutoff': 'm3_normal_general_x',
  'normal_standard_upper_tail': 'm3_normal_standard_z',
  'normal_warranty_lower_percentile': 'm3_normal_general_x',
  'contvdisc_area_is_probability': 'm3_uniform_interval',
  'uniform_commute_wait_time': 'm3_uniform_interval',

  // Module 4 (23 templates)
  'ci_interpretation_statement': 'm4_ci_mean_z',
  'ci_level_vs_width': 'm4_ci_mean_z',
  'ci_plausible_claim': 'm4_ci_mean_z',
  'clt_small_n_not_justified': 'm4_clt_sampling_dist_mean',
  'n_halve_margin_quadruple': 'm4_sample_size_mean',
  'n_round_up_misconception': 'm4_sample_size_mean',
  'se_what_n_changes': 'm4_sample_size_mean',
  'ci_mean_sample_estimate': 'm4_ci_mean_t',
  'ci_mean_t_small_sample_90': 'm4_ci_mean_t',
  'ci_mean_z_known_sigma': 'm4_ci_mean_z',
  'ci_proportion_recommend': 'm4_ci_proportion',
  'n_mean_margin_five': 'm4_sample_size_mean',
  'n_proportion_conservative': 'm4_sample_size_prop',
  'n_proportion_pilot': 'm4_sample_size_prop',
  'se_n_for_target_se': 'm4_sample_size_mean',
  'se_quadruple_n_halves': 'm4_sample_size_mean',
  'se_survey_design_compare': 'm4_sample_size_mean',
  'clt_finite_population_correction': 'm4_clt_sampling_dist_mean',
  'clt_individual_vs_mean': 'm4_clt_sampling_dist_mean',
  'clt_mean_order_value': 'm4_clt_sampling_dist_mean',
  'clt_proportion_coupon_users': 'm4_clt_sampling_dist_prop',
  'clt_skewed_service_times': 'm4_clt_sampling_dist_mean',
  'se_proportion_depends_on_p': 'm4_clt_sampling_dist_prop',

  // Module 5 (22 templates)
  'errors_alpha_beta_tradeoff': 'm5_excel_tail_function_guide',
  'errors_define_in_context': 'm5_excel_tail_function_guide',
  'errors_identify_after_the_fact': 'm5_excel_tail_function_guide',
  'errors_larger_n_reduces_beta': 'm5_excel_tail_function_guide',
  'errors_which_is_worse': 'm5_excel_tail_function_guide',
  'hyp_alpha_changes_decision': 'm5_excel_tail_function_guide',
  'hyp_excel_tail_function': 'm5_excel_tail_function_guide',
  'hyp_formulate_proportion': 'm5_hyp_test_prop_z',
  'hyp_formulate_right_tailed_sales': 'm5_excel_tail_function_guide',
  'hyp_formulate_two_tailed_calibration': 'm5_excel_tail_function_guide',
  'hyp_formulate_underfill': 'm5_excel_tail_function_guide',
  'hyp_pvalue_interpretation': 'm5_excel_tail_function_guide',
  'hyp_steps_and_conclusion': 'm5_excel_tail_function_guide',
  'twosample_choose_design': 'm5_twosample_independent_t',
  'hyp_proportion_left_tailed': 'm5_hyp_test_prop_z',
  'hyp_t_mean_right_tailed': 'm5_hyp_test_mean_t',
  'hyp_z_mean_two_tailed': 'm5_hyp_test_mean_z',
  'hypothesis_mean_one_sample': 'm5_hyp_test_mean_t',
  'twosample_independent_means_layouts': 'm5_twosample_independent_t',
  'twosample_paired_training': 'm5_twosample_paired_t',
  'twosample_read_excel_output': 'm5_twosample_independent_t',
  'twosample_two_proportion_ads': 'm5_twosample_prop_z',

  // Module 6 (28 templates)
  'reg_extrapolation_warning': 'm6_reg_bivariate_fit',
  'reg_interpret_slope_intercept': 'm6_reg_bivariate_fit',
  'regprob_autocorrelation': 'm6_reg_diagnostics_problems',
  'regprob_heteroscedasticity': 'm6_reg_diagnostics_problems',
  'regprob_multicollinearity': 'm6_reg_diagnostics_problems',
  'regprob_nonlinearity': 'm6_reg_diagnostics_problems',
  'regprob_omitted_variable_causation': 'm6_reg_diagnostics_problems',
  'regprob_outlier_influence': 'm6_reg_diagnostics_problems',
  'reg_fit_raw_points_price_units': 'm6_reg_bivariate_fit',
  'reg_residual_observed_point': 'm6_reg_residuals_se',
  'regression_sales_advertising': 'm6_reg_bivariate_fit',
  'mreg_adjusted_r2_weak_predictor': 'm6_mreg_adj_r2_f',
  'mreg_dummy_variable': 'm6_mreg_dummy_variables',
  'mreg_interpret_coefficient_ceteris_paribus': 'm6_mreg_equation_prediction',
  'mreg_overall_f_vs_individual_t': 'm6_mreg_adj_r2_f',
  'mreg_which_predictors_significant': 'm6_mreg_adj_r2_f',
  'mreg_write_equation_predict': 'm6_mreg_equation_prediction',
  'reg_predict_from_excel_coefficients': 'm6_reg_slope_inference_t',
  'regout_compare_models_fit': 'm6_mreg_adj_r2_f',
  'regout_compute_slope_ci': 'm6_reg_slope_ci',
  'regout_r_from_r2_sign': 'm6_reg_anova_f',
  'regout_r2_from_anova': 'm6_reg_anova_f',
  'regout_se_from_sse': 'm6_reg_residuals_se',
  'regout_significance_f': 'm6_reg_anova_f',
  'regout_slope_ci_includes_zero': 'm6_reg_slope_ci',
  'regout_slope_ci_interpretation': 'm6_reg_slope_ci',
  'regout_slope_t_pvalue': 'm6_reg_slope_inference_t',
  'regout_t_squared_equals_f': 'm6_reg_anova_f',
};

/**
 * Fallback mapping by familyId and moduleId when template ID is not explicitly listed.
 */
const FAMILY_MODULE_FALLBACK_MAP: Record<string, string> = {
  'descriptive_stats:module_1': 'm1_center_measures',
  'set_probability:module_1': 'm1_prob_complement',
  'concept_check:module_1': 'm1_contingency_independence',

  'binomial_pmf:module_2': 'm2_binom_exact_pmf',
  'poisson_pmf:module_2': 'm2_poisson_exact_pmf',
  'discrete_rv:module_2': 'm2_discrete_rv_moments',
  'linear_combination:module_2': 'm2_linear_comb_affine',
  'concept_check:module_2': 'm2_binom_parameters',

  'normal_dist:module_3': 'm3_normal_general_x',
  'uniform_dist:module_3': 'm3_uniform_interval',
  'exponential_dist:module_3': 'm3_exponential_cdf',
  'concept_check:module_3': 'm3_concept_continuous_vs_discrete',

  'sampling_distribution:module_4': 'm4_clt_sampling_dist_mean',
  'confidence_interval:module_4': 'm4_ci_mean_t',
  'sample_size:module_4': 'm4_sample_size_mean',
  'concept_check:module_4': 'm4_ci_mean_z',

  'hypothesis_test:module_5': 'm5_hyp_test_mean_t',
  'two_sample_test:module_5': 'm5_twosample_independent_t',
  'concept_check:module_5': 'm5_excel_tail_function_guide',

  'linear_regression:module_6': 'm6_reg_bivariate_fit',
  'regression_output:module_6': 'm6_reg_anova_f',
  'concept_check:module_6': 'm6_reg_diagnostics_problems',
};

/**
 * Fallback mapping by familyId only.
 */
const FAMILY_FALLBACK_MAP: Record<string, string> = {
  'descriptive_stats': 'm1_center_measures',
  'set_probability': 'm1_prob_complement',
  'binomial_pmf': 'm2_binom_exact_pmf',
  'poisson_pmf': 'm2_poisson_exact_pmf',
  'discrete_rv': 'm2_discrete_rv_moments',
  'linear_combination': 'm2_linear_comb_affine',
  'normal_dist': 'm3_normal_general_x',
  'uniform_dist': 'm3_uniform_interval',
  'exponential_dist': 'm3_exponential_cdf',
  'sampling_distribution': 'm4_clt_sampling_dist_mean',
  'confidence_interval': 'm4_ci_mean_t',
  'sample_size': 'm4_sample_size_mean',
  'hypothesis_test': 'm5_hyp_test_mean_t',
  'two_sample_test': 'm5_twosample_independent_t',
  'linear_regression': 'm6_reg_bivariate_fit',
  'regression_output': 'm6_reg_anova_f',
};

/**
 * Retrieve all registered software recipes.
 */
export function getAllRecipes(): SoftwareRecipe[] {
  return SOFTWARE_RECIPES;
}

/**
 * Retrieve a recipe by its unique ID.
 */
export function getRecipeById(id: string): SoftwareRecipe | undefined {
  return SOFTWARE_RECIPES.find((r) => r.id === id);
}

/**
 * Retrieve all recipes associated with a particular curriculum module.
 */
export function getRecipesByModule(moduleId: string): SoftwareRecipe[] {
  return SOFTWARE_RECIPES.filter((r) => r.moduleId === moduleId);
}

/**
 * Retrieve all recipes associated with a problem family.
 */
export function getRecipesByFamily(familyId: string): SoftwareRecipe[] {
  return SOFTWARE_RECIPES.filter((r) => r.familyId === familyId);
}

/**
 * Query criteria for resolving a software recipe from a problem or session context.
 */
export interface RecipeLookupQuery {
  id?: string;
  template_id?: string;
  family_id?: string;
  module_id?: string;
  title?: string;
}

/**
 * Resolve the optimal Software Recipe for an active drill, question, or template.
 *
 * Evaluation order:
 * 1. Direct template ID match in TEMPLATE_TO_RECIPE_MAP
 * 2. Recipe ID directly matching query.id
 * 3. Exact (family_id + module_id) match in fallback table
 * 4. Exact family_id match in fallback table
 * 5. Title keyword search
 * 6. Module default fallback (first recipe of module)
 */
export function getRecipeForTemplate(query: RecipeLookupQuery): SoftwareRecipe | undefined {
  const targetId = query.template_id || query.id;

  // 1. Direct template ID lookup
  if (targetId && TEMPLATE_TO_RECIPE_MAP[targetId]) {
    const matched = getRecipeById(TEMPLATE_TO_RECIPE_MAP[targetId]);
    if (matched) return matched;
  }

  // 2. Query ID itself is a recipe ID
  if (targetId) {
    const directRecipe = getRecipeById(targetId);
    if (directRecipe) return directRecipe;
  }

  // 3. Family + Module combination
  if (query.family_id && query.module_id) {
    const key = `${query.family_id}:${query.module_id}`;
    if (FAMILY_MODULE_FALLBACK_MAP[key]) {
      const matched = getRecipeById(FAMILY_MODULE_FALLBACK_MAP[key]);
      if (matched) return matched;
    }
  }

  // 4. Family alone
  if (query.family_id && FAMILY_FALLBACK_MAP[query.family_id]) {
    const matched = getRecipeById(FAMILY_FALLBACK_MAP[query.family_id]);
    if (matched) return matched;
  }

  // 5. Keyword search in title
  if (query.title) {
    const lowerTitle = query.title.toLowerCase();
    for (const recipe of SOFTWARE_RECIPES) {
      if (recipe.applicableKeywords?.some((kw) => lowerTitle.includes(kw.toLowerCase()))) {
        return recipe;
      }
    }
  }

  // 6. Module fallback
  if (query.module_id) {
    const moduleRecipes = getRecipesByModule(query.module_id);
    if (moduleRecipes.length > 0) {
      return moduleRecipes[0];
    }
  }

  return undefined;
}
