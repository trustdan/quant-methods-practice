import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { StageStrip, StageInfo } from './StageStrip';

const MOCK_STAGES: StageInfo[] = [
  { id: 'target', number: 1, label: 'Target', completed: true },
  { id: 'model', number: 2, label: 'Model', completed: false },
  { id: 'conditions', number: 3, label: 'Conditions' },
];

describe('StageStrip', () => {
  it('renders all stages with numbers and labels', () => {
    const onSelect = vi.fn();
    render(<StageStrip stages={MOCK_STAGES} currentStageIndex={1} onSelectStage={onSelect} />);

    expect(screen.getByText('1/7')).toBeInTheDocument();
    expect(screen.getByText('Target')).toBeInTheDocument();
    expect(screen.getByText('2/7')).toBeInTheDocument();
    expect(screen.getByText('Model')).toBeInTheDocument();
  });

  it('marks current active stage and completed stages', () => {
    const onSelect = vi.fn();
    render(<StageStrip stages={MOCK_STAGES} currentStageIndex={1} onSelectStage={onSelect} />);

    const buttons = screen.getAllByRole('tab');
    expect(buttons[0]).toHaveClass('completed');
    expect(buttons[1]).toHaveClass('active');
    expect(buttons[2]).not.toHaveClass('active');
  });

  it('calls onSelectStage callback when clicked', () => {
    const onSelect = vi.fn();
    render(<StageStrip stages={MOCK_STAGES} currentStageIndex={0} onSelectStage={onSelect} />);

    const buttons = screen.getAllByRole('tab');
    fireEvent.click(buttons[2]);
    expect(onSelect).toHaveBeenCalledWith(2);
  });
});
