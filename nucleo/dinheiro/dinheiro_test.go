package dinheiro

import "testing"

func TestCentavos(t *testing.T) {
	casos := map[string]int64{
		"R$ 1,50":         150,
		"1,5":             150,
		"R$ 15,00":        1500,
		"15,0":            1500,
		"R$ 150,00":       15000,
		"R$ 150":          15000,
		"R$ 1.250,00":     125000,
		"1250,5":          125050,
		"1 250,00":        125000,
		"R$ 1.500":        150000,
		"R$ 1.500.000":    150000000,
		"15.50":           1550,
		"R$ 0,01":         1,
		"R$ 1.234.567,89": 123456789,
	}
	for s, quer := range casos {
		c, ok := Centavos(s)
		if !ok || c != quer {
			t.Errorf("Centavos(%q) = %d %v, quero %d", s, c, ok, quer)
		}
	}
	for _, s := range []string{"", "abc", "1,505", "1.50,00", "R$"} {
		if c, ok := Centavos(s); ok {
			t.Errorf("Centavos(%q) = %d, deveria recusar", s, c)
		}
	}
}

func TestOsTresDoBug(t *testing.T) {
	// Antes, os três viravam "150" e eram tratados como o mesmo valor.
	pares := [][2]string{{"R$ 1,50", "R$ 150,00"}, {"R$ 1,5", "R$ 15,00"}, {"R$ 15,00", "R$ 150,00"}, {"R$ 1,50", "R$ 150"}}
	for _, p := range pares {
		if iguais, ok := Iguais(p[0], p[1]); !ok || iguais {
			t.Errorf("%s e %s não são iguais", p[0], p[1])
		}
	}
	if iguais, ok := Iguais("R$ 150", "150,00"); !ok || !iguais {
		t.Error("R$ 150 e 150,00 são iguais")
	}
}

func TestCentavosPonto(t *testing.T) {
	casos := map[string]int64{"10.50": 1050, "1.5": 150, "1000": 100000, "0.01": 1}
	for s, quer := range casos {
		if c, ok := CentavosPonto(s); !ok || c != quer {
			t.Errorf("%q = %d %v", s, c, ok)
		}
	}
	if _, ok := CentavosPonto("1.505"); ok {
		t.Error("três casas não vale")
	}
}

func TestFormatar(t *testing.T) {
	casos := map[int64]string{150: "R$ 1,50", 1500: "R$ 15,00", 125000: "R$ 1.250,00", 5: "R$ 0,05", 123456789: "R$ 1.234.567,89"}
	for c, quer := range casos {
		if s := Formatar(c); s != quer {
			t.Errorf("Formatar(%d) = %s, quero %s", c, s, quer)
		}
	}
}
