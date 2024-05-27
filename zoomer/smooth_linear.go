package zoomer

import "github.com/tinne26/mipix/internal"

var _ Zoomer = (*SmoothLinear)(nil)

// An implementation of the [Zoomer] interface that uses
// linear interpolation for zoom speeds, but with some
// additional smoothing factors. Very handcrafted, can
// "spring" a bit (overshoot the target and rebound).
//
// The implementation is tick-rate independent.
type SmoothLinear struct {
	speed float64
	adjustedTarget float64
}

// Implements [Zoomer].
func (self *SmoothLinear) Reset() {
	self.adjustedTarget = internal.GetCurrentZoom()
	self.speed = 0.0
}

// Implements [Zoomer].
func (self *SmoothLinear) Update(currentZoom, targetZoom float64) float64 {
	const MaxZoomTracking float64 = 5.0

	// The idea behind the maths is the following: using linear interpolation
	// for speeds already results in smooth changes (integrating a linear
	// function gives you a quadratic one). There are some problems with a
	// naive approach, though:
	// - Massive changes can lead to non-smooth deltas. We correct for this
	//   by setting a maximum distance value and clamping.
	// - Sudden target changes, which are not common with movement, but are
	//   common with zoom changes, can look unpleasant. What we do here is
	//   not registering a new target directly, but instead get closer and
	//   closer to it progressively, with self.adjustedTarget. There are
	//   still some edge cases, but we smooth that with an extra speed
	//   interpolation.

	updateDelta := 1.0/float64(internal.GetUPS())
	if targetZoom != self.adjustedTarget {
		distance := targetZoom - self.adjustedTarget
		normDist := internal.Clamp(distance, -MaxZoomTracking, MaxZoomTracking)
		targetApproximation := normDist*1.6*updateDelta
		if internal.Abs(targetZoom - currentZoom) < internal.Abs(distance - targetApproximation) {
			self.adjustedTarget = currentZoom
		} else {	
			self.adjustedTarget += targetApproximation
		}
	}

	newSpeed := internal.LinearInterp(0.0, self.adjustedTarget - currentZoom, 0.15)
	self.speed = internal.LinearInterp(self.speed, newSpeed, 3.0*updateDelta)
	speed := self.speed*updateDelta*20.0
	return speed 
}
