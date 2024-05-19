package mipix

import "math"
import "image"

import "github.com/hajimehoshi/ebiten/v2"

var pkgController controller
func init() {
	pkgController.zoomCurrent = 1.0
	pkgController.zoomTarget = 1.0
	pkgController.zoomStart = 1.0
	pkgController.tickRate = 1
	pkgController.lastFlushCoordinatesTick = 0xFFFF_FFFF_FFFF_FFFF
}

type controller struct {
	// core state
	game Game
	queuedDraws []queuedDraw
	reusableCanvas *ebiten.Image // this preserves the highest size requested by resolution or zooms
	logicalWidth  int
	logicalHeight int
	hiResWidth  int
	hiResHeight int
	prevHiResCanvasWidth  int // used to update layoutHasChanged even on unexpected cases *
	prevHiResCanvasHeight int // used to update layoutHasChanged even on unexpected cases
	layoutHasChanged bool
	inDraw bool
	stretchingEnabled bool
	scalingFilter ScalingFilter
	// * https://github.com/hajimehoshi/ebiten/issues/2978
	
	// camera
	lastFlushCoordinatesTick uint64
	cameraArea image.Rectangle
	
	// tracking
	tracker Tracker
	trackerCurrentX float64
	trackerCurrentY float64
	trackerTargetX float64
	trackerTargetY float64
	trackerPrevSpeedX float64
	trackerPrevSpeedY float64

	// zoom
	zoomTransitionElapsed TicksDuration
	zoomTransitionDuration TicksDuration
	zoomCurrent float64
	zoomStart float64
	zoomTarget float64

	// shake
	shaker Shaker
	shakeElapsed TicksDuration
	shakeFadeIn TicksDuration
	shakeDuration TicksDuration
	shakeFadeOut TicksDuration
	shakeOffsetX float64
	shakeOffsetY float64

	// ticks
	currentTick uint64
	tickRate uint64

	// shaders
	shaderOpts ebiten.DrawTrianglesShaderOptions
	shaderVertices []ebiten.Vertex
	shaderVertIndices []uint16
	shaders [scalingFilterEndSentinel]*ebiten.Shader

	// debug
	debugInfo []string
	debugOffscreen *Offscreen
}

// --- ebiten.Game implementation ---

func (self *controller) Update() error {
	self.currentTick += self.tickRate
	err := self.game.Update()
	if err != nil { return err }
	self.cameraFlushCoordinates()
	return nil
}

func (self *controller) Draw(hiResCanvas *ebiten.Image) {
	self.inDraw = true

	// get bounds and update hi res canvas size
	hiResBounds := hiResCanvas.Bounds()
	hiResWidth, hiResHeight := hiResBounds.Dx(), hiResBounds.Dy()
	if hiResWidth != self.prevHiResCanvasWidth || hiResHeight != self.prevHiResCanvasHeight {
		// * TODO: all this is kind of a temporary hack until ebitengine
		//   can guarantee that layout returned sizes and draw received
		//   canvas sizes will match
		self.prevHiResCanvasWidth  = hiResWidth
		self.prevHiResCanvasHeight = hiResHeight
		self.layoutHasChanged = true
	}

	logicalCanvas := self.getLogicalCanvas()
	activeCanvas  := self.getActiveHiResCanvas(hiResCanvas)
	self.game.Draw(logicalCanvas)
	
	var drawIndex int = 0
	var prevDrawWasHiRes bool = false
	for drawIndex < len(self.queuedDraws) {
		if self.queuedDraws[drawIndex].IsHighResolution() {
			if !prevDrawWasHiRes {
				self.projectLogical(logicalCanvas, activeCanvas)
			}
			self.queuedDraws[drawIndex].hiResFunc(hiResCanvas, activeCanvas)
			prevDrawWasHiRes = true
		} else {
			if prevDrawWasHiRes {
				logicalCanvas.Clear()
				prevDrawWasHiRes = false
			}
			self.queuedDraws[drawIndex].logicalFunc(logicalCanvas)
		}
		drawIndex += 1
	}
	self.queuedDraws = self.queuedDraws[ : 0]

	// final projection
	if !prevDrawWasHiRes { self.projectLogical(logicalCanvas, activeCanvas) }
	self.debugDrawAll(activeCanvas)
	self.inDraw = false
}

