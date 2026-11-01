import { Navigate } from 'react-router-dom'
import { useAuth } from '../context/useAuth'

export function RedirigirInicio() {
  const { estaAutenticado, cargando } = useAuth()
  if (cargando) {
    return <p className="text-center text-slate-600">Cargando...</p>
  }
  return <Navigate to={estaAutenticado ? '/formulario' : '/login'} replace />
}