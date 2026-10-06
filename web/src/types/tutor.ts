export type TutorAction = 'hint' | 'explain' | 'follow_up';

export type FollowUpKind =
  | 'explain_differently'
  | 'worked_example'
  | 'why_condition_matters'
  | 'compare_concepts'
  | 'custom';

export type TutorEventType =
  | 'started'
  | 'text_delta'
  | 'complete'
  | 'fallback'
  | 'cancelled'
  | 'error';

export interface TutorEventDTO {
  type: TutorEventType;
  request_id: string;
  session_id?: string;
  instance_id?: string;
  stage_id?: string;
  delta?: string;
  text?: string;
  fallback_label?: string;
  error?: string;
  timestamp: string;
}

export interface ProviderInfoDTO {
  title: string;
  concepts: string[];
  provider: string;
  model: string;
  route: string;
  thread_id?: string;
  parent_note_id?: string;
  fallback_label?: string;
  advisory_status?: string;
}

export interface SavedExplanationDTO {
  id: string;
  raw_markdown: string;
  origin_instance_id: string;
  origin_stage_id: string;
  topic: string;
  provider_info: ProviderInfoDTO;
  created_at: string;
  updated_at: string;
}

export interface TutorDraftDTO {
  id: string;
  context_json: string;
  recovery_text: string;
  updated_at: string;
}
