package mathengine_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/trustdan/quant-methods-practice/internal/mathengine"
)

func TestEventCanonicalization(t *testing.T) {
	lt := mathengine.LessThan(2).Canonicalize()
	if lt.Relation != mathengine.RelationAtMost || lt.K != 1 {
		t.Errorf("LessThan(2).Canonicalize() = %+v, want AtMost(1)", lt)
	}

	gt := mathengine.GreaterThan(2).Canonicalize()
	if gt.Relation != mathengine.RelationAtLeast || gt.K != 3 {
		t.Errorf("GreaterThan(2).Canonicalize() = %+v, want AtLeast(3)", gt)
	}

	eq := mathengine.Exactly(2).Canonicalize()
	if eq.Relation != mathengine.RelationEqual || eq.K != 2 {
		t.Errorf("Exactly(2).Canonicalize() = %+v, want Exactly(2)", eq)
	}
}

func TestEventLaTeXAndDescription(t *testing.T) {
	tests := []struct {
		event     mathengine.DiscreteEvent
		wantLaTeX string
		wantCanon string
		wantDesc  string
	}{
		{mathengine.Exactly(2), "X = 2", "X = 2", "exactly 2"},
		{mathengine.AtMost(2), "X \\le 2", "X \\le 2", "at most 2"},
		{mathengine.LessThan(2), "X < 2", "X \\le 1", "fewer than 2"},
		{mathengine.AtLeast(2), "X \\ge 2", "X \\ge 2", "at least 2"},
		{mathengine.GreaterThan(2), "X > 2", "X \\ge 3", "more than 2"},
		{mathengine.BetweenInclusive(1, 3), "1 \\le X \\le 3", "1 \\le X \\le 3", "between 1 and 3 inclusive"},
	}

	for _, tt := range tests {
		if got := tt.event.LaTeX("X"); got != tt.wantLaTeX {
			t.Errorf("event.LaTeX(X) = %q, want %q", got, tt.wantLaTeX)
		}
		if got := tt.event.CanonicalLaTeX("X"); got != tt.wantCanon {
			t.Errorf("event.CanonicalLaTeX(X) = %q, want %q", got, tt.wantCanon)
		}
		if got := tt.event.Description(); got != tt.wantDesc {
			t.Errorf("event.Description() = %q, want %q", got, tt.wantDesc)
		}
	}
}

func TestParseEventRelation(t *testing.T) {
	cases := []struct {
		input string
		want  mathengine.EventRelation
	}{
		{"equal", mathengine.RelationEqual},
		{"exactly", mathengine.RelationEqual},
		{"==", mathengine.RelationEqual},
		{"=", mathengine.RelationEqual},
		{"at_most", mathengine.RelationAtMost},
		{"at most", mathengine.RelationAtMost},
		{"<=", mathengine.RelationAtMost},
		{"less_than", mathengine.RelationLessThan},
		{"fewer than", mathengine.RelationLessThan},
		{"<", mathengine.RelationLessThan},
		{"at_least", mathengine.RelationAtLeast},
		{"at least", mathengine.RelationAtLeast},
		{">=", mathengine.RelationAtLeast},
		{"greater_than", mathengine.RelationGreaterThan},
		{"more than", mathengine.RelationGreaterThan},
		{">", mathengine.RelationGreaterThan},
		{"between", mathengine.RelationBetweenInclusive},
	}

	for _, c := range cases {
		rel, err := mathengine.ParseEventRelation(c.input)
		if err != nil {
			t.Errorf("ParseEventRelation(%q) returned unexpected error: %v", c.input, err)
		}
		if rel != c.want {
			t.Errorf("ParseEventRelation(%q) = %q, want %q", c.input, rel, c.want)
		}
	}

	if _, err := mathengine.ParseEventRelation("unknown_rel"); err == nil {
		t.Errorf("ParseEventRelation(unknown_rel) expected error, got nil")
	}
}

func TestEvaluateBinomialEvents(t *testing.T) {
	// n=4, p=0.5:
	// P(X=0) = 1/16 = 0.0625
	// P(X=1) = 4/16 = 0.25
	// P(X=2) = 6/16 = 0.375
	// P(X=3) = 4/16 = 0.25
	// P(X=4) = 1/16 = 0.0625
	dist, err := mathengine.NewBinomial(4, 0.5)
	if err != nil {
		t.Fatalf("NewBinomial failed: %v", err)
	}

	tests := []struct {
		name    string
		event   mathengine.DiscreteEvent
		wantVal float64
		wantRat string
	}{
		{"exactly 2", mathengine.Exactly(2), 0.375, "3/8"},
		{"at most 2", mathengine.AtMost(2), 0.6875, "11/16"},
		{"less than 2", mathengine.LessThan(2), 0.3125, "5/16"},
		{"at least 2", mathengine.AtLeast(2), 0.6875, "11/16"},
		{"greater than 2", mathengine.GreaterThan(2), 0.3125, "5/16"},
		{"between 1 and 3", mathengine.BetweenInclusive(1, 3), 0.875, "7/8"},
	}

	pRat := big.NewRat(1, 2)
	for _, tt := range tests {
		got := tt.event.EvaluateBinomial(dist)
		if math.Abs(got-tt.wantVal) > 1e-12 {
			t.Errorf("%s EvaluateBinomial = %v, want %v", tt.name, got, tt.wantVal)
		}

		rat, err := tt.event.EvaluateBinomialRat(dist, pRat)
		if err != nil {
			t.Fatalf("%s EvaluateBinomialRat error: %v", tt.name, err)
		}
		expectedRat, _ := new(big.Rat).SetString(tt.wantRat)
		if rat.Cmp(expectedRat) != 0 {
			t.Errorf("%s EvaluateBinomialRat = %s, want %s", tt.name, rat.RatString(), tt.wantRat)
		}
	}
}
