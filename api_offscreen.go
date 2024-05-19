package mipix

import "image"
import "image/color"

import "github.com/hajimehoshi/ebiten/v2"

type Offscreen struct {
	canvas *ebiten.Image
	width int
	height int
	drawImageOpts ebiten.DrawImageOptions
}

func NewOffscreen(width, height int) *Offscreen {
	return &Offscreen{
		canvas: ebiten.NewImage(width, height),
		width: width, height: height,
	}
}

func (self *Offscreen) Target() *ebiten.Image {
	return self.canvas
}

func (self *Offscreen) Size() (width, height int) {
	return self.width, self.height
}

// Equivalent to ebitengine's image draw.
func (self *Offscreen) Draw(source *ebiten.Image, opts *ebiten.DrawImageOptions) {
	self.canvas.DrawImage(source, opts)
}

// Simpler version of [Offscreen.Draw]().
func (self *Offscreen) DrawAt(source *ebiten.Image, x, y int) {
	self.drawImageOpts.GeoM.Translate(float64(x), float64(y))
	self.canvas.DrawImage(source, &self.drawImageOpts)
	self.drawImageOpts.GeoM.Reset()
}

// Similar to ebitengine's image.Fill(), but with BlendSourceOver instead of BlendCopy.
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
// canvas. 
func (self *Offscreen) Project(target *ebiten.Image) {
	pkgController.project(self.canvas, target)
}
