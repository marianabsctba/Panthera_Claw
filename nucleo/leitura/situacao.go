package leitura

import (
	"strings"

	"github.com/marianabsctba/Panthera_Claw/nucleo/bancos"
)

// Situações de um Pix que ainda não é (ou deixou de ser) pagamento.
const (
	SituacaoAgendado  = "agendado"
	SituacaoPendente  = "pendente"
	SituacaoCancelado = "cancelado"
	SituacaoDevolvido = "devolvido"
)

// Golpe comum: agendar o Pix, mandar o comprovante do agendamento (parece
// de pagamento) e cancelar depois. As expressões estão normalizadas e são
// procuradas como palavras inteiras, na ordem de gravidade.
var situacoes = []struct {
	situacao string
	termos   []string
}{
	{SituacaoCancelado, []string{"cancelado", "cancelada", "agendamento cancelado", "estornado", "estornada", "falhou", "nao realizado", "nao realizada", "recusado", "recusada", "rejeitado", "rejeitada", "nao efetuado", "nao efetuada"}},
	{SituacaoDevolvido, []string{"pix devolvido", "comprovante de devolucao", "devolucao de pix", "devolucao do pix", "valor devolvido"}},
	{SituacaoPendente, []string{"pendente", "em processamento", "processando", "em analise", "aguardando", "aguardando confirmacao", "nao concluido", "nao concluida", "em andamento"}},
	{SituacaoAgendado, []string{"agendado", "agendada", "agendamento", "pix agendado", "programado", "programada", "pix programado", "agendado para", "data do agendamento", "data agendada"}},
}

// LerSituacao procura no texto sinais de que o Pix não foi concluído.
// Devolve a situação e a linha onde ela aparece.
func LerSituacao(texto string) (string, string) {
	linhas := quebrar(texto)
	for _, s := range situacoes {
		for _, l := range linhas {
			n := " " + bancos.Normalizar(l) + " "
			for _, t := range s.termos {
				if strings.Contains(n, " "+t+" ") {
					return s.situacao, l
				}
			}
		}
	}
	return "", ""
}
