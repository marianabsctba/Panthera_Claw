import { useEffect, useState } from "react";
import { AlertTriangle, ArrowLeft, CalendarClock, Check, HelpCircle, Info, Pencil, Printer, SearchCheck, Send, Volume2, X, XCircle } from "lucide-react";
import { gravarPreferencia, lerPreferencia } from "./preferencias";
import type { Achado, Resultado, Veredito } from "./tipos";
import { falar, temVoz } from "./voz";

const cabecalhos: Record<Veredito, { classe: string; icone: React.ReactNode; titulo: string; texto: string }> = {
  falso: {
    classe: "falso", icone: <XCircle size={72} aria-hidden />,
    titulo: "Este comprovante é falso",
    texto: "Ele tem erros que um comprovante verdadeiro não teria.",
  },
  nao_pago: {
    classe: "nao-pago", icone: <CalendarClock size={72} aria-hidden />,
    titulo: "Isto ainda não é um pagamento",
    texto: "O comprovante mostra um Pix agendado, pendente, cancelado ou devolvido. O dinheiro pode nunca chegar.",
  },
  suspeito: {
    classe: "suspeito", icone: <AlertTriangle size={72} aria-hidden />,
    titulo: "Cuidado: este comprovante tem sinais suspeitos",
    texto: "Isso não prova que é falso, mas confira no extrato do seu banco antes de entregar qualquer coisa.",
  },
  sem_erros: {
    classe: "sem-erros", icone: <SearchCheck size={72} aria-hidden />,
    titulo: "Não encontramos erros",
    texto: "Mesmo assim, só o extrato do seu banco confirma que o dinheiro entrou. Confira antes de entregar.",
  },
  incompleto: {
    classe: "incompleto", icone: <HelpCircle size={72} aria-hidden />,
    titulo: "Não deu para conferir",
    texto: "Não achamos o código da transação neste comprovante. Peça o comprovante completo ou confira no extrato.",
  },
};

const ordem: Record<string, number> = { falso: 0, nao_pago: 1, alerta: 2, ok: 3, info: 4 };

const icones: Record<string, React.ReactNode> = {
  falso: <X size={28} aria-hidden />,
  nao_pago: <CalendarClock size={26} aria-hidden />,
  alerta: <AlertTriangle size={26} aria-hidden />,
  ok: <Check size={28} aria-hidden />,
  info: <Info size={26} aria-hidden />,
};

const rotulosPeso: Record<string, string> = { falso: "Erro", nao_pago: "Não é pagamento", alerta: "Atenção", ok: "Certo", info: "Informação" };

interface Props {
  resultado: Resultado;
  corrigir: () => void;
  outro: () => void;
}

function imprimir() {
  document.querySelectorAll("details").forEach((d) => { d.open = true; });
  window.print();
}

export default function ResultadoTela({ resultado: r, corrigir, outro }: Props) {
  const cab = cabecalhos[r.veredito];
  const achados = [...(r.achados ?? [])].sort((a, b) => ordem[a.peso] - ordem[b.peso]);
  const importantes = achados.filter((a) => a.peso === "falso" || a.peso === "nao_pago" || a.peso === "alerta");

  const ouvir = () => {
    const partes = [cab.titulo + ".", cab.texto, ...importantes.map((a) => a.titulo)];
    falar(partes.join(" "));
  };

  const [vozSempre, setVozSempre] = useState(() => lerPreferencia("claw:voz") === "sim");
  useEffect(() => {
    if (vozSempre) ouvir();
    // Só ao abrir o resultado.
  }, []);
  const mudarVoz = (sim: boolean) => {
    setVozSempre(sim);
    gravarPreferencia("claw:voz", sim ? "sim" : "nao");
    if (sim) ouvir();
  };

  const mandar = async () => {
    const texto = [
      `Conferi um comprovante de Pix${r.campos.valor ? ` de ${r.campos.valor}` : ""} no Panthera Claw.`,
      `Resultado: ${cab.titulo}.`,
      ...importantes.map((a) => `- ${a.titulo}`),
      "",
      "https://marianabsctba.github.io/Panthera_Claw/",
    ].join("\n");
    if (navigator.share) {
      try { await navigator.share({ text: texto }); return; } catch { /* cancelado: cai no WhatsApp */ }
    }
    window.open(`https://wa.me/?text=${encodeURIComponent(texto)}`, "_blank", "noopener");
  };

  return (
    <section className="resultado">
      <div className={`veredito ${cab.classe}`} role="status">
        {cab.icone}
        <h1>{cab.titulo}</h1>
        <p>{cab.texto}</p>
      </div>

      <div className="acoes-resultado">
        {temVoz && <button className="botao-secundario" onClick={ouvir}><Volume2 size={24} aria-hidden /> Ouvir o resultado</button>}
        <button className="botao-secundario" onClick={() => void mandar()}><Send size={24} aria-hidden /> Mandar para alguém da família</button>
      </div>
      {temVoz && (
        <label className="opcao-grande">
          <input type="checkbox" checked={vozSempre} onChange={(e) => mudarVoz(e.target.checked)} />
          Sempre ler o resultado em voz alta
        </label>
      )}

      <h2>O que encontramos</h2>
      <ul className="achados">
        {achados.map((a, i) => <ItemAchado key={i} a={a} />)}
      </ul>

      <h2>O que fazer agora</h2>
      <ol className="passos">
        {r.veredito === "nao_pago" ? (
          <>
            <li>Não entregue nada por enquanto.</li>
            <li>Espere o dinheiro aparecer no extrato do seu banco. Agendamento pode ser cancelado a qualquer hora.</li>
            <li>Se a pessoa insistir para você entregar antes, desconfie.</li>
          </>
        ) : r.veredito === "falso" ? (
          <>
            <li>Não entregue o produto nem devolva troco.</li>
            <li>Abra o aplicativo do seu banco e veja que o dinheiro não entrou.</li>
            <li>Se já entregou, guarde este comprovante e registre um boletim de ocorrência.</li>
          </>
        ) : (
          <>
            <li>Abra o aplicativo do seu banco.</li>
            <li>Veja no extrato se o valor {r.campos.valor ? `de ${r.campos.valor} ` : ""}entrou.</li>
            <li>Só entregue depois de ver o dinheiro na sua conta.</li>
          </>
        )}
      </ol>

      <button className="botao-gigante" onClick={outro}><ArrowLeft size={30} aria-hidden /> Conferir outro comprovante</button>
      <div className="acoes-resultado">
        <button className="botao-secundario" onClick={corrigir}><Pencil size={22} aria-hidden /> Corrigir os dados lidos</button>
        <button className="botao-secundario" onClick={imprimir}><Printer size={22} aria-hidden /> Imprimir</button>
      </div>

      <Detalhes r={r} />
    </section>
  );
}

