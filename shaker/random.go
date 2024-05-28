package shaker

import "math/rand/v2"

import "github.com/tinne26/mipix/internal"

var _ Shaker = (*Random)(nil)

// Very basic implementation of a [Shaker] using random values.
//
// This shaker is not particularly nice; good screen shakes
// generally try to create a continuous movement, avoiding
// sharp direction changes and so on. Random is cheap (still
// classic and effective, though).
//
// The implementation is tick-rate independent.
type Random struct {
	fromX, fromY float64
	toX, toY float64
	
	elapsed float64
	travelTime float64
	axisRatio float64
	zoomCompensated bool
	initialized bool
}

func (self *Random) ensureInitialized() {
	if !self.initialized {
		self.rollNewTarget()
		if self.axisRatio == 0.0 {
			self.axisRatio = 0.02
		}
		if self.travelTime == 0 {
			self.travelTime = 0.03
		}
		self.initialized = true
	}
}

// Sets the maximum travel distance between shake points, as
// a ratio to be applied to the game's logical resolution.
//
// For example, if the game resolution is 64x64 and you set
// a maximum shake of 0.25, the shake will range within [-8,
// +8] for both axes.
//
// Reasonable values range from 0.005 to 0.3. Values <= 0.0
// will be silently discarded and the default of 0.02 will
// be restored.
func (self *Random) SetMaxMotionRange(axisRatio float64) {
	if axisRatio <= 0.0 {
		self.axisRatio = 0.02
	} else {
		self.axisRatio = axisRatio
	}
}

// The range of motion of the shaker is based on the logical
// resolution of the game. This means that when zooming in or
// out, the shaking effect will become more or less pronounced,
// respectively. If you want the shaking to maintain the same
// relative magnitude regardless of zoom level, set zoom
// compensated to true.
func (self *Random) SetZoomCompensated(compensated bool) {
	self.zoomCompensated = compensated
}

// Change the travel time between generated shake points. Defaults to 0.1.
func (self *Random) SetTravelTime(travelTime float64) {
	if travelTime <= 0 { panic("travel time must be strictly positive") }
	self.travelTime = travelTime
}

// Implements the [Shaker] interface.
func (self *Random) GetShakeOffsets(level float64) (float64, float64) {
	self.ensureInitialized()
	if level == 0.0 {
		self.elapsed = 0.0
		self.rollNewTarget()
		self.fromX, self.fromY = 0.0, 0.0
		return 0.0, 0.0
	}

	t := self.elapsed/self.travelTime
	x := internal.QuadInOutInterp(self.fromX, self.toX, t)
	y := internal.QuadInOutInterp(self.fromY, self.toY, t)
	self.elapsed += 1.0/float64(internal.GetUPS())
	if self.elapsed >= self.travelTime {
		self.rollNewTarget()
		for self.elapsed >= self.travelTime {
			self.elapsed -= self.travelTime
		}
	} 

	w, h := internal.GetResolution()
	axisRange := float64(min(w, h))*self.axisRatio
	x, y = x*axisRange, y*axisRange
	if self.zoomCompensated {
		currentZoom := internal.GetCurrentZoom()
		x /= currentZoom
		y /= currentZoom
	}
	if level == 1.0 { return x, y }
	return internal.CubicSmoothstepInterp(0, x, level), internal.CubicSmoothstepInterp(0, y, level)
}

func (self *Random) rollNewTarget() {
	self.fromX, self.fromY = self.toX, self.toY
	self.toX = rand.Float64() - 0.5
	self.toY = rand.Float64() - 0.5
}
