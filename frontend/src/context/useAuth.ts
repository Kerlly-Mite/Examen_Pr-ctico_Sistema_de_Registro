import { useContext } from 'react'
import { AuthContext } from './auth-context'

export function useAuth() {
  const contexto = useContext(AuthContext)
  if (contexto === null) {
    throw new Error('useAuth debe usarse dentro de <AuthProvider>')
  }
  return contexto
}