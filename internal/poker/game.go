package poker

import (
	"errors"
	"fmt"
	"math/rand"
)

const StartChips = 1000
const BetSize = 20

var streets = []string{"プリフロップ", "フロップ", "ターン", "リバー"}

type Point struct {
	Stage string  `json:"stage"`
	Value float64 `json:"value"`
}
type View struct {
	Hand        int      `json:"hand"`
	Stage       string   `json:"stage"`
	Player      []Card   `json:"player"`
	CPU         []Card   `json:"cpu"`
	Board       []Card   `json:"board"`
	Stacks      [2]int   `json:"stacks"`
	Bets        [2]int   `json:"bets"`
	Pot         int      `json:"pot"`
	Dealer      int      `json:"dealer"`
	Call        int      `json:"call"`
	BetTo       int      `json:"betTo"`
	Actions     []string `json:"actions"`
	Done        bool     `json:"done"`
	CanNext     bool     `json:"canNext"`
	Result      string   `json:"result"`
	HandName    string   `json:"handName"`
	CPUName     string   `json:"cpuName"`
	Equity      Equity   `json:"equity"`
	History     []Point  `json:"history"`
	Log         []string `json:"log"`
	Outcome     string   `json:"outcome"`
	Variant     string   `json:"variant"`
	Title       string   `json:"title"`
	Stages      []string `json:"stages"`
	StageIndex  int      `json:"stageIndex"`
	BoardGame   bool     `json:"boardGame"`
	Drawing     bool     `json:"drawing"`
	CPUHand     []Card   `json:"cpuHand"`
	RuleSummary string   `json:"ruleSummary"`
	EquityNote  string   `json:"equityNote"`
}
type Game struct {
	rng                                             *rand.Rand
	deck                                            []int
	hole                                            [2][]int
	board                                           []int
	stacks                                          [2]int
	bets                                            [2]int
	acted                                           [2]bool
	target, turn, stage, pot, hand, dealer, samples int
	done, reveal                                    bool
	result                                          string
	outcome                                         string
	log                                             []string
	eq                                              Equity
	history                                         []Point
	variant                                         string
	drawing                                         bool
	dead                                            []int
}

