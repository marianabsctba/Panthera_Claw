// Funciona sem internet depois da primeira visita e recebe o comprovante
// compartilhado pelo WhatsApp (ou outro aplicativo) no celular.
const CACHE = "claw-v1";
const COMPARTILHADO = "claw-compartilhado";

self.addEventListener("install", () => self.skipWaiting());
self.addEventListener("activate", (e) => {
  e.waitUntil(
    caches.keys()
      .then((ks) => Promise.all(ks.filter((k) => k !== CACHE && k !== COMPARTILHADO).map((k) => caches.delete(k))))
      .then(() => self.clients.claim()),
  );
});

self.addEventListener("fetch", (e) => {
  const url = new URL(e.request.url);
  if (url.origin !== location.origin) return;

  if (e.request.method === "POST" && url.pathname.endsWith("/compartilhar")) {
    e.respondWith(
      (async () => {
        const dados = await e.request.formData();
        const arquivo = dados.get("arquivo");
        if (arquivo instanceof File) {
          const c = await caches.open(COMPARTILHADO);
          await c.put("arquivo", new Response(arquivo, {
            headers: { "Content-Type": arquivo.type || "application/octet-stream", "X-Nome": encodeURIComponent(arquivo.name) },
          }));
        }
        return Response.redirect(new URL("./?compartilhado=1", self.registration.scope).href, 303);
      })(),
    );
    return;
  }
  if (e.request.method !== "GET") return;

  // Arquivos com nome fixo (páginas, núcleo): rede primeiro, para pegar a
  // versão nova; cache quando estiver sem internet. Arquivos com resumo no
  // nome e os da leitura de imagem não mudam: cache primeiro.
  const fixo = !url.pathname.includes("/assets/") && !url.pathname.includes("/ocr/");
  const guardar = (r) => {
    if (r.ok) { const c = r.clone(); caches.open(CACHE).then((k) => k.put(e.request, c)); }
    return r;
  };
  if (fixo) {
    e.respondWith(
      fetch(e.request).then(guardar).catch(() =>
        caches.match(e.request).then((r) => r || (e.request.mode === "navigate" ? caches.match(new URL("./", self.registration.scope).href) : Response.error())),
      ),
    );
    return;
  }
  e.respondWith(caches.match(e.request).then((achado) => achado || fetch(e.request).then(guardar)));
});
