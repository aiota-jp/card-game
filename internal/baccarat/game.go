package baccarat

import (
	"fmt"
	"math"
	"math/rand"
)

const Unit = 100
const Initial = 1000 * Unit
const Goal = 2000 * Unit
const Minimum = 20 * Unit
const Samples = 3000

var Bets = []int{20, 50, 100, 200}
var names = []string{"あなた", "MISAKI", "AYAKA", "RINA"}
var ids = []string{"you", "misaki", "ayaka", "rina"}

type Card struct {
	Rank   string `json:"rank"`
	Suit   string `json:"suit"`
	Red    bool   `json:"red"`
	Hidden bool   `json:"hidden,omitempty"`
	Value  int    `json:"-"`
}
type Seat struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Chips  float64 `json:"chips"`
	Stake  float64 `json:"stake"`
	Banker bool    `json:"banker"`
}
type Hand struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Owner  string  `json:"owner"`
	Cards  []Card  `json:"cards"`
	Points *int    `json:"points"`
	Stake  float64 `json:"stake"`
	Result string  `json:"result"`
}
type Odds struct {
	Player  float64 `json:"player"`
	Banker  float64 `json:"banker"`
	Tie     float64 `json:"tie"`
	Samples int     `json:"samples"`
	Margin  float64 `json:"margin"`
}
type Record struct {
	Hand    int     `json:"hand"`
	Outcome string  `json:"outcome"`
	Profit  float64 `json:"profit"`
}
type View struct {
	Variant    string   `json:"variant"`
	Name       string   `json:"name"`
	Phase      string   `json:"phase"`
	Hand       int      `json:"hand"`
	Role       string   `json:"role"`
	Banker     int      `json:"banker"`
	Seats      []Seat   `json:"seats"`
	Hands      []Hand   `json:"hands"`
	Chips      float64  `json:"chips"`
	Goal       float64  `json:"goal"`
	Bet        float64  `json:"bet"`
	MaxBet     float64  `json:"maxBet"`
	Profit     float64  `json:"profit"`
	Commission float64  `json:"commission"`
	Target     string   `json:"target"`
	Done       bool     `json:"done"`
	Cleared    bool     `json:"cleared"`
	CanNext    bool     `json:"canNext"`
	CanDeal    bool     `json:"canDeal"`
	Actions    []string `json:"actions"`
	Hint       string   `json:"hint"`
	Outcome    string   `json:"outcome"`
	Result     string   `json:"result"`
	Pending    string   `json:"pending"`
	Log        []string `json:"log"`
	History    []Record `json:"history"`
	Remaining  int      `json:"remaining"`
	Decks      int      `json:"decks"`
	Odds       *Odds    `json:"odds"`
	Wins       int      `json:"wins"`
	Losses     int      `json:"losses"`
	Ties       int      `json:"ties"`
	Stages     []string `json:"stages"`
	StageIndex int      `json:"stageIndex"`
}
type Game struct {
	names                                 [4]string
	ids                                   [4]string
	rng                                   *rand.Rand
	variant                               string
	decks                                 int
	shoe                                  []Card
	balances                              [4]int
	stakes                                [4]int
	banker, preferred                     int
	humanBank                             bool
	bankCards                             []Card
	playerCards                           [2][]Card
	owners                                [2]int
	decided                               [2]bool
	pending, hand                         int
	phase, target, outcome, result        string
	wager, profit, commission, startHuman int
	revealed, cleared, rotate             bool
	log                                   []string
	history                               []Record
	wins, losses, ties                    int
}

