package mipix

var defaultSimpleZoomer *SimpleZoomer
var _ Zoomer = (*SimpleZoomer)(nil)

// A default implementation of the [Zoomer] interface.
type SimpleZoomer struct {
	speed float64
	adjustedTarget float64
}

// Implements [Zoomer].
func (self *SimpleZoomer) Reset() {
	self.adjustedTarget, _ = Camera().GetZoom()
	self.speed = 0.0
}

// Implements [Zoomer].
func (self *SimpleZoomer) Update(currentZoom, targetZoom float64) float64 {
	const MaxZoomTracking float64 = 5.0

	// The idea behind the maths is the following: using linear interpolation
	// for speeds already results in smooth changes (integrating a linear
	// function gives you a quadratic one). There are some problems with a
	// naive approach, though:
	// - Massive changes can lead to non-smooth deltas. We correct for this
	//   by setting a maximum distance value and clamping.
	// - Sudden target changes, which are not common with movement, but are
	//   common with zoom changes, can look unpleasant. What we do here is
	//   not registering a new target directly, but instead get closer and
	//   closer to it progressively, with self.adjustedTarget. There are
	//   still some edge cases, but we smooth that with an extra speed
	//   interpolation.

	updateDelta := 1.0/float64(Tick().UPS())
	if targetZoom != self.adjustedTarget {
		distance := targetZoom - self.adjustedTarget
		normDist := clamp(distance, -MaxZoomTracking, MaxZoomTracking)
		targetApproximation := normDist*1.6*updateDelta
		if abs(targetZoom - currentZoom) < abs(distance - targetApproximation) {
			self.adjustedTarget = currentZoom
		} else {	
			self.adjustedTarget += targetApproximation
		}
	}

	newSpeed := linearInterp(0.0, self.adjustedTarget - currentZoom, 0.15)
	self.speed = linearInterp(self.speed, newSpeed, 3.0*updateDelta)
	return self.speed*updateDelta*20.0
}

// Another implementation of the [Zoomer] interface.
//
// Unlike [SimpleZoomer], it doesn't have any bouncing,
// but turn arounds mid-transition are less smooth. It
// also has a customizable speed factor. It still presents
// a quadratic in/out response.
type SimpleZoomer2 struct {
	speedFactorOffset float64
	adjustedTarget float64
}

// Implements [Zoomer].
func (self *SimpleZoomer2) Reset() {
	self.adjustedTarget, _ = Camera().GetZoom()
}

// Speed factor must be strictly positive. Defaults to 1.0.
func (self *SimpleZoomer2) SetSpeedFactor(factor float64) {
	if factor <= 0 { panic("zoom speed factor must be strictly positive") }
	self.speedFactorOffset = factor - 1.0
}

// Implements [Zoomer].
func (self *SimpleZoomer2) Update(currentZoom, targetZoom float64) float64 {
	const MaxZoomTracking float64 = 5.0

	updateDelta := 1.0/float64(Tick().UPS())
	if targetZoom != self.adjustedTarget {
		var dir float64 = 1.0
		if targetZoom < self.adjustedTarget { dir = -1.0 }

		distance := targetZoom - self.adjustedTarget
		normDist := clamp(distance, -MaxZoomTracking, MaxZoomTracking)
		targetApproximation := normDist*1.6*updateDelta + dir*updateDelta/2.0
		if abs(targetZoom - currentZoom) < abs(distance - targetApproximation) {
			self.adjustedTarget = currentZoom
			return 0.0
		} else {	
			self.adjustedTarget += targetApproximation
			switch dir {
			case +1.0: self.adjustedTarget = min(self.adjustedTarget, targetZoom)
			case -1.0: self.adjustedTarget = max(self.adjustedTarget, targetZoom)
			}
		}
	}

	change := linearInterp(0.0, self.adjustedTarget - currentZoom, 2.6*updateDelta)*(1.0 + self.speedFactorOffset)
	if change < 0 {
		change -= updateDelta/3.0
		change = max(change, targetZoom - currentZoom)
	} else if change > 0 {
		change += updateDelta/3.0
		change = min(change, targetZoom - currentZoom)
	}
	return change
}

