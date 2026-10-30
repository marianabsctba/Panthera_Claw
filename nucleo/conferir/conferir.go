// Package conferir junta tudo e dá o resultado: compara o que está impresso
// no comprovante com o que o código da transação revela, e olha o arquivo.
//
// A regra de ouro: um erro de coerência prova que o comprovante é falso;
// a falta de erros não prova que é verdadeiro. Só o extrato confirma que o
// dinheiro entrou.
package conferir

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/arquivo"
	"github.com/marianabsctba/Panthera_Claw/nucleo/bancos"
	"github.com/marianabsctba/Panthera_Claw/nucleo/dinheiro"
	"github.com/marianabsctba/Panthera_Claw/nucleo/leitura"
	"github.com/marianabsctba/Panthera_Claw/nucleo/pix"
)

// Pesos de um achado.
const (
	Falso   = "falso"    // impossível num comprovante verdadeiro
	Alerta  = "alerta"   // suspeito, mas pode ter explicação
	NaoPago = "nao_pago" // agendado, pendente, cancelado ou devolvido
	OK      = "ok"       // conferido e certo
	Info    = "info"     // informação, sem peso
)

// Vereditos.
const (
	VereditoFalso      = "falso"
	VereditoSuspeito   = "suspeito"
	VereditoSemErros   = "sem_erros"
	VereditoIncompleto = "incompleto"
	VereditoNaoPago    = "nao_pago"
)

// Achado é uma conferência, com uma frase simples e o detalhe técnico.
type Achado struct {
	Codigo  string `json:"codigo"`
	Peso    string `json:"peso"`
	Titulo  string `json:"titulo"`
	Detalhe string `json:"detalhe,omitempty"`
}

// Visto é um comprovante conferido antes neste aparelho.
type Visto struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
	Valor  string `json:"valor"`
	Quando string `json:"quando"`
}

// Entrada é tudo o que a conferência recebe.
type Entrada struct {
	Campos  leitura.Campos `json:"campos"`
	Arquivo *arquivo.Info  `json:"arquivo,omitempty"`
	// PDFSemTexto indica um PDF que é só uma imagem (foto ou print
	// colado), sem o texto que o aplicativo do banco grava.
	PDFSemTexto bool `json:"pdfSemTexto"`
	// Sobrepostos são textos do PDF desenhados um em cima do outro, no
	// mesmo lugar: o jeito comum de editar um PDF é cobrir o valor antigo
	// com um retângulo e escrever o novo por cima.
	Sobrepostos []Sobreposicao `json:"sobrepostos,omitempty"`
	Vistos      []Visto        `json:"vistos,omitempty"`
	Agora       time.Time      `json:"agora"`
}

// Sobreposicao é um par de textos no mesmo lugar da página.
type Sobreposicao struct {
	Cima  string `json:"cima"`
	Baixo string `json:"baixo"`
}

// Codigo é o que o código da transação revela.
type Codigo struct {
	pix.E2E
	Banco     string `json:"banco,omitempty"`
	Marca     string `json:"marca,omitempty"`
	Brasilia  string `json:"brasilia"`
	UTC       string `json:"utc"`
	Conhecido bool   `json:"conhecido"`
}

// Resultado é a conclusão da conferência.
type Resultado struct {
	Veredito string         `json:"veredito"`
	Achados  []Achado       `json:"achados"`
	Codigo   *Codigo        `json:"codigo,omitempty"`
	Campos   leitura.Campos `json:"campos"`
	Arquivo  *arquivo.Info  `json:"arquivo,omitempty"`
}

const formato = "02/01/2006 às 15:04"

