package tracker

// Update(...) always returns (target - current).
var Instant Tracker = instantTracker{}

type instantTracker struct{}
func (instantTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	return targetX - currentX, targetY - currentY
}
