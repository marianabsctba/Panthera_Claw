// Compila o núcleo em Go para WebAssembly, copia o carregador do Go e os
// arquivos da leitura de imagem, que são servidos pelo próprio site (nada
// vem de outro endereço).
import { execFileSync } from "node:child_process";
import { copyFileSync, cpSync, mkdirSync } from "node:fs";
import { join } from "node:path";

const raiz = join(import.meta.dirname, "..", "..");
const web = join(raiz, "web");
const env = { ...process.env, GOOS: "js", GOARCH: "wasm" };
mkdirSync(join(web, "public", "ocr"), { recursive: true });
execFileSync(
  "go",
  ["build", "-trimpath", "-ldflags=-s -w", "-o", join(web, "public", "claw.wasm"), "./cmd/claw-wasm"],
  { cwd: raiz, env, stdio: "inherit" },
);
const goroot = execFileSync("go", ["env", "GOROOT"], { encoding: "utf8" }).trim();
copyFileSync(join(goroot, "lib", "wasm", "wasm_exec.js"), join(web, "src", "wasm_exec.js"));

const nm = join(web, "node_modules");
const ocr = join(web, "public", "ocr");
copyFileSync(join(nm, "tesseract.js", "dist", "worker.min.js"), join(ocr, "worker.min.js"));
for (const f of ["tesseract-core-lstm.wasm.js", "tesseract-core-simd-lstm.wasm.js", "tesseract-core-relaxedsimd-lstm.wasm.js"]) {
  copyFileSync(join(nm, "tesseract.js-core", f), join(ocr, f));
}
copyFileSync(join(nm, "@tesseract.js-data", "por", "4.0.0_best_int", "por.traineddata.gz"), join(ocr, "por.traineddata.gz"));
// Fontes padrão do PDF (Helvetica e afins), para desenhar comprovantes que
// não embutem a fonte.
cpSync(join(nm, "pdfjs-dist", "standard_fonts"), join(web, "public", "pdf", "fontes"), { recursive: true });
console.log("núcleo compilado");
