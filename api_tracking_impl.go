package mipix

// import "github.com/hajimehoshi/ebiten/v2"

var (
	FrozenTracker  Tracker = frozenTracker{}
	InstantTracker Tracker = instantTracker{}
	LinearTracker  Tracker = linearTracker{}
)

type frozenTracker struct {}
func (frozenTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	return 0, 0
}

type instantTracker struct{}
func (instantTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	return targetX - currentX, targetY - currentY
}

func linearInterp(a, b, t float64) float64 { return (a + (b - a)*t) }
func smoothInterp(a, b, t float64) float64 { // related: https://iquilezles.org/articles/smoothsteps
	t = clamp(t, 0, 1)
	return linearInterp(a, b, t*t*(3.0 - 2.0*t))
}
func quadInterp(a, b, t float64) float64 {
	return linearInterp(a, b, quadInOut(t))
}
func quadInOut(t float64) float64 {
	t = clamp(t, 0, 1)
	if t < 0.5 { return 2*t*t }
	t = 2*t - 1
	return -0.5*(t*(t - 2) - 1)
}

// A simple linear interpolation tracker.
type linearTracker struct {}

// TODO: add inertia/tension on slightly more advanced version, and/or min/max speeds
func (self linearTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	ratio := 0.094 // (60.0/float64(ebiten.TPS()))*N
	
	var resultSpeedX, resultSpeedY float64
	for range Tick().GetRate() {
		speedX := self.updateComponent(currentX, targetX, ratio)
		currentX += speedX
		resultSpeedX += speedX
		speedY := self.updateComponent(currentY, targetY, ratio)
		currentY += speedY
		resultSpeedY += speedY
	}
	return resultSpeedX, resultSpeedY
}

func (linearTracker) updateComponent(current, target, ratio float64) float64 {
	const MinSpeed = 0.064
	
	// determine base speed
	if target > current { // going right
		speed := linearInterp(0, target - current, ratio)
		if speed >= MinSpeed { return speed }
		return min(MinSpeed, target - current)
	} else { // going left
		speed := linearInterp(0, target - current, ratio)
		if -speed >= MinSpeed { return speed }
		return max(-MinSpeed, target - current)
	}
}

// TODO
type SpringTracker struct {
	k float64
	initialized bool
}

// Sets the spring constant. This must be manually invoked.
// Some common reference values:
//  - 640x360: 
//  - 33x32: 
func (self *SpringTracker) SetStiffness(k float64) {
	self.k = k
}
func (self *SpringTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	const timescaling = 0.02 
	extensionX, extensionY := (targetX - currentX), (targetY - currentY)
	// this might be best characterized by horzTravel, vertTravel,
	// maxHorzSpeed, maxVertSpeed (to reach horzTravel and vertTravel),
	// and horzSpringRatio (below 1, "going back from an extended position").
	// I think I can initialize reasonable values for all those, and
	// then its only about whether I can figure out all the remaining
	// parameters.
	return self.k*extensionX*timescaling, self.k*extensionY*timescaling
}
