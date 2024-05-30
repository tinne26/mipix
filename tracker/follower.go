package tracker

import "github.com/tinne26/mipix/internal"

type follower struct {
	engaged bool
	
	elapsedMatch float64
	elapsedHalt float64
	
	matchRequiredDuration float64
	haltRequiredDuration float64
	matchErrorMargin float64

	initialized bool
}

func (self *follower) initialize() {
	self.initialized = true
	if self.matchRequiredDuration == 0.0 {
		self.matchRequiredDuration = 1.0
	}
	if self.haltRequiredDuration == 0.0 {
		self.haltRequiredDuration = 0.5
	}
	self.matchErrorMargin = 0.05
}

func (self *follower) SetTimes(engage, disengage float64) {
	if engage < disengage { panic("engage time must be >= disengage") }
	self.matchRequiredDuration = engage
	self.haltRequiredDuration  = disengage
}

func (self *follower) IsEngaged() bool {
	return self.engaged
}

func (self *follower) Update(changeX, changeY, prevSpeedX, prevSpeedY float64) {
	if !self.initialized { self.initialize() }

	// helper values
	updateDelta := 1.0/float64(internal.GetUPS())
	speedX, speedY := internal.Abs(changeX/updateDelta), internal.Abs(changeY/updateDelta)
	w, h := internal.GetResolution()
	w64, h64 := float64(w), float64(h)
	zoom := internal.GetCurrentZoom()
	normWidth, normHeight := w64/zoom, h64/zoom
	
	// update elapsed match / halt
	if speedX/normWidth <= self.matchErrorMargin && speedY/normHeight <= self.matchErrorMargin {
		self.elapsedHalt += updateDelta
	} else {
		self.elapsedHalt = 0.0
	}
	halted := (self.elapsedHalt >= self.haltRequiredDuration)
	
	if !halted && self.elapsedMatch < self.matchRequiredDuration {
		normSpeedUDiffX := internal.Abs(speedX - internal.Abs(prevSpeedX))/normWidth
		normSpeedUDiffY := internal.Abs(speedY - internal.Abs(prevSpeedY))/normWidth
		if normSpeedUDiffX <= self.matchErrorMargin && normSpeedUDiffY <= self.matchErrorMargin {
			self.elapsedMatch += updateDelta
		} else {
			self.elapsedMatch = 0.0
		}
	} else if halted {
		self.elapsedMatch = 0.0
	}
	
	if halted || self.elapsedMatch < self.matchRequiredDuration {
		self.engaged = false
	} else {
		self.engaged = true
	}
}	
