package tracker

import "github.com/tinne26/mipix/internal"

type Spring struct {
	spring internal.Spring
	speedX, speedY float64
	initialized bool
}

func (self *Spring) initialize() {
	self.initialized = true
	if !self.spring.IsInitialized() {
		self.spring.SetParameters(0.55, 4.5)
	}
}

func (self *Spring) SetParameters(damping, power float64) {
	self.spring.SetParameters(damping, power)
	self.initialized = true
}

func (self *Spring) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	// initialization
	if !self.initialized { self.initialize() }
	
	// stabilization case
	if internal.Abs(targetX - currentX) < 0.001 && internal.Abs(targetY - currentY) < 0.001 {
		self.speedX, self.speedY = 0.0, 0.0
		return targetX - currentX, targetY - currentY
	}
	
	// get resolution
	w, h := internal.GetResolution()
	widthF64, heightF64 := float64(w), float64(h)
	
	// advance with spring
	var newX, newY float64
	newX, self.speedX = self.spring.Update(currentX/widthF64, targetX/widthF64, self.speedX)
	newY, self.speedY = self.spring.Update(currentY/heightF64, targetY/heightF64, self.speedY)
	newX *= widthF64
	newY *= heightF64

	// normalize change by zoom level
	zoom := internal.GetCurrentZoom()
	return (newX - currentX)*zoom, (newY - currentY)*zoom
}
