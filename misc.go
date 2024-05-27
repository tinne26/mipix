package mipix

import "image"

import "github.com/hajimehoshi/ebiten/v2"

import "github.com/tinne26/mipix/internal"

// Helper type used for zoom transitions and shakes.
type TicksDuration = internal.TicksDuration

const ZeroTicks TicksDuration = 0

// Utility method, syntax sugar for [ebiten.Image.SubImage]().
func SubImage(source *ebiten.Image, minX, minY, maxX, maxY int) *ebiten.Image {
	return source.SubImage(image.Rect(minX, minY, maxX, maxY)).(*ebiten.Image)
}

// Quick alias to the control key for use with [AccessorDebug.Printfk]().
const Ctrl = ebiten.KeyControl

// internal usage
const maxUint32 = 0xFFFF_FFFF
