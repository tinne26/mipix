# mipix

[![Go Reference](https://pkg.go.dev/badge/github.com/tinne26/mipix.svg)](https://pkg.go.dev/github.com/tinne26/mipix)

**WIP: EARLY DEVELOPMENT STAGES**

A package to assist the development of Ebitengine pixel art games.

This package allows you to implement your game working with logical canvases and ignoring `Game.Layout()` completely. Scaling is managed internally with a pixel art aware scaling algorithm, and support for camera movement, zoom and screenshakes are also available through the API.

## Context

This package implements the second model described on [lopix](https://github.com/tinne26/lopix). If `lopix` implements the simplest model for pixel art games, `mipix` is slightly more advanced and provides a much more practical foundation to build pixel art games.

While side scrollers can be implemented with this model, it's not ideal. Any character that "floats while moving", as if sliding through ice, changing animation frames while the position changes smoothly with floating point values, needs to be drawn on a high resolution canvas. There's basic suport for this, but it's not as efficient or prioritized as possible.

## Features

- Draw pixel art logically without having to worry so much about device scale factors, DPI, scaling and projections.
- You can still interleave high resolution and logical draws if needed.
- Basic camera with smooth tracking, zoom and screenshakes. Most behaviors are customizable through interfaces.
