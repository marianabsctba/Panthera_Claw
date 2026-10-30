package conferir

import (
	"testing"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/arquivo"
	"github.com/marianabsctba/Panthera_Claw/nucleo/leitura"
	"github.com/marianabsctba/Panthera_Claw/nucleo/pix"
)

var agora = time.Date(2026, 10, 29, 18, 0, 0, 0, pix.Brasilia)

// Nubank, 29/10/2026 14:32 em Brasília = 17:32 UTC.
var certo = leitura.Campos{
	ID: "E18236120202610291732abc12XYZ789", Valor: "R$ 150,00", DataHora: "29/10/2026 14:32:40",
	BancoPagador: "NU PAGAMENTOS - IP", Chave: "maria@exemplo.com.br",
}

func codigos(r Resultado, peso string) map[string]bool {
	m := map[string]bool{}
	for _, a := range r.Achados {
		if a.Peso == peso {
			m[a.Codigo] = true
		}
	}
	return m
}

func TestComprovanteCoerente(t *testing.T) {
	r := Conferir(Entrada{Campos: certo, Agora: agora})
	if r.Veredito != VereditoSemErros {
		t.Fatalf("%s %+v", r.Veredito, r.Achados)
	}
	ok := codigos(r, OK)
	for _, c := range []string{"id-ok", "banco-ok", "horario-ok", "chave-ok"} {
		if !ok[c] {
			t.Errorf("faltou %s", c)
		}
	}
	if r.Codigo.Marca != "Nubank" || r.Codigo.Brasilia != "29/10/2026 às 14:32" {
		t.Errorf("%+v", r.Codigo)
	}
}

func TestBancoDiferente(t *testing.T) {
	c := certo
	c.ID = "E00000000202610291732abc12XYZ789" // Banco do Brasil
	r := Conferir(Entrada{Campos: c, Agora: agora})
	if r.Veredito != VereditoFalso || !codigos(r, Falso)["banco-diferente"] {
		t.Fatalf("%s %+v", r.Veredito, r.Achados)
	}
}

func TestBancoPeloEmissor(t *testing.T) {
	c := certo
	c.BancoPagador, c.Emissor = "", "Itaú"
	r := Conferir(Entrada{Campos: c, Agora: agora})
	if !codigos(r, Falso)["banco-diferente"] {
		t.Fatalf("%+v", r.Achados)
	}
}

func TestHorarios(t *testing.T) {
	casos := map[string]string{
		"29/10/2026 14:33": "horario-ok",
		"29/10/2026 17:32": "horario-brasilia", // o falsário pôs o horário de Brasília no código
		"29/10/2026 13:32": "horario-fuso",
		"29/10/2026 15:10": "horario-proximo",
		"29/10/2026 08:05": "horario-diferente",
		"29/10/2026":       "data-ok",
		"28/10/2026":       "data-diferente",
	}
	for data, quer := range casos {
		c := certo
		c.DataHora = data
		r := Conferir(Entrada{Campos: c, Agora: agora})
		achou := false
		for _, a := range r.Achados {
			if a.Codigo == quer {
				achou = true
			}
		}
		if !achou {
			t.Errorf("%s: queria %s, veio %+v", data, quer, r.Achados)
		}
	}
	c := certo
	c.DataHora = "29/10/2026 17:32"
	if r := Conferir(Entrada{Campos: c, Agora: agora}); r.Veredito != VereditoFalso {
		t.Errorf("horário de Brasília no código tem que dar falso: %s", r.Veredito)
	}
}

func TestFuturoEAntesDoPix(t *testing.T) {
	c := certo
	c.ID, c.DataHora = "E18236120202612011732abc12XYZ789", ""
	if r := Conferir(Entrada{Campos: c, Agora: agora}); !codigos(r, Falso)["futuro"] {
		t.Errorf("futuro: %+v", r.Achados)
	}
	c.ID = "E18236120201901011732abc12XYZ789"
	if r := Conferir(Entrada{Campos: c, Agora: agora}); !codigos(r, Falso)["antes-do-pix"] {
		t.Errorf("antes do Pix: %+v", r.Achados)
	}
}

func TestIDInvalidoEAusente(t *testing.T) {
	c := certo
	c.ID = "E1823612020261029"
	if r := Conferir(Entrada{Campos: c, Agora: agora}); r.Veredito != VereditoFalso {
		t.Errorf("%s", r.Veredito)
	}
	c.ID, c.Chave = "", ""
	if r := Conferir(Entrada{Campos: c, Agora: agora}); r.Veredito != VereditoIncompleto {
		t.Errorf("%s", r.Veredito)
	}
}

func TestChaveInvalida(t *testing.T) {
	c := certo
	c.Chave = "529.982.247-24"
	if r := Conferir(Entrada{Campos: c, Agora: agora}); !codigos(r, Falso)["chave-invalida"] {
		t.Errorf("%+v", r.Achados)
	}
}

func TestArquivoEditado(t *testing.T) {
	a := &arquivo.Info{Tipo: "pdf", Editor: "Canva", Programas: []string{"Canva"}, Salvamentos: 2, SHA256: "x",
		Alteracao: &arquivo.Alteracao{Objetos: 2, Pagina: true}}
	r := Conferir(Entrada{Campos: certo, Arquivo: a, Agora: agora})
	al := codigos(r, Alerta)
	if r.Veredito != VereditoSuspeito || !al["editor"] || !al["pdf-alterado"] {
		t.Fatalf("%s %+v", r.Veredito, r.Achados)
	}
}

