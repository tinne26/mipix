package mipix

import "fmt"

import "github.com/hajimehoshi/ebiten/v2"

var _ fmt.Formatter

// --- game ---

// Equivalent to [ebiten.Game], but without the Layout() method.
// Pixel art games operate with a logical base resolution set
// through [SetResolution]() instead, and in mipix's case, the
// logical canvas passed to draw is only big enough to comprise
// the relevant area to be drawn.
type Game interface {
	// Updates the game logic
	Update() error

	// Draws the game contents
	Draw(logicalCanvas *ebiten.Image)
}

// Equivalent to [ebiten.RunGame](), but expecting a mipix [Game]
// instead of an [ebiten.Game].
//
// Will panic if invoked before [SetResolution]().
func Run(game Game) error {
	return pkgController.run(game)
}

// --- core ---

// Returns the game's base resolution. See [SetResolution]()
// for more details.
func GetResolution() (width, height int) {
	return pkgController.getResolution()
}

// Sets the game's base resolution. This defines the game's
// aspect ratio and logical canvas size at whole coordinates
// and zoom = 1.0.
func SetResolution(width, height int) {
	pkgController.setResolution(width, height)
}

// Ignore this function unless you are already using [QueueHiResDraw]().
// This function is only relevant when trying to interleave logical
// and high resolution draws.
//
// The canvas passed to the callback will be preemptively cleared if
// the previous draw was a high resolution draw.
//
// Must only be called from [Game].Draw() or successive draw callbacks. 
func QueueDraw(handler func(logicalCanvas *ebiten.Image)) {
	pkgController.queueDraw(handler)
}

// Schedules the given handler to be invoked after the current
// drawing function and any other queued draws finish.
//
// The viewport passed to the handler is the full game screen canvas,
// including any possibly unused borders, while the target is a subimage
// corresponding to the active area of the viewport.
//
// Using this function is necessary if you want to render high resolution
// graphics. This includes vectorial UI, some shader effects and others.
//
// Must only be called from [Game].Draw() or successive draw callbacks.
// See also [QueueDraw](). 
func QueueHiResDraw(handler func(viewport, target *ebiten.Image)) {
	pkgController.queueHiResDraw(handler)
}

// Returns whether a layout change has happened on the current tick.
// Layout changes happen whenever the game window is resized in windowed
// mode, the game switches between windowed and fullscreen modes, or
// the device scale factor changes (possibly due to a monitor change).
//
// This function is relevant if you need to redraw game borders manually
// and efficiently, or if you are only redrawing the screen when
// something changes.
func LayoutHasChanged() bool {
	return pkgController.layoutHasChanged
}

// --- high resolution drawing ---

// See [HiRes]().
type AccessorHiRes struct{}

// Provides access to high resolution drawing methods in
// a structured manner. Use through method chaining, e.g.:
//   mipix.HiRes().Draw()
func HiRes() AccessorHiRes { return AccessorHiRes{} }

// Draws the given source into the target at the given logical coordinates.
// These logical coordinates have the camera origin automatically subtracted.
//
// Notice that mipix's main focus is not high resolution drawing, and this
// method is not expected to be used more than a dozen times per frame.
// If you are only drawing the main character or a few entities at floating
// point positions, using this method should be fine. If you are trying to
// draw every element of your game with this, or relying on this for a
// particle system, you are misusing mipix.
//
// Many more high resolution drawing features could be provided, and some
// might be added in the future, but this is not the main goal of the project.
//
// All that being said, this is not a recommendation to avoid this method.
// This method is perfectly functional and a very practical tool in many
// scenarios.
func (self AccessorHiRes) Draw(target, source *ebiten.Image, x, y float64) {
	pkgController.hiResDraw(target, source, x, y)
}

// Similar to [AccessorHiRes.Draw](), but horizontally flipped.
func (self AccessorHiRes) DrawHorzFlip(target, source *ebiten.Image, x, y float64) {
	pkgController.hiResDrawHorzFlip(target, source, x, y)
}

