package poker

import (
	"math"
	"math/rand"
)

type Equity struct {
	Win     float64 `json:"win"`
	Tie     float64 `json:"tie"`
	Loss    float64 `json:"loss"`
	Share   float64 `json:"share"`
	Samples int     `json:"samples"`
	Margin  float64 `json:"margin"`
}

// Estimate deliberately accepts only visible cards. The actual CPU hand/deck
// cannot leak into the estimate. Assumption: one uniformly random opponent.
func Estimate(hole, board []int, trials int, rng *rand.Rand) Equity {
	var known [52]bool
	for _, c := range append(append([]int{}, hole...), board...) {
		known[c] = true
	}
	deck := make([]int, 0, 50)
	for c := 0; c < 52; c++ {
		if !known[c] {
			deck = append(deck, c)
		}
	}
	win, tie := 0, 0
	var own, other [7]int
	copy(own[:2], hole)
	copy(own[2:], board)
	copy(other[2:], board)
	need := 2 + 5 - len(board)
	for i := 0; i < trials; i++ {
		// Partial Fisher-Yates: sampling without replacement.
		for j := 0; j < need; j++ {
			k := j + rng.Intn(len(deck)-j)
			deck[j], deck[k] = deck[k], deck[j]
		}
		other[0], other[1] = deck[0], deck[1]
		for j := len(board); j < 5; j++ {
			own[j+2] = deck[2+j-len(board)]
			other[j+2] = own[j+2]
		}
		a, b := Best(own[:]), Best(other[:])
		if a > b {
			win++
		} else if a == b {
			tie++
		}
	}
	return equityCounts(win, tie, trials)
}
func equityCounts(win, tie, trials int) Equity {
	n := float64(trials)
	// Conservative 95% Monte Carlo error bound for any bounded [0,1] result.
	return Equity{100 * float64(win) / n, 100 * float64(tie) / n, 100 * float64(trials-win-tie) / n, 100 * (float64(win) + float64(tie)/2) / n, trials, 98 / math.Sqrt(n)}
}
