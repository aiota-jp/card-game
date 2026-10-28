package main

import (
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"io/fs"
	"os"
	"strings"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// Allow the requested launch greeting inside this application only.
	_ = os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", strings.TrimSpace(os.Getenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS")+" --autoplay-policy=no-user-gesture-required"))
	platformPrepare()
	window := newWindowController()
	system := &SystemApp{}
	web, err := fs.Sub(assets, "frontend")
	if err != nil {
		panic(err)
	}
	if err := wails.Run(&options.App{
		Title: "CARD GAME LOUNGE — MISAKI / AYAKA / RINA / SHIZUE / KATHY", Width: 1024, Height: 700, StartHidden: true,
		MinWidth: 320, MinHeight: 240, BackgroundColour: &options.RGBA{R: 12, G: 18, B: 29, A: 255},
		Windows:   &windows.Options{WindowClassName: "CardGameLoungeWindow"},
		OnStartup: system.startup, OnDomReady: window.ready, OnBeforeClose: window.beforeClose, OnShutdown: window.shutdown,
		AssetServer: &assetserver.Options{Assets: newSuiteAssets(web)}, Bind: []interface{}{system, NewPokerApp(), NewBlackjackApp(), NewBaccaratApp()},
	}); err != nil {
		panic(err)
	}
}
