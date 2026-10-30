import { useCallback, useEffect, useRef, useState } from "react";
import { ClipboardPaste, Download, FileUp, Github, Loader2, Lock, Minus, Plus } from "lucide-react";
import ComoFunciona from "./ComoFunciona";
import Conferencia from "./Conferencia";
import ResultadoTela from "./Resultado";
import { guardar, vistos } from "./historico";
import { ler, type Leitura } from "./leitor";
import { nucleo } from "./nucleo";
import type { Campos, InfoArquivo, Resultado } from "./tipos";
import { calar } from "./voz";

type Tela =
  | { nome: "inicio" }
  | { nome: "lendo"; etapa: string }
  | { nome: "conferir"; leitura: Leitura; campos: Campos; info: InfoArquivo }
  | { nome: "resultado"; leitura: Leitura; campos: Campos; info: InfoArquivo; resultado: Resultado }
  | { nome: "erro"; mensagem: string };

interface EventoInstalar extends Event { prompt(): Promise<void> }

const TAMANHOS = [1, 1.15, 1.3, 1.5];

const podeColar = typeof navigator !== "undefined" && !!navigator.clipboard?.read;

// Alguns programas só entregam a imagem pela lista de itens, não pela de
// arquivos.
function arquivoColado(d: DataTransfer): File | null {
  const f = Array.from(d.files)[0];
  if (f) return f;
  for (const item of Array.from(d.items)) {
    if (item.kind === "file") {
      const g = item.getAsFile();
      if (g) return g;
    }
  }
  return null;
}

function pareceComprovante(texto: string): boolean {
  const junto = texto.replace(/\s/g, "");
  return /[EeDd][0-9OoIl]{20}/.test(junto) || (/R\$/.test(texto) && /\d{2}\/\d{2}\/\d{2,4}|\d{1,2} (de )?[a-zç]{3}/i.test(texto));
}

function tamanhoSalvo(): number {
  try {
    const n = Number(localStorage.getItem("claw:tamanho"));
    return TAMANHOS.includes(n) ? n : 1;
  } catch {
    return 1;
  }
}

