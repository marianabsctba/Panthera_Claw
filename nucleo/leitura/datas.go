package leitura

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/bancos"
	"github.com/marianabsctba/Panthera_Claw/nucleo/pix"
)

// DataHora é uma data lida do comprovante, no horário de Brasília.
type DataHora struct {
	Instante time.Time
	TemHora  bool
}

var meses = map[string]time.Month{
	"jan": 1, "janeiro": 1, "fev": 2, "fevereiro": 2, "mar": 3, "marco": 3,
	"abr": 4, "abril": 4, "mai": 5, "maio": 5, "jun": 6, "junho": 6,
	"jul": 7, "julho": 7, "ago": 8, "agosto": 8, "set": 9, "setembro": 9,
	"out": 10, "outubro": 10, "nov": 11, "novembro": 11, "dez": 12, "dezembro": 12,
}

var (
	// 29/10/2026 14:32, 29/10/26 às 14:32:10, 29-10-2026 - 14h32
	reNumerica = regexp.MustCompile(`(\d{1,2})[/.-](\d{1,2})[/.-](\d{2,4})(?:\D{0,6}?(\d{1,2})[:h](\d{2})(?::(\d{2}))?)?`)
	// 29 out 2026 14:32, 29 de outubro de 2026 às 14:32 (texto já normalizado)
	reExtenso = regexp.MustCompile(`(\d{1,2}) (?:de )?([a-z]{3,9}) (?:de )?(\d{4})(?:\D{0,6}?(\d{1,2})[ h:](\d{2})(?: (\d{2}))?)?`)
)

// LerDataHora acha a primeira data (com hora, se houver) num trecho.
func LerDataHora(s string) (DataHora, bool) {
	if m := reNumerica.FindStringSubmatch(s); m != nil {
		if d, ok := montar(m[1], "", m[2], m[3], m[4], m[5], m[6]); ok {
			return d, true
		}
	}
	n := bancos.Normalizar(s)
	if m := reExtenso.FindStringSubmatch(n); m != nil {
		if d, ok := montar(m[1], m[2], "", m[3], m[4], m[5], m[6]); ok {
			return d, true
		}
	}
	return DataHora{}, false
}

func montar(dia, mesNome, mesNum, ano, h, min, seg string) (DataHora, bool) {
	d, _ := strconv.Atoi(dia)
	var m time.Month
	if mesNome != "" {
		var ok bool
		if m, ok = meses[strings.ToLower(mesNome)]; !ok {
			return DataHora{}, false
		}
	} else {
		n, _ := strconv.Atoi(mesNum)
		m = time.Month(n)
	}
	a, _ := strconv.Atoi(ano)
	if a < 100 {
		a += 2000
	}
	hh, mm, ss := 0, 0, 0
	tem := h != ""
	if tem {
		hh, _ = strconv.Atoi(h)
		mm, _ = strconv.Atoi(min)
		ss, _ = strconv.Atoi(seg)
	}
	if m < 1 || m > 12 || d < 1 || d > 31 || hh > 23 || mm > 59 || ss > 59 {
		return DataHora{}, false
	}
	t := time.Date(a, m, d, hh, mm, ss, 0, pix.Brasilia)
	if t.Day() != d {
		return DataHora{}, false // 31/02 e afins
	}
	return DataHora{Instante: t, TemHora: tem}, true
}
