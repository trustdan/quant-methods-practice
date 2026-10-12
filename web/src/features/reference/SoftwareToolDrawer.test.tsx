import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SoftwareToolDrawer } from './SoftwareToolDrawer';
import { SoftwareRecipe } from './softwareRecipes';

const mockRecipe: SoftwareRecipe = {
  id: 'test_binom_recipe',
  moduleId: 'module_2',
  familyId: 'binomial',
  title: 'Binomial Probability Distribution',
  statisticalTask: 'Calculate exact point probability of k successes in n independent Bernoulli trials.',
  relevantTemplateIds: ['test_binom_template'],
  applicableKeywords: ['binomial', 'bernoulli'],
  excel: {
    functionName: 'BINOM.DIST',
    genericSyntax: '=BINOM.DIST(number_s, trials, prob_s, FALSE)',
    argumentDescriptions: [
      { name: 'number_s', description: 'Number of successes (k).' },
      { name: 'trials', description: 'Number of independent Bernoulli trials (n).' },
      { name: 'prob_s', description: 'Success probability per trial (p).' },
      { name: 'cumulative', description: 'FALSE for exact PMF P(X = k); TRUE for cumulative CDF P(X <= k).' },
    ],
    notes: 'Requires 0 <= prob_s <= 1 and positive integers for trials.',
    commonPitfalls: 'Forgetting FALSE in exact PMF questions defaults to cumulative.',
    exampleFormula: '=BINOM.DIST(3, 10, 0.40, FALSE)',
  },
  r: {
    functionName: 'dbinom, pbinom',
    genericSyntax: 'dbinom(x, size, prob)',
    argumentDescriptions: [
      { name: 'x', description: 'Number of successes (k).' },
      { name: 'size', description: 'Number of trials (n).' },
      { name: 'prob', description: 'Probability of success (p).' },
    ],
    notes: 'Use dbinom for point probability; pbinom for cumulative CDF.',
    commonPitfalls: 'Parameter order in R is (x, size, prob) vs Excel (number_s, trials, prob_s).',
    exampleFormula: 'dbinom(3, size = 10, prob = 0.40)',
  },
};

describe('SoftwareToolDrawer Component', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('renders collapsed by default with accessible trigger and shortcut indicator', () => {
    render(<SoftwareToolDrawer recipe={mockRecipe} />);

    const trigger = screen.getByRole('button', { name: /Software Syntax Guide/i });
    expect(trigger).toBeInTheDocument();
    expect(trigger).toHaveAttribute('aria-expanded', 'false');

    // Shortcut badge 'u' is present
    expect(screen.getByText('u')).toBeInTheDocument();

    // Content should not be visible when collapsed
    expect(screen.queryByTestId('software-syntax-code')).not.toBeInTheDocument();
  });

  it('expands when clicked and renders Excel formula by default', () => {
    render(<SoftwareToolDrawer recipe={mockRecipe} />);

    const trigger = screen.getByRole('button', { name: /Software Syntax Guide/i });
    fireEvent.click(trigger);

    expect(trigger).toHaveAttribute('aria-expanded', 'true');

    // Recipe title and task
    expect(screen.getByText('Binomial Probability Distribution')).toBeInTheDocument();
    expect(screen.getByText(/calculate exact point probability/i)).toBeInTheDocument();

    // Excel formula rendered
    const code = screen.getByTestId('software-syntax-code');
    expect(code).toHaveTextContent('=BINOM.DIST(number_s, trials, prob_s, FALSE)');

    // Argument descriptions rendered
    expect(screen.getByText('number_s')).toBeInTheDocument();
    expect(screen.getByText('Number of successes (k).')).toBeInTheDocument();

    // Common pitfall callout rendered
    expect(screen.getByText(/forgetting false in exact pmf questions/i)).toBeInTheDocument();
  });

  it('switches to R tab on click and updates syntax and arguments', () => {
    render(<SoftwareToolDrawer recipe={mockRecipe} defaultExpanded={true} />);

    const rTab = screen.getByRole('tab', { name: /R \/ RStudio/i });
    expect(rTab).toHaveAttribute('aria-selected', 'false');

    fireEvent.click(rTab);

    expect(rTab).toHaveAttribute('aria-selected', 'true');
    const code = screen.getByTestId('software-syntax-code');
    expect(code).toHaveTextContent('dbinom(x, size, prob)');

    // Arguments updated to R parameters
    expect(screen.getByText('size')).toBeInTheDocument();
    expect(screen.getByText('Number of trials (n).')).toBeInTheDocument();

    // Local storage preference updated
    expect(localStorage.getItem('quant_preferred_software_tool')).toBe('r');
  });

  it('honors existing stored tool preference from localStorage on mount', () => {
    localStorage.setItem('quant_preferred_software_tool', 'r');

    render(<SoftwareToolDrawer recipe={mockRecipe} defaultExpanded={true} />);

    const rTab = screen.getByRole('tab', { name: /R \/ RStudio/i });
    expect(rTab).toHaveAttribute('aria-selected', 'true');

    const code = screen.getByTestId('software-syntax-code');
    expect(code).toHaveTextContent('dbinom(x, size, prob)');
  });

  it('supports keyboard arrow navigation between Excel and R tabs', () => {
    render(<SoftwareToolDrawer recipe={mockRecipe} defaultExpanded={true} />);

    const excelTab = screen.getByRole('tab', { name: /Microsoft Excel/i });
    expect(excelTab).toHaveAttribute('aria-selected', 'true');

    // Press ArrowRight on active tab
    fireEvent.keyDown(excelTab, { key: 'ArrowRight' });

    const rTab = screen.getByRole('tab', { name: /R \/ RStudio/i });
    expect(rTab).toHaveAttribute('aria-selected', 'true');

    // Press ArrowLeft to return to Excel
    fireEvent.keyDown(rTab, { key: 'ArrowLeft' });
    expect(excelTab).toHaveAttribute('aria-selected', 'true');
  });

  it('copies syntax to clipboard on copy button click with visual feedback', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    });

    render(<SoftwareToolDrawer recipe={mockRecipe} defaultExpanded={true} />);

    const copyBtn = screen.getByRole('button', { name: /Copy Syntax/i });
    expect(copyBtn).toHaveTextContent(/copy syntax/i);

    fireEvent.click(copyBtn);

    expect(writeTextMock).toHaveBeenCalledWith('=BINOM.DIST(number_s, trials, prob_s, FALSE)');

    // Visual feedback "Copied!"
    await waitFor(() => {
      expect(screen.getByText('Copied!')).toBeInTheDocument();
    });
  });

  it('renders withheld notice when in timed assessment exam mode', () => {
    render(
      <SoftwareToolDrawer
        recipe={mockRecipe}
        defaultExpanded={true}
        withheld={true}
        withheldReason="Software formulas are withheld during timed assessments."
      />
    );

    // Withheld alert banner
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText('Software Aid Withheld')).toBeInTheDocument();
    expect(screen.getByText('Software formulas are withheld during timed assessments.')).toBeInTheDocument();

    // Formula syntax is NOT shown
    expect(screen.queryByTestId('software-syntax-code')).not.toBeInTheDocument();
  });

  it('dynamically resolves recipe from templateId when recipe prop is not passed directly', () => {
    render(
      <SoftwareToolDrawer
        templateId="binomial_fair_coin_exactly_two"
        defaultExpanded={true}
      />
    );

    // Resolves binomial recipe
    const code = screen.getByTestId('software-syntax-code');
    expect(code).toBeInTheDocument();
    expect(code.textContent).toContain('BINOM.DIST');
  });
});
