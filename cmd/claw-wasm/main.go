//go:build js && wasm

// Ponte entre o núcleo em Go e a interface no navegador. Roda dentro de um
// Web Worker; o comprovante fica na memória deste aparelho e nunca é
// enviado a lugar nenhum.
package main

import (
	"encoding/json"
	"syscall/js"
	"time"

	"github.com/marianabsctba/Panthera_Claw/nucleo/arquivo"
	"github.com/marianabsctba/Panthera_Claw/nucleo/conferir"
	"github.com/marianabsctba/Panthera_Claw/nucleo/leitura"
)

const versao = "0.1.0"

func bytesDe(v js.Value) []byte {
	b := make([]byte, v.Get("length").Int())
	js.CopyBytesToGo(b, v)
	return b
}

func jsonJS(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return js.Global().Get("Error").New(err.Error())
	}
	return string(b)
}

// metodo protege a ponte: um pânico vira erro em vez de derrubar o worker.
func metodo(f func(a []js.Value) any) js.Func {
	return js.FuncOf(func(_ js.Value, a []js.Value) (v any) {
		defer func() {
			if r := recover(); r != nil {
				v = js.Global().Get("Error").New("erro interno")
			}
		}()
		return f(a)
	})
}

func main() {
	api := js.Global().Get("Object").New()

	// ler(texto) -> JSON dos campos
	api.Set("ler", metodo(func(a []js.Value) any {
		return jsonJS(leitura.Ler(a[0].String()))
	}))

	// analisar(nome, bytes) -> JSON do arquivo
	api.Set("analisar", metodo(func(a []js.Value) any {
		return jsonJS(arquivo.Analisar(a[0].String(), bytesDe(a[1])))
	}))

	// conferir(entradaJSON) -> JSON do resultado
	api.Set("conferir", metodo(func(a []js.Value) any {
		var e conferir.Entrada
		if err := json.Unmarshal([]byte(a[0].String()), &e); err != nil {
			return js.Global().Get("Error").New(err.Error())
		}
		e.Agora = time.Now()
		return jsonJS(conferir.Conferir(e))
	}))

	api.Set("versao", versao)
	js.Global().Set("pantheraClaw", api)
	select {}
}
