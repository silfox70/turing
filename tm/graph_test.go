package tm

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestGraphIncremento(t *testing.T) {
	g := mustParse(t, incremento).Graph()

	if g.Synthetic {
		t.Error("tre stati: atteso grafo completo, non sintetico")
	}
	want := []string{"q0", "q1", "qf"}
	if len(g.Nodes) != len(want) {
		t.Fatalf("%d nodi, attesi %d", len(g.Nodes), len(want))
	}
	for i, w := range want {
		if g.Nodes[i].Name != w {
			t.Errorf("nodo %d: %q, atteso %q", i, g.Nodes[i].Name, w)
		}
	}
	if !g.Nodes[0].Start || !g.Nodes[0].SelfLop {
		t.Error("q0: atteso iniziale e con cappio")
	}
	if !g.Nodes[2].Final || g.Nodes[2].SelfLop {
		t.Error("qf: atteso finale e senza cappio")
	}

	// Le sei regole si raggruppano in quattro archi.
	byPair := map[string][]string{}
	for _, e := range g.Edges {
		byPair[e.From+"->"+e.To] = e.Labels
	}
	cases := map[string][]string{
		"q0->q0": {"(0;0;R)", "(1;1;R)"},
		"q0->q1": {"(_;_;L)"},
		"q1->q1": {"(1;0;L)"},
		"q1->qf": {"(0;1;N)", "(_;1;N)"},
	}
	if len(g.Edges) != len(cases) {
		t.Errorf("%d archi, attesi %d", len(g.Edges), len(cases))
	}
	for pair, want := range cases {
		got := strings.Join(byPair[pair], " ")
		if got != strings.Join(want, " ") {
			t.Errorf("%s: etichette %q, attese %q", pair, got, strings.Join(want, " "))
		}
	}
}

func TestDisposizioneInCerchio(t *testing.T) {
	g := mustParse(t, incremento).Graph()
	for _, n := range g.Nodes {
		d := math.Hypot(n.X-0.5, n.Y-0.5)
		if math.Abs(d-0.42) > 1e-9 {
			t.Errorf("%s: distanza dal centro %.4f, attesa 0.42", n.Name, d)
		}
		if n.X < 0 || n.X > 1 || n.Y < 0 || n.Y > 1 {
			t.Errorf("%s: posizione (%.3f, %.3f) fuori da [0,1]", n.Name, n.X, n.Y)
		}
	}
	if g.Nodes[0].Y >= 0.5 {
		t.Error("il primo nodo dovrebbe stare in alto")
	}
}

func TestSogliaSintetica(t *testing.T) {
	// Catena di n stati: q0 -> q1 -> ... -> qn.
	catena := func(n int) string {
		var b strings.Builder
		b.WriteString("inizio: q0\n")
		for i := 0; i < n; i++ {
			fmt.Fprintf(&b, "q%d 1 1 R q%d\n", i, i+1)
		}
		return b.String()
	}
	for _, c := range []struct {
		stati     int
		sintetico bool
	}{{MaxDrawableStates, false}, {MaxDrawableStates + 1, true}} {
		g := mustParse(t, catena(c.stati-1)).Graph()
		if len(g.Nodes) != c.stati {
			t.Fatalf("%d nodi, attesi %d", len(g.Nodes), c.stati)
		}
		if g.Synthetic != c.sintetico {
			t.Errorf("%d stati: sintetico %v, atteso %v", c.stati, g.Synthetic, c.sintetico)
		}
	}
}

func TestStatiDeterministici(t *testing.T) {
	// Più stati finali: l'ordine non deve dipendere dall'iterazione di mappa.
	src := "inizio: q0\nfinali: qz qy qx\nq0 1 1 R q0\n"
	first := mustParse(t, src).Graph().String()
	for i := 0; i < 20; i++ {
		if got := mustParse(t, src).Graph().String(); got != first {
			t.Fatalf("grafo non deterministico:\n%s\n---\n%s", first, got)
		}
	}
}
