package mipix

import "math"
import "image"

import "github.com/tinne26/mipix/internal"
import "github.com/tinne26/mipix/zoomer"
import "github.com/tinne26/mipix/tracker"
import "github.com/tinne26/mipix/shaker"

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

func (self *controller) cameraGetTracker() tracker.Tracker {
	return self.tracker
}

func (self *controller) cameraSetTracker(tracker tracker.Tracker) {
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
	if self.redrawManaged && (x != self.trackerCurrentX || y != self.trackerCurrentY) {
		self.needsRedraw = true
	}
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
	var camTracker tracker.Tracker = self.tracker
	if camTracker == nil { camTracker = tracker.LinearTracker }
	self.trackerPrevSpeedX, self.trackerPrevSpeedY = camTracker.Update(
		self.trackerCurrentX, self.trackerCurrentY,
		self.trackerTargetX, self.trackerTargetY,
		self.trackerPrevSpeedX, self.trackerPrevSpeedY,
	)
	self.trackerCurrentX += self.trackerPrevSpeedX
	self.trackerCurrentY += self.trackerPrevSpeedY
	
	if self.redrawManaged && (self.trackerPrevSpeedX != 0 || self.trackerPrevSpeedY != 0) {
		self.needsRedraw = true
	}
}

// --- zoom ---

func (self *controller) updateZoom() {
	zoomer := self.cameraGetInternalZoomer()
	change := zoomer.Update(self.zoomCurrent, self.zoomTarget)
	if math.IsNaN(change) { panic("zoomer returned NaN") }
	self.zoomCurrent += change
	internal.CurrentZoom = self.zoomCurrent
	if self.zoomCurrent < 0.005 || self.zoomCurrent > 500.0 {
		panic("something is wrong with the zoomer: after last update, zoom went outside [0.005, 500.0]")
	}
	
	if self.redrawManaged && change != 0 {
		self.needsRedraw = true
	}
}

func (self *controller) cameraGetInternalZoomer() zoomer.Zoomer {
	if self.zoomer != nil { return self.zoomer }
	if defaultZoomer == nil {
		defaultZoomer = &zoomer.Quadratic{}
		defaultZoomer.Reset()
	}
	return defaultZoomer
}

func (self *controller) updateShake() {
	if !self.cameraIsShaking() { return }
	var camShaker shaker.Shaker = self.shaker
	if camShaker == nil {
		if defaultShaker == nil {
			defaultShaker = &shaker.Random{}
		}
		camShaker = defaultShaker
	}
	activity := self.getShakeActivity()
	shakeX, shakeY := camShaker.GetShakeOffsets(activity)
	self.shakeElapsed += TicksDuration(self.tickRate)
	if self.redrawManaged && (shakeX != self.shakeOffsetX || shakeY != self.shakeOffsetY) {
		self.needsRedraw = true
	}
	self.shakeOffsetX, self.shakeOffsetY = shakeX, shakeY
}

func (self *controller) cameraZoom(newZoomLevel float64) {
	if self.inDraw { panic("can't zoom during draw stage") }
	self.zoomTarget = newZoomLevel
}

func (self *controller) cameraZoomReset(zoomLevel float64) {
	if self.inDraw { panic("can't reset zoom during draw stage") }
	self.zoomCurrent, self.zoomTarget, internal.CurrentZoom = zoomLevel, zoomLevel, zoomLevel
	self.cameraGetInternalZoomer().Reset()
}

func (self *controller) cameraGetZoomer() zoomer.Zoomer {
	return self.zoomer
}

func (self *controller) cameraSetZoomer(zoomer zoomer.Zoomer) {
	if self.inDraw { panic("can't change zoomer during draw stage") }
	self.zoomer = zoomer
}

func (self *controller) cameraGetZoom() (current, target float64) {
	return self.zoomCurrent, self.zoomTarget
}

// ---- screenshake ----

func (self *controller) cameraSetShaker(shaker shaker.Shaker) {
	if self.inDraw { panic("can't set shaker during draw stage") }
	self.shaker = shaker
}

func (self *controller) cameraGetShaker() shaker.Shaker {
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
