package mipix

import _ "embed"

import "github.com/hajimehoshi/ebiten/v2"

// TODO: consider using quasilyte's minifier and paste code directly

//go:embed filters/aa_sampling_soft.kage
var _aaSamplingSoft []byte

//go:embed filters/aa_sampling_sharp.kage
var _aaSamplingSharp []byte

//go:embed filters/nearest.kage
var _nearest []byte

//go:embed filters/hermite.kage
var _hermite []byte

//go:embed filters/bicubic.kage
var _bicubic []byte

//go:embed filters/bilinear.kage
var _bilinear []byte

//go:embed filters/src_hermite.kage
var _srcHermite []byte

//go:embed filters/src_bicubic.kage
var _srcBicubic []byte

//go:embed filters/src_bilinear.kage
var _srcBilinear []byte

var pkgSrcKageFilters [scalingFilterEndSentinel][]byte
func init() {
	pkgSrcKageFilters[Nearest] = _nearest
	pkgSrcKageFilters[AASamplingSoft] = _aaSamplingSoft
	pkgSrcKageFilters[AASamplingSharp] = _aaSamplingSharp
	pkgSrcKageFilters[Hermite] = _hermite
	pkgSrcKageFilters[Bicubic] = _bicubic
	pkgSrcKageFilters[Bilinear] = _bilinear
	pkgSrcKageFilters[SrcHermite] = _srcHermite
	pkgSrcKageFilters[SrcBicubic] = _srcBicubic
	pkgSrcKageFilters[SrcBilinear] = _srcBilinear
}

func (self *controller) compileShader(filter ScalingFilter) {
	var err error
	self.shaders[filter], err = ebiten.NewShader(pkgSrcKageFilters[filter])
	if err != nil {
		panic("Failed to compile shader for '" + filter.String() + "' filter: " + err.Error())
	}
	if self.shaderOpts.Uniforms == nil {
		self.initShaderProperties()
	}
}

func (self *controller) initShaderProperties() {
	self.shaderVertices = make([]ebiten.Vertex, 4)
	self.shaderVertIndices = []uint16{0, 1, 3, 3, 1, 2}
	self.shaderOpts.Uniforms = make(map[string]interface{}, 2)
	for i := range 4 { // doesn't matter unless I start doing color scaling
		self.shaderVertices[i].ColorR = 1.0
		self.shaderVertices[i].ColorG = 1.0
		self.shaderVertices[i].ColorB = 1.0
		self.shaderVertices[i].ColorA = 1.0
	}
}
