package tracker

import "github.com/tinne26/mipix/internal"

// A tracker that uses a [Parametrized] implementation as its base,
// which you can access and configure directly as a struct field,
// and then adds a catch up mechanism that triggers after you move
// for some time in a more or less consistent speed.
type Tailer struct {
	Parametrized Parametrized
	follower follower
	corrector corrector
}

// Once the catching up mechanism is triggered, it uses a static
// acceleration to progressively change speed. Reasonable values
// range between [0.1, 0.5]. The default is 0.2.
func (self *Tailer) SetCatchUpAcceleration(acceleration float64) {
	self.corrector.SetAcceleration(acceleration)
}

// Specifies how long it takes the catch up mechanism to engage
// and disengage. The values are given in seconds:
//  - Engaging happens when we have been moving the target
//    at a more or less consistent speed for a while.
//  - Disengaging happens when we have been stopped for a while.
// The default values are 1.0, 0.5. The disengage time must
// always be <= than the engage time.
func (self *Tailer) SetCatchUpTimes(engage, disengage float64) {
	if engage < disengage { panic("engage time must be >= disengage") }
	self.follower.SetTimes(engage, disengage)
}

// Implements [Tracker].
func (self *Tailer) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	// pre-subtract correction
	w, h := internal.GetResolution()
	w64, h64 := float64(w), float64(h)
	zoom := internal.GetCurrentZoom()
	updateDelta := 1.0/(float64(internal.GetUPS()))
	relCurrentX := currentX - (self.corrector.speedX*w64*updateDelta)/zoom
	relCurrentY := currentY - (self.corrector.speedY*h64*updateDelta)/zoom

	// basic parametrized update
	changeX, changeY := self.Parametrized.Update(relCurrentX, relCurrentY, targetX, targetY, prevSpeedX, prevSpeedY)
	
	// follower correction
	self.follower.Update(changeX, changeY, prevSpeedX, prevSpeedY)
	if self.follower.IsEngaged() {
		self.corrector.Update(targetX - currentX, targetY - currentY)
	} else { // deceleration case
		self.corrector.Update(0.0, 0.0) // *
		// * here we could just self.corrector.Decelerate() too,
		//   but I kinda prefer the results with Update(0.0, 0.0)
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

