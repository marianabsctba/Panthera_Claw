// Package bancos conhece as instituições do Pix pela lista pública do Banco
// Central (gerada por ferramentas/bancos) e reconhece o nome que aparece
// impresso num comprovante, que quase nunca é o nome oficial.
package bancos

import (
	_ "embed"
	"encoding/json"
	"strings"
	"unicode"
)

//go:embed bancos.json
var bruto []byte

// Banco é uma instituição da lista do Banco Central.
type Banco struct {
	ISPB     string `json:"ispb"`
	Nome     string `json:"nome"`
	CNPJ     string `json:"cnpj,omitempty"`
	Tipo     string `json:"tipo,omitempty"`
	Compe    string `json:"compe,omitempty"`
	Situacao string `json:"situacao"` // ativo, adesao ou str (só emite boleto)
}

// Grupo junta as instituições de uma mesma marca: o comprovante diz
// "PicPay", mas a marca tem mais de um ISPB.
type Grupo struct {
	Marca string          `json:"marca"`
	ISPBs map[string]bool `json:"-"`
}

var (
	porISPB = map[string]Banco{}
	grupos  []*Grupo
)

func init() {
	var lista []Banco
	if err := json.Unmarshal(bruto, &lista); err != nil {
		panic(err)
	}
	for _, b := range lista {
		porISPB[b.ISPB] = b
	}
	for _, m := range marcas {
		g := &Grupo{Marca: m.marca, ISPBs: map[string]bool{}}
		for _, b := range lista {
			if contem(Normalizar(b.Nome), m.oficiais) {
				g.ISPBs[b.ISPB] = true
			}
		}
		grupos = append(grupos, g)
	}
}

// PorISPB devolve a instituição dona do código de 8 dígitos.
func PorISPB(ispb string) (Banco, bool) {
	b, ok := porISPB[ispb]
	return b, ok
}

// Marca devolve o grupo de marca ao qual o ISPB pertence, se houver.
func Marca(ispb string) *Grupo {
	for _, g := range grupos {
		if g.ISPBs[ispb] {
			return g
		}
	}
	return nil
}

// Reconhecer encontra a marca citada num trecho de texto do comprovante.
// Devolve nil quando nenhuma marca conhecida aparece.
func Reconhecer(texto string) *Grupo {
	n := Normalizar(texto)
	if n == "" {
		return nil
	}
	var melhor *Grupo
	maior := 0
	for i, m := range marcas {
		for _, t := range m.termos {
			if strings.Contains(" "+n+" ", " "+t+" ") && len(t) > maior {
				melhor, maior = grupos[i], len(t)
			}
		}
	}
	if melhor != nil {
		return melhor
	}
	// Sem marca conhecida: tenta o nome oficial inteiro da lista.
	for _, b := range porISPB {
		if b.Situacao == "str" {
			continue
		}
		nb := Normalizar(b.Nome)
		if len(nb) >= 6 && strings.Contains(" "+n+" ", " "+nb+" ") {
			return &Grupo{Marca: b.Nome, ISPBs: map[string]bool{b.ISPB: true}}
		}
	}
	return nil
}

// Normalizar deixa o texto em minúsculas, sem acento e só com letras,
// números e espaços simples.
func Normalizar(s string) string {
	var b strings.Builder
	espaco := true
	for _, r := range s {
		if base, ok := semAcento[unicode.ToLower(r)]; ok {
			r = base
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			espaco = false
		default:
			if !espaco {
				b.WriteByte(' ')
				espaco = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func contem(n string, termos []string) bool {
	for _, t := range termos {
		if strings.Contains(" "+n+" ", " "+t+" ") {
			return true
		}
	}
	return false
}

var semAcento = map[rune]rune{
	'á': 'a', 'à': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a',
	'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
	'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
	'ó': 'o', 'ò': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
	'ç': 'c', 'ñ': 'n',
}

// PorCompe devolve a instituição pelo número de banco (o que aparece nos
// três primeiros dígitos do boleto).
func PorCompe(compe string) (Banco, bool) {
	for _, b := range porISPB {
		if b.Compe == compe {
			return b, true
		}
	}
	return Banco{}, false
}
