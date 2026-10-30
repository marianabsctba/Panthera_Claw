// Gera nucleo/bancos/bancos.json a partir das listas públicas do Banco
// Central: participantes do Pix (ISO-8859-1, separado por ponto e vírgula)
// e participantes do STR (UTF-8, separado por vírgula), que traz o número
// do banco usado em boletos.
//
//	go run ./ferramentas/bancos -pix dados/participantes-pix.csv -str dados/participantes-str.csv
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

type banco struct {
	ISPB     string `json:"ispb"`
	Nome     string `json:"nome"`
	CNPJ     string `json:"cnpj,omitempty"`
	Tipo     string `json:"tipo,omitempty"`
	Compe    string `json:"compe,omitempty"`
	Situacao string `json:"situacao"`
}

func latin1(b []byte) string {
	r := make([]rune, len(b))
	for i, c := range b {
		r[i] = rune(c)
	}
	return string(r)
}

func main() {
	pix := flag.String("pix", "dados/participantes-pix.csv", "lista de participantes do Pix")
	str := flag.String("str", "dados/participantes-str.csv", "lista de participantes do STR")
	saida := flag.String("saida", "nucleo/bancos/bancos.json", "arquivo gerado")
	flag.Parse()

	bancos := map[string]*banco{}

	bp, err := os.ReadFile(*pix)
	falhar(err)
	situacao := ""
	for _, linha := range strings.Split(strings.ReplaceAll(latin1(bp), "\r", ""), "\n") {
		switch {
		case strings.HasPrefix(linha, "Lista de participantes ativos"):
			situacao = "ativo"
			continue
		case strings.HasPrefix(linha, "Lista de instituições em processo de adesão"):
			situacao = "adesao"
			continue
		}
		c := strings.Split(linha, ";")
		if len(c) < 5 || situacao == "" || len(strings.TrimSpace(c[2])) != 8 {
			continue
		}
		ispb := strings.TrimSpace(c[2])
		if _, ok := bancos[ispb]; ok {
			continue
		}
		bancos[ispb] = &banco{
			ISPB: ispb, Nome: strings.TrimSpace(c[1]), CNPJ: strings.TrimSpace(c[3]),
			Tipo: strings.TrimSpace(c[4]), Situacao: situacao,
		}
	}

	bs, err := os.ReadFile(*str)
	falhar(err)
	bs = bytes.TrimPrefix(bs, []byte("\xef\xbb\xbf"))
	linhas, err := csv.NewReader(bytes.NewReader(bs)).ReadAll()
	falhar(err)
	for _, c := range linhas[1:] {
		if len(c) < 3 {
			continue
		}
		ispb, compe := strings.TrimSpace(c[0]), strings.TrimSpace(c[2])
		if compe == "n/a" {
			compe = ""
		}
		if b, ok := bancos[ispb]; ok {
			b.Compe = compe
		} else if compe != "" {
			// Fora do Pix, mas emite boleto: entra para o número do banco.
			bancos[ispb] = &banco{ISPB: ispb, Nome: strings.TrimSpace(c[1]), Compe: compe, Situacao: "str"}
		}
	}

	lista := make([]*banco, 0, len(bancos))
	for _, b := range bancos {
		lista = append(lista, b)
	}
	sort.Slice(lista, func(i, j int) bool { return lista[i].ISPB < lista[j].ISPB })
	j, err := json.MarshalIndent(lista, "", " ")
	falhar(err)
	falhar(os.WriteFile(*saida, append(j, '\n'), 0o644))
	fmt.Printf("%d instituições gravadas em %s\n", len(lista), *saida)
}

func falhar(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
