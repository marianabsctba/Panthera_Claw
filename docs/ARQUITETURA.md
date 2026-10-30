# Arquitetura

Um núcleo em Go, duas caras: site estático no GitHub Pages (núcleo compilado para WebAssembly) e linha de comando nativa. Tudo roda no aparelho de quem confere.

```
print/PDF ──> leitura (pdf.js ou Tesseract.js) ──> campos ──> confirmação ──> núcleo (WASM) ──> resultado
                                                     ▲                            │
                                    nucleo/leitura ──┘           lista do BCB, histórico local
```

## Núcleo (`nucleo/`)

| Pacote | Papel |
|---|---|
| `bancos` | Lista de participantes do Pix e do STR (gerada por `ferramentas/bancos`, embutida no binário) e marcas: liga "Nubank" aos ISPBs da Nu Pagamentos e da Nu Financeira. |
| `pix` | Lê o EndToEndId (com correção de confusões de leitura só na parte numérica) e classifica chaves Pix (CPF, CNPJ inclusive alfanumérico, telefone, e-mail, aleatória). |
| `leitura` | Tira os campos do texto: rótulos na mesma linha ou na seguinte, seções de origem e destino, datas numéricas e por extenso, correção do @ mal lido. |
| `arquivo` | Metadados do arquivo: SHA-256, programas declarados (PDF Info, XMP, EXIF, PNG), versões gravadas no PDF e o texto escrito em cada versão posterior (fluxos Flate e ASCII85), assinatura digital, marcas de IA (C2PA, IPTC, prompt no PNG). |
| `boleto` | Linha digitável e código de barras, bancário e de arrecadação: dígitos verificadores (módulo 10 e 11), banco, valor e vencimento. |
| `brcode` | Pix copia e cola no padrão EMV: campos, chave, recebedor, cobrança dinâmica e CRC16/CCITT-FALSE. |
| `conferir` | Junta tudo e devolve achados com peso (`falso`, `nao_pago`, `alerta`, `ok`, `info`) e o veredito; para boleto e copia e cola, `errado`, `suspeito` ou `confira`. |

Decisões:

- **Horário de Brasília fixo em UTC−3.** Não há horário de verão desde 2019, antes do Pix, então o deslocamento é exato para qualquer comprovante.
- **Tolerância de horário:** até 5 minutos é coerente (o código guarda o minuto em que a transação começou). Diferença de 3 horas exatas é o erro típico de quem monta o código com o horário de Brasília. 1 ou 2 horas exatas podem ser outro fuso do Brasil (atenção). Até 2 horas é atenção; mais que isso é erro.
- **Chave com forma estranha não é erro**, porque pode ser leitura ruim da imagem. Dígito verificador ou DDD impossível é erro.
- **Texto sobreposto no PDF:** o site compara as caixas dos textos que o pdf.js extrai; dois textos diferentes ocupando mais de 40% da mesma área indicam algo escrito por cima. Se os dois são valores em reais e diferentes, é erro.
- **Versão nova do PDF:** só conta como nova a frase que não existia na primeira versão, porque alguns programas regravam a página inteira ao salvar por cima.
- **Marca de IA é erro, ausência de marca não diz nada.** Comprovante de banco nunca é gerado nem editado por IA.
- **Sem análise de pixels no print.** A análise do nível de erro (ELA) foi medida em prints recomprimidos como no WhatsApp e não separou editados de verdadeiros; foi removida.
- **Situação do Pix por palavras do comprovante** (agendado, pendente, cancelado, devolvido): gera o veredito "ainda não é um pagamento", abaixo de "falso" e acima de "suspeito".
- **QR Code de Pix numa imagem** vai direto para a conferência do copia e cola (lido no aparelho com jsQR): comprovante não traz QR de cobrança.
- **Mesmo código com o mesmo valor em outro arquivo** é o mesmo Pix enviado como PDF e como print: informação, não fraude.

## Site (`web/`)

React + Vite, sem servidor. O núcleo roda num Web Worker. A leitura de imagem (Tesseract.js, núcleo LSTM e modelo `por` best_int) e o pdf.js são servidos pelo próprio site; `scripts/nucleo.mjs` compila o WASM e copia esses arquivos para `public/`.

- PDF com pelo menos 20 caracteres de texto é lido direto. Os pedaços de texto são agrupados por altura em linhas; vão grande vira dois espaços (coluna).
- Imagem é ampliada para cerca de 1600 px de largura, convertida para cinza e invertida se o fundo for escuro.
- O service worker guarda o site para uso sem internet e recebe arquivos do compartilhamento do Android (`share_target`).

Acessibilidade: letra base de 20 px com ajuste em quatro tamanhos, alvos de toque de pelo menos 48 px, um passo por tela, resultado com ícone, cor e texto (nunca só cor), leitura em voz alta pela voz pt-BR do aparelho.

## Próximas fases

3. Relatório para B.O. e processo com resumo do arquivo e carimbo de tempo; conferência em lote na linha de comando.
4. Modelos de layout por banco, abertos para contribuição.
