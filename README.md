# mipix

[![Go Reference](https://pkg.go.dev/badge/github.com/tinne26/mipix.svg)](https://pkg.go.dev/github.com/tinne26/mipix)

**WIP: EARLY DEVELOPMENT STAGES**

A package to assist the development of Ebitengine pixel art games.

This package allows you to implement your game working with logical canvases and ignoring `Game.Layout()` completely. Scaling is managed internally with pixel art aware scaling algorithms, and support for camera movement, zoom and screen shakes are also available through the API.

## Features

- Draw pixel art logically without having to deal with device scale factors, DPI, scaling and projections yourself.
- Basic camera with smooth tracking, zoom and screen shakes. Most behaviors can be customized through interfaces.
- Interleave high resolution and logical draws when necessary.

## Context

This package implements the second model described on [lopix](https://github.com/tinne26/lopix). If `lopix` implements the simplest model for pixel art games, `mipix` is slightly more advanced and provides a much more practical foundation to build pixel art games:
- You are expected to place most of your game world elements at integer coordinates.
- Your draw method receives a logically sized offscreen corresponding to a specific area of your game world. This area can vary based on the current camera position, zoom and shake, but you are simply given the task of *rendering a logical area of your game world*, directly and in pure pixels, to a logically sized canvas.
- For UI and other camera-independent parts of your game, you can create `mipix.Offscreen`s manually, again with pure pixel based sizes, render on them in a straight-forward manner and then use the built-in `Offscreen.Project()` method to deal with the scaling and stretching and filtering and all that nonsense.

While side scrollers can be implemented with this model, that's probably not ideal. In most side-scrollers, developers use floating point positions for characters, which are smoothly "sliding through the floor" as animations change. Doing this requires drawing the characters on a high resolution canvas. The API offers basic support for this, but it's not the main focus of the package. If you have *many* fractionally positioned characters and elements, the `mipix` model might not be the best match.

## Code example

I haven't written examples oriented to end users yet, the only examples available are a bit overkill as they were designed to help me debug and test features. See [https://github.com/tinne26/lopix-examples].

## WIP status

- Unimplemented: SpringTracker, moaar trackers and shakers and gen-zers and so on.
- Undertested: stretching, varying TPS levels, tick rates != 1.
- Unclear: still debating whether to expose interfaces for zooming or not. Current model is not robust and flexible enough.
- Missing features: probably a couple window size management functions (though we ideally need [ebiten/issues/2978](https://github.com/hajimehoshi/ebiten/issues/2978) resolved first), maaaaybe `HiRes().DrawTriangles()`, or some form of `DrawEachAt(...)`..? I'm surprisingly happy with the feature set at the moment, to be honest.
- Annoyances: practically speaking, drawing UI on top of mipix is a bit annoying. Ideally, you would use bitmap fonts and a bitmap font renderer, but no good solution exists for this yet. I'm working on it on a separate project, though, so hopefully someday...
