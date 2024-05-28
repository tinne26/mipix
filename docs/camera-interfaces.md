# Camera interfaces

I just wanted to write a small document talking about camera tracking and zooming. After writing a few implementations for [`Zoomer`](https://pkg.go.dev/github.com/tinne26/mipix/zoomer#Zoomer) and [`Tracker`](https://pkg.go.dev/github.com/tinne26/mipix/tracker#Tracker) interfaces, I realized there are many more subtleties than I initially guessed.

For zooming, for example, there are a lot of questions you might want to ask yourself:
- Should zooming be bouncier and playful or more rigid and serious?
- Should zooming be snapier or smoother?
- Should zooming vary more in speed during transitions to look more dynamic, or should it be more consistent?
- Should zooming be compensated to look more perceptually linear?
- Should zoom-ins and zoom-outs be symmetric or asymmetric?
- Should zooming have a speed limit?
- Should zoom starts and ends have additional softening?
- Do I care about quick zoom level stabilization when transition speed is halting?

Depending on the mood of your game, the zoom range, zoom use frequency, zoom control (manual vs automatic) and so on, these questions are not just rhetorical! It's not that hard to imagine different games for basically any combination of answers from the previous questions.

While a functional zoom or tracker can be created with a simple `lerp(current, target, 0.1)`, trying to tailor the implementations to your specific game should not be underestimated. Even if mipix provides a few different implementations, there are many small decisions that will make much more sense when they are made for a specific game and context. Some people might consider this a waste of time, but after having spent more time with it I can clearly see how there's no "universal" solution; it's a very rich space to explore —if you want to—, and even if no one else might care, it's not meaningless that you do.