func (self *controller) getLogicalCanvas() *ebiten.Image {
	width  := self.cameraArea.Dx()
	height := self.cameraArea.Dy()

	if self.reusableCanvas == nil {
		self.reusableCanvas = ebiten.NewImage(width, height)
		return self.reusableCanvas
	} else {
		bounds := self.reusableCanvas.Bounds()
		availableWidth, availableHeight := bounds.Dx(), bounds.Dy()
		if width == availableWidth && height == availableHeight {
			return self.reusableCanvas
		} else if width <= availableWidth && height <= availableHeight {
			rect := image.Rect(0, 0, width, height)
			canvas := self.reusableCanvas.SubImage(rect).(*ebiten.Image)
			if ebiten.IsScreenClearedEveryFrame() { canvas.Clear() } // TODO: is this the best place to do it?
			return canvas
		} else { // insufficient width or height
			self.reusableCanvas = ebiten.NewImage(width, height)
			return self.reusableCanvas
		}
	}
}

func (self *controller) getActiveHiResCanvas(hiResCanvas *ebiten.Image) *ebiten.Image {
	// trivial case if stretching is used
	if self.stretchingEnabled { return hiResCanvas }

	// crop margins based on aspect ratios
	hiBounds := hiResCanvas.Bounds()
	hiWidth, hiHeight := hiBounds.Dx(), hiBounds.Dy()
	hiAspectRatio := float64(hiWidth)/float64(hiHeight)
	loAspectRatio := float64(self.logicalWidth)/float64(self.logicalHeight)

	switch {
	case hiAspectRatio == loAspectRatio: // just scaling
		return hiResCanvas
	case hiAspectRatio  > loAspectRatio: // horz margins
		xMargin := int((float64(hiWidth) - loAspectRatio*float64(hiHeight))/2.0)
		return SubImage(hiResCanvas, xMargin, 0, hiWidth - xMargin, hiHeight)
	case loAspectRatio  > hiAspectRatio: // vert margins
		yMargin := int((float64(hiHeight) - float64(hiWidth)/loAspectRatio)/2.0)
		return SubImage(hiResCanvas, 0, yMargin, hiWidth, hiHeight - yMargin)
	default:
		panic("unreachable")
	}
}

func (self *controller) Layout(logicWinWidth, logicWinHeight int) (int, int) {
	monitor := ebiten.Monitor()
	scale := monitor.DeviceScaleFactor()
	hiResWidth  := int(float64(logicWinWidth)*scale)
	hiResHeight := int(float64(logicWinHeight)*scale)
	self.layoutHasChanged = false
	if hiResWidth != self.hiResWidth || hiResHeight != self.hiResHeight {
		self.layoutHasChanged = true
		self.hiResWidth, self.hiResHeight = hiResWidth, hiResHeight
	}
	return self.hiResWidth, self.hiResHeight
}

func (self *controller) LayoutF(logicWinWidth, logicWinHeight float64) (float64, float64) {
	monitor := ebiten.Monitor()
	scale := monitor.DeviceScaleFactor()
	outWidth  := math.Ceil(logicWinWidth*scale)
	outHeight := math.Ceil(logicWinHeight*scale)
	self.layoutHasChanged = false
	if int(outWidth) != self.hiResWidth || int(outHeight) != self.hiResHeight {
		self.layoutHasChanged = true
		self.hiResWidth, self.hiResHeight = int(outWidth), int(outHeight)
	}
	return outWidth, outHeight
}

// --- run and queued draws ---

func (self *controller) run(game Game) error {
	self.game = game
	if self.logicalWidth == 0 || self.logicalHeight == 0 {
		panic("must set the game resolution with mipix.SetResolution(width, height) before mipix.Run()")
	}
	self.trackerCurrentX = self.trackerTargetX
	self.trackerCurrentY = self.trackerTargetY
	return ebiten.RunGame(self)
}