// Conferir faz todas as verificações.
func Conferir(e Entrada) Resultado {
	if e.Agora.IsZero() {
		e.Agora = time.Now()
	}
	r := Resultado{Campos: e.Campos, Arquivo: e.Arquivo}
	add := func(cod, peso, titulo, detalhe string) {
		r.Achados = append(r.Achados, Achado{cod, peso, titulo, detalhe})
	}

	c := e.Campos
	var cod *pix.E2E
	if strings.TrimSpace(c.ID) == "" {
		add("id-ausente", Info, "Não encontramos o código da transação (ID), que é o que mais ajuda a conferir.",
			"O código tem 32 caracteres e começa com a letra E.")
	} else if e2e, err := pix.LerE2E(c.ID); err != nil {
		add("id-invalido", Falso, "O código da transação não segue o padrão do Pix: "+err.Error()+".",
			"Código lido: "+pix.Normalizar(c.ID))
	} else {
		cod = &e2e
		r.Codigo = descrever(e2e)
		det := "Código " + e2e.Texto
		if len(e2e.Correcoes) > 0 {
			det += ". Ajustes de leitura: " + strings.Join(e2e.Correcoes, "; ")
		}
		add("id-ok", OK, "O código da transação segue o padrão do Pix.", det)
		conferirCodigo(e, e2e, r.Codigo, add)
	}

	if strings.TrimSpace(c.Chave) != "" {
		ch := pix.LerChave(c.Chave)
		switch ch.Tipo {
		case pix.ChaveInvalida:
			if ch.Impossivel {
				add("chave-invalida", Falso, "A chave Pix do comprovante não pode existir: "+ch.Motivo+".", "Chave: "+c.Chave)
			} else {
				add("chave-estranha", Alerta, "A chave Pix está escrita de um jeito estranho. Confira se foi lida certo.", "Chave: "+c.Chave+" ("+ch.Motivo+")")
			}
		case pix.ChaveMascarada:
			add("chave-mascarada", Info, "A chave Pix aparece parcialmente escondida, como de costume.", "Chave: "+c.Chave)
		default:
			add("chave-ok", OK, "A chave Pix tem formato válido ("+nomeChave[ch.Tipo]+").", "Chave: "+c.Chave)
		}
	}

	if c.Situacao != "" {
		add("situacao-"+c.Situacao, NaoPago, frasesSituacao[c.Situacao], "No comprovante: "+c.SituacaoTrecho)
	}

	if c.Valor != "" && valorZero(c.Valor) {
		add("valor-zero", Alerta, "O valor do comprovante é zero.", "Valor: "+c.Valor)
	}

	conferirArquivo(e, add)
	conferirSobrepostos(e, add)
	conferirVistos(e, cod, add)

	r.Veredito = veredito(r.Achados, cod != nil)
	return r
}

var frasesSituacao = map[string]string{
	leitura.SituacaoAgendado:  "O comprovante é de um Pix agendado, não de um pagamento feito. Um agendamento pode ser cancelado a qualquer momento antes da data.",
	leitura.SituacaoPendente:  "O comprovante mostra um Pix que ainda não foi concluído (pendente ou em processamento). O dinheiro pode não chegar.",
	leitura.SituacaoCancelado: "O comprovante mostra um Pix cancelado ou que não foi realizado.",
	leitura.SituacaoDevolvido: "O comprovante é de um Pix devolvido, não de um pagamento.",
}

var nomeChave = map[string]string{
	pix.ChaveCPF: "CPF", pix.ChaveCNPJ: "CNPJ", pix.ChaveTelefone: "telefone",
	pix.ChaveEmail: "e-mail", pix.ChaveAleatoria: "chave aleatória",
}

func descrever(e pix.E2E) *Codigo {
	d := &Codigo{E2E: e, Brasilia: e.NoHorarioDeBrasilia().Format(formato), UTC: e.Instante.Format("02/01/2006 15:04") + " UTC"}
	if b, ok := bancos.PorISPB(e.ISPB); ok && b.Situacao != "str" {
		d.Banco, d.Conhecido = b.Nome, true
		if g := bancos.Marca(e.ISPB); g != nil {
			d.Marca = g.Marca
		}
	}
	return d
}

