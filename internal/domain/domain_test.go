package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

func TestValidateID(t *testing.T) {
	validIDs := []string{
		"binomial_pmf",
		"define_variable",
		"n4_p_half",
		"a",
		"step1_first",
	}
	for _, id := range validIDs {
		if !domain.IsValidID(id) {
			t.Errorf("expected %q to be valid ID", id)
		}
		if err := domain.ValidateID("test", id); err != nil {
			t.Errorf("expected %q to pass ValidateID, got %v", id, err)
		}
	}

	invalidIDs := []string{
		"",
		"1first",
		"_leading_underscore",
		"-hyphen",
		"CamelCase",
		"with spaces",
		"special!char",
	}
	for _, id := range invalidIDs {
		if domain.IsValidID(id) {
			t.Errorf("expected %q to be invalid ID", id)
		}
		if err := domain.ValidateID("test", id); err == nil {
			t.Errorf("expected %q to fail ValidateID", id)
		}
	}
}

func TestExpectedAnswerUnmarshal(t *testing.T) {
	t.Run("valid choice", func(t *testing.T) {
		data := []byte(`{"kind": "choice", "option_id": "count_heads"}`)
		var ans domain.ExpectedAnswer
		if err := json.Unmarshal(data, &ans); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if ans.Kind != domain.AnswerKindChoice || ans.OptionID != "count_heads" || ans.Value != nil {
			t.Fatalf("unexpected parsed choice: %+v", ans)
		}
	})

	t.Run("choice with unknown field rejected", func(t *testing.T) {
		data := []byte(`{"kind": "choice", "option_id": "count_heads", "value": 0.5}`)
		var ans domain.ExpectedAnswer
		if err := json.Unmarshal(data, &ans); err == nil {
			t.Fatal("expected error for extra field in choice expected_answer, got nil")
		}
	})

	t.Run("valid numeric", func(t *testing.T) {
		data := []byte(`{"kind": "numeric", "value": 0.375, "units": "probability"}`)
		var ans domain.ExpectedAnswer
		if err := json.Unmarshal(data, &ans); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}
		if ans.Kind != domain.AnswerKindNumeric || ans.Value == nil || *ans.Value != 0.375 || ans.Units != "probability" {
			t.Fatalf("unexpected parsed numeric: %+v", ans)
		}
	})

	t.Run("numeric with unknown field rejected", func(t *testing.T) {
		data := []byte(`{"kind": "numeric", "value": 0.375, "units": "probability", "option_id": "opt1"}`)
		var ans domain.ExpectedAnswer
		if err := json.Unmarshal(data, &ans); err == nil {
			t.Fatal("expected error for extra field in numeric expected_answer, got nil")
		}
	})

	t.Run("numeric with invalid units rejected", func(t *testing.T) {
		data := []byte(`{"kind": "numeric", "value": 0.375, "units": "dollars"}`)
		var ans domain.ExpectedAnswer
		if err := json.Unmarshal(data, &ans); err == nil {
			t.Fatal("expected error for invalid units, got nil")
		}
	})

	t.Run("unknown kind rejected", func(t *testing.T) {
		data := []byte(`{"kind": "unsupported_kind"}`)
		var ans domain.ExpectedAnswer
		if err := json.Unmarshal(data, &ans); err == nil {
			t.Fatal("expected error for unknown kind, got nil")
		}
	})
}
