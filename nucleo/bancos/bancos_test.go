package bancos

import "testing"

func TestMarcasTemInstituicoes(t *testing.T) {
	for i, g := range grupos {
		if len(g.ISPBs) == 0 {
			t.Errorf("marca %q sem nenhuma instituição na lista", marcas[i].marca)
		}
	}
}

func TestReconhecer(t *testing.T) {
	casos := map[string]string{
		"NU PAGAMENTOS - IP":         "Nubank",
		"Instituição: Itaú Unibanco": "Itaú",
		"PAGSEGURO INTERNET IP S.A.": "PagBank",
		"BANCO INTER":                "Inter",
		"Caixa Econômica Federal":    "Caixa",
		"Sicoob Credicoop":           "Sicoob",
		"Banco do Brasil S.A.":       "Banco do Brasil",
	}
	for texto, marca := range casos {
		g := Reconhecer(texto)
		if g == nil || g.Marca != marca {
			t.Errorf("Reconhecer(%q) = %v, quero %s", texto, g, marca)
		}
	}
	if g := Reconhecer("internet banking"); g != nil {
		t.Errorf("internet não pode virar Inter: %v", g.Marca)
	}
}

func TestPorISPB(t *testing.T) {
	b, ok := PorISPB("18236120")
	if !ok || b.Compe != "260" {
		t.Fatalf("Nu Pagamentos: %+v %v", b, ok)
	}
	if g := Marca("18236120"); g == nil || g.Marca != "Nubank" {
		t.Fatalf("marca do 18236120: %v", g)
	}
}
