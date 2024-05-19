package mipix

import "math"

import "github.com/hajimehoshi/ebiten/v2"

// project from a logical canvas to a high resolution one
func (self *controller) project(from, to *ebiten.Image) {
	if !self.inDraw { panic("can't project images outside draw stage") }

	// compile shader if necessary
	if self.shaders[self.scalingFilter] == nil {
		self.compileShader(self.scalingFilter)
	}

	// set up vertices
	dstBounds := to.Bounds()
	self.shaderVertices[0].DstX = float32(dstBounds.Min.X)
	self.shaderVertices[0].DstY = float32(dstBounds.Min.Y)
	self.shaderVertices[1].DstX = float32(dstBounds.Max.X)
	self.shaderVertices[1].DstY = self.shaderVertices[0].DstY
	self.shaderVertices[2].DstX = self.shaderVertices[1].DstX
	self.shaderVertices[2].DstY = float32(dstBounds.Max.Y)
	self.shaderVertices[3].DstX = self.shaderVertices[0].DstX
	self.shaderVertices[3].DstY = self.shaderVertices[2].DstY

	srcBounds := from.Bounds()
	self.shaderVertices[0].SrcX = float32(srcBounds.Min.X)
	self.shaderVertices[0].SrcY = float32(srcBounds.Min.Y)
	self.shaderVertices[1].SrcX = float32(srcBounds.Max.X)
	self.shaderVertices[1].SrcY = self.shaderVertices[0].SrcY
	self.shaderVertices[2].SrcX = self.shaderVertices[1].SrcX
	self.shaderVertices[2].SrcY = float32(srcBounds.Max.Y)
	self.shaderVertices[3].SrcX = self.shaderVertices[0].SrcX
	self.shaderVertices[3].SrcY = self.shaderVertices[2].SrcY

	self.shaderOpts.Images[0] = from
	self.shaderOpts.Uniforms["SourceRelativeTextureUnitX"] = float32(srcBounds.Dx())/float32(dstBounds.Dx())
	self.shaderOpts.Uniforms["SourceRelativeTextureUnitY"] = float32(srcBounds.Dy())/float32(dstBounds.Dy())
	to.DrawTrianglesShader(
		self.shaderVertices, self.shaderVertIndices,
		self.shaders[self.scalingFilter], &self.shaderOpts,
	)
	self.shaderOpts.Images[0] = nil
}

func (self *controller) projectLogical(from, to *ebiten.Image) {
	if !self.inDraw { panic("can't project images outside draw stage") }

	// compile shader if necessary
	if self.shaders[self.scalingFilter] == nil {
		self.compileShader(self.scalingFilter)
	}

	// set up vertices
	dstBounds := to.Bounds()
	self.shaderVertices[0].DstX = float32(dstBounds.Min.X)
	self.shaderVertices[0].DstY = float32(dstBounds.Min.Y)
	self.shaderVertices[1].DstX = float32(dstBounds.Max.X)
	self.shaderVertices[1].DstY = self.shaderVertices[0].DstY
	self.shaderVertices[2].DstX = self.shaderVertices[1].DstX
	self.shaderVertices[2].DstY = float32(dstBounds.Max.Y)
	self.shaderVertices[3].DstX = self.shaderVertices[0].DstX
	self.shaderVertices[3].DstY = self.shaderVertices[2].DstY

	cminX, cminY, cmaxX, cmaxY := self.cameraAreaF64()
	fractCamMinX := cminX - math.Floor(cminX)
	fractCamMinY := cminY - math.Floor(cminY)
	fractCamMaxX := cmaxX - math.Floor(cmaxX)
	fractCamMaxY := cmaxY - math.Floor(cmaxY)
	if fractCamMaxX != 0.0 { fractCamMaxX = 1.0 - fractCamMaxX }
	if fractCamMaxY != 0.0 { fractCamMaxY = 1.0 - fractCamMaxY }

	srcBounds := from.Bounds()
	self.shaderVertices[0].SrcX = float32(float64(srcBounds.Min.X) + fractCamMinX)
	self.shaderVertices[0].SrcY = float32(float64(srcBounds.Min.Y) + fractCamMinY)
	self.shaderVertices[1].SrcX = float32(float64(srcBounds.Max.X) - fractCamMaxX)
	self.shaderVertices[1].SrcY = self.shaderVertices[0].SrcY
	self.shaderVertices[2].SrcX = self.shaderVertices[1].SrcX
	self.shaderVertices[2].SrcY = float32(float64(srcBounds.Max.Y) - fractCamMaxY)
	self.shaderVertices[3].SrcX = self.shaderVertices[0].SrcX
	self.shaderVertices[3].SrcY = self.shaderVertices[2].SrcY

	self.shaderOpts.Images[0] = from
	self.shaderOpts.Uniforms["SourceRelativeTextureUnitX"] = float32(srcBounds.Dx())/float32(dstBounds.Dx())
	self.shaderOpts.Uniforms["SourceRelativeTextureUnitY"] = float32(srcBounds.Dy())/float32(dstBounds.Dy())
	to.DrawTrianglesShader(
		self.shaderVertices, self.shaderVertIndices,
		self.shaders[self.scalingFilter], &self.shaderOpts,
	)
	self.shaderOpts.Images[0] = nil
}