func conferirCodigo(e Entrada, e2e pix.E2E, d *Codigo, add func(string, string, string, string)) {
	c := e.Campos
	if e2e.Tipo == "D" {
		add("devolucao", Alerta, "O código é de uma devolução de Pix, não de um pagamento.", "O código começa com D.")
	}
	switch {
	case e2e.Instante.After(e.Agora.Add(10 * time.Minute)):
		add("futuro", Falso, "A data escondida no código da transação ainda não chegou ("+d.Brasilia+").", "Instante do código: "+d.UTC)
	case e2e.Instante.Before(pix.Lancamento):
		add("antes-do-pix", Falso, "A data escondida no código ("+d.Brasilia+") é de antes de o Pix existir.", "O Pix começou em 03/11/2020 (fase restrita) e foi aberto a todos em 16/11/2020.")
	}

	// Banco: o código diz quem gerou a transação.
	if !d.Conhecido {
		add("banco-desconhecido", Alerta, "O código aponta para uma instituição que não está na lista oficial do Pix.",
			"ISPB "+e2e.ISPB+" não consta na lista de participantes do Banco Central.")
	} else {
		nomeBanco := d.Banco
		if d.Marca != "" {
			nomeBanco = d.Marca
		}
		impresso, origemImpresso := c.BancoPagador, "banco de quem pagou"
		if impresso == "" {
			impresso, origemImpresso = c.Emissor, "aplicativo que gerou o comprovante"
		}
		g := bancos.Reconhecer(impresso)
		switch {
		case impresso == "":
			add("banco-codigo", Info, "Pelo código, o Pix saiu do "+nomeBanco+".", "ISPB "+e2e.ISPB+": "+d.Banco)
		case g == nil:
			add("banco-codigo", Info, "Pelo código, o Pix saiu do "+nomeBanco+". Não reconhecemos o banco escrito no comprovante ("+impresso+").",
				"ISPB "+e2e.ISPB+": "+d.Banco)
		case !g.ISPBs[e2e.ISPB] && c.BancoPagador != "":
			add("banco-diferente", Falso, "O comprovante diz que o Pix saiu do "+g.Marca+", mas o código da transação é do "+nomeBanco+".",
				"Comparado com o "+origemImpresso+": "+impresso+". ISPB "+e2e.ISPB+": "+d.Banco)
		case !g.ISPBs[e2e.ISPB]:
			// A marca do topo é um indício mais fraco que o banco escrito na
			// parte de quem pagou: o comprovante pode ser do recebedor.
			add("banco-diferente-topo", Alerta, "O comprovante parece ser do "+g.Marca+", mas o código da transação é do "+nomeBanco+".",
				"Comparado com o "+origemImpresso+": "+impresso+". ISPB "+e2e.ISPB+": "+d.Banco)
		default:
			add("banco-ok", OK, "O código da transação é do mesmo banco do comprovante ("+g.Marca+").", "ISPB "+e2e.ISPB+": "+d.Banco)
		}
	}

	if c.DataHora == "" {
		return
	}
	dh, ok := leitura.LerDataHora(c.DataHora)
	if !ok {
		add("data-ilegivel", Info, "Não conseguimos entender a data do comprovante.", "Data lida: "+c.DataHora)
		return
	}
	idBR := e2e.NoHorarioDeBrasilia()
	if !dh.TemHora {
		a, m, dd := dh.Instante.Date()
		b, n, ee := idBR.Date()
		if a == b && m == n && dd == ee {
			add("data-ok", OK, "A data do código bate com a do comprovante.", "Código: "+d.Brasilia)
			return
		}
		meiaNoite := idBR.Hour() < 2 || idBR.Hour() >= 22
		peso := Falso
		if meiaNoite {
			peso = Alerta
		}
		add("data-diferente", peso, "A data escondida no código ("+idBR.Format("02/01/2006")+") não bate com a do comprovante ("+dh.Instante.Format("02/01/2006")+").",
			"Código: "+d.Brasilia+" (Brasília)")
		return
	}

	// O código guarda o minuto em que a transação começou; o comprovante
	// costuma mostrar o mesmo minuto ou o seguinte.
	dif := dh.Instante.Sub(e2e.Instante)
	perto := func(alvo time.Duration) bool { return abs(dif-alvo) <= 5*time.Minute }
	impresso := dh.Instante.Format("15:04")
	doCodigo := idBR.Format("15:04")
	det := fmt.Sprintf("Código: %s (Brasília), %s. Comprovante: %s.", d.Brasilia, d.UTC, dh.Instante.Format(formato))
	switch {
	case perto(0):
		add("horario-ok", OK, "O horário escondido no código bate com o do comprovante ("+impresso+").", det)
	case perto(3 * time.Hour):
		add("horario-brasilia", Falso, "O código da transação foi montado com o horário de Brasília, mas o Pix usa sempre o horário universal (UTC). Isso é típico de comprovante falsificado.", det)
	case perto(-time.Hour) || perto(-2*time.Hour) || perto(time.Hour):
		add("horario-fuso", Alerta, "O horário do comprovante ("+impresso+") difere do código ("+doCodigo+" em Brasília) em horas exatas. Pode ser outro fuso do Brasil, mas os bancos costumam usar o de Brasília.", det)
	case abs(dif) <= 2*time.Hour:
		add("horario-proximo", Alerta, fmt.Sprintf("O horário do comprovante (%s) está %s distante do horário escondido no código (%s).", impresso, duracao(abs(dif)), doCodigo), det)
	default:
		add("horario-diferente", Falso, "O horário escondido no código da transação ("+d.Brasilia+") não bate com o do comprovante ("+dh.Instante.Format(formato)+").", det)
	}
}

