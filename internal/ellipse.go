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
