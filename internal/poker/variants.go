package poker

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
)

const rankBase = 759375

type Variant struct {
	ID, Title, Summary string
	Stages             []string
	Hole               int
	Board              bool
	Stud               bool
}

func Rule(id string) (Variant, bool) {
	v := map[string]Variant{
		"texas": {"texas", "テキサスホールデム", "手札2枚と共有5枚から最強の5枚。ブラインド5/10。", []string{"プリフロップ", "フロップ", "ターン", "リバー"}, 2, true, false},
		"omaha": {"omaha", "オマハ", "ハイのみ。手札4枚のうち必ず2枚、共有5枚のうち必ず3枚を使用。ブラインド5/10。", []string{"プリフロップ", "フロップ", "ターン", "リバー"}, 4, true, false},
		"short": {"short", "ショートデック", "6〜Aの36枚。フラッシュはフルハウスより上。A-6-7-8-9もストレート。アンティ5＋ボタンブラインド5。", []string{"プリフロップ", "フロップ", "ターン", "リバー"}, 2, true, false},
		"draw":  {"draw", "ファイブカードドロー", "手札5枚。交換は1回、0〜5枚。交換前後にベット。ブラインド5/10。", []string{"交換前", "カード交換", "交換後"}, 5, false, false},
		"stud":  {"stud", "セブンカードスタッド", "最初に裏2枚＋表1枚。4〜6枚目は表、7枚目は裏。最強の5枚。アンティ5・ブリングイン5。", []string{"3枚目", "4枚目", "5枚目", "6枚目", "7枚目"}, 3, false, true},
		"razz":  {"razz", "ラズ", "A-to-5ロー。Aは低く、ストレート・フラッシュは無視。最良は5-4-3-2-A。アンティ5・ブリングイン5。", []string{"3枚目", "4枚目", "5枚目", "6枚目", "7枚目"}, 3, false, true},
	}
	value, ok := v[id]
	return value, ok
}
func (g *Game) rule() Variant {
	v, ok := Rule(g.variant)
	if !ok {
		v, _ = Rule("texas")
	}
	return v
}
func (g *Game) deckForVariant() []int {
	out := []int{}
	for c := 0; c < 52; c++ {
		if g.rule().ID != "short" || c%13 >= 4 {
			out = append(out, c)
		}
	}
	g.rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}
