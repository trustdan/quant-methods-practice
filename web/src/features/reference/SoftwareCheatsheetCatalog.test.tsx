import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SoftwareCheatsheetCatalog } from './SoftwareCheatsheetCatalog';

describe('SoftwareCheatsheetCatalog', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders all recipes by default and displays toolbar', () => {
    render(<SoftwareCheatsheetCatalog />);

    expect(screen.getByText(/Statistical Software Cheatsheet \(Excel & R\)/)).toBeInTheDocument();
    expect(screen.getByText(/All Modules \(50\)/)).toBeInTheDocument();
  });

  it('filters recipes by module pill', () => {
    render(<SoftwareCheatsheetCatalog />);

    const m6Btn = screen.getByRole('button', { name: /M6 Regression/ });
    fireEvent.click(m6Btn);

    expect(screen.getByText(/Simple Linear Regression: OLS Slope/)).toBeInTheDocument();
    expect(screen.queryByText(/Measures of Center: Mean, Median, and Mode/)).not.toBeInTheDocument();
  });

  it('switches between Side-by-Side, Excel Only, and R Only display modes', () => {
    render(<SoftwareCheatsheetCatalog />);

    const rOnlyBtn = screen.getByRole('button', { name: /R Only/ });
    fireEvent.click(rOnlyBtn);

    expect(screen.queryByText(/📊 Excel:/)).not.toBeInTheDocument();
    expect(screen.getAllByText(/🔷 R:/).length).toBeGreaterThan(0);

    const sideBySideBtn = screen.getByRole('button', { name: /Side-by-Side/ });
    fireEvent.click(sideBySideBtn);

    expect(screen.getAllByText(/📊 Excel:/).length).toBeGreaterThan(0);
    expect(screen.getAllByText(/🔷 R:/).length).toBeGreaterThan(0);
  });

  it('triggers window.print on Print button click', () => {
    const printSpy = vi.spyOn(window, 'print').mockImplementation(() => {});
    render(<SoftwareCheatsheetCatalog />);

    const printBtn = screen.getByRole('button', { name: /Print \/ PDF/ });
    fireEvent.click(printBtn);

    expect(printSpy).toHaveBeenCalledTimes(1);
  });

  it('expands recipe details showing argument descriptions and pitfalls', () => {
    render(<SoftwareCheatsheetCatalog />);

    const detailButtons = screen.getAllByRole('button', { name: /Details ▼/ });
    fireEvent.click(detailButtons[0]);

    expect(screen.getAllByText(/Parameters & Arguments:/).length).toBeGreaterThan(0);
  });
});
