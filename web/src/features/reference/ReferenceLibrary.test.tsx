import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ReferenceLibrary } from './ReferenceLibrary';

describe('ReferenceLibrary', () => {
  it('renders topic cards and allows filtering by category', () => {
    const handleClose = vi.fn();
    render(<ReferenceLibrary onClose={handleClose} />);

    expect(screen.getByText('Reference Library')).toBeInTheDocument();
    expect(screen.getByText(/Kolmogorov Probability Axioms/)).toBeInTheDocument();
    expect(screen.getByText(/Binomial Distribution/)).toBeInTheDocument();

    // Click "Distributions" category filter
    const distBtn = screen.getByRole('button', { name: 'Distributions' });
    fireEvent.click(distBtn);

    expect(screen.getByText(/Binomial Distribution/)).toBeInTheDocument();
    expect(screen.getByText(/Poisson Distribution/)).toBeInTheDocument();
    expect(screen.queryByText(/Kolmogorov Probability Axioms/)).not.toBeInTheDocument();

    // Click "Return to Practice"
    const returnBtn = screen.getByRole('button', { name: /Return to Practice/ });
    fireEvent.click(returnBtn);
    expect(handleClose).toHaveBeenCalledTimes(1);
  });

  it('filters topics by search query', () => {
    render(<ReferenceLibrary />);

    const searchInput = screen.getByPlaceholderText(/Search formulas/);
    fireEvent.change(searchInput, { target: { value: 'Poisson' } });

    expect(screen.getByText(/Poisson Distribution/)).toBeInTheDocument();
    expect(screen.queryByText(/Binomial Distribution/)).not.toBeInTheDocument();
  });

  it('expands details and shows pitfalls on click', () => {
    render(<ReferenceLibrary />);

    const detailsBtn = screen.getAllByRole('button', { name: /Details ▼/ })[0];
    fireEvent.click(detailsBtn);

    expect(screen.getByText(/Common Exam Pitfalls & Misconceptions/)).toBeInTheDocument();
  });

  it('switches to Software Cheatsheet tab and displays Excel & R recipes', () => {
    render(<ReferenceLibrary />);

    const softwareTab = screen.getByRole('tab', { name: /Software Cheatsheet/ });
    fireEvent.click(softwareTab);

    expect(screen.getByText(/Statistical Software Cheatsheet \(Excel & R\)/)).toBeInTheDocument();
    expect(screen.getByText(/Complete reference catalog of all 50 canonical statistical software recipes/)).toBeInTheDocument();

    // Verify search in cheatsheet
    const searchInput = screen.getByPlaceholderText(/Search Excel & R functions/);
    fireEvent.change(searchInput, { target: { value: 'BINOM.DIST' } });

    expect(screen.getAllByText(/BINOM\.DIST/).length).toBeGreaterThan(0);

    // Verify display mode toggle
    const excelOnlyBtn = screen.getByRole('button', { name: /Excel Only/ });
    fireEvent.click(excelOnlyBtn);
    expect(screen.queryByText(/🔷 R:/)).not.toBeInTheDocument();
  });
});
