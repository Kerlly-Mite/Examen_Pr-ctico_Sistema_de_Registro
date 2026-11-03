import { useInstallPrompt } from './useInstallPrompt'

/** Botón "Instalar aplicación": solo aparece cuando el navegador permite instalar la PWA. */
export function InstallButton() {
  const { puedeInstalar, instalar } = useInstallPrompt()
  if (!puedeInstalar) return null

  return (
    <button
      type="button"
      onClick={instalar}
      className="rounded bg-oro-400 px-3 py-1 text-sm font-semibold text-biess-900 hover:bg-oro-500 focus:outline-none focus:ring-2 focus:ring-white"
    >
      <span className="hidden sm:inline">Instalar aplicación</span>
      <span className="sm:hidden" aria-label="Instalar aplicación">
        ⬇ Instalar
      </span>
    </button>
  )
}
