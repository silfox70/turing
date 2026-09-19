package tm

import (
	"strings"
	"testing"
)

const incremento = `# Incremento binario
inizio: q0
finali: qf
# stato letto scrive mossa nuovo
q0 0 0 R q0
q0 1 1 R q0
q0 _ _ L q1
q1 1 0 L q1
q1 0 1 N qf
q1 _ 1 N qf
`

func mustParse(t *testing.T, src string) *Program {
	t.Helper()
	p, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return p
}

func TestIncremento(t *testing.T) {
	p := mustParse(t, incremento)
	cases := []struct {
		in, want string
		steps    int
	}{
		{"1011", "1100", 8},
		{"111", "1000", 8}, // il riporto esce a sinistra, nella cella -1
		{"0", "1", 3},
		{"", "1", 2}, // nastro vuoto: vale zero
	}
	for _, c := range cases {
		m := New(p, c.in)
		if h := m.Run(1000); h != FinalState {
			t.Errorf("%q: arresto %v, atteso stato finale", c.in, h)
		}
		if got := m.Tape().Content(); got != c.want {
			t.Errorf("%q: nastro %q, atteso %q", c.in, got, c.want)
		}
		if m.Steps() != c.steps {
			t.Errorf("%q: %d passi, attesi %d", c.in, m.Steps(), c.steps)
		}
	}
}

func TestIndietro(t *testing.T) {
	p := mustParse(t, incremento)
	m := New(p, "111")
	m.Run(1000)
	for m.Back() {
	}
	if m.Tape().Content() != "111" || m.Pos() != 0 || m.State() != "q0" || m.Steps() != 0 {
		t.Errorf("dopo Back: nastro %q pos %d stato %s passi %d",
			m.Tape().Content(), m.Pos(), m.State(), m.Steps())
	}
}

func TestNessunaRegola(t *testing.T) {
	src := strings.Replace(incremento, "q1 _ 1 N qf\n", "", 1)
	m := New(mustParse(t, src), "111")
	if h := m.Run(1000); h != NoRule {
		t.Errorf("arresto %v, atteso nessuna regola", h)
	}
}

func TestDeterminismo(t *testing.T) {
	_, err := Parse(incremento + "q0 0 1 R q0\n")
	if err == nil || !strings.Contains(err.Error(), "duplicata") {
		t.Errorf("atteso errore di regola duplicata, ottenuto %v", err)
	}
}

func TestErroriMultipli(t *testing.T) {
	_, err := Parse("q0 1 1 X q0\nq0 1\n")
	if err == nil || strings.Count(err.Error(), "riga") != 2 {
		t.Errorf("attesi due errori, ottenuto %v", err)
	}
}