func Valid(id string) bool { return id == "punto" || id == "chemin" || id == "banque" }
func Name(id string) string {
	return map[string]string{"punto": "プントバンコ", "chemin": "シュマン・ド・フェール", "banque": "バカラ・バンク"}[id]
}
func New(id string, seed int64) *Game {
	if !Valid(id) {
		panic("invalid variant")
	}
	g := &Game{rng: rand.New(rand.NewSource(seed)), variant: id, phase: "betting", banker: 1, preferred: 1, owners: [2]int{0, -1}}
	copy(g.names[:], names)
	copy(g.ids[:], ids)
	for i := range g.balances {
		g.balances[i] = Initial
	}
	g.decks = 8
	if id == "chemin" {
		g.decks = 6
	}
	if id == "banque" {
		g.decks = 3
	}
	g.shuffle()
	return g
}
func (g *Game) shuffle() {
	g.shoe = nil
	for d := 0; d < g.decks; d++ {
		for s, suit := range []string{"♠", "♥", "♣", "♦"} {
			for r, rank := range []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"} {
				v := r + 1
				if v >= 10 {
					v = 0
				}
				g.shoe = append(g.shoe, Card{Rank: rank, Suit: suit, Red: s == 1 || s == 3, Value: v})
			}
		}
	}
	g.rng.Shuffle(len(g.shoe), func(i, j int) { g.shoe[i], g.shoe[j] = g.shoe[j], g.shoe[i] })
}
func Points(cards []Card) int {
	sum := 0
	for _, c := range cards {
		sum += c.Value
	}
	return sum % 10
}
func (g *Game) draw() Card { c := g.shoe[0]; g.shoe = g.shoe[1:]; return c }
func (g *Game) Configure(host string, bank bool) error {
	if g.hand > 0 || g.phase != "betting" {
		return fmt.Errorf("参加する役は最初の配札前に選んでください")
	}
	copy(g.names[:], names)
	copy(g.ids[:], ids)
	if host == "shizue" || host == "kathy" {
		g.ids[1] = host
		if host == "shizue" {
			g.names[1] = "SHIZUE"
		} else {
			g.names[1] = "KATHY"
		}
	}
	index := -1
	for i, id := range g.ids {
		if host == id && i > 0 {
			index = i
		}
	}
	if index < 0 {
		return fmt.Errorf("進行役が見つかりません")
	}
	g.preferred = index
	g.humanBank = bank && g.variant == "banque"
	g.banker = index
	if g.humanBank {
		g.banker = 0
	}
	g.previewOwners()
	return nil
}
func (g *Game) previewOwners() {
	g.owners = [2]int{0, -1}
	if g.banker == 0 {
		g.owners[0] = g.preferred
	}
	if g.variant == "banque" {
		for i := 1; i < 4; i++ {
			if i != g.banker && i != g.owners[0] && g.balances[i] >= Minimum {
				g.owners[1] = i
				break
			}
		}
	}
}
func (g *Game) maxBet() int {
	if g.variant == "punto" {
		return g.balances[0]
	}
	if g.banker == 0 {
		return g.balances[0]
	}
	m := min(g.balances[0], g.balances[g.banker])
	if g.variant == "banque" {
		m = min(g.balances[0], g.balances[g.banker]-Minimum)
	}
	return max(0, m)
}
func (g *Game) ready(bank int) bool {
	if g.balances[0] < Minimum || g.cleared {
		return false
	}
	if g.variant == "punto" {
		return true
	}
	if bank < 0 || g.balances[bank] < Minimum {
		return false
	}
	if g.variant == "banque" && g.balances[bank] < Minimum*2 {
		return false
	}
	available := 0
	for i, b := range g.balances {
		if i != bank && b >= Minimum {
			available++
		}
	}
	if g.variant == "banque" {
		return available >= 2
	}
	return available >= 1
}
func (g *Game) nextBanker() int {
	for step := 1; step <= 4; step++ {
		i := (g.banker + step) % 4
		if g.balances[i] >= Minimum {
			return i
		}
	}
	return -1
}
func (g *Game) futureBanker() int {
	if g.variant == "chemin" && g.rotate {
		return g.nextBanker()
	}
	return g.banker
}
func (g *Game) Deal(amount int, target string) error {
	if g.phase != "betting" {
		return fmt.Errorf("次のハンドへ進んでから配札してください")
	}
	if !g.ready(g.banker) {
		return fmt.Errorf("参加者の持ち点が不足しています。新しいセッションで再開できます")
	}
	ok := false
	for _, v := range Bets {
		ok = ok || amount == v
	}
	bet := amount * Unit
	if !ok || bet > g.maxBet() {
		return fmt.Errorf("持ち点の範囲でベットを選んでください")
	}
	if g.variant == "punto" && target != "player" && target != "banker" && target != "tie" {
		return fmt.Errorf("賭ける場所を選んでください")
	}
	// Prepare all stakes before touching money, so invalid deals are atomic.
	stakes := [4]int{}
	owners := [2]int{0, -1}
	budget := g.balances[g.banker]
	if g.variant == "chemin" {
		order := []int{0, 1, 2, 3}
		if g.banker == 0 {
			order = []int{1, 2, 3}
		}
		for _, i := range order {
			if i == g.banker {
				continue
			}
			n := min(bet, g.balances[i], budget)
			n = n / Unit * Unit
			if i == 0 {
				n = bet
			}
			if n < Minimum {
				continue
			}
			stakes[i] = n
			budget -= n
		}
		if g.banker == 0 {
			owners[0] = -1
			for i := 1; i < 4; i++ {
				if stakes[i] > 0 {
					owners[0] = i
					break
				}
			}
		}
	} else if g.variant == "banque" {
		if g.banker == 0 {
			owners[0] = g.preferred
			for i := 1; i < 4; i++ {
				if i != owners[0] && g.balances[i] >= Minimum {
					owners[1] = i
					break
				}
			}
			if g.balances[owners[0]] < Minimum {
				for i := 1; i < 4; i++ {
					if g.balances[i] >= Minimum && i != owners[1] {
						owners[0] = i
						break
					}
				}
			}
			if owners[1] < 0 {
				return fmt.Errorf("左右の参加者の持ち点が不足しています")
			}
			n := min(bet, g.balances[owners[0]], budget-Minimum)
			n = n / Unit * Unit
			if n < Minimum {
				return fmt.Errorf("右側の持ち点が不足しています")
			}
			stakes[owners[0]] = n
			budget -= n
			n = min(bet, g.balances[owners[1]], budget)
			n = n / Unit * Unit
			if n < Minimum {
				return fmt.Errorf("左側の持ち点が不足しています")
			}
			stakes[owners[1]] = n
		} else {
			for i := 1; i < 4; i++ {
				if i != g.banker && g.balances[i] >= Minimum {
					owners[1] = i
					break
				}
			}
			if owners[1] < 0 {
				return fmt.Errorf("左側の参加者の持ち点が不足しています")
			}
			stakes[0] = bet
			budget -= bet
			n := min(bet, g.balances[owners[1]], budget)
			n = n / Unit * Unit
			if n < Minimum {
				return fmt.Errorf("左側をカバーする持ち点が不足しています")
			}
			stakes[owners[1]] = n
		}
	}
	if len(g.shoe) <= 16 {
		g.shuffle()
	}
	g.hand++
	g.wager = bet
	g.target = target
	g.startHuman = g.balances[0]
	g.profit = 0
	g.commission = 0
	g.rotate = false
	g.revealed = false
	g.outcome = ""
	g.result = ""
	g.bankCards = nil
	g.playerCards = [2][]Card{}
	g.decided = [2]bool{}
	g.pending = -1
	g.stakes = stakes
	g.owners = owners
	g.log = []string{fmt.Sprintf("ハンド %d · %s", g.hand, Name(g.variant))}
	if g.variant == "punto" {
		g.balances[0] -= bet
		g.stakes[0] = bet
	} else {
		sum := 0
		for i, n := range stakes {
			if i != g.banker {
				g.balances[i] -= n
				sum += n
			}
		}
		g.stakes[g.banker] = sum
		g.balances[g.banker] -= sum
		g.log = append(g.log, g.names[g.banker]+"がバンカーを担当")
	}
	for round := 0; round < 2; round++ {
		g.playerCards[0] = append(g.playerCards[0], g.draw())
		if g.variant == "banque" {
			g.playerCards[1] = append(g.playerCards[1], g.draw())
		}
		g.bankCards = append(g.bankCards, g.draw())
	}
	natural := Points(g.bankCards) >= 8 || Points(g.playerCards[0]) >= 8
	if g.variant == "banque" {
		natural = natural || Points(g.playerCards[1]) >= 8
	}
	if natural {
		g.log = append(g.log, "ナチュラル8/9 · 追加カードなし")
		g.finish()
		return nil
	}
	if g.variant == "punto" {
		g.phase = "dealt"
		return nil
	}
	g.progress()
	return nil
}

