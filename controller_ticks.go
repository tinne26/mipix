package mipix

func (self *controller) tickNow() uint64 {
	return self.currentTick
}

func (self *controller) tickSetRate(rate int) {
	if rate < 1 || rate > 256 { panic("tick rate must be within [1, 256]") }
	self.tickRate = uint64(rate)
}

func (self *controller) tickGetRate() int {
	return int(self.tickRate)
}
