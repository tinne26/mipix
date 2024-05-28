package shaker

import "math/rand/v2"

import "github.com/tinne26/mipix/internal"

// idea: just place points anywhere and change them every N ticks. just that.
// and SetParameters(damping, power float64)
// this one can also have individual motion ranges for x and y, that will be
// better

var _ Shaker = (*Spring)(nil)

// A [Shaker] implementation based on spring simulations.
// It's nothing special, but it has its own flavor. Common
// configurations remind me of boxes falling from the closet,
// or driving through a bad road (e.g., params {0.1, 40.0}
// and motion ranges {0.02, 0.01}).
//
// The implementation is tick-rate independent.
type Spring struct {
	spring internal.Spring
	x, y float64
	xSpeed, ySpeed float64
	xTarget, yTarget float64

	xRatio, yRatio float64
	zoomCompensation float64
	initialized bool
}

func (self *Spring) ensureInitialized() {
	if self.initialized { return }
	self.initialized = true
	if self.xRatio == 0.0 { self.xRatio = 0.02 }
	if self.yRatio == 0.0 { self.yRatio = 0.01 }
	if !self.spring.IsInitialized() {
		self.spring.SetParameters(0.1, 40.0)//0.25, 80.0)
	}
}

// Same idea as [Random.SetMaxMotionRange](), but with two separate
// factors for each axis. Defaults to 0.02 for both axes, but it
// often will need tweaking if you also adjust the spring parameters.
func (self *Spring) SetMaxMotionRange(xRatio, yRatio float64) {
	if xRatio <= 0.0 {
		self.xRatio = 0.03
	} else {
		self.xRatio = xRatio
	}
	if yRatio <= 0.0 {
		self.yRatio = 0.03
	} else {
		self.yRatio = yRatio
	}
}

// Same idea as [Bezier.SetZoomCompensation](). Defaults to 0.
func (self *Spring) SetZoomCompensation(compensation float64) {
	if compensation < 0 || compensation > 1.0 {
		panic("zoom compensation factor must be in [0, 1]")
	}
	self.zoomCompensation = compensation
}

// Sets the internal spring simulation parameters.
// Defaults are (0.25, 80.0), but it depends a lot on
// the motion ranges too.
func (self *Spring) SetParameters(damping, power float64) {
	if damping < 0.0 || damping > 1.0 {
		panic("damping must be in [0, 1] range")
	}
	if power <= 0.0 {
		panic("power must be strictly positive")
	}
	self.spring.SetParameters(damping, power)
}

// Implements the [Shaker] interface.
func (self *Spring) GetShakeOffsets(level float64) (float64, float64) {
	self.ensureInitialized()
	if level == 0.0 {
		self.x, self.y = 0.0, 0.0
		self.xSpeed, self.ySpeed = 0.0, 0.0
		self.rerollTarget()
		return 0.0, 0.0
	}
	
	// bézier conic curve interpolation
	self.x, self.xSpeed = self.spring.Update(self.x, self.xTarget, self.xSpeed)
	self.y, self.ySpeed = self.spring.Update(self.y, self.yTarget, self.ySpeed)
	if internal.Abs(self.xTarget - self.x) < 0.08 && internal.Abs(self.yTarget - self.y) < 0.08 {
		self.rerollTarget()
	}
	
	// translate interpolated point to real screen distances
	w, h := internal.GetResolution()
	w64, h64 := float64(w), float64(h)
	zoom := internal.GetCurrentZoom()
	xOffset, yOffset := self.x*w64*self.xRatio, self.y*h64*self.yRatio
	if self.zoomCompensation != 0.0 {
		compensatedZoom := 1.0 + (zoom - 1.0)*self.zoomCompensation
		xOffset /= compensatedZoom
		yOffset /= compensatedZoom
	}
	if level != 1.0 {
		xOffset *= level
		yOffset *= level
	}
	
	return xOffset, yOffset
}

func (self *Spring) rerollTarget() {
	self.xTarget, self.yTarget = rand.Float64() - 0.5, rand.Float64() - 0.5
}
