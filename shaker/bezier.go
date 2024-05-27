package shaker

// use points within an ellipse, and connect a - b - c through conic bézier
// what about sine-like for y and noisy for x?
// and Math.sqrt(1 - Math.pow(x - 1, 2)) for circular interpolation in the top-left quadrant
// ...
// ellipse equation: (x/horzAxisLength)^2+(y/vertAxisLength)^2 = 1
// just intersect with line equation and solve the system.
// so, just make a simple function that returns the value given
// axis lengths and angle. you first find the intersection point
// and then use a^2 + b^2 = h^2.
// with bézier conic curves, sliding 2/3 points and rng angles

