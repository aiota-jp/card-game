package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"os"
	"strings"
)

type SystemApp struct{ ctx context.Context }

func (a *SystemApp) startup(ctx context.Context) { a.ctx = ctx }

type preferencesBackup struct {
	Format   string            `json:"format"`
	Settings map[string]string `json:"settings"`
}

func validPreference(key, value string) bool {
	switch key {
	case "cardlounge.lounge.character":
		return value == "misaki" || value == "ayaka" || value == "rina" || value == "shizue" || value == "kathy"
	case "cardlounge.lounge.voice", "cardlounge.lounge.expression", "cardlounge.sound.enabled", "cardlounge.music.enabled":
		return value == "true" || value == "false"
	case "cardlounge.lounge.aivis-default-v1", "cardlounge.lounge.rate-default-v2", "cardlounge.lounge.rate-default-v3", "cardlounge.lounge.volume-default-v2":
		return value == "1"
	case "cardlounge.lounge.volume", "cardlounge.sound.volume", "cardlounge.music.volume", "cardlounge.lounge.rate":
		var literal any
		if json.Unmarshal([]byte(value), &literal) != nil {
			return false
		}
		n, numeric := literal.(float64)
		if !numeric {
			return false
		}
		max := 100.0
		min := 0.0
		if strings.HasSuffix(key, ".rate") {
			min = .8
			max = 2.5
		}
		return n >= min && n <= max
	}
	if strings.HasPrefix(key, "poker.tutorial.completed.v1.") {
		return len(key) < 100 && value == "1"
	}
	return false
}
func validatePreferences(raw string) error {
	if len(raw) > 1048576 {
		return errors.New("バックアップが大きすぎます")
	}
	var b preferencesBackup
	if json.Unmarshal([]byte(raw), &b) != nil || b.Format != "CardGameLoungeBackup/1" || b.Settings == nil || len(b.Settings) > 100 {
		return errors.New("このゲームの有効なバックアップではありません")
	}
	for k, v := range b.Settings {
		if !validPreference(k, v) {
			return fmt.Errorf("設定の値が不正です: %s", k)
		}
	}
	return nil
}
func (a *SystemApp) SaveBackup(raw string) (string, error) {
	if err := validatePreferences(raw); err != nil {
		return "", err
	}
	path, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{Title: "バックアップを保存", DefaultFilename: "card-game-lounge-backup.json", Filters: []wruntime.FileFilter{{DisplayName: "JSON バックアップ", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return "", err
	}
	if err = os.WriteFile(path, []byte(raw), 0600); err != nil {
		return "", err
	}
	return path, nil
}
func (a *SystemApp) LoadBackup() (string, error) {
	path, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{Title: "バックアップを復元", Filters: []wruntime.FileFilter{{DisplayName: "JSON バックアップ", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 1048577))
	if err != nil {
		return "", err
	}
	if err = validatePreferences(string(raw)); err != nil {
		return "", err
	}
	return string(raw), nil
}
func (a *SystemApp) Exit() { wruntime.Quit(a.ctx) }
