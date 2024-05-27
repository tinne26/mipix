package internal

import "github.com/hajimehoshi/ebiten/v2"

var BridgedLogicalWidth int
var BridgedLogicalHeight int
var CurrentZoom float64
var CurrentTPU uint64 // ticks per update

func GetCurrentZoom() float64 {
	return CurrentZoom
}

func GetResolution() (int, int) {
	return BridgedLogicalWidth, BridgedLogicalHeight
}

func GetUPS() int {
	return ebiten.TPS()
}

func GetTPU() uint64 {
	return CurrentTPU
}
