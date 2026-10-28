package main

import (
	"cardlounge/internal/baccarat"
	"fmt"
	"sync"
	"time"
)

type BaccaratApp struct {
	mu       sync.Mutex
	game     *baccarat.Game
	sessions map[string]*baccarat.Game
	host     string
	bank     bool
}

func NewBaccaratApp() *BaccaratApp {
	g := baccarat.New("punto", time.Now().UnixNano())
	return &BaccaratApp{game: g, sessions: map[string]*baccarat.Game{"punto:misaki": g}, host: "misaki"}
}
func (a *BaccaratApp) OpenGame(id, host string) (baccarat.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if host != "misaki" && host != "ayaka" && host != "rina" && host != "shizue" && host != "kathy" {
		return baccarat.View{}, fmt.Errorf("進行役が見つかりません")
	}
	if !baccarat.Valid(id) {
		return baccarat.View{}, fmt.Errorf("未対応のゲームです")
	}
	key := id + ":" + host
	if a.sessions[key] == nil {
		g := baccarat.New(id, time.Now().UnixNano())
		if err := g.Configure(host, false); err != nil {
			return baccarat.View{}, err
		}
		a.sessions[key] = g
	}
	g := a.sessions[key]
	a.game = g
	a.host = host
	a.bank = g.View().Role == "banker"
	return g.View(), nil
}
func (a *BaccaratApp) SetRole(bank bool) (baccarat.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Configure(a.host, bank); err != nil {
		return baccarat.View{}, err
	}
	a.bank = bank
	return a.game.View(), nil
}
func (a *BaccaratApp) Snapshot() baccarat.View {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.game.View()
}
func (a *BaccaratApp) Deal(bet int, target string) (baccarat.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Deal(bet, target); err != nil {
		return baccarat.View{}, err
	}
	return a.game.View(), nil
}
func (a *BaccaratApp) Act(action string) (baccarat.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Act(action); err != nil {
		return baccarat.View{}, err
	}
	return a.game.View(), nil
}
func (a *BaccaratApp) NextHand() (baccarat.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Next(); err != nil {
		return baccarat.View{}, err
	}
	return a.game.View(), nil
}
func (a *BaccaratApp) Reset() baccarat.View {
	a.mu.Lock()
	defer a.mu.Unlock()
	v := a.game.View()
	g := baccarat.New(v.Variant, time.Now().UnixNano())
	g.Configure(a.host, a.bank)
	a.sessions[v.Variant+":"+a.host] = g
	a.game = g
	return g.View()
}
