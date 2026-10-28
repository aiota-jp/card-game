package main

import (
	"cardlounge/internal/blackjack"
	"sync"
	"time"
)

type BlackjackApp struct {
	mu   sync.Mutex
	game *blackjack.Game
}

func NewBlackjackApp() *BlackjackApp {
	return &BlackjackApp{game: blackjack.New(time.Now().UnixNano())}
}
func (a *BlackjackApp) Snapshot() blackjack.View {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.game.View()
}
func (a *BlackjackApp) Deal(bet int) (blackjack.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Deal(bet); err != nil {
		return blackjack.View{}, err
	}
	return a.game.View(), nil
}
func (a *BlackjackApp) Act(action string) (blackjack.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Act(action); err != nil {
		return blackjack.View{}, err
	}
	return a.game.View(), nil
}
func (a *BlackjackApp) NextHand() (blackjack.View, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.game.Next(); err != nil {
		return blackjack.View{}, err
	}
	return a.game.View(), nil
}
func (a *BlackjackApp) Reset() blackjack.View {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.game = blackjack.New(time.Now().UnixNano())
	return a.game.View()
}
