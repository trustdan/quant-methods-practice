import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { PracticeDrill } from './PracticeDrill';
import { DEFAULT_BINOMIAL_SESSION } from './defaultSession';
import { PublicSessionView } from '../../types/practice';

describe('PracticeDrill Component', () => {
  it('renders scenario title, description, and assumptions', () => {
    const onSend = vi.fn();
    const onNav = vi.fn();
    const onReset = vi.fn();

    render(
      <PracticeDrill
        session={DEFAULT_BINOMIAL_SESSION}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    expect(screen.getByText('Exactly two heads in four tosses')).toBeInTheDocument();
    expect(screen.getByText(/A fair coin is tossed four times independently/)).toBeInTheDocument();
    expect(screen.getByText(/• Four fixed trials/)).toBeInTheDocument();
    expect(screen.getByText(/• Constant head probability 0.5/)).toBeInTheDocument();
  });

  it('renders stage 1 prompt and choice options', () => {
    const onSend = vi.fn();
    const onNav = vi.fn();
    const onReset = vi.fn();

    render(
      <PracticeDrill
        session={DEFAULT_BINOMIAL_SESSION}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    expect(screen.getByText(/Stage 1\/7: Target/)).toBeInTheDocument();
    expect(screen.getByText(/What should \$X\$ represent\?/)).toBeInTheDocument();
    expect(screen.getByText(/The number of heads in four tosses/)).toBeInTheDocument();
    expect(screen.getByText(/The probability of a head on one toss/)).toBeInTheDocument();
  });

  it('selects choice option and submits command', async () => {
    const onSend = vi.fn().mockResolvedValue({ success: true });
    const onNav = vi.fn();
    const onReset = vi.fn();

    render(
      <PracticeDrill
        session={DEFAULT_BINOMIAL_SESSION}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    const optionBtn = screen.getByText(/The number of heads in four tosses/);
    fireEvent.click(optionBtn);

    const submitBtn = screen.getByRole('button', { name: /Submit Answer/ });
    fireEvent.click(submitBtn);

    expect(onSend).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'submit_answer',
        stage_id: 'define_variable',
        answer: {
          kind: 'choice',
          option_id: 'count_heads',
        },
      })
    );
  });

  it('requests offline hint when hint button clicked', () => {
    const onSend = vi.fn().mockResolvedValue({ success: true });
    const onNav = vi.fn();
    const onReset = vi.fn();

    render(
      <PracticeDrill
        session={DEFAULT_BINOMIAL_SESSION}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    const hintBtn = screen.getByRole('button', { name: /Offline Hint/ });
    fireEvent.click(hintBtn);

    expect(onSend).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'request_hint',
        stage_id: 'define_variable',
      })
    );
  });

  it('displays causal hint when stage is in retry status', () => {
    const onSend = vi.fn();
    const onNav = vi.fn();
    const onReset = vi.fn();

    const retrySession: PublicSessionView = JSON.parse(JSON.stringify(DEFAULT_BINOMIAL_SESSION));
    retrySession.stages[0].status = 'retry';
    retrySession.stages[0].active_hint = 'Which quantity changes from one four-toss experiment to another?';

    render(
      <PracticeDrill
        session={retrySession}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    expect(screen.getByText('1 Retry Remaining')).toBeInTheDocument();
    expect(screen.getByText(/Causal Hint & Guidance:/)).toBeInTheDocument();
    expect(screen.getByText(/Which quantity changes from one four-toss experiment to another\?/)).toBeInTheDocument();
  });

  it('renders numeric input on Stage 6 (Calculate)', () => {
    const onSend = vi.fn().mockResolvedValue({ success: true });
    const onNav = vi.fn();
    const onReset = vi.fn();

    const calcSession: PublicSessionView = JSON.parse(JSON.stringify(DEFAULT_BINOMIAL_SESSION));
    calcSession.current_stage_index = 5;
    calcSession.stages[5].status = 'active';

    render(
      <PracticeDrill
        session={calcSession}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    expect(screen.getByText(/Stage 6\/7: Calculate/)).toBeInTheDocument();
    const input = screen.getByPlaceholderText(/e\.g\. 0\.375, 37\.5%, or 3\/8/);
    expect(input).toBeInTheDocument();

    fireEvent.change(input, { target: { value: '0.375' } });
    const submitBtn = screen.getByRole('button', { name: /^Submit$/ });
    fireEvent.click(submitBtn);

    expect(onSend).toHaveBeenCalledWith(
      expect.objectContaining({
        type: 'submit_answer',
        stage_id: 'calculate_probability',
        answer: {
          kind: 'numeric',
          numeric_raw: '0.375',
        },
      })
    );
  });

  it('renders invalid input diagnostic notice without consuming attempts', () => {
    const onSend = vi.fn();
    const onNav = vi.fn();
    const onReset = vi.fn();

    const noticeSession: PublicSessionView = JSON.parse(JSON.stringify(DEFAULT_BINOMIAL_SESSION));
    noticeSession.current_stage_index = 5;
    noticeSession.stages[5].status = 'active';
    noticeSession.stages[5].invalid_input_notice = 'Commas are ambiguous; please use period as decimal point.';

    render(
      <PracticeDrill
        session={noticeSession}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    expect(screen.getByText(/Input Format Notice:/)).toBeInTheDocument();
    expect(screen.getByText(/Commas are ambiguous; please use period as decimal point\./)).toBeInTheDocument();
  });

  it('renders full drill recap when session is completed', () => {
    const onSend = vi.fn();
    const onNav = vi.fn();
    const onReset = vi.fn();

    const completedSession: PublicSessionView = JSON.parse(JSON.stringify(DEFAULT_BINOMIAL_SESSION));
    completedSession.completed = true;
    completedSession.recap = {
      title: 'Exactly two heads in four tosses',
      scenario_markdown: 'A fair coin is tossed four times independently.',
      total_stages: 7,
      first_try_count: 6,
      retry_count: 1,
      revealed_count: 0,
      canonical_derivation: {
        n: 4,
        p: 0.5,
        k: 2,
        canonical_probability: 0.375,
        canonical_rational: '3/8',
        mean: 2,
        variance: 1,
        std_dev: 1,
        expression_tex: '\\binom{4}{2}(0.5)^2(0.5)^2',
        calculation_tex: '6 \\times 0.0625 = 0.375',
        event_tex: 'X = 2',
      },
      stage_summaries: [
        {
          stage_number: 1,
          stage_id: 'define_variable',
          label: 'Target',
          prompt_markdown: 'What should $X$ represent?',
          learner_answer: 'The number of heads in four tosses',
          canonical_answer: 'The number of heads in four tosses',
          outcome: 'first_try',
          explanation_markdown: '$X$ counts heads in the whole experiment.',
        },
      ],
    };

    render(
      <PracticeDrill
        session={completedSession}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    expect(screen.getByText('Drill Completed')).toBeInTheDocument();
    expect(screen.getByText('Canonical Mathematical Derivation')).toBeInTheDocument();
    expect(screen.getByText('7 / 7')).toBeInTheDocument();
    expect(screen.getByText('6')).toBeInTheDocument();
    expect(screen.getByText(/Stage 1: Target/)).toBeInTheDocument();
  });

  it('preserves unsent answer draft when navigating between stages', () => {
    const onSend = vi.fn().mockResolvedValue({ success: true });
    const onNav = vi.fn();
    const onReset = vi.fn();

    // Start with session where stage 1 and stage 2 are active
    const session: PublicSessionView = JSON.parse(JSON.stringify(DEFAULT_BINOMIAL_SESSION));
    session.stages[0].status = 'active';
    session.stages[1].status = 'active';

    const { rerender } = render(
      <PracticeDrill
        session={session}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    // Select option on stage 0
    const opt1 = screen.getByText(/The number of heads in four tosses/);
    fireEvent.click(opt1);
    expect(screen.getByText('Option selected')).toBeInTheDocument();

    // Navigate to stage 1 (Model)
    const sessionStage1 = { ...session, current_stage_index: 1 };
    rerender(
      <PracticeDrill
        session={sessionStage1}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    expect(screen.getByText(/Stage 2\/7: Model/)).toBeInTheDocument();

    // Navigate back to stage 0 (Target)
    const sessionStage0 = { ...session, current_stage_index: 0 };
    rerender(
      <PracticeDrill
        session={sessionStage0}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    expect(screen.getByText(/Stage 1\/7: Target/)).toBeInTheDocument();
    // Invariant: Unsent answer is preserved across navigation!
    expect(screen.getByText('Option selected')).toBeInTheDocument();
  });

  it('shows locked stage notice when clicking an unavailable future stage', () => {
    const onSend = vi.fn();
    const onNav = vi.fn();
    const onReset = vi.fn();

    render(
      <PracticeDrill
        session={DEFAULT_BINOMIAL_SESSION}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    // Default session: only Stage 1 is active, Stages 2-7 are unvisited (locked)
    const tabs = screen.getAllByRole('tab');
    // Click Stage 4
    fireEvent.click(tabs[3]);

    // Explains why it is unavailable without navigating
    expect(onNav).not.toHaveBeenCalled();
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/Complete previous stages before jumping ahead/)).toBeInTheDocument();
  });

  it('allows moving choice selection with j and k keys', () => {
    const onSend = vi.fn();
    const onNav = vi.fn();
    const onReset = vi.fn();

    render(
      <PracticeDrill
        session={DEFAULT_BINOMIAL_SESSION}
        onSendCommand={onSend}
        onNavigateStage={onNav}
        onResetDrill={onReset}
      />
    );

    // Press 'j' to select first option / move down
    fireEvent.keyDown(window, { key: 'j' });
    expect(screen.getByText('Option selected')).toBeInTheDocument();

    // Press 'k' to move selection up
    fireEvent.keyDown(window, { key: 'k' });
    expect(screen.getByText('Option selected')).toBeInTheDocument();
  });
});
