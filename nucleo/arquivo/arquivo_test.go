package arquivo

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
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
