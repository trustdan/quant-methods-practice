package mathengine

import "fmt"

// EmpiricalFrequency describes recorded experiments, without estimating a model's
// independence or constant-p conditions from the observations.
func EmpiricalFrequency(successes []int, trials, target int) (matching int, frequency float64, err error) {
	if trials < 1 || target < 0 || target > trials || len(successes) == 0 {
		return 0, 0, fmt.Errorf("need experiments and a target within the trial support")
	}
	for _, count := range successes {
		if count < 0 || count > trials {
			return 0, 0, fmt.Errorf("success count outside [0,%d]", trials)
		}
		if count == target {
			matching++
		}
	}
	return matching, float64(matching) / float64(len(successes)), nil
}
