// Package brcode lê o Pix copia e cola (o texto por trás do QR Code do
// Pix), no padrão EMV do Banco Central, e confere a soma de verificação.
package brcode

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Codigo é o que o Pix copia e cola revela.
type Codigo struct {
	Texto     string `json:"texto"`
	Dinamico  bool   `json:"dinamico"` // valor e recebedor vêm do banco na hora
	Chave     string `json:"chave,omitempty"`
	Info      string `json:"info,omitempty"`
	URL       string `json:"url,omitempty"`
	Dominio   string `json:"dominio,omitempty"`
	Valor     string `json:"valor,omitempty"` // como no código: 10.50
	Nome      string `json:"nome,omitempty"`
	Cidade    string `json:"cidade,omitempty"`
	TxID      string `json:"txid,omitempty"`
	CRC       string `json:"crc"`
	CRCCerto  string `json:"crcCerto"`
	CRCValido bool   `json:"crcValido"`
	EhPix     bool   `json:"ehPix"`
}

// Parece diz se o texto tem cara de Pix copia e cola.
func Parece(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "000201") && strings.Contains(strings.ToLower(s), "br.gov.bcb.pix")
}

// Ler interpreta o código. Erro só quando nem a estrutura dá para ler.
func Ler(s string) (Codigo, error) {
	s = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", ""))
	c := Codigo{Texto: s}
	campos, err := tlv(s)
	if err != nil {
		return c, err
	}
	for _, f := range campos {
		switch f.id {
		case "01":
			c.Dinamico = f.valor == "12"
		case "26":
			sub, err := tlv(f.valor)
			if err != nil {
				continue
			}
			for _, g := range sub {
				switch g.id {
				case "00":
					c.EhPix = strings.EqualFold(g.valor, "br.gov.bcb.pix")
				case "01":
					c.Chave = g.valor
				case "02":
					c.Info = g.valor
				case "25":
					c.URL = g.valor
					c.Dinamico = true
					if u, err := url.Parse("https://" + strings.TrimPrefix(strings.TrimPrefix(g.valor, "https://"), "http://")); err == nil {
						c.Dominio = u.Hostname()
					}
				}
			}
		case "54":
			c.Valor = f.valor
		case "59":
			c.Nome = f.valor
		case "60":
			c.Cidade = f.valor
		case "62":
			if sub, err := tlv(f.valor); err == nil {
				for _, g := range sub {
					if g.id == "05" {
						c.TxID = g.valor
					}
				}
			}
		case "63":
			c.CRC = strings.ToUpper(f.valor)
		}
	}
	if i := strings.LastIndex(s, "6304"); i >= 0 {
		c.CRCCerto = fmt.Sprintf("%04X", crc16(s[:i+4]))
	}
	c.CRCValido = c.CRC != "" && c.CRC == c.CRCCerto
	return c, nil
}

// ValorReais devolve o valor no formato brasileiro, ou "".
func (c Codigo) ValorReais() string {
	if c.Valor == "" {
		return ""
	}
	f, err := strconv.ParseFloat(c.Valor, 64)
	if err != nil {
		return c.Valor
	}
	centavos := int64(f*100 + 0.5)
	inteiro := strconv.FormatInt(centavos/100, 10)
	var b strings.Builder
	for k, r := range inteiro {
		if k > 0 && (len(inteiro)-k)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return fmt.Sprintf("R$ %s,%02d", b.String(), centavos%100)
}

type campo struct{ id, valor string }

func tlv(s string) ([]campo, error) {
	var out []campo
	for k := 0; k < len(s); {
		if k+4 > len(s) {
			return out, fmt.Errorf("o código termina no meio de um campo")
		}
		n, err := strconv.Atoi(s[k+2 : k+4])
		if err != nil || k+4+n > len(s) {
			return out, fmt.Errorf("o código está quebrado perto da posição %d", k+1)
		}
		out = append(out, campo{s[k : k+2], s[k+4 : k+4+n]})
		k += 4 + n
	}
	return out, nil
}

// crc16 é o CRC-16/CCITT-FALSE (polinômio 0x1021, início 0xFFFF).
func crc16(s string) uint16 {
	crc := uint16(0xFFFF)
	for k := 0; k < len(s); k++ {
		crc ^= uint16(s[k]) << 8
		for b := 0; b < 8; b++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
