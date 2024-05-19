package mipix

import "math"
import "image"

func (self *controller) cameraAreaGet() image.Rectangle {
	return self.cameraArea
}

func (self *controller) cameraAreaF64() (minX, minY, maxX, maxY float64) {
	zoomedWidth  := float64(self.logicalWidth )/self.zoomCurrent
	zoomedHeight := float64(self.logicalHeight)/self.zoomCurrent
	minX = self.trackerCurrentX - zoomedWidth /2.0 + self.shakeOffsetX
	minY = self.trackerCurrentY - zoomedHeight/2.0 + self.shakeOffsetY
	return minX, minY, minX + zoomedWidth, minY + zoomedHeight
}

func (self *controller) updateCameraArea() {
	minX, minY, maxX, maxY := self.cameraAreaF64()
	self.cameraArea = image.Rect(
		int(math.Floor(minX)), int(math.Floor(minY)),
		int(math.Ceil( maxX)), int(math.Ceil( maxY)),
	)
}

// ---- tracking ----

func (self *controller) cameraGetTracker() Tracker {
	return self.tracker
}

func (self *controller) cameraSetTracker(tracker Tracker) {
	if self.inDraw { panic("can't set tracker during draw stage") }
	self.tracker = tracker
}

func (self *controller) cameraNotifyCoordinates(x, y float64) {
	if self.inDraw { panic("can't notify tracking coordinates during draw stage") }
	self.trackerTargetX, self.trackerTargetY = x, y
}

func (self *controller) cameraResetCoordinates(x, y float64) {
	if self.inDraw { panic("can't reset camera coordinates during draw stage") }
	self.trackerTargetX , self.trackerTargetY  = x, y
	self.trackerCurrentX, self.trackerCurrentY = x, y
	self.updateCameraArea()
}

func (self *controller) cameraFlushCoordinates() {
	if self.lastFlushCoordinatesTick == self.currentTick { return }
	self.lastFlushCoordinatesTick = self.currentTick
	self.updateZoom()
	self.updateTracking()
	self.updateShake()
	self.updateCameraArea()
}

func (self *controller) updateTracking() {
	var tracker Tracker = self.tracker
	if tracker == nil { tracker = InstantTracker } // could use linear/default tracker with 0.1 to 8 speed limits
	self.trackerPrevSpeedX, self.trackerPrevSpeedY = tracker.Update(
		self.trackerCurrentX, self.trackerCurrentY,
		self.trackerTargetX, self.trackerTargetY,
		self.trackerPrevSpeedX, self.trackerPrevSpeedY,
	)
	self.trackerCurrentX += self.trackerPrevSpeedX
	self.trackerCurrentY += self.trackerPrevSpeedY
}

// --- zoom ---

func (self *controller) updateZoom() {
	if self.zoomTransitionElapsed >= self.zoomTransitionDuration {
		self.zoomCurrent = self.zoomTarget
		self.zoomStart = self.zoomTarget
	} else {
		t := float64(self.zoomTransitionElapsed)/float64(self.zoomTransitionDuration)
		self.zoomCurrent = quadInterp(self.zoomStart, self.zoomTarget, t) // TODO: make customizable? quad is nice though
		self.zoomTransitionElapsed += 1
	}
}

func (self *controller) updateShake() {
	if !self.cameraIsShaking() { return }
	var shaker Shaker = self.shaker
	if shaker == nil {
		shortSide := min(self.logicalWidth, self.logicalHeight)
		shakeRange := float64(shortSide)/80.0
		defaultSimpleShaker.SetRange(shakeRange, shakeRange)
		defaultSimpleShaker.rollNewTarget()
		shaker = defaultSimpleShaker
	}
	activity := self.getShakeActivity()
	self.shakeOffsetX, self.shakeOffsetY = shaker.GetShakeOffsets(activity)
	self.shakeElapsed += 1
}

