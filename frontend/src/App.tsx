import { useEffect, useState } from 'react'
import { http } from './services/http'
import type { Pais } from './types/catalogo'

type Estado =
  | { fase: 'cargando' }
  | { fase: 'ok'; total: number }
  | { fase: 'error'; mensaje: string }

// Pantalla temporal de la Fase 0-1: comprueba que React, Tailwind, el proxy y la API funcionan.
// Se reemplaza por el enrutador y las pantallas reales en las Fases 6-7.
export default function App() {
  const [estado, setEstado] = useState<Estado>({ fase: 'cargando' })

  useEffect(() => {
    http<{ data: Pais[] }>('/catalogos/paises')
      .then((r) => setEstado({ fase: 'ok', total: r.data.length }))
      .catch((e: Error) => setEstado({ fase: 'error', mensaje: e.message }))
  }, [])

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-50 p-4">
      <section className="w-full max-w-md rounded-2xl bg-white p-6 shadow-md">
        <h1 className="text-xl font-semibold text-slate-800">Postulación Directorio BIESS</h1>
        <p className="mt-1 text-sm text-slate-500">Verificación de conexión con la API</p>

        <div className="mt-6" aria-live="polite">
          {estado.fase === 'cargando' && (
            <p className="flex items-center gap-2 text-slate-600">
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-blue-600" />
              Conectando...
            </p>
          )}
          {estado.fase === 'ok' && (
            <p className="rounded-lg bg-green-50 p-3 text-green-700">
              API conectada: {estado.total} países cargados desde PostgreSQL.
            </p>
          )}
          {estado.fase === 'error' && (
            <p className="rounded-lg bg-red-50 p-3 text-red-700">{estado.mensaje}</p>
          )}
        </div>
      </section>
    </main>
  )
}
