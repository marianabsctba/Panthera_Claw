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
}

export type Peso = "falso" | "alerta" | "ok" | "info";

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

export type Veredito = "falso" | "suspeito" | "sem_erros" | "incompleto";

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
  vistos?: Visto[];
}
