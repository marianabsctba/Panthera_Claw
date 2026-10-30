// Lê um QR Code de uma imagem, no próprio aparelho.
import jsQR from "jsqr";

export async function lerQR(arquivo: Blob): Promise<string | null> {
  const bitmap = await createImageBitmap(arquivo);
  // Tenta no tamanho original e reduzido: foto grande demais atrapalha.
  for (const largura of [bitmap.width, 1200, 800]) {
    if (largura > bitmap.width) continue;
    const esc = largura / bitmap.width;
    const tela = document.createElement("canvas");
    tela.width = Math.round(bitmap.width * esc);
    tela.height = Math.round(bitmap.height * esc);
    const ctx = tela.getContext("2d", { willReadFrequently: true })!;
    ctx.drawImage(bitmap, 0, 0, tela.width, tela.height);
    const img = ctx.getImageData(0, 0, tela.width, tela.height);
    const achado = jsQR(img.data, img.width, img.height, { inversionAttempts: "attemptBoth" });
    if (achado?.data) return achado.data;
  }
  return null;
}

// Pix copia e cola ou linha de boleto (44, 47 ou 48 dígitos).
export function pareceCodigo(texto: string): boolean {
  const t = texto.trim();
  if (t.startsWith("000201") && /br\.gov\.bcb\.pix/i.test(t)) return true;
  const d = t.replace(/[\s.\-]/g, "");
  return /^\d+$/.test(d) && [44, 47, 48].includes(d.length);
}
