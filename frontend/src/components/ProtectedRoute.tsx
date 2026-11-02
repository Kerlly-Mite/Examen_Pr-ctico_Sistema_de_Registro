import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../context/useAuth'
import { Cargando } from './Cargando'

export function ProtectedRoute() {
  const { estaAutenticado, cargando } = useAuth()
  const ubicacion = useLocation()

  if (cargando) {
    return <Cargando texto="Cargando sesión..." />
  }

  if (!estaAutenticado) {
    return <Navigate to="/login" replace state={{ desde: ubicacion.pathname }} />
  }

  return <Outlet />
}