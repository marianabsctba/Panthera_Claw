// Package leitura tira do texto de um comprovante (vindo do PDF ou da
// leitura da imagem) os campos que serão conferidos. Os bancos não seguem
// um modelo único, então a leitura procura rótulos conhecidos e seções de
// "origem" e "destino"; o que não for achado fica vazio para a pessoa
// completar.
package leitura

import (
	"regexp"
	"strings"

	"github.com/marianabsctba/Panthera_Claw/nucleo/bancos"
	"github.com/marianabsctba/Panthera_Claw/nucleo/dinheiro"
	"github.com/marianabsctba/Panthera_Claw/nucleo/pix"
)

// Campos são os dados do comprovante, como aparecem impressos.
type Campos struct {
	ID             string `json:"id"`
	Valor          string `json:"valor"`
	DataHora       string `json:"dataHora"`
	BancoPagador   string `json:"bancoPagador"`
	BancoRecebedor string `json:"bancoRecebedor"`
	NomePagador    string `json:"nomePagador"`
	NomeRecebedor  string `json:"nomeRecebedor"`
	Chave          string `json:"chave"`
	// Emissor é a marca que aparece no topo do comprovante (o aplicativo
	// que gerou), usada quando a seção de quem pagou não diz o banco.
	Emissor string `json:"emissor"`
	// IDDuvidoso indica que o final do código, na imagem, tinha mais ou
	// menos caracteres do que deveria: a pessoa precisa conferir.
	IDDuvidoso bool `json:"idDuvidoso,omitempty"`
	// Situacao indica Pix agendado, pendente, cancelado ou devolvido, e
	// SituacaoTrecho a linha do comprovante que diz isso.
	Situacao       string `json:"situacao,omitempty"`
	SituacaoTrecho string `json:"situacaoTrecho,omitempty"`
}

// A leitura de imagem costuma trocar o @ por "(Q", "(@" ou "©".
var reArrobaLida = regexp.MustCompile(`^([A-Za-z0-9._%+-]+)\s*(?:\(Q\)?|\(@\)?|\(a\)|©)\s*([A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,})$`)

var reArrobaSuja = regexp.MustCompile(`\s*\(?@\)?\s*`)

func corrigirEmail(s string) string {
	if strings.Contains(s, "@") {
		return reArrobaSuja.ReplaceAllString(s, "@")
	}
	if m := reArrobaLida.FindStringSubmatch(s); m != nil {
		return m[1] + "@" + m[2]
	}
	// Em letra grossa, o @ vira um Q maiúsculo solto no meio de um e-mail
	// todo em minúsculas: "mariaQexemplo.com.br".
	if m := reArrobaQ.FindStringSubmatch(s); m != nil && strings.ToLower(m[1]+m[2]) == m[1]+m[2] {
		return m[1] + "@" + m[2]
	}
	return s
}

var reArrobaQ = regexp.MustCompile(`^([a-z0-9._%+-]+)\s*Q\s*([a-z0-9-]+(?:\.[a-z0-9-]+)*\.[a-z]{2,})$`)

var (
	reID = regexp.MustCompile(`[EeDd][0-9OoQDIil|!SsZzBGb]{20}[0-9A-Za-z]{11}`)
	// "R$" (que a leitura de imagem às vezes lê como RS ou R5) seguido do
	// valor; centavos com uma ou duas casas, ou sem centavos.
	reValor = regexp.MustCompile(`R\s?[$S5]\s*-?\s*(\d{1,3}(?:\.\d{3})+(?:,\d{1,2})?|\d+(?:,\d{1,2})?)`)
)

type secao int

const (
	nenhuma secao = iota
	origem
	destino
)

var (
	rotulosOrigem  = []string{"origem", "pagador", "quem pagou", "de", "dados de quem pagou", "dados do pagador", "conta de origem", "remetente", "debitado de", "pago por"}
	rotulosDestino = []string{"destino", "recebedor", "quem recebeu", "para", "dados de quem recebeu", "dados do recebedor", "conta de destino", "favorecido", "beneficiario", "dados do favorecido", "destinatario", "enviado para"}
	rotulosBanco   = []string{"instituicao", "instituicao financeira", "banco", "instituicao de pagamento"}
	rotulosNome    = []string{"nome"}
	rotulosChave   = []string{"chave", "chave pix"}
	rotulosValor   = []string{"valor", "valor pago", "valor da transferencia", "valor do pix", "total"}
	rotulosData    = []string{"data", "data e hora", "data da transferencia", "data do pagamento", "realizado em", "efetuado em", "quando", "data da transacao", "horario"}
)

