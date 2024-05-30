package internal

import "cmp"
import "image"
import "image/color"

import "github.com/hajimehoshi/ebiten/v2"

func Clamp[T cmp.Ordered](x, a, b T) T {
	if x <= a { return a }
	if x >= b { return b }
	return x
}

func ClampTowardsZero[T float64 | float32 | int | int16 | int32 | int64](x, clampReference T) T {
	if clampReference > 0 { return min(x, clampReference) }
	return max(x, clampReference)
}

func Abs[T float64 | float32 | int | int8 | int16 | int32 | int64](x T) T {
	if x >= 0 { return x }
	return -x
}

// It doesn't take zero into account, but this is intentional.
func Sign[T float64 | float32 | int | int8 | int16 | int32 | int64](x T) T {
	if x >= 0 { return +1 }
	return -1
}

// --- color ---

func toRGBAf32(clr color.Color) (r, g, b, a float32) {
	r16, g16, b16, a16 := clr.RGBA()
	return float32(r16)/65535.0, float32(g16)/65535.0, float32(b16)/65535.0, float32(a16)/65535.0
}

// --- interpolation ---

func TAt(x, a, b float64) float64 {
	if a < b {
		if x <  a { return 0.0 }
		if x >= b { return 1.0 }
	} else {
		if x <  b { return 1.0 }
		if x >= a { return 0.0 }
	}
	return (x - a)/(b - a)
}

func LinearInterp(a, b, t float64) float64 { return (a + (b - a)*t) }
func CubicSmoothstepInterp(a, b, t float64) float64 { // related: https://iquilezles.org/articles/smoothsteps
	t = Clamp(t, 0, 1)
	return LinearInterp(a, b, t*t*(3.0 - 2.0*t))
}
func QuadInOutInterp(a, b, t float64) float64 {
	return LinearInterp(a, b, QuadInOut(t))
}
func QuadInOut(t float64) float64 {
	t = Clamp(t, 0, 1)
	if t < 0.5 { return 2*t*t }
	t = 2*t - 1
	return -0.5*(t*(t - 2) - 1)
}

func QuadDvInOut(t float64) float64 {
	t = Clamp(t, 0, 1)
	if t <= 0.5 { return 4*t }
	return 4 - 4*t
}

func EaseInQuad(t float64) float64 {
	t = Clamp(t, 0, 1)
	return t*t
}

func CubicOutInterp(a, b, t float64) float64 {
	return LinearInterp(a, b, EaseOutCubic(t))
}
func EaseOutCubic(t float64) float64 {
	t = Clamp(t, 0, 1)
	omt := 1 - t
	return 1 - omt*omt*omt
}

func EaseOutQuad(t float64) float64 {
	t = Clamp(t, 0, 1)
	omt := 1 - t
	return 1 - omt*omt
}

// --- triangles drawing ---

var pkgMask1x1 *ebiten.Image
var pkgFillVertices []ebiten.Vertex
var pkgFillVertIndices []uint16
var pkgFillTrianglesOpts ebiten.DrawTrianglesOptions

func init() {
	pkgMask1x1 = ebiten.NewImage(1, 1)
	pkgFillVertices = make([]ebiten.Vertex, 4)
	pkgFillVertIndices = []uint16{0, 1, 3, 3, 1, 2}
	for i := range 4 {
		pkgFillVertices[i].SrcX = 0.5
		pkgFillVertices[i].SrcY = 0.5
	}
}

func FillOver(target *ebiten.Image, fillColor color.Color) {
	FillOverRect(target, target.Bounds(), fillColor)
}

func FillOverRect(target *ebiten.Image, bounds image.Rectangle, fillColor color.Color) {
	if bounds.Empty() { return }
	r, g, b, a := toRGBAf32(fillColor)
	for i := range 4 {
		pkgFillVertices[i].ColorR = r
		pkgFillVertices[i].ColorG = g
		pkgFillVertices[i].ColorB = b
		pkgFillVertices[i].ColorA = a
	}

	minX, minY := float32(bounds.Min.X), float32(bounds.Min.Y)
	maxX, maxY := float32(bounds.Max.X), float32(bounds.Max.Y)
	pkgFillVertices[0].DstX = minX
	pkgFillVertices[0].DstY = minY
	pkgFillVertices[1].DstX = maxX
	pkgFillVertices[1].DstY = minY
	pkgFillVertices[2].DstX = maxX
	pkgFillVertices[2].DstY = maxY
	pkgFillVertices[3].DstX = minX
	pkgFillVertices[3].DstY = maxY
	target.DrawTriangles(pkgFillVertices, pkgFillVertIndices, pkgMask1x1, &pkgFillTrianglesOpts)
}
