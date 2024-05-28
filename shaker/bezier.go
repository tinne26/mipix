package shaker

import "github.com/tinne26/mipix/internal"

var _ Shaker = (*Bezier)(nil)

// Implementation of a [Shaker] using bézier curves in
// strange ways.
//
// This shaker has a fair share of personality. I would
// say it's quite biased and unpleasant, like someone 
// throwing a tantrum.
//
// The implementation is tick-rate independent.
type Bezier struct {
	ax, ay float64
	bx, by float64
	ctrlx, ctrly float64
	
	elapsed float64
	travelTime float64
	axisRatio float64
	zoomCompensation float64
	initialized bool
}

// To preserve resolution independence, shakers often simulate the
// shaking within a [-0.5, 0.5] space and only later scale it. For
// example, if you have a resolution of 32x32 and set a motion
// scale of 0.25, the shaking will range within [-4, +4] in both
// axes.
// 
// Defaults to 0.05.
func (self *Bezier) SetMotionScale(axisScalingFactor float64) {
	if axisScalingFactor <= 0.0 { panic("axisScalingFactor must be strictly positive") }
	self.axisRatio = axisScalingFactor
}

// The range of motion of most shakers is based on the logical
// resolution of the game. This means that when zooming in or
// out, the shaking effect will become more or less pronounced,
// respectively. If you want the shaking to maintain the same
// relative magnitude regardless of zoom level, change the zoom
// compensation from 0 (the default) to 1.
func (self *Bezier) SetZoomCompensation(compensation float64) {
	if compensation < 0 || compensation > 1.0 {
		panic("zoom compensation factor must be in [0, 1]")
	}
	self.zoomCompensation = compensation
}

// Change the travel time between generated shake points. Defaults to 0.1.
func (self *Bezier) SetTravelTime(travelTime float64) {
	if travelTime <= 0 { panic("travel time must be strictly positive") }
	self.travelTime = travelTime
}

// Implements the [Shaker] interface.
func (self *Bezier) GetShakeOffsets(level float64) (float64, float64) {
	self.ensureInitialized()
	if level == 0.0 {
		self.elapsed = 0.0
		self.rerollOriginPoints()
		return 0.0, 0.0
	}
	
	// bézier conic curve interpolation
	t := self.elapsed/self.travelTime
	lerp := func(x1, y1, x2, y2, t float64) (float64, float64) {
		return internal.LinearInterp(x1, x2, t), internal.LinearInterp(y1, y2, t)
	}
	ocx, ocy := lerp(self.ax, self.ay, self.ctrlx, self.ctrly, t) // origin to control
	cfx, cfy := lerp(self.ctrlx, self.ctrly, self.bx, self.by, t) // control to end
	ix , iy  := lerp(ocx, ocy, cfx, cfy, t) // interpolated point

	// roll new point, slide previous
	self.elapsed += 1.0/float64(internal.GetUPS())
	if self.elapsed >= self.travelTime {
		self.ax, self.ay = self.bx, self.by
		self.ctrlx, self.ctrly = self.rollNewPoint()
		self.bx, self.by = self.rollNewPoint()
		for self.elapsed >= self.travelTime {
			self.elapsed -= self.travelTime
		}
	}
	
	// translate interpolated point to real screen distances
	w, h := internal.GetResolution()
	w64, h64 := float64(w), float64(h)
	zoom := internal.GetCurrentZoom()
	xOffset, yOffset := ix*w64*self.axisRatio, iy*h64*self.axisRatio
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

func (self *Bezier) ensureInitialized() {
	if self.initialized { return }
	self.initialized = true
	if self.axisRatio == 0.0 {
		self.axisRatio = 0.05
	}
	if self.travelTime == 0.0 {
		self.travelTime = 0.1
	}
	self.rerollOriginPoints()
}

func (self *Bezier) rerollOriginPoints() {
	self.ax, self.ay = 0.0, 0.0
	self.bx, self.by = self.rollNewPoint()
	self.ctrlx, self.ctrly = self.rollNewPoint()
}

func (self *Bezier) rollNewPoint() (float64, float64) {
	return internal.RollPointWithinEllipse(1.0, 1.0)
}
