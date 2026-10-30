// Cliente do worker do núcleo: cada chamada vira uma promessa.
import type { Campos, Entrada, InfoArquivo, Resultado, ResultadoCodigo } from "./tipos";

const worker = new Worker(new URL("./nucleo.worker.ts", import.meta.url), { type: "module" });
let seq = 0;
const pendentes = new Map<number, { ok: (v: unknown) => void; falha: (e: Error) => void }>();

worker.onmessage = (e: MessageEvent<{ id: number; ok: boolean; valor?: unknown; erro?: string }>) => {
  const p = pendentes.get(e.data.id);
  if (!p) return;
  pendentes.delete(e.data.id);
  if (e.data.ok) p.ok(e.data.valor);
  else p.falha(new Error(e.data.erro ?? "erro"));
};

function chamar<T>(metodo: string, ...args: unknown[]): Promise<T> {
  const id = ++seq;
  return new Promise<T>((ok, falha) => {
    pendentes.set(id, { ok: ok as (v: unknown) => void, falha });
    worker.postMessage({ id, metodo, args });
  });
}

export const nucleo = {
  ler: async (texto: string) => JSON.parse(await chamar<string>("ler", texto)) as Campos,
  analisar: async (nome: string, dados: Uint8Array) => JSON.parse(await chamar<string>("analisar", nome, dados)) as InfoArquivo,
  conferir: async (e: Entrada) => JSON.parse(await chamar<string>("conferir", JSON.stringify(e))) as Resultado,
  conferirCodigo: async (texto: string) => JSON.parse(await chamar<string>("conferirCodigo", texto)) as ResultadoCodigo,
};