func New(seed int64) *Game {
	return NewVariant("texas", seed)
}
func NewVariant(id string, seed int64) *Game {
	g := &Game{rng: rand.New(rand.NewSource(seed)), stacks: [2]int{StartChips, StartChips}, samples: 5000, variant: id}
	g.Next()
	return g
}
func (g *Game) add(s string) { g.log = append(g.log, s) }
func (g *Game) Next() error {
	if g.hand > 0 && !g.done {
		return errors.New("現在のハンドを終了してください")
	}
	if g.stacks[0] == 0 || g.stacks[1] == 0 {
		return errors.New("チップがなくなりました。「セッションをリセット」を選んでください")
	}
	g.hand++
	g.dealer = (g.hand - 1) % 2
	g.deck = g.deckForVariant()
	g.board = []int{}
	g.log = []string{}
	g.history = []Point{}
	g.stage = 0
	g.pot = 0
	g.done = false
	g.reveal = false
	g.result = ""
	g.outcome = ""
	g.bets = [2]int{}
	g.acted = [2]bool{}
	g.drawing = false
	g.dead = nil
	g.hole = [2][]int{}
	for i := 0; i < g.rule().Hole; i++ {
		for p := 0; p < 2; p++ {
			g.hole[p] = append(g.hole[p], g.draw())
		}
	}
	// Match short stacks: no side pots in this heads-up trainer.
	cap := min(10, min(g.stacks[0], g.stacks[1]))
	sb := min(5, cap)
	if g.rule().Stud || g.rule().ID == "short" {
		ante := min(5, min(g.stacks[0], g.stacks[1]))
		for p := 0; p < 2; p++ {
			g.stacks[p] -= ante
			g.pot += ante
		}
		first := g.dealer
		if g.rule().Stud {
			first = g.bringIn()
		}
		forced := min(5, min(g.stacks[0], g.stacks[1]))
		g.pay(first, forced)
		g.target = forced
		g.turn = 1 - first
		g.add(fmt.Sprintf("ハンド #%d · アンティ%d / 強制ベット%d", g.hand, ante, forced))
	} else {
		g.pay(g.dealer, sb)
		g.pay(1-g.dealer, cap)
		g.target = cap
		g.turn = g.dealer
		g.add(fmt.Sprintf("ハンド #%d 開始 · ブラインド %d / %d", g.hand, sb, cap))
	}
	g.updateEquity()
	g.drive()
	return nil
}
func (g *Game) draw() int    { c := g.deck[0]; g.deck = g.deck[1:]; return c }
func (g *Game) pay(p, n int) { g.stacks[p] -= n; g.bets[p] += n; g.pot += n }
func (g *Game) updateEquity() {
	// A separate RNG ensures neither the actual CPU hand nor previous betting
	// changes the visible-card estimate or future shuffled cards.
	seed := int64(17)
	for _, c := range append(append([]int{}, g.hole[0]...), g.board...) {
		seed = seed*53 + int64(c+1)
	}
	for _, c := range append(g.public(1), g.dead...) {
		seed = seed*53 + int64(c+1)
	}
	g.eq = EstimateVariant(g.rule().ID, g.hole[0], g.board, g.public(1), g.dead, g.samples, rand.New(rand.NewSource(seed)))
	g.history = append(g.history, Point{g.rule().Stages[g.stage], g.eq.Share})
}
func (g *Game) betTo() int {
	step := BetSize
	if g.stage == 0 || (g.rule().Stud && g.stage == 1) {
		step = 10
	}
	return min(g.target+step, min(g.bets[0]+g.stacks[0], g.bets[1]+g.stacks[1]))
}
func (g *Game) actions(p int) []string {
	if g.done || g.drawing || g.turn != p {
		return []string{}
	}
	a := []string{"fold"}
	if g.target > g.bets[p] {
		a = append(a, "call")
	} else {
		a = append(a, "check")
	}
	if g.target < 40 && g.betTo() > g.target {
		if g.target == 0 {
			a = append(a, "bet")
		} else {
			a = append(a, "raise")
		}
	}
	return a
}
func (g *Game) Act(action string) error {
	ok := false
	for _, a := range g.actions(0) {
		if a == action {
			ok = true
		}
	}
	if !ok {
		return errors.New("その操作は現在選べません")
	}
	g.step(0, action)
	g.drive()
	return nil
}
func (g *Game) step(p int, action string) {
	who := []string{"あなた", "MISAKI"}[p]
	switch action {
	case "fold":
		g.add(who + "：フォールド")
		g.finish(1-p, false)
		return
	case "check":
		g.acted[p] = true
		g.add(who + "：チェック")
	case "call":
		n := g.target - g.bets[p]
		g.pay(p, n)
		g.acted[p] = true
		g.add(fmt.Sprintf("%s：コール %d", who, n))
	case "bet", "raise":
		n := g.betTo()
		g.pay(p, n-g.bets[p])
		g.target = n
		g.acted[p] = true
		g.acted[1-p] = false
		g.add(fmt.Sprintf("%s：ベット合計 %d", who, n))
	}
	g.turn = 1 - p
}
func (g *Game) drive() {
	for !g.done {
		if g.drawing {
			return
		}
		if g.bets[0] == g.bets[1] && ((g.acted[0] && g.acted[1]) || g.stacks[0] == 0 || g.stacks[1] == 0) {
			if g.stage == len(g.rule().Stages)-1 {
				g.showdown()
				return
			}
			g.stage++
			if g.rule().ID == "draw" {
				g.drawing = true
				g.add("── カード交換：0〜5枚を選択 ──")
				g.updateEquity()
				return
			}
			if g.rule().Stud {
				for p := 0; p < 2; p++ {
					g.hole[p] = append(g.hole[p], g.draw())
				}
			} else {
				n := 1
				if g.stage == 1 {
					n = 3
				}
				for i := 0; i < n; i++ {
					g.board = append(g.board, g.draw())
				}
			}
			g.bets = [2]int{}
			g.acted = [2]bool{}
			g.target = 0
			g.turn = 1 - g.dealer
			if g.rule().Stud {
				g.turn = g.firstExposed()
			}
			g.add("── " + g.rule().Stages[g.stage] + " ──")
			g.updateEquity()
			continue
		}
		if g.turn == 0 {
			return
		}
		g.step(1, g.cpuAction())
	}
}
func (g *Game) cpuAction() string {
	a := g.actions(1)
	has := func(s string) bool {
		for _, x := range a {
			if x == s {
				return true
			}
		}
		return false
	}
	// Lightweight heuristic CPU, not a trained or optimal poker agent.
	strength := .35
	if g.rule().ID == "razz" {
		sum := 0
		seen := map[int]bool{}
		for _, c := range g.hole[1] {
			r := lowRank(c)
			sum += r
			if seen[r] {
				sum += 5
			}
			seen[r] = true
		}
		strength = 1 - float64(sum)/float64(len(g.hole[1])*17)
	} else if !g.rule().Board || len(g.board) >= 3 {
		value := Score(g.rule().ID, g.hole[1], g.board)
		strength = .25 + float64(value/rankBase)*.14
		if len(g.hole[1]) < 5 {
			strength = .45
		}
	} else if len(g.board) == 0 {
		x, y := g.hole[1][0], g.hole[1][1]
		strength = float64(x%13+y%13+4) / 40
		if x%13 == y%13 {
			strength += .3
		}
		if x/13 == y/13 {
			strength += .08
		}
	}
	if has("call") {
		if strength < .42 && g.rng.Float64() < .22 {
			return "fold"
		}
		return "call"
	}
	if has("bet") && g.rng.Float64() < strength*.6 {
		return "bet"
	}
	if has("raise") && strength > .65 && g.rng.Float64() < .2 {
		return "raise"
	}
	return "check"
}
func (g *Game) showdown() {
	a := Score(g.rule().ID, g.hole[0], g.board)
	b := Score(g.rule().ID, g.hole[1], g.board)
	winner := -1
	if a > b {
		winner = 0
	} else if b > a {
		winner = 1
	}
	g.finish(winner, true)
}
func (g *Game) finish(winner int, reveal bool) {
	g.done = true
	g.reveal = reveal
	g.outcome = "tie"
	if winner == 0 {
		g.outcome = "win"
	} else if winner == 1 {
		g.outcome = "loss"
	}
	if winner < 0 {
		g.stacks[0] += g.pot / 2
		g.stacks[1] += g.pot / 2
		g.stacks[1-g.dealer] += g.pot % 2
		g.result = "引き分け · ポットを分配"
	} else {
		g.stacks[winner] += g.pot
		g.result = fmt.Sprintf("%sの勝ち · %d チップ獲得", []string{"あなた", "MISAKI"}[winner], g.pot)
	}
	g.add(g.result)
}
func (g *Game) View() View {
	cpu := []Card{}
	cpuName := "非公開"
	if g.reveal {
		cpu = cards(g.hole[1])
		cpuName = scoreName(g.rule().ID, g.hole[1], g.board)
	}
	stage := g.rule().Stages[g.stage]
	if g.done {
		stage = "ハンド終了"
	}
	player := cards(g.hole[0])
	display := make([]Card, 0, len(g.hole[1]))
	for i, c := range g.hole[1] {
		up := g.rule().Stud && i >= 2 && i <= 5
		value := card(-1)
		if up || g.reveal {
			value = card(c)
		}
		value.Up = up
		display = append(display, value)
	}
	if g.rule().Stud {
		for i := range player {
			player[i].Up = i >= 2 && i <= 5
		}
		if !g.reveal {
			cpuName = fmt.Sprintf("公開%d枚 / 非公開%d枚", len(g.public(1)), len(g.hole[1])-len(g.public(1)))
		}
	}
	note := "相手1人・ランダムな手札を想定。MISAKIの実際の非公開手札は計算に使いません。"
	if g.rule().Stud {
		note = "MISAKIの公開カードも反映。残りの非公開手札と今後のカードをランダムに仮定します。"
	}
	if g.rule().ID == "draw" {
		note = "現在の5枚を維持した場合のショーダウン勝率。交換前は交換後の予測ではありません。相手の交換方針は推定しません。"
	}
	return View{Hand: g.hand, Stage: stage, Player: player, CPU: cpu, Board: cards(g.board), Stacks: g.stacks, Bets: g.bets, Pot: g.pot, Dealer: g.dealer, Call: max(0, g.target-g.bets[0]), BetTo: g.betTo(), Actions: g.actions(0), Done: g.done, CanNext: g.done && g.stacks[0] > 0 && g.stacks[1] > 0, Result: g.result, HandName: scoreName(g.rule().ID, g.hole[0], g.board), CPUName: cpuName, Equity: g.eq, History: append([]Point{}, g.history...), Log: append([]string{}, g.log...), Outcome: g.outcome, Variant: g.rule().ID, Title: g.rule().Title, Stages: g.rule().Stages, StageIndex: g.stage, BoardGame: g.rule().Board, Drawing: g.drawing, CPUHand: display, RuleSummary: g.rule().Summary, EquityNote: note}
}
