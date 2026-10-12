package mathengine

import (
	"math"
	"testing"
)

func TestInclusiveQuantileFixtures(t *testing.T) {
	// [1,3,7,9]: h=.75 at p=.25, giving 1+.75*(3-1)=2.5.
	data := []float64{9, 1, 7, 3}
	for _, tc := range []struct{ p, want float64 }{{0, 1}, {.25, 2.5}, {.5, 5}, {.75, 7.5}, {1, 9}} {
		got, err := InclusiveQuantile(data, tc.p)
		if err != nil || math.Abs(got-tc.want) > 1e-12 {
			t.Fatalf("p=%g got %g, %v; want %g", tc.p, got, err, tc.want)
		}
	}
	if data[0] != 9 {
		t.Fatal("quantile mutated input")
	}
	for _, data := range [][]float64{nil, {1, math.NaN()}, {math.Inf(1)}} {
		if _, err := InclusiveQuantile(data, .25); err == nil {
			t.Fatal("invalid data accepted")
		}
	}
	if _, err := InclusiveQuantile([]float64{1}, 1.1); err == nil {
		t.Fatal("invalid percentile accepted")
	}
}

func TestCurriculumIndependentFixtures(t *testing.T) {
	cases := []struct {
		family string
		p      map[string]interface{}
		want   map[string]float64
	}{
		{"descriptive_stats", map[string]interface{}{"operation": "z_score", "x": 140, "mean": 100, "sd": 20}, map[string]float64{"z": 2}},
		{"descriptive_stats", map[string]interface{}{"operation": "weighted_mean", "data": []float64{10, 20, 40}, "weights": []float64{100, 50, 50}}, map[string]float64{"mean": 20}},
		{"descriptive_stats", map[string]interface{}{"operation": "quartiles_inc", "data": []float64{2, 3, 4, 5, 6, 7, 8, 9, 30}}, map[string]float64{"q1": 4, "q3": 8, "iqr": 4, "lower_fence": -2, "upper_fence": 14}},
		{"discrete_rv", map[string]interface{}{"operation": "compare_risk", "values": []float64{0, 200}, "probabilities": []float64{.5, .5}, "values_b": []float64{80, 120}, "probabilities_b": []float64{.5, .5}}, map[string]float64{"mean": 100, "mean_b": 100, "sd": 100, "sd_b": 20}},
		{"discrete_rv", map[string]interface{}{"operation": "pmf", "values": []float64{0, 1000, 10000}, "probabilities": []float64{.97, .025, .005}}, map[string]float64{"mean": 75, "variance": 519375}},
		{"discrete_rv", map[string]interface{}{"operation": "pmf", "values": []float64{0, 1, 2, 3, 4}, "probabilities": []float64{.1, .2, .3, .25, .15}, "k": 2, "lower": 1, "upper": 3}, map[string]float64{"geq": .7, "between": .75}},
		{"linear_combination", map[string]interface{}{"operation": "iid_sum_vs_scale", "n": 5, "mean_x": 20, "var_x": 9}, map[string]float64{"mean_sum": 100, "var_sum": 45, "var_scaled": 225, "sd_scaled": 15}},
		{"linear_combination", map[string]interface{}{"operation": "portfolio", "a": .6, "b": .4, "mean_x": 8, "mean_y": 4, "var_x": 100, "var_y": 25, "rho": .3}, map[string]float64{"mean": 6.4, "covariance": 15, "variance": 47.2}},
		{"two_sample_test", map[string]interface{}{"operation": "paired", "before": []float64{40, 42, 45, 39, 50, 48, 43, 46}, "after": []float64{42, 46, 48, 44, 51, 52, 45, 49}, "alpha": .05}, map[string]float64{"mean_diff": 3, "sd_diff": math.Sqrt(12.0 / 7), "se": math.Sqrt(3.0 / 14), "t": math.Sqrt(42), "df": 7}},
		{"two_sample_test", map[string]interface{}{"operation": "pooled", "n1": 20, "n2": 20, "mean1": 52, "mean2": 48, "var1": 36, "var2": 64, "alpha": .05}, map[string]float64{"pooled_variance": 50, "se": math.Sqrt(5), "t": 4 / math.Sqrt(5), "df": 38}},
		{"two_sample_test", map[string]interface{}{"operation": "proportions", "n1": 600, "n2": 600, "x1": 48, "x2": 72, "alpha": .05}, map[string]float64{"pooled_p": .1, "se": math.Sqrt(.0003), "z": -4 / math.Sqrt(3)}},
	}
	for _, tc := range cases {
		t.Run(tc.family+"/"+tc.p["operation"].(string), func(t *testing.T) {
			got, err := CurriculumCalculations(tc.family, tc.p)
			if err != nil {
				t.Fatal(err)
			}
			for key, want := range tc.want {
				if math.Abs(got[key]-want) > 1e-9*math.Max(1, math.Abs(want)) {
					t.Errorf("%s=%g want %g", key, got[key], want)
				}
			}
		})
	}
}

