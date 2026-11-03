# PWA — Postulación Directorio BIESS

La aplicación es una **Progressive Web App**: se puede instalar en el celular, la tablet o el computador, abre
sin conexión (la interfaz) y avisa cuando hay una versión nueva.

## Qué incluye

| Pieza | Archivo | Para qué sirve |
|---|---|---|
| Manifiesto | `vite.config.ts` (genera `manifest.webmanifest`) | Nombre, colores, íconos y modo `standalone` para instalar |
| Íconos | `public/icons/*`, `public/favicon.svg` | 192 y 512 px, maskable y apple-touch |
| Service worker | `src/sw.ts` | Guarda el "app shell" y sirve `offline.html` si falla una navegación |
| Página sin conexión | `public/offline.html` | Respaldo cuando no hay internet |
| Registro y avisos | `src/pwa/PwaStatus.tsx` | Registra el service worker, avisa "Sin conexión" y "Nueva versión" |
| Botón instalar | `src/pwa/InstallButton.tsx`, `useInstallPrompt.ts` | Aparece solo cuando el navegador permite instalar |

## Qué se guarda en caché y qué NO

- **Sí:** HTML, JavaScript, CSS e íconos generados en el build (el "app shell"). Por eso la aplicación abre sin conexión.
- **No, nunca:** nada de `/api/*`. Login, registro, formulario, documentos PDF y el formulario descargable llevan datos
  personales y el JWT: pasan siempre directo a la red. El service worker ni siquiera tiene una ruta para `/api`.
- Las escrituras (guardar, enviar) no se encolan: sin conexión fallan con aviso, para no enviar una postulación
  sin que la persona lo sepa.

## Actualizaciones

`registerType: 'prompt'`: cuando hay una versión nueva se muestra "Hay una nueva versión disponible" con los botones
**Luego** y **Actualizar**. Así nadie pierde un formulario a medio llenar por un refresco automático.

## Cómo probarla

El service worker **no corre con `npm run dev`**. Se prueba con el build de producción:

```bash
cd frontend
npm run build
npm run preview        # http://localhost:5180 (el backend debe estar corriendo)
```

En Chrome, con la página abierta, pulsa F12 y revisa en **Application**:

1. **Manifest:** nombre, íconos y "Installable" sin errores.
2. **Service Workers:** estado *activated and is running*.
3. Marca **Offline**, recarga: la aplicación abre y aparece el aviso "Sin conexión".
4. Icono de instalar en la barra de direcciones (o el botón "Instalar aplicación" del encabezado).

Para una medición formal: **Lighthouse** (pestaña del mismo F12) con la categoría *Progressive Web App*.

## Requisito para instalar fuera de tu computador

Los navegadores solo permiten instalar y registrar service workers en **HTTPS** (o en `localhost`).
Por eso el despliegue en el VPS debe servirse con HTTPS.
