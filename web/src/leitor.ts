// Transforma o arquivo do comprovante em texto, sem sair do aparelho. PDF
// com texto é lido direto (exato); imagem, ou PDF que é só imagem, passa
// pela leitura de imagem.
import * as pdfjs from "pdfjs-dist";
import workerPdf from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import type { Worker as Ocr } from "tesseract.js";

pdfjs.GlobalWorkerOptions.workerSrc = workerPdf;

export interface Leitura {
  texto: string;
  origem: "pdf" | "imagem" | "texto";
  pdfSemTexto: boolean;
  // Confiança da leitura de imagem, de 0 a 100. No PDF é sempre 100.
  confianca: number;
  // Imagem para mostrar ao lado dos campos lidos.
  previa: string;
  // Textos do PDF desenhados um em cima do outro (o de cima veio depois).
  sobrepostos?: { cima: string; baixo: string }[];
}

export type Etapa = (texto: string) => void;

export async function ler(arquivo: File, etapa: Etapa): Promise<Leitura> {
  const dados = new Uint8Array(await arquivo.arrayBuffer());
  const ehPdf = arquivo.type === "application/pdf" || /\.pdf$/i.test(arquivo.name)
    || new TextDecoder().decode(dados.slice(0, 1024)).includes("%PDF-");
  if (ehPdf) return lerPdf(dados, etapa);
  etapa("Lendo a imagem do comprovante");
  const bitmap = await createImageBitmap(arquivo);
  const tela = preparar(bitmap);
  const { texto, confianca } = await reconhecer(tela, etapa);
  return { texto, origem: "imagem", pdfSemTexto: false, confianca, previa: tela.toDataURL("image/jpeg", 0.85) };
}

async function lerPdf(dados: Uint8Array, etapa: Etapa): Promise<Leitura> {
  etapa("Abrindo o PDF");
  const tarefa = pdfjs.getDocument({
    data: dados.slice(),
    disableFontFace: true,
    standardFontDataUrl: new URL(`${import.meta.env.BASE_URL}pdf/fontes/`, location.href).href,
  });
  const doc = await tarefa.promise;
  try {
    const linhas: string[] = [];
    const sobrepostos: { cima: string; baixo: string }[] = [];
    const paginas = Math.min(doc.numPages, 3);
    for (let n = 1; n <= paginas; n++) {
      const pagina = await doc.getPage(n);
      const conteudo = await pagina.getTextContent();
      linhas.push(...juntarEmLinhas(conteudo.items as ItemTexto[]));
      sobrepostos.push(...acharSobrepostos(conteudo.items as ItemTexto[]));
    }
    const texto = linhas.join("\n");
    const primeira = await doc.getPage(1);
    const tela = await desenhar(primeira, 2);
    if (texto.replace(/\s/g, "").length >= 20) {
      return { texto, origem: "pdf", pdfSemTexto: false, confianca: 100, previa: tela.toDataURL("image/jpeg", 0.85), sobrepostos };
    }
    // PDF sem texto: é uma foto ou um print salvo como PDF.
    etapa("O PDF é uma imagem. Lendo a imagem");
    const { texto: lido, confianca } = await reconhecer(preparar(tela), etapa);
    return { texto: lido, origem: "imagem", pdfSemTexto: true, confianca, previa: tela.toDataURL("image/jpeg", 0.85) };
  } finally {
    void tarefa.destroy();
  }
}

interface ItemTexto { str: string; transform: number[]; width: number; height: number }

// O PDF guarda pedaços de texto soltos com posição. Junta os que estão na
// mesma altura numa linha; distância grande entre eles vira dois espaços,
// que a leitura entende como coluna ("Valor  R$ 10,00").
function juntarEmLinhas(itens: ItemTexto[]): string[] {
  const pedacos = itens.filter((i) => i.str.trim() !== "")
    .map((i) => ({ s: i.str, x: i.transform[4], y: i.transform[5], w: i.width, h: Math.abs(i.transform[3]) || i.height || 10 }));
  pedacos.sort((a, b) => b.y - a.y || a.x - b.x);
  const linhas: (typeof pedacos)[] = [];
  for (const p of pedacos) {
    const l = linhas[linhas.length - 1];
    if (l && Math.abs(l[0].y - p.y) < Math.max(2, p.h * 0.4)) l.push(p);
    else linhas.push([p]);
  }
  return linhas.map((l) => {
    l.sort((a, b) => a.x - b.x);
    let s = l[0].s;
    for (let k = 1; k < l.length; k++) {
      const vao = l[k].x - (l[k - 1].x + l[k - 1].w);
      s += (vao > l[k].h * 1.5 ? "  " : vao > l[k].h * 0.15 ? " " : "") + l[k].s;
    }
    return s.trim();
  });
}

