// Conferência de boleto e de Pix copia e cola (QR Code): entrada e
// resultado.
import { useRef, useState } from "react";
import { AlertTriangle, ArrowLeft, Check, Image as Imagem, Info, QrCode, Receipt, Search, ShieldCheck, Volume2, X, XCircle } from "lucide-react";
import type { Achado, ResultadoCodigo } from "./tipos";
import { falar, temVoz } from "./voz";

interface EntradaProps {
  tipo: "boleto" | "pix";
  conferir: (texto: string) => void;
  lerImagem: (f: File) => void;
  voltar: () => void;
  erro?: string;
}

export function CodigoEntrada({ tipo, conferir, lerImagem, voltar, erro }: EntradaProps) {
  const [texto, setTexto] = useState("");
  const entrada = useRef<HTMLInputElement>(null);
  const boleto = tipo === "boleto";
  return (
    <section className="conferencia">
      <h1>{boleto ? "Conferir um boleto" : "Conferir um QR Code de Pix"}</h1>
      <p className="grande">
        {boleto
          ? "Cole ou digite a linha digitável do boleto (os números que ficam em cima do código de barras)."
          : "Escolha a imagem do QR Code ou cole o código do Pix copia e cola."}
      </p>
      {erro && <p className="aviso-erro" role="alert">{erro}</p>}
      {!boleto && (
        <>
          <input ref={entrada} type="file" accept="image/*" hidden
            onChange={(e) => { const f = e.target.files?.[0]; e.target.value = ""; if (f) lerImagem(f); }} />
          <button className="botao-gigante" onClick={() => entrada.current?.click()}>
            <Imagem size={32} aria-hidden /> Escolher imagem do QR Code
          </button>
          <p className="dica">ou cole o código abaixo:</p>
        </>
      )}
      <form className="campos" onSubmit={(e) => { e.preventDefault(); if (texto.trim()) conferir(texto); }}>
        <label>
          <span>{boleto ? "Linha digitável" : "Pix copia e cola"}</span>
          <textarea className="codigo" rows={boleto ? 3 : 4} value={texto} onChange={(e) => setTexto(e.target.value)}
            spellCheck={false} autoCapitalize="off" inputMode={boleto ? "numeric" : "text"}
            placeholder={boleto ? "00190.00009 01234.567890 12345.678901 2 12340000010000" : "00020126..."} />
        </label>
        <button type="submit" className={boleto ? "botao-gigante" : "botao-secundario largo"} disabled={!texto.trim()}>
          <Search size={boleto ? 32 : 24} aria-hidden /> Conferir
        </button>
        <button type="button" className="botao-secundario" onClick={voltar}><ArrowLeft size={22} aria-hidden /> Voltar</button>
      </form>
    </section>
  );
}

const cabecalhos = {
  errado: { classe: "falso", icone: <XCircle size={72} aria-hidden />, titulo: "Não pague: este código está errado",
    texto: "Ele tem um erro que um código verdadeiro não teria. Se você copiou certinho, ele foi alterado." },
  suspeito: { classe: "suspeito", icone: <AlertTriangle size={72} aria-hidden />, titulo: "Cuidado com este código",
    texto: "Encontramos algo estranho. Confira com quem mandou antes de pagar." },
  confira: { classe: "sem-erros", icone: <ShieldCheck size={72} aria-hidden />, titulo: "O código está inteiro. Agora confira:",
    texto: "Veja se os dados abaixo são de quem você espera pagar. Se não forem, não pague." },
};

const icones: Record<string, React.ReactNode> = {
  falso: <X size={28} aria-hidden />, alerta: <AlertTriangle size={26} aria-hidden />,
  ok: <Check size={28} aria-hidden />, info: <Info size={26} aria-hidden />,
};
const ordem: Record<string, number> = { falso: 0, alerta: 1, info: 2, ok: 3 };

interface ResultadoProps {
  r: ResultadoCodigo;
  outro: () => void;
  inicio: () => void;
}

export function CodigoResultado({ r, outro, inicio }: ResultadoProps) {
  const cab = cabecalhos[r.veredito];
  const achados = [...r.achados].sort((a: Achado, b: Achado) => ordem[a.peso] - ordem[b.peso]);
  const ouvir = () => falar([cab.titulo, cab.texto, ...r.resumo.map((l) => `${l.rotulo}: ${l.valor}.`),
    ...achados.filter((a) => a.peso === "falso" || a.peso === "alerta").map((a) => a.titulo)].join(" "));
  return (
    <section className="resultado">
      <div className={`veredito ${cab.classe}`} role="status">
        {cab.icone}
        <h1>{cab.titulo}</h1>
        <p>{cab.texto}</p>
      </div>
      {r.resumo.length > 0 && (
        <dl className="resumo">
          {r.resumo.map((l, i) => (
            <div key={i}><dt>{l.rotulo}</dt><dd>{l.valor}</dd></div>
          ))}
        </dl>
      )}
      {temVoz && (
        <div className="acoes-resultado">
          <button className="botao-secundario" onClick={ouvir}><Volume2 size={24} aria-hidden /> Ouvir o resultado</button>
        </div>
      )}
      <h2>O que encontramos</h2>
      <ul className="achados">
        {achados.map((a, i) => (
          <li key={i} className={`achado ${a.peso}`}>
            <span className="selo">{icones[a.peso]}</span>
            <span>{a.titulo}</span>
          </li>
        ))}
      </ul>
      <button className="botao-gigante" onClick={outro}>
        {r.tipo === "boleto" ? <Receipt size={30} aria-hidden /> : <QrCode size={30} aria-hidden />}
        {r.tipo === "boleto" ? " Conferir outro boleto" : " Conferir outro QR Code"}
      </button>
      <div className="acoes-resultado">
        <button className="botao-secundario" onClick={inicio}><ArrowLeft size={22} aria-hidden /> Voltar ao início</button>
      </div>
      <details className="detalhes">
        <summary>Detalhes técnicos</summary>
        <table><tbody>
          {r.boleto && <>
            <tr><th>Linha digitável</th><td>{r.boleto.linha}</td></tr>
            <tr><th>Código de barras</th><td>{r.boleto.barras}</td></tr>
          </>}
          {r.pix && <>
            <tr><th>Código</th><td>{r.pix.texto}</td></tr>
            <tr><th>Soma de verificação</th><td>no código {r.pix.crc || "(ausente)"}, calculada {r.pix.crcCerto}</td></tr>
            {r.pix.txid && <tr><th>Identificador</th><td>{r.pix.txid}</td></tr>}
            {r.pix.url && <tr><th>Endereço da cobrança</th><td>{r.pix.url}</td></tr>}
          </>}
          {r.achados.filter((a) => a.detalhe).map((a, i) => <tr key={i}><th>{a.codigo}</th><td>{a.detalhe}</td></tr>)}
        </tbody></table>
      </details>
    </section>
  );
}
