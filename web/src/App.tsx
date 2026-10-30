import { useCallback, useEffect, useRef, useState } from "react";
import { Download, FileUp, Github, Loader2, Lock, Minus, Plus } from "lucide-react";
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
      campos, arquivo: info, pdfSemTexto: leitura.pdfSemTexto,
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

  // Colar (Ctrl+V) um print copiado.
  useEffect(() => {
    const colar = (e: ClipboardEvent) => {
      const f = Array.from(e.clipboardData?.files ?? [])[0];
      if (f && (tela.nome === "inicio" || tela.nome === "erro")) { e.preventDefault(); void processar(f); }
    };
    window.addEventListener("paste", colar);
    return () => window.removeEventListener("paste", colar);
  }, [tela.nome, processar]);

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