// --- resolution ---

func (self *controller) getResolution() (width, height int) {
	return self.logicalWidth, self.logicalHeight
}

func (self *controller) setResolution(width, height int) {
	if self.inDraw { panic("can't change resolution during draw stage") }
	if width < 1 || height < 1 { panic("game resolution must be at least (1, 1)") }
	if width != self.logicalWidth || height != self.logicalHeight {
		self.logicalWidth, self.logicalHeight = width, height
		self.updateCameraArea()
	}
}

// --- scaling ---

func (self *controller) scalingSetFilter(filter ScalingFilter) {
	if self.inDraw { panic("can't change scaling filter during draw stage") }
	self.scalingFilter = filter
	if self.shaders[filter] == nil {
		self.compileShader(filter)
	}
}

func (self *controller) scalingGetFilter() ScalingFilter {
	return self.scalingFilter
}

func (self *controller) scalingSetStretchingAllowed(allowed bool) {
	if self.inDraw { panic("can't change stretching mode during draw stage") }
	self.stretchingEnabled = allowed
}

func (self *controller) scalingGetStretchingAllowed() bool {
	return self.stretchingEnabled
}

// --- hi res ---

func (self *controller) hiResDraw(target, source *ebiten.Image, x, y float64) {
	if !self.inDraw { panic("can't mipix.HiRes().Draw() outside draw stage") }

	// view culling
	camMinX, camMinY, camMaxX, camMaxY := self.cameraAreaF64() // TODO: this is wasteful per draw
	if x > camMaxX || y > camMaxY { return }
	sourceBounds := source.Bounds()
	sourceWidth, sourceHeight := float64(sourceBounds.Dx()), float64(sourceBounds.Dy())
	if x + sourceWidth  < camMinX { return } // outside view
	if y + sourceHeight < camMinY { return } // outside view

	// compile shader if necessary
	if self.shaders[self.scalingFilter] == nil {
		self.compileShader(self.scalingFilter)
	}

	// set triangle vertex coordinates
	targetBounds := target.Bounds()
	targetMinX , targetMinY   := float64(targetBounds.Min.X), float64(targetBounds.Min.Y)
	targetWidth, targetHeight := float64(targetBounds.Dx()), float64(targetBounds.Dy())
	xFactor := self.zoomCurrent*targetWidth/float64(self.logicalWidth)
	yFactor := self.zoomCurrent*targetHeight/float64(self.logicalHeight)
	srcProjMinX := (x - camMinX)*xFactor
	srcProjMinY := (y - camMinY)*yFactor
	srcProjMaxX := srcProjMinX + sourceWidth*xFactor
	srcProjMaxY := srcProjMinY + sourceHeight*yFactor
	self.shaderVertices[0].DstX = float32(targetMinX + srcProjMinX)
	self.shaderVertices[0].DstY = float32(targetMinY + srcProjMinY)
	self.shaderVertices[1].DstX = float32(targetMinX + srcProjMaxX)
	self.shaderVertices[1].DstY = self.shaderVertices[0].DstY
	self.shaderVertices[2].DstX = self.shaderVertices[1].DstX
	self.shaderVertices[2].DstY = float32(targetMinY + srcProjMaxY)
	self.shaderVertices[3].DstX = self.shaderVertices[0].DstX
	self.shaderVertices[3].DstY = self.shaderVertices[2].DstY

	self.shaderVertices[0].SrcX = float32(sourceBounds.Min.X)
	self.shaderVertices[0].SrcY = float32(sourceBounds.Min.Y)
	self.shaderVertices[1].SrcX = float32(sourceBounds.Max.X)
	self.shaderVertices[1].SrcY = self.shaderVertices[0].SrcY
	self.shaderVertices[2].SrcX = self.shaderVertices[1].SrcX
	self.shaderVertices[2].SrcY = float32(sourceBounds.Max.Y)
	self.shaderVertices[3].SrcX = self.shaderVertices[0].SrcX
	self.shaderVertices[3].SrcY = self.shaderVertices[2].SrcY

	self.shaderOpts.Images[0] = source
	self.shaderOpts.Uniforms["SourceRelativeTextureUnitX"] = float32(float64(self.logicalWidth)/targetWidth)
	self.shaderOpts.Uniforms["SourceRelativeTextureUnitY"] = float32(float64(self.logicalHeight)/targetHeight)
	target.DrawTrianglesShader(
		self.shaderVertices, self.shaderVertIndices,
		self.shaders[self.scalingFilter], &self.shaderOpts,
	)
	self.shaderOpts.Images[0] = nil
}

