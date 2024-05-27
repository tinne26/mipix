package zoomer

import "github.com/tinne26/mipix/internal"

var _ Zoomer = (*Spring)(nil)

// Springy zoom. By default, it barely overshoots, but
// you can set it to be more or less bouncy if you want.
//
// The implementation is tick-rate independent.
type Spring struct {
	spring internal.Spring
	speed float64
	maxTargetDistance float64
	initialized bool
	zoomCompensation float64
}

func (self *Spring) ensureInitialized() {
	if self.initialized { return }
	self.spring.SetParameters(0.85, 2.5)
	self.initialized = true
}

// Damping values must be in [0.0, 1.0] range.
// Power depends on damping, but must be strictly positive.
// Defaults are (0.85, 2.5).
func (self *Spring) SetParameters(damping, power float64) {
	if damping < 0.0 || damping > 1.0 {
		panic("damping must be in [0, 1] range")
	}
	if power <= 0.0 {
		panic("power must be strictly positive")
	}
	self.spring.SetParameters(damping, power)
	self.initialized = true
}

// See [Constant.SetZoomCompensated]() for context. Compensating
// zooms with the spring zoomer will lead to overshoot on zoom ins
// and undershoot on zoom outs for normal to high power values.
//
// Parameters will also have to be adjusted, as the results get
// very different. I recommend starting at (0.87, 1.6) if you
// set zoom compensation to 1. I also like (0.8, 1.5) at 0.6.
//
// The compensation is also a bit more sophisticated than on
// [Constant], so if you really expect specific results, just
// dive directly into the code.
// 
// Defaults to 0.
func (self *Spring) SetZoomCompensation(compensation float64) {
	if compensation < 0 || compensation > 1.0 {
		panic("zoom compensation must be in [0, 1] range")
	}
	self.zoomCompensation = compensation
}

// Can help tame maximum speeds if desired. Setting it to 0.0
// disables the maximum target distance.
func (self *Spring) SetMaxTargetDistance(maxDistance float64) {
	if maxDistance < 0.0 { panic("max target distance must be >= 0.0") }
	self.maxTargetDistance = maxDistance
}

// Implements [Zoomer].
func (self *Spring) Reset() {
	self.speed = 0.0
}

// Implements [Zoomer].
func (self *Spring) Update(currentZoom, targetZoom float64) float64 {
	if currentZoom == targetZoom && self.speed == 0.0 { return 0.0 }
	
	self.ensureInitialized()
	targetZoom = self.limitTargetDistance(currentZoom, targetZoom)
	newPosition, newSpeed := self.spring.Update(currentZoom, targetZoom, self.speed)
	
	// clean up case, don't keep oscillating on super small
	// changes, it interferes with efficient GPU usage
	if internal.Abs(targetZoom - newPosition) < 0.001 && internal.Abs(newSpeed) < (1.0/float64(internal.GetUPS())) {
		self.speed = 0.0
		return targetZoom - currentZoom
	}
	
	self.speed = newSpeed
	change := (newPosition - currentZoom)

	// zoom compensation is not done directly, but in a smoothed way.
	// as we get close to the target, the compensation is also relaxed
	if self.zoomCompensation > 0 {
		const Threshold = 0.333
		dist := internal.Abs(targetZoom - currentZoom)
		currentZoom = 1.0 + (currentZoom - 1.0)*self.zoomCompensation // *
		// * I personally like this softening to not make the spring
		//   so lifeless, but this could totally be customized.
		if dist <= Threshold {
			t := internal.EaseInQuad(dist*(1.0/Threshold))
			change *= internal.LinearInterp(1.0, currentZoom, t)
		} else {
			change *= currentZoom
		}
	}

	return change
}

func (self *Spring) limitTargetDistance(currentZoom, targetZoom float64) float64 {
	if self.maxTargetDistance <= 0 { return targetZoom }
	distance := (targetZoom - currentZoom)
	if internal.Abs(distance) <= self.maxTargetDistance { return targetZoom }
	switch {
	case distance > 0: targetZoom = currentZoom + self.maxTargetDistance
	case distance < 0: targetZoom = currentZoom - self.maxTargetDistance
	}
	return targetZoom
}
