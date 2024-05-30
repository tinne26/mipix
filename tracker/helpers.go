package tracker

import "github.com/tinne26/mipix/internal"

// the parameters are too complicated to understand for the average lib user,
// so this is like a subpar version of Parametrized, which was made clearer
func computeLinComponent(current, target, minAdvance, maxAdvance, refMaxDist float64) float64 {
	// determine base speed
	if target > current { // going right
		dist := min(target - current, refMaxDist)
		t := internal.TAt(dist, 0, refMaxDist)
		advance := internal.LinearInterp(0, maxAdvance, t)
		if advance >= minAdvance { return advance }
		return min(minAdvance, dist)
	} else { // going left
		dist := min(current - target, refMaxDist)
		t := internal.TAt(dist, 0, refMaxDist)
		advance := internal.LinearInterp(0, maxAdvance, t)
		if advance >= minAdvance { return -advance }
		return -min(minAdvance, dist)
	}
}

func sim(predictedChange, actualChange float64, maxErrorForZeroSimilarity float64) float64 {
	predictionError := internal.Abs(actualChange - predictedChange)
	if predictionError > maxErrorForZeroSimilarity { return 0.0 }
	return 1.0 - predictionError/maxErrorForZeroSimilarity
}