func (self *controller) hiResDrawHorzFlip(target, source *ebiten.Image, x, y float64) {
	if !self.inDraw { panic("can't mipix.HiRes().Draw() outside draw stage") }

	// view culling
	camMinX, camMinY, camMaxX, camMaxY := self.cameraAreaF64() // TODO: this is per draw
	if x > camMaxX || y > camMaxY { return }
	sourceBounds := source.Bounds()
	sourceWidth, sourceHeight := float64(sourceBounds.Dx()), float64(sourceBounds.Dy())
	if x + sourceWidth  < camMinX { return } // outside view
	if y + sourceHeight < camMinY { return } // outside view

	// compile shader if necessary
	if self.shaders[self.scalingFilter] == nil {
		self.compileShader(self.scalingFilter)
	}

	// set triangle vertex coordinates
	targetBounds := target.Bounds()
	targetMinX, targetMinY := float64(targetBounds.Min.X), float64(targetBounds.Min.Y)
	targetWidth, targetHeight := float64(targetBounds.Dx()), float64(targetBounds.Dy())
	xFactor := self.zoomCurrent*targetWidth/float64(self.logicalWidth)
	yFactor := self.zoomCurrent*targetHeight/float64(self.logicalHeight)
	srcProjMinX := (x - camMinX)*xFactor
	srcProjMinY := (y - camMinY)*yFactor
	srcProjMaxX := srcProjMinX + sourceWidth*xFactor
	srcProjMaxY := srcProjMinY + sourceHeight*yFactor
	self.shaderVertices[0].DstX = float32(targetMinX + srcProjMaxX)
	self.shaderVertices[0].DstY = float32(targetMinY + srcProjMinY)
	self.shaderVertices[1].DstX = float32(targetMinX + srcProjMinX)
	self.shaderVertices[1].DstY = self.shaderVertices[0].DstY
	self.shaderVertices[2].DstX = self.shaderVertices[1].DstX
	self.shaderVertices[2].DstY = float32(targetMinY + srcProjMaxY)
	self.shaderVertices[3].DstX = self.shaderVertices[0].DstX
	self.shaderVertices[3].DstY = self.shaderVertices[2].DstY
	
	self.shaderVertices[0].SrcX = float32(sourceBounds.Min.X)
	self.shaderVertices[0].SrcY = float32(sourceBounds.Min.Y)
	self.shaderVertices[1].SrcX = float32(sourceBounds.Max.X)
	self.shaderVertices[1].SrcY = self.shaderVertices[0].SrcY
	self.shaderVertices[2].SrcX = self.shaderVertices[1].SrcX
	self.shaderVertices[2].SrcY = float32(sourceBounds.Max.Y)
	self.shaderVertices[3].SrcX = self.shaderVertices[0].SrcX
	self.shaderVertices[3].SrcY = self.shaderVertices[2].SrcY

	self.shaderOpts.Images[0] = source
	self.shaderOpts.Uniforms["SourceRelativeTextureUnitX"] = float32(float64(self.logicalWidth)/targetWidth)
	self.shaderOpts.Uniforms["SourceRelativeTextureUnitY"] = float32(float64(self.logicalHeight)/targetHeight)
	target.DrawTrianglesShader(
		self.shaderVertices, self.shaderVertIndices,
		self.shaders[self.scalingFilter], &self.shaderOpts,
	)
	self.shaderOpts.Images[0] = nil
}