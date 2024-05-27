package shaker

import "math/rand/v2"

import "github.com/tinne26/mipix/internal"

// Very basic implementation of a [Shaker] using random values.
type Random struct {
	fromX, fromY float64
	toX, toY float64
	ticksElapsed uint64
	
	ptpTicks uint64
	axisRatio float64
	zoomCompensated bool
	initialized bool
}

// Sets the maximum travel distance between shake points, as
// a ratio to be applied to the game's logical resolution.
//
// For example, if the game resolution is 64x64 and you set
// a maximum shake of 0.25, the shake will range within [-8,
// +8] for both axes.
//
// Reasonable values range from 0.05 to 0.3. Values <= 0.0
// will be silently discarded and the default of 0.15 will
// be restored.
func (self *Random) SetMaxMotionRange(axisRatio float64) {
	if axisRatio <= 0.0 {
		self.axisRatio = 0.0
		self.ensureInitialized()
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

// Sets the number of ticks the shaker takes to go from one randomly
// rolled point to the next. At 60TPS, reasonable values range from
// 3 to 30 (or 12 to 120 at 240TPS). Zero is not allowed.
func (self *Random) SetPointToPointTicks(ticks TicksDuration) {
	if ticks == 0 { panic("ticks must be > 0") }
	self.ptpTicks = uint64(ticks)
}

// Implements the [Shaker] interface.
func (self *Random) GetShakeOffsets(level float64) (float64, float64) {
	self.ensureInitialized()
	t := float64(self.ticksElapsed)/float64(self.ptpTicks)
	x := internal.QuadInOutInterp(self.fromX, self.toX, t)
	y := internal.QuadInOutInterp(self.fromY, self.toY, t)
	self.ticksElapsed += internal.GetTPU()
	if self.ticksElapsed >= self.ptpTicks {
		self.rollNewTarget()
		for self.ticksElapsed >= self.ptpTicks {
			self.ticksElapsed -= self.ptpTicks
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

func (self *Random) ensureInitialized() {
	if !self.initialized {
		self.rollNewTarget()
		if self.axisRatio == 0.0 {
			self.axisRatio = 0.02
		}
		if self.ptpTicks == 0 {
			ticksPerSecond := float64(internal.GetTPU()*uint64(internal.GetUPS()))
			self.ptpTicks = max(uint64(ticksPerSecond*0.01), 1)
		}
		self.initialized = true
	}
}

func (self *Random) rollNewTarget() {
	self.fromX, self.fromY = self.toX, self.toY
	self.toX = rand.Float64() - 0.5
	self.toY = rand.Float64() - 0.5
}
