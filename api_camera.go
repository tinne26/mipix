package mipix

import "image"

import "github.com/tinne26/mipix/zoomer"
import "github.com/tinne26/mipix/tracker"
import "github.com/tinne26/mipix/shaker"

// See [Camera]().
type AccessorCamera struct{}

// Provides access to camera-related functionality in a structured
// manner. Use through method chaining, e.g.:
//   mipix.Camera().Zoom(2.0)
func Camera() AccessorCamera { return AccessorCamera{} } 

// --- tracking ---

// Returns the current tracker. See [AccessorCamera.SetTracker]()
// for more details.
func (AccessorCamera) GetTracker() tracker.Tracker {
	return pkgController.cameraGetTracker()
}

// Sets the tracker in charge of updating the camera position.
// By default the tracker is nil, and tracking is handled
// by a fallback [trackr.LinearTracker].
func (AccessorCamera) SetTracker(tracker tracker.Tracker) {
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
// since the camera might be zoomed or shaking.
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

// Sets a new target zoom level. The transition from the current
// zoom level to the new one is managed by a [zoomer.Zoomer].
func (AccessorCamera) Zoom(newZoomLevel float64) {
	pkgController.cameraZoom(newZoomLevel)
}

// Returns the current [zoomer.Zoomer] interface.
// See [AccessorCamera.SetZoomer]() for more details.
func (AccessorCamera) GetZoomer() zoomer.Zoomer {
	return pkgController.cameraGetZoomer()
}

// Sets the [zoomer.Zoomer] in charge of updating camera zoom levels.
// By default the zoomer is nil, and zoom levels are handled
// by a fallback [SimpleZoomer].
func (AccessorCamera) SetZoomer(zoomer zoomer.Zoomer) {
	pkgController.cameraSetZoomer(zoomer)
}

// Returns the current and target zoom levels.
func (AccessorCamera) GetZoom() (current, target float64) {
	return pkgController.cameraGetZoom()
}

// --- screen shaking ---

// Returns the current screen shaker interface.
// See [AccessorCamera.SetShaker]() for more details.
func (AccessorCamera) GetShaker() shaker.Shaker {
	return pkgController.cameraGetShaker()
}

// Sets a shaker. By default the screen shaker interface is
// nil, and shakes are handled by a fallback [shaker.SimpleShaker].
func (AccessorCamera) SetShaker(shaker shaker.Shaker) {
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
