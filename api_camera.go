package mipix

import "image"

// See [Camera]().
type AccessorCamera struct{}

// Provides access to camera-related functionality in a structured
// manner. Use through method chaining, e.g.:
//   mipix.Camera().Zoom(2.0, mipix.TicksDuration(30))
func Camera() AccessorCamera { return AccessorCamera{} } 

// --- tracking ---

// Trackers are an interface used for updating the camera position.
// Given current and target coordinates, a tracker must return
// the advance for a single update.
//
// Related to [AccessorCamera.SetTracker]().
type Tracker interface {
	Update(currentX, currentY, targetX, targetY, prevSpeedX, prevSpeedY float64) (float64, float64)
}

// Returns the current tracker. See [AccessorCamera.SetTracker]()
// for more details.
func (AccessorCamera) GetTracker() Tracker {
	return pkgController.cameraGetTracker()
}

// Sets the tracker in charge of updating the camera position.
// By default the tracker is nil, and tracking is handled
// by a fallback [InstantTracker].
func (AccessorCamera) SetTracker(tracker Tracker) {
	pkgController.cameraSetTracker(tracker)
}

// Feeds the camera the latest target coordinates to point
// to. The camera might take a while to reach them, depending
// on the current [Tracker] behavior.
//
// You can pass coordinates as many times as you want, the
// target position is always the most recent pair.
func (AccessorCamera) NotifyCoordinates(x, y float64) {
	pkgController.cameraNotifyCoordinates(x, y)
}

// Immediately resets the camera coordinates.
// Commonly used when changing scenes or maps.
func (AccessorCamera) ResetCoordinates(x, y float64) {
	pkgController.cameraResetCoordinates(x, y)
}

// This method allows updating the [AccessorCamera.Area]()
// even during [Game].Update(). By default, this happens
// automatically after [Game].Update(), but flushing the
// coordinates can force an earlier update.
//
// Notice that only one camera update can happen per tick,
// so the automatic camera update will be skipped if you
// flush coordinates manually during [Game].Update(). 
// Calling this method multiple times during the same update 
// will only update coordinates on the first invocation.
//
// If you don't need this feature, it's better to forget about
// this method. This is only necessary if you need the camera
// area to remain consistent during update and draw(s), in which
// case you update the player position first, then notify the
// coordinates and finally flush them.
func (AccessorCamera) FlushCoordinates() {
	pkgController.cameraFlushCoordinates()
}

// Returns the logical area of the game that has to be
// rendered on [Game].Draw()'s canvas or successive logical
// draws. Notice that this can change after each [Game].Update(),
// since the camera might be zoomed or shaked.
//
// Notice that the area will typically be slightly different
// between [Game].Update() and [Game].Draw(). If you need more
// manual control over that, see [AccessorCamera.FlushCoordinates]().
func (AccessorCamera) Area() image.Rectangle {
	return pkgController.cameraAreaGet()
}

// Similar to [AccessorCamera.Area](), but without rounding up
// the coordinates and returning the exact values. This is rarely
// necessary in practice outside debugging.
func (AccessorCamera) AreaF64() (minX, minY, maxX, maxY float64) {
	return pkgController.cameraAreaF64()
}

// --- zoom ---

// Begins a transition from the current zoom level to the new given
// value. Common zoom values range between 0.5 and 3. For immediate
// transitions, use [ZeroTicks].
func (AccessorCamera) Zoom(toZoomLevel float64, transition TicksDuration) {
	pkgController.cameraZoom(toZoomLevel, transition)
}

// Same as [AccessorCamera.Zoom](), but the given transition is not absolute,
// but given as if you were starting the zoom from the 'ifCurrent' zoom level.
// For example, say we want a transition from x1.0 to x2.0 in 60 ticks. If
// we are already at x1.5 and we want it to take the proportional 30 ticks
// instead, we can use ZoomFrom(1.0, 2.0, 60) to handle that automatically.
func (AccessorCamera) ZoomFrom(ifCurrent, toZoomLevel float64, transition TicksDuration) {
	pkgController.cameraZoomFrom(ifCurrent, toZoomLevel, transition)
}

// Returns the current and target zoom levels.
func (AccessorCamera) GetZoom() (current, target float64) {
	return pkgController.cameraGetZoom()
}

// TODO: interpolator for camera would be better, without so many problems
// for smoothing and so on. There are too many different use-cases and
// in-between switching and so on.

// When zoom levels are changed in the middle of another zoom transition,
// the sudden swing can result unpleasant. This is why the camera has
// zoom swing smoothing enabled by default, making the current zoom
// transition reduce speed quickly before turning towards the other
// direction. This can add some extra ticks to zoom transitions, but
// generally leads to more natural results.
//
// This function allows you to turn off/on this feature.
// func (AccessorCamera) SmoothZoomSwings(smooth bool) {
// 	pkgController.cameraSmoothZoomSwings(smooth)
// }

// --- screen shaking ---

// Shakers are an interface used to implement screen shakes.
// Given a level that transitions linearly between 0 and 1
// during the fade in and fade out stages, GetShakeOffsets()
// returns the logical offsets for the camera.
// 
// Related to [AccessorCamera.SetShaker](). See [SimpleShaker]
// for a default implementation.
type Shaker interface {
	GetShakeOffsets(level float64) (float64, float64)
}

// Returns the current screen shaker interface.
// See [AccessorCamera.SetShaker]() for more details.
func (AccessorCamera) GetShaker() Shaker {
	return pkgController.cameraGetShaker()
}

// Sets a shaker. By default the screen shaker interface is
// nil, and shakes are handled by a fallback [SimpleShaker].
func (AccessorCamera) SetShaker(shaker Shaker) {
	pkgController.cameraSetShaker(shaker)
}

// Starts a screen shake. The screen will continue shaking
// indefinitely; you must use [AccessorCamera.EndShake]()
// to stop it again.
func (AccessorCamera) StartShake(fadeIn TicksDuration) {
	pkgController.cameraStartShake(fadeIn)
}

// Stop a screen shake. This can even be used to fade out
// triggered shakes early, or to ensure that no shakes
// remain active after screen transitions or others.
func (AccessorCamera) EndShake(fadeOut TicksDuration) {
	pkgController.cameraEndShake(fadeOut)
}

// Returns whether any screen shaking is happening.
func (AccessorCamera) IsShaking() bool {
	return pkgController.cameraIsShaking()
}

// Triggers a screenshake with specific fade in, duration and
// fade out tick durations.
//
// TODO: mention if this overrides previously active shakes,
// or resets anything or whatever.
func (AccessorCamera) TriggerShake(fadeIn, duration, fadeOut TicksDuration) {
	pkgController.cameraTriggerShake(fadeIn, duration, fadeOut)
}
