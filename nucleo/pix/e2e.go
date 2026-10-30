// Package pix lê os identificadores e as chaves do Pix segundo as regras
// públicas do Banco Central, sem consultar banco nenhum.
package pix

import (
	"fmt"
	"strings"
	"time"
)

// E2E é o identificador fim a fim de uma transação Pix (EndToEndId):
//
//	E 18236120 202610291732 abc12XYZ789
//	│ │        │            └ sequencial (11 caracteres)
//	│ │        └ data e hora em UTC (AAAAMMDDHHMM)
//	│ └ ISPB do participante que gerou o identificador
//	└ E para pagamento, D para devolução
type E2E struct {
	Texto      string    `json:"texto"`
	Tipo       string    `json:"tipo"` // "E" ou "D"
	ISPB       string    `json:"ispb"`
	Instante   time.Time `json:"instante"` // em UTC
	Sequencial string    `json:"sequencial"`
	// Correcoes registra o que foi ajustado por confusão de leitura (letra
	// O lida no lugar do zero, por exemplo). Só a parte numérica é ajustada.
	Correcoes []string `json:"correcoes,omitempty"`
}

// Brasilia é o horário oficial de Brasília. Sem horário de verão desde
// 2019, antes do lançamento do Pix, então o deslocamento fixo é exato.
var Brasilia = time.FixedZone("Brasília", -3*3600)

// Lancamento é o início da operação do Pix: a fase restrita, com parte dos
// clientes, começou em 03/11/2020; o lançamento geral foi em 16/11/2020.
var Lancamento = time.Date(2020, 11, 3, 0, 0, 0, 0, Brasilia)

// ErroE2E explica, em linguagem simples, por que o texto não é um
// identificador possível.
type ErroE2E struct {
	Motivo string
}

func (e *ErroE2E) Error() string { return e.Motivo }

// confusoes troca, só na parte que deve ser número, as letras que a
// leitura de imagem costuma confundir com dígitos.
var confusoes = map[rune]rune{
	'O': '0', 'o': '0', 'Q': '0', 'D': '0',
	'I': '1', 'l': '1', '|': '1', 'i': '1', '!': '1',
	'Z': '2', 'z': '2',
	'S': '5', 's': '5',
	'G': '6', 'b': '6',
	'B': '8',
}

// LerE2E interpreta um identificador. Espaços e quebras de linha no meio
// são ignorados, porque o código costuma aparecer quebrado em duas linhas.
func LerE2E(s string) (E2E, error) {
	var limpo []rune
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == '-' {
			continue
		}
		limpo = append(limpo, r)
	}
	e := E2E{}
	if len(limpo) != 32 {
		return e, &ErroE2E{fmt.Sprintf("o código tem %d caracteres, mas o do Pix tem sempre 32", len(limpo))}
	}
	switch limpo[0] {
	case 'E', 'D':
	case 'e', 'd':
		e.Correcoes = append(e.Correcoes, fmt.Sprintf("%c lido como %c", limpo[0], limpo[0]-32))
		limpo[0] -= 32
	default:
		return e, &ErroE2E{fmt.Sprintf("o código começa com %q, mas o do Pix começa com E (ou D, na devolução)", string(limpo[0]))}
	}
	for i := 1; i <= 20; i++ {
		r := limpo[i]
		if r >= '0' && r <= '9' {
			continue
		}
		if d, ok := confusoes[r]; ok {
			e.Correcoes = append(e.Correcoes, fmt.Sprintf("posição %d: %c lido como %c", i+1, r, d))
			limpo[i] = d
			continue
		}
		return e, &ErroE2E{fmt.Sprintf("a posição %d do código tem %q, mas ali só pode haver número", i+1, string(r))}
	}
	for i := 21; i < 32; i++ {
		r := limpo[i]
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return e, &ErroE2E{fmt.Sprintf("a posição %d do código tem %q, que não é letra nem número", i+1, string(r))}
		}
	}
	t := string(limpo)
	e.Texto, e.Tipo, e.ISPB, e.Sequencial = t, t[:1], t[1:9], t[21:]
	inst, err := time.ParseInLocation("200601021504", t[9:21], time.UTC)
	if err != nil {
		return e, &ErroE2E{fmt.Sprintf("a data dentro do código (%s) não existe no calendário", formatarBruto(t[9:21]))}
	}
	e.Instante = inst
	return e, nil
}

func formatarBruto(d string) string {
	return d[6:8] + "/" + d[4:6] + "/" + d[0:4] + " " + d[8:10] + ":" + d[10:12]
}

// NoHorarioDeBrasilia devolve o instante do identificador em Brasília.
func (e E2E) NoHorarioDeBrasilia() time.Time { return e.Instante.In(Brasilia) }

// Parece diz se um trecho de texto tem a forma de um E2E, para achá-lo no
// meio do comprovante.
func Parece(s string) bool {
	_, err := LerE2E(s)
	return err == nil
}

// Normalizar tira espaços, para comparar identificadores.
func Normalizar(s string) string {
	return strings.Join(strings.Fields(s), "")
}
