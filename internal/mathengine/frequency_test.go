package mathengine

import "testing"

func TestEmpiricalFrequency(t *testing.T) {
	for _, tc := range []struct {
		counts   []int
		matching int
		value    float64
	}{{[]int{2, 1, 4, 2, 0}, 2, .4}, {[]int{2, 2, 2}, 3, 1}, {[]int{0, 1, 3, 4}, 0, 0}, {[]int{1, 2, 3}, 1, 1.0 / 3}} {
		n, p, err := EmpiricalFrequency(tc.counts, 4, 2)
		if err != nil || n != tc.matching || p != tc.value {
			t.Fatalf("%v: %d %g %v", tc.counts, n, p, err)
		}
	}
	for _, counts := range [][]int{nil, {-1}, {5}} {
		if _, _, err := EmpiricalFrequency(counts, 4, 2); err == nil {
			t.Fatalf("accepted %v", counts)
		}
	}
	if _, _, err := EmpiricalFrequency([]int{2}, 4, 5); err == nil {
		t.Fatal("accepted impossible event")
	}
}
