package mathengine

import (
	"fmt"
	"math"
)

// SampleSizeMeanResult holds the result of sample size planning for a population mean.
type SampleSizeMeanResult struct {
	Sigma           float64 `json:"sigma"`
	MarginOfError   float64 `json:"margin_of_error"`
	ConfidenceLevel float64 `json:"confidence_level"`
	CriticalZ       float64 `json:"critical_z"`
	ExactN          float64 `json:"exact_n"`
	RequiredN       int     `json:"required_n"` // always ceil
	ExcelFormula    string  `json:"excel_formula"`
}

// SampleSizeProportionResult holds the result of sample size planning for a population proportion.
type SampleSizeProportionResult struct {
	P               float64 `json:"p"`
	Conservative    bool    `json:"conservative"`
	MarginOfError   float64 `json:"margin_of_error"`
	ConfidenceLevel float64 `json:"confidence_level"`
	CriticalZ       float64 `json:"critical_z"`
	ExactN          float64 `json:"exact_n"`
	RequiredN       int     `json:"required_n"` // always ceil
	ExcelFormula    string  `json:"excel_formula"`
}

// TargetSEResult holds the result of calculating sample size for a target standard error.
type TargetSEResult struct {
	Sigma        float64 `json:"sigma"`
	TargetSE     float64 `json:"target_se"`
	ExactN       float64 `json:"exact_n"`
	RequiredN    int     `json:"required_n"`
	ExcelFormula string  `json:"excel_formula"`
}

// ComputeSampleSizeMean computes the minimum sample size required to estimate a population mean
// within a specified margin of error E at confidence level (1 - alpha).
// Formula: n = ceil( (z * sigma / E)^2 ).
func ComputeSampleSizeMean(sigma, marginOfError, confLevel float64) (*SampleSizeMeanResult, error) {
	if math.IsNaN(sigma) || math.IsInf(sigma, 0) || sigma <= 0 {
		return nil, fmt.Errorf("sigma must be strictly positive, got %v", sigma)
	}
	if math.IsNaN(marginOfError) || math.IsInf(marginOfError, 0) || marginOfError <= 0 {
		return nil, fmt.Errorf("margin of error must be strictly positive, got %v", marginOfError)
	}
	if math.IsNaN(confLevel) || math.IsInf(confLevel, 0) || confLevel <= 0 || confLevel >= 1 {
		return nil, fmt.Errorf("confidence level must be in open interval (0, 1), got %v", confLevel)
	}

	alpha := 1.0 - confLevel
	norm := StandardNormal()
	zCrit, err := norm.Quantile(1.0 - 0.5*alpha)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate critical z-value: %w", err)
	}

	exactN := math.Pow(zCrit*sigma/marginOfError, 2)
	requiredN := int(math.Ceil(exactN))

	excelFormula := fmt.Sprintf("=ROUNDUP((NORM.S.INV(%.4g) * %.4g / %.4g)^2, 0)", 1.0-0.5*alpha, sigma, marginOfError)

	return &SampleSizeMeanResult{
		Sigma:           sigma,
		MarginOfError:   marginOfError,
		ConfidenceLevel: confLevel,
		CriticalZ:       zCrit,
		ExactN:          exactN,
		RequiredN:       requiredN,
		ExcelFormula:    excelFormula,
	}, nil
}

// ComputeSampleSizeProportion computes the minimum sample size required to estimate a population proportion
// within a specified margin of error E at confidence level (1 - alpha).
// If prior estimate p is not provided (or <= 0), conservative planning uses p = 0.50 (maximizing p(1-p) = 0.25).
// Formula: n = ceil( (z^2 * p * (1-p)) / E^2 ).
func ComputeSampleSizeProportion(p, marginOfError, confLevel float64) (*SampleSizeProportionResult, error) {
	if math.IsNaN(marginOfError) || math.IsInf(marginOfError, 0) || marginOfError <= 0 {
		return nil, fmt.Errorf("margin of error must be strictly positive, got %v", marginOfError)
	}
	if math.IsNaN(confLevel) || math.IsInf(confLevel, 0) || confLevel <= 0 || confLevel >= 1 {
		return nil, fmt.Errorf("confidence level must be in open interval (0, 1), got %v", confLevel)
	}

	conservative := false
	effP := p
	if math.IsNaN(p) || p <= 0 || p >= 1 {
		conservative = true
		effP = 0.50
	}

	alpha := 1.0 - confLevel
	norm := StandardNormal()
	zCrit, err := norm.Quantile(1.0 - 0.5*alpha)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate critical z-value: %w", err)
	}

	exactN := (math.Pow(zCrit, 2) * effP * (1.0 - effP)) / math.Pow(marginOfError, 2)
	requiredN := int(math.Ceil(exactN))

	excelFormula := fmt.Sprintf("=ROUNDUP(NORM.S.INV(%.4g)^2 * %.4g * (1 - %.4g) / %.4g^2, 0)", 1.0-0.5*alpha, effP, effP, marginOfError)

	return &SampleSizeProportionResult{
		P:               effP,
		Conservative:    conservative,
		MarginOfError:   marginOfError,
		ConfidenceLevel: confLevel,
		CriticalZ:       zCrit,
		ExactN:          exactN,
		RequiredN:       requiredN,
		ExcelFormula:    excelFormula,
	}, nil
}

// ComputeSampleSizeTargetSE computes the minimum sample size required for the standard error
// of the sample mean to be at or below targetSE: sigma / sqrt(n) <= targetSE  =>  n >= (sigma / targetSE)^2.
// Always rounds up to the next integer.
func ComputeSampleSizeTargetSE(sigma, targetSE float64) (*TargetSEResult, error) {
	if math.IsNaN(sigma) || math.IsInf(sigma, 0) || sigma <= 0 {
		return nil, fmt.Errorf("sigma must be strictly positive, got %v", sigma)
	}
	if math.IsNaN(targetSE) || math.IsInf(targetSE, 0) || targetSE <= 0 {
		return nil, fmt.Errorf("target standard error must be strictly positive, got %v", targetSE)
	}

	exactN := math.Pow(sigma/targetSE, 2)
	requiredN := int(math.Ceil(exactN))

	excelFormula := fmt.Sprintf("=ROUNDUP((%.4g / %.4g)^2, 0)", sigma, targetSE)

	return &TargetSEResult{
		Sigma:        sigma,
		TargetSE:     targetSE,
		ExactN:       exactN,
		RequiredN:    requiredN,
		ExcelFormula: excelFormula,
	}, nil
}

// StandardErrorMean computes sigma / sqrt(n).
func StandardErrorMean(sigma float64, n int) (float64, error) {
	if math.IsNaN(sigma) || math.IsInf(sigma, 0) || sigma <= 0 {
		return 0, fmt.Errorf("sigma must be strictly positive, got %v", sigma)
	}
	if n <= 0 {
		return 0, fmt.Errorf("sample size n must be positive, got %d", n)
	}
	return sigma / math.Sqrt(float64(n)), nil
}

// StandardErrorProportion computes sqrt(p * (1 - p) / n).
func StandardErrorProportion(p float64, n int) (float64, error) {
	if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 || p >= 1 {
		return 0, fmt.Errorf("proportion p must be in (0, 1), got %v", p)
	}
	if n <= 0 {
		return 0, fmt.Errorf("sample size n must be positive, got %d", n)
	}
	return math.Sqrt(p * (1.0 - p) / float64(n)), nil
}
