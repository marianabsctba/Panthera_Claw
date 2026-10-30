// Comando claw: confere comprovantes Pix pela linha de comando.
//
//	claw conferir -texto comprovante.txt comprovante.pdf
//	claw conferir -id E18236120202610291732abc12XYZ789 -data "29/10/2026 14:32" -banco Nubank
//	claw codigo E18236120202610291732abc12XYZ789
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/arquivo"
	"github.com/marianabsctba/Panthera_Claw/nucleo/conferir"
	"github.com/marianabsctba/Panthera_Claw/nucleo/leitura"
	"github.com/marianabsctba/Panthera_Claw/nucleo/pix"
)

const versao = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		ajuda()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "conferir":
		os.Exit(cmdConferir(os.Args[2:]))
	case "codigo":
		os.Exit(cmdCodigo(os.Args[2:]))
	case "versao", "-v", "--version":
		fmt.Println("claw", versao)
	default:
		ajuda()
		os.Exit(2)
	}
}

func ajuda() {
	fmt.Fprint(os.Stderr, `Panthera Claw: confere comprovantes Pix.

  claw conferir [opções] [arquivo]   confere um comprovante
  claw codigo ID                     mostra o que o código da transação revela
  claw versao

Opções de conferir:
  -texto arquivo.txt   texto do comprovante (copiado do PDF ou lido da imagem)
  -id, -valor, -data, -banco, -chave   campos digitados (têm prioridade sobre o texto)
  -json                resultado em JSON

Saída: 0 sem erros, 1 comprovante falso ou suspeito, 2 erro de uso.
`)
}

func cmdConferir(args []string) int {
	fs := flag.NewFlagSet("conferir", flag.ExitOnError)
	texto := fs.String("texto", "", "")
	id := fs.String("id", "", "")
	valor := fs.String("valor", "", "")
	data := fs.String("data", "", "")
	banco := fs.String("banco", "", "")
	chave := fs.String("chave", "", "")
	saidaJSON := fs.Bool("json", false, "")
	fs.Parse(args)

	var c leitura.Campos
	if *texto != "" {
		b, err := os.ReadFile(*texto)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		c = leitura.Ler(string(b))
	}
	for _, p := range []struct{ v, campo *string }{{id, &c.ID}, {valor, &c.Valor}, {data, &c.DataHora}, {banco, &c.BancoPagador}, {chave, &c.Chave}} {
		if *p.v != "" {
			*p.campo = *p.v
		}
	}

	e := conferir.Entrada{Campos: c, Agora: time.Now()}
	if fs.NArg() > 0 {
		b, err := os.ReadFile(fs.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		a := arquivo.Analisar(filepath.Base(fs.Arg(0)), b)
		e.Arquivo = &a
	}
	r := conferir.Conferir(e)

	if *saidaJSON {
		j, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(j))
	} else {
		imprimir(r)
	}
	if r.Veredito == conferir.VereditoFalso || r.Veredito == conferir.VereditoSuspeito {
		return 1
	}
	return 0
}

var titulos = map[string]string{
	conferir.VereditoFalso:      "COMPROVANTE FALSO",
	conferir.VereditoSuspeito:   "COMPROVANTE SUSPEITO",
	conferir.VereditoSemErros:   "NENHUM ERRO ENCONTRADO (confirme no extrato)",
	conferir.VereditoIncompleto: "NÃO DEU PARA CONFERIR (falta o código da transação)",
}

var marcas = map[string]string{conferir.Falso: "✗", conferir.Alerta: "!", conferir.OK: "✓", conferir.Info: "·"}

func imprimir(r conferir.Resultado) {
	fmt.Println(titulos[r.Veredito])
	fmt.Println()
	for _, a := range r.Achados {
		fmt.Printf("  %s %s\n", marcas[a.Peso], a.Titulo)
		if a.Detalhe != "" {
			fmt.Printf("    %s\n", a.Detalhe)
		}
	}
	if r.Arquivo != nil {
		fmt.Printf("\n  arquivo %s (%s), SHA-256 %s\n", r.Arquivo.Nome, r.Arquivo.Tipo, r.Arquivo.SHA256)
	}
}

func cmdCodigo(args []string) int {
	if len(args) == 0 {
		ajuda()
		return 2
	}
	e, err := pix.LerE2E(strings.Join(args, ""))
	if err != nil {
		fmt.Fprintln(os.Stderr, "código inválido:", err)
		return 1
	}
	r := conferir.Conferir(conferir.Entrada{Campos: leitura.Campos{ID: e.Texto}, Agora: time.Now()})
	d := r.Codigo
	fmt.Printf("tipo        %s\n", map[string]string{"E": "pagamento", "D": "devolução"}[e.Tipo])
	banco := "desconhecido na lista do Banco Central"
	if d.Conhecido {
		banco = d.Banco
		if d.Marca != "" {
			banco += " (" + d.Marca + ")"
		}
	}
	fmt.Printf("instituição %s  ISPB %s\n", banco, e.ISPB)
	fmt.Printf("quando      %s (Brasília), %s\n", d.Brasilia, d.UTC)
	fmt.Printf("sequencial  %s\n", e.Sequencial)
	for _, c := range e.Correcoes {
		fmt.Printf("ajuste      %s\n", c)
	}
	return 0
}
