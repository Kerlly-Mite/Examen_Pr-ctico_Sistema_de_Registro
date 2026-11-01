import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../context/useAuth'

export function ProtectedRoute() {
  const { estaAutenticado, cargando } = useAuth()
  const ubicacion = useLocation()

  if (cargando) {
    return <p className="p-6 text-center text-slate-600">Cargando sesión...</p>
  }

  if (!estaAutenticado) {
    return <Navigate to="/login" replace state={{ desde: ubicacion.pathname }} />
  }

  return <Outlet />
}