package results

import (
	"math"
	"slices"
)

// Summarize describes samples in the order they were taken. Jitter is the mean absolute
// difference between consecutive samples, so it measures how unsteady the path was.
func Summarize(samples []float64) Stats {
	if len(samples) == 0 {
		return Stats{}
	}
	var jitter, sum float64
	prev := samples[0]
	for _, v := range samples {
		sum += v
		jitter += math.Abs(v - prev)
		prev = v
	}
	sorted := slices.Clone(samples)
	slices.Sort(sorted)
	out := Stats{
		N:      len(sorted),
		Min:    sorted[0],
		Median: Quantile(sorted, 0.5),
		P90:    Quantile(sorted, 0.9),
		P99:    Quantile(sorted, 0.99),
		Max:    sorted[len(sorted)-1],
		Mean:   sum / float64(len(sorted)),
	}
	if len(samples) > 1 {
		out.Jitter = jitter / float64(len(samples)-1)
	}
	return out
}

// Quantile reads q, from 0 to 1, from sorted samples by linear interpolation.
func Quantile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	pos := q * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return sorted[lo] + (sorted[hi]-sorted[lo])*(pos-float64(lo))
}