// --- scaling ---

// See [Scaling]().
type AccessorScaling struct{}

// Provides access to scaling-related functionality in a structured
// manner. Use through method chaining, e.g.:
//   mipix.Scaling().SetFilter(mipix.Hermite)
func Scaling() AccessorScaling { return AccessorScaling{} }

// See [AccessorScaling.SetFilter]().
//
// Multiple filter options are provided mostly as comparison points.
// In general, sticking to [AASamplingSoft] is recommended.
type ScalingFilter uint8
const (
	// Anti-aliased pixel art point sampling. Good default, reasonably
	// performant, decent balance between sharpness and stability during
	// zooms and small movements.
	AASamplingSoft ScalingFilter = iota

	// Like AASamplingSoft, but slightly sharper and slightly less stable
	// during zooms and small movements.
	AASamplingSharp

	// No interpolation. Sharpest and fastest filter, but can lead
	// to distorted geometry. Very unstable, zooming and small movements
	// will be really jumpy and ugly.
	Nearest

	// Slightly blurrier than AASamplingSoft and more unstable than
	// AASamplingSharp. Still provides fairly decent results at
	// reasonable performance.
	Hermite

	// The most expensive filter by quite a lot. Slightly less sharp than
	// Hermite, but quite a bit more stable. Might slightly misrepresent
	// some colors throughout high contrast areas.
	Bicubic

	// Offered mostly for comparison purposes. Slightly blurrier than
	// Hermite, but quite a bit more stable.
	Bilinear

	// Offered for comparison purposes only. Non high-resolution aware
	// scaling filter, more similar to what naive scaling will look like.
	SrcHermite

	// Offered for comparison purposes only. Non high-resolution aware
	// scaling filter, more similar to what naive scaling will look like.
	SrcBicubic

	// Offered for comparison purposes only. Non high-resolution aware
	// scaling filter, more similar to what naive scaling will look like.
	// This is what Ebitengine will do by default with the FilterLinear
	// filter.
	SrcBilinear

	scalingFilterEndSentinel
)

// Returns a string representation of the scaling filter.
func (self ScalingFilter) String() string {
	switch self {
	case AASamplingSoft  : return "AASamplingSoft"
	case AASamplingSharp : return "AASamplingSharp"
	case Nearest  : return "Nearest"
	case Hermite  : return "Hermite"
	case Bicubic  : return "Bicubic"
	case Bilinear : return "Bilinear"
	case SrcHermite  : return "SrcHermite"
	case SrcBicubic  : return "SrcBicubic"
	case SrcBilinear : return "SrcBilinear"
	default:
		panic("invalid ScalingFilter")
	}
}

// Set to true to completely fill the screen no matter how ugly
// it gets. By default, stretching is disabled. In general,
// you only want to expose stretching as a setting for players.
// 
// Must only be called during initialization or [Game].Update().
func (AccessorScaling) SetStretchingAllowed(allowed bool) {
	pkgController.scalingSetStretchingAllowed(allowed)
}

// Returns whether stretching is allowed for screen scaling.
// See [AccessorScaling.SetStretchingAllowed]() for more details.
func (AccessorScaling) GetStretchingAllowed() bool {
	return pkgController.scalingGetStretchingAllowed()
}

// Changes the scaling filter. The default is [AASamplingSoft].
//
// Must only be called during initialization or [Game].Update().
//
// The first time you set a filter explicitly, its shader will also
// be compiled. This means that this function can be effectively used
// to precompile the relevant shaders. Otherwise, the shader will be
// compiled the first time it's needed to draw something.
func (AccessorScaling) SetFilter(filter ScalingFilter) {
	pkgController.scalingSetFilter(filter)
}

// Returns the current scaling filter. The default is [AASamplingSoft].
func (AccessorScaling) GetFilter() ScalingFilter {
	return pkgController.scalingGetFilter()
}

// --- conversions ---

// See [Convert]().
type AccessorConvert struct{}

