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
| `arquivo` | Metadados do arquivo: SHA-256, programas declarados (PDF Info, XMP, EXIF, PNG), versões gravadas no PDF, assinatura digital. |
| `conferir` | Junta tudo e devolve achados com peso (`falso`, `alerta`, `ok`, `info`) e o veredito. |

Decisões:

- **Horário de Brasília fixo em UTC−3.** Não há horário de verão desde 2019, antes do Pix, então o deslocamento é exato para qualquer comprovante.
- **Tolerância de horário:** até 5 minutos é coerente (o código guarda o minuto em que a transação começou). Diferença de 3 horas exatas é o erro típico de quem monta o código com o horário de Brasília. 1 ou 2 horas exatas podem ser outro fuso do Brasil (atenção). Até 2 horas é atenção; mais que isso é erro.
- **Chave com forma estranha não é erro**, porque pode ser leitura ruim da imagem. Dígito verificador ou DDD impossível é erro.
- **Mesmo código com o mesmo valor em outro arquivo** é o mesmo Pix enviado como PDF e como print: informação, não fraude.

## Site (`web/`)

React + Vite, sem servidor. O núcleo roda num Web Worker. A leitura de imagem (Tesseract.js, núcleo LSTM e modelo `por` best_int) e o pdf.js são servidos pelo próprio site; `scripts/nucleo.mjs` compila o WASM e copia esses arquivos para `public/`.

- PDF com pelo menos 20 caracteres de texto é lido direto. Os pedaços de texto são agrupados por altura em linhas; vão grande vira dois espaços (coluna).
- Imagem é ampliada para cerca de 1600 px de largura, convertida para cinza e invertida se o fundo for escuro.
- O service worker guarda o site para uso sem internet e recebe arquivos do compartilhamento do Android (`share_target`).

Acessibilidade: letra base de 20 px com ajuste em quatro tamanhos, alvos de toque de pelo menos 48 px, um passo por tela, resultado com ícone, cor e texto (nunca só cor), leitura em voz alta pela voz pt-BR do aparelho.

## Próximas fases

2. Boleto (linha digitável contra código de barras, dígitos verificadores, fator de vencimento reiniciado em 22/02/2025) e Pix copia e cola (CRC16 e recebedor).
3. Relatório para B.O. e processo com resumo do arquivo e carimbo de tempo; conferência em lote na linha de comando.
4. Modelos de layout por banco, abertos para contribuição.
