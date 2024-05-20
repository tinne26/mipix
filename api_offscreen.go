package mipix

import "image"
import "image/color"

import "github.com/hajimehoshi/ebiten/v2"

// Offscreens are logically sized canvases that you can draw on
// and later project to a high resolution screen. By default,
// your game world is not drawn on manual offscreens, but rather
// the canvas received through [Game].Draw(). What offscreens can
// be useful for is drawing pixel-perfect UI and other
// camera-independent elements of your game.
//
// Creating an offscreen involves creating an [*ebiten.Image], so
// you want to store and reuse them. They also have to be manually
// cleared when needed.
type Offscreen struct {
	canvas *ebiten.Image
	width int
	height int
	drawImageOpts ebiten.DrawImageOptions
}

// Creates a new offscreen with the given logical size.
//
// Never do this per frame, always reuse the offscreens.
func NewOffscreen(width, height int) *Offscreen {
	return &Offscreen{
		canvas: ebiten.NewImage(width, height),
		width: width, height: height,
	}
}

// Returns the underlying canvas.
func (self *Offscreen) Target() *ebiten.Image {
	return self.canvas
}

// Returns the size of the offscreen.
func (self *Offscreen) Size() (width, height int) {
	return self.width, self.height
}

// Equivalent to [ebiten.Image.DrawImage]().
func (self *Offscreen) Draw(source *ebiten.Image, opts *ebiten.DrawImageOptions) {
	self.canvas.DrawImage(source, opts)
}

// Handy version of [Offscreen.Draw]().
func (self *Offscreen) DrawAt(source *ebiten.Image, x, y int) {
	self.drawImageOpts.GeoM.Translate(float64(x), float64(y))
	self.canvas.DrawImage(source, &self.drawImageOpts)
	self.drawImageOpts.GeoM.Reset()
}

// Similar to [ebiten.Image.Fill](), but with BlendSourceOver instead of BlendCopy.
func (self *Offscreen) Coat(fillColor color.Color) {
	fillOverRect(self.canvas, self.canvas.Bounds(), fillColor)
}

// Similar to [Offscreen.Coat](), but restricted to a specific rectangular area.
func (self *Offscreen) CoatRect(bounds image.Rectangle, fillColor color.Color) {
	fillOverRect(self.canvas, bounds, fillColor)
}

// Clears the underlying canvas.
func (self *Offscreen) Clear() {
	self.canvas.Clear()
}

// Projects the offscreen into the given target. In almost all
// cases, you want the target to be the active high resolution
// target (the second argument of a [QueueHiResDraw]() handler).
func (self *Offscreen) Project(target *ebiten.Image) {
	pkgController.project(self.canvas, target)
}
