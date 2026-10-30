package conferir

import (
	"strings"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/bancos"
	"github.com/marianabsctba/Panthera_Claw/nucleo/boleto"
	"github.com/marianabsctba/Panthera_Claw/nucleo/brcode"
	"github.com/marianabsctba/Panthera_Claw/nucleo/pix"
)

// Vereditos de boleto e de Pix copia e cola.
const (
	CodigoErrado   = "errado"   // dígito ou soma de verificação não bate: não pague
	CodigoSuspeito = "suspeito" // algo estranho
	CodigoConfira  = "confira"  // código íntegro: confira quem vai receber
)

// Linha é um par rótulo e valor do resumo, para mostrar em letra grande.
type Linha struct {
	Rotulo string `json:"rotulo"`
	Valor  string `json:"valor"`
}

// ResultadoCodigo é a conferência de um boleto ou de um Pix copia e cola.
type ResultadoCodigo struct {
	Tipo     string         `json:"tipo"` // boleto ou pix
	Veredito string         `json:"veredito"`
	Resumo   []Linha        `json:"resumo"`
	Achados  []Achado       `json:"achados"`
	Boleto   *boleto.Boleto `json:"boleto,omitempty"`
	Pix      *brcode.Codigo `json:"pix,omitempty"`
}

// ConferirCodigo decide se o texto é Pix copia e cola ou linha de boleto.
func ConferirCodigo(s string, agora time.Time) ResultadoCodigo {
	if brcode.Parece(s) || strings.HasPrefix(strings.TrimSpace(s), "000201") {
		return ConferirPixCodigo(s)
	}
	return ConferirBoleto(s, agora)
}

// ConferirBoleto confere a linha digitável ou o código de barras.
func ConferirBoleto(s string, agora time.Time) ResultadoCodigo {
	r := ResultadoCodigo{Tipo: "boleto", Resumo: []Linha{}, Achados: []Achado{}}
	add := func(cod, peso, titulo, detalhe string) {
		r.Achados = append(r.Achados, Achado{cod, peso, titulo, detalhe})
	}
	b, err := boleto.Ler(s, agora)
	if err != nil {
		add("tamanho", Falso, "Isto não tem o tamanho de uma linha de boleto: "+err.Error()+".", "")
		r.Veredito = CodigoErrado
		return r
	}
	r.Boleto = &b
	for _, e := range b.Erros {
		add("dv", Falso, "A linha do boleto tem erro: "+e+". Se você copiou certinho, o boleto foi alterado. Não pague.", "Linha: "+b.Linha)
	}

	if b.Tipo == boleto.Bancario {
		nome := "banco " + b.Banco
		if bc, ok := bancos.PorCompe(b.Banco); ok {
			nome = bc.Nome
			if g := bancos.Marca(bc.ISPB); g != nil {
				nome = g.Marca
			}
			r.Resumo = append(r.Resumo, Linha{"Banco que vai receber", nome + " (" + b.Banco + ")"})
			add("banco", Info, "Este boleto paga para uma conta no "+nome+". Confira se é o mesmo banco do logotipo do boleto: golpistas trocam a linha para um banco deles.", "Número do banco: "+b.Banco+" ("+bc.Nome+")")
		} else {
			r.Resumo = append(r.Resumo, Linha{"Banco que vai receber", "número " + b.Banco + " (desconhecido)"})
			add("banco-desconhecido", Alerta, "O número do banco no boleto ("+b.Banco+") não está na lista do Banco Central.", "")
		}
		if b.Vencimento != "" {
			r.Resumo = append(r.Resumo, Linha{"Vencimento", b.Vencimento})
			// Compara datas no horário de Brasília: o boleto vence no fim do dia.
			hoje := agora.In(pix.Brasilia).Format("2006-01-02")
			if v, err := time.Parse("02/01/2006", b.Vencimento); err == nil && v.Format("2006-01-02") < hoje {
				add("vencido", Info, "O vencimento ("+b.Vencimento+") já passou. Pode haver juros, ou o boleto pode ter sido trocado por um antigo.", "")
			}
		}
	} else {
		r.Resumo = append(r.Resumo, Linha{"Tipo", "conta ou tributo (" + b.Segmento + ")"})
	}
	if b.Valor != "" {
		r.Resumo = append([]Linha{{"Valor", b.Valor}}, r.Resumo...)
	} else {
		add("sem-valor", Info, "O código não traz o valor: ele será pedido ou mostrado na hora de pagar.", "")
	}
	if len(b.Erros) == 0 {
		add("dv-ok", OK, "Os dígitos de conferência da linha estão certos.", "")
	}
	r.Veredito = vereditoCodigo(r.Achados)
	return r
}

