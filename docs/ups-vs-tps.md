# UPS vs TPS

The API for mipix exposes some methods like [`mipix.Tick().UPS()`](https://pkg.go.dev/github.com/tinne26/mipix#AccessorTick.UPS), [`mipix.Tick().TPS()`](https://pkg.go.dev/github.com/tinne26/mipix#AccessorTick.TPS), [`mipix.Tick().SetRate()`](https://pkg.go.dev/github.com/tinne26/mipix#AccessorTick.SetRate) and [`mipix.Tick().GetRate()`](https://pkg.go.dev/github.com/tinne26/mipix#AccessorTick.GetRate) that might have you confused. What does all this mean?

> [!IMPORTANT]
> Ebitengine, unlike most game engines, uses a tick-based fixed timestep loop for updating game logic. If you are only familiar with "delta times" and this sounds strange to you, [tinne26/tps-vs-fps](https://github.com/tinne26/tps-vs-fps) explains the topic in more detail.
>
> The rest of this document assumes you understand the TPS model.

UPS stands for "updates per second". Instead of a fixed timestep loop with *ticks*, where one tick corresponds to one update, mipix uses a slightly more sophisticated system where you still use a fixed timestep loop with N updates per second, but each update might simulate *more than one tick*.

The default `ebiten.TPS() == 60`, in this model, corresponds to `UPS() == 60` and `TPU() == 1` (`TPU` means "ticks per update", which mipix refers to as "tick rate").

The reason for this model to exist is to *better support high-refresh display rates and/or reduce input latency*, but without forcing the user to run the game at a higher simulation rate if they don't want to.

> [!NOTE]
> Not all games are a *good match for higher refresh rates*. You might have a very pure pixel art game with animations that only run at 8 frames per second, all in sync. Trying to artificially support high refresh rate displays in this case would be silly. This document explains a feature that's interesting for *most games*, but not all of them. Use your common sense before joining the feature hype train and all that.

As you know, many modern displays can run at 120Hz, 144Hz or 240Hz. We call these "high refresh rate displays". If we want to support these on Ebitengine, we have a couple options:
- Interpolate positions smoothly between the current and previous updates. This is not always so simple to do and will often introduce extra latency.
- Set a higher `TPS`. With more granular simulation steps, we have something new to show on each frame even on a high refresh rate display.

There are some arguments and use-cases for the first approach, but we will be exploring the second option. Can't we do that already with Ebitengine's model? Just run the game at 240 ticks per second!

That's correct. The main problem with doing that is that you are paying the cost whether you have a 240Hz display or not. Now, this is not always a real problem, maybe your logic is simple enough that that's perfectly acceptable, but a general solution needs to be more flexible.

The general solution is the UPS model, which you should already be starting to understand by now:
- Out internal logic will always run at 240 ticks per second, but the number of updates per second is not necessarily 240. It might be 120 UPS, with a tick rate of 2 ticks per update instead (still a total of 240 ticks per second, but with less updates), or 60 UPS, with a tick rate of 4 TPU.

And this is what mipix does. The old `ebiten.TPS` is hidden under `mipix.Tick().UPS()` now, and we can control the tick rate independently too.

The 240 internal tick rate is a good recommendation, because it makes sense for 60Hz, 120Hz and 240Hz, but you can totally explore something different, like 40, 80, 160 and so on. I don't know why, but you could do that if you want.

## Tick-rate independent and update-rate independent

In mipix, there are multiple functions that use ticks for transitions. For example, shakes can have fade ins and fade outs that are measured in ticks. The reason to stick to ticks instead of delta times is what was already laid out on [tinne26/tps-vs-fps](https://github.com/tinne26/tps-vs-fps): determinism and simplicity. I don't have anything against delta times, but they go against Ebitengine's ethos.

Well, it's all a lie. Plenty of mipix interface implementations are based on times instead of ticks.

Many interface implementations in mipix mention being tick-rate independent or update-rate independent:
- Tick-rate independent means that no matter what tick rate or UPS you set, the result will be perceptually the same. This is usually guaranteed with default [`Tracker`](https://pkg.go.dev/github.com/tinne26/mipix/tracker#Tracker) implementations. This is generally only recommended for visuals, not game logic.
- Update-rate independent is a less strict promise, as changing the total amount of ticks per second will change the results too. This is used with [`Shaker`](https://pkg.go.dev/github.com/tinne26/mipix/shaker#Shaker) and [`Zoomer`](https://pkg.go.dev/github.com/tinne26/mipix/zoomer#Zoomer), as they use tick-based transitions.

> [!CAUTION]
> In practice, the precision of many algorithms depends on the number of iterations it simulates. For very high simulation rates, results would typically converge with a high degree of accuracy, but... if you use 30UPS, some things might not go so smoothly. It's always advisable to test all your configurations.

## Summary and conclusions

- The UPS model (update per second) allows multiple ticks per update in order to better support high refresh rate displays and lower latencies when required, in a configurable manner, at runtime.
- If you are making a non-toy project and are interested in high refresh rates, consider starting your project at 240 ticks per second internally, typically with a default of 60UPS@4TPU. If you later decide to support 120UPS@2TPU and 240UPS@1TPU, everything is already compatible.
- For game logic, prioritize working with ticks. Advance your internal tick counters with `mipix.Tick().GetRate()` per update. For visuals, if you want to make something tick-rate independent, you can multiply values by the update delta (the time you should be simulating during an update, which is `1.0/float64(mipix.Tick().UPS())`). If you want to make it update-rate independent, use the same technique you use for game logic.
