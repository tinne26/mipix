package tracker

import "github.com/tinne26/mipix/internal"

type springCorrector struct {
	spring internal.Spring
	speedX, speedY float64 // logical, relative to resolution and zoom
	initialized bool
}

func (self *springCorrector) initialize() {
	self.initialized = true
	if !self.spring.IsInitialized() {
		self.spring.SetParameters(0.9, 1.75)
	}
}

func (self *springCorrector) SetParameters(damping, power float64) {
	self.spring.SetParameters(damping, power)
}

func (self *springCorrector) Update(errorX, errorY float64) {
	if !self.initialized { self.initialize() }

	updateDelta := 1.0/float64(internal.GetUPS())
	w, h := internal.GetResolution()
	w64, h64 := float64(w), float64(h)
	errorX /= w64
	errorY /= h64

	_, self.speedX = self.spring.Update(0.0, errorX, self.speedX)
	_, self.speedY = self.spring.Update(0.0, errorY, self.speedY)
	if internal.Abs(self.speedX) < 0.12*updateDelta {
		self.speedX = 0.0
	}
	if internal.Abs(self.speedY) < 0.12*updateDelta {
		self.speedY = 0.0
	}
}
