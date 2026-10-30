// Package boleto lê a linha digitável (ou o código de barras) de boletos
// bancários e de arrecadação (contas de consumo, tributos) e confere os
// dígitos verificadores, segundo as regras da Febraban.
package boleto

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/dinheiro"
)

// Tipos de boleto.
const (
	Bancario    = "bancario"    // 47 dígitos, banco nos 3 primeiros
	Arrecadacao = "arrecadacao" // 48 dígitos, começa com 8
)

// Boleto é o que a linha revela.
type Boleto struct {
	Tipo   string `json:"tipo"`
	Linha  string `json:"linha"`
	Barras string `json:"barras"`
	// Bancário
	Banco      string `json:"banco,omitempty"` // número do banco (COMPE)
	Moeda      string `json:"moeda,omitempty"`
	Vencimento string `json:"vencimento,omitempty"` // dd/mm/aaaa; vazio se não tiver
	// Arrecadação
	Segmento string `json:"segmento,omitempty"`
	Empresa  string `json:"empresa,omitempty"` // identificação do órgão ou empresa
	// Valor em centavos; 0 quando o boleto não traz valor.
	Centavos int64  `json:"centavos"`
	Valor    string `json:"valor,omitempty"`
	// Erros de dígito verificador, em linguagem simples.
	Erros []string `json:"erros,omitempty"`
}

var segmentos = map[byte]string{
	'1': "prefeitura", '2': "saneamento", '3': "energia elétrica e gás", '4': "telecomunicações",
	'5': "órgão do governo", '6': "carnê ou outra empresa", '7': "multa de trânsito", '9': "uso exclusivo do banco",
}

// Ler interpreta a linha digitável (47 ou 48 dígitos) ou o código de
// barras (44 dígitos). Pontos, espaços e traços são ignorados.
func Ler(s string, hoje time.Time) (Boleto, error) {
	d := soDigitos(s)
	switch {
	case len(d) == 47:
		return lerBancario(d, hoje), nil
	case len(d) == 48 && d[0] == '8':
		return lerArrecadacao(d), nil
	case len(d) == 44 && d[0] == '8':
		return lerArrecadacaoBarras(d), nil
	case len(d) == 44:
		return lerBancarioBarras(d, hoje), nil
	}
	return Boleto{}, fmt.Errorf("o código tem %d números; a linha do boleto tem 47 (bancário) ou 48 (conta de consumo e tributo)", len(d))
}

func lerBancario(d string, hoje time.Time) Boleto {
	b := Boleto{Tipo: Bancario, Linha: d}
	c1, c2, c3, c4, c5 := d[0:10], d[10:21], d[21:32], d[32:33], d[33:47]
	for k, c := range []string{c1, c2, c3} {
		if modulo10(c[:len(c)-1]) != c[len(c)-1] {
			b.Erros = append(b.Erros, fmt.Sprintf("o dígito de conferência do bloco %d da linha está errado", k+1))
		}
	}
	b.Barras = c1[0:4] + c4 + c5 + c1[4:9] + c2[0:10] + c3[0:10]
	completarBancario(&b, hoje)
	return b
}

func lerBancarioBarras(d string, hoje time.Time) Boleto {
	b := Boleto{Tipo: Bancario, Barras: d}
	campoLivre := d[19:44]
	c1 := d[0:4] + campoLivre[0:5]
	c2 := campoLivre[5:15]
	c3 := campoLivre[15:25]
	b.Linha = c1 + string(modulo10(c1)) + c2 + string(modulo10(c2)) + c3 + string(modulo10(c3)) + d[4:5] + d[5:19]
	completarBancario(&b, hoje)
	return b
}

func completarBancario(b *Boleto, hoje time.Time) {
	br := b.Barras
	b.Banco, b.Moeda = br[0:3], br[3:4]
	if dvGeralBancario(br[:4]+br[5:]) != br[4] {
		b.Erros = append(b.Erros, "o dígito verificador geral do boleto está errado")
	}
	fator, _ := strconv.Atoi(br[5:9])
	b.Vencimento = vencimento(fator, hoje)
	b.Centavos, _ = strconv.ParseInt(br[9:19], 10, 64)
	b.Valor = reais(b.Centavos)
}

