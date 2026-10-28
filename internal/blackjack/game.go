package blackjack

import (
	"fmt"
	"math"
	"math/rand"
)

const StartingChips = 1000
const Goal = 2000
const Samples = 4000

var Bets = []int{10, 20, 50, 100, 200}

type Card struct {
	Rank   string `json:"rank"`
	Suit   string `json:"suit"`
	Red    bool   `json:"red"`
	Hidden bool   `json:"hidden,omitempty"`
	Value  int    `json:"-"`
}
type Odds struct {
	Bust       float64 `json:"bust"`
	Win        float64 `json:"win"`
	Tie        float64 `json:"tie"`
	Loss       float64 `json:"loss"`
	DealerBust float64 `json:"dealerBust"`
	Samples    int     `json:"samples"`
	Margin     float64 `json:"margin"`
}
type View struct {
	Variant     string   `json:"variant"`
	Hand        int      `json:"hand"`
	Phase       string   `json:"phase"`
	Player      []Card   `json:"player"`
	Dealer      []Card   `json:"dealer"`
	Total       int      `json:"total"`
	Soft        bool     `json:"soft"`
	DealerTotal int      `json:"dealerTotal"`
	DealerSoft  bool     `json:"dealerSoft"`
	Chips       int      `json:"chips"`
	Bet         int      `json:"bet"`
	Profit      int      `json:"profit"`
	Goal        int      `json:"goal"`
	Cleared     bool     `json:"cleared"`
	Done        bool     `json:"done"`
	CanNext     bool     `json:"canNext"`
	Outcome     string   `json:"outcome"`
	Result      string   `json:"result"`
	Actions     []string `json:"actions"`
	Odds        Odds     `json:"odds"`
	Remaining   int      `json:"remaining"`
	Shuffled    bool     `json:"shuffled"`
	Wins        int      `json:"wins"`
	Losses      int      `json:"losses"`
	Ties        int      `json:"ties"`
	Log         []string `json:"log"`
	Stages      []string `json:"stages"`
	StageIndex  int      `json:"stageIndex"`
}
type Game struct {
	rng                                          *rand.Rand
	shoe                                         []Card
	player, dealer                               []Card
	known                                        [11]int
	revealed                                     bool
	chips, bet, hand, profit, wins, losses, ties int
	phase, outcome, result                       string
	log                                          []string
	cleared, shuffled                            bool
}

