import { useSearchParams } from 'react-router-dom'

export function RestablecerPage() {
  const [params] = useSearchParams()
  const token = params.get('token')

  return (
    <section className="rounded-lg bg-white p-6 shadow">
      <h1 className="text-xl font-semibold">Restablecer contraseña</h1>
      {token ? (
        <p className="mt-2 text-slate-600">Pantalla pendiente (rama feature/frontend-auth).</p>
      ) : (
        <p className="mt-2 text-red-600">El enlace no es válido: falta el token.</p>
      )}
    </section>
  )
}