function ItemAchado({ a }: { a: Achado }) {
  return (
    <li className={`achado ${a.peso}`}>
      <span className="selo" title={rotulosPeso[a.peso]}>{icones[a.peso]}</span>
      <span>
        <span className="sr">{rotulosPeso[a.peso]}: </span>
        {a.titulo}
      </span>
    </li>
  );
}

function Detalhes({ r }: { r: Resultado }) {
  const linha = (rotulo: string, valor?: string | number | boolean) =>
    valor === undefined || valor === "" ? null : <tr><th>{rotulo}</th><td>{String(valor)}</td></tr>;
  return (
    <details className="detalhes">
      <summary>Detalhes técnicos (para perito ou advogado)</summary>
      {r.codigo && (
        <>
          <h3>Código da transação</h3>
          <table><tbody>
            {linha("Código", r.codigo.texto)}
            {linha("Tipo", r.codigo.tipo === "E" ? "pagamento" : "devolução")}
            {linha("ISPB", r.codigo.ispb)}
            {linha("Instituição", r.codigo.conhecido ? r.codigo.banco : "fora da lista do Banco Central")}
            {linha("Data e hora (Brasília)", r.codigo.brasilia)}
            {linha("Data e hora (UTC)", r.codigo.utc)}
            {linha("Sequencial", r.codigo.sequencial)}
            {r.codigo.correcoes?.length ? linha("Ajustes de leitura", r.codigo.correcoes.join("; ")) : null}
          </tbody></table>
        </>
      )}
      <h3>Dados do comprovante</h3>
      <table><tbody>
        {linha("Valor", r.campos.valor)}
        {linha("Data e hora", r.campos.dataHora)}
        {linha("Banco de quem pagou", r.campos.bancoPagador)}
        {linha("Nome de quem pagou", r.campos.nomePagador)}
        {linha("Banco de quem recebeu", r.campos.bancoRecebedor)}
        {linha("Nome de quem recebeu", r.campos.nomeRecebedor)}
        {linha("Chave Pix", r.campos.chave)}
        {linha("Marca no topo", r.campos.emissor)}
      </tbody></table>
      {r.arquivo && (
        <>
          <h3>Arquivo</h3>
          <table><tbody>
            {linha("Nome", r.arquivo.nome)}
            {linha("Tipo", r.arquivo.tipo.toUpperCase())}
            {linha("Tamanho", `${r.arquivo.tamanho.toLocaleString("pt-BR")} bytes`)}
            {linha("SHA-256", r.arquivo.sha256)}
            {r.arquivo.programas?.length ? linha("Programas registrados", r.arquivo.programas.join("; ")) : null}
            {linha("Criado em", r.arquivo.criado)}
            {linha("Alterado em", r.arquivo.alterado)}
            {r.arquivo.tipo === "pdf" ? linha("Versões gravadas", r.arquivo.salvamentos) : null}
            {r.arquivo.assinado ? linha("Assinatura digital", "presente") : null}
            {r.arquivo.alteracao ? linha("Objetos gravados depois", r.arquivo.alteracao.objetos) : null}
            {r.arquivo.alteracao?.textos?.length ? linha("Texto escrito depois", r.arquivo.alteracao.textos.join(" | ")) : null}
            {r.arquivo.ia ? linha("Marca de inteligência artificial", `${r.arquivo.ia.tipo === "editada" ? "editado" : "gerado"} (${r.arquivo.ia.origem}${r.arquivo.ia.gerador ? ", " + r.arquivo.ia.gerador : ""})`) : null}
          </tbody></table>
        </>
      )}
      <h3>Cada conferência</h3>
      <table><tbody>
        {r.achados.map((a, i) => (
          <tr key={i}><th>{rotulosPeso[a.peso]}</th><td>{a.titulo}{a.detalhe ? <><br /><span className="suave">{a.detalhe}</span></> : null}</td></tr>
        ))}
      </tbody></table>
      <p className="suave">Conferido em {new Date().toLocaleString("pt-BR")}. A conferência roda inteira neste aparelho.</p>
    </details>
  );
}
