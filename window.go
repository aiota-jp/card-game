package main

import (
	"cardlounge/internal/windowstate"
	"context"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type windowCapture struct {
	Bounds               windowstate.Rect
	Display              windowstate.Display
	Maximized, Minimized bool
}
type windowController struct {
	mu     sync.Mutex
	path   string
	handle uintptr
	state  windowstate.State
	cancel context.CancelFunc
}

func newWindowController() *windowController {
	root, err := os.UserConfigDir()
	if err != nil {
		return &windowController{}
	}
	return &windowController{path: filepath.Join(root, "CardGameLounge", "window.json")}
}
func (w *windowController) ready(ctx context.Context) {
	// Never flash a 1000px-tall window before applying display bounds.
	defer wruntime.WindowShow(ctx)
	w.handle = platformHandle()
	screens, preferred := platformDisplays(ctx, w.handle)
	w.state = windowstate.Optimal(windowstate.Load(w.path), screens, preferred)
	logicalW := int(float64(w.state.Work.Width) / w.state.Scale)
	logicalH := int(float64(w.state.Work.Height) / w.state.Scale)
	wruntime.WindowSetMinSize(ctx, max(160, min(600, logicalW-64)), max(120, min(360, logicalH-96)))
	platformApply(ctx, w.handle, w.state.Normal)
	if w.state.Maximized {
		wruntime.WindowMaximise(ctx)
	}
	pollCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-pollCtx.Done():
				return
			case <-ticker.C:
				w.capture(ctx)
			}
		}
	}()
}
func (w *windowController) capture(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	current, ok := platformCapture(ctx, w.handle)
	if !ok || current.Minimized {
		return
	}
	w.state.Maximized = current.Maximized
	if !current.Maximized {
		w.state.Normal = current.Bounds
		w.state.DisplayID = current.Display.ID
		w.state.Work = current.Display.Work
		w.state.Scale = current.Display.Scale
	} else if current.Display.ID != w.state.DisplayID {
		w.state = windowstate.Fit(w.state, []windowstate.Display{current.Display}, 0)
		w.state.Maximized = true
	}
}
func (w *windowController) beforeClose(ctx context.Context) bool {
	if w.cancel != nil {
		w.cancel()
	}
	w.capture(ctx)
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.path != "" && w.state.Version == 1 {
		if err := windowstate.Save(w.path, w.state); err != nil {
			log.Printf("window preferences: %v", err)
		}
	}
	return false
}
func (w *windowController) shutdown(context.Context) {
	if w.cancel != nil {
		w.cancel()
	}
}