// ConferirPixCodigo confere o Pix copia e cola (o texto do QR Code).
func ConferirPixCodigo(s string) ResultadoCodigo {
	r := ResultadoCodigo{Tipo: "pix", Resumo: []Linha{}, Achados: []Achado{}}
	add := func(cod, peso, titulo, detalhe string) {
		r.Achados = append(r.Achados, Achado{cod, peso, titulo, detalhe})
	}
	c, err := brcode.Ler(s)
	if err != nil {
		add("quebrado", Falso, "Este código Pix está quebrado ("+err.Error()+"). Não pague com ele.", "")
		r.Veredito = CodigoErrado
		return r
	}
	r.Pix = &c
	if !c.EhPix {
		add("nao-pix", Falso, "Este código não é de Pix.", "")
	}
	if !c.CRCValido {
		add("crc", Falso, "O código foi alterado ou está incompleto: a conferência final não bate. Não pague com ele.",
			"Soma no código: "+c.CRC+". Soma calculada: "+c.CRCCerto+".")
	} else {
		add("crc-ok", OK, "O código está inteiro: a conferência final bate.", "")
	}

	if c.Nome != "" {
		r.Resumo = append(r.Resumo, Linha{"Quem vai receber", c.Nome})
	} else {
		add("sem-nome", Alerta, "O código não diz o nome de quem vai receber.", "")
	}
	if c.Cidade != "" {
		r.Resumo = append(r.Resumo, Linha{"Cidade", c.Cidade})
	}
	if v := c.ValorReais(); v != "" {
		r.Resumo = append([]Linha{{"Valor", v}}, r.Resumo...)
	}
	if c.Chave != "" {
		r.Resumo = append(r.Resumo, Linha{"Chave Pix", c.Chave})
		ch := pix.LerChave(c.Chave)
		switch {
		case ch.Tipo == pix.ChaveInvalida && ch.Impossivel:
			add("chave-invalida", Falso, "A chave Pix deste código não pode existir: "+ch.Motivo+".", "")
		case ch.Tipo == pix.ChaveInvalida:
			add("chave-estranha", Alerta, "A chave Pix deste código tem um formato estranho.", "Chave: "+c.Chave)
		}
	}
	if c.Dinamico && c.Dominio != "" {
		r.Resumo = append(r.Resumo, Linha{"Cobrança gerada por", c.Dominio})
		add("dinamico", Info, "É uma cobrança: o valor e o nome de quem recebe vêm do banco na hora de pagar. Confira o nome que aparece no seu aplicativo antes de confirmar.", "Endereço da cobrança: "+c.URL)
	}
	add("confira-nome", Info, "Antes de pagar, confira se o nome de quem vai receber é de quem você espera. No golpe do QR Code trocado, tudo funciona, mas o dinheiro vai para outra pessoa.", "")
	r.Veredito = vereditoCodigo(r.Achados)
	return r
}

func vereditoCodigo(as []Achado) string {
	alerta := false
	for _, a := range as {
		if a.Peso == Falso {
			return CodigoErrado
		}
		if a.Peso == Alerta {
			alerta = true
		}
	}
	if alerta {
		return CodigoSuspeito
	}
	return CodigoConfira
}