// Dois textos diferentes ocupando o mesmo lugar: quem edita PDF costuma
// cobrir o valor antigo com um retângulo e escrever o novo por cima. A ordem
// dos itens é a ordem de desenho, então o posterior está por cima.
function acharSobrepostos(itens: ItemTexto[]): { cima: string; baixo: string }[] {
  const caixas = itens.filter((i) => i.str.trim().length >= 2).map((i) => {
    const h = Math.abs(i.transform[3]) || i.height || 10;
    return { s: i.str.trim(), x0: i.transform[4], x1: i.transform[4] + i.width, y0: i.transform[5] - h * 0.2, y1: i.transform[5] + h * 0.8 };
  });
  const achados: { cima: string; baixo: string }[] = [];
  for (let a = 0; a < caixas.length; a++) {
    for (let b = a + 1; b < caixas.length; b++) {
      const p = caixas[a], q = caixas[b];
      if (p.s === q.s) continue; // mesmo texto duas vezes é negrito simulado
      const iw = Math.min(p.x1, q.x1) - Math.max(p.x0, q.x0);
      const ih = Math.min(p.y1, q.y1) - Math.max(p.y0, q.y0);
      if (iw <= 0 || ih <= 0) continue;
      const menor = Math.min((p.x1 - p.x0) * (p.y1 - p.y0), (q.x1 - q.x0) * (q.y1 - q.y0));
      if (menor > 0 && (iw * ih) / menor > 0.4) achados.push({ cima: q.s, baixo: p.s });
    }
  }
  return achados.slice(0, 10);
}

async function desenhar(pagina: pdfjs.PDFPageProxy, escala: number): Promise<HTMLCanvasElement> {
  const vp = pagina.getViewport({ scale: escala });
  const tela = document.createElement("canvas");
  tela.width = Math.ceil(vp.width);
  tela.height = Math.ceil(vp.height);
  const ctx = tela.getContext("2d")!;
  ctx.fillStyle = "#fff";
  ctx.fillRect(0, 0, tela.width, tela.height);
  await pagina.render({ canvas: tela, canvasContext: ctx, viewport: vp }).promise;
  return tela;
}

// Prepara a imagem para a leitura: tamanho bom para letras de aplicativo,
// tons de cinza e, se o comprovante for de fundo escuro, cores invertidas.
function preparar(origem: ImageBitmap | HTMLCanvasElement): HTMLCanvasElement {
  const largura = origem.width;
  const escala = largura < 1200 ? Math.min(3, 1600 / largura) : largura > 2600 ? 2600 / largura : 1;
  const tela = document.createElement("canvas");
  tela.width = Math.round(origem.width * escala);
  tela.height = Math.round(origem.height * escala);
  const ctx = tela.getContext("2d", { willReadFrequently: true })!;
  ctx.imageSmoothingQuality = "high";
  ctx.drawImage(origem, 0, 0, tela.width, tela.height);
  const img = ctx.getImageData(0, 0, tela.width, tela.height);
  const d = img.data;
  let soma = 0;
  for (let k = 0; k < d.length; k += 4) {
    const cinza = 0.299 * d[k] + 0.587 * d[k + 1] + 0.114 * d[k + 2];
    d[k] = d[k + 1] = d[k + 2] = cinza;
    soma += cinza;
  }
  if (soma / (d.length / 4) < 110) {
    for (let k = 0; k < d.length; k += 4) d[k] = d[k + 1] = d[k + 2] = 255 - d[k];
  }
  ctx.putImageData(img, 0, 0);
  return tela;
}

let ocr: Promise<Ocr> | null = null;

function abrirOcr(etapa: Etapa): Promise<Ocr> {
  if (!ocr) {
    const base = new URL(`${import.meta.env.BASE_URL}ocr/`, location.href).href;
    ocr = (async () => {
      const { createWorker, OEM } = await import("tesseract.js");
      etapa("Preparando a leitura (na primeira vez demora um pouco mais)");
      return createWorker("por", OEM.LSTM_ONLY, {
        workerPath: base + "worker.min.js",
        corePath: base,
        langPath: base,
        workerBlobURL: false,
        gzip: true,
      });
    })();
    ocr.catch(() => { ocr = null; });
  }
  return ocr;
}

async function reconhecer(tela: HTMLCanvasElement, etapa: Etapa): Promise<{ texto: string; confianca: number }> {
  const w = await abrirOcr(etapa);
  etapa("Lendo as letras do comprovante");
  const { data } = await w.recognize(tela);
  return { texto: data.text, confianca: data.confidence };
}
