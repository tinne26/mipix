package mipix

import "github.com/tinne26/mipix/tracker"
import "github.com/tinne26/mipix/zoomer"
import "github.com/tinne26/mipix/shaker"

var defaultZoomer *zoomer.Quadratic
var defaultTracker *tracker.SpringTailer
var defaultShaker *shaker.Random