func conferirArquivo(e Entrada, add func(string, string, string, string)) {
	a := e.Arquivo
	if a == nil {
		return
	}
	if m := a.IA; m != nil {
		quem := ""
		if m.Gerador != "" {
			quem = " (" + m.Gerador + ")"
		}
		det := map[string]string{
			"c2pa":     "Selo de procedência C2PA (Content Credentials) dentro do arquivo.",
			"iptc":     "Etiqueta IPTC de origem digital: mídia gerada por algoritmo treinado.",
			"prompt":   "O pedido (prompt) usado para gerar a imagem está gravado no arquivo.",
			"programa": "Programa registrado no arquivo: " + strings.Join(a.Programas, "; "),
		}[m.Origem]
		if m.Tipo == "editada" {
			add("ia-editada", Falso, "O próprio arquivo declara que foi editado com inteligência artificial"+quem+".", det)
		} else {
			add("ia-gerada", Falso, "O próprio arquivo declara que foi criado por inteligência artificial"+quem+". Comprovante de banco não é feito assim.", det)
		}
	}
	if a.Editor != "" {
		add("editor", Alerta, "O arquivo passou por um programa de edição, e não veio direto do aplicativo do banco.",
			"Programa registrado no arquivo: "+a.Editor+" ("+strings.Join(a.Programas, "; ")+")")
	}
	if a.Tipo == "pdf" {
		if a.Salvamentos > 1 && !a.Assinado {
			conferirAlteracao(e, a, add)
		}
		if e.PDFSemTexto {
			add("pdf-imagem", Alerta, "O PDF é só uma imagem, sem o texto que o aplicativo do banco costuma gravar.", "Nenhum texto selecionável no documento.")
		}
	}
}

func conferirAlteracao(e Entrada, a *arquivo.Info, add func(string, string, string, string)) {
	det := fmt.Sprintf("%d versões gravadas no mesmo arquivo.", a.Salvamentos)
	alt := a.Alteracao
	if alt == nil || !alt.Pagina {
		add("pdf-salvo-de-novo", Info, "O PDF foi salvo de novo depois de criado, mas sem mudar o que aparece na página.", det)
		return
	}
	if len(alt.Textos) > 0 {
		det += " Texto escrito na versão nova: " + strings.Join(alt.Textos, " | ")
	}
	for _, t := range alt.Textos {
		if _, ok := valorNoTexto(t); ok {
			add("pdf-valor-alterado", Falso, "O valor foi escrito no PDF depois que ele foi criado ("+strings.TrimSpace(t)+"). Comprovante de banco não é alterado assim.", det)
			return
		}
	}
	add("pdf-alterado", Alerta, "O PDF foi alterado depois de criado: algo foi escrito ou desenhado na página por cima do original.", det)
}