// Another implementation of the [Zoomer] interface.
//
// Unlike [SimpleZoomer] and [SimpleZoomer2], this one is
// purely linear. Speeds can be modulated through
// [SimpleZoomer3.SetSpeed]().
//
// Notice that zooms might still not look "linear", but that's
// because going from x1.0 to x2.0 zoom doesn't result in /2.0
// the surface to draw, but /4.0.
type SimpleZoomer3 struct {
	speed float64
	speedTransitionIni float64
	speedTransitionEnd float64
	speedTransitionTicks TicksDuration
	speedTransitionTicksElapsed TicksDuration
}

const simpleZoomer3DefaultSpeedOffset = 1.0

func (self *SimpleZoomer3) Reset() {
	// nothing to do here
}

func (self *SimpleZoomer3) Update(currentZoom, targetZoom float64) float64 {
	if self.speedTransitionTicksElapsed < self.speedTransitionTicks {
		self.speedTransitionTicksElapsed += TicksDuration(Tick().GetRate())
		self.speedTransitionTicksElapsed  = min(self.speedTransitionTicksElapsed, self.speedTransitionTicks)
		t := float64(self.speedTransitionTicksElapsed)/float64(self.speedTransitionTicks)
		self.speed  = linearInterp(self.speedTransitionIni, self.speedTransitionEnd, t)
		self.speed -= simpleZoomer3DefaultSpeedOffset
	}

	if targetZoom == currentZoom { return 0.0 }
	updateSpeed := (self.speed + simpleZoomer3DefaultSpeedOffset)*(1.0/float64(Tick().UPS()))
	if currentZoom < targetZoom {
		return min(updateSpeed, targetZoom - currentZoom)
	} else {
		return max(-updateSpeed, targetZoom - currentZoom)
	}
}

// Speed unit is pixels per second.
func (self *SimpleZoomer3) SetSpeed(newSpeed float64, transition TicksDuration) {
	if newSpeed < 0 { newSpeed = -newSpeed }
	self.speedTransitionTicks = transition
	self.speedTransitionTicksElapsed = 0
	if transition == ZeroTicks {
		self.speed = newSpeed - simpleZoomer3DefaultSpeedOffset
	} else {
		self.speedTransitionIni = self.speed
		self.speedTransitionEnd = newSpeed
	}
}

// Similar to [SimpleZoomer3], but made look truly linear by
// multiplying the speed by the current zoom level.
type SimpleZoomer4 struct {
	speed float64
	speedTransitionIni float64
	speedTransitionEnd float64
	speedTransitionTicks TicksDuration
	speedTransitionTicksElapsed TicksDuration
}

const simpleZoomer4DefaultSpeedOffset = 1.0

func (self *SimpleZoomer4) Reset() {
	// nothing to do here
}

func (self *SimpleZoomer4) Update(currentZoom, targetZoom float64) float64 {
	if self.speedTransitionTicksElapsed < self.speedTransitionTicks {
		self.speedTransitionTicksElapsed += TicksDuration(Tick().GetRate())
		self.speedTransitionTicksElapsed  = min(self.speedTransitionTicksElapsed, self.speedTransitionTicks)
		t := float64(self.speedTransitionTicksElapsed)/float64(self.speedTransitionTicks)
		self.speed  = linearInterp(self.speedTransitionIni, self.speedTransitionEnd, t)
		self.speed -= simpleZoomer4DefaultSpeedOffset
	}

	if targetZoom == currentZoom { return 0.0 }
	updateSpeed := currentZoom*(self.speed + simpleZoomer4DefaultSpeedOffset)*(1.0/float64(Tick().UPS()))
	if currentZoom < targetZoom {
		return min(updateSpeed, targetZoom - currentZoom)
	} else {
		return max(-updateSpeed, targetZoom - currentZoom)
	}
}

// Speed unit is pixels per second.
func (self *SimpleZoomer4) SetSpeed(newSpeed float64, transition TicksDuration) {
	if newSpeed < 0 { newSpeed = -newSpeed }
	self.speedTransitionTicks = transition
	self.speedTransitionTicksElapsed = 0
	if transition == ZeroTicks {
		self.speed = newSpeed - simpleZoomer4DefaultSpeedOffset
	} else {
		self.speedTransitionIni = self.speed
		self.speedTransitionEnd = newSpeed
	}
}
