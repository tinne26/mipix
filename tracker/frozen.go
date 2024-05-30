package tracker

// Update(...) always returns (0, 0).
var Frozen Tracker = frozenTracker{}

type frozenTracker struct {}
func (frozenTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	return 0, 0
}