// vencimento converte o fator (dias a partir de 07/10/1997). O fator chegou
// a 9999 em 21/02/2025 e recomeçou em 1000 em 22/02/2025; vale o ciclo mais
// perto de hoje.
func vencimento(fator int, hoje time.Time) string {
	if fator == 0 {
		return ""
	}
	antigo := time.Date(1997, 10, 7, 0, 0, 0, 0, time.UTC).AddDate(0, 0, fator)
	novo := time.Date(2025, 2, 22, 0, 0, 0, 0, time.UTC).AddDate(0, 0, fator-1000)
	v := antigo
	if fator >= 1000 && abs(novo.Sub(hoje)) < abs(antigo.Sub(hoje)) {
		v = novo
	}
	return v.Format("02/01/2006")
}

func lerArrecadacao(d string) Boleto {
	b := Boleto{Tipo: Arrecadacao, Linha: d}
	mod := modDoTipo(d[2])
	var barras strings.Builder
	for k := 0; k < 4; k++ {
		bloco := d[k*12 : k*12+12]
		if mod(bloco[:11]) != bloco[11] {
			b.Erros = append(b.Erros, fmt.Sprintf("o dígito de conferência do bloco %d da linha está errado", k+1))
		}
		barras.WriteString(bloco[:11])
	}
	b.Barras = barras.String()
	completarArrecadacao(&b)
	return b
}

func lerArrecadacaoBarras(d string) Boleto {
	b := Boleto{Tipo: Arrecadacao, Barras: d}
	mod := modDoTipo(d[2])
	var linha strings.Builder
	for k := 0; k < 4; k++ {
		bloco := d[k*11 : k*11+11]
		linha.WriteString(bloco)
		linha.WriteByte(mod(bloco))
	}
	b.Linha = linha.String()
	completarArrecadacao(&b)
	return b
}

func completarArrecadacao(b *Boleto) {
	br := b.Barras
	if modDoTipo(br[2])(br[:3]+br[4:]) != br[3] {
		b.Erros = append(b.Erros, "o dígito verificador geral do código está errado")
	}
	b.Segmento = segmentos[br[1]]
	b.Empresa = br[15:19]
	// Com o terceiro dígito 6 ou 8, o valor é em reais; 7 ou 9, é referência.
	if br[2] == '6' || br[2] == '8' {
		b.Centavos, _ = strconv.ParseInt(br[4:15], 10, 64)
		b.Valor = reais(b.Centavos)
	}
}

func modDoTipo(c byte) func(string) byte {
	if c == '8' || c == '9' {
		return modulo11Arrecadacao
	}
	return modulo10
}

// modulo10: pesos 2 e 1 alternados da direita para a esquerda, somando os
// algarismos dos produtos.
func modulo10(s string) byte {
	soma, peso := 0, 2
	for k := len(s) - 1; k >= 0; k-- {
		p := int(s[k]-'0') * peso
		soma += p/10 + p%10
		peso = 3 - peso
	}
	return byte('0' + (10-soma%10)%10)
}

// dvGeralBancario: módulo 11 com pesos 2 a 9; resto 0, 1 ou 10 vira 1.
func dvGeralBancario(s string) byte {
	soma, peso := 0, 2
	for k := len(s) - 1; k >= 0; k-- {
		soma += int(s[k]-'0') * peso
		peso++
		if peso > 9 {
			peso = 2
		}
	}
	dv := 11 - soma%11
	if dv == 0 || dv == 10 || dv == 11 {
		dv = 1
	}
	return byte('0' + dv)
}

// modulo11Arrecadacao: pesos 2 a 9; resto 0 ou 1 vira 0, resto 10 vira 1.
func modulo11Arrecadacao(s string) byte {
	soma, peso := 0, 2
	for k := len(s) - 1; k >= 0; k-- {
		soma += int(s[k]-'0') * peso
		peso++
		if peso > 9 {
			peso = 2
		}
	}
	switch r := soma % 11; r {
	case 0, 1:
		return '0'
	case 10:
		return '1'
	default:
		return byte('0' + 11 - r)
	}
}

func reais(c int64) string {
	if c == 0 {
		return ""
	}
	return dinheiro.Formatar(c)
}

func soDigitos(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
