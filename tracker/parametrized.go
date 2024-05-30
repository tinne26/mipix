package tracker

import "github.com/tinne26/mipix/internal"

var _ Tracker = (*Parametrized)(nil)

// A configurable linear tracker.
type Parametrized struct {
	// max speed basic parameters
	maxScreensPerSecond float64
	screensToMaxSpeed float64

	// quirky extra parameters
	minScreensPerSecond float64 // instant tracking below this
	screensToMinSpeed float64 // frozen tracking below this

	// initialization
	initialized bool
}

func (self *Parametrized) initialize() {
	self.initialized = true
	if self.maxScreensPerSecond == 0.0 {
		self.maxScreensPerSecond = 2.0
	}
	if self.screensToMaxSpeed == 0.0 {
		self.screensToMaxSpeed = 0.5
	}
}

//  - maxScreensPerSecond is the maximum speed you want to allow the camera to move at.
//    It defaults to 2.0 screens per second.
//  - screensToMaxSpeed is the distance at which the maximum speed is reached. The
//    difference between the camera's target and the current position must be >= screensToMaxSpeed
//    for the camera to reach maximum speed. The default is 0.5.
func (self *Parametrized) SetMaxSpeed(maxScreensPerSecond, screensToMaxSpeed float64) {
	if maxScreensPerSecond <= 0.0 {
		panic("maxScreensPerSecond must be > 0")
	}
	if screensToMaxSpeed <= 0.0 {
		panic("screensToMaxSpeed must be > 0")
	}

	self.maxScreensPerSecond = maxScreensPerSecond
	self.screensToMaxSpeed   = screensToMaxSpeed
}

// If set, speeds below the given threshold will result in instantaneous tracking.
func (self *Parametrized) SetInstantTrackingBelow(screensPerSecond float64) {
	if screensPerSecond < 0 { panic("'screensPerSecond' can't be a negative speed") }
	self.minScreensPerSecond = screensPerSecond
}

// If set, the tracking error (difference between camera's target and current positions)
// must reach the given distance in screens before the tracker starts moving.
func (self *Parametrized) SetFrozenTrackingBelow(screens float64) {
	if screens < 0 { panic("'screens' must be a non-negative distance") }
	self.screensToMinSpeed = screens
}

// Implements [Tracker].
func (self *Parametrized) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	if !self.initialized { self.initialize() }

	w, h := internal.GetResolution()
	widthF64, heightF64 := float64(w), float64(h)
	
	updateDelta := 1.0/float64(internal.GetUPS())
	xAdvance := self.updateComponent(currentX, targetX, widthF64 , updateDelta)
	yAdvance := self.updateComponent(currentY, targetY, heightF64, updateDelta)
	return xAdvance, yAdvance
}

func (self *Parametrized) updateComponent(current, target, screen, updateDelta float64) float64 {
	distance := target - current
	zoomedScreen := screen/internal.GetCurrentZoom()
	
	// frozen tracking
	frozenDistance := self.screensToMinSpeed*zoomedScreen
	if distance >= 0 {
		if distance <= frozenDistance { return 0 }
		distance -= frozenDistance
	} else { // distance < 0
		if -distance <= frozenDistance { return 0 }
		distance += frozenDistance
	}

	// compute speed
	t := internal.TAt(internal.Abs(distance)*internal.GetCurrentZoom(), 0, self.screensToMaxSpeed*zoomedScreen)
	normSpeed := internal.LinearInterp(self.minScreensPerSecond, self.maxScreensPerSecond, t)
	change := normSpeed*zoomedScreen*updateDelta

	// clamp
	if distance >= 0 {
		return min(distance, change)
	} else { // distance < 0
		return max(distance, -change)
	}
}
