package mipix

// See [Redraw]().
type AccessorRedraw struct{}

// Provides access to methods for efficient GPU usage in
// a structured manner. Use through method chaining, e.g.:
//   mipix.Redraw().SetManaged(true)
//
// In some games and applications it's possible to spare
// GPU by using [ebiten.SetScreenClearedEveryFrame](false)
// and omitting redundant draw calls.
//
// The redraw accessor allows you to synchronize this
// process with mipix itself, as there are some projections
// that would otherwise fall outside your control.
//
// By default, redraws are executed on every frame. If you
// want to manage them more efficiently, you can do the
// following:
//  - Make sure to disable ebitengine's screen clear.
//  - Opt into managed redraws with [AccessorRedraw.SetManaged](true).
//  - Whenever a redraw becomes necessary, issue an
//    [AccessorRedraw.Request]().
//  - On [Game].Draw(), if ![AccessorRedraw.Pending](), skip the draw.
func Redraw() AccessorRedraw {
	return AccessorRedraw{}
}

// Enables or disables manual redraw management. By default,
// redraw management is disabled and the screen is redrawn
// every frame.
//
// Must only be called during initialization or [Game].Update().
func (AccessorRedraw) SetManaged(managed bool) {
	pkgController.redrawSetManaged(managed)
}

// Returns whether manual redraw management is enabled or not.
func (AccessorRedraw) IsManaged() bool {
	return pkgController.redrawIsManaged()
}

// Notifies mipix that the next [Game].Draw() needs to be
// projected to the screen. Requests are typically issued
// when relevant input or events are detected during
// [Game].Update().
//
// Zoom and camera changes are also auto-detected.
//
// This function can be called multiple times within a single
// update, it's only doing the equivalent of "needs redraw = true".
func (AccessorRedraw) Request() {
	pkgController.redrawRequest()
}

// Returns whether a redraw is still pending. Notice that
// besides explicit requests, a redraw can also be pending
// due to a canvas resize, the modification of the scaling
// properties, etc.
//
// You would typically use this method right at the start
// [Game].Draw(), returning early if !mipix.Redraw().Pending().
func (AccessorRedraw) Pending() bool {
	return pkgController.redrawPending()
}

// Signal the redraw manager to clear both the logical screen
// and the high resolution canvas before the next [Game].Draw().
func (AccessorRedraw) ScheduleClear() {
	pkgController.redrawScheduleClear()
}
