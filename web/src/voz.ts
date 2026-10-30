// Lê o resultado em voz alta, com a voz em português do aparelho.
export const temVoz = typeof window !== "undefined" && "speechSynthesis" in window;

export function falar(texto: string) {
  if (!temVoz) return;
  speechSynthesis.cancel();
  const u = new SpeechSynthesisUtterance(texto);
  u.lang = "pt-BR";
  u.rate = 0.9;
  const voz = speechSynthesis.getVoices().find((v) => v.lang.toLowerCase().startsWith("pt-br"))
    ?? speechSynthesis.getVoices().find((v) => v.lang.toLowerCase().startsWith("pt"));
  if (voz) u.voice = voz;
  speechSynthesis.speak(u);
}

export function calar() {
  if (temVoz) speechSynthesis.cancel();
}