export default function App() {
  const [tela, setTela] = useState<Tela>({ nome: "inicio" });
  const [como, setComo] = useState(location.hash === "#como");
  const [tamanho, setTamanho] = useState(tamanhoSalvo);
  const [instalar, setInstalar] = useState<EventoInstalar | null>(null);
  const [arrastando, setArrastando] = useState(false);
  const entrada = useRef<HTMLInputElement>(null);

  useEffect(() => {
    document.documentElement.style.fontSize = `${tamanho * 100}%`;
    try { localStorage.setItem("claw:tamanho", String(tamanho)); } catch { /* sem armazenamento */ }
  }, [tamanho]);

  useEffect(() => {
    const f = () => setComo(location.hash === "#como");
    window.addEventListener("hashchange", f);
    const i = (e: Event) => { e.preventDefault(); setInstalar(e as EventoInstalar); };
    window.addEventListener("beforeinstallprompt", i);
    return () => { window.removeEventListener("hashchange", f); window.removeEventListener("beforeinstallprompt", i); };
  }, []);

  const conferir = useCallback(async (leitura: Leitura, campos: Campos, info: InfoArquivo) => {
    const resultado = await nucleo.conferir({
      campos, arquivo: info, pdfSemTexto: leitura.pdfSemTexto, sobrepostos: leitura.sobrepostos,
      vistos: vistos().filter((v) => v.sha256 !== info.sha256),
    });
    guardar(resultado);
    setTela({ nome: "resultado", leitura, campos, info, resultado });
    window.scrollTo(0, 0);
  }, []);

  const processar = useCallback(async (arquivo: File) => {
    calar();
    if (location.hash) history.replaceState(null, "", location.pathname + location.search);
    setComo(false);
    setTela({ nome: "lendo", etapa: "Abrindo o comprovante" });
    try {
      const dados = new Uint8Array(await arquivo.arrayBuffer());
      const info = await nucleo.analisar(arquivo.name || "comprovante", dados);
      const leitura = await ler(arquivo, (etapa) => setTela({ nome: "lendo", etapa }));
      const campos = await nucleo.ler(leitura.texto);
      // Texto do PDF é exato; o que veio da imagem a pessoa confere antes.
      if (leitura.origem === "pdf" && campos.id) await conferir(leitura, campos, info);
      else setTela({ nome: "conferir", leitura, campos, info });
    } catch (e) {
      console.error(e);
      setTela({ nome: "erro", mensagem: "Não conseguimos abrir este arquivo. Tente uma foto (print) ou o PDF do comprovante." });
    }
  }, [conferir]);

  // Comprovante compartilhado por outro aplicativo (celular com o site instalado).
  useEffect(() => {
    if (!new URLSearchParams(location.search).has("compartilhado")) return;
    history.replaceState(null, "", location.pathname);
    void (async () => {
      const c = await caches.open("claw-compartilhado");
      const r = await c.match("arquivo");
      if (!r) return;
      await c.delete("arquivo");
      const nome = decodeURIComponent(r.headers.get("X-Nome") ?? "comprovante");
      await processar(new File([await r.blob()], nome, { type: r.headers.get("Content-Type") ?? "" }));
    })();
  }, [processar]);

  // Texto colado (copiado de um PDF ou digitado): vai direto para a
  // conferência dos campos, sem imagem.
  const processarTexto = useCallback(async (texto: string) => {
    calar();
    setComo(false);
    setTela({ nome: "lendo", etapa: "Lendo o texto colado" });
    const info = await nucleo.analisar("texto colado", new TextEncoder().encode(texto));
    const campos = await nucleo.ler(texto);
    setTela({ nome: "conferir", leitura: { texto, origem: "texto", pdfSemTexto: false, confianca: 100, previa: "" }, campos, info });
  }, []);

  // Colar (Ctrl+V) em qualquer tela: imagem, PDF ou texto do comprovante.
  useEffect(() => {
    const colar = (e: ClipboardEvent) => {
      if (tela.nome === "lendo" || !e.clipboardData) return;
      const f = arquivoColado(e.clipboardData);
      if (f) { e.preventDefault(); void processar(f); return; }
      // Dentro de um campo, colar texto é só preencher o campo.
      const alvo = e.target as HTMLElement | null;
      if (alvo?.closest("input, textarea, [contenteditable]")) return;
      const texto = e.clipboardData.getData("text/plain");
      if (pareceComprovante(texto)) { e.preventDefault(); void processarTexto(texto); }
    };
    document.addEventListener("paste", colar);
    return () => document.removeEventListener("paste", colar);
  }, [tela.nome, processar, processarTexto]);

  // Botão "Colar comprovante", para quem não usa Ctrl+V.
  const colarDoBotao = async () => {
    try {
      for (const item of await navigator.clipboard.read()) {
        const tipo = item.types.find((t) => t.startsWith("image/") || t === "application/pdf");
        if (tipo) {
          const b = await item.getType(tipo);
          await processar(new File([b], tipo === "application/pdf" ? "colado.pdf" : "colado." + tipo.split("/")[1], { type: tipo }));
          return;
        }
      }
      const texto = await navigator.clipboard.readText();
      if (pareceComprovante(texto)) { await processarTexto(texto); return; }
      setTela({ nome: "erro", mensagem: "Não achamos um comprovante copiado. Copie a imagem do comprovante e toque em Colar de novo." });
    } catch {
      setTela({ nome: "erro", mensagem: "O navegador não deixou colar por este botão. Use o botão Escolher comprovante ou as teclas Ctrl e V juntas." });
    }
  };

  const recomecar = () => { calar(); setTela({ nome: "inicio" }); window.scrollTo(0, 0); };

  const soltar = (e: React.DragEvent) => {
    e.preventDefault();
    setArrastando(false);
    const f = e.dataTransfer.files[0];
    if (f) void processar(f);
  };

  return (
    <div className="app" onDragOver={(e) => { e.preventDefault(); setArrastando(true); }} onDragLeave={() => setArrastando(false)} onDrop={soltar}>
      <header className="topo">
        <a className="marca" href="./" onClick={(e) => { e.preventDefault(); history.replaceState(null, "", location.pathname); setComo(false); recomecar(); }}>
          <img src={`${import.meta.env.BASE_URL}pantera.webp`} alt="" />
          <span className="nome">PANTHERA CLAW</span>
        </a>
        <div className="tamanho" role="group" aria-label="Tamanho da letra">
          <button aria-label="Diminuir a letra" disabled={tamanho === TAMANHOS[0]} onClick={() => setTamanho(TAMANHOS[Math.max(0, TAMANHOS.indexOf(tamanho) - 1)])}><Minus size={18} /> A</button>
          <button aria-label="Aumentar a letra" disabled={tamanho === TAMANHOS[TAMANHOS.length - 1]} onClick={() => setTamanho(TAMANHOS[Math.min(TAMANHOS.length - 1, TAMANHOS.indexOf(tamanho) + 1)])}><Plus size={18} /> A</button>
        </div>
      </header>

      <main className="conteudo">
        {como ? (
          <ComoFunciona voltar={() => { history.replaceState(null, "", location.pathname); setComo(false); }} />
        ) : tela.nome === "inicio" || tela.nome === "erro" ? (
          <section className="inicio">
            <h1>Esse comprovante de Pix é verdadeiro?</h1>
            <p className="grande">Escolha o comprovante que você recebeu. A gente confere em poucos segundos.</p>
            {tela.nome === "erro" && <p className="aviso-erro" role="alert">{tela.mensagem}</p>}
            <input ref={entrada} type="file" accept="image/*,application/pdf,.pdf" hidden
              onChange={(e) => { const f = e.target.files?.[0]; e.target.value = ""; if (f) void processar(f); }} />
            <button className={`botao-gigante ${arrastando ? "ativo" : ""}`} onClick={() => entrada.current?.click()}>
              <FileUp size={36} aria-hidden /> Escolher comprovante
            </button>
            {podeColar && (
              <button className="botao-secundario largo" onClick={() => void colarDoBotao()}>
                <ClipboardPaste size={24} aria-hidden /> Colar comprovante copiado
              </button>
            )}
            <p className="dica">Pode ser o print da tela ou o PDF. No computador, você também pode arrastar o arquivo ou colar com Ctrl+V.</p>
            <p className="privacidade"><Lock size={20} aria-hidden /> O comprovante não sai do seu aparelho.</p>
            {instalar && (
              <button className="botao-secundario" onClick={async () => { await instalar.prompt(); setInstalar(null); }}>
                <Download size={22} aria-hidden /> Instalar no celular
              </button>
            )}
            <a className="link-grande" href="#como">Como funciona e como usar pelo WhatsApp</a>
          </section>
        ) : tela.nome === "lendo" ? (
          <section className="lendo" aria-live="polite">
            <Loader2 size={64} className="gira" aria-hidden />
            <p className="grande">{tela.etapa}…</p>
            <p className="dica">Isso leva alguns segundos.</p>
          </section>
        ) : tela.nome === "conferir" ? (
          <Conferencia
            leitura={tela.leitura}
            campos={tela.campos}
            voltar={recomecar}
            confirmar={(campos) => void conferir(tela.leitura, campos, tela.info)}
          />
        ) : (
          <ResultadoTela
            resultado={tela.resultado}
            corrigir={() => setTela({ nome: "conferir", leitura: tela.leitura, campos: tela.campos, info: tela.info })}
            outro={recomecar}
          />
        )}
      </main>

      <footer className="rodape">
        <a href="#como">Como funciona</a>
        <a href="https://github.com/marianabsctba/Panthera_Claw" target="_blank" rel="noreferrer"><Github size={16} aria-hidden /> Código aberto</a>
      </footer>
    </div>
  );
}

