package mipix

import "github.com/tinne26/mipix/internal"

func (self *controller) convertToRelativeCoords(x, y int) (float64, float64) {
	xMargin, yMargin := self.hackyGetMargins()
	relX := (float64(x) - xMargin)/(float64(self.hiResWidth ) - xMargin*2)
	relY := (float64(y) - yMargin)/(float64(self.hiResHeight) - yMargin*2)
	return internal.Clamp(relX, 0.0, 1.0), internal.Clamp(relY, 0.0, 1.0)
}

func (self *controller) convertToLogicalCoords(x, y int) (float64, float64) {
	rx, ry := self.convertToRelativeCoords(x, y)
	minX, minY, _, _ := self.cameraAreaF64()
	return minX + rx*float64(self.logicalWidth)/self.zoomCurrent, minY + ry*float64(self.logicalHeight)/self.zoomCurrent
}

func (self *controller) hackyGetMargins() (float64, float64) {
	if self.stretchingEnabled { return 0, 0 }
	var hiWidth, hiHeight int
	if self.inDraw {
		hiWidth  = self.prevHiResCanvasWidth
		hiHeight = self.prevHiResCanvasHeight
	} else {
		hiWidth  = self.hiResWidth
		hiHeight = self.hiResHeight
	}

	hiAspectRatio := float64(hiWidth)/float64(hiHeight)
	loAspectRatio := float64(self.logicalWidth)/float64(self.logicalHeight)
	switch {
	case hiAspectRatio == loAspectRatio: // just scaling
		return 0, 0
	case hiAspectRatio  > loAspectRatio: // horz margins
		xMargin := int((float64(hiWidth) - loAspectRatio*float64(hiHeight))/2.0)
		return float64(xMargin), 0
	case loAspectRatio  > hiAspectRatio: // vert margins
		yMargin := int((float64(hiHeight) - float64(hiWidth)/loAspectRatio)/2.0)
		return 0, float64(yMargin)
	default:
		panic("unreachable")
	}
}
