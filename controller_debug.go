package mipix

import "fmt"

import "github.com/hajimehoshi/ebiten/v2"
import "github.com/hajimehoshi/ebiten/v2/ebitenutil" // not desirable, but let's ignore it for the moment

func (self *controller) debugDrawf(format string, args ...any) {
	self.debugInfo = append(self.debugInfo, fmt.Sprintf(format, args...))
}

func (self *controller) debugPrintfr(firstTick, lastTick uint64, format string, args ...any) {
	if self.currentTick >= firstTick && self.currentTick <= lastTick {
		fmt.Printf(format, args...)
	}
}

func (self *controller) debugPrintfe(everyNTicks uint64, format string, args ...any) {
	if self.currentTick % everyNTicks == 0 {
		fmt.Printf(format, args...)
	}
}

func (self *controller) debugPrintfk(key ebiten.Key, format string, args ...any) {
	if ebiten.IsKeyPressed(key) {
		fmt.Printf(format, args...)
	}
}

// --- internal ---

func (self *controller) debugDrawAll(target *ebiten.Image) {
	if len(self.debugInfo) == 0 { return }

	// determine offscreen size
	targetBounds := target.Bounds()
	targetWidth, targetHeight := float64(targetBounds.Dx()), float64(targetBounds.Dy())
	height := 256/ebiten.Monitor().DeviceScaleFactor()
	width  := height*(targetWidth/targetHeight)
	offWidth, offHeight := int(width), int(height)

	// create offscreen if necessary
	if self.debugOffscreen == nil {
		self.debugOffscreen = NewOffscreen(offWidth, offHeight)
	} else {
		currWidth, currHeight := self.debugOffscreen.Size()
		if currWidth != offWidth || currHeight != offHeight {
			self.debugOffscreen = NewOffscreen(offWidth, offHeight)
		} else { // (unless skip draw, but debug is only called if needsRedraw)
			self.debugOffscreen.Clear()
		}
	}

	// draw info to offscreen and project
	for i, info := range self.debugInfo {
		ebitenutil.DebugPrintAt(self.debugOffscreen.Target(), info, 1, 1 + i*12)
	}
	self.debugOffscreen.Project(target)

	// clear debug info
	self.debugInfo = self.debugInfo[ : 0]
}
