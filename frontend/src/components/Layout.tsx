import { Link, Outlet } from 'react-router-dom'
import { useAuth } from '../context/useAuth'
import { EstadoApi } from './EstadoApi'

export function Layout() {
  const { estaAutenticado, usuario, cerrarSesion } = useAuth()

  return (
    <div className="flex min-h-screen flex-col bg-slate-50 text-slate-800">
      <header className="bg-biess-800 text-white print:hidden">
        <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3">
          <Link to="/" className="text-sm font-semibold tracking-wide">
            IESS · BIESS
          </Link>
          <div className="flex items-center gap-4 text-sm">
            <span className="hidden text-white/90 md:inline">Concurso de méritos y oposición</span>
            {estaAutenticado && (
              <div className="flex items-center gap-3 border-l border-white/30 pl-4">
                {usuario && <span className="hidden sm:inline">{usuario.nombres}</span>}
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
        </div>
      </header>

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8">
        <Outlet />
      </main>

      <footer className="px-4 py-3 text-center text-xs print:hidden">
        <EstadoApi />
      </footer>
    </div>
  )
}