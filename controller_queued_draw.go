package mipix

import "github.com/hajimehoshi/ebiten/v2"

type queuedDraw struct {
	hiResFunc func(*ebiten.Image, *ebiten.Image)
	logicalFunc func(*ebiten.Image)
}

func (self *queuedDraw) IsHighResolution() bool {
	return self.hiResFunc != nil
}

func (self *controller) queueDraw(handler func(*ebiten.Image)) {
	if !self.inDraw { panic("can't queue draw outside draw stage") }
	self.queuedDraws = append(self.queuedDraws, queuedDraw{ logicalFunc: handler })
}

func (self *controller) queueHiResDraw(handler func(*ebiten.Image, *ebiten.Image)) {
	if !self.inDraw { panic("can't queue draw outside draw stage") }
	self.queuedDraws = append(self.queuedDraws, queuedDraw{ hiResFunc: handler })
}
