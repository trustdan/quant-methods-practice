import { describe, it, expect } from 'vitest';
import { generateWorkedSoftwareCalls, formatParamNumber } from './softwareWorkedCalls';
import { getRecipeById } from './softwareRecipes';

describe('softwareWorkedCalls', () => {
  it('formats parameter numbers cleanly', () => {
    expect(formatParamNumber(4)).toBe('4');
    expect(formatParamNumber(0.375)).toBe('0.375');
    expect(formatParamNumber(0.50000)).toBe('0.5');
    expect(formatParamNumber(60000.0)).toBe('60000');
    expect(formatParamNumber('2.5')).toBe('2.5');
  });

  it('interpolates binomial exact PMF recipe (m2_binom_exact_pmf)', () => {
    const recipe = getRecipeById('m2_binom_exact_pmf')!;
    expect(recipe).toBeDefined();

    const output = generateWorkedSoftwareCalls({
      recipe,
      parameters: { n: 4, p: 0.5, k: 2 },
      expectedAnswer: 0.375,
    });

    expect(output).not.toBeNull();
    expect(output?.excel.call).toBe('=BINOM.DIST(2, 4, 0.5, FALSE)');
    expect(output?.excel.result).toBe('0.375');
    expect(output?.r.call).toBe('dbinom(2, size = 4, prob = 0.5)');
    expect(output?.r.result).toBe('0.375');
  });

  it('interpolates binomial cumulative CDF upper tail', () => {
    const recipe = getRecipeById('m2_binom_cumulative_cdf')!;
    expect(recipe).toBeDefined();

    const output = generateWorkedSoftwareCalls({
      recipe,
      parameters: { n: 10, p: 0.4, k: 3, tail: 'greater_equal' },
      expectedAnswer: 0.83271,
    });

    expect(output).not.toBeNull();
    expect(output?.excel.call).toBe('=1 - BINOM.DIST(2, 10, 0.4, TRUE)');
    expect(output?.r.call).toContain('pbinom(2, size = 10, prob = 0.4, lower.tail = FALSE)');
  });

  it('interpolates Poisson exact PMF recipe (m2_poisson_exact_pmf)', () => {
    const recipe = getRecipeById('m2_poisson_exact_pmf')!;
    expect(recipe).toBeDefined();

    const output = generateWorkedSoftwareCalls({
      recipe,
      parameters: { lambda: 3.0, k: 2 },
      expectedAnswer: 0.22404,
    });

    expect(output).not.toBeNull();
    expect(output?.excel.call).toBe('=POISSON.DIST(2, 3, FALSE)');
    expect(output?.r.call).toBe('dpois(2, lambda = 3)');
  });

  it('interpolates normal CDF and normal inverse quantiles', () => {
    const normGen = getRecipeById('m3_normal_general_x')!;
    expect(normGen).toBeDefined();

    const cdfOutput = generateWorkedSoftwareCalls({
      recipe: normGen,
      parameters: { mu: 60000, sigma: 8000, x: 50000 },
      expectedAnswer: 0.10565,
    });
    expect(cdfOutput?.excel.call).toBe('=NORM.DIST(50000, 60000, 8000, TRUE)');
    expect(cdfOutput?.r.call).toBe('pnorm(50000, mean = 60000, sd = 8000)');

    const invOutput = generateWorkedSoftwareCalls({
      recipe: normGen,
      parameters: { mu: 60000, sigma: 8000, p: 0.02 },
      expectedAnswer: 43570,
    });
    expect(invOutput?.excel.call).toBe('=NORM.INV(0.02, 60000, 8000)');
    expect(invOutput?.r.call).toContain('qnorm(0.02, mean = 60000, sd = 8000)');
  });

  it('interpolates CLT standard error and confidence interval', () => {
    const seRecipe = getRecipeById('m4_clt_sampling_dist_mean')!;
    expect(seRecipe).toBeDefined();

    const seOutput = generateWorkedSoftwareCalls({
      recipe: seRecipe,
      parameters: { sigma: 10, n: 25 },
      expectedAnswer: 2,
    });
    expect(seOutput?.excel.call).toBe('=10 / SQRT(25)');
    expect(seOutput?.r.call).toBe('10 / sqrt(25)');

    const ciNorm = getRecipeById('m4_ci_mean_z')!;
    expect(ciNorm).toBeDefined();

    const ciOutput = generateWorkedSoftwareCalls({
      recipe: ciNorm,
      parameters: { alpha: 0.05, sigma: 12, n: 36 },
      expectedAnswer: 3.92,
    });
    expect(ciOutput?.excel.call).toBe('=CONFIDENCE.NORM(0.05, 12, 36)');
  });

  it('resolves and interpolates via template lookup query', () => {
    const output = generateWorkedSoftwareCalls({
      lookup: { template_id: 'binomial_fair_coin_exactly_two' },
      parameters: { n: 4, p: 0.5, k: 2 },
      expectedAnswer: { value: 0.375 },
    });

    expect(output).not.toBeNull();
    expect(output?.recipeId).toBe('m2_binom_exact_pmf');
    expect(output?.excel.call).toBe('=BINOM.DIST(2, 4, 0.5, FALSE)');
    expect(output?.excel.result).toBe('0.375');
  });

  it('provides safe fallback for unmapped recipes', () => {
    const recipe = getRecipeById('m6_reg_diagnostics_problems')!;
    expect(recipe).toBeDefined();

    const output = generateWorkedSoftwareCalls({
      recipe,
      expectedAnswer: '0.12',
    });
    expect(output).not.toBeNull();
    expect(output?.excel.call).toBeDefined();
    expect(output?.r.call).toBeDefined();
  });
});
