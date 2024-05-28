package shaker

import "math"
import "math/rand/v2"

import "github.com/tinne26/mipix/internal"

var _ Shaker = (*Balanced)(nil)

// Cubic bézier curves with start and end points at (0, 0),
// in some kind of circular oscillation.
//
// You can set the motion scale to 0.02 and travel time to
// 1.4 for a soft ship-like motion. The default is more like
// a dampened earthquake. The shaking is fairly homogeneous
// and consistent in speed and motion. You could probably
// dive into the code and tweak a few things here and there
// to destabilize speed and get something more similar to a
// drunken effect. You probably shouldn't, but you could.
//
// The implementation is tick-rate independent.
type Balanced struct {
	rads float64
	cx1, cy1 float64
	cx2, cy2 float64

	elapsed float64
	travelTime float64
	axisRatio float64
	zoomCompensation float64
	initialized bool
}

func (self *Balanced) ensureInitialized() {
	if self.initialized { return }
	self.initialized = true
	if self.axisRatio == 0.0 {
		self.axisRatio = 0.01
	}
	if self.travelTime == 0.0 {
		self.travelTime = 0.05
	}
	self.rads = rand.Float64()*2.0*math.Pi
	self.rerollControlPoints()
}

// To preserve resolution independence, shakers often simulate the
// shaking within a [-0.5, 0.5] space and only later scale it. For
// example, if you have a resolution of 32x32 and set a motion
// scale of 0.25, the shaking will range within [-4, +4] in both
// axes.
// 
// Defaults to 0.01.
func (self *Balanced) SetMotionScale(axisScalingFactor float64) {
	if axisScalingFactor <= 0.0 { panic("axisScalingFactor must be strictly positive") }
	self.axisRatio = axisScalingFactor
}

// The range of motion of most shakers is based on the logical
// resolution of the game. This means that when zooming in or
// out, the shaking effect will become more or less pronounced,
// respectively. If you want the shaking to maintain the same
// relative magnitude regardless of zoom level, change the zoom
// compensation from 0 (the default) to 1.
func (self *Balanced) SetZoomCompensation(compensation float64) {
	if compensation < 0 || compensation > 1.0 {
		panic("zoom compensation factor must be in [0, 1]")
	}
	self.zoomCompensation = compensation
}

// Change the travel time between generated shake points. Defaults to 0.05.
func (self *Balanced) SetTravelTime(travelTime float64) {
	if travelTime <= 0 { panic("travel time must be strictly positive") }
	self.travelTime = travelTime
}

// Implements the [Shaker] interface.
func (self *Balanced) GetShakeOffsets(level float64) (float64, float64) {
	self.ensureInitialized()
	if level == 0.0 {
		self.elapsed = 0.0
		self.rads = rand.Float64()*2.0*math.Pi
		self.rerollControlPoints()
		return 0.0, 0.0
	}
	
	// bézier cubic curve interpolation
	t := self.elapsed/self.travelTime
	lerp := func(x1, y1, x2, y2, t float64) (float64, float64) {
		return internal.LinearInterp(x1, x2, t), internal.LinearInterp(y1, y2, t)
	}
	oc1x , oc1y  := lerp(0.0, 0.0, self.cx1, self.cy1, t)           // origin to control 1
	c1c2x, c1c2y := lerp(self.cx1, self.cy1, self.cx2, self.cy2, t) // control 1 to control 2
	c2fx , c2fy  := lerp(self.cx2, self.cy2, 0.0, 0.0, t)           // control 2 to end
	iox  , ioy   := lerp(oc1x, oc1y, c1c2x, c1c2y, t) // first interpolation from origin
	ifx  , ify   := lerp(c1c2x, c1c2y, c2fx, c2fy, t) // second interpolation to end
	ix   , iy    := lerp(iox, ioy, ifx, ify, t)       // interpolated

	// roll new point, slide previous
	self.elapsed += 1.0/float64(internal.GetUPS())
	if self.elapsed >= self.travelTime {
		self.rerollControlPoints()
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

func (self *Balanced) rerollControlPoints() {
	length := 0.8 + rand.Float64()*0.2
	sin, cos := math.Sincos(self.rads)
	self.cx1 = cos*length
	self.cy1 = sin*length
	
	// shift angle for the exit direction, which will
	// also be used as the entry direction for the next
	// point (with an 180 degree offset)
	self.rads += math.Pi*rand.Float64()*0.3333 // yes, shift in a consistent direction
	if self.rads >= 2.0*math.Pi { self.rads -= 2.0*math.Pi }

	sin, cos = math.Sincos(self.rads)
	self.cx2 = cos*length
	self.cy2 = sin*length

	// apply offset for next angle
	self.rads += math.Pi
	if self.rads >= 2.0*math.Pi { self.rads -= 2.0*math.Pi }
}
