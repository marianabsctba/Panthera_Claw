package arquivo

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"os"
	"testing"
)

func TestPDFEditadoNoWord(t *testing.T) {
	d := []byte("%PDF-1.7\n1 0 obj<</Producer (Microsoft\\256 Word para Microsoft 365)/CreationDate (D:20261029143200-03'00')/ModDate (D:20261029150000-03'00')>>endobj\n%%EOF\n2 0 obj<<>>endobj\n%%EOF\n")
	i := Analisar("c.pdf", d)
	if i.Tipo != "pdf" || i.Editor != "Microsoft Word" || i.Salvamentos != 2 {
		t.Fatalf("%+v", i)
	}
	if i.Criado != "29/10/2026 14:32:00" || i.Alterado != "29/10/2026 15:00:00" {
		t.Fatalf("datas %+v", i)
	}
}

func TestPDFUTF16(t *testing.T) {
	d := []byte("%PDF-1.4\n<</Producer (\xfe\xff\x00i\x00T\x00e\x00x\x00t)>>\n%%EOF")
	if i := Analisar("c.pdf", d); len(i.Programas) != 1 || i.Programas[0] != "iText" || i.Editor != "" {
		t.Fatalf("%+v", i)
	}
}

func TestPDFLinearizado(t *testing.T) {
	d := []byte("%PDF-1.4\n1 0 obj<</Linearized 1>>endobj\n%%EOF\nresto\n%%EOF")
	if i := Analisar("c.pdf", d); i.Salvamentos != 1 {
		t.Fatalf("%+v", i)
	}
}

func TestJPEGComSoftware(t *testing.T) {
	// EXIF mínimo em little endian com uma entrada Software.
	var tiff bytes.Buffer
	tiff.WriteString("II*\x00")
	binary.Write(&tiff, binary.LittleEndian, uint32(8))
	binary.Write(&tiff, binary.LittleEndian, uint16(1))
	sw := "Adobe Photoshop 25.0\x00"
	binary.Write(&tiff, binary.LittleEndian, uint16(0x0131))
	binary.Write(&tiff, binary.LittleEndian, uint16(2))
	binary.Write(&tiff, binary.LittleEndian, uint32(len(sw)))
	binary.Write(&tiff, binary.LittleEndian, uint32(8+2+12+4))
	binary.Write(&tiff, binary.LittleEndian, uint32(0))
	tiff.WriteString(sw)
	seg := append([]byte("Exif\x00\x00"), tiff.Bytes()...)
	var j bytes.Buffer
	j.Write([]byte{0xFF, 0xD8, 0xFF, 0xE1})
	binary.Write(&j, binary.BigEndian, uint16(len(seg)+2))
	j.Write(seg)
	j.Write([]byte{0xFF, 0xDA, 0, 2})
	i := Analisar("c.jpg", j.Bytes())
	if i.Tipo != "jpeg" || i.Editor != "Photoshop" {
		t.Fatalf("%+v", i)
	}
}

func TestPNGComSoftware(t *testing.T) {
	var p bytes.Buffer
	p.WriteString("\x89PNG\r\n\x1a\n")
	bloco := func(tipo string, corpo []byte) {
		binary.Write(&p, binary.BigEndian, uint32(len(corpo)))
		p.WriteString(tipo)
		p.Write(corpo)
		binary.Write(&p, binary.BigEndian, crc32.ChecksumIEEE(append([]byte(tipo), corpo...)))
	}
	bloco("tEXt", []byte("Software\x00Canva"))
	bloco("IEND", nil)
	if i := Analisar("c.png", p.Bytes()); i.Tipo != "png" || i.Editor != "Canva" {
		t.Fatalf("%+v", i)
	}
}

// PDFs fictícios: o original, uma edição salva por cima (valor coberto e
// reescrito) e um salvamento que só mudou metadados.
func TestAlteracoesDoPDF(t *testing.T) {
	ler := func(n string) Info {
		d, err := os.ReadFile("testdata/" + n)
		if err != nil {
			t.Fatal(err)
		}
		return Analisar(n, d)
	}
	if a := ler("original.pdf").Alteracao; a != nil {
		t.Fatalf("original com alteração: %+v", a)
	}
	a := ler("editado-incremental.pdf").Alteracao
	if a == nil || !a.Pagina || len(a.Textos) != 1 || a.Textos[0] != "R$ 890,00" {
		t.Fatalf("edição: %+v", a)
	}
	if a := ler("salvo-so-metadados.pdf").Alteracao; a == nil || a.Pagina {
		t.Fatalf("só metadados: %+v", a)
	}
}

func pngCom(chave, valor string) []byte {
	var p bytes.Buffer
	p.WriteString("\x89PNG\r\n\x1a\n")
	bloco := func(tipo string, corpo []byte) {
		binary.Write(&p, binary.BigEndian, uint32(len(corpo)))
		p.WriteString(tipo)
		p.Write(corpo)
		binary.Write(&p, binary.BigEndian, crc32.ChecksumIEEE(append([]byte(tipo), corpo...)))
	}
	bloco("tEXt", []byte(chave+"\x00"+valor))
	bloco("IEND", nil)
	return p.Bytes()
}

func TestMarcasDeIA(t *testing.T) {
	casos := []struct {
		nome   string
		dados  []byte
		tipo   string
		origem string
		gerado string
	}{
		{"sd.png", pngCom("parameters", "comprovante pix nubank, Steps: 30"), "gerada", "prompt", "Stable Diffusion"},
		{"comfy.png", pngCom("workflow", "{}"), "gerada", "prompt", "ComfyUI"},
		{"iptc.jpg", []byte("\xff\xd8\xff<x:xmpmeta><Iptc4xmpExt:DigitalSourceType>http://cv.iptc.org/newscodes/digitalsourcetype/trainedAlgorithmicMedia</Iptc4xmpExt:DigitalSourceType></x:xmpmeta>"), "gerada", "iptc", ""},
		{"editada.jpg", []byte("\xff\xd8\xff<x:xmpmeta>digitalsourcetype/compositeWithTrainedAlgorithmicMedia</x:xmpmeta>"), "editada", "iptc", ""},
		{"c2pa.png", append(pngCom("Software", "x"), []byte("\x00jumb\x00c2pa\x00c2pa.claim\x00\x6fclaim_generatorx\x16ChatGPT/1.0 c2pa-rs/0.3")...), "gerada", "c2pa", "ChatGPT"},
		{"mj.png", pngCom("Software", "Midjourney v7"), "gerada", "programa", "Midjourney"},
	}
	for _, c := range casos {
		m := Analisar(c.nome, c.dados).IA
		if m == nil || m.Tipo != c.tipo || m.Origem != c.origem || m.Gerador != c.gerado {
			t.Errorf("%s: %+v", c.nome, m)
		}
	}
	// Selo C2PA de câmera ou editor comum não é IA.
	if m := Analisar("camera.jpg", []byte("\xff\xd8\xff c2pa.claim claim_generator Leica M11-P")).IA; m != nil {
		t.Errorf("câmera virou IA: %+v", m)
	}
	if m := Analisar("print.png", pngCom("Software", "Android")).IA; m != nil {
		t.Errorf("print comum virou IA: %+v", m)
	}
}
