package windowstate

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
)

// Coordinates are physical pixels on Windows, including window decorations.
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
type Display struct {
	ID    string
	Work  Rect
	Scale float64
}
type State struct {
	Version   int     `json:"version"`
	Normal    Rect    `json:"normal"`
	Work      Rect    `json:"work"`
	DisplayID string  `json:"displayId"`
	Scale     float64 `json:"scale"`
	Maximized bool    `json:"maximized"`
}

func scale(s float64) float64 {
	if math.IsNaN(s) || math.IsInf(s, 0) || s < .5 || s > 8 {
		return 1
	}
	return s
}
func valid(r Rect) bool {
	return r.Width >= 160 && r.Height >= 120 && r.Width <= 32768 && r.Height <= 32768 && r.X >= -100000 && r.X <= 100000 && r.Y >= -100000 && r.Y <= 100000
}
func overlap(a, b Rect) int {
	return max(0, min(a.X+a.Width, b.X+b.Width)-max(a.X, b.X)) * max(0, min(a.Y+a.Height, b.Y+b.Height)-max(a.Y, b.Y))
}
func Fit(saved State, displays []Display, preferred int) State {
	if len(displays) == 0 {
		displays = []Display{{"fallback", Rect{0, 0, 1024, 720}, 1}}
	}
	preferred = max(0, min(preferred, len(displays)-1))
	selected := preferred
	matched := false
	usable := saved.Version == 1 && valid(saved.Normal)
	if usable {
		for i, d := range displays {
			if saved.DisplayID != "" && d.ID == saved.DisplayID {
				selected = i
				matched = true
				break
			}
		}
		if !matched {
			area := 0
			for i, d := range displays {
				if a := overlap(saved.Normal, d.Work); a > area {
					area = a
					selected = i
					matched = true
				}
			}
		}
	}
	d := displays[selected]
	s := scale(d.Scale)
	work := d.Work
	if work.Width < 160 || work.Height < 120 {
		work = Rect{0, 0, 1024, 720}
	}
	margin := min(int(16*s), min(work.Width/24, work.Height/24))
	available := Rect{work.X + margin, work.Y + margin, work.Width - 2*margin, work.Height - 2*margin}
	width, height := min(int(float64(work.Width)*.92), int(1600*s)), min(int(float64(work.Height)*.92), int(1000*s))
	if usable {
		factor := s / scale(saved.Scale)
		width = int(math.Round(float64(saved.Normal.Width) * factor))
		height = int(math.Round(float64(saved.Normal.Height) * factor))
	}
	width = max(min(int(640*s), available.Width), min(width, available.Width))
	height = max(min(int(460*s), available.Height), min(height, available.Height))
	x, y := available.X+(available.Width-width)/2, available.Y+(available.Height-height)/2
	if usable && matched {
		factor := s / scale(saved.Scale)
		x = work.X + int(math.Round(float64(saved.Normal.X-saved.Work.X)*factor))
		y = work.Y + int(math.Round(float64(saved.Normal.Y-saved.Work.Y)*factor))
	}
	x = max(available.X, min(x, available.X+available.Width-width))
	y = max(available.Y, min(y, available.Y+available.Height-height))
	return State{1, Rect{x, y, width, height}, work, d.ID, s, usable && saved.Maximized}
}

// Optimal uses the selected monitor rather than an old window's dimensions.
// Start centered at 90% of the previous normal size, leaving desktop space.
func Optimal(saved State, displays []Display, preferred int) State {
	selected := Fit(saved, displays, preferred)
	result := Fit(State{}, []Display{{selected.DisplayID, selected.Work, selected.Scale}}, 0)
	result.Normal.Width = max(160, result.Normal.Width*9/10)
	result.Normal.Height = max(120, result.Normal.Height*9/10)
	result.Normal.X = result.Work.X + (result.Work.Width-result.Normal.Width)/2
	result.Normal.Y = result.Work.Y + (result.Work.Height-result.Normal.Height)/2
	result.Maximized = false
	return result
}
func Load(path string) State {
	data, err := os.ReadFile(path)
	if err != nil || len(data) > 16384 {
		return State{}
	}
	var s State
	if json.Unmarshal(data, &s) != nil {
		return State{}
	}
	return s
}
func Save(path string, s State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".window-*.json")
	if err != nil {
		return err
	}
	tmp := file.Name()
	defer os.Remove(tmp)
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return os.WriteFile(path, data, 0600)
	}
	return nil
}
