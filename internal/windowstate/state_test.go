package windowstate

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func inside(a, b Rect) bool {
	return a.X >= b.X && a.Y >= b.Y && a.X+a.Width <= b.X+b.Width && a.Y+a.Height <= b.Y+b.Height
}
func TestFitDisplaysAndDPI(t *testing.T) {
	screens := []Display{{"laptop", Rect{0, 0, 1366, 728}, 1}, {"external", Rect{-2560, 0, 2560, 1400}, 1.5}}
	first := Fit(State{}, screens, 0)
	if !inside(first.Normal, screens[0].Work) || first.Normal.Height >= 728 {
		t.Fatal(first)
	}
	saved := State{Version: 1, Normal: Rect{-2400, 100, 1500, 1100}, Work: screens[1].Work, DisplayID: "external", Scale: 1.5}
	restored := Fit(saved, screens, 0)
	if restored.Normal != saved.Normal {
		t.Fatalf("negative monitor coordinates not restored: %+v", restored)
	}
	// Same monitor moved to the right and changed DPI. Keep logical size and fit.
	changed := Display{"external", Rect{1920, 0, 3840, 2080}, 2}
	resized := Fit(saved, []Display{screens[0], changed}, 0)
	if resized.DisplayID != "external" || resized.Normal.Width != 2000 || !inside(resized.Normal, changed.Work) {
		t.Fatalf("DPI change: %+v", resized)
	}
	// Monitor unplugged: move the previous window onto the remaining display.
	removed := Fit(saved, screens[:1], 0)
	if removed.DisplayID != "laptop" || !inside(removed.Normal, screens[0].Work) {
		t.Fatalf("removed monitor: %+v", removed)
	}
}
func TestFitsSmallWorkAreasAndCorruptPreferences(t *testing.T) {
	rng := rand.New(rand.NewSource(74))
	for i := 0; i < 3000; i++ {
		work := Rect{rng.Intn(6000) - 3000, rng.Intn(2000) - 1000, 320 + rng.Intn(3700), 240 + rng.Intn(2100)}
		saved := State{Version: 1, Normal: Rect{rng.Intn(30000) - 15000, rng.Intn(30000) - 15000, rng.Intn(30000), rng.Intn(30000)}, Scale: float64(rng.Intn(8)) / 2, Maximized: true}
		result := Fit(saved, []Display{{"test", work, .75 + float64(rng.Intn(6))*.25}}, 0)
		if !inside(result.Normal, work) || result.Normal.Width <= 0 || result.Normal.Height <= 0 {
			t.Fatalf("window escaped work area: %+v work=%+v", result, work)
		}
	}
	invalid := State{Version: 999, Normal: Rect{0, 0, 900, 700}, Maximized: true}
	if Fit(invalid, nil, 0).Maximized {
		t.Fatal("invalid state restored maximized")
	}
}
func TestSettingsRoundTripAndRecovery(t *testing.T) {
	file := filepath.Join(t.TempDir(), "settings", "window.json")
	state := Fit(State{}, []Display{{"main", Rect{0, 40, 1920, 1000}, 1.25}}, 0)
	state.Maximized = true
	for i := 0; i < 2; i++ {
		if err := Save(file, state); err != nil {
			t.Fatal(err)
		}
		if got := Load(file); got != state {
			t.Fatalf("preferences changed: %+v", got)
		}
	}
	if err := os.WriteFile(file, []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := Load(file); got.Version != 0 {
		t.Fatal("corrupt settings accepted")
	}
	if got := Load(file + "missing"); got.Version != 0 {
		t.Fatal("missing preferences accepted")
	}
}

func TestOptimalUsesMonitorWorkArea(t *testing.T) {
	for _, screen := range []Display{{"laptop", Rect{0, 0, 1366, 728}, 1}, {"fullHD", Rect{0, 0, 1920, 1040}, 1}, {"scaled4K", Rect{-3840, 0, 3840, 2080}, 2}, {"small", Rect{0, 0, 1024, 700}, 1}} {
		saved := State{Version: 1, Normal: Rect{0, 0, 640, 460}, DisplayID: screen.ID, Work: screen.Work, Scale: screen.Scale}
		result := Optimal(saved, []Display{screen}, 0)
		if result.Maximized || !inside(result.Normal, screen.Work) || result.Normal.Width <= 640 || result.Normal.Height <= 460 {
			t.Fatalf("monitor sizing: %+v", result)
		}
	}
	monitors := []Display{{"main", Rect{0, 0, 1920, 1040}, 1}, {"left", Rect{-2560, 0, 2560, 1400}, 1.5}}
	saved := State{Version: 1, Normal: Rect{-2500, 60, 800, 600}, DisplayID: "left", Work: monitors[1].Work, Scale: 1.5}
	if result := Optimal(saved, monitors, 0); result.DisplayID != "left" || !inside(result.Normal, monitors[1].Work) {
		t.Fatal(result)
	}
	if result := Optimal(saved, monitors[:1], 0); result.DisplayID != "main" || !inside(result.Normal, monitors[0].Work) {
		t.Fatal(result)
	}
}