func New(seed int64) *Game {
	g := &Game{rng: rand.New(rand.NewSource(seed)), chips: StartingChips, phase: "betting"}
	g.shuffle()
	return g
}
func (g *Game) shuffle() {
	g.shoe = nil
	g.known = [11]int{}
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	for d := 0; d < 6; d++ {
		for s, suit := range []string{"♠", "♥", "♣", "♦"} {
			for i, rank := range ranks {
				v := min(i+1, 10)
				g.shoe = append(g.shoe, Card{Rank: rank, Suit: suit, Red: s == 1 || s == 3, Value: v})
				g.known[v]++
			}
		}
	}
	g.rng.Shuffle(len(g.shoe), func(i, j int) { g.shoe[i], g.shoe[j] = g.shoe[j], g.shoe[i] })
	g.shuffled = true
}
func Total(cards []Card) (int, bool) {
	sum, aces := 0, 0
	for _, c := range cards {
		sum += c.Value
		if c.Value == 1 {
			aces++
		}
	}
	soft := aces > 0 && sum+10 <= 21
	if soft {
		sum += 10
	}
	return sum, soft
}
func totalValues(cards []int) (int, bool) {
	sum, aces := 0, 0
	for _, v := range cards {
		sum += v
		if v == 1 {
			aces++
		}
	}
	soft := aces > 0 && sum+10 <= 21
	if soft {
		sum += 10
	}
	return sum, soft
}
func (g *Game) draw(visible bool) Card {
	c := g.shoe[0]
	g.shoe = g.shoe[1:]
	if visible {
		g.known[c.Value]--
	}
	return c
}
func (g *Game) reveal() {
	if !g.revealed && len(g.dealer) > 1 {
		g.known[g.dealer[1].Value]--
		g.revealed = true
	}
}
func (g *Game) Deal(bet int) error {
	if g.phase == "playing" {
		return fmt.Errorf("進行中のハンドを先に終えてください")
	}
	if g.cleared {
		return fmt.Errorf("目標達成済みです。新しいセッションで遊べます")
	}
	allowed := false
	for _, n := range Bets {
		allowed = allowed || n == bet
	}
	if !allowed || bet > g.chips {
		return fmt.Errorf("持ち点の範囲でベット額を選んでください")
	}
	g.shuffled = false
	if len(g.shoe) <= 78 {
		g.shuffle()
	}
	g.hand++
	g.bet = bet
	g.chips -= bet
	g.profit = 0
	g.phase = "playing"
	g.outcome = ""
	g.result = ""
	g.revealed = false
	g.player = nil
	g.dealer = nil
	g.log = []string{fmt.Sprintf("%dチップをベット", bet)}
	if g.shuffled {
		g.log = append(g.log, "6デッキのシューをシャッフルしました")
	}
	g.player = append(g.player, g.draw(true))
	g.dealer = append(g.dealer, g.draw(true))
	g.player = append(g.player, g.draw(true))
	g.dealer = append(g.dealer, g.draw(false))
	p, _ := Total(g.player)
	d, _ := Total(g.dealer)
	if p == 21 || d == 21 {
		g.reveal()
		if p == 21 && d == 21 {
			g.settle("tie", bet, "双方ブラックジャック · 引き分け")
		} else if p == 21 {
			g.settle("win", bet*5/2, "ブラックジャック！ 3:2配当")
		} else {
			g.settle("loss", 0, "ディーラーがブラックジャック")
		}
	}
	return nil
}
func (g *Game) Act(action string) error {
	if g.phase != "playing" {
		return fmt.Errorf("ベットしてカードを配ってください")
	}
	switch action {
	case "hit":
		g.player = append(g.player, g.draw(true))
		g.log = append(g.log, "ヒット · カードを1枚引きました")
		p, _ := Total(g.player)
		if p > 21 {
			g.reveal()
			g.settle("loss", 0, "バースト · 21を超えました")
		} else if p == 21 {
			g.finishDealer()
		}
	case "stand":
		g.log = append(g.log, "スタンド · この合計で勝負")
		g.finishDealer()
	case "double":
		if len(g.player) != 2 || g.chips < g.bet {
			return fmt.Errorf("ダブルダウンは最初の2枚と追加の持ち点が必要です")
		}
		g.chips -= g.bet
		g.bet *= 2
		g.player = append(g.player, g.draw(true))
		g.log = append(g.log, "ダブルダウン · ベットを2倍にして1枚だけ引きました")
		p, _ := Total(g.player)
		if p > 21 {
			g.reveal()
			g.settle("loss", 0, "ダブルダウンでバースト")
		} else {
			g.finishDealer()
		}
	default:
		return fmt.Errorf("未対応の操作です")
	}
	return nil
}
func (g *Game) finishDealer() {
	g.reveal()
	for {
		d, _ := Total(g.dealer)
		if d >= 17 {
			break
		}
		g.dealer = append(g.dealer, g.draw(true))
	}
	p, _ := Total(g.player)
	d, _ := Total(g.dealer)
	if d > 21 {
		g.settle("win", g.bet*2, "ディーラーがバースト · あなたの勝利")
	} else if p > d {
		g.settle("win", g.bet*2, "あなたの合計が上 · あなたの勝利")
	} else if p == d {
		g.settle("tie", g.bet, "同じ合計 · ベットを返却")
	} else {
		g.settle("loss", 0, "ディーラーの合計が上")
	}
}
func (g *Game) settle(outcome string, returned int, message string) {
	g.chips += returned
	g.profit = returned - g.bet
	g.outcome = outcome
	g.result = message
	g.phase = "done"
	switch outcome {
	case "win":
		g.wins++
	case "loss":
		g.losses++
	default:
		g.ties++
	}
	g.cleared = g.chips >= Goal
	g.log = append(g.log, fmt.Sprintf("%s · 収支 %+dチップ", message, g.profit))
	if g.cleared {
		g.log = append(g.log, "2,000チップ達成！ セッションクリア")
	}
}
func (g *Game) Next() error {
	if g.phase != "done" {
		return fmt.Errorf("ハンドを終えてから次へ進めます")
	}
	if g.cleared || g.chips < 10 {
		return fmt.Errorf("新しいセッションで再開してください")
	}
	g.phase = "betting"
	g.player = nil
	g.dealer = nil
	g.result = ""
	g.outcome = ""
	g.bet = 0
	g.profit = 0
	g.log = nil
	return nil
}
func (g *Game) View() View {
	p, soft := Total(g.player)
	d, ds := Total(g.dealer)
	dealer := append([]Card{}, g.dealer...)
	if !g.revealed && len(dealer) > 1 {
		dealer[1] = Card{Hidden: true}
		d = 0
		ds = false
	}
	actions := []string{}
	if g.phase == "playing" {
		actions = []string{"hit", "stand"}
		if len(g.player) == 2 && g.chips >= g.bet {
			actions = append(actions, "double")
		}
	}
	odds := Odds{}
	if g.phase == "playing" {
		odds = g.odds()
	}
	stage := 0
	if g.phase == "done" {
		stage = 1
	}
	return View{Variant: "blackjack", Hand: g.hand, Phase: g.phase, Player: append([]Card{}, g.player...), Dealer: dealer, Total: p, Soft: soft, DealerTotal: d, DealerSoft: ds, Chips: g.chips, Bet: g.bet, Profit: g.profit, Goal: Goal, Cleared: g.cleared, Done: g.phase == "done", CanNext: g.phase == "done" && !g.cleared && g.chips >= 10, Outcome: g.outcome, Result: g.result, Actions: actions, Odds: odds, Remaining: len(g.shoe), Shuffled: g.shuffled, Wins: g.wins, Losses: g.losses, Ties: g.ties, Log: append([]string{}, g.log...), Stages: []string{"プレイ", "結果"}, StageIndex: stage}
}

