// Worker que carrega o núcleo em Go (WebAssembly), para a tela não travar.
import "./wasm_exec.js";

declare const Go: new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): Promise<void> };
type Api = Record<string, (...a: unknown[]) => unknown>;

const pronto: Promise<Api> = (async () => {
  const go = new Go();
  const url = new URL(`${import.meta.env.BASE_URL}claw.wasm`, self.location.origin);
  const { instance } = await WebAssembly.instantiateStreaming(fetch(url), go.importObject);
  void go.run(instance);
  return (self as unknown as { pantheraClaw: Api }).pantheraClaw;
})();

self.onmessage = async (e: MessageEvent<{ id: number; metodo: string; args: unknown[] }>) => {
  const { id, metodo, args } = e.data;
  try {
    const api = await pronto;
    const valor = api[metodo](...args);
    if (valor instanceof Error) throw valor;
    (self as unknown as Worker).postMessage({ id, ok: true, valor });
  } catch (err) {
    (self as unknown as Worker).postMessage({ id, ok: false, erro: err instanceof Error ? err.message : String(err) });
  }
};
