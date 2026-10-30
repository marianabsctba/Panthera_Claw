// Comprovantes já conferidos neste aparelho, para achar o mesmo código em
// comprovantes diferentes. Fica só no navegador.
import type { Resultado, Visto } from "./tipos";

const CHAVE = "claw:vistos";
const LIMITE = 300;

export function vistos(): Visto[] {
  try {
    return JSON.parse(localStorage.getItem(CHAVE) ?? "[]") as Visto[];
  } catch {
    return [];
  }
}

export function guardar(r: Resultado) {
  if (!r.codigo || !r.arquivo) return;
  const lista = vistos().filter((v) => !(v.sha256 === r.arquivo!.sha256 && v.id === r.codigo!.texto));
  lista.unshift({ id: r.codigo.texto, sha256: r.arquivo.sha256, valor: r.campos.valor, quando: new Date().toLocaleDateString("pt-BR") });
  try {
    localStorage.setItem(CHAVE, JSON.stringify(lista.slice(0, LIMITE)));
  } catch {
    // Sem armazenamento (janela anônima): a conferência segue sem histórico.
  }
}

export function apagar() {
  try {
    localStorage.removeItem(CHAVE);
  } catch {
    // nada a fazer
  }
}
