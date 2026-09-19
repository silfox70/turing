// Package tm implementa il motore di una macchina di Turing deterministica
// a nastro infinito in entrambe le direzioni, con istruzioni a quintuple.
// Il package non fa input/output: carica programmi, esegue passi e
// restituisce resoconti. La visualizzazione sta altrove.
package tm

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Blank è il simbolo che occupa ogni cella mai scritta.
const Blank = '_'

// Move è il movimento della testina: i tre movimenti di Turing.
type Move int8

const (
	Left  Move = -1
	Stay  Move = 0
	Right Move = 1
)

func (m Move) String() string {
	switch m {
	case Left:
		return "L"
	case Right:
		return "R"
	default:
		return "N"
	}
}

// Rule è una quintupla: (stato, letto) -> (scritto, movimento, nuovo stato).
type Rule struct {
	From  string
	Read  rune
	Write rune
	Move  Move
	To    string
	Line  int // riga del sorgente, utile per messaggi e visualizzazione
}

type key struct {
	state string
	sym   rune
}

// Program è un programma già validato: al massimo una regola per ogni
// coppia (stato, simbolo), quindi la macchina è deterministica.
type Program struct {
	Start  string
	Finals map[string]bool
	Rules  []Rule // nell'ordine del sorgente
	table  map[key]int
}

// Lookup restituisce la regola per la coppia (stato, simbolo), se esiste.
func (p *Program) Lookup(state string, sym rune) (Rule, bool) {
	i, ok := p.table[key{state, sym}]
	if !ok {
		return Rule{}, false
	}
	return p.Rules[i], true
}

var moves = map[string]Move{
	"L": Left, "R": Right, "N": Stay,
	"S": Left, "D": Right, "F": Stay, // sinistra, destra, fermo
}

