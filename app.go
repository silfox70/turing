package main

import (
	"context"
	"fmt"

	"github.com/silfox70/turing/tm"
)

// maxRadius limita quante celle il frontend può chiedere attorno alla
// testina: la finestra la decide la pagina, che sa quanto è larga, ma
// non deve poter chiedere una fotografia sterminata a ogni passo.
const maxRadius = 200

// runLimit è la rete di sicurezza per le macchine che non si fermano.
const runLimit = 1_000_000

// App è il collegamento fra il motore e la finestra. Tutti i metodi
// pubblici di questa struct diventano funzioni JavaScript.
type App struct {
	ctx     context.Context
	prog    *tm.Program
	machine *tm.Machine
	radius  int
}

func NewApp() *App {
	return &App{radius: 12}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Cell è una cella della finestra di nastro mostrata attorno alla testina.
type Cell struct {
	Pos  int    `json:"pos"`
	Sym  string `json:"sym"`
	Head bool   `json:"head"`
}

// Snapshot è la fotografia di ciò che va mostrato dopo ogni passo.
// Non contiene il grafo né la tavola delle regole, che non cambiano
// durante l'esecuzione e viaggiano una sola volta con Load.
type Snapshot struct {
	State   string `json:"state"`
	Pos     int    `json:"pos"`
	Steps   int    `json:"steps"`
	Cells   []Cell `json:"cells"`
	Halted  bool   `json:"halted"`
	Reason  string `json:"reason"`
	Final   bool   `json:"final"`   // fermo in uno stato finale
	Next    int    `json:"next"`    // indice della regola che scatterà, -1 se nessuna
	Last    string `json:"last"`    // resoconto dell'ultimo passo eseguito
	Content string `json:"content"` // contenuto non bianco del nastro
	Error   string `json:"error"`   // vuoto se tutto bene
}

// Program è ciò che viaggia una volta sola, al caricamento.
type Program struct {
	Rules    []RuleView `json:"rules"`
	Graph    *tm.Graph  `json:"graph"`
	Legend   string     `json:"legend"`
	Snapshot Snapshot   `json:"snapshot"`
	Error    string     `json:"error"`
}

// RuleView è una riga della tavola delle regole, già in forma testuale.
type RuleView struct {
	From  string `json:"from"`
	Read  string `json:"read"`
	Write string `json:"write"`
	Move  string `json:"move"`
	To    string `json:"to"`
	Label string `json:"label"`
	Line  int    `json:"line"`
}

// Load compila il programma e prepara la macchina sull'input dato.
// In caso di errore restituisce i messaggi del caricatore, tutti insieme.
func (a *App) Load(src, input string) Program {
	p, err := tm.Parse(src)
	if err != nil {
		return Program{Error: err.Error()}
	}
	a.prog = p
	a.machine = tm.New(p, input)

	rules := make([]RuleView, len(p.Rules))
	for i, r := range p.Rules {
		rules[i] = RuleView{
			From:  r.From,
			Read:  string(r.Read),
			Write: string(r.Write),
			Move:  r.Move.String(),
			To:    r.To,
			Label: r.Label(),
			Line:  r.Line,
		}
	}
	return Program{
		Rules:    rules,
		Graph:    p.Graph(),
		Legend:   tm.Legend,
		Snapshot: a.snapshot(""),
	}
}

// SetRadius fissa quante celle mostrare per lato attorno alla testina.
// La pagina lo chiama quando la finestra cambia dimensione.
func (a *App) SetRadius(r int) Snapshot {
	switch {
	case r < 1:
		a.radius = 1
	case r > maxRadius:
		a.radius = maxRadius
	default:
		a.radius = r
	}
	return a.snapshot("")
}

// Step esegue un passo e restituisce la nuova fotografia.
func (a *App) Step() Snapshot {
	if a.machine == nil {
		return Snapshot{Error: "nessun programma caricato"}
	}
	s, ok := a.machine.Step()
	if !ok {
		return a.snapshot("")
	}
	return a.snapshot(describe(s))
}

// Back disfa l'ultimo passo.
func (a *App) Back() Snapshot {
	if a.machine == nil {
		return Snapshot{Error: "nessun programma caricato"}
	}
	a.machine.Back()
	return a.snapshot("")
}

// Run esegue fino all'arresto o fino a n passi; con n a zero usa il
// limite di sicurezza. Serve per saltare avanti senza una fotografia
// per ogni passo.
func (a *App) Run(n int) Snapshot {
	if a.machine == nil {
		return Snapshot{Error: "nessun programma caricato"}
	}
	if n <= 0 || n > runLimit {
		n = runLimit
	}
	before := a.machine.Steps()
	a.machine.Run(n)
	snap := a.snapshot("")
	if !snap.Halted && a.machine.Steps()-before >= n {
		snap.Error = fmt.Sprintf("limite di %d passi raggiunto: la macchina potrebbe non fermarsi mai", n)
	}
	return snap
}

// Reset riporta la macchina alla configurazione iniziale sull'input dato,
// senza ricompilare il programma.
func (a *App) Reset(input string) Snapshot {
	if a.prog == nil {
		return Snapshot{Error: "nessun programma caricato"}
	}
	a.machine = tm.New(a.prog, input)
	return a.snapshot("")
}

func describe(s tm.Step) string {
	return fmt.Sprintf("passo %d · (%s, %c) → scrive %c, %s, %s",
		s.N, s.OldState, s.OldSym, s.Rule.Write, moveWord(s.Rule.Move), s.Rule.To)
}

func moveWord(m tm.Move) string {
	switch m {
	case tm.Left:
		return "sinistra"
	case tm.Right:
		return "destra"
	default:
		return "fermo"
	}
}

// snapshot costruisce la fotografia dallo stato corrente della macchina.
func (a *App) snapshot(last string) Snapshot {
	if a.machine == nil {
		return Snapshot{Error: "nessun programma caricato", Next: -1}
	}
	m := a.machine
	halt := m.Halted()

	cells := make([]Cell, 0, 2*a.radius+1)
	for p := m.Pos() - a.radius; p <= m.Pos()+a.radius; p++ {
		cells = append(cells, Cell{Pos: p, Sym: string(m.Tape().Get(p)), Head: p == m.Pos()})
	}

	next := -1
	if r, ok := m.Next(); ok {
		for i, c := range a.prog.Rules {
			if c.Line == r.Line {
				next = i
				break
			}
		}
	}

	return Snapshot{
		State:   m.State(),
		Pos:     m.Pos(),
		Steps:   m.Steps(),
		Cells:   cells,
		Halted:  halt != tm.NotHalted,
		Reason:  halt.String(),
		Final:   halt == tm.FinalState,
		Next:    next,
		Last:    last,
		Content: m.Tape().Content(),
	}
}
