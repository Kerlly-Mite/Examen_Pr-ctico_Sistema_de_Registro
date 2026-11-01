import { Link, Outlet } from 'react-router-dom'
import { useAuth } from '../context/useAuth'
import { EstadoApi } from './EstadoApi'

export function Layout() {
  const { estaAutenticado, usuario, cerrarSesion } = useAuth()

  return (
    <div className="flex min-h-screen flex-col bg-slate-100 text-slate-800">
      <header className="bg-blue-900 text-white">
        <div className="mx-auto flex max-w-4xl items-center justify-between px-4 py-3">
          <Link to="/" className="font-semibold">
            Postulación Directorio BIESS
          </Link>
          {estaAutenticado && (
            <div className="flex items-center gap-3 text-sm">
              {usuario && <span>{usuario.nombres}</span>}
              <button
                type="button"
                onClick={cerrarSesion}
                className="rounded bg-white/20 px-3 py-1 hover:bg-white/30"
              >
                Cerrar sesión
              </button>
            </div>
          )}
        </div>
      </header>

      <main className="mx-auto w-full max-w-4xl flex-1 px-4 py-8">
        <Outlet />
      </main>

      <footer className="px-4 py-3 text-center text-sm">
        <EstadoApi />
      </footer>
    </div>
  )
}