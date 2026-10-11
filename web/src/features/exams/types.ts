import { SubmittedAnswerDTO, PublicNumericPolicyView } from '../../types/practice';

export type ExamStatus = 'active' | 'completed' | 'expired' | 'abandoned';

export interface PublicOptionView {
  id: string;
  text_markdown: string;
}

export interface PublicStageView {
  id: string;
  number: number;
  kind: 'choice' | 'numeric';
  prompt_markdown: string;
  options?: PublicOptionView[];
  numeric_policy?: PublicNumericPolicyView;
  submitted_answer?: SubmittedAnswerDTO;
  is_answered: boolean;
}

export interface PublicQuestionView {
  index: number;
  question_id: string;
  title: string;
  scenario_markdown: string;
  assumptions?: string[];
  stages: PublicStageView[];
  is_answered: boolean;
}

export interface ExamStageReport {
  id: string;
  kind: 'choice' | 'numeric';
  prompt_markdown: string;
  options?: Array<{ id: string; text_markdown: string }>;
  submitted_answer?: SubmittedAnswerDTO;
  expected_answer: {
    kind: 'choice' | 'numeric';
    option_id?: string;
    value?: number;
    units?: string;
  };
  is_correct: boolean;
  feedback_markdown: string;
  explanation_markdown: string;
}

export interface ExamQuestionReport {
  index: number;
  question_id: string;
  template_id: string;
  title: string;
  scenario_markdown: string;
  assumptions?: string[];
  stage_reports: ExamStageReport[];
  correct_stages: number;
  total_stages: number;
  is_fully_correct: boolean;
}

export interface ExamReport {
  total_questions: number;
  total_stages: number;
  correct_stages: number;
  score_percent: number;
  time_spent_seconds: number;
  expired: boolean;
  question_reports: ExamQuestionReport[];
}

export interface PublicExamView {
  id: string;
  title: string;
  revision: number;
  status: ExamStatus;
  duration_minutes: number;
  allowed_aid: string;
  started_at: string;
  deadline_at?: string;
  completed_at?: string;
  time_remaining_seconds?: number;
  current_question_index: number;
  total_questions: number;
  answered_questions: number;
  questions: PublicQuestionView[];
  report?: ExamReport;
}

export interface ExamSummary {
  id: string;
  title: string;
  status: ExamStatus;
  duration_minutes: number;
  question_count: number;
  created_at: string;
  completed_at?: string;
  score_percent?: number;
  correct_stages?: number;
  total_stages?: number;
}

export interface ExamCommandDTO {
  command_id: string;
  expected_revision: number;
  type: 'save_answer' | 'navigate' | 'submit_exam' | 'abandon_exam';
  question_index?: number;
  stage_id?: string;
  answer?: SubmittedAnswerDTO;
}