func TestCurriculumDirectionalTails(t *testing.T) {
	// df=1 is Cauchy: at t=-1, CDF=.25 and right survival=.75 exactly.
	p := map[string]interface{}{"operation": "paired", "before": []float64{0, 0}, "after": []float64{-2, 0}, "alpha": .05}
	r, err := CurriculumCalculations("two_sample_test", p)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]float64{"t": -1, "p_left": .25, "p_right": .75, "p_two": .5} {
		if math.Abs(r[key]-want) > 1e-10 {
			t.Errorf("%s=%g want %g", key, r[key], want)
		}
	}
	// The opposite directional alternative must not use half the two-tail p.
	if r["p_right"] <= .5 {
		t.Fatal("wrong-direction alternative was reported significant")
	}
}

func TestCurriculumRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		family string
		p      map[string]interface{}
	}{
		{"descriptive_stats", map[string]interface{}{"operation": "made_up"}},
		{"descriptive_stats", map[string]interface{}{"operation": "z_score", "x": 2, "mean": 0, "sd": 0}},
		{"descriptive_stats", map[string]interface{}{"operation": "z_score", "x": 2, "mean": 0, "sd": 1, "answer": 2}},
		{"descriptive_stats", map[string]interface{}{"operation": "z_score", "sd": 1}},
		{"descriptive_stats", map[string]interface{}{"operation": "weighted_mean", "data": []float64{1, 2}, "weights": []float64{1, -1}}},
		{"discrete_rv", map[string]interface{}{"operation": "pmf", "values": []float64{0, 1}, "probabilities": []float64{.4, .4}}},
		{"discrete_rv", map[string]interface{}{"operation": "pmf", "values": []float64{0, 0}, "probabilities": []float64{.5, .5}}},
		{"discrete_rv", map[string]interface{}{"operation": "pmf", "values": []float64{0, 1}, "probabilities": []float64{-.000000001, 1.000000001}}},
		{"linear_combination", map[string]interface{}{"operation": "iid_sum_vs_scale", "n": 1.5, "mean_x": 2, "var_x": 1}},
		{"two_sample_test", map[string]interface{}{"operation": "paired", "before": []float64{1, 2}, "after": []float64{2}, "alpha": .05}},
		{"two_sample_test", map[string]interface{}{"operation": "proportions", "n1": 20, "n2": 20, "x1": 1, "x2": 2, "alpha": .05}},
		{"descriptive_stats", map[string]interface{}{"operation": "z_score", "x": math.Inf(1), "mean": 0, "sd": 1}},
	}
	for i, tc := range cases {
		if _, err := CurriculumCalculations(tc.family, tc.p); err == nil {
			t.Errorf("case %d accepted invalid input", i)
		}
	}
}
