package arquivo

import (
	"bytes"
	"strings"
)

// MarcaIA é a declaração, gravada no próprio arquivo, de que ele foi gerado
// ou editado por inteligência artificial. Só conta o que o arquivo declara:
// sem marca não quer dizer que não houve IA (o WhatsApp, por exemplo, apaga
// os metadados ao recomprimir a imagem).
type MarcaIA struct {
	Tipo    string `json:"tipo"`    // gerada ou editada
	Origem  string `json:"origem"`  // c2pa, iptc, prompt ou programa
	Gerador string `json:"gerador"` // quando identificado
}

// Selo de procedência C2PA (Content Credentials), gravado por ChatGPT,
// Firefly, Designer e outros. Os rótulos aparecem em texto dentro do
// manifesto, em qualquer formato de arquivo.
var rotulosC2PA = [][]byte{[]byte("c2pa.claim"), []byte("c2pa.signature"), []byte("c2pa.assertions")}

// Códigos IPTC de origem digital.
var (
	iptcGerada  = []byte("trainedAlgorithmicMedia")
	iptcEditada = []byte("compositeWithTrainedAlgorithmicMedia")
)

// geradores conhecidos, procurados no selo C2PA e nos campos de programa.
var geradores = []struct{ trecho, nome string }{
	{"chatgpt", "ChatGPT"},
	{"openai", "OpenAI"},
	{"dall-e", "DALL·E"},
	{"dall·e", "DALL·E"},
	{"sora", "Sora"},
	{"firefly", "Adobe Firefly"},
	{"microsoft designer", "Microsoft Designer"},
	{"bing image creator", "Bing Image Creator"},
	{"gemini", "Gemini"},
	{"imagen", "Imagen"},
	{"meta ai", "Meta AI"},
	{"midjourney", "Midjourney"},
	{"stable diffusion", "Stable Diffusion"},
	{"stability ai", "Stability AI"},
	{"comfyui", "ComfyUI"},
	{"novelai", "NovelAI"},
	{"invokeai", "InvokeAI"},
	{"ideogram", "Ideogram"},
	{"leonardo.ai", "Leonardo"},
	{"flux", "FLUX"},
	{"grok", "Grok"},
	{"recraft", "Recraft"},
	{"runway", "Runway"},
}

// Chaves de texto que geradores de imagem gravam no PNG com o prompt.
var chavesPrompt = map[string]string{
	"parameters":        "Stable Diffusion",
	"prompt":            "ComfyUI",
	"workflow":          "ComfyUI",
	"sd-metadata":       "InvokeAI",
	"invokeai_metadata": "InvokeAI",
	"Dream":             "Stable Diffusion",
}

// marcaIA procura as marcas no arquivo inteiro e nos programas declarados.
func marcaIA(d []byte, programas []string, chavesPNG []string) *MarcaIA {
	c2pa := false
	gerador := ""
	for _, r := range rotulosC2PA {
		if k := bytes.Index(d, r); k >= 0 {
			c2pa = true
			// O manifesto fica todo junto: procura o gerador só ao redor.
			gerador = acharGerador(strings.ToLower(trechoLegivel(d[max(0, k-65536):min(len(d), k+65536)])))
			break
		}
	}
	for _, p := range programas {
		if g := acharGerador(strings.ToLower(p)); g != "" {
			gerador = g
			break
		}
	}

	switch {
	case bytes.Contains(d, iptcEditada):
		return &MarcaIA{Tipo: "editada", Origem: origem(c2pa, "iptc"), Gerador: gerador}
	case bytes.Contains(d, iptcGerada):
		return &MarcaIA{Tipo: "gerada", Origem: origem(c2pa, "iptc"), Gerador: gerador}
	}
	for _, k := range chavesPNG {
		if g, ok := chavesPrompt[k]; ok {
			if gerador == "" {
				gerador = g
			}
			return &MarcaIA{Tipo: "gerada", Origem: "prompt", Gerador: gerador}
		}
	}
	if gerador != "" {
		o := "programa"
		if c2pa {
			o = "c2pa"
		}
		return &MarcaIA{Tipo: "gerada", Origem: o, Gerador: gerador}
	}
	return nil
}

func origem(c2pa bool, outra string) string {
	if c2pa {
		return "c2pa"
	}
	return outra
}

func acharGerador(s string) string {
	for _, g := range geradores {
		if strings.Contains(s, g.trecho) {
			return g.nome
		}
	}
	return ""
}

// trechoLegivel junta as sequências de texto legível do arquivo (o selo
// C2PA guarda o nome do gerador como texto dentro do CBOR).
func trechoLegivel(d []byte) string {
	var b strings.Builder
	run := 0
	inicio := 0
	for k, c := range d {
		if c >= 0x20 && c < 0x7f {
			if run == 0 {
				inicio = k
			}
			run++
			continue
		}
		if run >= 4 {
			b.Write(d[inicio:k])
			b.WriteByte(' ')
		}
		run = 0
		if b.Len() > 1<<20 {
			break
		}
	}
	if run >= 4 {
		b.Write(d[inicio:])
	}
	return b.String()
}
