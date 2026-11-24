package results

import (
	"math"
	"testing"
)

func TestSummarizeReadsOrderForJitterAndSortsForQuantiles(t *testing.T) {
	s := Summarize([]float64{10, 30, 20, 40})
	if s.N != 4 || s.Min != 10 || s.Max != 40 || s.Median != 25 || s.Mean != 25 {
		t.Fatalf("stats = %+v", s)
	}
	// |30-10| + |20-30| + |40-20| over three steps.
	if math.Abs(s.Jitter-50.0/3) > 1e-9 {
		t.Fatalf("jitter = %v", s.Jitter)
	}
	if Summarize(nil) != (Stats{}) {
		t.Fatal("no samples should summarise to nothing")
	}
}

func TestQuantileInterpolates(t *testing.T) {
	sorted := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if got := Quantile(sorted, 0.9); math.Abs(got-9.1) > 1e-9 {
		t.Fatalf("p90 = %v", got)
	}
	if !math.IsNaN(Quantile(nil, 0.5)) {
		t.Fatal("an empty quantile should be NaN")
	}
}
