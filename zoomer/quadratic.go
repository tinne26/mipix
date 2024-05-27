package zoomer

import "github.com/tinne26/mipix/internal"

// After writing a few zoomers, I saw I liked quadratic in/out curves
// more than springs. This is some kind of v2 for [SmoothLinear] and
// [RoughLinear], with less magic hardcoded numbers, less bouncing by
// default, but still smooth turns mid-transition.
//
// The zoom can still bounce if the target is suddenly changed, but
// it's more stable in general. Like, it's purely quadratic in/out
// if we don't change targets mid-transition and there are enough
// updates per second for the simulation to be accurate.
type Quadratic struct {
	speed float64
	acceleration float64 // absolute value, always positive, configurable
	maxSpeed float64 // absolute value, always positive, configurable
	initialized bool
}

func (self *Quadratic) ensureInitialized() {
	if self.initialized { return }
	self.initialized = true
	if self.maxSpeed == 0.0 {
		self.maxSpeed = 5.0
	}
	if self.acceleration == 0.0 {
		self.acceleration = 3.66
	}
}

// The default is 3.66. Reasonable values range between [0.3, 16.0].
func (self *Quadratic) SetAcceleration(acceleration float64) {
	if acceleration < 0.01 { panic("acceleration can't be < 0.01") }
	self.acceleration = acceleration
}

// The default is 5.0.
func (self *Quadratic) SetMaxSpeed(maxSpeed float64) {
	if maxSpeed < 0.5 { panic("maxSpeed can't be < 0.5") }
	self.maxSpeed = maxSpeed
}

// Implements [Zoomer].
func (self *Quadratic) Reset() {
	self.speed = 0.0
}

// Implements [Zoomer].
func (self *Quadratic) Update(currentZoom, targetZoom float64) float64 {
	if currentZoom == targetZoom { return 0.0 }
	self.ensureInitialized()

	// compute predicted distance if starting to decelerate now
	// (the calculation is fairly simple: we know that the speed
	// is speed = acceleration*t, then t = speed/acceleration,
	// and the integral of this is acceleration*(speed^2/2.0),
	// which gives us the distance)
	distance := (targetZoom - currentZoom)
	predicted := (self.speed*self.speed)/(2.0*internal.Abs(self.acceleration))
	target := internal.Abs(distance)
	
	// update speed
	updateDelta := 1.0/float64(internal.GetUPS())
	if predicted < target {
		if distance >= 0 {
			self.speed += self.acceleration*updateDelta
			self.speed  = min(self.speed, +self.maxSpeed)
		} else {
			self.speed -= self.acceleration*updateDelta
			self.speed  = max(self.speed, -self.maxSpeed)
		}
	} else {
		if distance >= 0 {
			self.speed -= self.acceleration*updateDelta
			self.speed  = max(self.speed, 0.0)
		} else {
			self.speed += self.acceleration*updateDelta
			self.speed  = min(self.speed, 0.0)
		}
	}

	// compute change and stabilize speed and acceleration if very near the target
	change := self.speed*updateDelta
	if internal.Abs(change) < 0.001 && internal.Abs(distance) < 0.001 {
		self.speed = 0.0
		return distance
	}
	return change
}
