// This package defines a [Zoomer] interface that the mipix
// camera can use to update its position, and provides a few
// default implementations.
//
// All provided implementations respect a few properties:
//  - Resolution independent: range of motion for the tracking
//    is not hardcoded, but proportional to the game's resolution.
//  - Update-rate independent: as long as the total ticks per
//    second remain the same, different Tick().UPS() values will
//    still reproduce the same results. See [ups-vs-tps] if you
//    need more context. Many implementations are actually also
//    tick-rate independent.
// These are nice properties for public implementations, but if you
// are writing your own, remember that most often these properties
// won't be relevant to you. You can ignore them and make your life
// easier if you are only getting started.
//
// Warning: avoid bringing cameras to 0.1 and similarly low zoom
// levels. At those levels, zoomer bounciness and overshoot can make
// your game collapse very easily. It's better to be at x3.0 zoom by
// default most of the time than going to super low values that
// might be unstable and dangerous to work with. There are many
// sources of unstability, like different update/tick rate
// configurations, changing zooms mid-transition and so on.
// Always strive to operate either with generous safety margins
// or very safe zoomers without bounciness.
//
// [ups-vs-tps]: https://github.com/tinne26/mipix/blob/main/docs/ups-vs-tps.md
package zoomer

import "github.com/tinne26/mipix/internal"

// The interface for mipix camera zooming.
// 
// Given current and target zoom levels, the Update() method
// returns the zoom change for a single update. Reset()
// is used to indicate an instantaneous zoom level reset
// instead.
type Zoomer interface {
	Reset()
	Update(currentZoom, targetZoom float64) (change float64)
}

// Alias for mipix.TicksDuration.
type TicksDuration = internal.TicksDuration
