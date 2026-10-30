// Preferências deste aparelho (tamanho da letra, tema, voz). Sem
// armazenamento (janela anônima), tudo funciona com o padrão.
export function lerPreferencia(chave: string): string | null {
  try {
    return localStorage.getItem(chave);
  } catch {
    return null;
  }
}

export function gravarPreferencia(chave: string, valor: string) {
  try {
    localStorage.setItem(chave, valor);
  } catch {
    // sem armazenamento: vale só nesta visita
  }
}
