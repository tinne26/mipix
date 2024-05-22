package mipix

var defaultSimpleZoomer *SimpleZoomer
var _ Zoomer = (*SimpleZoomer)(nil)

// A default implementation of the [Zoomer] interface.
type SimpleZoomer struct {
	speed float64
	adjustedTarget float64
}

// Implements [Zoomer].
func (self *SimpleZoomer) Reset() {
	self.adjustedTarget, _ = Camera().GetZoom()
	self.speed = 0.0
}

// Implements [Zoomer].
func (self *SimpleZoomer) Update(currentZoom, targetZoom float64) float64 {
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

	updateDelta := 1.0/float64(Tick().UPS())
	if targetZoom != self.adjustedTarget {
		distance := targetZoom - self.adjustedTarget
		normDist := clamp(distance, -MaxZoomTracking, MaxZoomTracking)
		targetApproximation := normDist*1.6*updateDelta
		if abs(targetZoom - currentZoom) < abs(distance - targetApproximation) {
			self.adjustedTarget = currentZoom
		} else {	
			self.adjustedTarget += targetApproximation
		}
	}

	newSpeed := linearInterp(0.0, self.adjustedTarget - currentZoom, 0.15)
	self.speed = linearInterp(self.speed, newSpeed, 3.0*updateDelta)
	return self.speed*updateDelta*20.0
}
