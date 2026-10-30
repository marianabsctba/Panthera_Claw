package pix

import (
	"testing"
	"time"
)

func TestLerE2E(t *testing.T) {
	e, err := LerE2E("E18236120202610291732abc12XYZ789")
	if err != nil {
		t.Fatal(err)
	}
	if e.ISPB != "18236120" || e.Tipo != "E" || e.Sequencial != "abc12XYZ789" {
		t.Fatalf("partes erradas: %+v", e)
	}
	if !e.Instante.Equal(time.Date(2026, 10, 29, 17, 32, 0, 0, time.UTC)) {
		t.Fatalf("instante %v", e.Instante)
	}
	if h := e.NoHorarioDeBrasilia().Format("15:04"); h != "14:32" {
		t.Fatalf("Brasília %s", h)
	}
}

func TestLerE2EQuebradoEComConfusao(t *testing.T) {
	// Quebrado em duas linhas e com O no lugar de zero na parte numérica.
	e, err := LerE2E("E18236120 2O261O291732\nabc12XYZ789")
	if err != nil {
		t.Fatal(err)
	}
	if e.Texto != "E18236120202610291732abc12XYZ789" || len(e.Correcoes) != 2 {
		t.Fatalf("%+v", e)
	}
}

func TestLerE2EInvalidos(t *testing.T) {
	for _, s := range []string{
		"E1823612020261029173",             // curto
		"X18236120202610291732abc12XYZ789", // começo errado
		"E18236120202613291732abc12XYZ789", // mês 13
		"E18236120202610291772abc12XYZ789", // minuto 72
		"E1823612020261029173#abc12XYZ789", // símbolo na parte numérica
		"E18236120202610291732abc12XYZ78#", // símbolo no sequencial
	} {
		if _, err := LerE2E(s); err == nil {
			t.Errorf("%s deveria ser inválido", s)
		}
	}
}

func TestChaves(t *testing.T) {
	casos := map[string]string{
		"529.982.247-25":                       ChaveCPF,
		"529.982.247-24":                       ChaveInvalida,
		"11.222.333/0001-81":                   ChaveCNPJ,
		"11.222.333/0001-82":                   ChaveInvalida,
		"12.ABC.345/01DE-35":                   ChaveCNPJ,
		"+55 41 99999-8888":                    ChaveTelefone,
		"+55 20 99999-8888":                    ChaveInvalida,
		"fulana@exemplo.com.br":                ChaveEmail,
		"123e4567-e89b-42d3-a456-556642440000": ChaveAleatoria,
		"***.982.247-**":                       ChaveMascarada,
		"qualquer coisa":                       ChaveInvalida,
	}
	for s, tipo := range casos {
		if c := LerChave(s); c.Tipo != tipo {
			t.Errorf("LerChave(%q) = %s (%s), quero %s", s, c.Tipo, c.Motivo, tipo)
		}
	}
}
