import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { QuestionStrip } from './QuestionStrip';

describe('QuestionStrip', () => {
  it('renders all question pills with correct active state', () => {
    const onSelect = vi.fn();
    render(<QuestionStrip totalQuestions={4} currentQuestionIndex={2} onSelectQuestion={onSelect} />);

    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(4);
    expect(buttons[2]).toHaveClass('active');
    expect(buttons[0]).not.toHaveClass('active');
  });

  it('triggers onSelectQuestion callback when a pill is clicked', () => {
    const onSelect = vi.fn();
    render(<QuestionStrip totalQuestions={3} currentQuestionIndex={0} onSelectQuestion={onSelect} />);

    const buttons = screen.getAllByRole('button');
    fireEvent.click(buttons[1]);
    expect(onSelect).toHaveBeenCalledWith(1);
  });
});
