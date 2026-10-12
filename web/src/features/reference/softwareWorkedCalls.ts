/**
 * Post-Submission Parameter Interpolation Engine (Phase 3)
 *
 * Takes a SoftwareRecipe, session/stage parameters, and expected answers to generate
 * exact, plugged-in Microsoft Excel formulas and R / RStudio calls with substituted numbers.
 *
 * Connects the learner's numerical calculation directly to real-world software calls.
 */

import { SoftwareRecipe, getRecipeForTemplate, RecipeLookupQuery } from './softwareRecipes';

export interface WorkedCallSpec {
  software: 'excel' | 'r';
  call: string;
  result?: string;
  description: string;
}

export interface WorkedSoftwareOutput {
  recipeId: string;
  recipeTitle: string;
  task: string;
  excel: WorkedCallSpec;
  r: WorkedCallSpec;
  explanation?: string;
}

export interface InterpolationInput {
  recipe?: SoftwareRecipe;
  lookup?: RecipeLookupQuery;
  parameters?: Record<string, unknown>;
  expectedAnswer?: number | string | { value?: number; units?: string; option_id?: string; [key: string]: unknown };
  stageLabel?: string;
  stageNumber?: number;
}

/**
 * Format a number cleanly: integer without decimal, floats rounded up to 5 decimals with trailing zeros trimmed.
 */
export function formatParamNumber(val: unknown, fallback = '0'): string {
  if (typeof val === 'number') {
    if (Number.isInteger(val)) return String(val);
    const rounded = parseFloat(val.toFixed(5));
    return String(rounded);
  }
  if (typeof val === 'string') {
    const num = parseFloat(val);
    if (!isNaN(num)) {
      if (Number.isInteger(num)) return String(num);
      return String(parseFloat(num.toFixed(5)));
    }
    return val;
  }
  return fallback;
}

/**
 * Format expected answer to string if available.
 */
function extractExpectedResult(ans: unknown): string | undefined {
  if (ans === undefined || ans === null) return undefined;
  if (typeof ans === 'number') {
    if (Number.isInteger(ans)) return String(ans);
    return String(parseFloat(ans.toFixed(5)));
  }
  if (typeof ans === 'string') return ans;
  if (typeof ans === 'object' && ans !== null && 'value' in ans) {
    const v = (ans as { value?: unknown }).value;
    if (typeof v === 'number') {
      if (Number.isInteger(v)) return String(v);
      return String(parseFloat(v.toFixed(5)));
    }
  }
  return undefined;
}

/**
 * Generate exact worked Excel and R calls by interpolating problem parameters.
 */
