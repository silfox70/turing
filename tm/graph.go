package tm

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// MaxDrawableStates è la soglia oltre la quale il grafo completo smette di
// essere leggibile: sopra questo numero di stati la vista passa al solo
// registro di stato (un nodo grande che cambia etichetta).
const MaxDrawableStates = 12

// Node è uno stato della macchina, con la posizione calcolata per il disegno.
type Node struct {
	Name    string  `json:"name"`
	X       float64 `json:"x"` // coordinate normalizzate in [0,1]
	Y       float64 `json:"y"`
	Start   bool    `json:"start"`
	Final   bool    `json:"final"`
	SelfLop bool    `json:"selfLoop"` // ha almeno una transizione su se stesso
}

// Edge è l'insieme delle transizioni che collegano la stessa coppia di stati.
// Labels contiene tutte le etichette (R;W;M), nell'ordine del sorgente.
type Edge struct {
	From   string   `json:"from"`
	To     string   `json:"to"`
	Labels []string `json:"labels"`
	Loop   bool     `json:"loop"` // From == To
}

// Graph è la struttura del programma pronta per essere disegnata.
// Se Synthetic è true gli stati sono troppi per il grafo completo e le
// posizioni dei nodi non sono significative: si disegna il solo registro.
type Graph struct {
	Nodes     []Node `json:"nodes"`
	Edges     []Edge `json:"edges"`
	Synthetic bool   `json:"synthetic"`
}

// Label restituisce l'etichetta (R;W;M) di una regola, nella stessa
// notazione usata nel sorgente del programma.
func (r Rule) Label() string {
	return fmt.Sprintf("(%c;%c;%s)", r.Read, r.Write, r.Move)
}

// States restituisce tutti gli stati del programma in ordine di prima
// apparizione nel sorgente: quelli di partenza delle regole, poi quelli
// raggiunti soltanto come destinazione, infine gli stati finali dichiarati
// che non compaiono in nessuna regola.
func (p *Program) States() []string {
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	add(p.Start)
	for _, r := range p.Rules {
		add(r.From)
	}
	for _, r := range p.Rules {
		add(r.To)
	}
	finals := make([]string, 0, len(p.Finals))
	for s := range p.Finals {
		finals = append(finals, s)
	}
	sort.Strings(finals) // l'ordine di una mappa non è deterministico
	for _, s := range finals {
		add(s)
	}
	return out
}

// Graph estrae la struttura del programma e ne calcola la disposizione.
// I nodi sono disposti in cerchio, in coordinate normalizzate: è una
// disposizione sempre corretta e mai ottima, sostituibile da sola.
func (p *Program) Graph() *Graph {
	states := p.States()
	g := &Graph{Synthetic: len(states) > MaxDrawableStates}

	loops := map[string]bool{}
	idx := map[string]int{}
	for _, r := range p.Rules {
		if r.From == r.To {
			loops[r.From] = true
		}
		k := r.From + "\x00" + r.To
		if i, ok := idx[k]; ok {
			g.Edges[i].Labels = append(g.Edges[i].Labels, r.Label())
			continue
		}
		idx[k] = len(g.Edges)
		g.Edges = append(g.Edges, Edge{
			From: r.From, To: r.To, Labels: []string{r.Label()}, Loop: r.From == r.To,
		})
	}

	g.Nodes = make([]Node, len(states))
	for i, s := range states {
		g.Nodes[i] = Node{
			Name:    s,
			Start:   s == p.Start,
			Final:   p.Finals[s],
			SelfLop: loops[s],
		}
		if g.Synthetic {
			continue // nella vista sintetica le posizioni non servono
		}
		// Primo nodo in alto, poi in senso orario; raggio 0.42 per
		// lasciare margine alle etichette degli archi.
		a := 2*math.Pi*float64(i)/float64(len(states)) - math.Pi/2
		g.Nodes[i].X = 0.5 + 0.42*math.Cos(a)
		g.Nodes[i].Y = 0.5 + 0.42*math.Sin(a)
	}
	return g
}

// Legend è la riga che scioglie la notazione delle etichette.
const Legend = "(R;W;M) = legge; scrive; muove — L sinistra, R destra, N fermo"

// String rende il grafo in forma testuale, utile per i test e per una
// vista a riga di comando.
func (g *Graph) String() string {
	var b strings.Builder
	for _, n := range g.Nodes {
		b.WriteString(n.Name)
		if n.Start {
			b.WriteString(" [iniziale]")
		}
		if n.Final {
			b.WriteString(" [finale]")
		}
		b.WriteString("\n")
	}
	for _, e := range g.Edges {
		fmt.Fprintf(&b, "%s -> %s %s\n", e.From, e.To, strings.Join(e.Labels, " "))
	}
	return b.String()
}
