package mipix

var (
	FrozenTracker  Tracker = frozenTracker{}  // Update(...) always returns (0, 0)
	InstantTracker Tracker = instantTracker{} // Update(...) always returns (target - current)
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

// A simple linear interpolation tracker.
type linearTracker struct {}

func (self linearTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	w, h := GetResolution()
	zoom, _ := Camera().GetZoom()
	widthF64, heightF64 := float64(w)/zoom, float64(h)/zoom
	
	updateDelta := 1.0/float64(Tick().UPS())
	maxHorzAdvance := 6.0*zoom*widthF64*updateDelta  // use higher values for a more rigid / strict tracking
	maxVertAdvance := 6.0*zoom*heightF64*updateDelta // use lower values for a more elastic / softer tracking
	minAdvance := 0.01*updateDelta
	refHorzMaxDist := 2.0*widthF64 // higher values lead to smoother tracking
	refVertMaxDist := 2.0*heightF64

	horzAdvance := self.computeComponent(currentX, targetX, minAdvance, maxHorzAdvance, refHorzMaxDist)
	vertAdvance := self.computeComponent(currentY, targetY, minAdvance, maxVertAdvance, refVertMaxDist)
	return horzAdvance, vertAdvance
}

func (linearTracker) computeComponent(current, target, minAdvance, maxAdvance, refMaxDist float64) float64 {
	// determine base speed
	if target > current { // going right
		dist := min(target - current, refMaxDist)
		advance := linearInterp(0, maxAdvance, tAt(dist, 0, refMaxDist))
		if advance >= minAdvance { return advance }
		return min(minAdvance, dist)
	} else { // going left
		dist := min(current - target, refMaxDist)
		advance := linearInterp(0, maxAdvance, tAt(dist, 0, refMaxDist))
		if advance >= minAdvance { return -advance }
		return -min(minAdvance, dist)
	}
}

// TODO: consider inertia/tension for more advanced tracker
// Probably like linear tracker, but with X/Y predictions to self-correct through time.
// type SimpleTracker struct {
// 	predictedX float64
// 	predictedY float64
// }

// func (self *SimpleTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
// 	// ...
// 	return 0, 0
// }

// TODO
// type SpringTracker struct {
// 	k float64
// 	initialized bool
// }

// // Sets the spring constant. This must be manually invoked.
// // Some common reference values:
// //  - 640x360: 
// //  - 33x32: 
// func (self *SpringTracker) SetStiffness(k float64) {
// 	self.k = k
// }
// func (self *SpringTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
// 	const timescaling = 0.02 
// 	extensionX, extensionY := (targetX - currentX), (targetY - currentY)
// 	// this might be best characterized by horzTravel, vertTravel,
// 	// maxHorzSpeed, maxVertSpeed (to reach horzTravel and vertTravel),
// 	// and horzSpringRatio (below 1, "going back from an extended position").
// 	// I think I can initialize reasonable values for all those, and
// 	// then its only about whether I can figure out all the remaining
// 	// parameters.
// 	return self.k*extensionX*timescaling, self.k*extensionY*timescaling
// }
