package brcode

import "testing"

// Exemplo do manual do BR Code do Banco Central.
const exemplo = "00020126580014br.gov.bcb.pix0136123e4567-e12b-12d1-a456-4266554400005204000053039865802BR5913Fulano de Tal6008BRASILIA62070503***63041D3D"

func TestExemploDoBancoCentral(t *testing.T) {
	c, err := Ler(exemplo)
	if err != nil {
		t.Fatal(err)
	}
	if !c.CRCValido || !c.EhPix || c.Chave != "123e4567-e12b-12d1-a456-426655440000" || c.Nome != "Fulano de Tal" || c.Cidade != "BRASILIA" {
		t.Fatalf("%+v", c)
	}
	if !Parece(exemplo) {
		t.Fatal("não reconheceu")
	}
}

func TestChaveTrocada(t *testing.T) {
	// O golpe do QR trocado: outra chave, sem recalcular a soma.
	trocado := exemplo[:40] + "9" + exemplo[41:]
	c, err := Ler(trocado)
	if err != nil {
		t.Fatal(err)
	}
	if c.CRCValido {
		t.Fatalf("troca não detectada: %+v", c)
	}
}

func TestComValorEDinamico(t *testing.T) {
	semCRC := f("00", "01") + f("01", "12") +
		f("26", f("00", "br.gov.bcb.pix")+f("25", "pix.exemplo.com.br/qr/v2/cobv/abc12")) +
		f("52", "0000") + f("53", "986") + f("54", "15.50") + f("58", "BR") + f("59", "Loja XY") + f("60", "SAO PAULO") +
		f("62", f("05", "***")) + "6304"
	codigo := semCRC + toHex(crc16(semCRC))
	c, err := Ler(codigo)
	if err != nil {
		t.Fatal(err)
	}
	if !c.CRCValido || !c.Dinamico || c.Dominio != "pix.exemplo.com.br" || c.ValorReais() != "R$ 15,50" {
		t.Fatalf("%+v %s", c, c.ValorReais())
	}
}

func TestQuebrado(t *testing.T) {
	if _, err := Ler("000201265"); err == nil {
		t.Fatal("deveria falhar")
	}
}

func f(id, v string) string {
	return id + string([]byte{byte('0' + len(v)/10), byte('0' + len(v)%10)}) + v
}

func toHex(v uint16) string {
	const h = "0123456789ABCDEF"
	return string([]byte{h[v>>12], h[v>>8&15], h[v>>4&15], h[v&15]})
}
