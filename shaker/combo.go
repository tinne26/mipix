package shaker

var _ Shaker = (*Combo)(nil)

// An example [Shaker] created by combining a [Balanced] and
// a [Random] shaker. This is only offered to showcase how
// easy it is to create new shakers by combining previously
// existing ones. Since this is only an example, no methods
// for configuring the parameters are exposed.
type Combo struct {
	balanced Balanced
	rand Random
	initialized bool
}

func (self *Combo) initialize() {
	self.initialized = true
	self.balanced.SetMotionScale(0.014)
	self.balanced.SetTravelTime(0.26)
	self.rand.SetTravelTime(0.03)
	self.rand.SetMotionScale(0.02)
}

// Implements [Shaker].
func (self *Combo) GetShakeOffsets(level float64) (float64, float64) {
	if !self.initialized { self.initialize() }
	bx, by := self.balanced.GetShakeOffsets(level)
	rx, ry := self.rand.GetShakeOffsets(level)
	return bx + rx, by + ry
}
