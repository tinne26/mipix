package internal

import "math"

type Spring struct {
	initialized bool
	damping   float64 // drag / resistance [0...1.0]
	frequency float64 // power / speed [0.001...Inf]

	lastUPS int
	tempAlpha  float64
	tempExp    float64
	tempCosExp float64
	tempSinExp float64
}

func (self *Spring) IsInitialized() bool {
	return self.initialized
}

func (self *Spring) SetParameters(damping, frequency float64) {
	if damping < 0.0 || damping > 1.0 {
		panic("damping must be in [0.0, 1.0] range")
	}
	if frequency < 0.001 {
		panic("frequency must be in >= 0.001")
	}
	if damping != self.damping || frequency != self.frequency {
		self.damping   = damping
		self.frequency = frequency
		self.recomputeExpensiveTerms()
		self.initialized = true
	}
}

func (self *Spring) recomputeExpensiveTerms() {
	self.lastUPS = GetUPS()
	delta := 1.0/float64(self.lastUPS)
	if self.damping >= 0.999 {
		self.tempExp   = math.Exp(-self.frequency*delta)
	} else {
		freqByDamp := self.frequency*self.damping
		self.tempAlpha = self.frequency*math.Sqrt(1.0 - self.damping*self.damping)
		self.tempExp    = math.Exp(-freqByDamp*delta)
		self.tempSinExp = math.Sin(self.tempAlpha*delta)*self.tempExp
		self.tempCosExp = math.Cos(self.tempAlpha*delta)*self.tempExp
	}
}

// Returns the new position and new speed.
func (self *Spring) Update(current, target, speed float64) (float64, float64) {
	if !self.initialized { panic("must Spring.SetParameters() before using") }
	
	if GetUPS() != self.lastUPS {
		self.recomputeExpensiveTerms()
	}

	var posPos, velVel, posVel, velPos float64
	if self.damping >= 0.999 {
		posVel = (1.0/float64(self.lastUPS))*self.tempExp
		expr := posVel*self.frequency
		velPos = -self.frequency*expr
		posPos = +expr + self.tempExp
		velVel = -expr + self.tempExp
	} else {
		freqByDamp := self.frequency*self.damping
		expr := self.tempSinExp*freqByDamp*(1.0/self.tempAlpha)
		posVel = +self.tempSinExp*(1.0/self.tempAlpha)
		velPos = -self.tempSinExp*self.tempAlpha - freqByDamp*expr
		posPos = self.tempCosExp + expr
		velVel = self.tempCosExp - expr
	}

	mirroredStart := current - target
	current = mirroredStart*posPos + speed*posVel + target
	speed   = mirroredStart*velPos + speed*velVel
	return current, speed
}
