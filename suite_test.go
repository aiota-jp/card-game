package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	wassets "github.com/wailsapp/wails/v2/pkg/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"io/fs"
	"net/http/httptest"
	"strings"
	"testing"
)

type testRuntime struct{}

func (testRuntime) DesktopIPC() []byte       { return []byte("/* desktop IPC */") }
func (testRuntime) WebsocketIPC() []byte     { return nil }
func (testRuntime) RuntimeDesktopJS() []byte { return []byte("/* runtime */") }

func TestWailsSuiteRouting(t *testing.T) {
	base, _ := fs.Sub(assets, "frontend")
	web := newSuiteAssets(base)
	server, err := wassets.NewAssetServer("", assetserver.Options{Assets: web}, false, nil, testRuntime{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"/", "/poker/game.html", "/blackjack/game.html", "/baccarat/game.html", "/poker/assets/rina-full.png", "/baccarat/assets/voice-rina/reward-01-expressive.wav"} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest("GET", "http://wails.localhost"+name, nil))
		if response.Code != 200 {
			t.Fatal(name, response.Code, response.Body.String())
		}
		injected := strings.Contains(response.Body.String(), "/wails/runtime.js")
		if injected != (name == "/") {
			t.Fatal("runtime must be injected into hub only", name)
		}
	}
}

func TestSuiteAssets(t *testing.T) {
	base, _ := fs.Sub(assets, "frontend")
	web := newSuiteAssets(base)
	for _, mode := range []string{"poker", "blackjack", "baccarat"} {
		html, err := fs.ReadFile(web, mode+"/game.html")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(html), "window.go=window.parent.go") {
			t.Fatal("missing parent binding", mode)
		}
		for _, folder := range []string{"voice", "voice-ayaka", "voice-rina", "voice-shizue", "voice-kathy"} {
			for _, file := range []string{"manifest.json", "reward-manifest.json"} {
				data, err := fs.ReadFile(web, mode+"/assets/"+folder+"/"+file)
				if err != nil {
					t.Fatal(err)
				}
				var manifest struct {
					Lines []struct {
						Files map[string]struct{ Path, SHA256 string }
					}
				}
				if err = json.Unmarshal(data, &manifest); err != nil {
					t.Fatal(err)
				}
				for _, line := range manifest.Lines {
					for _, audio := range line.Files {
						data, err := fs.ReadFile(web, mode+"/"+audio.Path)
						if err != nil {
							t.Fatal(err)
						}
						sum := sha256.Sum256(data)
						if hex.EncodeToString(sum[:]) != audio.SHA256 {
							t.Fatal("voice hash", mode, audio.Path)
						}
					}
				}
			}
		}
	}
	for _, bad := range []string{"../go.mod", "/index.html", "poker/../../go.mod"} {
		if _, err := web.Open(bad); err == nil {
			t.Fatal("invalid path accepted", bad)
		}
	}
}

func TestSeparateSessions(t *testing.T) {
	p, b, c := NewPokerApp(), NewBlackjackApp(), NewBaccaratApp()
	before := p.Snapshot()
	if _, err := p.Act("fold"); err != nil {
		t.Fatal(err)
	}
	settled := p.Snapshot()
	if !settled.Done {
		t.Fatal("poker did not settle")
	}
	if _, err := b.Deal(50); err != nil {
		t.Fatal(err)
	}
	bj := b.Snapshot()
	if _, err := c.OpenGame("banque", "rina"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SetRole(true); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Deal(50, "banker"); err != nil {
		t.Fatal(err)
	}
	if p.Snapshot().Stacks != settled.Stacks || p.Snapshot().Hand != before.Hand {
		t.Fatal("poker session changed")
	}
	if b.Snapshot().Chips != bj.Chips || b.Snapshot().Hand != bj.Hand {
		t.Fatal("blackjack session changed")
	}
	b.Reset()
	if p.Snapshot().Stacks != settled.Stacks || c.Snapshot().Variant != "banque" {
		t.Fatal("reset affected other games")
	}
}

func TestLobbyGreetingAudio(t *testing.T) {
	data, err := assets.ReadFile("frontend/lobby-voice/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Lines []struct {
			ID, Character, Game, Text string
			Files                     map[string]struct {
				Path, SHA256 string
				Duration     float64
			}
		}
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Lines) != 15 {
		t.Fatal("expected fifteen complete greetings")
	}
	seen := map[string]bool{}
	for _, line := range manifest.Lines {
		if seen[line.ID] {
			t.Fatal("duplicate greeting", line.ID)
		}
		seen[line.ID] = true
		if !strings.Contains(line.Text, "こんばんは") || !strings.Contains(line.Text, "趣味") || !strings.Contains(line.Text, "最近") || !strings.Contains(line.Text, "三つのゲームから") {
			t.Fatal("incomplete introduction", line.ID)
		}
		for _, mode := range []string{"expressive", "calm"} {
			audio, ok := line.Files[mode]
			if !ok {
				t.Fatal("missing audio mode", line.ID, mode)
			}
			data, err := assets.ReadFile("frontend/" + audio.Path)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(data)
			if hex.EncodeToString(sum[:]) != audio.SHA256 || len(data) < 44 || string(data[:4]) != "RIFF" || audio.Duration < 10 {
				t.Fatal("invalid greeting audio", line.ID, mode)
			}
		}
	}
	for _, character := range []string{"misaki", "ayaka", "rina", "shizue", "kathy"} {
		for _, game := range []string{"poker", "blackjack", "baccarat"} {
			if !seen[character+"-"+game] {
				t.Fatal("missing invitation", character, game)
			}
		}
	}
}

func TestMisakiReferenceVoiceProfile(t *testing.T) {
	base, _ := fs.Sub(assets, "frontend")
	web := newSuiteAssets(base)
	paths := []string{"lobby-voice/manifest.json"}
	for _, game := range []string{"poker", "blackjack", "baccarat"} {
		paths = append(paths, game+"/assets/voice/manifest.json", game+"/assets/voice/reward-manifest.json")
	}
	for _, path := range paths {
		data, err := fs.ReadFile(web, path)
		if err != nil {
			t.Fatal(err)
		}
		var obj struct {
			Lines []struct {
				Character string
				Files     map[string]struct {
					PostPitchScale, IntonationScale float64
					StyleID                         int `json:"style_id"`
				}
			}
		}
		if err = json.Unmarshal(data, &obj); err != nil {
			t.Fatal(err)
		}
		for _, line := range obj.Lines {
			if line.Character != "" && line.Character != "misaki" {
				continue
			}
			for _, file := range line.Files {
				if file.PostPitchScale != 1.10 || file.IntonationScale <= 0 || file.StyleID != 1388823424 {
					t.Fatal("MISAKI reference voice profile missing", path, file)
				}
			}
		}
	}
}
