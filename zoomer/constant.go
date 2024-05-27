package zoomer

import "github.com/tinne26/mipix/internal"

var _ Zoomer = (*Constant)(nil)

// A very simple zoomer that modifies the zoom at a constant
// speed, which can be changed through [Constant.SetSpeed]().
//
// By default, the change is not purely linear, though, it's
// multiplied by the current zoom level. This can be changed
// through [Constant.SetZoomCompensated](false). The reason
// for this is that going from x1.0 to x2.0 zoom doesn't result
// in /2.0 the surface to draw, but /4.0. Therefore, to have
// a perceptually linear change in zoom, using a linear speed
// doesn't quite work, we also need to multiply the change by
// the current zoom level.
//
// The implementation is update-rate independent.
type Constant struct {
	speedTransitionIni float64
	speedTransitionEnd float64
	speedTransitionLength TicksDuration
	speedTransitionElapsed TicksDuration
	zoomCompensationDisabled bool
}

// The default speed is 1.0, reasonable values range between [0.5, 3.0].
//
// The method also requires a second parameter indicating the duration of
// the transition from the old speed to the new one, in ticks.
func (self *Constant) SetSpeed(newSpeed float64, transition TicksDuration) {
	if newSpeed < 0 { newSpeed = -newSpeed }
	if transition == 0 {
		self.speedTransitionEnd = newSpeed - constantDefaultSpeedOffset
	} else {
		self.speedTransitionIni = self.getCurrentSpeed() - constantDefaultSpeedOffset
		self.speedTransitionEnd = newSpeed - constantDefaultSpeedOffset
	}
	self.speedTransitionLength = transition
	self.speedTransitionElapsed = 0
}

// With zoom compensation, the zoom looks perceptually linear.
// Without zoom compensation, zooming in seems to progressively slow
// down, and zooming out seems to progressively speed up. This is
// explained in more detail on the documentation of [Constant] itself.
// 
// Defaults to true.
func (self *Constant) SetZoomCompensated(compensated bool) {
	self.zoomCompensationDisabled = !compensated
}

const constantDefaultSpeedOffset = 1.0

func (self *Constant) Reset() {
	self.speedTransitionIni = 0.0
	self.speedTransitionEnd = 0.0
	self.speedTransitionLength = 0
	self.speedTransitionElapsed = 0
}

func (self *Constant) Update(currentZoom, targetZoom float64) float64 {
	speed := self.getAndAdvanceCurrentSpeed()
	if targetZoom == currentZoom { return 0.0 }
	updateSpeed := (speed + constantDefaultSpeedOffset)*(1.0/float64(internal.GetUPS()))
	if !self.zoomCompensationDisabled { updateSpeed *= currentZoom }
	if currentZoom < targetZoom {
		return min(updateSpeed, targetZoom - currentZoom)
	} else {
		return max(-updateSpeed, targetZoom - currentZoom)
	}
}

func (self *Constant) getCurrentSpeed() float64 {
	var speed float64
	if self.speedTransitionElapsed < self.speedTransitionLength {
		t := float64(self.speedTransitionElapsed)/float64(self.speedTransitionLength)
		speed = internal.LinearInterp(self.speedTransitionIni, self.speedTransitionEnd, t)
	} else {
		speed = self.speedTransitionEnd
	}
	return speed + constantDefaultSpeedOffset
}

func (self *Constant) getAndAdvanceCurrentSpeed() float64 {
	speed := self.getCurrentSpeed()
	if self.speedTransitionElapsed < self.speedTransitionLength {
		self.speedTransitionElapsed += TicksDuration(internal.GetTPU())
		self.speedTransitionElapsed  = min(self.speedTransitionElapsed, self.speedTransitionLength)
	}
	return speed
}
