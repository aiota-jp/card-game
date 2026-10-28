package poker

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func parse(s string) int {
	ranks := "23456789TJQKA"
	suits := "shdc"
	r, u := -1, -1
	for i, c := range ranks {
		if byte(c) == s[0] {
			r = i
		}
	}
	for i, c := range suits {
		if byte(c) == s[1] {
			u = i
		}
	}
	if r < 0 || u < 0 {
		panic(s)
	}
	return u*13 + r
}
func hand(ss ...string) []int {
	out := []int{}
	for _, s := range ss {
		out = append(out, parse(s))
	}
	return out
}
func score(ss ...string) int { h := hand(ss...); return Five([5]int(h)) }

func TestHandCategoriesAndKickers(t *testing.T) {
	examples := [][]string{
		{"As", "Kd", "9c", "7h", "3s"},
		{"As", "Ad", "9c", "7h", "3s"},
		{"As", "Ad", "9c", "9h", "3s"},
		{"As", "Ad", "Ac", "7h", "3s"},
		{"As", "2d", "3c", "4h", "5s"},
		{"As", "Js", "9s", "7s", "3s"},
		{"As", "Ad", "Ac", "7h", "7s"},
		{"As", "Ad", "Ac", "Ah", "3s"},
		{"9s", "Ts", "Js", "Qs", "Ks"},
	}
	last := -1
	for category, ss := range examples {
		got := score(ss...)
		if got/759375 != category || got <= last {
			t.Fatalf("category %d: %d", category, got)
		}
		last = got
	}
	if score("As", "2d", "3c", "4h", "5s") >= score("2s", "3d", "4c", "5h", "6s") {
		t.Fatal("wheel must be six-high straight's inferior")
	}
	if score("As", "Ad", "Kc", "Qh", "9s") <= score("Ah", "Ac", "Kd", "Qs", "8s") {
		t.Fatal("last kicker ignored")
	}
	if score("Ks", "Kd", "Qc", "Qh", "As") >= score("Ah", "Ac", "Jd", "Js", "2s") {
		t.Fatal("two pair order")
	}
	if score("As", "Kd", "9c", "7h", "3s") != score("Ah", "Kc", "9d", "7s", "3h") {
		t.Fatal("suits must not break ties")
	}
}

func TestBestOfSeven(t *testing.T) {
	// Two triplets must resolve to aces full of kings.
	got := Best(hand("As", "Ad", "Ac", "Ks", "Kd", "Kc", "2s"))
	if got != score("As", "Ad", "Ac", "Ks", "Kd") {
		t.Fatal("full house from two triplets")
	}
	if Best(hand("As", "Ks", "Qs", "Js", "Ts", "2d", "3h")) != score("As", "Ks", "Qs", "Js", "Ts") {
		t.Fatal("royal flush")
	}
}

func TestAllFiveCardCategoryCounts(t *testing.T) {
	// Exact distribution over all C(52,5) hands. Straight flush includes royals.
	want := [9]int{1302540, 1098240, 123552, 54912, 10200, 5108, 3744, 624, 40}
	var got [9]int
	for a := 0; a < 48; a++ {
		for b := a + 1; b < 49; b++ {
			for c := b + 1; c < 50; c++ {
				for d := c + 1; d < 51; d++ {
					for e := d + 1; e < 52; e++ {
						got[Five([5]int{a, b, c, d, e})/759375]++
					}
				}
			}
		}
	}
	if got != want {
		t.Fatalf("five-card distribution: got %v want %v", got, want)
	}
}

func TestEquityKnownOutcomes(t *testing.T) {
	e := Estimate(hand("As", "Ks"), hand("Qs", "Js", "Ts", "2d", "3h"), 500, rand.New(rand.NewSource(1)))
	if e.Win != 100 || e.Tie != 0 || e.Share != 100 {
		t.Fatalf("unbeatable hand: %+v", e)
	}
	e = Estimate(hand("2h", "3d"), hand("As", "Ks", "Qs", "Js", "Ts"), 500, rand.New(rand.NewSource(2)))
	if e.Tie != 100 || e.Share != 50 {
		t.Fatalf("shared royal: %+v", e)
	}
	e = Estimate(hand("Ah", "Ad"), nil, 2000, rand.New(rand.NewSource(3)))
	if e.Share < 79 || e.Share > 89 {
		t.Fatalf("AA heads-up sanity: %+v", e)
	}
	if math.Abs(e.Win+e.Tie+e.Loss-100) > 1e-8 {
		t.Fatal("probabilities do not sum to 100")
	}
}

func testGame(seed int64) *Game {
	g := &Game{rng: rand.New(rand.NewSource(seed)), stacks: [2]int{1000, 1000}, samples: 80}
	g.Next()
	return g
}
func TestHiddenCardsAndInvalidActions(t *testing.T) {
	g := testGame(42)
	v := g.View()
	if len(v.CPU) != 0 {
		t.Fatal("CPU hand leaked")
	}
	before := g.View()
	if g.Act("invented") == nil {
		t.Fatal("invalid action accepted")
	}
	if !reflect.DeepEqual(before, g.View()) {
		t.Fatal("invalid action mutated game")
	}
	eq := g.eq
	g.hole[1] = []int{1, 2}
	g.updateEquity()
	if !reflect.DeepEqual(eq, g.eq) {
		t.Fatal("equity depends on secret CPU cards")
	}
}

func TestRandomPlayConservesChipsAndTerminates(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	g := testGame(123)
	for round := 0; round < 120; round++ {
		for step := 0; !g.done; step++ {
			if step > 30 {
				t.Fatal("hand failed to terminate")
			}
			v := g.View()
			if len(v.Actions) == 0 {
				t.Fatal("player turn has no actions")
			}
			if err := g.Act(v.Actions[rng.Intn(len(v.Actions))]); err != nil {
				t.Fatal(err)
			}
			total := g.stacks[0] + g.stacks[1]
			if !g.done {
				total += g.pot
			}
			if total != 2000 || g.stacks[0] < 0 || g.stacks[1] < 0 {
				t.Fatalf("chips not conserved: %+v", g.View())
			}
		}
		if len(g.board) > 5 {
			t.Fatal("too many board cards")
		}
		seen := map[int]bool{}
		all := append(append(append([]int{}, g.hole[0]...), g.hole[1]...), g.board...)
		for _, c := range all {
			if seen[c] {
				t.Fatal("duplicate dealt card")
			}
			seen[c] = true
		}
		if g.stacks[0] == 0 || g.stacks[1] == 0 {
			g.stacks = [2]int{1000, 1000}
		}
		if err := g.Next(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestShortStackAllInRunout(t *testing.T) {
	for _, stacks := range [][2]int{{1, 1999}, {1999, 1}, {5, 1995}, {1995, 5}, {13, 1987}, {1987, 13}} {
		g := testGame(19)
		g.done = true
		g.stacks = stacks
		if err := g.Next(); err != nil {
			t.Fatal(err)
		}
		for steps := 0; !g.done; steps++ {
			if steps > 20 {
				t.Fatal("all-in stalled")
			}
			a := "check"
			if g.target > g.bets[0] {
				a = "call"
			}
			if err := g.Act(a); err != nil {
				t.Fatal(err)
			}
		}
		if len(g.board) != 5 || !g.reveal {
			t.Fatal("all-in did not reach showdown")
		}
		if g.stacks[0]+g.stacks[1] != 2000 {
			t.Fatal("all-in lost chips")
		}
	}
}
