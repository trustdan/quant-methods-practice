import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { WorkedSoftwareCallBox } from './WorkedSoftwareCallBox';
import { getRecipeById } from './softwareRecipes';

describe('WorkedSoftwareCallBox', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('renders worked Excel call by default with substituted numbers and result', () => {
    const recipe = getRecipeById('m2_binom_exact_pmf')!;
    render(
      <WorkedSoftwareCallBox
        recipe={recipe}
        parameters={{ n: 4, p: 0.5, k: 2 }}
        expectedAnswer={0.375}
      />
    );

    expect(screen.getByText(/Worked Software Call/)).toBeInTheDocument();
    expect(screen.getByText(/Exactly two heads in four tosses|Exact Point Probability/i)).toBeInTheDocument();
    expect(screen.getByText(/=BINOM\.DIST\(2, 4, 0\.5, FALSE\)/)).toBeInTheDocument();
    expect(screen.getByText(/=\s*0\.375/)).toBeInTheDocument();
  });

  it('switches to R tab and displays dbinom code with evaluated result', () => {
    const recipe = getRecipeById('m2_binom_exact_pmf')!;
    render(
      <WorkedSoftwareCallBox
        recipe={recipe}
        parameters={{ n: 4, p: 0.5, k: 2 }}
        expectedAnswer={0.375}
      />
    );

    const rTab = screen.getByRole('tab', { name: /R/ });
    fireEvent.click(rTab);

    expect(screen.getByText(/dbinom\(2, size = 4, prob = 0\.5\)/)).toBeInTheDocument();
    expect(screen.getByText(/\[1\]\s*0\.375/)).toBeInTheDocument();
    expect(localStorage.getItem('quant_preferred_software_tool')).toBe('r');
  });

  it('toggles side-by-side mode displaying both Excel and R cards simultaneously', () => {
    const recipe = getRecipeById('m2_binom_exact_pmf')!;
    render(
      <WorkedSoftwareCallBox
        recipe={recipe}
        parameters={{ n: 4, p: 0.5, k: 2 }}
        expectedAnswer={0.375}
        allowSideBySide={true}
      />
    );

    const sideBySideBtn = screen.getByRole('button', { name: /Side-by-Side/ });
    fireEvent.click(sideBySideBtn);

    expect(screen.getByText(/📊 Excel Call/)).toBeInTheDocument();
    expect(screen.getByText(/🔷 R Call/)).toBeInTheDocument();
    expect(screen.getByText(/=BINOM\.DIST\(2, 4, 0\.5, FALSE\)/)).toBeInTheDocument();
    expect(screen.getByText(/dbinom\(2, size = 4, prob = 0\.5\)/)).toBeInTheDocument();
  });

  it('copies software syntax to clipboard on button click', async () => {
    const writeTextMock = vi.fn().mockResolvedValue(undefined);
    Object.assign(navigator, {
      clipboard: { writeText: writeTextMock },
    });

    const recipe = getRecipeById('m2_binom_exact_pmf')!;
    render(
      <WorkedSoftwareCallBox
        recipe={recipe}
        parameters={{ n: 4, p: 0.5, k: 2 }}
        expectedAnswer={0.375}
      />
    );

    const copyBtn = screen.getByRole('button', { name: /Copy Formula/ });
    fireEvent.click(copyBtn);

    expect(writeTextMock).toHaveBeenCalledWith('=BINOM.DIST(2, 4, 0.5, FALSE)');
    expect(await screen.findByText(/✓ Copied/)).toBeInTheDocument();
  });
});
