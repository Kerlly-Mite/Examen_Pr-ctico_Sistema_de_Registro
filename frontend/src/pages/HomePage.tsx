import { Link, Navigate } from 'react-router-dom'
import { useAuth } from '../context/useAuth'

const SECTORES = [
  { titulo: 'Afiliados', texto: 'Se valida tu afiliación con el Historial Laboral.' },
  { titulo: 'Jubilados', texto: 'Se valida que acredites jubilación.' },
  { titulo: 'Empleadores', texto: 'Se valida tu calidad de empleador.' },
]

export function HomePage() {
  const { estaAutenticado, cargando } = useAuth()

  if (cargando) return <p className="text-center text-slate-600">Cargando...</p>
  if (estaAutenticado) return <Navigate to="/formulario" replace />

  return (
    <section className="space-y-8 py-4">
      <div className="max-w-2xl space-y-4">
        <h1 className="text-3xl font-bold text-biess-800 sm:text-4xl">
          Directorio del BIESS: postulación en línea
        </h1>
        <p className="text-slate-600">
          Postula como miembro principal y alterno en representación de afiliados, jubilados o empleadores, según
          la Resolución C.D. 701.
        </p>
        <div className="flex flex-wrap gap-3">
          <Link
            to="/registro"
            className="rounded bg-oro-500 px-5 py-2 font-medium text-slate-900 hover:bg-oro-600"
          >
            Registrarse
          </Link>
          <Link
            to="/login"
            className="rounded bg-biess-800 px-5 py-2 font-medium text-white hover:bg-biess-900"
          >
            Iniciar sesión
          </Link>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {SECTORES.map((s) => (
          <div key={s.titulo} className="rounded-lg border border-slate-200 bg-white p-4">
            <h2 className="font-semibold text-biess-800">{s.titulo}</h2>
            <p className="mt-1 text-sm text-slate-600">{s.texto}</p>
          </div>
        ))}
      </div>
    </section>
  )
}