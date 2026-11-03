/// <reference lib="webworker" />
// src/sw.ts — Service worker de Postulación BIESS (Workbox, estrategia injectManifest).
//
// REGLAS DE SEGURIDAD (la app maneja datos personales, JWT y documentos):
//  1. Solo se guarda en caché el "app shell": HTML, JS, CSS e íconos generados en el build.
//  2. NUNCA se cachea nada de /api/*: login, registro, formulario, documentos y PDF pasan siempre
//     directo a la red. Como no hay ruta para /api, el service worker no las intercepta.
//  3. Las escrituras (POST/PUT/DELETE) no se interceptan ni se encolan: sin conexión fallan con aviso.
import { cleanupOutdatedCaches, createHandlerBoundToURL, matchPrecache, precacheAndRoute } from 'workbox-precaching'
import { NavigationRoute, registerRoute, setCatchHandler } from 'workbox-routing'

declare const self: ServiceWorkerGlobalScope

// 1) App shell: lo que genera el build (el plugin inyecta aquí la lista de archivos).
precacheAndRoute(self.__WB_MANIFEST)
cleanupOutdatedCaches()

// 2) Navegaciones de la SPA -> index.html precacheado (la app abre sin conexión).
//    Se excluye /api para no devolver HTML a una llamada de API.
registerRoute(new NavigationRoute(createHandlerBoundToURL('/index.html'), { denylist: [/^\/api\//] }))

// 3) Respaldo: si una navegación falla del todo, se muestra offline.html (precacheada).
setCatchHandler(async ({ request }) => {
  if (request.destination === 'document') {
    const offline = await matchPrecache('/offline.html')
    if (offline) return offline
  }
  return Response.error()
})

// 4) Actualización controlada: el service worker nuevo espera hasta que el usuario pulsa "Actualizar".
self.addEventListener('message', (event) => {
  if (event.data?.type === 'SKIP_WAITING') void self.skipWaiting()
})
self.addEventListener('activate', () => {
  void self.clients.claim()
})
