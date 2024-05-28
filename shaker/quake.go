package shaker

import "math/rand/v2"

import "github.com/tinne26/mipix/internal"

var _ Shaker = (*Quake)(nil)

// Implementation of a [Shaker] with consistently oscillating
// movement in both axes, but with some irregularities in speed
// and travel distance. It's interesting because it has those
// unpredictable variances within a very predictable and rather
// smooth motion pattern.
//
// The implementation is tick-rate independent.
type Quake struct {
	x float64 // [-0.5, +0.5]
	y float64 // [-0.5, +0.5]
	fromX, fromY float64
	towardsX, towardsY float64 // can be cut short from [-0.5, 0.5]
	xSpeedIni, xSpeedEnd float64
	ySpeedIni, ySpeedEnd float64
	minSpeed, maxSpeed float64 // absolute values
	axisRatio float64
	zoomCompensation float64
	initialized bool
}

func (self *Quake) ensureInitialized() {
	if self.initialized { return }
	self.initialized = true
	if self.minSpeed == 0.0 {
		self.minSpeed = 5.0
	}
	if self.maxSpeed == 0.0 {
		self.maxSpeed = self.minSpeed*4.6
	}
	if self.axisRatio == 0.0 {
		self.axisRatio = 0.0225
	}
	self.towardsX, self.xSpeedIni, self.xSpeedEnd = self.reroll(0.0, 0.0)
	self.towardsY, self.ySpeedIni, self.ySpeedEnd = self.reroll(0.0, 0.0)
}

// Internally, both x and y oscillate at their own speeds.
// While going from side to side, these speeds can change
// between any random values within [minSpeed, maxSpeed].
//
// This method allows you to configure those speeds.
// The default values are (5.0, 23.0).
func (self *Quake) SetSpeedRange(minSpeed, maxSpeed float64) {
	if minSpeed <= 0.0 { panic("minSpeed must be strictly positive") }
	if maxSpeed < minSpeed { panic("maxSpeed must be >= than minSpeed") }
	self.minSpeed, self.maxSpeed = minSpeed, maxSpeed
}

// To preserve resolution independence, shakers often simulate the
// shaking within a [-0.5, 0.5] space and only later scale it. For
// example, if you have a resolution of 32x32 and set a motion
// scale of 0.25, the shaking will range within [-4, +4] in both
// axes.
// 
// Defaults to 0.0225.
func (self *Quake) SetMotionScale(axisScalingFactor float64) {
	if axisScalingFactor <= 0.0 { panic("axisScalingFactor must be strictly positive") }
	self.axisRatio = axisScalingFactor
}

// The range of motion of most shakers is based on the logical
// resolution of the game. This means that when zooming in or
// out, the shaking effect will become more or less pronounced,
// respectively. If you want the shaking to maintain the same
// relative magnitude regardless of zoom level, change the zoom
// compensation from 0 (the default) to 1.
func (self *Quake) SetZoomCompensation(compensation float64) {
	if compensation < 0 || compensation > 1.0 {
		panic("zoom compensation factor must be in [0, 1]")
	}
	self.zoomCompensation = compensation
}

// Implements the [Shaker] interface.
func (self *Quake) GetShakeOffsets(level float64) (float64, float64) {
	self.ensureInitialized()
	if level == 0.0 {
		self.x, self.y = 0.0, 0.0
		self.fromX, self.fromY = 0.0, 0.0
		self.towardsX, self.xSpeedIni, self.xSpeedEnd = self.reroll(0.0, 0.0)
		self.towardsY, self.ySpeedIni, self.ySpeedEnd = self.reroll(0.0, 0.0)
		return 0.0, 0.0
	}
	
	// update x/y
	updateDelta := 1.0/float64(internal.GetUPS())
	t := internal.TAt(self.x, self.fromX, self.towardsX)
	self.x += internal.LinearInterp(self.xSpeedIni, self.xSpeedEnd, t)*updateDelta
	if internal.TAt(self.x, self.fromX, self.towardsX) >= 1.0 {
		self.fromX = self.x
		self.towardsX, self.xSpeedIni, self.xSpeedEnd = self.reroll(self.x, self.xSpeedEnd)
	}
	t = internal.TAt(self.y, self.fromY, self.towardsY)
	self.y += internal.LinearInterp(self.ySpeedIni, self.ySpeedEnd, t)*updateDelta
	if internal.TAt(self.y, self.fromY, self.towardsY) >= 1.0 {
		self.fromY = self.y
		self.towardsY, self.ySpeedIni, self.ySpeedEnd = self.reroll(self.y, self.ySpeedEnd)
	}
	
	// translate interpolated point to real screen offsets
	w, h := internal.GetResolution()
	w64, h64 := float64(w), float64(h)
	zoom := internal.GetCurrentZoom()
	xOffset, yOffset := self.x*w64*self.axisRatio, self.y*h64*self.axisRatio
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

func (self *Quake) reroll(value, speed float64) (target, iniSpeed, endSpeed float64) {
	if value > 0.0 || (value == 0.0 && rand.Float64() < 0.5) {
		target = -(0.05 + rand.Float64()*0.45)
		iniSpeed = -max(internal.Abs(speed), self.minSpeed)
		endSpeed = -(self.minSpeed + rand.Float64()*(self.maxSpeed - self.minSpeed))
		speedDiff := (endSpeed - iniSpeed)*(internal.Abs(target - value))
		endSpeed = iniSpeed + speedDiff
	} else { // value < 0.0
		target = (0.05 + rand.Float64()*0.45)
		iniSpeed = max(internal.Abs(speed), self.minSpeed)
		endSpeed = (self.minSpeed + rand.Float64()*(self.maxSpeed - self.minSpeed))
		speedDiff := (endSpeed - iniSpeed)*(internal.Abs(target - value))
		endSpeed = iniSpeed + speedDiff
	}
	return
}
