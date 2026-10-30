// Mostra o que foi lido da imagem ao lado do comprovante, para a pessoa
// confirmar ou corrigir antes do resultado. Erro de leitura fica visível,
// nunca escondido dentro da conclusão.
import { useState } from "react";
import { ArrowLeft, Check, ZoomIn } from "lucide-react";
import type { Leitura } from "./leitor";
import type { Campos } from "./tipos";

interface Props {
  leitura: Leitura;
  campos: Campos;
  voltar: () => void;
  confirmar: (c: Campos) => void;
}

export default function Conferencia({ leitura, campos, voltar, confirmar }: Props) {
  const [c, setC] = useState<Campos>(campos);
  const [ampliada, setAmpliada] = useState(false);
  const id = c.id.replace(/\s/g, "");
  const muda = (k: keyof Campos) => (e: React.ChangeEvent<HTMLInputElement>) => setC({ ...c, [k]: e.target.value });

  return (
    <section className="conferencia">
      <h1>Confira o que a gente leu</h1>
      <p className="grande">Compare com a imagem. Se algo estiver diferente, é só corrigir.</p>
      {leitura.confianca < 70 && leitura.origem === "imagem" && (
        <p className="aviso-leve">A imagem não está muito nítida. Olhe com atenção, principalmente o código da transação.</p>
      )}

      <button className={`previa ${ampliada ? "ampliada" : ""}`} onClick={() => setAmpliada(!ampliada)} aria-label={ampliada ? "Diminuir a imagem" : "Aumentar a imagem"}>
        <img src={leitura.previa} alt="Imagem do comprovante" />
        {!ampliada && <span className="lupa"><ZoomIn size={22} aria-hidden /> Toque para aumentar</span>}
      </button>

      <form className="campos" onSubmit={(e) => { e.preventDefault(); confirmar({ ...c, id }); }}>
        <label>
          <span>Código da transação (ID)</span>
          <input className="codigo" value={c.id} onChange={muda("id")} spellCheck={false} autoCapitalize="off" autoComplete="off"
            placeholder="Começa com a letra E" />
          <small className={id.length === 32 && !campos.idDuvidoso ? "certo" : id.length ? "errado" : ""}>
            {!id.length
              ? "Não encontramos o código. Se ele aparecer na imagem, digite aqui."
              : id.length !== 32
                ? `${id.length} de 32 caracteres. Confira letra por letra.`
                : campos.idDuvidoso && c.id === campos.id
                  ? "O final do código pode ter sido lido errado. Compare as últimas letras com a imagem."
                  : "32 de 32 caracteres"}
          </small>
        </label>
        <label>
          <span>Valor</span>
          <input value={c.valor} onChange={muda("valor")} inputMode="decimal" placeholder="R$ 0,00" />
        </label>
        <label>
          <span>Data e hora</span>
          <input value={c.dataHora} onChange={muda("dataHora")} placeholder="dia/mês/ano hora:minuto" />
        </label>
        <label>
          <span>Banco de quem pagou</span>
          <input value={c.bancoPagador} onChange={muda("bancoPagador")} placeholder={c.emissor || "Nome do banco"} />
        </label>
        <label>
          <span>Chave Pix de quem recebeu</span>
          <input value={c.chave} onChange={muda("chave")} spellCheck={false} autoCapitalize="off" placeholder="Opcional" />
        </label>

        <button type="submit" className="botao-gigante"><Check size={32} aria-hidden /> Está certo, conferir</button>
        <button type="button" className="botao-secundario" onClick={voltar}><ArrowLeft size={22} aria-hidden /> Escolher outro comprovante</button>
      </form>
    </section>
  );
}
