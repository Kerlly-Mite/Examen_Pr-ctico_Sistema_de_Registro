import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { AuthContext } from './auth-context'
import type { AuthContextValue } from './auth-context'
import { api, guardarToken, obtenerToken, registrarManejadorNoAutorizado } from '../api/client'
import { RUTAS } from '../api/rutas'
import type { Postulante } from '../types/api'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => obtenerToken())
  const [usuario, setUsuario] = useState<Postulante | null>(null)
  const [cargando, setCargando] = useState<boolean>(() => obtenerToken() !== null)

  const cerrarSesion = useCallback(() => {
    guardarToken(null)
    setToken(null)
    setUsuario(null)
  }, [])

  // El login ya devuelve el postulante, no hace falta otra petición
  const iniciarSesion = useCallback((nuevoToken: string, postulante: Postulante) => {
    guardarToken(nuevoToken)
    setToken(nuevoToken)
    setUsuario(postulante)
  }, [])

  // Al recargar la página: si hay token guardado, se valida con GET /auth/me
  useEffect(() => {
    if (obtenerToken() === null) return
    api
      .get<Postulante>(RUTAS.me)
      .then(setUsuario)
      .catch(cerrarSesion)
      .finally(() => setCargando(false))
  }, [cerrarSesion])

  // Si cualquier petición responde "no_autorizado", se cierra la sesión
  useEffect(() => {
    registrarManejadorNoAutorizado(cerrarSesion)
    return () => registrarManejadorNoAutorizado(null)
  }, [cerrarSesion])

  const valor = useMemo<AuthContextValue>(
    () => ({
      token,
      usuario,
      cargando,
      estaAutenticado: token !== null,
      iniciarSesion,
      cerrarSesion,
    }),
    [token, usuario, cargando, iniciarSesion, cerrarSesion],
  )

  return <AuthContext.Provider value={valor}>{children}</AuthContext.Provider>
}