// third=-1 means the player stood. Chemin adds two optional banker cells.
func BankerDraw(b, third int) bool {
	if third < 0 {
		return b <= 5
	}
	switch {
	case b <= 2:
		return true
	case b == 3:
		return third != 8
	case b == 4:
		return third >= 2 && third <= 7
	case b == 5:
		return third >= 4 && third <= 7
	case b == 6:
		return third == 6 || third == 7
	}
	return false
}
func BankerOptional(b, third int) bool { return b == 3 && third == 9 || b == 5 && third == 4 }
func (g *Game) playerCount() int {
	if g.variant == "banque" {
		return 2
	}
	return 1
}
func (g *Game) applyPlayer(i int, hit bool) {
	if hit {
		g.playerCards[i] = append(g.playerCards[i], g.draw())
		g.log = append(g.log, []string{"右側／プレイヤー", "左側"}[i]+"が3枚目を引きました")
	} else {
		g.log = append(g.log, []string{"右側／プレイヤー", "左側"}[i]+"は2枚で止めました")
	}
	g.decided[i] = true
}
func (g *Game) progress() {
	for i := 0; i < g.playerCount(); i++ {
		if g.decided[i] {
			continue
		}
		p := Points(g.playerCards[i])
		if p == 5 && g.owners[i] == 0 {
			g.phase = "player"
			g.pending = i
			return
		}
		hit := p <= 4
		if p == 5 {
			hit = g.rng.Intn(2) == 0
		}
		g.applyPlayer(i, hit)
	}
	b := Points(g.bankCards)
	third := -1
	if len(g.playerCards[0]) == 3 {
		third = g.playerCards[0][2].Value
	}
	if g.banker == 0 && (g.variant == "banque" || BankerOptional(b, third)) {
		g.phase = "banker"
		return
	}
	hit := BankerDraw(b, third)
	if g.variant == "banque" {
		hit = b <= 5
	}
	if hit {
		g.bankCards = append(g.bankCards, g.draw())
		g.log = append(g.log, "バンカーが3枚目を引きました")
	}
	g.finish()
}
func (g *Game) Act(action string) error {
	if action == "retire" {
		if g.phase != "betting" || g.variant != "chemin" || g.banker != 0 {
			return fmt.Errorf("今はバンカーを辞退できません")
		}
		next := g.nextBanker()
		if next < 0 || next == 0 {
			return fmt.Errorf("交代できる参加者がいません")
		}
		g.banker = next
		g.rotate = false
		return nil
	}
	if g.phase == "dealt" && action == "continue" {
		p := Points(g.playerCards[0])
		if p <= 5 {
			g.playerCards[0] = append(g.playerCards[0], g.draw())
		}
		third := -1
		if len(g.playerCards[0]) == 3 {
			third = g.playerCards[0][2].Value
		}
		if BankerDraw(Points(g.bankCards), third) {
			g.bankCards = append(g.bankCards, g.draw())
		}
		g.finish()
		return nil
	}
	if action != "hit" && action != "stand" {
		return fmt.Errorf("この場面では実行できない操作です")
	}
	if g.phase == "player" {
		g.applyPlayer(g.pending, action == "hit")
		g.progress()
		return nil
	}
	if g.phase == "banker" {
		if action == "hit" {
			g.bankCards = append(g.bankCards, g.draw())
		}
		g.finish()
		return nil
	}
	return fmt.Errorf("カードを選ぶ場面ではありません")
}
func compare(p, b int) string {
	if p > b {
		return "player"
	}
	if p < b {
		return "banker"
	}
	return "tie"
}
func (g *Game) finish() {
	g.revealed = true
	g.phase = "done"
	b := Points(g.bankCards)
	winner := compare(Points(g.playerCards[0]), b)
	if g.variant == "punto" {
		returned := 0
		switch {
		case winner == "tie" && g.target != "tie":
			returned = g.wager
		case winner == g.target:
			if winner == "tie" {
				returned = g.wager * 9
			} else if winner == "banker" {
				g.commission = g.wager * 5 / 100
				returned = g.wager*2 - g.commission
			} else {
				returned = g.wager * 2
			}
		}
		g.balances[0] += returned
	} else {
		bankStart := g.balances[g.banker] + g.stakes[g.banker]
		for i, stake := range g.stakes {
			if i == g.banker || stake == 0 {
				continue
			}
			side := 0
			if g.variant == "banque" && i == g.owners[1] {
				side = 1
			}
			result := compare(Points(g.playerCards[side]), b)
			switch result {
			case "player":
				g.balances[i] += stake * 2
			case "banker":
				g.balances[g.banker] += stake * 2
			case "tie":
				g.balances[i] += stake
				g.balances[g.banker] += stake
			}
		}
		net := g.balances[g.banker] - bankStart
		if net > 0 {
			rate := 5
			if g.variant == "banque" {
				rate = 2
			}
			g.commission = net * rate / 100
			g.balances[g.banker] -= g.commission
		}
		g.rotate = g.variant == "chemin" && winner == "player"
	}
	g.profit = g.balances[0] - g.startHuman
	g.outcome = "tie"
	if g.profit > 0 {
		g.outcome = "win"
		g.wins++
	} else if g.profit < 0 {
		g.outcome = "loss"
		g.losses++
	} else {
		g.ties++
	}
	g.cleared = g.balances[0] >= Goal
	g.result = map[string]string{"win": "あなたの勝利", "loss": "今回は負け", "tie": "引き分け／収支ゼロ"}[g.outcome]
	g.log = append(g.log, fmt.Sprintf("%s · 収支 %+.2fチップ", g.result, money(g.profit)))
	if g.commission > 0 {
		g.log = append(g.log, fmt.Sprintf("バンカー勝ち分の手数料 %.2fチップ", money(g.commission)))
	}
	if g.rotate {
		g.log = append(g.log, "バンカーが敗れたため次のハンドで交代します")
	}
	g.history = append(g.history, Record{Hand: g.hand, Outcome: g.outcome, Profit: money(g.profit)})
	if len(g.history) > 60 {
		g.history = g.history[len(g.history)-60:]
	}
}
func (g *Game) Next() error {
	if g.phase != "done" {
		return fmt.Errorf("ハンドを終えてから進んでください")
	}
	next := g.futureBanker()
	if !g.ready(next) {
		return fmt.Errorf("持ち点不足または目標達成です。新しいセッションで再開できます")
	}
	g.banker = next
	g.rotate = false
	g.previewOwners()
	g.phase = "betting"
	g.bankCards = nil
	g.playerCards = [2][]Card{}
	g.stakes = [4]int{}
	g.wager = 0
	g.profit = 0
	g.commission = 0
	g.outcome = ""
	g.result = ""
	g.log = nil
	return nil
}
func money(v int) float64 { return float64(v) / Unit }
func (g *Game) handView(id, label string, cards []Card, owner int, stake int) Hand {
	shown := append([]Card{}, cards...)
	visible := g.revealed || g.variant == "punto" || owner == 0
	var points *int
	if len(cards) > 0 && visible {
		n := Points(cards)
		points = &n
	}
	if !visible {
		for i := 0; i < len(shown) && i < 2; i++ {
			shown[i] = Card{Hidden: true}
		}
	}
	name := ""
	if owner >= 0 {
		name = g.names[owner]
	}
	h := Hand{ID: id, Name: label, Owner: name, Cards: shown, Points: points, Stake: money(stake)}
	if g.revealed && id != "banker" {
		h.Result = compare(Points(cards), Points(g.bankCards))
	}
	return h
}
func (g *Game) View() View {
	role := "player"
	if g.variant != "punto" && g.banker == 0 {
		role = "banker"
	}
	seats := []Seat{}
	for i, name := range g.names {
		seats = append(seats, Seat{ID: g.ids[i], Name: name, Chips: money(g.balances[i]), Stake: money(g.stakes[i]), Banker: g.variant != "punto" && i == g.banker})
	}
	bankOwner := g.banker
	bankStake := g.stakes[g.banker]
	if g.variant == "punto" {
		bankOwner = -1
		bankStake = 0
		if g.target == "banker" {
			bankStake = g.wager
		}
	}
	hands := []Hand{g.handView("banker", "バンカー", g.bankCards, bankOwner, bankStake)}
	owner := g.owners[0]
	if g.variant == "punto" {
		owner = -1
	}
	stake := g.stakes[0]
	if g.variant == "punto" && g.target != "player" {
		stake = 0
	}
	if g.variant != "punto" {
		stake = 0
		for i, n := range g.stakes {
			if i != g.banker {
				stake += n
			}
		}
	}
	if g.variant == "banque" && g.owners[0] >= 0 {
		stake = g.stakes[g.owners[0]]
	}
	hands = append(hands, g.handView("right", map[bool]string{true: "右のプレイヤー", false: "プレイヤー"}[g.variant == "banque"], g.playerCards[0], owner, stake))
	if g.variant == "banque" {
		stake = 0
		if g.owners[1] >= 0 {
			stake = g.stakes[g.owners[1]]
		}
		hands = append(hands, g.handView("left", "左のプレイヤー", g.playerCards[1], g.owners[1], stake))
	}
	actions := []string{}
	hint := "ベット額を選んで配札します"
	pending := ""
	switch g.phase {
	case "dealt":
		actions = []string{"continue"}
		hint = "追加カードはルールに従って自動で配られます"
	case "player":
		actions = []string{"hit", "stand"}
		pending = hands[g.pending+1].Name
		hint = pending + "が5点です。3枚目を引くか止めるか選べます"
	case "banker":
		actions = []string{"hit", "stand"}
		pending = "バンカー"
		hint = "あなたがバンカーです。公開情報を見て3枚目を選びます"
	case "done":
		hint = g.result
		if !g.cleared && !g.ready(g.futureBanker()) {
			hint += " · 参加者の持ち点不足でセッション終了"
		}
	}
	if g.phase == "betting" && g.variant == "chemin" && g.banker == 0 {
		actions = []string{"retire"}
		hint = "あなたがバンカーです。基準額を選ぶとCPUが資金の範囲で賭けます"
	}
	var odds *Odds
	if g.variant == "punto" && (g.phase == "betting" || g.phase == "dealt") {
		v := g.odds()
		odds = &v
	}
	stage := 0
	if g.phase == "done" {
		stage = 1
	}
	return View{Variant: g.variant, Name: Name(g.variant), Phase: g.phase, Hand: g.hand, Role: role, Banker: g.banker, Seats: seats, Hands: hands, Chips: money(g.balances[0]), Goal: money(Goal), Bet: money(g.wager), MaxBet: money(g.maxBet()), Profit: money(g.profit), Commission: money(g.commission), Target: g.target, Done: g.phase == "done", Cleared: g.cleared, CanNext: g.phase == "done" && g.ready(g.futureBanker()), CanDeal: g.phase == "betting" && g.ready(g.banker), Actions: actions, Hint: hint, Outcome: g.outcome, Result: g.result, Pending: pending, Log: append([]string{}, g.log...), History: append([]Record{}, g.history...), Remaining: len(g.shoe), Decks: g.decks, Odds: odds, Wins: g.wins, Losses: g.losses, Ties: g.ties, Stages: []string{"プレイ", "結果"}, StageIndex: stage}
}
func drawValue(pool *[10]int, rng *rand.Rand) int {
	total := 0
	for _, n := range pool {
		total += n
	}
	r := rng.Intn(total)
	for v, n := range pool {
		r -= n
		if r < 0 {
			pool[v]--
			return v
		}
	}
	panic("unreachable")
}
func (g *Game) odds() Odds {
	counts := [10]int{}
	for _, c := range g.shoe {
		counts[c.Value]++
	}
	if g.phase == "betting" && len(g.shoe) <= 16 {
		counts = [10]int{}
		counts[0] = 16 * g.decks
		for i := 1; i < 10; i++ {
			counts[i] = 4 * g.decks
		}
	}
	var seed int64 = 73
	for _, n := range counts {
		seed = seed*31 + int64(n)
	}
	if g.phase == "dealt" {
		seed = seed*31 + int64(Points(g.playerCards[0])*10+Points(g.bankCards))
	}
	rng := rand.New(rand.NewSource(seed))
	pWin, bWin, tie := 0, 0, 0
	for n := 0; n < Samples; n++ {
		pool := counts
		p, b := 0, 0
		if g.phase == "dealt" {
			p = Points(g.playerCards[0])
			b = Points(g.bankCards)
		} else {
			p = drawValue(&pool, rng)
			b = drawValue(&pool, rng)
			p = (p + drawValue(&pool, rng)) % 10
			b = (b + drawValue(&pool, rng)) % 10
		}
		if p < 8 && b < 8 {
			third := -1
			if p <= 5 {
				third = drawValue(&pool, rng)
				p = (p + third) % 10
			}
			if BankerDraw(b, third) {
				b = (b + drawValue(&pool, rng)) % 10
			}
		}
		switch compare(p, b) {
		case "player":
			pWin++
		case "banker":
			bWin++
		default:
			tie++
		}
	}
	return Odds{Player: float64(pWin) * 100 / Samples, Banker: float64(bWin) * 100 / Samples, Tie: float64(tie) * 100 / Samples, Samples: Samples, Margin: 1.96 * math.Sqrt(.25/Samples) * 100}
}
