package domain

// QuestionInstance represents an immutable instantiated question snapshot
// bound to a specific seed and parameter set.
type QuestionInstance struct {
	ID               string                 `json:"id"`
	TemplateID       string                 `json:"template_id"`
	TemplateVersion  int                    `json:"template_version"`
	Seed             int64                  `json:"seed"`
	Parameters       map[string]interface{} `json:"parameters"`
	Title            string                 `json:"title"`
	ScenarioMarkdown string                 `json:"scenario_markdown"`
	Assumptions      []string               `json:"assumptions,omitempty"`
	SettingGroup     string                 `json:"setting_group,omitempty"`
	Stages           []StageInstance        `json:"stages"`
}

// StageInstance represents an instantiated stage within a QuestionInstance,
// preserving the exact option ordering and configuration shown to the learner.
type StageInstance struct {
	ID                  string         `json:"id"`
	Kind                StageKind      `json:"kind"`
	PromptMarkdown      string         `json:"prompt_markdown"`
	Options             []Option       `json:"options"`
	ExpectedAnswer      ExpectedAnswer `json:"expected_answer"`
	EvidenceConceptIDs  []string       `json:"evidence_concept_ids"`
	ExplanationMarkdown string         `json:"explanation_markdown"`
	NumericPolicy       *NumericPolicy `json:"numeric_policy,omitempty"`
}