func (self *controller) cameraZoom(newZoomLevel float64, transition TicksDuration) {
	if self.inDraw { panic("can't zoom during draw stage") }
	if newZoomLevel <= 0.0 { panic("can't zoom <= 0.0") }

	if transition == ZeroTicks {
		self.zoomCurrent = newZoomLevel
		self.zoomStart   = newZoomLevel
		self.zoomTarget  = newZoomLevel
	} else {
		self.zoomStart  = self.zoomCurrent
		self.zoomTarget = newZoomLevel
	}
	self.zoomTransitionDuration = transition
	self.zoomTransitionElapsed  = 0
}

func (self *controller) cameraZoomFrom(ifZoom, newZoomLevel float64, transition TicksDuration) {
	referenceDist := newZoomLevel - ifZoom
	if referenceDist == 0 { self.cameraZoom(newZoomLevel, 0) ; return }
	if referenceDist < 0 { referenceDist = -referenceDist }
	actualDist := newZoomLevel - self.zoomCurrent
	if actualDist == 0 { self.cameraZoom(newZoomLevel, 0) ; return }
	if actualDist < 0 { actualDist = -actualDist }
	relativeTransition := float64(transition)*(actualDist/referenceDist)
	// TODO: swing smoothing
	self.cameraZoom(newZoomLevel, TicksDuration(relativeTransition))
}

func (self *controller) cameraGetZoom() (current, target float64) {
	return self.zoomCurrent, self.zoomTarget
}

// ---- screenshake ----

func (self *controller) cameraSetShaker(shaker Shaker) {
	if self.inDraw { panic("can't set shaker during draw stage") }
	self.shaker = shaker
}

func (self *controller) cameraGetShaker() Shaker {
	return self.shaker
}

func (self *controller) cameraStartShake(fadeIn TicksDuration) {
	if self.inDraw { panic("can't start shake during draw stage") }
	activity := self.getShakeActivity()
	self.shakeFadeIn = fadeIn
	self.shakeDuration = maxUint32
	self.shakeFadeOut = 0
	self.shakeElapsed = TicksDuration(float64(fadeIn)*activity)
}

func (self *controller) cameraEndShake(fadeOut TicksDuration) {
	if self.inDraw { panic("can't end shake during draw stage") }
	activity := self.getShakeActivity()
	self.shakeDuration = self.shakeElapsed - self.shakeFadeIn
	self.shakeFadeOut  = fadeOut
	self.shakeElapsed  = self.shakeFadeIn + self.shakeDuration
	self.shakeElapsed += TicksDuration(float64(fadeOut)*(1.0 - activity))
}

func (self *controller) cameraTriggerShake(fadeIn, duration, fadeOut TicksDuration) {
	if self.inDraw { panic("can't trigger shake during draw stage") }
	self.cameraStartShake(fadeIn)
	self.shakeDuration = duration
	self.shakeFadeOut  = fadeOut // TODO: maybe triggered shakes shouldn't stop pre-existing continuous shakes?
}

func (self *controller) cameraIsShaking() bool {
	if self.shakeElapsed == 0 {
		return self.shakeFadeIn > 0 || self.shakeDuration > 0
	} else {
		if self.shakeElapsed < self.shakeDuration { return true }
		return self.shakeElapsed < (self.shakeFadeIn + self.shakeDuration + self.shakeFadeOut)
	}
}

func (self *controller) getShakeActivity() float64 {
	if self.shakeElapsed == 0 { return 0 }
	if self.shakeElapsed < self.shakeFadeIn {
		return float64(self.shakeElapsed)/float64(self.shakeFadeIn)
	} else {
		elapsed := self.shakeElapsed - self.shakeFadeIn
		if elapsed <= self.shakeDuration { return 1.0 } // shake in progress
		elapsed -= self.shakeDuration
		if elapsed >= self.shakeFadeOut { return 0.0 }
		return 1.0 - float64(elapsed)/float64(self.shakeFadeOut)
	}
}
