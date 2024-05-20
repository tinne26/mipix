package mipix

import "math/rand/v2"

import "github.com/hajimehoshi/ebiten/v2"

var defaultSimpleShaker *SimpleShaker = &SimpleShaker{}
var _ Shaker = (*SimpleShaker)(nil)

type SimpleShaker struct {
	horzRange, vertRange float64
	fromX, fromY float64
	toX, toY float64
	progress float64
}

// Sets the maximum horizontal and vertical range of the shaker.
//
// As an example, when no shaker is set mipix initializes a
// SimpleShaker with a range of 1/80 of the shortest axis of
// the logical resolution.
func (self *SimpleShaker) SetRange(horz, vert float64) {
	self.horzRange, self.vertRange = horz, vert
}

// Implements the [Shaker] interface.
func (self *SimpleShaker) GetShakeOffsets(activity float64) (float64, float64) {
	// notice: completely sh*t temporary implementation
	// what about sine-like for y and noisy for x?
	// and Math.sqrt(1 - Math.pow(x - 1, 2)) for circular interpolation in the top-left quadrant
	// ...
	// ellipse equation: (x/horzAxisLength)^2+(y/vertAxisLength)^2 = 1
	// just intersect with line equation and solve the system.
	// so, just make a simple function that returns the value given
	// axis lengths and angle. you first find the intersection point
	// and then use a^2 + b^2 = h^2.
	// with bézier conic curves, sliding 2/3 points and rng angles

	x := smoothInterp(self.fromX, self.toX, self.progress)
	y := smoothInterp(self.fromY, self.toY, self.progress)
	self.progress += (40.0*float64(Tick().GetRate()))/float64(ebiten.TPS())
	if self.progress >= 1.0 {
		self.rollNewTarget()
		self.progress = 0.0
	}
	if activity == 1.0 { return x, y }
	return smoothInterp(0, x, activity), smoothInterp(0, y, activity)
}

func (self *SimpleShaker) rollNewTarget() {
	self.fromX, self.fromY = self.toX, self.toY
	self.toX = rand.Float64()*self.horzRange - self.horzRange/2.0
	self.toY = rand.Float64()*self.vertRange - self.vertRange/2.0
}

