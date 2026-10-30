package boleto

import (
	"strings"
	"testing"
	"time"
)

var hoje = time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)

// montarBancario gera uma linha válida a partir das partes, para testar
// sem depender de boleto real.
func montarBancario(banco, fator, valor, livre string) string {
	semDV := banco + "9" + fator + valor + livre
	dv := dvGeralBancario(semDV)
	barras := semDV[:4] + string(dv) + semDV[4:]
	return lerBancarioBarras(barras, hoje).Linha
}

func TestExemploPublicoDoBancoDoBrasil(t *testing.T) {
	// Exemplo de linha digitável publicado em manuais de cobrança.
	b, err := Ler("00190.50095 40144.816069 06809.350314 3 37370000000100", hoje)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Erros) != 0 || b.Banco != "001" || b.Valor != "R$ 1,00" {
		t.Fatalf("%+v", b)
	}
}

func TestBancarioValido(t *testing.T) {
	// Fator 3235 no ciclo novo: 22/02/2025 + 2235 dias.
	linha := montarBancario("237", "3235", "0000015000", "1234567890123456789012345")
	b, err := Ler(linha, hoje)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Erros) != 0 || b.Banco != "237" || b.Valor != "R$ 150,00" {
		t.Fatalf("%+v", b)
	}
	quer := time.Date(2025, 2, 22, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 2235).Format("02/01/2006")
	if b.Vencimento != quer {
		t.Fatalf("vencimento %s, quero %s", b.Vencimento, quer)
	}
	// Ida e volta pelo código de barras.
	if c, _ := Ler(b.Barras, hoje); c.Linha != b.Linha {
		t.Fatalf("barras -> linha: %s != %s", c.Linha, b.Linha)
	}
}

func TestBancarioAdulterado(t *testing.T) {
	linha := montarBancario("237", "3235", "0000015000", "1234567890123456789012345")
	// Troca o valor (R$ 150,00 -> R$ 950,00) sem recalcular o dígito geral.
	adulterada := linha[:37] + "9" + linha[38:]
	b, _ := Ler(adulterada, hoje)
	if len(b.Erros) == 0 {
		t.Fatalf("adulteração não detectada: %+v", b)
	}
	// Troca no meio do bloco 2: o dígito do bloco acusa.
	trocada := linha[:12] + string('0'+(linha[12]-'0'+1)%10) + linha[13:]
	if b, _ := Ler(trocada, hoje); !strings.Contains(strings.Join(b.Erros, " "), "bloco 2") {
		t.Fatalf("%+v", b.Erros)
	}
}

func TestArrecadacao(t *testing.T) {
	// Monta um código de barras de conta de energia (segmento 3, valor em
	// reais com módulo 10) e confere a linha gerada.
	semDV := "836" + "00000012345" + "0001" + "1234567890123456789012345"[:25]
	semDV = semDV[:43]
	barras := semDV[:3] + string(modulo10(semDV)) + semDV[3:]
	b, err := Ler(barras, hoje)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Erros) != 0 || b.Segmento != "energia elétrica e gás" || b.Valor != "R$ 123,45" {
		t.Fatalf("%+v", b)
	}
	c, _ := Ler(b.Linha, hoje)
	if len(c.Erros) != 0 || c.Barras != barras {
		t.Fatalf("linha -> barras: %+v", c)
	}
	errada := c.Linha[:5] + string('0'+(c.Linha[5]-'0'+1)%10) + c.Linha[6:]
	if d, _ := Ler(errada, hoje); len(d.Erros) == 0 {
		t.Fatal("erro na linha de arrecadação não detectado")
	}
}

func TestTamanhoErrado(t *testing.T) {
	if _, err := Ler("1234", hoje); err == nil {
		t.Fatal("deveria recusar")
	}
}
