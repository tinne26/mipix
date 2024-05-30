package tracker

import "github.com/tinne26/mipix/internal"

// Note: I could have a generic Tailer[T], but anyone who wants
// to write their own stuff can figure it out.

// Like [Tailer], but based on a [Spring] tracker, and the catch
// up algorithm also uses a spring instead of quadratic easings.
//
// Example settings configuration for a gentle tracker:
//   springTailer := tracker.SpringTailer{}
//   springTailer.Spring.SetParameters(0.8, 2.4)
//   springTailer.SetCatchUpParameters(0.9, 1.75)
type SpringTailer struct {
	Spring Spring
	follower follower
	corrector springCorrector
	// note: with a quadratic corrector, the following
	// settings are nice: spring {0.666, 2.0}, accel. {0.15}
}

// Sets the spring parameters for the catch up corrector.
// The default values are (0.9, 1.75).
func (self *SpringTailer) SetCatchUpParameters(damping, power float64) {
	self.corrector.SetParameters(damping, power)
}

// See [Tailer.SetCatchUpTimes](). The default values are (1.0, 0.5).
func (self *SpringTailer) SetCatchUpTimes(engage, disengage float64) {
	if engage < disengage { panic("engage time must be >= disengage") }
	self.follower.SetTimes(engage, disengage)
}

// Implements [Tracker].
func (self *SpringTailer) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	// pre-subtract correction
	w, h := internal.GetResolution()
	w64, h64 := float64(w), float64(h)
	zoom := internal.GetCurrentZoom()
	updateDelta := 1.0/(float64(internal.GetUPS()))
	relCurrentX := currentX - (self.corrector.speedX*w64*updateDelta)/zoom
	relCurrentY := currentY - (self.corrector.speedY*h64*updateDelta)/zoom

	// basic parametrized update
	changeX, changeY := self.Spring.Update(relCurrentX, relCurrentY, targetX, targetY, prevSpeedX, prevSpeedY)
	
	// follower correction
	self.follower.Update(changeX, changeY, prevSpeedX, prevSpeedY)
	if self.follower.IsEngaged() {
		self.corrector.Update(targetX - currentX, targetY - currentY)
	} else { // deceleration case
		self.corrector.Update(0.0, 0.0)
	}
	correctorChangeX := (self.corrector.speedX*w64*updateDelta)/zoom
	correctorChangeY := (self.corrector.speedY*h64*updateDelta)/zoom
	changeX += correctorChangeX
	changeY += correctorChangeY

	// stabilization pass
	if self.corrector.speedX == 0.0 && self.corrector.speedY == 0.0 &&
	   internal.Abs(changeX) < 0.12*updateDelta && internal.Abs(changeY) < 0.12*updateDelta &&
		internal.Abs(targetX - (currentX + changeX)) < (0.25/zoom)*updateDelta &&
		internal.Abs(targetY - (currentY + changeY)) < (0.25/zoom)*updateDelta {
		return targetX - currentX, targetY - currentY
	}

	return changeX, changeY
}