func conferirSobrepostos(e Entrada, add func(string, string, string, string)) {
	for _, s := range e.Sobrepostos {
		cima, ok1 := valorNoTexto(s.Cima)
		baixo, ok2 := valorNoTexto(s.Baixo)
		if ok1 && ok2 && cima != baixo {
			add("valor-por-baixo", Falso, "Há outro valor escondido por baixo do valor do comprovante: por baixo de "+s.Cima+" está escrito "+s.Baixo+".",
				"Dois textos no mesmo lugar da página. O de cima foi escrito depois.")
			return
		}
	}
	if len(e.Sobrepostos) > 0 {
		s := e.Sobrepostos[0]
		add("texto-por-baixo", Alerta, "Há texto escondido por baixo de outro texto no PDF, sinal de que algo foi escrito por cima.",
			fmt.Sprintf("Por cima: %q. Por baixo: %q.", s.Cima, s.Baixo))
	}
}

// Um valor em reais dentro de um texto: com "R$" (centavos opcionais) ou
// com vírgula decimal de uma ou duas casas.
var reValorTexto = regexp.MustCompile(`R\$\s*\d{1,3}(?:\.\d{3})+(?:,\d{1,2})?|R\$\s*\d+(?:,\d{1,2})?|\d{1,3}(?:\.\d{3})+,\d{1,2}|\d+,\d{1,2}`)

// valorNoTexto acha o primeiro valor em reais no texto, em centavos.
func valorNoTexto(t string) (int64, bool) {
	for _, m := range reValorTexto.FindAllStringIndex(t, -1) {
		// Não vale pedaço de número maior ("1,505" ou "12,34,56").
		if m[1] < len(t) && t[m[1]] >= '0' && t[m[1]] <= '9' {
			continue
		}
		if c, ok := dinheiro.Centavos(t[m[0]:m[1]]); ok {
			return c, true
		}
	}
	return 0, false
}

func conferirVistos(e Entrada, cod *pix.E2E, add func(string, string, string, string)) {
	if cod == nil {
		return
	}
	for _, v := range e.Vistos {
		if pix.Normalizar(v.ID) != cod.Texto {
			continue
		}
		if e.Arquivo != nil && v.SHA256 == e.Arquivo.SHA256 {
			add("ja-conferido", Info, "Você já conferiu este mesmo arquivo em "+v.Quando+".", "")
			return
		}
		// Só acusa com os dois valores lidos com certeza; na dúvida, informa.
		if iguais, ok := dinheiro.Iguais(v.Valor, e.Campos.Valor); ok && !iguais {
			add("id-repetido", Falso, "Este mesmo código de transação já apareceu em outro comprovante, com outro valor ("+v.Valor+"), conferido em "+v.Quando+".",
				"Cada Pix tem um código único.")
			return
		}
		add("mesmo-codigo", Info, "Um comprovante com este mesmo código já foi conferido em "+v.Quando+".", "Pode ser o mesmo Pix enviado em outro arquivo.")
		return
	}
}

func veredito(as []Achado, temCodigo bool) string {
	alerta, naoPago := false, false
	for _, a := range as {
		switch a.Peso {
		case Falso:
			return VereditoFalso
		case Alerta:
			alerta = true
		case NaoPago:
			naoPago = true
		}
	}
	switch {
	case naoPago:
		return VereditoNaoPago
	case alerta:
		return VereditoSuspeito
	case !temCodigo:
		return VereditoIncompleto
	}
	return VereditoSemErros
}

func valorZero(v string) bool {
	c, ok := dinheiro.Centavos(v)
	return ok && c == 0
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func duracao(d time.Duration) string {
	m := int(d.Round(time.Minute).Minutes())
	if m < 60 {
		return fmt.Sprintf("%d minutos", m)
	}
	return fmt.Sprintf("%dh%02d", m/60, m%60)
}
