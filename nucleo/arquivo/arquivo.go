// Package arquivo olha o arquivo do comprovante em si, sem ler o conteúdo:
// que programa gerou, se foi salvo de novo depois de criado, e o resumo
// criptográfico que identifica aquele arquivo exato.
package arquivo

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Info é o que foi encontrado no arquivo.
type Info struct {
	Nome    string `json:"nome"`
	Tipo    string `json:"tipo"` // pdf, jpeg, png, webp ou outro
	Tamanho int    `json:"tamanho"`
	SHA256  string `json:"sha256"`
	// Programas que gravaram o arquivo, como declarados nele mesmo.
	Programas []string `json:"programas,omitempty"`
	Criado    string   `json:"criado,omitempty"`
	Alterado  string   `json:"alterado,omitempty"`
	// Salvamentos conta as versões gravadas num PDF (atualização
	// incremental). Um PDF tirado do aplicativo do banco tem uma só.
	Salvamentos int  `json:"salvamentos,omitempty"`
	Assinado    bool `json:"assinado,omitempty"`
	// Editor é o primeiro programa reconhecido como editor, se houver.
	Editor string `json:"editor,omitempty"`
	// Alteracao diz o que mudou nas versões gravadas depois da primeira.
	Alteracao *Alteracao `json:"alteracao,omitempty"`
	// IA é a marca de geração ou edição por inteligência artificial
	// declarada no arquivo, se houver.
	IA *MarcaIA `json:"ia,omitempty"`

	chavesPNG []string
}

// Analisar identifica o tipo e lê os metadados que existirem.
func Analisar(nome string, dados []byte) Info {
	s := sha256.Sum256(dados)
	i := Info{Nome: nome, Tamanho: len(dados), SHA256: hex.EncodeToString(s[:]), Tipo: "outro"}
	switch {
	case bytes.HasPrefix(dados, []byte("%PDF-")) || bytes.Contains(dados[:min(len(dados), 1024)], []byte("%PDF-")):
		i.Tipo = "pdf"
		pdf(&i, dados)
	case bytes.HasPrefix(dados, []byte{0xFF, 0xD8, 0xFF}):
		i.Tipo = "jpeg"
		jpeg(&i, dados)
	case bytes.HasPrefix(dados, []byte("\x89PNG\r\n\x1a\n")):
		i.Tipo = "png"
		png(&i, dados)
	case len(dados) > 12 && string(dados[:4]) == "RIFF" && string(dados[8:12]) == "WEBP":
		i.Tipo = "webp"
	}
	xmp(&i, dados)
	i.Programas = unicos(i.Programas)
	i.IA = marcaIA(dados, i.Programas, i.chavesPNG)
	for _, p := range i.Programas {
		if e := Editor(p); e != "" {
			i.Editor = e
			break
		}
	}
	return i
}

var (
	reProdutor = regexp.MustCompile(`/(Producer|Creator)\s*\(((?:\\.|[^\\)])*)\)`)
	reData     = regexp.MustCompile(`/(CreationDate|ModDate)\s*\(D:(\d{14})`)
)

func pdf(i *Info, d []byte) {
	for _, m := range reProdutor.FindAllSubmatch(d, -1) {
		i.Programas = append(i.Programas, textoPDF(m[2]))
	}
	for _, m := range reData.FindAllSubmatch(d, -1) {
		v := string(m[2])
		v = v[6:8] + "/" + v[4:6] + "/" + v[0:4] + " " + v[8:10] + ":" + v[10:12] + ":" + v[12:14]
		if string(m[1]) == "CreationDate" && i.Criado == "" {
			i.Criado = v
		} else if string(m[1]) == "ModDate" {
			i.Alterado = v
		}
	}
	i.Salvamentos = bytes.Count(d, []byte("%%EOF"))
	// Um PDF "linearizado" (otimizado para a web) já nasce com duas marcas
	// de fim; não é edição.
	linearizado := bytes.Contains(d[:min(len(d), 2048)], []byte("/Linearized"))
	if linearizado && i.Salvamentos > 1 {
		i.Salvamentos--
	}
	i.Assinado = bytes.Contains(d, []byte("/ByteRange")) && bytes.Contains(d, []byte("/Sig"))
	if i.Salvamentos > 1 {
		i.Alteracao = alteracoes(d, linearizado)
	}
}

