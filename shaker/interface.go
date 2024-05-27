// This package defines a [Shaker] interface that the mipix
// camera can use to perform screen shakes and provides
// a few default implementations.
//
// All provided implementations respect a few properties:
//  - Resolution independent: range of motion for the shakes
//    is not hardcoded, but proportional to the game's resolution.
//  - Update-rate independent: as long as the total ticks per
//    second remain the same, different Tick().UPS() values will
//    still reproduce the same results. See [ups-vs-tps] if you
//    need more context.
// These are nice properties to have for public implementations,
// but if you write your own, remember that most often these properties
// won't be relevant to you. Make your life easier and ignore them if
// you are only getting started.
//
// [ups-vs-tps]: https://github.com/tinne26/mipix/blob/main/docs/ups-vs-tps.md
package shaker

import "github.com/tinne26/mipix/internal"

// The interface for mipix screen shakers.
//
// Given a level that transitions linearly between 0 and 1
// during the fade in and fade out stages, GetShakeOffsets()
// returns the logical offsets for the camera.
//
// Minor detail: all built-in implementations happen to normalize
// the fade in/out level with a cubic smoothstep, just to make
// things nicer.
type Shaker interface {
	GetShakeOffsets(level float64) (float64, float64)
}

// Alias equivalent to mipix.TicksDuration.
type TicksDuration = internal.TicksDuration