func TestRepetido(t *testing.T) {
	a := &arquivo.Info{Tipo: "png", SHA256: "novo"}
	vistos := []Visto{{ID: certo.ID, SHA256: "antigo", Valor: "R$ 1.500,00", Quando: "28/10/2026"}}
	if r := Conferir(Entrada{Campos: certo, Arquivo: a, Vistos: vistos, Agora: agora}); !codigos(r, Falso)["id-repetido"] {
		t.Fatalf("%+v", r.Achados)
	}
	vistos[0].Valor = "R$ 150,00"
	if r := Conferir(Entrada{Campos: certo, Arquivo: a, Vistos: vistos, Agora: agora}); r.Veredito != VereditoSemErros {
		t.Fatalf("mesmo Pix em outro arquivo não é fraude: %+v", r.Achados)
	}
}

func TestChaveMalLidaNaoEhFalso(t *testing.T) {
	c := certo
	c.Chave = "maria(Qexemplo.com.br"
	r := Conferir(Entrada{Campos: c, Agora: agora})
	if r.Veredito != VereditoSuspeito || !codigos(r, Alerta)["chave-estranha"] {
		t.Fatalf("%s %+v", r.Veredito, r.Achados)
	}
}

func TestPDFComValorEscritoDepois(t *testing.T) {
	a := &arquivo.Info{Tipo: "pdf", Salvamentos: 2, SHA256: "x",
		Alteracao: &arquivo.Alteracao{Objetos: 2, Pagina: true, Textos: []string{"R$ 1.500,00"}}}
	if r := Conferir(Entrada{Campos: certo, Arquivo: a, Agora: agora}); !codigos(r, Falso)["pdf-valor-alterado"] {
		t.Fatalf("%+v", r.Achados)
	}
	a.Alteracao = &arquivo.Alteracao{Objetos: 1}
	if r := Conferir(Entrada{Campos: certo, Arquivo: a, Agora: agora}); r.Veredito != VereditoSemErros {
		t.Fatalf("salvar de novo sem mexer na página não é suspeito: %+v", r.Achados)
	}
}

func TestValorPorBaixo(t *testing.T) {
	e := Entrada{Campos: certo, Agora: agora, Sobrepostos: []Sobreposicao{{Cima: "R$ 150,00", Baixo: "R$ 1,50"}}}
	if r := Conferir(e); !codigos(r, Falso)["valor-por-baixo"] {
		t.Fatalf("%+v", r.Achados)
	}
	e.Sobrepostos = []Sobreposicao{{Cima: "Maria", Baixo: "José"}}
	if r := Conferir(e); !codigos(r, Alerta)["texto-por-baixo"] {
		t.Fatalf("%+v", r.Achados)
	}
}

func TestArquivoDeIA(t *testing.T) {
	a := &arquivo.Info{Tipo: "png", SHA256: "x", IA: &arquivo.MarcaIA{Tipo: "gerada", Origem: "c2pa", Gerador: "ChatGPT"}}
	r := Conferir(Entrada{Campos: certo, Arquivo: a, Agora: agora})
	if r.Veredito != VereditoFalso || !codigos(r, Falso)["ia-gerada"] {
		t.Fatalf("%+v", r.Achados)
	}
}

func TestNaoPago(t *testing.T) {
	c := certo
	c.Situacao, c.SituacaoTrecho = "agendado", "Pix agendado para 30/10/2026"
	r := Conferir(Entrada{Campos: c, Agora: agora})
	if r.Veredito != VereditoNaoPago || !codigos(r, NaoPago)["situacao-agendado"] {
		t.Fatalf("%s %+v", r.Veredito, r.Achados)
	}
	// Falso continua mais forte que agendado.
	c.Chave = "529.982.247-24"
	if r := Conferir(Entrada{Campos: c, Agora: agora}); r.Veredito != VereditoFalso {
		t.Fatalf("%s", r.Veredito)
	}
}

func TestConferirBoleto(t *testing.T) {
	r := ConferirCodigo("00190.50095 40144.816069 06809.350314 3 37370000000100", agora)
	if r.Tipo != "boleto" || r.Veredito != CodigoConfira || len(r.Resumo) == 0 || r.Resumo[0].Valor != "R$ 1,00" {
		t.Fatalf("%+v", r)
	}
	if r := ConferirCodigo("00190.50095 40144.816069 06809.350314 3 37370000000900", agora); r.Veredito != CodigoErrado {
		t.Fatalf("valor trocado: %+v", r)
	}
}

func TestConferirPixCodigo(t *testing.T) {
	ok := "00020126580014br.gov.bcb.pix0136123e4567-e12b-12d1-a456-4266554400005204000053039865802BR5913Fulano de Tal6008BRASILIA62070503***63041D3D"
	r := ConferirCodigo(ok, agora)
	if r.Tipo != "pix" || r.Veredito != CodigoConfira || r.Resumo[0].Valor != "Fulano de Tal" {
		t.Fatalf("%+v", r)
	}
	trocado := ok[:40] + "9" + ok[41:]
	if r := ConferirCodigo(trocado, agora); r.Veredito != CodigoErrado {
		t.Fatalf("%+v", r)
	}
}
