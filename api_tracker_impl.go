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

	horzAdvance := computeLinComponent(currentX, targetX, minAdvance, maxHorzAdvance, refHorzMaxDist)
	vertAdvance := computeLinComponent(currentY, targetY, minAdvance, maxVertAdvance, refVertMaxDist)
	return horzAdvance, vertAdvance
}

func computeLinComponent(current, target, minAdvance, maxAdvance, refMaxDist float64) float64 {
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

// A default implementation of the [Zoomer] interface
// using linear interpolation and basic movement prediction.
type SimpleTracker struct {
	xCompensation float64
	yCompensation float64
}

var _ Tracker = (*SimpleTracker)(nil)

// Implements [Tracker].
func (self *SimpleTracker) Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64) {
	w, h := GetResolution()
	zoom, _ := Camera().GetZoom()
	widthF64, heightF64 := float64(w)/zoom, float64(h)/zoom
	
	// TODO: the use of zoom here is not very clean
	updateDelta := 1.0/float64(Tick().UPS())
	maxHorzAdvance := 6.0*zoom*widthF64*updateDelta  // use higher values for a more rigid / strict tracking
	maxVertAdvance := 6.0*zoom*heightF64*updateDelta // use lower values for a more elastic / softer tracking
	minAdvance := 0.01*updateDelta
	refHorzMaxDist := 3.5*widthF64 // higher values lead to smoother tracking
	refVertMaxDist := 3.5*heightF64

	horzAdvance := computeLinComponent(currentX, targetX, minAdvance, maxHorzAdvance, refHorzMaxDist)
	vertAdvance := computeLinComponent(currentY, targetY, minAdvance, maxVertAdvance, refVertMaxDist)

	// compensation for automatic following...
	const ErrorRatio = 0.3 // how much of the error we want to allow correcting
	const RecoverySpeed = 1.4 // how fast we want to recover a neutral position / no compensation
	const CompensationPull = 0.008 // higher values lead to compensation ramping up faster *
	// * this could be expanded into its own function, e.g. an easyInQuad might work,
	//   with t based on error relative to widthF64. alternatively, I feel a min error
	//   might be allowed instead of the error ratio and so on. this allows going for
	//   higher ramp ups that better split the effects of natural linear following and
	//   "wait let me catch you". but I might need to store the 0 compensation point
	//   for this to reset. yeah, currently discontinuous movements lead to janky
	//   behavior, so we definitely need something smarter and smoother

	xError := (targetX - currentX)*ErrorRatio
	switch {
	case self.xCompensation > 0 && xError < self.xCompensation:
		self.xCompensation -= RecoverySpeed*updateDelta
		if self.xCompensation < 0.001 { self.xCompensation = 0.0 }
	case self.xCompensation < 0 && xError > self.xCompensation:
		self.xCompensation += RecoverySpeed*updateDelta
		if self.xCompensation > -0.001 { self.xCompensation = 0.0 }
	default:
		self.xCompensation += sign(xError)*widthF64*CompensationPull*updateDelta
		self.xCompensation = clampTowardsZero(self.xCompensation, xError)
	}

	yError := (targetY - currentY)*ErrorRatio
	switch {
	case self.yCompensation > 0 && yError < self.yCompensation:
		self.yCompensation -= RecoverySpeed*updateDelta
		if self.yCompensation < 0.001 { self.yCompensation = 0.0 }
	case self.yCompensation < 0 && yError > self.yCompensation:
		self.yCompensation += RecoverySpeed*updateDelta
		if self.yCompensation > -0.001 { self.yCompensation = 0.0 }
	default:
		self.yCompensation += sign(yError)*heightF64*CompensationPull*updateDelta
		self.yCompensation = clampTowardsZero(self.yCompensation, yError)
	}
	
	return horzAdvance + self.xCompensation*18*updateDelta, vertAdvance + self.yCompensation*18*updateDelta
}

func sim(predictedChange, actualChange float64, maxErrorForZeroSimilarity float64) float64 {
	predictionError := abs(actualChange - predictedChange)
	if predictionError > maxErrorForZeroSimilarity { return 0.0 }
	return 1.0 - predictionError/maxErrorForZeroSimilarity
}
