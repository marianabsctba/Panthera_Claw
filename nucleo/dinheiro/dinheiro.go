// Package dinheiro lê e escreve valores em reais. Toda comparação de valor
// passa por centavos: comparar só os algarismos confundia R$ 1,50 com
// R$ 150 e R$ 15,0 com R$ 150,00.
package dinheiro

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Com vírgula: a vírgula separa os centavos (1 ou 2 casas) e os pontos
	// ou espaços separam milhares. Ex.: 1.250,00  1250,5  1 250,00
	reVirgula = regexp.MustCompile(`^(\d{1,3}(?:[.\s]\d{3})+|\d+),(\d{1,2})$`)
	// Sem vírgula, com ponto de milhar: 1.250  1.250.000
	reMilhar = regexp.MustCompile(`^\d{1,3}(?:\.\d{3})+$`)
	// Sem vírgula, ponto como decimal (1 ou 2 casas): 15.5  15.50
	rePonto   = regexp.MustCompile(`^(\d+)\.(\d{1,2})$`)
	reInteiro = regexp.MustCompile(`^\d+$`)
)

// Centavos lê um valor escrito à brasileira, com ou sem "R$".
func Centavos(s string) (int64, bool) {
	t := limpar(s)
	switch {
	case t == "":
		return 0, false
	case reVirgula.MatchString(t):
		m := reVirgula.FindStringSubmatch(t)
		return juntar(m[1], m[2])
	case reMilhar.MatchString(t):
		return juntar(t, "")
	case rePonto.MatchString(t):
		m := rePonto.FindStringSubmatch(t)
		return juntar(m[1], m[2])
	case reInteiro.MatchString(t):
		return juntar(t, "")
	}
	return 0, false
}

// CentavosPonto lê o valor do Pix copia e cola, onde o ponto é sempre o
// separador decimal ("10.50", "1.5", "1000").
func CentavosPonto(s string) (int64, bool) {
	t := strings.TrimSpace(s)
	inteiro, frac, _ := strings.Cut(t, ".")
	if !reInteiro.MatchString(inteiro) || (frac != "" && !regexp.MustCompile(`^\d{1,2}$`).MatchString(frac)) {
		return 0, false
	}
	return juntar(inteiro, frac)
}

// Formatar escreve centavos como "R$ 1.250,00".
func Formatar(c int64) string {
	neg := c < 0
	if neg {
		c = -c
	}
	s := strconv.FormatInt(c/100, 10)
	var b strings.Builder
	for k, r := range s {
		if k > 0 && (len(s)-k)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	out := "R$ " + b.String() + "," + strconv.FormatInt(c%100/10, 10) + strconv.FormatInt(c%10, 10)
	if neg {
		out = "-" + out
	}
	return out
}

// Iguais diz se dois textos representam o mesmo valor. Se algum não puder
// ser lido, não afirma nada (devolve ok falso).
func Iguais(a, b string) (iguais, ok bool) {
	ca, oka := Centavos(a)
	cb, okb := Centavos(b)
	if !oka || !okb {
		return false, false
	}
	return ca == cb, true
}

// Parece diz se o texto é um valor em reais: tem "R$" ou vírgula decimal.
func Parece(s string) bool {
	if _, ok := Centavos(s); !ok {
		return false
	}
	return strings.Contains(strings.ToUpper(s), "R$") || strings.Contains(s, ",")
}

func limpar(s string) string {
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(strings.TrimPrefix(t, "R$"), "r$")
	t = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t), "-"))
	t = strings.ReplaceAll(t, " ", " ")
	return strings.TrimSpace(t)
}

func juntar(inteiro, frac string) (int64, bool) {
	inteiro = strings.NewReplacer(".", "", " ", "").Replace(inteiro)
	for len(frac) < 2 {
		frac += "0"
	}
	r, err := strconv.ParseInt(inteiro, 10, 64)
	if err != nil || r > 1e12 {
		return 0, false
	}
	c, _ := strconv.ParseInt(frac, 10, 64)
	return r*100 + c, true
}
