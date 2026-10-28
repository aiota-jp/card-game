package blackjack

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func stacked(t *testing.T, values ...int) *Game {
	t.Helper()
	g := New(7)
	deck := append([]Card{}, g.shoe...)
	front := []Card{}
	for _, v := range values {
		found := -1
		for i, c := range deck {
			if c.Value == v {
				found = i
				break
			}
		}
		if found < 0 {
			t.Fatal("no rank", v)
		}
		front = append(front, deck[found])
		deck = append(deck[:found], deck[found+1:]...)
	}
	g.shoe = append(front, deck...)
	return g
}
func deal(t *testing.T, g *Game) {
	t.Helper()
	if err := g.Deal(50); err != nil {
		t.Fatal(err)
	}
}
func TestAceTotals(t *testing.T) {
	for _, tc := range []struct {
		cards []int
		total int
		soft  bool
	}{{[]int{1, 6}, 17, true}, {[]int{1, 1, 9}, 21, true}, {[]int{1, 1, 9, 10}, 21, false}, {[]int{1, 10, 10}, 21, false}, {[]int{1, 1, 1, 9}, 12, false}, {[]int{10, 9, 5}, 24, false}} {
		got, soft := totalValues(tc.cards)
		if got != tc.total || soft != tc.soft {
			t.Fatalf("%v => %d/%v", tc.cards, got, soft)
		}
	}
}
func TestNaturalsAndPayouts(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cards   []int
		outcome string
		chips   int
	}{{"player", []int{1, 9, 10, 7}, "win", 1075}, {"dealer", []int{9, 10, 8, 1}, "loss", 950}, {"both", []int{1, 10, 10, 1}, "tie", 1000}} {
		t.Run(tc.name, func(t *testing.T) {
			g := stacked(t, tc.cards...)
			deal(t, g)
			if g.phase != "done" || g.outcome != tc.outcome || g.chips != tc.chips {
				t.Fatalf("%+v", g.View())
			}
		})
	}
}
func TestDealerStandsOnSoft17(t *testing.T) {
	g := stacked(t, 10, 1, 7, 6, 10)
	deal(t, g)
	left := len(g.shoe)
	if err := g.Act("stand"); err != nil {
		t.Fatal(err)
	}
	if len(g.shoe) != left || g.outcome != "tie" || g.chips != 1000 {
		t.Fatal("soft17 was hit or settled incorrectly")
	}
}
func TestDealerDrawAndPlayerBust(t *testing.T) {
	g := stacked(t, 10, 10, 8, 6, 5)
	deal(t, g)
	g.Act("stand")
	d, _ := Total(g.dealer)
	if d != 21 || g.outcome != "loss" || g.chips != 950 {
		t.Fatal("dealer did not draw to 21")
	}
	g = stacked(t, 10, 9, 9, 7, 5)
	deal(t, g)
	g.Act("hit")
	if g.phase != "done" || g.outcome != "loss" || g.chips != 950 || !g.revealed || len(g.dealer) != 2 {
		t.Fatal("player bust settlement")
	}
}
func TestDoubleDown(t *testing.T) {
	g := stacked(t, 5, 9, 6, 7, 10, 2)
	deal(t, g)
	if err := g.Act("double"); err != nil {
		t.Fatal(err)
	}
	if g.bet != 100 || g.chips != 1100 || len(g.player) != 3 || g.phase != "done" {
		t.Fatalf("double win: %+v", g.View())
	}
	g = stacked(t, 10, 9, 6, 7, 10)
	deal(t, g)
	g.Act("double")
	if g.chips != 900 || g.profit != -100 {
		t.Fatal("double bust debit")
	}
	g = stacked(t, 2, 9, 3, 7, 4)
	deal(t, g)
	g.Act("hit")
	chips, left := g.chips, len(g.shoe)
	if g.Act("double") == nil || g.chips != chips || len(g.shoe) != left {
		t.Fatal("late double accepted")
	}
	g = stacked(t, 5, 9, 6, 7)
	g.chips = 50
	deal(t, g)
	if g.Act("double") == nil || g.chips != 0 {
		t.Fatal("unfunded double accepted")
	}
}
func TestInvalidActionsDoNotMutate(t *testing.T) {
	g := New(2)
	for _, bet := range []int{0, -10, 15, 5000} {
		if g.Deal(bet) == nil || g.chips != 1000 || g.hand != 0 || len(g.shoe) != 312 {
			t.Fatal("invalid wager changed state")
		}
	}
	if g.Act("hit") == nil || g.Next() == nil {
		t.Fatal("invalid phase accepted")
	}
	deal(t, g)
	left := len(g.shoe)
	if g.Deal(50) == nil || len(g.shoe) != left {
		t.Fatal("mid-hand wager")
	}
	if g.Act("split") == nil {
		t.Fatal("unsupported split")
	}
}
func TestViewAndOddsDoNotLeakHoleCard(t *testing.T) {
	g := stacked(t, 10, 10, 6, 7, 2)
	deal(t, g)
	a := g.View()
	if !a.Dealer[1].Hidden || a.Dealer[1].Rank != "" || a.DealerTotal != 0 {
		t.Fatal("hole leaked")
	}
	clone := *g
	clone.dealer = append([]Card{}, g.dealer...)
	clone.shoe = append([]Card{}, g.shoe...)
	clone.dealer[1], clone.shoe[0] = clone.shoe[0], clone.dealer[1]
	b := clone.View()
	if !reflect.DeepEqual(a.Odds, b.Odds) {
		t.Fatal("prediction used actual hole")
	}
	if math.Abs(a.Odds.Win+a.Odds.Tie+a.Odds.Loss-100) > .00001 {
		t.Fatal("probabilities do not sum to 100")
	}
	if a.Odds.Bust <= 0 || a.Odds.Bust >= 100 {
		t.Fatal("invalid bust probability")
	}
	if !reflect.DeepEqual(a, g.View()) {
		t.Fatal("view changed random state")
	}
}
func TestExactBustProbabilityWithPeek(t *testing.T) {
	g := stacked(t, 10, 10, 6, 7)
	deal(t, g)
	g.known = [11]int{}
	g.known[1] = 2
	g.known[2] = 1
	g.known[10] = 1
	odds := g.odds()
	if math.Abs(odds.Bust-100.0/6) > 1e-9 {
		t.Fatalf("got %v want 16.6667", odds.Bust)
	}
}
func TestGoalAndNewHand(t *testing.T) {
	g := stacked(t, 1, 9, 10, 7)
	g.chips = 1950
	deal(t, g)
	if !g.cleared || g.chips != 2025 || g.Next() == nil || g.Deal(50) == nil {
		t.Fatal("goal not terminal")
	}
	g = stacked(t, 10, 9, 9, 7, 10)
	deal(t, g)
	g.Act("stand")
	chips := g.chips
	if err := g.Next(); err != nil {
		t.Fatal(err)
	}
	if g.phase != "betting" || len(g.player) != 0 || len(g.dealer) != 0 || g.chips != chips {
		t.Fatal("next hand reset bankroll")
	}
}
func TestShuffleAndLongRunInvariants(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	g := New(17)
	total := 0
	for i := 1; i <= 10; i++ {
		total += g.known[i]
	}
	if total != 312 || g.known[10] != 96 {
		t.Fatal("six deck composition")
	}
	g.shoe = g.shoe[:78]
	if err := g.Deal(10); err != nil {
		t.Fatal(err)
	}
	if !g.shuffled || len(g.shoe) != 308 {
		t.Fatal("cut-card shuffle")
	}
	for session := 0; session < 20; session++ {
		g = New(int64(session))
		for hand := 0; hand < 400; hand++ {
			if g.chips < 10 || g.cleared {
				break
			}
			before := g.chips
			bet := 10
			if before >= 50 {
				bet = 50
			}
			if err := g.Deal(bet); err != nil {
				t.Fatal(err)
			}
			for g.phase == "playing" {
				p, _ := Total(g.player)
				action := "stand"
				if p < 16 {
					action = "hit"
				}
				if len(g.player) == 2 && g.chips >= g.bet && rng.Intn(8) == 0 {
					action = "double"
				}
				if err := g.Act(action); err != nil {
					t.Fatal(err)
				}
			}
			if g.chips < 0 || g.chips != before+g.profit || g.profit < -g.bet || g.wins+g.losses+g.ties != g.hand {
				t.Fatalf("accounting failure hand %d", g.hand)
			}
			if !g.cleared && g.chips >= 10 {
				if err := g.Next(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
