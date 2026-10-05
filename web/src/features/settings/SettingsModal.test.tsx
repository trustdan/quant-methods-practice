import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SettingsModal } from './SettingsModal';

describe('SettingsModal', () => {
  it('does not render when closed', () => {
    const { container } = render(
      <SettingsModal isOpen={false} onClose={vi.fn()} onStartSession={vi.fn()} />
    );
    expect(container.firstChild).toBeNull();
  });

  it('allows selecting question count, module filters, and starting session', async () => {
    const handleClose = vi.fn();
    const handleStart = vi.fn().mockResolvedValue(undefined);

    render(
      <SettingsModal
        isOpen={true}
        onClose={handleClose}
        onStartSession={handleStart}
        currentSettings={{ question_count: 10, module_ids: ['module_1'], intensity: 'standard' }}
      />
    );

    expect(screen.getByText('Session Settings & Module Picker')).toBeInTheDocument();

    // Select 5 questions
    const btn5 = screen.getByRole('button', { name: '5 Questions' });
    fireEvent.click(btn5);

    // Toggle module_2 checkbox
    const mod2Checkbox = screen.getByLabelText(/Module 2/);
    fireEvent.click(mod2Checkbox);

    // Click Start New Session
    const startBtn = screen.getByRole('button', { name: 'Start New Session' });
    fireEvent.click(startBtn);

    expect(handleStart).toHaveBeenCalledWith({
      question_count: 5,
      module_ids: ['module_1', 'module_2'],
      intensity: 'standard',
      seed: undefined,
    });
  });
});
