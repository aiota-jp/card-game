package baccarat

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func stacked(t *testing.T, id string, values ...int) *Game {
	t.Helper()
	g := New(id, 7)
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
			t.Fatal("missing rank")
		}
		front = append(front, deck[found])
		deck = append(deck[:found], deck[found+1:]...)
	}
	g.shoe = append(front, deck...)
	return g
}
func deal(t *testing.T, g *Game, target string) {
	t.Helper()
	if e := g.Deal(50, target); e != nil {
		t.Fatal(e)
	}
}
func sum(g *Game) int {
	n := 0
	for _, b := range g.balances {
		n += b
	}
	return n
}
func TestPointsAndDecks(t *testing.T) {
	if Points([]Card{{Value: 7}, {Value: 6}}) != 3 || Points([]Card{{Value: 0}, {Value: 1}}) != 1 {
		t.Fatal("points")
	}
	for _, tc := range []struct {
		id    string
		decks int
	}{{"punto", 8}, {"chemin", 6}, {"banque", 3}} {
		g := New(tc.id, 1)
		zeros := 0
		for _, c := range g.shoe {
			if c.Value == 0 {
				zeros++
			}
		}
		if len(g.shoe) != 52*tc.decks || zeros != 16*tc.decks {
			t.Fatal(tc.id, "composition")
		}
	}
}
func TestBankerThirdCardTable(t *testing.T) {
	for b := 0; b < 10; b++ {
		for p := -1; p < 10; p++ {
			want := false
			if p < 0 {
				want = b <= 5
			} else {
				switch b {
				case 0, 1, 2:
					want = true
				case 3:
					want = p != 8
				case 4:
					want = p >= 2 && p <= 7
				case 5:
					want = p >= 4 && p <= 7
				case 6:
					want = p == 6 || p == 7
				}
			}
			if BankerDraw(b, p) != want {
				t.Fatalf("bank=%d third=%d", b, p)
			}
		}
	}
	for b := 0; b < 8; b++ {
		for p := -1; p < 10; p++ {
			if BankerOptional(b, p) != (b == 3 && p == 9 || b == 5 && p == 4) {
				t.Fatal("optional table")
			}
		}
	}
}
func TestPuntoPayoutsAndNaturals(t *testing.T) {
	for _, tc := range []struct {
		cards      []int
		target     string
		chips, fee int
	}{{[]int{4, 2, 4, 3}, "player", 105000, 0}, {[]int{4, 4, 4, 5}, "banker", 104750, 250}, {[]int{4, 3, 4, 5}, "player", 100000, 0}, {[]int{4, 3, 4, 5}, "banker", 100000, 0}, {[]int{4, 3, 4, 5}, "tie", 140000, 0}, {[]int{4, 4, 4, 5}, "tie", 95000, 0}} {
		g := stacked(t, "punto", tc.cards...)
		deal(t, g, tc.target)
		if g.phase != "done" || g.balances[0] != tc.chips || g.commission != tc.fee || len(g.bankCards) != 2 || len(g.playerCards[0]) != 2 {
			t.Fatalf("payout: %+v", g.View())
		}
	}
}
func TestPuntoAutomaticDraw(t *testing.T) {
	g := stacked(t, "punto", 2, 3, 3, 2, 4, 2)
	deal(t, g, "player")
	if g.phase != "dealt" || Points(g.playerCards[0]) != 5 {
		t.Fatal("initial phase")
	}
	if e := g.Act("hit"); e == nil {
		t.Fatal("manual hit in punto")
	}
	if e := g.Act("continue"); e != nil {
		t.Fatal(e)
	}
	if Points(g.playerCards[0]) != 9 || Points(g.bankCards) != 7 || len(g.bankCards) != 3 || g.balances[0] != 105000 {
		t.Fatal("third cards")
	}
}
func TestCheminChoiceAndRotation(t *testing.T) {
	g := stacked(t, "chemin", 4, 3, 1, 3, 2)
	deal(t, g, "player")
	v := g.View()
	if g.phase != "player" || v.Hands[0].Points != nil || !v.Hands[0].Cards[0].Hidden || *v.Hands[1].Points != 5 {
		t.Fatal("hidden hands or 5 choice")
	}
	if e := g.Act("hit"); e != nil {
		t.Fatal(e)
	}
	if g.balances != [4]int{105000, 85000, 105000, 105000} || g.commission != 0 || !g.rotate {
		t.Fatalf("shared stakes: %v", g.balances)
	}
	if e := g.Next(); e != nil {
		t.Fatal(e)
	}
	if g.banker != 2 {
		t.Fatal("banker did not rotate")
	}
	g = stacked(t, "chemin", 4, 3, 1, 3)
	deal(t, g, "player")
	g.Act("stand")
	if g.balances != [4]int{95000, 114250, 95000, 95000} || g.commission != 750 || g.rotate {
		t.Fatal("bank commission or continuation")
	}
}
func TestCheminHumanBankerOptionalAndRetire(t *testing.T) {
	g := stacked(t, "chemin", 2, 3, 0, 0, 9)
	g.banker = 0
	deal(t, g, "player")
	if g.phase != "banker" {
		t.Fatal("3 vs third9 not optional")
	}
	v := g.View()
	if v.Role != "banker" || v.Hands[1].Points != nil || !v.Hands[1].Cards[0].Hidden || v.Hands[1].Cards[2].Value != 9 {
		t.Fatal("public third or private hand")
	}
	g.Act("stand")
	if g.balances[0] != 114250 || g.commission != 750 {
		t.Fatal("banker settlement")
	}
	g.Next()
	if e := g.Act("retire"); e != nil || g.banker != 1 {
		t.Fatal("retire")
	}
	g = stacked(t, "chemin", 4, 3, 1, 3)
	g.banker = 3
	g.owners[0] = 0
	deal(t, g, "player")
	g.Act("hit")
	if g.rotate {
		g.Next()
		if g.banker != 0 {
			t.Fatal("rotation to human")
		}
	}
}
func TestBanqueBothHandsAndNaturalStopsAll(t *testing.T) {
	g := stacked(t, "banque", 4, 2, 3, 1, 2, 3, 2)
	deal(t, g, "player")
	if g.phase != "player" || len(g.View().Hands) != 3 {
		t.Fatal("two player hands")
	}
	g.Act("stand")
	if g.balances != [4]int{95000, 104900, 100000, 100000} || g.commission != 100 {
		t.Fatalf("separate settlement: %v fee %d", g.balances, g.commission)
	}
	if e := g.Next(); e != nil || g.banker != 1 {
		t.Fatal("fixed banker")
	}
	g = stacked(t, "banque", 2, 4, 3, 3, 5, 4)
	deal(t, g, "player")
	if g.phase != "done" || len(g.shoe) != 150 || g.balances != [4]int{95000, 100000, 105000, 100000} || g.commission != 0 {
		t.Fatal("left natural must stop all hands")
	}
}
func TestBanqueHumanBankFreedomAndNetCommission(t *testing.T) {
	for _, hit := range []bool{false, true} {
		g := stacked(t, "banque", 2, 3, 2, 2, 3, 3, 6, 4)
		if e := g.Configure("misaki", true); e != nil {
			t.Fatal(e)
		}
		deal(t, g, "player")
		if g.phase != "banker" || g.banker != 0 || g.View().Hands[1].Points != nil {
			t.Fatal("bank freedom/privacy")
		}
		action := "stand"
		want := 100000
		if hit {
			action = "hit"
			want = 109800
		}
		if e := g.Act(action); e != nil {
			t.Fatal(e)
		}
		if g.balances[0] != want {
			t.Fatalf("action %s chips %d", action, g.balances[0])
		}
		fee := 0
		if hit {
			fee = 200
		}
		if g.commission != fee {
			t.Fatal("net commission")
		}
	}
}
func TestInvalidActionsAndFundingAreAtomic(t *testing.T) {
	for _, id := range []string{"punto", "chemin", "banque"} {
		g := New(id, 2)
		before := g.balances
		for _, bet := range []int{-1, 0, 30, 5000} {
			if g.Deal(bet, "player") == nil || g.balances != before || g.hand != 0 {
				t.Fatal("invalid bet changed state")
			}
		}
		if g.Act("hit") == nil || g.Next() == nil {
			t.Fatal("invalid phase")
		}
		if id == "banque" {
			g.balances[g.banker] = 3000
			before = g.balances
			if g.Deal(20, "player") == nil || g.balances != before {
				t.Fatal("underfunded bank")
			}
		}
	}
	g := New("punto", 1)
	if g.Deal(50, "bogus") == nil || g.hand != 0 {
		t.Fatal("invalid target")
	}
}
func TestPuntoPredictionUsesCountsNotOrder(t *testing.T) {
	g := New("punto", 11)
	a := g.View()
	clone := *g
	clone.shoe = append([]Card{}, g.shoe...)
	for i, j := 0, len(clone.shoe)-1; i < j; i, j = i+1, j-1 {
		clone.shoe[i], clone.shoe[j] = clone.shoe[j], clone.shoe[i]
	}
	b := clone.View()
	if !reflect.DeepEqual(a.Odds, b.Odds) || math.Abs(a.Odds.Player+a.Odds.Banker+a.Odds.Tie-100) > 1e-8 {
		t.Fatal("prediction leaked future order")
	}
	if !reflect.DeepEqual(a, g.View()) {
		t.Fatal("view mutated state")
	}
}
func TestGoalAndCutCard(t *testing.T) {
	g := stacked(t, "punto", 4, 2, 4, 3)
	g.balances[0] = 195000
	deal(t, g, "player")
	if !g.cleared || g.Next() == nil {
		t.Fatal("goal")
	}
	for _, id := range []string{"punto", "chemin", "banque"} {
		g = New(id, 3)
		g.shoe = g.shoe[:16]
		if e := g.Deal(20, "player"); e != nil {
			t.Fatal(e)
		}
		if len(g.shoe) < g.decks*52-9 {
			t.Fatal("shuffle threshold")
		}
	}
}
func TestLongRunMoneyAndCardInvariants(t *testing.T) {
	rng := rand.New(rand.NewSource(39))
	for _, id := range []string{"punto", "chemin", "banque"} {
		for session := 0; session < 20; session++ {
			g := New(id, int64(session))
			if id == "banque" {
				g.Configure("misaki", session%2 == 0)
			}
			for hand := 0; hand < 250; hand++ {
				if !g.ready(g.banker) || g.maxBet() < Minimum {
					break
				}
				before := sum(g)
				start := g.balances[0]
				if e := g.Deal(20, "banker"); e != nil {
					t.Fatal(id, e)
				}
				for steps := 0; g.phase != "done" && steps < 5; steps++ {
					action := "stand"
					if g.phase == "dealt" {
						action = "continue"
					} else if rng.Intn(2) == 0 {
						action = "hit"
					}
					if e := g.Act(action); e != nil {
						t.Fatal(id, e)
					}
				}
				if g.phase != "done" || g.profit != g.balances[0]-start || len(g.bankCards) > 3 || g.wins+g.losses+g.ties != g.hand {
					t.Fatal(id, "state invariant")
				}
				for _, c := range g.playerCards {
					if len(c) > 3 {
						t.Fatal("too many cards")
					}
				}
				for _, b := range g.balances {
					if b < 0 {
						t.Fatal(id, "negative balance")
					}
				}
				if id != "punto" && sum(g) != before-g.commission {
					t.Fatalf("%s lost money: before %d after %d fee %d", id, before, sum(g), g.commission)
				}
				if g.cleared || !g.ready(g.futureBanker()) {
					break
				}
				if e := g.Next(); e != nil {
					t.Fatal(id, e)
				}
			}
		}
	}
}
