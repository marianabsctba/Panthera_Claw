// Espelho dos tipos JSON do núcleo em Go.

export interface Campos {
  id: string;
  valor: string;
  dataHora: string;
  bancoPagador: string;
  bancoRecebedor: string;
  nomePagador: string;
  nomeRecebedor: string;
  chave: string;
  emissor: string;
  idDuvidoso?: boolean;
  situacao?: string;
  situacaoTrecho?: string;
}

export interface InfoArquivo {
  nome: string;
  tipo: "pdf" | "jpeg" | "png" | "webp" | "outro";
  tamanho: number;
  sha256: string;
  programas?: string[];
  criado?: string;
  alterado?: string;
  salvamentos?: number;
  assinado?: boolean;
  editor?: string;
  alteracao?: { objetos: number; pagina: boolean; textos?: string[] };
  ia?: { tipo: "gerada" | "editada"; origem: string; gerador: string };
}

export type Peso = "falso" | "nao_pago" | "alerta" | "ok" | "info";

export interface Achado {
  codigo: string;
  peso: Peso;
  titulo: string;
  detalhe?: string;
}

export interface Codigo {
  texto: string;
  tipo: "E" | "D";
  ispb: string;
  instante: string;
  sequencial: string;
  correcoes?: string[];
  banco?: string;
  marca?: string;
  brasilia: string;
  utc: string;
  conhecido: boolean;
}

export type Veredito = "falso" | "nao_pago" | "suspeito" | "sem_erros" | "incompleto";

export interface Resultado {
  veredito: Veredito;
  achados: Achado[];
  codigo?: Codigo;
  campos: Campos;
  arquivo?: InfoArquivo;
}

export interface Visto {
  id: string;
  sha256: string;
  valor: string;
  quando: string;
}

export interface Entrada {
  campos: Campos;
  arquivo?: InfoArquivo;
  pdfSemTexto: boolean;
  sobrepostos?: { cima: string; baixo: string }[];
  vistos?: Visto[];
}

export interface ResultadoCodigo {
  tipo: "boleto" | "pix";
  veredito: "errado" | "suspeito" | "confira";
  resumo: { rotulo: string; valor: string }[];
  achados: Achado[];
  boleto?: { linha: string; barras: string; tipo: string; banco?: string; vencimento?: string; valor?: string; erros?: string[] };
  pix?: { texto: string; chave?: string; nome?: string; cidade?: string; valor?: string; url?: string; dominio?: string; txid?: string; crc: string; crcCerto: string; crcValido: boolean; dinamico: boolean };
}
