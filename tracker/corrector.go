package tracker

import "github.com/tinne26/mipix/internal"

type corrector struct {
	speedX, speedY float64 // logical, relative to resolution and zoom
	acceleration float64	
	initialized bool
}

func (self *corrector) initialize() {
	self.initialized = true
	self.acceleration = 0.2
}

// Reasonable values are typically in the [0.05, 0.5] range.
func (self *corrector) SetAcceleration(acceleration float64) {
	if acceleration < 0.0 { panic("acceleration can't be < 0.0") }
	self.initialized = true
	self.acceleration = acceleration
}

func (self *corrector) Update(errorX, errorY float64) {
	if !self.initialized { self.initialize() }

	w, h := internal.GetResolution()
	w64, h64 := float64(w), float64(h)
	//zoom := internal.GetCurrentZoom()
	errorX /= w64
	errorY /= h64

	predictedX := (self.speedX*self.speedX)/(2.0*self.acceleration)
	predictedY := (self.speedY*self.speedY)/(2.0*self.acceleration)
	targetX := internal.Abs(errorX)
	targetY := internal.Abs(errorY)
	
	updateDelta := 1.0/float64(internal.GetUPS())
	speedChange := self.acceleration*updateDelta
	margin := (speedChange*speedChange)/(2.0*self.acceleration)

	// update speeds
	// (NOTICE: the code could be greatly shortened, but
   // I'm happy with it being long, simple and obvious
   // in this particular case)
	if errorX < 0.0 { // wanna go left
		if self.speedX > 0.0 {
			self.speedX -= speedChange // decelerate to turn around
		} else if predictedX < targetX - margin {
			self.speedX -= speedChange // accelerate
		} else if predictedX > targetX {
			self.speedX += speedChange // decelerate
		} else if -self.speedX <= speedChange + 0.0005 {
			self.speedX = 0.0 // stabilization
		}
	} else if errorX > 0.0 { // wanna go right
		if self.speedX < 0.0 {
			self.speedX += speedChange // decelerate to turn around
		} else if predictedX < targetX - margin {
			self.speedX += speedChange // accelerate
		} else if predictedX > targetX {
			self.speedX -= speedChange // decelerate
		} else if self.speedX <= speedChange + 0.0005 {
			self.speedX = 0.0 // stabilization
		}
	} else { // errorX == 0.0
		if self.speedX > 0.0 && predictedX >= margin {
			self.speedX -= speedChange
		} else if self.speedX < 0.0 && predictedX >= margin {
			self.speedX += speedChange
		} else {
			self.speedX = 0.0
		}
	}
	
	if errorY < 0.0 { // wanna go up
		if self.speedY > 0.0 {
			self.speedY -= speedChange // decelerate to turn around
		} else if predictedY < targetY - margin {
			self.speedY -= speedChange // accelerate
		} else if predictedY > targetY {
			self.speedY += speedChange // decelerate
		} else if -self.speedY <= speedChange + 0.0005 {
			self.speedY = 0.0 // stabilization
		}
	} else if errorY > 0.0 { // wanna go down
		if self.speedY < 0.0 {
			self.speedY += speedChange // decelerate to turn around
		} else if predictedY < targetY - margin {
			self.speedY += speedChange // accelerate
		} else if predictedY > targetY {
			self.speedY -= speedChange // decelerate
		} else if self.speedY <= speedChange + 0.0005 {
			self.speedY = 0.0 // stabilization
		}
	} else { // errorY == 0.0
		if self.speedY > 0.0 && predictedY >= margin {
			self.speedY -= speedChange
		} else if self.speedY < 0.0 && predictedY >= margin {
			self.speedY += speedChange
		} else {
			self.speedY = 0.0
		}
	}
}

func (self *corrector) Decelerate() {
	updateDelta := 1.0/float64(internal.GetUPS())
	speedChange := self.acceleration*updateDelta
	
	if self.speedX != 0.0 {
		if self.speedX > 0.0 {
			self.speedX = max(0.0, self.speedX - speedChange)
		} else {
			self.speedX = min(0.0, self.speedX + speedChange)
		}
	}
	if self.speedY != 0.0 {
		if self.speedY > 0.0 {
			self.speedY = max(0.0, self.speedY - speedChange)
		} else {
			self.speedY = min(0.0, self.speedY + speedChange)
		}
	}

}
