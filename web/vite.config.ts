import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// No GitHub Pages o site fica em /Panthera_Claw/.
export default defineConfig({
  base: "/Panthera_Claw/",
  plugins: [react()],
  worker: { format: "es" },
  build: { outDir: "dist", sourcemap: false, target: "es2022", chunkSizeWarningLimit: 2000 },
  optimizeDeps: { exclude: ["tesseract.js"] },
});
