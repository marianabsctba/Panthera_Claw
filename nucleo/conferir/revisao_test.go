package conferir

import (
	"fmt"
	"testing"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/arquivo"
	"github.com/marianabsctba/Panthera_Claw/nucleo/brcode"
	"github.com/marianabsctba/Panthera_Claw/nucleo/leitura"
	"github.com/marianabsctba/Panthera_Claw/nucleo/pix"
)

// Casos achados na revisão de código.

func TestTelefoneSemMais55NaoEhCPF(t *testing.T) {
	for _, chave := range []string{"(11) 98765-4321", "11987654321", "(41) 3333-4444"} {
		c := certo
		c.Chave = chave
		if r := Conferir(Entrada{Campos: c, Agora: agora}); r.Veredito == VereditoFalso {
			t.Errorf("%s deu falso: %+v", chave, r.Achados)
		}
	}
	// CPF escrito como CPF e com dígito errado continua impossível.
	c := certo
	c.Chave = "529.982.247-24"
	if r := Conferir(Entrada{Campos: c, Agora: agora}); r.Veredito != VereditoFalso {
		t.Errorf("CPF impossível: %s", r.Veredito)
	}
}

func TestNomeDePessoaNaoViraBanco(t *testing.T) {
	l := leitura.Ler("Comprovante de Pix\nPara\nCora Lima\nR$ 50,00\nID: E18236120202610291732abc12XYZ789\n29/10/2026 14:32")
	if l.Emissor != "" {
		t.Fatalf("emissor %q", l.Emissor)
	}
	if r := Conferir(Entrada{Campos: l, Agora: agora}); r.Veredito == VereditoFalso {
		t.Fatalf("%+v", r.Achados)
	}
}

func TestEntradasQueDerrubavam(t *testing.T) {
	if _, err := brcode.Ler("00020126-1br.gov.bcb.pix"); err == nil {
		t.Error("tamanho negativo deveria dar erro")
	}
	arquivo.Analisar("x.jpg", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0, 1, 2, 3})
	arquivo.Analisar("x.jpg", []byte{0xFF, 0xD8, 0xFF, 0xE1, 0, 1})
	if r := ConferirBoleto("123", agora); r.Resumo == nil || r.Achados == nil {
		t.Error("resumo e achados não podem ser nulos (a tela quebra)")
	}
}

func TestVencimentoNoProprioDiaANoite(t *testing.T) {
	// Boleto que vence hoje, conferido às 22h em Brasília (01h UTC do dia seguinte).
	noite := time.Date(2026, 10, 29, 22, 0, 0, 0, pix.Brasilia)
	r := ConferirBoleto(linhaQueVenceEm(t, noite), noite)
	vencido := false
	for _, a := range r.Achados {
		vencido = vencido || a.Codigo == "vencido"
	}
	if vencido {
		t.Fatalf("vence hoje, não está vencido: %+v", r.Boleto)
	}
}

func TestNomeComAcentoNoCopiaECola(t *testing.T) {
	// Gerador que conta o tamanho em caracteres: "JOSÉ" tem 4.
	semCRC := "000201" + "26330014br.gov.bcb.pix0111" + "52998224725" + "52040000" + "5303986" + "5802BR" + "5904JOSÉ" + "6008CURITIBA" + "6304"
	c, err := brcode.Ler(semCRC + "0000")
	if err != nil || c.Nome != "JOSÉ" || c.Cidade != "CURITIBA" {
		t.Fatalf("%v %+v", err, c)
	}
}

func TestPixFaseRestrita(t *testing.T) {
	c := certo
	c.ID, c.DataHora = "E18236120202011051732abc12XYZ789", ""
	if r := Conferir(Entrada{Campos: c, Agora: agora}); codigos(r, Falso)["antes-do-pix"] {
		t.Fatal("05/11/2020 já tinha Pix (fase restrita)")
	}
}

func TestIDComLetrasConfundidasEhAchado(t *testing.T) {
	// "i" e "D" na parte numérica, que a correção de leitura já conhece.
	if l := leitura.Ler("ID: E1823612O2026i0291732abc12XYZ789"); l.ID == "" {
		t.Fatal("código não achado")
	}
}

// linhaQueVenceEm monta uma linha de boleto bancário válida que vence no
// dia dado (fator do ciclo iniciado em 22/02/2025).
func linhaQueVenceEm(t *testing.T, dia time.Time) string {
	t.Helper()
	d := time.Date(dia.Year(), dia.Month(), dia.Day(), 0, 0, 0, 0, time.UTC)
	fator := 1000 + int(d.Sub(time.Date(2025, 2, 22, 0, 0, 0, 0, time.UTC)).Hours()/24)
	r := ConferirBoleto(montarBarras("237", fator, "0000015000", "1234567890123456789012345"), dia)
	if r.Boleto == nil || len(r.Boleto.Erros) > 0 {
		t.Fatalf("linha de teste inválida: %+v", r)
	}
	return r.Boleto.Linha
}

// montarBarras monta um código de barras bancário, achando o dígito geral
// por tentativa (a função de cálculo é interna do pacote boleto).
func montarBarras(banco string, fator int, valor, livre string) string {
	f := fmt.Sprintf("%04d", fator)
	for dv := '1'; dv <= '9'; dv++ {
		b := banco + "9" + string(dv) + f + valor + livre
		if r := ConferirBoleto(b, agora); r.Boleto != nil && len(r.Boleto.Erros) == 0 {
			return b
		}
	}
	return ""
}
