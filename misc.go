package mipix

import "cmp"
import "image"
import "image/color"

import "github.com/hajimehoshi/ebiten/v2"

// Helper type used for zoom transitions and others.
type TicksDuration uint32

const ZeroTicks TicksDuration = 0

// Utility method, syntax sugar for [ebiten.Image].SubImage().
func SubImage(source *ebiten.Image, minX, minY, maxX, maxY int) *ebiten.Image {
	return source.SubImage(image.Rect(minX, minY, maxX, maxY)).(*ebiten.Image)
}

const Ctrl = ebiten.KeyControl

// --- private use ---

const maxUint32 = 0xFFFF_FFFF

func clamp[T cmp.Ordered](x, a, b T) T {
	if x <= a { return a }
	if x >= b { return b }
	return x
}

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

func fillOver(target *ebiten.Image, fillColor color.Color) {
	fillOverRect(target, target.Bounds(), fillColor)
}

func fillOverRect(target *ebiten.Image, bounds image.Rectangle, fillColor color.Color) {
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

func toRGBAf32(clr color.Color) (r, g, b, a float32) {
	r16, g16, b16, a16 := clr.RGBA()
	return float32(r16)/65535.0, float32(g16)/65535.0, float32(b16)/65535.0, float32(a16)/65535.0
}
