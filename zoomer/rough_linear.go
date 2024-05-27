package zoomer

import "github.com/tinne26/mipix/internal"

var _ Zoomer = (*RoughLinear)(nil)

// Somewhat similar to [SmoothLinear], but without any bouncing
// and rougher turns mid-transition.
//
// It also has a customizable speed factor.
//
// The implementation is tick-rate independent.
type RoughLinear struct {
	speedFactorOffset float64
	adjustedTarget float64
}

// Speed factor must be strictly positive. Defaults to 1.0.
func (self *RoughLinear) SetSpeedFactor(factor float64) {
	if factor <= 0 { panic("zoom speed factor must be strictly positive") }
	self.speedFactorOffset = factor - 1.0
}

// Implements [Zoomer].
func (self *RoughLinear) Reset() {
	self.adjustedTarget = internal.GetCurrentZoom()
}

// Implements [Zoomer].
func (self *RoughLinear) Update(currentZoom, targetZoom float64) float64 {
	const MaxZoomTracking float64 = 5.0

	updateDelta := 1.0/float64(internal.GetUPS())
	if targetZoom != self.adjustedTarget {
		var dir float64 = 1.0
		if targetZoom < self.adjustedTarget { dir = -1.0 }

		distance := targetZoom - self.adjustedTarget
		normDist := internal.Clamp(distance, -MaxZoomTracking, MaxZoomTracking)
		targetApproximation := normDist*1.6*updateDelta + dir*updateDelta/2.0
		if internal.Abs(targetZoom - currentZoom) < internal.Abs(distance - targetApproximation) {
			self.adjustedTarget = currentZoom
			return 0.0
		} else {	
			self.adjustedTarget += targetApproximation
			switch dir {
			case +1.0: self.adjustedTarget = min(self.adjustedTarget, targetZoom)
			case -1.0: self.adjustedTarget = max(self.adjustedTarget, targetZoom)
			}
		}
	}

	var a, b, t float64 = 0.0, self.adjustedTarget - currentZoom, 2.6*updateDelta
	change := internal.LinearInterp(a, b, t)*(1.0 + self.speedFactorOffset)
	if change < 0 {
		change -= updateDelta/3.0
		change = max(change, targetZoom - currentZoom)
	} else if change > 0 {
		change += updateDelta/3.0
		change = min(change, targetZoom - currentZoom)
	}
	return change
}
