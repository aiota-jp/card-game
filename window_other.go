//go:build !windows

package main

import (
	"cardlounge/internal/windowstate"
	"context"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func platformPrepare()        {}
func platformHandle() uintptr { return 0 }
func platformDisplays(ctx context.Context, _ uintptr) ([]windowstate.Display, int) {
	screens, err := wruntime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return nil, 0
	}
	// Wails v2 exposes logical screen sizes but no monitor origins/work areas on
	// these platforms. Use a conservative margin for desktop panels.
	preferred := 0
	result := make([]windowstate.Display, 0, len(screens))
	for i, s := range screens {
		if s.IsCurrent {
			preferred = i
		}
		result = append(result, windowstate.Display{ID: "logical", Work: windowstate.Rect{Width: s.Size.Width, Height: max(240, s.Size.Height-64)}, Scale: 1})
	}
	return result, preferred
}
func platformCapture(ctx context.Context, _ uintptr) (windowCapture, bool) {
	screens, index := platformDisplays(ctx, 0)
	if len(screens) == 0 {
		return windowCapture{}, false
	}
	x, y := wruntime.WindowGetPosition(ctx)
	width, height := wruntime.WindowGetSize(ctx)
	return windowCapture{windowstate.Rect{X: x, Y: y, Width: width, Height: height}, screens[index], wruntime.WindowIsMaximised(ctx), wruntime.WindowIsMinimised(ctx)}, true
}
func platformApply(ctx context.Context, _ uintptr, r windowstate.Rect) {
	wruntime.WindowSetSize(ctx, r.Width, r.Height)
	wruntime.WindowSetPosition(ctx, r.X, r.Y)
}