// Ler extrai os campos de um texto com uma informação por linha.
func Ler(texto string) Campos {
	linhas := quebrar(texto)
	c := Campos{}

	c.ID = acharID(texto)
	c.IDDuvidoso = duvidoso(texto, c.ID)
	c.Situacao, c.SituacaoTrecho = LerSituacao(texto)

	sec := nenhuma
	for i, l := range linhas {
		n := bancos.Normalizar(l)
		if r, ok := ehRotulo(n, rotulosOrigem); ok && r == n {
			sec = origem
			continue
		}
		if r, ok := ehRotulo(n, rotulosDestino); ok && r == n {
			sec = destino
			continue
		}
		valor := func(rotulos []string) string { return valorDoRotulo(linhas, i, l, n, rotulos) }
		if v := valor(rotulosBanco); v != "" {
			atribuir(sec, &c.BancoPagador, &c.BancoRecebedor, v)
		}
		if v := valor(rotulosNome); v != "" {
			atribuir(sec, &c.NomePagador, &c.NomeRecebedor, v)
		}
		if v := valor(rotulosChave); v != "" && c.Chave == "" {
			c.Chave = corrigirEmail(v)
		}
		if v := valor(rotulosValor); v != "" && c.Valor == "" {
			c.Valor = acharValor(v)
		}
		if v := valor(rotulosData); v != "" && c.DataHora == "" {
			if _, ok := LerDataHora(v); ok {
				c.DataHora = v
			}
		}
	}

	if c.Valor == "" {
		c.Valor = acharValor(texto)
	}
	if c.DataHora == "" {
		for _, l := range linhas {
			if d, ok := LerDataHora(l); ok && d.TemHora {
				c.DataHora = l
				break
			}
		}
	}
	// O emissor é a primeira marca reconhecida no alto do comprovante.
	for i, l := range linhas {
		if i >= 6 {
			break
		}
		if g := bancos.ReconhecerLogo(l); g != nil {
			c.Emissor = g.Marca
			break
		}
	}
	return c
}

// acharValor devolve o primeiro valor em reais do texto, já no formato
// "R$ 1.250,00", ou "".
func acharValor(texto string) string {
	m := reValor.FindStringSubmatch(texto)
	if m == nil {
		return ""
	}
	if c, ok := dinheiro.Centavos(m[1]); ok {
		return dinheiro.Formatar(c)
	}
	return "R$ " + m[1]
}

func quebrar(texto string) []string {
	var out []string
	for _, l := range strings.Split(strings.ReplaceAll(texto, "\r", ""), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// acharID procura o identificador no texto inteiro, sem espaços (ele costuma
// vir quebrado em duas linhas). Entre vários candidatos, prefere o que tem
// data válida.
func acharID(texto string) string {
	junto := strings.Join(strings.Fields(texto), "")
	cands := reID.FindAllString(junto, -1)
	for _, c := range cands {
		if pix.Parece(c) {
			return c
		}
	}
	if len(cands) > 0 {
		return cands[0]
	}
	return ""
}

// duvidoso confere, no texto original, se o código estava inteiro numa
// palavra de 32 caracteres ou quebrado em 21 + 11. Outro tamanho indica
// letra a mais ou a menos na leitura.
func duvidoso(texto, id string) bool {
	if len(id) != 32 {
		return false
	}
	palavras := strings.Fields(texto)
	for k, p := range palavras {
		if len(p) < 21 || !strings.EqualFold(p[:21], id[:21]) {
			continue
		}
		if len(p) == 21 {
			return k+1 >= len(palavras) || len(palavras[k+1]) != 11
		}
		return len(p) != 32
	}
	return false
}

// ehRotulo diz se a linha normalizada começa com um dos rótulos, e qual.
// Prefere o rótulo mais longo.
func ehRotulo(n string, rotulos []string) (string, bool) {
	melhor := ""
	for _, r := range rotulos {
		if (n == r || strings.HasPrefix(n, r+" ")) && len(r) > len(melhor) {
			melhor = r
		}
	}
	return melhor, melhor != ""
}

// valorDoRotulo devolve o valor de um rótulo: o resto da mesma linha,
// separado por dois pontos ou por coluna ("Instituição: Nubank"), ou a linha
// seguinte ("Instituição" / "Nubank"). Sem separador, "Banco do Brasil" não
// vira o rótulo "banco" com valor "do Brasil".
func valorDoRotulo(linhas []string, i int, original, n string, rotulos []string) string {
	if m := reSeparador.FindStringSubmatchIndex(original); m != nil {
		esq, dir := original[:m[0]], strings.TrimSpace(original[m[1]:])
		if r, ok := ehRotulo(bancos.Normalizar(esq), rotulos); ok && r == bancos.Normalizar(esq) {
			return dir
		}
	}
	r, ok := ehRotulo(n, rotulos)
	if !ok {
		return ""
	}
	if r != n {
		// Coluna que a leitura de imagem devolve como uma linha só, com um
		// espaço simples: "Instituição ITAÚ UNIBANCO S.A.". Não vale para
		// rótulos que também começam nomes ("Banco do Brasil").
		if ambiguos[r] {
			return ""
		}
		palavras := strings.Fields(original)
		if k := len(strings.Fields(r)); len(palavras) > k {
			return strings.Join(palavras[k:], " ")
		}
		return ""
	}
	if i+1 < len(linhas) {
		prox := linhas[i+1]
		if _, ehOutro := rotuloQualquer(bancos.Normalizar(prox)); !ehOutro {
			return prox
		}
	}
	return ""
}

var ambiguos = map[string]bool{"banco": true, "de": true, "para": true, "total": true, "quando": true}

var reSeparador = regexp.MustCompile(`\s*:\s*|\t+|\s{2,}`)

func rotuloQualquer(n string) (string, bool) {
	for _, g := range [][]string{rotulosOrigem, rotulosDestino, rotulosBanco, rotulosNome, rotulosChave, rotulosValor, rotulosData} {
		if r, ok := ehRotulo(n, g); ok && r == n {
			return r, true
		}
	}
	return "", false
}

func atribuir(sec secao, pagador, recebedor *string, v string) {
	switch sec {
	case origem:
		if *pagador == "" {
			*pagador = v
		}
	case destino:
		if *recebedor == "" {
			*recebedor = v
		}
	}
}