func encode(category int, ranks ...int) int {
	value := category
	for i := 0; i < 5; i++ {
		value *= 15
		if i < len(ranks) {
			value += ranks[i]
		}
	}
	return value
}
func FiveShort(cs [5]int) int {
	value := Five(cs)
	cat := value / rankBase
	mask := 0
	flush := true
	for _, c := range cs {
		mask |= 1 << (c%13 + 2)
		flush = flush && c/13 == cs[0]/13
	}
	if mask == (1<<14 | 1<<6 | 1<<7 | 1<<8 | 1<<9) {
		if flush {
			return encode(8, 9)
		}
		return encode(4, 9)
	}
	if cat == 5 {
		return value + rankBase
	}
	if cat == 6 {
		return value - rankBase
	}
	return value
}
func lowRank(c int) int {
	r := c%13 + 2
	if r == 14 {
		return 1
	}
	return r
}
func lowKey(cs []int) int {
	var count [15]int
	for _, c := range cs {
		count[lowRank(c)]++
	}
	groups := []int{}
	for r := 13; r >= 1; r-- {
		if count[r] > 0 {
			groups = append(groups, r)
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		a, b := groups[i], groups[j]
		if count[a] != count[b] {
			return count[a] > count[b]
		}
		return a > b
	})
	cat := 0
	if count[groups[0]] == 4 {
		cat = 5
	} else if count[groups[0]] == 3 {
		cat = 3
		if len(groups) > 1 && count[groups[1]] == 2 {
			cat = 4
		}
	} else if count[groups[0]] == 2 {
		cat = 1
		if len(groups) > 1 && count[groups[1]] == 2 {
			cat = 2
		}
	}
	return encode(cat, groups...)
}
func FiveLow(cs [5]int) int { return 7*rankBase - lowKey(cs[:]) }
func bestWith(cs []int, five func([5]int) int) int {
	best := 0
	for a := 0; a < len(cs)-4; a++ {
		for b := a + 1; b < len(cs)-3; b++ {
			for c := b + 1; c < len(cs)-2; c++ {
				for d := c + 1; d < len(cs)-1; d++ {
					for e := d + 1; e < len(cs); e++ {
						best = max(best, five([5]int{cs[a], cs[b], cs[c], cs[d], cs[e]}))
					}
				}
			}
		}
	}
	return best
}
func Score(id string, hole, board []int) int {
	switch id {
	case "omaha":
		best := 0
		for a := 0; a < len(hole)-1; a++ {
			for b := a + 1; b < len(hole); b++ {
				for c := 0; c < len(board)-2; c++ {
					for d := c + 1; d < len(board)-1; d++ {
						for e := d + 1; e < len(board); e++ {
							best = max(best, Five([5]int{hole[a], hole[b], board[c], board[d], board[e]}))
						}
					}
				}
			}
		}
		return best
	case "short":
		return bestWith(append(append([]int{}, hole...), board...), FiveShort)
	case "razz":
		return bestWith(hole, FiveLow)
	default:
		return Best(append(append([]int{}, hole...), board...))
	}
}
func scoreName(id string, hole, board []int) string {
	if id == "omaha" && len(board) < 3 {
		return "共有カード3枚から役を判定"
	}
	if len(hole)+len(board) < 5 {
		if id == "razz" {
			return "ローを作成中（Aは低い）"
		}
		return "5枚がそろうと役を判定"
	}
	value := Score(id, hole, board)
	if id == "razz" {
		key := 7*rankBase - value
		cat := key / rankBase
		if cat > 0 {
			return []string{"", "ワンペア", "ツーペア", "スリーカード", "フルハウス", "フォーカード"}[cat] + "（ロー）"
		}
		ranks := make([]string, 5)
		for i := 4; i >= 0; i-- {
			r := key % 15
			key /= 15
			ranks[i] = []string{"", "A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}[r]
		}
		return strings.Join(ranks, "-") + " ロー"
	}
	cat := value / rankBase
	if id == "short" {
		if cat == 5 {
			return "フルハウス"
		}
		if cat == 6 {
			return "フラッシュ"
		}
	}
	return names[cat]
}
func (g *Game) public(p int) []int {
	out := []int{}
	if !g.rule().Stud {
		return out
	}
	for i, c := range g.hole[p] {
		if i >= 2 && i <= 5 {
			out = append(out, c)
		}
	}
	return out
}
func exposed(cs []int, low bool) int {
	if low {
		return 7*rankBase - lowKey(cs)
	}
	return -lowKeyHigh(cs)
}
func lowKeyHigh(cs []int) int {
	// For up to four exposed cards, only pairs/trips/quads and kickers matter.
	var count [15]int
	for _, c := range cs {
		count[c%13+2]++
	}
	ranks := []int{}
	for r := 14; r >= 2; r-- {
		if count[r] > 0 {
			ranks = append(ranks, r)
		}
	}
	sort.Slice(ranks, func(i, j int) bool {
		a, b := ranks[i], ranks[j]
		if count[a] != count[b] {
			return count[a] > count[b]
		}
		return a > b
	})
	cat := 0
	if count[ranks[0]] == 4 {
		cat = 4
	} else if count[ranks[0]] == 3 {
		cat = 3
	} else if count[ranks[0]] == 2 {
		cat = 1
		if len(ranks) > 1 && count[ranks[1]] == 2 {
			cat = 2
		}
	}
	return -encode(cat, ranks...)
}
func (g *Game) firstExposed() int {
	a, b := exposed(g.public(0), g.rule().ID == "razz"), exposed(g.public(1), g.rule().ID == "razz")
	if a == b {
		return 1 - g.dealer
	}
	if a > b {
		return 0
	}
	return 1
}
func (g *Game) bringIn() int {
	rank := func(c int) int {
		r := c%13 + 2
		if g.rule().ID == "razz" {
			r = lowRank(c)
		}
		return r*4 + 3 - c/13
	}
	a, b := rank(g.hole[0][2]), rank(g.hole[1][2])
	if g.rule().ID == "razz" {
		if a > b {
			return 0
		}
		return 1
	}
	if a < b {
		return 0
	}
	return 1
}
func DrawChoice(hand []int) []int {
	var count [15]int
	var suits [4]int
	for _, c := range hand {
		count[c%13+2]++
		suits[c/13]++
	}
	if Best(hand)/rankBase >= 4 {
		return []int{}
	}
	keep := map[int]bool{}
	group := false
	for i, c := range hand {
		if count[c%13+2] >= 2 {
			keep[i] = true
			group = true
		}
	}
	if !group {
		for suit, n := range suits {
			if n == 4 {
				for i, c := range hand {
					if c/13 == suit {
						keep[i] = true
					}
				}
				group = true
				break
			}
		}
	}
	if !group {
		for i, c := range hand {
			if c%13+2 >= 12 {
				keep[i] = true
			}
		}
	}
	discard := []int{}
	for i := range hand {
		if !keep[i] {
			discard = append(discard, i)
		}
	}
	return discard
}
func EstimateVariant(id string, hole, board, up, dead []int, trials int, rng *rand.Rand) Equity {
	v, ok := Rule(id)
	if !ok {
		v, _ = Rule("texas")
	}
	var known [52]bool
	for _, list := range [][]int{hole, board, up, dead} {
		for _, c := range list {
			known[c] = true
		}
	}
	deck := []int{}
	for c := 0; c < 52; c++ {
		if !known[c] && (id != "short" || c%13 >= 4) {
			deck = append(deck, c)
		}
	}
	total := v.Hole
	if v.Stud {
		total = 7
	}
	own := make([]int, total)
	other := make([]int, total)
	copy(own, hole)
	copy(other, up)
	finalBoard := make([]int, 0, 5)
	if v.Board {
		finalBoard = make([]int, 5)
		copy(finalBoard, board)
	}
	needOwn := total - len(hole)
	needOther := total - len(up)
	needBoard := len(finalBoard) - len(board)
	need := needOwn + needOther + needBoard
	win, tie := 0, 0
	for i := 0; i < trials; i++ {
		for j := 0; j < need; j++ {
			k := j + rng.Intn(len(deck)-j)
			deck[j], deck[k] = deck[k], deck[j]
		}
		cursor := 0
		copy(own[len(hole):], deck[cursor:cursor+needOwn])
		cursor += needOwn
		copy(other[len(up):], deck[cursor:cursor+needOther])
		cursor += needOther
		if needBoard > 0 {
			copy(finalBoard[len(board):], deck[cursor:cursor+needBoard])
		}
		a, b := Score(id, own, finalBoard), Score(id, other, finalBoard)
		if a > b {
			win++
		} else if a == b {
			tie++
		}
	}
	return equityCounts(win, tie, trials)
}
func (g *Game) Draw(indices []int) error {
	if !g.drawing || g.done || g.rule().ID != "draw" {
		return fmt.Errorf("今はカード交換できません")
	}
	selected := map[int]bool{}
	for _, i := range indices {
		if i < 0 || i >= 5 || selected[i] {
			return fmt.Errorf("交換カードの選択が不正です")
		}
		selected[i] = true
	}
	if len(indices) > 5 {
		return fmt.Errorf("交換は5枚までです")
	}
	// Decide the CPU's exchange before any replacement is dealt.
	cpu := DrawChoice(g.hole[1])
	for i := 0; i < 5; i++ {
		if selected[i] {
			g.dead = append(g.dead, g.hole[0][i])
			g.hole[0][i] = g.draw()
		}
	}
	for _, i := range cpu {
		g.hole[1][i] = g.draw()
	}
	g.add(fmt.Sprintf("あなた：%d枚交換 / MISAKI：%d枚交換", len(indices), len(cpu)))
	g.drawing = false
	g.stage = 2
	g.bets = [2]int{}
	g.acted = [2]bool{}
	g.target = 0
	g.turn = 1 - g.dealer
	g.updateEquity()
	g.drive()
	return nil
}
