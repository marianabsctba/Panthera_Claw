# Panthera Claw

Confira se um comprovante de Pix é falso em poucos segundos, sem ele sair do seu aparelho.

Comprovante falso é um dos golpes mais comuns do Brasil: a pessoa "paga", manda um print editado e leva o produto. O Panthera Claw lê o comprovante (PDF ou print) e compara o que está escrito com o que o **código da transação** revela. Feito para qualquer pessoa usar, inclusive idosos: letra grande, um passo por tela, resultado em uma frase e leitura em voz alta.

**Site:** https://marianabsctba.github.io/Panthera_Claw/

## Como funciona

Todo Pix tem um identificador fim a fim (EndToEndId) de 32 caracteres, com formato definido pelo Banco Central:

```
E 18236120 202610291732 k7Hq2Lm9Pzx
│ │        │            └ sequencial (11)
│ │        └ data e hora em UTC (AAAAMMDDHHMM)
│ └ ISPB da instituição que gerou o identificador
└ E para pagamento, D para devolução
```

Quem falsifica quase nunca acerta esses detalhes. O Panthera Claw confere:

| Conferência | Peso |
|---|---|
| Código com formato do Pix (tamanho, data que existe no calendário) | erro se falhar |
| Banco do código igual ao banco de quem pagou (lista oficial de participantes do Banco Central) | erro se falhar |
| Horário do código (UTC) igual ao horário impresso (Brasília); código montado com horário de Brasília | erro se falhar |
| Data do código no futuro ou antes de o Pix existir | erro |
| Chave Pix possível (dígitos verificadores de CPF e CNPJ, inclusive o CNPJ alfanumérico; DDD) | erro se falhar |
| Mesmo código em outro comprovante com outro valor, já conferido neste aparelho | erro |
| PDF com outro valor escondido por baixo do valor (valor antigo coberto e reescrito) | erro |
| PDF salvo de novo com o valor escrito na versão nova | erro |
| Arquivo que declara ter sido gerado ou editado por IA (selo C2PA, etiqueta IPTC `trainedAlgorithmicMedia`, prompt de Stable Diffusion ou ComfyUI no PNG, programa gerador) | erro |
| Comprovante de Pix agendado, pendente, cancelado ou devolvido | "ainda não é um pagamento" |
| Arquivo gravado por editor de documento ou imagem; PDF alterado depois de criado; PDF que é só imagem; texto escondido por baixo de outro | atenção |

**Resultado:** "falso" quando há erro de coerência; "suspeito" quando só há sinais de atenção; "não encontramos erros" quando passa em tudo. Nunca "verdadeiro": um falsificador cuidadoso pode acertar tudo, e só o extrato confirma que o dinheiro entrou.

Leitura do arquivo:

- **PDF com texto:** lido direto (exato), sem etapa de confirmação.
- **Print ou PDF que é só imagem:** leitura de imagem local (Tesseract.js, modelo em português servido pelo próprio site). A pessoa confere os dados lidos ao lado da imagem antes do resultado. Na parte numérica do código, letras que a leitura confunde com dígitos (O e 0, I e 1, S e 5) são corrigidas pelo formato.

## Boleto e QR Code de Pix

- **Boleto:** cole ou digite a linha digitável (47 dígitos no bancário, 48 em contas e tributos) ou o código de barras. Confere os dígitos verificadores de cada bloco e o geral, mostra o banco que vai receber (pelo número do banco na lista do Banco Central), o valor e o vencimento (com o fator de vencimento reiniciado em 22/02/2025). Linha alterada sem recalcular os dígitos dá "não pague".
- **Pix copia e cola / QR Code:** escolha a imagem do QR Code (lida no aparelho) ou cole o código. Confere a soma de verificação (CRC16) e mostra quem vai receber, cidade, chave e valor. Golpe do QR trocado: o código pode estar inteiro e mesmo assim levar o dinheiro para outra pessoa, por isso o resultado pede para conferir o nome.

## Privacidade

Nada sai do aparelho. A política de segurança do site proíbe conexão com qualquer outro endereço. Para avisar de código repetido, o navegador guarda localmente só o código, o valor e a data de cada conferência (dá para apagar na página "Como funciona").

## Usar pelo WhatsApp

No Android, depois de instalar o site (botão "Instalar no celular"), o Panthera Claw aparece na lista de compartilhamento: abra o comprovante no WhatsApp, toque em Compartilhar e escolha Panthera Claw.

## Linha de comando

```bash
go install github.com/marianabsctba/Panthera_Claw/cmd/claw@latest
```

```bash
claw codigo E18236120202610291732k7Hq2Lm9Pzx
claw conferir -id E18236120202610291732k7Hq2Lm9Pzx -data "29/10/2026 14:32" -banco Nubank
claw conferir -texto comprovante.txt comprovante.pdf
claw conferir -json -texto comprovante.txt
claw boleto 00190.50095 40144.816069 06809.350314 3 37370000000100
claw pix "00020126580014br.gov.bcb.pix..."
```

Saída 0 sem erros, 1 falso ou suspeito, 2 erro de uso.

## Desenvolvimento

Requisitos: Go 1.27 e Node 22.

```bash
go test ./...
```

```bash
cd web && npm ci && npm run dev
```

Atualizar a lista de instituições (baixe as listas mais recentes do Banco Central para `dados/`):

```bash
go run ./ferramentas/bancos
```

Veja a [arquitetura](docs/ARQUITETURA.md).

## Limites conhecidos

- **Valor editado num print.** Se alguém faz um Pix verdadeiro de valor baixo e edita só o valor no print, o código, o banco e o horário continuam certos. Testamos a análise do nível de erro de compressão (ELA) palavra a palavra: depois que o WhatsApp reduz e recomprime a imagem, ela não separou prints editados de verdadeiros e ainda acusou um verdadeiro em modo escuro. Por isso ficou de fora. No PDF a edição é detectável (texto por baixo, versão salva por cima), e é lá que ela é conferida.
- **IA sem marca.** Só conta o que o arquivo declara. O WhatsApp apaga metadados ao recomprimir, então uma imagem de IA enviada como foto chega sem marca. Não usamos classificador de "parece IA": erra demais para um resultado que diz "falso".

- Os bancos não seguem um modelo único de comprovante. A leitura procura rótulos comuns ("Instituição", "Origem", "Destino", "Chave Pix"); campos não achados ficam para a pessoa completar.
- Na leitura de imagem, os 11 últimos caracteres do código (sequencial) misturam letras e números sem padrão, e a leitura pode trocar letras parecidas (q e g, Z e 2) mesmo com o tamanho certo. Testamos reler só a linha do código, ampliada e restrita a letras e números: não acertou nenhum dos casos errados, então ficou de fora. A tela mostra o final em destaque e pede para conferir. As conferências de banco e horário usam só os 21 primeiros, que o formato confere.
- Modelos de layout por banco ainda não existem: dependem de comprovantes reais para calibrar.

## Licença

[GPLv3 ou posterior](LICENSE). As listas em `dados/` são públicas, do Banco Central do Brasil.
