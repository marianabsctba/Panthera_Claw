package arquivo

import (
	"bytes"
	"compress/zlib"
	"encoding/ascii85"
	"io"
	"regexp"
	"strings"
)

// Alteracao descreve o que um PDF salvo de novo (atualização incremental)
// mudou depois da primeira versão.
type Alteracao struct {
	Objetos int `json:"objetos"`
	// Pagina indica que a alteração trouxe conteúdo de página novo: algo
	// que aparece na tela foi escrito ou desenhado depois.
	Pagina bool `json:"pagina"`
	// Textos legíveis escritos na alteração (fontes sem subconjunto).
	Textos []string `json:"textos,omitempty"`
}

var (
	reObjeto   = regexp.MustCompile(`(?s)(\d+)\s+(\d+)\s+obj\b(.*?)\bendobj`)
	reStream   = regexp.MustCompile(`(?s)stream\r?\n(.*?)\r?\n?endstream`)
	reTj       = regexp.MustCompile(`\(((?:\\.|[^\\)])*)\)\s*(?:Tj|'|")`)
	reTJ       = regexp.MustCompile(`(?s)\[(.*?)\]\s*TJ`)
	reLiteral  = regexp.MustCompile(`\(((?:\\.|[^\\)])*)\)`)
	reOperador = regexp.MustCompile(`\b(BT|Tj|TJ|re\s+f|Do)\b`)
)

// alteracoes olha cada versão gravada depois da primeira.
func alteracoes(d []byte, linearizado bool) *Alteracao {
	fins := todos(d, []byte("%%EOF"))
	pular := 1
	if linearizado {
		pular = 2
	}
	if len(fins) <= pular {
		return nil
	}
	a := &Alteracao{}
	inicio := fins[pular-1]
	// Textos que já existiam na primeira versão não contam como novos:
	// alguns programas regravam a página inteira ao salvar por cima.
	antes := map[string]bool{}
	for _, c := range conteudos(d[:inicio], nil) {
		for _, t := range textosDoConteudo(c) {
			antes[t] = true
		}
	}
	for _, c := range conteudos(d[inicio:], &a.Objetos) {
		a.Pagina = true
		for _, t := range textosDoConteudo(c) {
			if !antes[t] {
				a.Textos = append(a.Textos, t)
			}
		}
	}
	if a.Objetos == 0 {
		return nil
	}
	return a
}

// conteudos devolve os fluxos de conteúdo de página (já descomprimidos)
// definidos no trecho, e conta os objetos.
func conteudos(trecho []byte, objetos *int) [][]byte {
	var out [][]byte
	for _, m := range reObjeto.FindAllSubmatch(trecho, -1) {
		if objetos != nil {
			*objetos++
		}
		corpo := m[3]
		s := reStream.FindSubmatch(corpo)
		if s == nil {
			continue
		}
		conteudo := s[1]
		dic := corpo[:bytes.Index(corpo, []byte("stream"))]
		if bytes.Contains(dic, []byte("/ASCII85Decode")) || bytes.Contains(dic, []byte("/A85")) {
			conteudo = de85(conteudo)
		}
		if bytes.Contains(dic, []byte("/FlateDecode")) || bytes.Contains(dic, []byte("/Fl ")) {
			if r, err := zlib.NewReader(bytes.NewReader(conteudo)); err == nil {
				if b, err := io.ReadAll(io.LimitReader(r, 4<<20)); err == nil || len(b) > 0 {
					conteudo = b
				}
			}
		}
		if reOperador.Match(conteudo) {
			out = append(out, conteudo)
		}
	}
	return out
}

// textosDoConteudo tira as strings legíveis dos operadores de texto.
func textosDoConteudo(c []byte) []string {
	var out []string
	for _, m := range reTj.FindAllSubmatch(c, -1) {
		if t := strings.TrimSpace(textoPDF(m[1])); t != "" {
			out = append(out, t)
		}
	}
	for _, m := range reTJ.FindAllSubmatch(c, -1) {
		var b strings.Builder
		for _, l := range reLiteral.FindAllSubmatch(m[1], -1) {
			b.WriteString(textoPDF(l[1]))
		}
		if t := strings.TrimSpace(b.String()); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// de85 decodifica o filtro ASCII85 dos fluxos de PDF (termina em "~>").
func de85(b []byte) []byte {
	b = bytes.TrimSpace(b)
	if i := bytes.Index(b, []byte("~>")); i >= 0 {
		b = b[:i]
	}
	b = bytes.TrimPrefix(b, []byte("<~"))
	out := make([]byte, 4*len(b)/5+4)
	n, _, err := ascii85.Decode(out, b, true)
	if err != nil && n == 0 {
		return b
	}
	return out[:n]
}

func todos(d, alvo []byte) []int {
	var pos []int
	for k := 0; ; {
		i := bytes.Index(d[k:], alvo)
		if i < 0 {
			return pos
		}
		pos = append(pos, k+i+len(alvo))
		k += i + len(alvo)
	}
}
