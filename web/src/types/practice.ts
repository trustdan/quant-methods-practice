export type StageProgressStatus = 'unvisited' | 'active' | 'retry' | 'completed';

export type StageKind = 'choice' | 'numeric';

export interface PublicOptionView {
  id: string;
  text_markdown: string;
}

export interface PublicNumericPolicyView {
  version: number;
  allowed_forms: string[];
  display_decimals?: number;
}

export interface ExpectedAnswerView {
  kind: StageKind;
  option_id?: string;
  value?: number;
  units?: string;
}

export interface SubmittedAnswerDTO {
  kind: StageKind;
  option_id?: string;
  numeric_raw?: string;
  normalized_value?: number;
  form?: string;
}

export interface StageAttemptDTO {
  id: string;
  session_id: string;
  instance_id: string;
  stage_id: string;
  attempt_number: number;
  submitted_answer: SubmittedAnswerDTO;
  assistance: string[];
  is_correct: boolean;
  feedback_markdown: string;
  created_at: string;
}

export interface PublicStageView {
  id: string;
  number: number;
  label: string;
  kind: StageKind;
  prompt_markdown: string;
  status: StageProgressStatus;
  options: PublicOptionView[];
  numeric_policy?: PublicNumericPolicyView;
  attempt_count: number;
  max_attempts: number;
  is_correct?: boolean;
  solved_on_retry: boolean;
  revealed: boolean;
  active_hint?: string;
  last_feedback?: string;
  explanation_markdown?: string;
  expected_answer?: ExpectedAnswerView;
  attempts: StageAttemptDTO[];
  invalid_input_notice?: string;
  draft_answer?: SubmittedAnswerDTO;
}

export interface StageSummary {
  stage_number: number;
  stage_id: string;
  label: string;
  prompt_markdown: string;
  learner_answer: string;
  canonical_answer: string;
  outcome: 'first_try' | 'retry' | 'revealed' | 'pending';
  misconception_triggered?: string;
  explanation_markdown: string;
}

export interface CanonicalDerivation {
  n: number;
  p: number;
  k: number;
  canonical_probability: number;
  canonical_rational?: string;
  mean: number;
  variance: number;
  std_dev: number;
  expression_tex: string;
  calculation_tex: string;
  event_tex: string;
}

export interface DrillRecap {
  title: string;
  scenario_markdown: string;
  total_stages: number;
  first_try_count: number;
  retry_count: number;
  revealed_count: number;
  canonical_derivation?: CanonicalDerivation;
  stage_summaries: StageSummary[];
}

export interface PublicQuestionInfo {
  index: number;
  title: string;
  status: 'pending' | 'in_progress' | 'completed' | 'skipped';
}

export interface PublicSessionView {
  id: string;
  template_id: string;
  revision: number;
  title: string;
  scenario_markdown: string;
  parameters: Record<string, any>;
  assumptions: string[];
  current_stage_index: number;
  completed: boolean;
  stages: PublicStageView[];
  recap?: DrillRecap;
  updated_at: string;
  current_question_index?: number;
  total_questions?: number;
  questions?: PublicQuestionInfo[];
  all_completed?: boolean;
}

export interface SessionCommandDTO {
  command_id: string;
  expected_revision: number;
  type: 'submit_answer' | 'request_hint' | 'navigate_stage' | 'navigate_question' | 'reset_drill' | 'save_draft' | 'clear_draft';
  stage_id?: string;
  answer?: SubmittedAnswerDTO;
  target_stage_index?: number;
  target_question_index?: number;
}

export interface CommandResultDTO {
  success: boolean;
  command_id: string;
  session_state: PublicSessionView;
  error_message?: string;
  invalid_input?: boolean;
}
