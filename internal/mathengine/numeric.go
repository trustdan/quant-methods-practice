package mathengine

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/trustdan/quant-methods-practice/internal/domain"
)

// ParsedNumber encapsulates a sanitized, parsed learner numeric response
// preserving both floating-point and exact rational representations where available.
type ParsedNumber struct {
	Raw      string             `json:"raw"`
	Form     domain.NumericForm `json:"form"`
	Value    float64            `json:"value"`
	ExactRat *big.Rat           `json:"exact_rat,omitempty"`
}

// GradeResult captures the deterministic evaluation outcome of a numeric answer.
type GradeResult struct {
	Correct       bool    `json:"correct"`
	SubmittedVal  float64 `json:"submitted_value"`
	ExpectedVal   float64 `json:"expected_value"`
	Diff          float64 `json:"diff"`
	ToleranceUsed float64 `json:"tolerance_used"`
	Reason        string  `json:"reason"`
}

var (
	// Disallow arbitrary LaTeX, expressions, or operators.
	disallowedExpressionPattern = regexp.MustCompile(`[+*^()\\$=a-df-zA-DF-Z]`)
)

func isFormAllowed(form domain.NumericForm, allowed []domain.NumericForm) bool {
	if len(allowed) == 0 {
		return true // Default: all supported forms allowed
	}
	for _, f := range allowed {
		if f == form {
			return true
		}
	}
	return false
}

// ParseNumericInput parses user input into a ParsedNumber according to allowed forms.
// Returns an error separately from wrong answers (validation feedback vs mathematical correctness).
// Explicitly rejects comma decimal separators and unevaluated expressions.
func ParseNumericInput(raw string, allowedForms []domain.NumericForm) (*ParsedNumber, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, errors.New("answer input cannot be empty")
	}

	// Invariant: Ambiguous comma separators are rejected with actionable guidance.
	if strings.Contains(trimmed, ",") {
		return nil, errors.New("ambiguous comma separator; use period '.' as decimal point (e.g. 0.375)")
	}

	// Invariant: Unevaluated expressions and LaTeX formulas are rejected.
	// Allow leading '+' or '-' but disallow binary operators like '1+2', LaTeX macros, or variables.
	testStr := trimmed
	if strings.HasPrefix(testStr, "+") || strings.HasPrefix(testStr, "-") {
		testStr = testStr[1:]
	}
	// Also ignore 'e' or 'E' exponent signs (e.g. 1e-3, 1E+3)
	cleanForCheck := strings.ReplaceAll(testStr, "e+", "")
	cleanForCheck = strings.ReplaceAll(cleanForCheck, "e-", "")
	cleanForCheck = strings.ReplaceAll(cleanForCheck, "E+", "")
	cleanForCheck = strings.ReplaceAll(cleanForCheck, "E-", "")
	cleanForCheck = strings.ReplaceAll(cleanForCheck, "e", "")
	cleanForCheck = strings.ReplaceAll(cleanForCheck, "E", "")
	cleanForCheck = strings.ReplaceAll(cleanForCheck, "%", "")
	cleanForCheck = strings.ReplaceAll(cleanForCheck, "/", "")

	if disallowedExpressionPattern.MatchString(cleanForCheck) {
		return nil, errors.New("expressions and LaTeX formulas are not evaluated; enter a decimal, explicit percent, or simple fraction")
	}

	// 1. Explicit percent: ends with '%'
	if strings.HasSuffix(trimmed, "%") {
		if !isFormAllowed(domain.NumericFormPercent, allowedForms) {
			return nil, errors.New("percentage format is not permitted for this stage")
		}
		numPart := strings.TrimSpace(trimmed[:len(trimmed)-1])
		if numPart == "" {
			return nil, errors.New("missing number before '%'")
		}
		val, err := strconv.ParseFloat(numPart, 64)
		if err != nil || math.IsNaN(val) || math.IsInf(val, 0) {
			return nil, fmt.Errorf("invalid percentage value %q", numPart)
		}

		var exactRat *big.Rat
		if rat, ok := new(big.Rat).SetString(numPart); ok {
			exactRat = new(big.Rat).Mul(rat, big.NewRat(1, 100))
		} else {
			exactRat = new(big.Rat).SetFloat64(val / 100.0)
		}

		return &ParsedNumber{
			Raw:      raw,
			Form:     domain.NumericFormPercent,
			Value:    val / 100.0,
			ExactRat: exactRat,
		}, nil
	}

	// 2. Simple integer fraction: contains '/'
	if strings.Contains(trimmed, "/") {
		if !isFormAllowed(domain.NumericFormFraction, allowedForms) {
			return nil, errors.New("fraction format is not permitted for this stage")
		}
		parts := strings.Split(trimmed, "/")
		if len(parts) != 2 {
			return nil, errors.New("invalid fraction format; expected numerator/denominator (e.g. 3/8)")
		}
		numStr := strings.TrimSpace(parts[0])
		denStr := strings.TrimSpace(parts[1])

		num := new(big.Int)
		if _, ok := num.SetString(numStr, 10); !ok {
			return nil, fmt.Errorf("fraction numerator %q must be an integer", numStr)
		}
		den := new(big.Int)
		if _, ok := den.SetString(denStr, 10); !ok {
			return nil, fmt.Errorf("fraction denominator %q must be an integer", denStr)
		}
		if den.Sign() == 0 {
			return nil, errors.New("fraction denominator cannot be zero")
		}

		rat := new(big.Rat).SetFrac(num, den)
		val, _ := rat.Float64()

		return &ParsedNumber{
			Raw:      raw,
			Form:     domain.NumericFormFraction,
			Value:    val,
			ExactRat: rat,
		}, nil
	}

	// 3. Decimal number
	if !isFormAllowed(domain.NumericFormDecimal, allowedForms) {
		return nil, errors.New("decimal format is not permitted for this stage")
	}
	val, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || math.IsNaN(val) || math.IsInf(val, 0) {
		return nil, fmt.Errorf("invalid decimal number %q", trimmed)
	}

	var exactRat *big.Rat
	if rat, ok := new(big.Rat).SetString(trimmed); ok {
		exactRat = rat
	} else {
		exactRat = new(big.Rat).SetFloat64(val)
	}

	return &ParsedNumber{
		Raw:      raw,
		Form:     domain.NumericFormDecimal,
		Value:    val,
		ExactRat: exactRat,
	}, nil
}

