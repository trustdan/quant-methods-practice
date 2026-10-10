import type { SubmittedAnswerDTO as SubmittedAnswer, PublicNumericPolicyView } from '../../types/practice';

export type AnswerMap = Record<string, SubmittedAnswer>;
export interface Dataset {
  rows: { experiment_id: string; successes: number }[];
  rule_version: number;
  reviewer?: string;
  source_note?: string;
}
export interface CasePreview {
  dataset: Dataset;
  questions: { id: string; title: string; scenario_markdown: string; assumptions: string[]; stages: {
    id: string; prompt_markdown: string; explanation_markdown: string;
    expected_answer: { option_id?: string; value?: number }; options: { id: string; text_markdown: string; hint_markdown?: string }[];
    numeric_policy?: { version: number; absolute_tolerance: number; relative_tolerance: number };
  }[] }[];
}
export interface Worksheet {
  id: string;
  revision: number;
  mode: 'full_solution' | 'dataset';
  status: 'draft' | 'retry' | 'completed';
  evidence: string;
  updated_at: string;
  dataset?: Dataset;
  questions: { id: string; title: string; template_id: string; template_version: number; seed: number; scenario_markdown: string; assumptions: string[] }[];
  items: {
    key: string; question_id: string; stage_id: string; kind: 'choice' | 'numeric'; prompt_markdown: string;
    status: string; options: { id: string; text_markdown: string }[]; draft_answer?: SubmittedAnswer;
    numeric_policy?: PublicNumericPolicyView & { absolute_tolerance: number; relative_tolerance: number };
    attempts: { attempt_number: number; is_correct: boolean; feedback_markdown: string; submitted_answer: SubmittedAnswer; assistance: string[] }[] | null;
    explanation_markdown?: string;
  }[];
}