// textoPDF decodifica uma string literal de PDF, inclusive em UTF-16.
func textoPDF(b []byte) string {
	var out []byte
	for k := 0; k < len(b); k++ {
		if b[k] == '\\' && k+1 < len(b) {
			k++
			switch c := b[k]; {
			case c >= '0' && c <= '7':
				v, n := 0, 0
				for n < 3 && k < len(b) && b[k] >= '0' && b[k] <= '7' {
					v = v*8 + int(b[k]-'0')
					k, n = k+1, n+1
				}
				k--
				out = append(out, byte(v))
			case c == 'n':
				out = append(out, '\n')
			default:
				out = append(out, c)
			}
			continue
		}
		out = append(out, b[k])
	}
	if len(out) >= 2 && out[0] == 0xFE && out[1] == 0xFF {
		var r []rune
		for k := 2; k+1 < len(out); k += 2 {
			r = append(r, rune(binary.BigEndian.Uint16(out[k:])))
		}
		return strings.TrimSpace(string(r))
	}
	if utf8.Valid(out) {
		return strings.TrimSpace(string(out))
	}
	// PDFDocEncoding coincide com o Latin-1 nos caracteres que importam.
	r := make([]rune, len(out))
	for k, c := range out {
		r[k] = rune(c)
	}
	return strings.TrimSpace(string(r))
}

// jpeg lê a marca "Software" do EXIF.
func jpeg(i *Info, d []byte) {
	for k := 2; k+4 < len(d); {
		if d[k] != 0xFF {
			return
		}
		marca := d[k+1]
		tam := int(binary.BigEndian.Uint16(d[k+2:]))
		if marca == 0xDA || tam < 2 || k+2+tam > len(d) {
			return
		}
		seg := d[k+4 : k+2+tam]
		if marca == 0xE1 && bytes.HasPrefix(seg, []byte("Exif\x00\x00")) {
			if s := exifSoftware(seg[6:]); s != "" {
				i.Programas = append(i.Programas, s)
			}
		}
		k += 2 + tam
	}
}

func exifSoftware(t []byte) string {
	if len(t) < 8 {
		return ""
	}
	var o binary.ByteOrder
	switch string(t[:2]) {
	case "II":
		o = binary.LittleEndian
	case "MM":
		o = binary.BigEndian
	default:
		return ""
	}
	ifd := int(o.Uint32(t[4:]))
	if ifd+2 > len(t) {
		return ""
	}
	n := int(o.Uint16(t[ifd:]))
	for e := 0; e < n; e++ {
		p := ifd + 2 + e*12
		if p+12 > len(t) {
			return ""
		}
		if o.Uint16(t[p:]) != 0x0131 { // Software
			continue
		}
		cnt := int(o.Uint32(t[p+4:]))
		var v []byte
		if cnt <= 4 {
			v = t[p+8 : p+8+cnt]
		} else {
			off := int(o.Uint32(t[p+8:]))
			if off+cnt > len(t) {
				return ""
			}
			v = t[off : off+cnt]
		}
		return strings.TrimSpace(strings.TrimRight(string(v), "\x00"))
	}
	return ""
}

// png lê os blocos de texto com a chave "Software".
func png(i *Info, d []byte) {
	for k := 8; k+12 <= len(d); {
		tam := int(binary.BigEndian.Uint32(d[k:]))
		tipo := string(d[k+4 : k+8])
		if k+12+tam > len(d) {
			return
		}
		corpo := d[k+8 : k+8+tam]
		if tipo == "tEXt" || tipo == "iTXt" || tipo == "zTXt" {
			if chave, _, ok := bytes.Cut(corpo, []byte{0}); ok {
				i.chavesPNG = append(i.chavesPNG, string(chave))
			}
		}
		if tipo == "tEXt" || tipo == "iTXt" {
			if chave, valor, ok := bytes.Cut(corpo, []byte{0}); ok && string(chave) == "Software" {
				if tipo == "iTXt" && len(valor) > 2 {
					// compressão, método, idioma e palavra traduzida
					partes := bytes.SplitN(valor[2:], []byte{0}, 3)
					if len(partes) == 3 {
						valor = partes[2]
					}
				}
				i.Programas = append(i.Programas, strings.TrimSpace(string(valor)))
			}
		}
		if tipo == "IEND" {
			return
		}
		k += 12 + tam
	}
}

var reXMP = regexp.MustCompile(`(?:xmp:CreatorTool|pdf:Producer|photoshop:History|stEvt:softwareAgent)(?:="([^"]+)"|>([^<]+)<)`)

// xmp procura os campos de programa do bloco XMP, que existe em PDF, JPEG
// e PNG.
func xmp(i *Info, d []byte) {
	if !bytes.Contains(d, []byte("x:xmpmeta")) && !bytes.Contains(d, []byte("<rdf:RDF")) {
		return
	}
	for _, m := range reXMP.FindAllSubmatch(d, -1) {
		v := string(m[1])
		if v == "" {
			v = string(m[2])
		}
		if v = strings.TrimSpace(v); v != "" {
			i.Programas = append(i.Programas, v)
		}
	}
}

func unicos(l []string) []string {
	visto := map[string]bool{}
	var out []string
	for _, s := range l {
		if s != "" && !visto[s] {
			visto[s] = true
			out = append(out, s)
		}
	}
	return out
}
