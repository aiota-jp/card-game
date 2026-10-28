package main

import (
	"cardlounge/internal/poker"
	"fmt"
	"sync"
	"time"
)

type PokerApp struct {
	mu       sync.Mutex
	game     *poker.Game
	sessions map[string]*poker.Game
}

func NewPokerApp() *PokerApp {
	game := poker.New(time.Now().UnixNano())
	return &PokerApp{game: game, sessions: map[string]*poker.Game{"texas": game}}
}
func (a *PokerApp) OpenGame(id string) (poker.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := poker.Rule(id); !ok {
		return poker.View{}, fmt.Errorf("未対応のゲームです")
	}
	if a.sessions[id] == nil {
		a.sessions[id] = poker.NewVariant(id, time.Now().UnixNano())
	}
	a.game = a.sessions[id]
	return a.game.View(), nil
}
func (a *PokerApp) Draw(indices []int) (poker.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Draw(indices); err != nil {
		return poker.View{}, err
	}
	return a.game.View(), nil
}
func (a *PokerApp) Snapshot() poker.View { a.mu.Lock(); defer a.mu.Unlock(); return a.game.View() }
func (a *PokerApp) Act(action string) (poker.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Act(action); err != nil {
		return poker.View{}, err
	}
	return a.game.View(), nil
}
func (a *PokerApp) NextHand() (poker.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Next(); err != nil {
		return poker.View{}, err
	}
	return a.game.View(), nil
}
func (a *PokerApp) Reset() poker.View {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := a.game.View().Variant
	a.game = poker.NewVariant(id, time.Now().UnixNano())
	a.sessions[id] = a.game
	return a.game.View()
}
