package poker

import "sort"

type Card struct {
	ID     int    `json:"id"`
	Rank   string `json:"rank"`
	Suit   string `json:"suit"`
	Red    bool   `json:"red"`
	Hidden bool   `json:"hidden"`
	Up     bool   `json:"up"`
}

func card(c int) Card {
	if c < 0 {
		return Card{ID: -1, Hidden: true}
	}
	return Card{ID: c, Rank: []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}[c%13], Suit: []string{"♠", "♥", "♦", "♣"}[c/13], Red: c/13 == 1 || c/13 == 2}
}
func cards(cs []int) []Card {
	out := make([]Card, 0, len(cs))
	for _, c := range cs {
		out = append(out, card(c))
	}
	return out
}

var names = []string{"ハイカード", "ワンペア", "ツーペア", "スリーカード", "ストレート", "フラッシュ", "フルハウス", "フォーカード", "ストレートフラッシュ"}

// Five produces a lexicographically comparable rank, including every kicker.
func Five(cs [5]int) int {
	var counts [15]int
	flush := true
	for _, c := range cs {
		counts[c%13+2]++
		if c/13 != cs[0]/13 {
			flush = false
		}
	}
	straight, run := 0, 0
	for r := 2; r <= 14; r++ {
		if counts[r] > 0 {
			run++
		} else {
			run = 0
		}
		if run >= 5 {
			straight = r
		}
	}
	if straight == 0 && counts[14] > 0 && counts[2] > 0 && counts[3] > 0 && counts[4] > 0 && counts[5] > 0 {
		straight = 5
	}
	encode := func(category int, ranks ...int) int {
		v := category
		for i := 0; i < 5; i++ {
			v *= 15
			if i < len(ranks) {
				v += ranks[i]
			}
		}
		return v
	}
	if flush && straight > 0 {
		return encode(8, straight)
	}
	// Sort ranks by group size, then rank. This also orders all kickers.
	groups := make([]int, 0, 5)
	for r := 14; r >= 2; r-- {
		if counts[r] > 0 {
			groups = append(groups, r)
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		a, b := groups[i], groups[j]
		if counts[a] != counts[b] {
			return counts[a] > counts[b]
		}
		return a > b
	})
	a := counts[groups[0]]
	if a == 4 {
		return encode(7, groups...)
	}
	if a == 3 && counts[groups[1]] == 2 {
		return encode(6, groups...)
	}
	if flush {
		return encode(5, groups...)
	}
	if straight > 0 {
		return encode(4, straight)
	}
	if a == 3 {
		return encode(3, groups...)
	}
	if a == 2 && counts[groups[1]] == 2 {
		return encode(2, groups...)
	}
	if a == 2 {
		return encode(1, groups...)
	}
	return encode(0, groups...)
}

func Best(cs []int) int {
	best := 0
	for a := 0; a < len(cs)-4; a++ {
		for b := a + 1; b < len(cs)-3; b++ {
			for c := b + 1; c < len(cs)-2; c++ {
				for d := c + 1; d < len(cs)-1; d++ {
					for e := d + 1; e < len(cs); e++ {
						v := Five([5]int{cs[a], cs[b], cs[c], cs[d], cs[e]})
						if v > best {
							best = v
						}
					}
				}
			}
		}
	}
	return best
}
func HandName(cs []int) string {
	if len(cs) < 5 {
		if len(cs) == 2 && cs[0]%13 == cs[1]%13 {
			return "ポケットペア"
		}
		return "役は共有カード公開後に判定"
	}
	return names[Best(cs)/759375]
}