// The hole card is included among unknown cards. Peek constraints condition
// both the next-card probability and simulations without using its real rank.
func (g *Game) holeAllowed(v int) bool {
	up := g.dealer[0].Value
	return !(up == 1 && v == 10 || up == 10 && v == 1)
}
func pick(counts *[11]int, rng *rand.Rand, allowed func(int) bool) int {
	n := 0
	for v := 1; v <= 10; v++ {
		if allowed == nil || allowed(v) {
			n += counts[v]
		}
	}
	if n == 0 {
		return 0
	}
	roll := rng.Intn(n)
	for v := 1; v <= 10; v++ {
		if allowed != nil && !allowed(v) {
			continue
		}
		roll -= counts[v]
		if roll < 0 {
			counts[v]--
			return v
		}
	}
	panic("unreachable")
}
func (g *Game) odds() Odds {
	var n, allowed int
	var seed int64 = 31
	for v := 1; v <= 10; v++ {
		n += g.known[v]
		seed = seed*37 + int64(g.known[v])
		if g.holeAllowed(v) {
			allowed += g.known[v]
		}
	}
	for _, c := range g.player {
		seed = seed*17 + int64(c.Value)
	}
	seed = seed*17 + int64(g.dealer[0].Value)
	values := []int{}
	for _, c := range g.player {
		values = append(values, c.Value)
	}
	bust := 0.0
	if n > 1 && allowed > 0 {
		for v := 1; v <= 10; v++ {
			probHole := 0.0
			if g.holeAllowed(v) {
				probHole = float64(g.known[v]) / float64(allowed)
			}
			t, _ := totalValues(append(append([]int{}, values...), v))
			if t > 21 {
				bust += (float64(g.known[v]) - probHole) / float64(n-1)
			}
		}
	}
	p, _ := Total(g.player)
	wins, ties, losses, dealerBust := 0, 0, 0, 0
	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < Samples; i++ {
		pool := g.known
		hole := pick(&pool, rng, g.holeAllowed)
		cards := []int{g.dealer[0].Value, hole}
		for {
			t, _ := totalValues(cards)
			if t >= 17 {
				break
			}
			v := pick(&pool, rng, nil)
			if v == 0 {
				break
			}
			cards = append(cards, v)
		}
		d, _ := totalValues(cards)
		if d > 21 {
			dealerBust++
			wins++
		} else if p > d {
			wins++
		} else if p == d {
			ties++
		} else {
			losses++
		}
	}
	return Odds{Bust: math.Max(0, math.Min(100, bust*100)), Win: float64(wins) * 100 / Samples, Tie: float64(ties) * 100 / Samples, Loss: float64(losses) * 100 / Samples, DealerBust: float64(dealerBust) * 100 / Samples, Samples: Samples, Margin: 1.96 * math.Sqrt(.25/Samples) * 100}
}
