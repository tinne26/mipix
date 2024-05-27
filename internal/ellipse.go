package internal

import "math"
import "math/rand"

func RollPointWithinEllipse(width, height float64) (float64, float64) {
	// the tangent of angles approaching 90degs goes to infinite,
	// so I'm limiting it to 89.99 degrees at most
	const AsymptoteMargin = 0.02*math.Pi/180.0
	angle := rand.Float64()*(math.Pi - AsymptoteMargin) - (math.Pi/2.0 - AsymptoteMargin/2)
	slope := math.Tan(angle)

	// get half of the width and height
	width  /= 2.0
	height /= 2.0

	// line equation is y = slope*x
	// ellipse equation is (x/halfWidth)^2 + (y/halfHeight)^2 = 1
	// if we solve the system, we get:
	x := (height*width)/math.Sqrt(height*height + (slope*slope)*width*width)
	y := slope*x
	if rand.Float64() < 0.5 { x = -x }
	x *= EaseOutQuad(rand.Float64())
	y *= EaseOutQuad(rand.Float64())
	return x, y
}

// (x/a)^2 + (y/b)^2 = 1
// tan(φ) * startPoint_X + n = startPoint_Y

// ellipse equation: (x/horzAxisLength)^2+(y/vertAxisLength)^2 = 1
// just intersect with line equation and solve the system.
// so, just make a simple function that returns the value given
// axis lengths and angle. you first find the intersection point
// and then use a^2 + b^2 = h^2.
// with bézier conic curves, sliding 2/3 points and rng angles
