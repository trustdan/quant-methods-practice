import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { Exams } from './Exams';
import { PublicExamView, ExamSummary } from './types';

const mockExam: PublicExamView = {
  id: 'exam_test_1',
  title: 'Midterm Test',
  revision: 1,
  status: 'active',
  duration_minutes: 30,
  allowed_aid: 'none',
  started_at: new Date().toISOString(),
  deadline_at: new Date(Date.now() + 1800000).toISOString(),
  time_remaining_seconds: 1800,
  current_question_index: 0,
  total_questions: 1,
  answered_questions: 0,
  questions: [
    {
      index: 0,
      question_id: 'q_1',
      title: 'Fair Coin Exactly Two',
      scenario_markdown: 'Toss a fair coin 4 times.',
      assumptions: ['Independent trials'],
      stages: [
        {
          id: 'stage_choice',
          number: 1,
          kind: 'choice',
          prompt_markdown: 'What distribution?',
          options: [
            { id: 'opt_binom', text_markdown: 'Binomial' },
            { id: 'opt_poiss', text_markdown: 'Poisson' },
          ],
          is_answered: false,
        },
        {
          id: 'stage_numeric',
          number: 2,
          kind: 'numeric',
          prompt_markdown: 'Find probability P(X=2)',
          numeric_policy: {
            version: 1,
            allowed_forms: ['decimal', 'percent', 'fraction'],
          },
          is_answered: false,
        },
      ],
      is_answered: false,
    },
  ],
};

const mockSummary: ExamSummary = {
  id: 'exam_past_1',
  title: 'Previous Exam',
  status: 'completed',
  duration_minutes: 30,
  question_count: 5,
  created_at: '2026-10-09T10:00:00Z',
  score_percent: 80.0,
  correct_stages: 4,
  total_stages: 5,
};

const mockResponse = (data: unknown, status = 200) =>
  ({
    ok: status < 400,
    status,
    json: async () => data,
    text: async () => JSON.stringify(data),
  } as Response);

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('Exams Feature UI', () => {
  it('renders creation setup and lists past assessments in history tab', async () => {
    vi.stubGlobal('fetch', vi.fn(async (url: string) => {
      if (url.endsWith('/api/exams')) {
        return mockResponse([mockSummary]);
      }
      return mockResponse({});
    }));

    render(<Exams active={true} onReturnToPractice={() => {}} />);

    expect(screen.getByText('Assessments & Exam Mode')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Begin Assessment' })).toBeInTheDocument();

    // Switch to history tab
    fireEvent.click(screen.getByRole('button', { name: /History/ }));
    await waitFor(() => {
      expect(screen.getByText('Previous Exam')).toBeInTheDocument();
      expect(screen.getByText('80.0%')).toBeInTheDocument();
    });
  });

  it('starts an assessment and handles question choice and numeric inputs', async () => {
    let currentExam = { ...mockExam };

    vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
      if (url.endsWith('/api/exams') && init?.method === 'POST') {
        return mockResponse(currentExam, 201);
      }
      if (url.includes('/commands') && init?.method === 'POST') {
        const cmd = JSON.parse(String(init?.body));
        if (cmd.type === 'save_answer' && cmd.answer?.option_id) {
          currentExam = {
            ...currentExam,
            revision: 2,
            answered_questions: 1,
            questions: [
              {
                ...currentExam.questions[0],
                is_answered: true,
                stages: [
                  {
                    ...currentExam.questions[0].stages[0],
                    is_answered: true,
                    submitted_answer: cmd.answer,
                  },
                  currentExam.questions[0].stages[1],
                ],
              },
            ],
          };
          return mockResponse(currentExam);
        }
        if (cmd.type === 'submit_exam') {
          return mockResponse({
            ...currentExam,
            status: 'completed',
            report: {
              total_questions: 1,
              total_stages: 2,
              correct_stages: 2,
              score_percent: 100.0,
              time_spent_seconds: 120,
              expired: false,
              question_reports: [
                {
                  index: 0,
                  question_id: 'q_1',
                  template_id: 'test',
                  title: 'Fair Coin Exactly Two',
                  scenario_markdown: 'Toss a fair coin 4 times.',
                  stage_reports: [
                    {
                      id: 'stage_choice',
                      kind: 'choice',
                      prompt_markdown: 'What distribution?',
                      is_correct: true,
                      feedback_markdown: 'Correct! Binomial distribution.',
                      explanation_markdown: 'Independent coin tosses.',
                      expected_answer: { kind: 'choice', option_id: 'opt_binom' },
                    },
                  ],
                  correct_stages: 2,
                  total_stages: 2,
                  is_fully_correct: true,
                },
              ],
            },
          });
        }
      }
      return mockResponse([]);
    }));

    render(<Exams active={true} onReturnToPractice={() => {}} />);

    // Click Begin Assessment
    fireEvent.click(screen.getByRole('button', { name: 'Begin Assessment' }));

    await waitFor(() => {
      expect(screen.getByText('Midterm Test')).toBeInTheDocument();
      expect(screen.getByText('Part 1')).toBeInTheDocument();
    });

    // Active exam must withhold software drawer
    expect(screen.getByText(/Withheld in Timed Assessment/)).toBeInTheDocument();

    // Select Binomial option radio button
    const radio = screen.getByDisplayValue('opt_binom');
    fireEvent.click(radio);

    // Click Submit Exam
    fireEvent.click(screen.getByRole('button', { name: 'Submit Exam' }));

    // Confirm in modal
    await waitFor(() => {
      expect(screen.getByText('Finish & Submit Assessment?')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button', { name: 'Yes, Submit Assessment' }));

    // Report should now be displayed with worked software call unlocked
    await waitFor(() => {
      expect(screen.getByText('100.0%')).toBeInTheDocument();
      expect(screen.getByText('Detailed Question Review')).toBeInTheDocument();
      expect(screen.getByText('Download Markdown Report')).toBeInTheDocument();
      expect(screen.getByText(/Worked Software Call/)).toBeInTheDocument();
    });
  });

  it('strictly withholds hints, solutions, and feedback during active assessment', async () => {
    vi.stubGlobal('fetch', vi.fn(async (url: string, init?: RequestInit) => {
      if (url.endsWith('/api/exams') && init?.method === 'POST') {
        return mockResponse(mockExam, 201);
      }
      return mockResponse([]);
    }));

    render(<Exams active={true} onReturnToPractice={() => {}} />);

    // Begin assessment
    fireEvent.click(screen.getByRole('button', { name: 'Begin Assessment' }));

    await waitFor(() => {
      expect(screen.getByText('Midterm Test')).toBeInTheDocument();
    });

    // Verify hints and solutions are not in DOM
    expect(screen.queryByText(/hint/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/explanation/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/expected/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/solution/i)).not.toBeInTheDocument();

    // Verify software aids are withheld
    expect(screen.getByText(/Withheld in Timed Assessment/)).toBeInTheDocument();

    // Verify option text is rendered cleanly
    expect(screen.getByText('Binomial')).toBeInTheDocument();
    expect(screen.getByText('Poisson')).toBeInTheDocument();
  });
});
