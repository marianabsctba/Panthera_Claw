package leitura

import (
	"testing"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/pix"
)

// Textos de exemplo montados à mão, em formatos comuns. Não são
// comprovantes reais.
const comRotulosEmLinhas = `nubank
Comprovante de transferência
29 OUT 2026 - 14:32:10
Valor
R$ 1.250,00
Tipo de transferência
Pix
Destino
Nome
Maria da Silva
CPF
•••.456.789-••
Instituição
BCO DO BRASIL S.A.
Chave Pix
maria@exemplo.com.br
Origem
Nome
João Souza
Instituição
NU PAGAMENTOS - IP
ID da transação:
E18236120202610291732
abc12XYZ789`

const emColunas = `Itaú
comprovante de pix
valor:  R$ 89,90
data:  29/10/2026 às 09:05
dados de quem recebeu
nome:  Padaria Pão Bom
instituição:  PAGSEGURO INTERNET IP S.A.
dados de quem pagou
nome:  Ana Lima
instituição:  ITAÚ UNIBANCO S.A.
ID da transação:  E60701190202610291205aBcDeFgHiJk`

func TestLerRotulosEmLinhas(t *testing.T) {
	c := Ler(comRotulosEmLinhas)
	quer := Campos{
		ID: "E18236120202610291732abc12XYZ789", Valor: "R$ 1.250,00", DataHora: "29 OUT 2026 - 14:32:10",
		BancoPagador: "NU PAGAMENTOS - IP", BancoRecebedor: "BCO DO BRASIL S.A.",
		NomePagador: "João Souza", NomeRecebedor: "Maria da Silva", Chave: "maria@exemplo.com.br", Emissor: "Nubank",
	}
	if c != quer {
		t.Fatalf("\nlido %+v\nquer %+v", c, quer)
	}
	d, ok := LerDataHora(c.DataHora)
	if !ok || !d.TemHora || !d.Instante.Equal(time.Date(2026, 10, 29, 14, 32, 10, 0, pix.Brasilia)) {
		t.Fatalf("data %+v %v", d, ok)
	}
}

func TestLerEmColunas(t *testing.T) {
	c := Ler(emColunas)
	if c.ID != "E60701190202610291205aBcDeFgHiJk" || c.Valor != "R$ 89,90" || c.DataHora != "29/10/2026 às 09:05" {
		t.Fatalf("%+v", c)
	}
	if c.BancoPagador != "ITAÚ UNIBANCO S.A." || c.BancoRecebedor != "PAGSEGURO INTERNET IP S.A." || c.Emissor != "Itaú" {
		t.Fatalf("%+v", c)
	}
}

func TestBancoDoBrasilNaoViraRotulo(t *testing.T) {
	c := Ler("Origem\nBanco do Brasil\nDestino\nInstituição: Caixa")
	if c.BancoPagador != "" || c.BancoRecebedor != "Caixa" {
		t.Fatalf("%+v", c)
	}
}

func TestDatas(t *testing.T) {
	casos := map[string]string{
		"29/10/2026 14:32":                            "2026-10-29 14:32",
		"29/10/26 às 14:32:10":                        "2026-10-29 14:32",
		"Realizado em 29 de outubro de 2026 às 14h32": "2026-10-29 14:32",
		"29 OUT 2026 - 14:32:10":                      "2026-10-29 14:32",
		"05/03/2026":                                  "2026-03-05 00:00",
	}
	for s, quer := range casos {
		d, ok := LerDataHora(s)
		if !ok || d.Instante.Format("2006-01-02 15:04") != quer {
			t.Errorf("LerDataHora(%q) = %v %v, quero %s", s, d.Instante, ok, quer)
		}
	}
	if _, ok := LerDataHora("31/02/2026"); ok {
		t.Error("31/02 não existe")
	}
}

func TestArrobaLidaErrado(t *testing.T) {
	for _, s := range []string{"maria(Qexemplo.com.br", "maria (@ exemplo.com.br", "maria©exemplo.com", "mariaQexemplo.com.br", "maria Qexemplo.com.br"} {
		if c := corrigirEmail(s); c != "maria@exemplo.com.br" && c != "maria@exemplo.com" {
			t.Errorf("%q virou %q", s, c)
		}
	}
	if c := corrigirEmail("QUALQUERQcoisa.com"); c != "QUALQUERQcoisa.com" {
		t.Errorf("mexeu em texto com maiúsculas: %q", c)
	}
	if c := corrigirEmail("12345678900"); c != "12345678900" {
		t.Errorf("mexeu em CPF: %q", c)
	}
}

// Como a leitura de imagem devolve um comprovante em colunas: rótulo e valor
// na mesma linha, separados por um espaço só.
const colunasLidasDaImagem = `14:33 4G 87%
Comprovante de transferência
29 out 2026 às 14:32
Valor
R$ 1.250,00
Destino
Nome Maria da Silva
CPF •••.456.789-••
Instituição BCO DO BRASIL S.A.
Chave Pix maria@exemplo.com.br
Origem
Nome João Souza
Instituição ITAÚ UNIBANCO S.A.
Agência 0001
ID da transação
E18236120202610291732
k7Hq2Lm9Pzx`

func TestColunasLidasDaImagem(t *testing.T) {
	c := Ler(colunasLidasDaImagem)
	if c.BancoPagador != "ITAÚ UNIBANCO S.A." || c.BancoRecebedor != "BCO DO BRASIL S.A." ||
		c.Chave != "maria@exemplo.com.br" || c.NomePagador != "João Souza" || c.ID != "E18236120202610291732k7Hq2Lm9Pzx" {
		t.Fatalf("%+v", c)
	}
}

func TestIDDuvidoso(t *testing.T) {
	if Ler(colunasLidasDaImagem).IDDuvidoso {
		t.Error("código inteiro marcado como duvidoso")
	}
	if !Ler("E18236120202610291732\nk7Hqa2Lm9Pzx").IDDuvidoso {
		t.Error("final com 12 caracteres deveria ser duvidoso")
	}
}

func TestSituacao(t *testing.T) {
	casos := map[string]string{
		"Comprovante de agendamento\nPix agendado para 30/10/2026": SituacaoAgendado,
		"Status: Em processamento":                                 SituacaoPendente,
		"Transferência cancelada":                                  SituacaoCancelado,
		"Comprovante de devolução":                                 SituacaoDevolvido,
		"Comprovante de transferência\nPix enviado":                "",
		"Sem pendências":                                           "",
	}
	for texto, quer := range casos {
		if s, _ := LerSituacao(texto); s != quer {
			t.Errorf("%q: %q, quero %q", texto, s, quer)
		}
	}
}

func TestValoresSemDuasCasas(t *testing.T) {
	casos := map[string]string{"Valor R$ 1,5": "R$ 1,50", "Valor: R$ 150": "R$ 150,00", "Total R$ 15,0": "R$ 15,00", "R$1.250,00": "R$ 1.250,00"}
	for texto, quer := range casos {
		if c := Ler(texto); c.Valor != quer {
			t.Errorf("%q: %q, quero %q", texto, c.Valor, quer)
		}
	}
}
