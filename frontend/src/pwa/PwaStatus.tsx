// Registra el service worker (solo en producción) y muestra los avisos "Sin conexión" y "Nueva versión".
import { useRegisterSW } from 'virtual:pwa-register/react'
import { useOnlineStatus } from './useOnlineStatus'

export function PwaStatus() {
  const online = useOnlineStatus()
  const {
    needRefresh: [hayActualizacion, setHayActualizacion],
    updateServiceWorker,
  } = useRegisterSW()

  if (online && !hayActualizacion) return null

  return (
    <div
      className="pointer-events-none fixed inset-x-0 bottom-0 z-[100] flex flex-col items-center gap-2 p-3 print:hidden"
      style={{ paddingBottom: 'max(0.75rem, env(safe-area-inset-bottom))' }}
    >
      {!online && (
        <div
          role="status"
          className="pointer-events-auto w-full max-w-md rounded-lg bg-slate-800 px-4 py-3 text-sm text-white shadow-lg"
        >
          📡 Sin conexión. Puede ver la aplicación, pero guardar y enviar su postulación requiere internet.
        </div>
      )}
      {hayActualizacion && (
        <div
          role="alert"
          className="pointer-events-auto flex w-full max-w-md items-center justify-between gap-3 rounded-lg border border-biess-100 bg-white px-4 py-3 text-sm text-slate-800 shadow-lg"
        >
          <span>Hay una nueva versión disponible.</span>
          <span className="flex shrink-0 gap-2">
            <button
              type="button"
              onClick={() => setHayActualizacion(false)}
              className="rounded px-3 py-2 text-slate-600 hover:bg-slate-100"
            >
              Luego
            </button>
            <button
              type="button"
              onClick={() => updateServiceWorker(true)}
              className="rounded bg-biess-800 px-3 py-2 font-semibold text-white hover:bg-biess-900"
            >
              Actualizar
            </button>
          </span>
        </div>
      )}
    </div>
  )
}