// GradeNumericAnswer evaluates a ParsedNumber against an expected numeric answer
// using the per-stage NumericPolicy (absolute/relative tolerance and stated rounding intervals).
// Invariant: Never compare formatted strings as numeric truth.
func GradeNumericAnswer(submitted *ParsedNumber, expected float64, policy domain.NumericPolicy) GradeResult {
	if submitted == nil {
		return GradeResult{
			Correct: false,
			Reason:  "no submitted answer provided",
		}
	}

	diff := math.Abs(submitted.Value - expected)
	allowedTol := math.Max(policy.AbsoluteTolerance, policy.RelativeTolerance*math.Abs(expected))

	// 1. Check direct tolerance inequality: |x - y| <= max(absTol, relTol * |y|)
	if diff <= allowedTol {
		return GradeResult{
			Correct:       true,
			SubmittedVal:  submitted.Value,
			ExpectedVal:   expected,
			Diff:          diff,
			ToleranceUsed: allowedTol,
			Reason:        "within acceptable tolerance",
		}
	}

	// 2. Check stated rounding policy if display_decimals is configured
	if policy.DisplayDecimals != nil {
		d := *policy.DisplayDecimals
		scale := math.Pow10(d)

		// Round half-to-even (banker's rounding) and round half-away-from-zero
		roundedExpEven := math.RoundToEven(expected*scale) / scale
		roundedExpAway := math.Round(expected*scale) / scale

		if math.Abs(submitted.Value-roundedExpEven) <= policy.AbsoluteTolerance ||
			math.Abs(submitted.Value-roundedExpAway) <= policy.AbsoluteTolerance {
			return GradeResult{
				Correct:       true,
				SubmittedVal:  submitted.Value,
				ExpectedVal:   expected,
				Diff:          diff,
				ToleranceUsed: allowedTol,
				Reason:        fmt.Sprintf("matches expected answer rounded to %d decimal places", d),
			}
		}

		// Check half-ulp interval implied by stated rounding [expected - 0.5*10^-d, expected + 0.5*10^-d]
		halfUlp := 0.5 / scale
		if submitted.Value >= (expected-halfUlp-policy.AbsoluteTolerance) &&
			submitted.Value <= (expected+halfUlp+policy.AbsoluteTolerance) {
			return GradeResult{
				Correct:       true,
				SubmittedVal:  submitted.Value,
				ExpectedVal:   expected,
				Diff:          diff,
				ToleranceUsed: halfUlp,
				Reason:        fmt.Sprintf("within rounding interval implied by stated %d decimal places", d),
			}
		}
	}

	// Invariant: A bare 37.5 is not silently interpreted as 37.5%
	// Generate diagnostic reasoning to guide the learner without altering numeric truth.
	reason := fmt.Sprintf("value %.6g differs from expected %.6g by %.6g (allowed tolerance: %.6g)",
		submitted.Value, expected, diff, allowedTol)

	if expected <= 1.0 && math.Abs((submitted.Value/100.0)-expected) <= allowedTol {
		reason += fmt.Sprintf("; hint: entered value %.6g appears to be a bare percentage; an explicit '%%' sign is required (e.g. %.6g%%)",
			submitted.Value, submitted.Value)
	}

	return GradeResult{
		Correct:       false,
		SubmittedVal:  submitted.Value,
		ExpectedVal:   expected,
		Diff:          diff,
		ToleranceUsed: allowedTol,
		Reason:        reason,
	}
}
