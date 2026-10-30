package pix

import (
	"regexp"
	"strings"
)

// Tipos de chave Pix.
const (
	ChaveCPF       = "cpf"
	ChaveCNPJ      = "cnpj"
	ChaveTelefone  = "telefone"
	ChaveEmail     = "email"
	ChaveAleatoria = "aleatoria"
	ChaveMascarada = "mascarada" // parte escondida com asteriscos
	ChaveInvalida  = "invalida"
)

// Chave é o resultado da leitura de uma chave Pix impressa.
type Chave struct {
	Texto  string `json:"texto"`
	Tipo   string `json:"tipo"`
	Motivo string `json:"motivo,omitempty"` // quando inválida
	// Impossivel indica que a chave não pode existir (dígito verificador
	// ou DDD errado). Sem isso, a chave só tem forma estranha, o que pode
	// ser erro de leitura da imagem.
	Impossivel bool `json:"impossivel,omitempty"`
}

var (
	reEmail     = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	reAleatoria = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	reCNPJ      = regexp.MustCompile(`^[0-9A-Z]{12}[0-9]{2}$`)
)

// DDDs válidos no Brasil (Anatel).
var ddds = map[string]bool{}

func init() {
	for _, d := range strings.Fields(`11 12 13 14 15 16 17 18 19 21 22 24 27 28 31 32 33 34 35 37 38
		41 42 43 44 45 46 47 48 49 51 53 54 55 61 62 63 64 65 66 67 68 69 71 73 74 75 77 79
		81 82 83 84 85 86 87 88 89 91 92 93 94 95 96 97 98 99`) {
		ddds[d] = true
	}
}

// LerChave classifica uma chave Pix e confere o que dá para conferir.
func LerChave(s string) Chave {
	s = strings.TrimSpace(s)
	c := Chave{Texto: s}
	switch {
	case s == "":
		c.Tipo, c.Motivo = ChaveInvalida, "chave vazia"
	case strings.ContainsAny(s, "*•"):
		c.Tipo = ChaveMascarada
	case reEmail.MatchString(s):
		c.Tipo = ChaveEmail
	case reAleatoria.MatchString(s):
		c.Tipo = ChaveAleatoria
	default:
		so := somenteDocumento(s)
		switch {
		case strings.HasPrefix(s, "+") || (len(so) == 13 && strings.HasPrefix(so, "55")):
			c.Tipo = ChaveTelefone
			d := strings.TrimPrefix(so, "55")
			if len(d) != 11 && len(d) != 10 {
				c.Tipo, c.Motivo = ChaveInvalida, "telefone com quantidade errada de dígitos"
			} else if !ddds[d[:2]] {
				c.Tipo, c.Motivo, c.Impossivel = ChaveInvalida, "o DDD "+d[:2]+" não existe", true
			}
		case len(so) == 11 && CPFValido(so):
			c.Tipo = ChaveCPF
		case (len(so) == 11 || len(so) == 10) && pareceTelefone(s, so):
			// Celular impresso sem o +55: "(11) 98765-4321".
			c.Tipo = ChaveTelefone
		case len(so) == 11:
			// Onze dígitos que não formam CPF válido. Só é impossível se
			// estiver escrito como CPF (000.000.000-00); sem formatação
			// pode ser telefone ou erro de leitura.
			c.Tipo, c.Motivo = ChaveInvalida, "esse CPF não existe: os dígitos finais não batem"
			c.Impossivel = reCPFFormatado.MatchString(s)
		case len(so) == 14:
			c.Tipo = ChaveCNPJ
			if !CNPJValido(so) {
				c.Tipo, c.Motivo, c.Impossivel = ChaveInvalida, "esse CNPJ não existe: os dígitos finais não batem", true
			}
		default:
			c.Tipo, c.Motivo = ChaveInvalida, "não tem a forma de nenhum tipo de chave Pix"
		}
	}
	return c
}

var (
	reCPFFormatado = regexp.MustCompile(`^\d{3}\.\d{3}\.\d{3}-\d{2}$`)
	reTelefone     = regexp.MustCompile(`^\(?\d{2}\)?[\s-]*9?\d{4}[\s-]?\d{4}$`)
)

// pareceTelefone: DDD existente e, no celular de 11 dígitos, o 9 na frente;
// ou escrito como telefone, com parênteses no DDD.
func pareceTelefone(s, so string) bool {
	if !ddds[so[:2]] || reCPFFormatado.MatchString(s) {
		return false
	}
	if strings.Contains(s, "(") {
		return reTelefone.MatchString(strings.TrimSpace(s))
	}
	return len(so) == 11 && so[2] == '9'
}

func somenteDocumento(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if r >= '0' && r <= '9' || r >= 'A' && r <= 'Z' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CPFValido confere os dois dígitos verificadores.
func CPFValido(s string) bool {
	if len(s) != 11 || strings.Count(s, s[:1]) == 11 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	dv := func(n int) byte {
		soma := 0
		for i := 0; i < n; i++ {
			soma += int(s[i]-'0') * (n + 1 - i)
		}
		r := soma * 10 % 11
		if r == 10 {
			r = 0
		}
		return byte('0' + r)
	}
	return dv(9) == s[9] && dv(10) == s[10]
}

// CNPJValido confere os dígitos verificadores, inclusive do CNPJ
// alfanumérico (letras valem o código ASCII menos 48).
func CNPJValido(s string) bool {
	s = strings.ToUpper(s)
	if !reCNPJ.MatchString(s) || strings.Count(s, s[:1]) == 14 {
		return false
	}
	dv := func(n int) byte {
		pesos := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}[13-n:]
		soma := 0
		for i := 0; i < n; i++ {
			soma += int(s[i]-'0') * pesos[i]
		}
		r := soma % 11
		if r < 2 {
			return '0'
		}
		return byte('0' + 11 - r)
	}
	return dv(12) == s[12] && dv(13) == s[13]
}