// Parse legge il formato testuale del programma:
//
//	# commento
//	inizio: q0
//	finali: qf
//	q0 1 1 R q0      (stato letto scritto mossa nuovo)
//
// Raccoglie tutti gli errori invece di fermarsi al primo.
func Parse(src string) (*Program, error) {
	p := &Program{Finals: map[string]bool{}, table: map[key]int{}}
	var errs []error

	for i, raw := range strings.Split(src, "\n") {
		n := i + 1
		line, _, _ := strings.Cut(raw, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		low := strings.ToLower(line)
		if strings.HasPrefix(low, "inizio:") {
			p.Start = strings.TrimSpace(line[len("inizio:"):])
			continue
		}
		if strings.HasPrefix(low, "finali:") {
			for _, s := range strings.FieldsFunc(line[len("finali:"):], func(r rune) bool {
				return r == ' ' || r == '\t' || r == ','
			}) {
				p.Finals[s] = true
			}
			continue
		}

		f := strings.Fields(line)
		if len(f) != 5 {
			errs = append(errs, fmt.Errorf("riga %d: servono 5 campi, trovati %d", n, len(f)))
			continue
		}
		mv, ok := moves[strings.ToUpper(f[3])]
		if !ok {
			errs = append(errs, fmt.Errorf("riga %d: mossa %q non valida, usa L, R o N", n, f[3]))
			continue
		}
		if utf8.RuneCountInString(f[1]) != 1 || utf8.RuneCountInString(f[2]) != 1 {
			errs = append(errs, fmt.Errorf("riga %d: i simboli devono essere di un carattere", n))
			continue
		}
		read, _ := utf8.DecodeRuneInString(f[1])
		write, _ := utf8.DecodeRuneInString(f[2])

		k := key{f[0], read}
		if j, dup := p.table[k]; dup {
			errs = append(errs, fmt.Errorf("riga %d: regola duplicata per (%s, %c), già definita alla riga %d",
				n, f[0], read, p.Rules[j].Line))
			continue
		}
		p.table[k] = len(p.Rules)
		p.Rules = append(p.Rules, Rule{From: f[0], Read: read, Write: write, Move: mv, To: f[4], Line: n})
	}

	if len(p.Rules) == 0 && len(errs) == 0 {
		errs = append(errs, errors.New("il programma non contiene regole"))
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	if p.Start == "" {
		p.Start = p.Rules[0].From
	}
	return p, nil
}

// Tape è il nastro infinito in entrambe le direzioni, fatto di due slice
// che crescono solo quando la testina scrive oltre la parte già allocata:
// right[i] è la cella i, left[i] è la cella -(i+1).
type Tape struct {
	right []rune
	left  []rune
}

// Get legge la cella p; le celle mai allocate valgono Blank.
func (t *Tape) Get(p int) rune {
	s, i := t.side(p)
	if i < len(*s) {
		return (*s)[i]
	}
	return Blank
}

// Set scrive nella cella p, allungando il nastro se necessario.
func (t *Tape) Set(p int, sym rune) {
	s, i := t.side(p)
	if i >= len(*s) {
		if sym == Blank {
			return // scrivere un bianco oltre il bordo non cambia nulla
		}
		for len(*s) <= i {
			*s = append(*s, Blank)
		}
	}
	(*s)[i] = sym
}

func (t *Tape) side(p int) (*[]rune, int) {
	if p >= 0 {
		return &t.right, p
	}
	return &t.left, -p - 1
}

// Bounds restituisce la prima e l'ultima cella non bianca;
// ok è false se il nastro è tutto bianco.
func (t *Tape) Bounds() (lo, hi int, ok bool) {
	lo, hi = 1, 0
	for p := -len(t.left); p < len(t.right); p++ {
		if t.Get(p) != Blank {
			if !ok {
				lo, ok = p, true
			}
			hi = p
		}
	}
	return lo, hi, ok
}

// Content restituisce il contenuto non bianco del nastro, dalla prima
// all'ultima cella scritta (i bianchi interni restano come '_').
func (t *Tape) Content() string {
	lo, hi, ok := t.Bounds()
	if !ok {
		return ""
	}
	var b strings.Builder
	for p := lo; p <= hi; p++ {
		b.WriteRune(t.Get(p))
	}
	return b.String()
}

// Step è il resoconto di un passo: tutto ciò che serve per mostrarlo
// e per disfarlo.
type Step struct {
	N        int    // numero del passo, da 1
	Pos      int    // posizione della testina prima del passo
	OldSym   rune   // simbolo che c'era nella cella
	OldState string // stato prima del passo
	Rule     Rule   // regola applicata
}

// Halt descrive perché la macchina è ferma.
type Halt int

const (
	NotHalted  Halt = iota // c'è una regola applicabile
	FinalState             // la macchina è in uno stato finale
	NoRule                 // nessuna regola per (stato, simbolo)
)

func (h Halt) String() string {
	switch h {
	case FinalState:
		return "stato finale"
	case NoRule:
		return "nessuna regola applicabile"
	default:
		return "in esecuzione"
	}
}

// Machine è una configurazione in esecuzione di un programma.
type Machine struct {
	prog    *Program
	tape    Tape
	pos     int
	state   string
	history []Step
}

// New crea una macchina con l'input scritto dalla cella 0 in poi e la
// testina sulla cella 0. Spazio e '_' nell'input sono bianchi.
func New(p *Program, input string) *Machine {
	m := &Machine{prog: p, state: p.Start}
	i := 0
	for _, r := range input {
		if r == ' ' {
			r = Blank
		}
		m.tape.Set(i, r)
		i++
	}
	return m
}

func (m *Machine) Program() *Program { return m.prog }
func (m *Machine) State() string     { return m.state }
func (m *Machine) Pos() int          { return m.pos }
func (m *Machine) Steps() int        { return len(m.history) }
func (m *Machine) Tape() *Tape       { return &m.tape }
func (m *Machine) History() []Step   { return m.history }

// Next restituisce la regola che verrebbe applicata al prossimo passo.
func (m *Machine) Next() (Rule, bool) {
	if m.prog.Finals[m.state] {
		return Rule{}, false
	}
	return m.prog.Lookup(m.state, m.tape.Get(m.pos))
}

// Halted dice se la macchina è ferma e perché.
func (m *Machine) Halted() Halt {
	if m.prog.Finals[m.state] {
		return FinalState
	}
	if _, ok := m.Next(); !ok {
		return NoRule
	}
	return NotHalted
}

// Step esegue un passo e ne restituisce il resoconto; ok è false se la
// macchina era già ferma.
func (m *Machine) Step() (Step, bool) {
	r, ok := m.Next()
	if !ok {
		return Step{}, false
	}
	s := Step{
		N:        len(m.history) + 1,
		Pos:      m.pos,
		OldSym:   m.tape.Get(m.pos),
		OldState: m.state,
		Rule:     r,
	}
	m.tape.Set(m.pos, r.Write)
	m.pos += int(r.Move)
	m.state = r.To
	m.history = append(m.history, s)
	return s, true
}

// Back disfa l'ultimo passo usando il suo resoconto.
func (m *Machine) Back() bool {
	if len(m.history) == 0 {
		return false
	}
	s := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	m.tape.Set(s.Pos, s.OldSym)
	m.pos = s.Pos
	m.state = s.OldState
	return true
}

// Run esegue fino all'arresto o fino a limit passi (0 = nessun limite).
// Restituisce il motivo dell'arresto: NotHalted significa limite raggiunto.
func (m *Machine) Run(limit int) Halt {
	for n := 0; limit == 0 || n < limit; n++ {
		if _, ok := m.Step(); !ok {
			return m.Halted()
		}
	}
	return m.Halted()
}
