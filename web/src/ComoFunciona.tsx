import { ArrowLeft } from "lucide-react";
import { apagar } from "./historico";

export default function ComoFunciona({ voltar }: { voltar: () => void }) {
  return (
    <section className="leitura">
      <h1>Como funciona</h1>

      <h2>O segredo está no código da transação</h2>
      <p>
        Todo Pix tem um código de 32 caracteres, que começa com a letra E. Dentro dele ficam escondidos o banco de onde
        o dinheiro saiu e a data e a hora exatas. Quem falsifica um comprovante quase nunca acerta esses detalhes.
      </p>
      <p>
        O Panthera Claw compara o que está escrito no comprovante com o que o código revela. Se não bater, o
        comprovante é falso.
      </p>
      <p>
        Também olhamos o próprio arquivo: se passou por um programa de edição, se o PDF tem outro valor escondido
        por baixo do valor que aparece, e se o arquivo declara que foi feito ou mexido por inteligência artificial.
      </p>

      <h2>O que a gente não consegue ver</h2>
      <p>
        Se alguém faz um Pix verdadeiro de valor pequeno e muda só o valor no print, o resto do comprovante continua
        certo. Por isso, antes de entregar, confira sempre no extrato do seu banco.
      </p>

      <h2>Agendamento não é pagamento</h2>
      <p>
        Um golpe comum: a pessoa agenda um Pix para amanhã, manda o comprovante do agendamento e cancela depois que
        leva o produto. Quando o comprovante diz agendado, pendente, cancelado ou devolvido, avisamos que aquilo ainda
        não é um pagamento.
      </p>

      <h2>Boleto e QR Code</h2>
      <p>
        Na tela inicial, em Outras conferências, dá para conferir a linha de um boleto e o QR Code (ou o código copia e
        cola) de um Pix antes de pagar. Mostramos para quem o dinheiro vai. Se não for quem você espera, não pague.
      </p>

      <h2>O que o resultado quer dizer</h2>
      <ul className="lista-grande">
        <li><strong>Falso:</strong> encontramos um erro que um comprovante verdadeiro não teria.</li>
        <li><strong>Ainda não é um pagamento:</strong> o Pix foi só agendado, está pendente ou foi cancelado.</li>
        <li><strong>Suspeito:</strong> há sinais estranhos. Pode ter explicação, então confira no extrato.</li>
        <li><strong>Não encontramos erros:</strong> o comprovante passou nas conferências. Mas um falsificador cuidadoso
          pode acertar tudo. Só o extrato do seu banco confirma que o dinheiro entrou.</li>
      </ul>

      <h2>Usar direto pelo WhatsApp</h2>
      <ol className="lista-grande">
        <li>Abra este site no celular e toque em <strong>Instalar no celular</strong> (ou, no menu do navegador, em
          <strong> Adicionar à tela inicial</strong>).</li>
        <li>No WhatsApp, abra o comprovante que você recebeu.</li>
        <li>Toque em <strong>Compartilhar</strong> e escolha <strong>Panthera Claw</strong>.</li>
      </ol>
      <p className="suave">No iPhone, salve a imagem e escolha o arquivo pelo botão da tela inicial.</p>

      <h2>Sua privacidade</h2>
      <p>
        O comprovante é lido e conferido dentro do seu aparelho. Nada é enviado para a internet. Para avisar quando um
        mesmo código aparecer em comprovantes diferentes, guardamos neste aparelho só o código, o valor e a data das
        conferências.
      </p>
      <p><button className="botao-secundario" onClick={() => { apagar(); alert("Histórico apagado."); }}>Apagar o histórico deste aparelho</button></p>

      <button className="botao-gigante" onClick={voltar}><ArrowLeft size={30} aria-hidden /> Voltar</button>
    </section>
  );
}
