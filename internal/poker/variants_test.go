package poker

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestVariantRanks(t *testing.T) {
	// A board-only royal is not an Omaha royal: exactly two hole cards required.
	board := hand("As", "Ks", "Qs", "Js", "Ts")
	if Score("omaha", hand("2h", "3h", "4d", "5d"), board) >= encode(8, 14) {
		t.Fatal("Omaha used zero hole cards")
	}
	if Score("omaha", hand("As", "Ks", "2h", "3d"), hand("Qs", "Js", "Ts", "9c", "9d")) != encode(8, 14) {
		t.Fatal("Omaha royal missing")
	}
	flush := [5]int(hand("As", "Js", "9s", "8s", "6s"))
	full := [5]int(hand("Ah", "Ad", "Ac", "Kh", "Kd"))
	if FiveShort(flush) <= FiveShort(full) || Five(flush) >= Five(full) {
		t.Fatal("short-deck flush order")
	}
	if FiveShort([5]int(hand("As", "6h", "7d", "8c", "9s"))) != encode(4, 9) {
		t.Fatal("short-deck ace-low straight")
	}
	if FiveShort([5]int(hand("As", "6s", "7s", "8s", "9s"))) != encode(8, 9) {
		t.Fatal("short-deck ace-low straight flush")
	}
	wheel := [5]int(hand("As", "2s", "3s", "4s", "5s"))
	six := [5]int(hand("As", "2h", "3d", "4c", "6s"))
	if FiveLow(wheel) <= FiveLow(six) {
		t.Fatal("Razz wheel should beat six low; flush/straight ignored")
	}
	if FiveLow([5]int(hand("As", "Ah", "2d", "3c", "4s"))) >= FiveLow([5]int(hand("Kh", "Qh", "Jh", "Th", "9h"))) {
		t.Fatal("Razz pair should lose to unpaired king low")
	}
	if scoreName("razz", wheel[:], nil) != "5-4-3-2-A ロー" {
		t.Fatal("Razz low name")
	}
}
func newTestVariant(id string, seed int64) *Game {
	g := &Game{variant: id, rng: rand.New(rand.NewSource(seed)), stacks: [2]int{1000, 1000}, samples: 60}
	g.Next()
	return g
}
func TestAllVariantsLegalPlay(t *testing.T) {
	for _, id := range []string{"texas", "omaha", "draw", "stud", "razz", "short"} {
		t.Run(id, func(t *testing.T) {
			g := newTestVariant(id, 31)
			rng := rand.New(rand.NewSource(44))
			for round := 0; round < 45; round++ {
				for step := 0; !g.done; step++ {
					if step > 50 {
						t.Fatal("hand did not terminate")
					}
					v := g.View()
					if v.Variant != id {
						t.Fatal("wrong variant")
					}
					if v.Drawing {
						if err := g.Draw(DrawChoice(g.hole[0])); err != nil {
							t.Fatal(err)
						}
					} else {
						a := g.actions(0)
						if len(a) == 0 {
							t.Fatal("stalled without actions")
						}
						choice := a[rng.Intn(len(a))]
						if err := g.Act(choice); err != nil {
							t.Fatal(err)
						}
					}
					total := g.stacks[0] + g.stacks[1]
					if !g.done {
						total += g.pot
					}
					if total != 2000 || g.stacks[0] < 0 || g.stacks[1] < 0 {
						t.Fatal("lost chips")
					}
				}
				seen := map[int]bool{}
				for _, cs := range [][]int{g.hole[0], g.hole[1], g.board, g.dead} {
					for _, c := range cs {
						if seen[c] {
							t.Fatal("duplicate card")
						}
						seen[c] = true
						if id == "short" && c%13 < 4 {
							t.Fatal("2-5 dealt in short deck")
						}
					}
				}
				if g.stacks[0] == 0 || g.stacks[1] == 0 {
					g.stacks = [2]int{1000, 1000}
				}
				if err := g.Next(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
func TestVariantHiddenInformation(t *testing.T) {
	for _, id := range []string{"omaha", "draw", "stud", "razz", "short"} {
		g := newTestVariant(id, 9)
		v := g.View()
		for i, c := range v.CPUHand {
			visible := g.rule().Stud && i == 2
			if visible {
				if c.Hidden || c.ID < 0 {
					t.Fatal("up card hidden")
				}
			} else if !c.Hidden || c.ID != -1 || c.Rank != "" || c.Suit != "" {
				t.Fatal("secret card leaked")
			}
		}
		g.updateEquity()
		before := g.eq
		g.hole[1][0], g.hole[1][1] = 50, 49
		g.updateEquity()
		if before != g.eq {
			t.Fatalf("%s equity leaked CPU hole cards", id)
		}
	}
}
func TestDrawValidationAndStandPat(t *testing.T) {
	g := newTestVariant("draw", 21)
	for !g.drawing && !g.done {
		action := "check"
		if g.target > g.bets[0] {
			action = "call"
		}
		g.Act(action)
	}
	if g.done {
		t.Skip("CPU folded before exchange")
	}
	before := g.View()
	for _, indices := range [][]int{{0, 0}, {-1}, {5}, {0, 1, 2, 3, 4, 5}} {
		if g.Draw(indices) == nil {
			t.Fatal("invalid exchange accepted")
		}
		if !reflect.DeepEqual(before, g.View()) {
			t.Fatal("invalid draw mutated state")
		}
	}
	own := append([]int{}, g.hole[0]...)
	if err := g.Draw([]int{}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(own, g.hole[0]) {
		t.Fatal("stand pat changed cards")
	}
	if g.Draw(nil) == nil {
		t.Fatal("second draw accepted")
	}
}
func TestVariantAllInAndShowdown(t *testing.T) {
	for _, id := range []string{"omaha", "draw", "stud", "razz", "short"} {
		for _, stacks := range [][2]int{{1, 1999}, {1999, 1}, {13, 1987}} {
			g := newTestVariant(id, 55)
			g.done = true
			g.stacks = stacks
			g.Next()
			for step := 0; !g.done; step++ {
				if step > 30 {
					t.Fatal("all-in stalled")
				}
				if g.drawing {
					g.Draw([]int{})
				} else {
					action := "check"
					if g.target > g.bets[0] {
						action = "call"
					}
					g.Act(action)
				}
			}
			if g.stacks[0]+g.stacks[1] != 2000 {
				t.Fatal("all-in lost chips")
			}
			if g.reveal {
				expected := g.rule().Hole
				if g.rule().Stud {
					expected = 7
				}
				if len(g.hole[0]) != expected {
					t.Fatal("wrong final card count")
				}
			}
		}
	}
}
