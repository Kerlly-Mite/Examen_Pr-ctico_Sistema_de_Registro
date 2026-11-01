import { createContext } from 'react'
import type { Postulante } from '../types/api'

export interface AuthContextValue {
  token: string | null
  usuario: Postulante | null
  cargando: boolean
  estaAutenticado: boolean
  iniciarSesion: (token: string, postulante: Postulante) => void
  cerrarSesion: () => void
}

export const AuthContext = createContext<AuthContextValue | null>(null)