export function generateWorkedSoftwareCalls(input: InterpolationInput): WorkedSoftwareOutput | null {
  const recipe = input.recipe || (input.lookup ? getRecipeForTemplate(input.lookup) : undefined);
  if (!recipe) return null;

  const p = input.parameters || {};
  const resultStr = extractExpectedResult(input.expectedAnswer);

  // Helper getters with common aliases
  const getNum = (keys: string[], def: number): number => {
    for (const k of keys) {
      if (p[k] !== undefined && p[k] !== null) {
        const parsed = Number(p[k]);
        if (!isNaN(parsed)) return parsed;
      }
    }
    return def;
  };

  const getStr = (keys: string[], def = ''): string => {
    for (const k of keys) {
      if (typeof p[k] === 'string') return p[k] as string;
    }
    return def;
  };

  switch (recipe.id) {
    // ----------------------------------------------------
    // MODULE 1: SET PROBABILITY & DESCRIPTIVE
    // ----------------------------------------------------
    case 'm1_center_measures': {
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: '=AVERAGE(A2:A50)  [or =MEDIAN(A2:A50)]',
          result: resultStr,
          description: 'Calculates sample mean (sensitive) or median (resistant to skew).',
        },
        r: {
          software: 'r',
          call: 'mean(x, na.rm = TRUE)  # or median(x, na.rm = TRUE)',
          result: resultStr,
          description: 'Base R arithmetic mean or 50th percentile median vector summary.',
        },
        explanation: 'In Excel use =AVERAGE() or =MEDIAN(); in R use mean() or median().',
      };
    }

    case 'm1_dispersion_sample': {
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: '=STDEV.S(A2:A50)  |  =VAR.S(A2:A50)',
          result: resultStr,
          description: 'Sample standard deviation and variance with Bessel correction (n - 1).',
        },
        r: {
          software: 'r',
          call: 'sd(x, na.rm = TRUE)  # or var(x, na.rm = TRUE)',
          result: resultStr,
          description: 'Unbiased sample standard deviation and variance in R.',
        },
        explanation: 'Always use sample functions with (n - 1) divisor for sample data.',
      };
    }

    case 'm1_prob_complement':
    case 'm1_complement_rule': {
      const probA = formatParamNumber(p.p_a ?? p.p ?? 0.25);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=1 - ${probA}`,
          result: resultStr,
          description: `Complement probability P(A') = 1 - P(A) with P(A) = ${probA}`,
        },
        r: {
          software: 'r',
          call: `1 - ${probA}`,
          result: resultStr,
          description: `Complement subtraction in R with p = ${probA}`,
        },
        explanation: `By the complement axiom, P(A^c) = 1 - P(A) = 1 - ${probA}.`,
      };
    }

    case 'm1_prob_union_addition':
    case 'm1_addition_union': {
      const pa = formatParamNumber(p.p_a ?? p.p1 ?? 0.4);
      const pb = formatParamNumber(p.p_b ?? p.p2 ?? 0.35);
      const pab = formatParamNumber(p.p_intersection ?? p.p_ab ?? 0.15);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=${pa} + ${pb} - ${pab}`,
          result: resultStr,
          description: `Union probability = P(A) + P(B) - P(A ∩ B)`,
        },
        r: {
          software: 'r',
          call: `${pa} + ${pb} - ${pab}`,
          result: resultStr,
          description: `General addition rule accounting for event overlap`,
        },
        explanation: `P(A ∪ B) = P(A) + P(B) - P(A ∩ B) = ${pa} + ${pb} - ${pab}.`,
      };
    }

    case 'm1_prob_conditional_mult':
    case 'm1_conditional_prob': {
      const pab = formatParamNumber(p.p_intersection ?? p.p_ab ?? 0.12);
      const pb = formatParamNumber(p.p_b ?? p.p2 ?? 0.30);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=${pab} / ${pb}`,
          result: resultStr,
          description: `Conditional probability P(A|B) = P(A ∩ B) / P(B)`,
        },
        r: {
          software: 'r',
          call: `${pab} / ${pb}`,
          result: resultStr,
          description: `Ratio of joint probability to prior conditioning probability`,
        },
        explanation: `P(A | B) = P(A ∩ B) / P(B) = ${pab} / ${pb}.`,
      };
    }

    // ----------------------------------------------------
    // MODULE 2: BINOMIAL & POISSON
    // ----------------------------------------------------
    case 'm2_binom_exact_pmf':
    case 'm2_binomial_pmf': {
      const n = getNum(['n', 'trials', 'trials_per_experiment'], 4);
      const pVal = getNum(['p', 'prob', 'prob_s', 'probability'], 0.5);
      const k = getNum(['k', 'x', 'target_successes', 'successes'], 2);
      const pStr = formatParamNumber(pVal);

      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=BINOM.DIST(${k}, ${n}, ${pStr}, FALSE)`,
          result: resultStr,
          description: `Exact probability P(X = ${k}) with n = ${n}, p = ${pStr}. (FALSE = exact PMF)`,
        },
        r: {
          software: 'r',
          call: `dbinom(${k}, size = ${n}, prob = ${pStr})`,
          result: resultStr,
          description: `Density / mass function dbinom() evaluated at k = ${k}`,
        },
        explanation: `Exact point probability P(X = ${k}): Excel uses =BINOM.DIST(k, n, p, FALSE); R uses dbinom(k, n, p).`,
      };
    }

    case 'm2_binom_cumulative_cdf':
    case 'm2_binomial_cdf': {
      const n = getNum(['n', 'trials'], 10);
      const pVal = getNum(['p', 'prob'], 0.2);
      const k = getNum(['k', 'x', 'threshold'], 2);
      const pStr = formatParamNumber(pVal);
      const isUpper = getStr(['tail', 'direction', 'inequality']).includes('greater') ||
                      getStr(['tail', 'direction', 'inequality']).includes('upper') ||
                      getStr(['tail', 'direction', 'inequality']).includes('at_least');

      if (isUpper) {
        const kMinus1 = Math.max(0, k - 1);
        return {
          recipeId: recipe.id,
          recipeTitle: recipe.title,
          task: recipe.statisticalTask,
          excel: {
            software: 'excel',
            call: `=1 - BINOM.DIST(${kMinus1}, ${n}, ${pStr}, TRUE)`,
            result: resultStr,
            description: `Upper tail P(X >= ${k}) = 1 - P(X <= ${kMinus1}) using cumulative TRUE`,
          },
          r: {
            software: 'r',
            call: `pbinom(${kMinus1}, size = ${n}, prob = ${pStr}, lower.tail = FALSE)`,
            result: resultStr,
            description: `Upper tail probability via lower.tail = FALSE with k - 1 = ${kMinus1}`,
          },
          explanation: `Cumulative upper tail P(X >= ${k}) = 1 - P(X <= ${kMinus1}).`,
        };
      }

      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=BINOM.DIST(${k}, ${n}, ${pStr}, TRUE)`,
          result: resultStr,
          description: `Cumulative probability P(X <= ${k}) with n = ${n}, p = ${pStr}. (TRUE = CDF)`,
        },
        r: {
          software: 'r',
          call: `pbinom(${k}, size = ${n}, prob = ${pStr})`,
          result: resultStr,
          description: `Cumulative distribution function pbinom() at k = ${k}`,
        },
        explanation: `Cumulative probability P(X <= ${k}): Excel uses =BINOM.DIST(k, n, p, TRUE); R uses pbinom(k, n, p).`,
      };
    }

    case 'm2_binom_parameters':
    case 'm2_binomial_moments': {
      const n = getNum(['n', 'trials'], 10);
      const pVal = getNum(['p', 'prob'], 0.3);
      const pStr = formatParamNumber(pVal);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `Mean: =${n} * ${pStr}  |  SD: =SQRT(${n} * ${pStr} * (1 - ${pStr}))`,
          result: resultStr,
          description: `Expected value E[X] = np and standard deviation σ = sqrt(np(1-p))`,
        },
        r: {
          software: 'r',
          call: `mu <- ${n} * ${pStr}; sigma <- sqrt(${n} * ${pStr} * (1 - ${pStr}))`,
          result: resultStr,
          description: `Analytical moment calculation in R`,
        },
        explanation: `Mean μ = np = ${n} × ${pStr}; σ = √(np(1-p)).`,
      };
    }

    case 'm2_poisson_exact_pmf':
    case 'm2_poisson_pmf': {
      const lambda = getNum(['lambda', 'rate', 'mean_rate'], 3.0);
      const k = getNum(['k', 'x', 'arrivals'], 2);
      const lamStr = formatParamNumber(lambda);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=POISSON.DIST(${k}, ${lamStr}, FALSE)`,
          result: resultStr,
          description: `Exact Poisson probability P(X = ${k}) with rate λ = ${lamStr}. (FALSE = exact PMF)`,
        },
        r: {
          software: 'r',
          call: `dpois(${k}, lambda = ${lamStr})`,
          result: resultStr,
          description: `Poisson mass function dpois() at k = ${k}`,
        },
        explanation: `Exact Poisson probability P(X = ${k}): Excel uses =POISSON.DIST(k, λ, FALSE); R uses dpois(k, λ).`,
      };
    }

    case 'm2_poisson_cumulative_cdf':
    case 'm2_poisson_cdf': {
      const lambda = getNum(['lambda', 'rate', 'mean_rate'], 4.0);
      const k = getNum(['k', 'x', 'threshold'], 3);
      const lamStr = formatParamNumber(lambda);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=POISSON.DIST(${k}, ${lamStr}, TRUE)`,
          result: resultStr,
          description: `Cumulative Poisson probability P(X <= ${k}) with λ = ${lamStr}. (TRUE = CDF)`,
        },
        r: {
          software: 'r',
          call: `ppois(${k}, lambda = ${lamStr})`,
          result: resultStr,
          description: `Cumulative Poisson probability ppois() at k = ${k}`,
        },
        explanation: `Cumulative Poisson probability P(X <= ${k}): Excel uses =POISSON.DIST(k, λ, TRUE); R uses ppois(k, λ).`,
      };
    }

    // ----------------------------------------------------
    // MODULE 3: CONTINUOUS DISTRIBUTIONS (NORMAL, UNIFORM, EXPONENTIAL)
    // ----------------------------------------------------
    case 'm3_normal_standard_z': {
      const z = getNum(['z', 'x', 'score'], 1.96);
      const zStr = formatParamNumber(z);
      const isUpper = getStr(['tail', 'direction']).includes('upper') ||
                      getStr(['tail', 'direction']).includes('greater');
      if (isUpper) {
        return {
          recipeId: recipe.id,
          recipeTitle: recipe.title,
          task: recipe.statisticalTask,
          excel: {
            software: 'excel',
            call: `=1 - NORM.S.DIST(${zStr}, TRUE)`,
            result: resultStr,
            description: `Standard normal upper tail P(Z >= ${zStr})`,
          },
          r: {
            software: 'r',
            call: `pnorm(${zStr}, lower.tail = FALSE)`,
            result: resultStr,
            description: `Standard normal upper tail via lower.tail = FALSE`,
          },
          explanation: `Standard normal upper tail probability P(Z >= ${zStr}).`,
        };
      }
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=NORM.S.DIST(${zStr}, TRUE)`,
          result: resultStr,
          description: `Standard normal cumulative probability Φ(${zStr})`,
        },
        r: {
          software: 'r',
          call: `pnorm(${zStr})`,
          result: resultStr,
          description: `Standard normal cumulative CDF at z = ${zStr}`,
        },
        explanation: `Standard normal CDF: Excel uses =NORM.S.DIST(z, TRUE); R uses pnorm(z).`,
      };
    }

    case 'm3_normal_general_x':
    case 'm3_normal_cdf':
    case 'm3_normal_inv': {
      const mu = getNum(['mu', 'mean'], 0);
      const sigma = getNum(['sigma', 'sd'], 1);
      const muStr = formatParamNumber(mu);
      const sigmaStr = formatParamNumber(sigma);
      const isInverse = p.p !== undefined && p.x === undefined;
      const isUpper = getStr(['tail', 'direction']).includes('upper') ||
                      getStr(['tail', 'direction']).includes('greater');

      if (isInverse) {
        const prob = getNum(['p', 'prob', 'probability', 'alpha'], 0.05);
        const pStr = formatParamNumber(prob);
        const callProb = isUpper ? `1 - ${pStr}` : pStr;
        return {
          recipeId: recipe.id,
          recipeTitle: recipe.title,
          task: recipe.statisticalTask,
          excel: {
            software: 'excel',
            call: mu === 0 && sigma === 1 ? `=NORM.S.INV(${callProb})` : `=NORM.INV(${callProb}, ${muStr}, ${sigmaStr})`,
            result: resultStr,
            description: `Inverse normal cutoff X with μ = ${muStr}, σ = ${sigmaStr}`,
          },
          r: {
            software: 'r',
            call: isUpper
              ? `qnorm(${pStr}, mean = ${muStr}, sd = ${sigmaStr}, lower.tail = FALSE)`
              : `qnorm(${pStr}, mean = ${muStr}, sd = ${sigmaStr})`,
            result: resultStr,
            description: `Quantile function qnorm() with mean = ${muStr}, sd = ${sigmaStr}`,
          },
          explanation: `Normal quantile cutoff: Excel uses =NORM.INV(prob, μ, σ); R uses qnorm(prob, mean = μ, sd = σ).`,
        };
      }

      const x = getNum(['x', 'x_cutoff', 'value'], 50000);
      const xStr = formatParamNumber(x);

      if (isUpper) {
        return {
          recipeId: recipe.id,
          recipeTitle: recipe.title,
          task: recipe.statisticalTask,
          excel: {
            software: 'excel',
            call: mu === 0 && sigma === 1 ? `=1 - NORM.S.DIST(${xStr}, TRUE)` : `=1 - NORM.DIST(${xStr}, ${muStr}, ${sigmaStr}, TRUE)`,
            result: resultStr,
            description: `General normal upper tail P(X >= ${xStr}) with μ = ${muStr}, σ = ${sigmaStr}`,
          },
          r: {
            software: 'r',
            call: `pnorm(${xStr}, mean = ${muStr}, sd = ${sigmaStr}, lower.tail = FALSE)`,
            result: resultStr,
            description: `General normal upper tail via lower.tail = FALSE`,
          },
          explanation: `Upper tail P(X >= ${xStr}) = 1 - P(X <= ${xStr}).`,
        };
      }

      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: mu === 0 && sigma === 1 ? `=NORM.S.DIST(${xStr}, TRUE)` : `=NORM.DIST(${xStr}, ${muStr}, ${sigmaStr}, TRUE)`,
          result: resultStr,
          description: `General normal cumulative probability P(X <= ${xStr}) with μ = ${muStr}, σ = ${sigmaStr}`,
        },
        r: {
          software: 'r',
          call: `pnorm(${xStr}, mean = ${muStr}, sd = ${sigmaStr})`,
          result: resultStr,
          description: `General normal CDF pnorm() evaluated at x = ${xStr}`,
        },
        explanation: `Normal CDF: Excel uses =NORM.DIST(x, μ, σ, TRUE); R uses pnorm(x, mean = μ, sd = σ).`,
      };
    }

    case 'm3_uniform_interval':
    case 'm3_uniform_cdf': {
      const a = getNum(['a', 'min', 'lower'], 0);
      const b = getNum(['b', 'max', 'upper'], 10);
      const x = getNum(['x', 'value'], 5);
      const aStr = formatParamNumber(a);
      const bStr = formatParamNumber(b);
      const xStr = formatParamNumber(x);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=(${xStr} - ${aStr}) / (${bStr} - ${aStr})`,
          result: resultStr,
          description: `Uniform probability (x - a) / (b - a) on [${aStr}, ${bStr}]`,
        },
        r: {
          software: 'r',
          call: `punif(${xStr}, min = ${aStr}, max = ${bStr})`,
          result: resultStr,
          description: `Cumulative uniform probability punif()`,
        },
        explanation: `Uniform CDF on [a, b]: Excel uses =(x - a)/(b - a); R uses punif(x, min, max).`,
      };
    }

    case 'm3_exponential_cdf': {
      const lam = getNum(['lambda', 'rate'], 0.5);
      const x = getNum(['x', 'time'], 3);
      const lamStr = formatParamNumber(lam);
      const xStr = formatParamNumber(x);
      const isUpper = getStr(['tail', 'direction']).includes('upper') ||
                      getStr(['tail', 'direction']).includes('greater');

      if (isUpper) {
        return {
          recipeId: recipe.id,
          recipeTitle: recipe.title,
          task: recipe.statisticalTask,
          excel: {
            software: 'excel',
            call: `=1 - EXPON.DIST(${xStr}, ${lamStr}, TRUE)`,
            result: resultStr,
            description: `Exponential survival probability P(X > ${xStr}) = e^(-λx)`,
          },
          r: {
            software: 'r',
            call: `pexp(${xStr}, rate = ${lamStr}, lower.tail = FALSE)`,
            result: resultStr,
            description: `Exponential upper tail via lower.tail = FALSE`,
          },
          explanation: `Exponential survival probability P(X > ${xStr}) = e^(-${lamStr} × ${xStr}).`,
        };
      }

      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=EXPON.DIST(${xStr}, ${lamStr}, TRUE)`,
          result: resultStr,
          description: `Exponential CDF P(X <= ${xStr}) with rate λ = ${lamStr}`,
        },
        r: {
          software: 'r',
          call: `pexp(${xStr}, rate = ${lamStr})`,
          result: resultStr,
          description: `Cumulative exponential CDF pexp()`,
        },
        explanation: `Exponential CDF: Excel uses =EXPON.DIST(x, λ, TRUE); R uses pexp(x, rate = λ).`,
      };
    }

    // ----------------------------------------------------
    // MODULE 4: CLT, CONFIDENCE INTERVALS & SAMPLE SIZE
    // ----------------------------------------------------
    case 'm4_clt_sampling_dist_mean':
    case 'm4_standard_error_mean': {
      const sigma = getNum(['sigma', 'sd', 'population_sd', 's'], 10);
      const n = getNum(['n', 'sample_size'], 25);
      const sigStr = formatParamNumber(sigma);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=${sigStr} / SQRT(${n})`,
          result: resultStr,
          description: `Standard error of the mean SE = σ / sqrt(n)`,
        },
        r: {
          software: 'r',
          call: `${sigStr} / sqrt(${n})`,
          result: resultStr,
          description: `Sample mean standard error calculation`,
        },
        explanation: `Standard error SE = σ / √n = ${sigStr} / √${n}.`,
      };
    }

    case 'm4_ci_mean_z':
    case 'm4_confidence_interval_mean_z': {
      const alpha = getNum(['alpha', 'sig_level'], 0.05);
      const sigma = getNum(['sigma', 'sd', 'population_sd'], 12);
      const n = getNum(['n', 'sample_size'], 36);
      const alphaStr = formatParamNumber(alpha);
      const sigStr = formatParamNumber(sigma);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=CONFIDENCE.NORM(${alphaStr}, ${sigStr}, ${n})`,
          result: resultStr,
          description: `Z margin of error E = z_(α/2) * (σ / sqrt(n))`,
        },
        r: {
          software: 'r',
          call: `qnorm(1 - ${alphaStr}/2) * ${sigStr} / sqrt(${n})`,
          result: resultStr,
          description: `Margin of error calculation via standard normal critical value`,
        },
        explanation: `Normal confidence margin: Excel uses =CONFIDENCE.NORM(α, σ, n); R multiplies qnorm(1 - α/2) by SE.`,
      };
    }

    case 'm4_ci_mean_t':
    case 'm4_confidence_interval_mean_t': {
      const alpha = getNum(['alpha', 'sig_level'], 0.05);
      const s = getNum(['s', 'sd', 'sample_sd'], 8.5);
      const n = getNum(['n', 'sample_size'], 16);
      const alphaStr = formatParamNumber(alpha);
      const sStr = formatParamNumber(s);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=CONFIDENCE.T(${alphaStr}, ${sStr}, ${n})`,
          result: resultStr,
          description: `Student's t margin of error E = t_(α/2, n-1) * (s / sqrt(n))`,
        },
        r: {
          software: 'r',
          call: `qt(1 - ${alphaStr}/2, df = ${n - 1}) * ${sStr} / sqrt(${n})`,
          result: resultStr,
          description: `Margin of error via Student t critical quantile qt()`,
        },
        explanation: `Student's t confidence margin: Excel uses =CONFIDENCE.T(α, s, n); R uses qt(1 - α/2, df = n-1) * SE.`,
      };
    }

    case 'm4_sample_size_mean': {
      const sigma = getNum(['sigma', 'sd', 'population_sd'], 15);
      const moe = getNum(['margin_of_error', 'moe', 'e', 'target_se'], 3);
      const sigStr = formatParamNumber(sigma);
      const moeStr = formatParamNumber(moe);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=ROUNDUP((1.96 * ${sigStr} / ${moeStr})^2, 0)`,
          result: resultStr,
          description: `Sample size required for 95% confidence ceiling n >= (z*σ / E)^2`,
        },
        r: {
          software: 'r',
          call: `ceiling((qnorm(0.975) * ${sigStr} / ${moeStr})^2)`,
          result: resultStr,
          description: `Sample size ceiling calculation using qnorm(0.975)`,
        },
        explanation: `Sample size formula n = ⌈(z × σ / E)²⌉: always rounded up to next integer.`,
      };
    }

    // ----------------------------------------------------
    // MODULE 5: HYPOTHESIS TESTING
    // ----------------------------------------------------
    case 'm5_hyp_test_mean_t':
    case 'm5_one_sample_t_test': {
      const xBar = getNum(['x_bar', 'mean', 'sample_mean'], 105);
      const mu0 = getNum(['mu0', 'null_value', 'hypothesized_mean'], 100);
      const s = getNum(['s', 'sd', 'sample_sd'], 12);
      const n = getNum(['n', 'sample_size'], 25);
      const xBarStr = formatParamNumber(xBar);
      const mu0Str = formatParamNumber(mu0);
      const sStr = formatParamNumber(s);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=(${xBarStr} - ${mu0Str}) / (${sStr} / SQRT(${n}))`,
          result: resultStr,
          description: `One-sample t test statistic t = (x̄ - μ0) / (s / sqrt(n))`,
        },
        r: {
          software: 'r',
          call: `(${xBarStr} - ${mu0Str}) / (${sStr} / sqrt(${n}))`,
          result: resultStr,
          description: `Analytical t test statistic calculation in R`,
        },
        explanation: `Test statistic t = (x̄ - μ0) / (s / √n) = (${xBarStr} - ${mu0Str}) / (${sStr} / √${n}).`,
      };
    }

    case 'm5_excel_tail_function_guide':
    case 'm5_p_value_t': {
      const tVal = getNum(['t', 't_stat', 'test_stat'], 2.15);
      const df = getNum(['df', 'n'], 24);
      const tStr = formatParamNumber(tVal);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=T.DIST.2T(ABS(${tStr}), ${df})`,
          result: resultStr,
          description: `Two-tailed Student's t p-value with |t| = ${tStr} and df = ${df}`,
        },
        r: {
          software: 'r',
          call: `2 * pt(-abs(${tStr}), df = ${df})`,
          result: resultStr,
          description: `Two-tailed Student's t distribution tail probability`,
        },
        explanation: `Two-tailed t p-value: Excel uses =T.DIST.2T(|t|, df); R uses 2 * pt(-|t|, df).`,
      };
    }

    // ----------------------------------------------------
    // MODULE 6: REGRESSION ANALYSIS & INFERENCE
    // ----------------------------------------------------
    case 'm6_reg_bivariate_fit':
    case 'm6_reg_slope_inference_t':
    case 'm6_mreg_equation_prediction': {
      const slope = getNum(['slope', 'b1'], 1.8);
      const intercept = getNum(['intercept', 'b0'], 12.8);
      const xPred = getNum(['x_pred', 'x', 'x_new'], 6.0);
      const mStr = formatParamNumber(slope);
      const bStr = formatParamNumber(intercept);
      const xStr = formatParamNumber(xPred);

      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=${bStr} + ${mStr} * ${xStr}`,
          result: resultStr,
          description: `Fitted OLS prediction ŷ = b0 + b1 * x at x = ${xStr}`,
        },
        r: {
          software: 'r',
          call: `${bStr} + ${mStr} * ${xStr}  # or predict(model, newdata = data.frame(x = ${xStr}))`,
          result: resultStr,
          description: `Model fitted evaluation in R`,
        },
        explanation: `Predicted value ŷ = ${bStr} + ${mStr}(${xStr}).`,
      };
    }

    case 'm6_reg_anova_table':
    case 'm6_mreg_adj_r2_f': {
      const ssr = getNum(['ssr'], 32.4);
      const sst = getNum(['sst'], 40.0);
      const ssrStr = formatParamNumber(ssr);
      const sstStr = formatParamNumber(sst);
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: `=${ssrStr} / ${sstStr}`,
          result: resultStr,
          description: `Coefficient of determination R² = SSR / SST`,
        },
        r: {
          software: 'r',
          call: `${ssrStr} / ${sstStr}  # or summary(model)$r.squared`,
          result: resultStr,
          description: `Proportion of total variation explained by model`,
        },
        explanation: `R² = SSR / SST = ${ssrStr} / ${sstStr}.`,
      };
    }

    // ----------------------------------------------------
    // DEFAULT GENERIC INTERPOLATION
    // ----------------------------------------------------
    default: {
      const excelCall = recipe.excel.exampleFormula || recipe.excel.genericSyntax;
      const rCall = recipe.r.exampleFormula || recipe.r.genericSyntax;
      return {
        recipeId: recipe.id,
        recipeTitle: recipe.title,
        task: recipe.statisticalTask,
        excel: {
          software: 'excel',
          call: excelCall,
          result: resultStr,
          description: recipe.excel.notes || 'Canonical Excel function call.',
        },
        r: {
          software: 'r',
          call: rCall,
          result: resultStr,
          description: recipe.r.notes || 'Canonical R function call.',
        },
        explanation: `Software syntax for ${recipe.title}.`,
      };
    }
  }
}