// Provides access to coordinate conversions in a structured
// manner. Use through method chaining, e.g.:
//   cx, cy := ebiten.CursorPosition()
//   lx, ly := mipix.Convert().ToLogicalCoords(cx, cy)
func Convert() AccessorConvert { return AccessorConvert{} }

// Transforms coordinates obtained from [ebiten.CursorPosition]() and similar
// functions to coordinates within the game's logical space.
func (AccessorConvert) ToLogicalCoords(x, y int) (float64, float64) {
	return pkgController.convertToLogicalCoords(x, y)
}

// Transforms coordinates obtained from [ebiten.CursorPosition]() and similar
// functions to relative coordinates between 0 and 1. 
func (AccessorConvert) ToRelativeCoords(x, y int) (float64, float64) {
	return pkgController.convertToRelativeCoords(x, y)
}

// --- debug ---

// See [Debug]().
type AccessorDebug struct{}

// Provides access to debugging functionality in a structured
// manner. Use through method chaining, e.g.:
//   mipix.Debug().Drawf("current tick: %d", mipix.Ticks().Now())
func Debug() AccessorDebug { return AccessorDebug{} }

// Similar to Printf debugging, but drawing the text on the top
// left of the screen instead. Multi-line text is not supported,
// use multiple Drawf commands in sequence instead.
//
// You can call this function at any point, even during [Game].Update.
// Strings will be queued and rendered at the end of the next draw.
func (AccessorDebug) Drawf(format string, args ...any) {
	pkgController.debugDrawf(format, args...)
}

// Similar to [fmt.Printf](), but expects two tick counts as the first
// arguments. The function will only print during the period elapsed
// between those two tick counts.
// Some examples:
//   mipix.Debug().Printfr(0, 0, "only print on the first tick\n")
//   mipix.Debug().Printfr(180, 300, "print from 3s to 5s lapse\n")
func (AccessorDebug) Printfr(firstTick, lastTick uint64, format string, args ...any) {
	pkgController.debugPrintfr(firstTick, lastTick, format, args...)
}

// Similar to [fmt.Printf](), but only prints every N ticks, where N
// is given as 'everyNTicks'. For example, in most games, using 
// N = 60 will lead to print once every 60 ticks. 61 is prime.
func (AccessorDebug) Printfe(everyNTicks uint64, format string, args ...any) {
	pkgController.debugPrintfe(everyNTicks, format, args...)
}

// Similar to [fmt.Printf](), but only prints if the given key is pressed.
// Common keys: [ebiten.KeyShiftLeft], [ebiten.KeyControl], [ebiten.KeyDigit1].
func (AccessorDebug) Printfk(key ebiten.Key, format string, args ...any) {
	pkgController.debugPrintfk(key, format, args...)
}

// --- ticks ---

// See [Tick]().
type AccessorTick struct{}

// Provides access to game tick functions in a structured
// manner. Use through method chaining, e.g.:
//   currentTick := mipix.Tick().Now()
func Tick() AccessorTick { return AccessorTick{} }

// Returns the current tick.
func (AccessorTick) Now() uint64 {
	return pkgController.tickNow()
}

// An advanced mechanism to improve support for high refresh
// rate displays.
//
// For context, the basic idea is to consider the game clock to
// always run at 240 ticks per second. This can be simulated
// perfectly through [ebiten.SetTPS](240), but it can also be
// reproduced with 120TPS (by advancing the internal clock by 2
// ticks on each update instead of 1), or 60TPS (advance by 4
// instead of 1). With some care, this allows games to preserve
// a perfectly determinist logic while providing players with 
// high refresh rate displays a smoother or more responsive
// experience.
//
// This is not necessary, relevant or appropriate for every game.
// This is not necessary, relevant or appropriate for every player.
func (AccessorTick) SetRate(tickRate int) {
	pkgController.tickSetRate(tickRate)
}

// Returns the current tick rate. Defaults to 1.
// See [AccessorTick.SetRate]() for more context.
func (AccessorTick) GetRate() int {
	return pkgController.tickGetRate()
}
