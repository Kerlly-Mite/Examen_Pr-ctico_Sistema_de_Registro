import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import {
  enviarPostulacion,
  guardarPostulacion,
  obtenerOpciones,
  obtenerPostulacion,
} from '../api/postulacion'
import { ApiError } from '../api/client'
import { aReq } from '../lib/postulacion'
import type { Bloqueo, OpcionesPostulacion, Postulacion, PostulacionReq } from '../types/postulacion'
import { FormularioContext } from './formulario-context'
import type { FormularioContextValue, ResultadoAccion } from './formulario-context'

export function FormularioProvider({ children }: { children: ReactNode }) {
  const [postulacion, setPostulacion] = useState<Postulacion | null>(null)
  const [opciones, setOpciones] = useState<OpcionesPostulacion | null>(null)
  const [cargando, setCargando] = useState(true)
  const [errorCarga, setErrorCarga] = useState<string | null>(null)
  const [guardando, setGuardando] = useState(false)
  const [bloqueoLocal, setBloqueoLocal] = useState<Bloqueo | null>(null)
  const [intento, setIntento] = useState(0)
  // Copia síncrona del estado: permite encadenar varios cambios y luego guardar sin datos viejos
  const actual = useRef<Postulacion | null>(null)

  const aplicar = useCallback((p: Postulacion | null) => {
    actual.current = p
    setPostulacion(p)
  }, [])

  useEffect(() => {
    let activo = true
    Promise.all([obtenerOpciones(), obtenerPostulacion()])
      .then(([o, p]) => {
        if (!activo) return
        setOpciones(o)
        aplicar(p)
        setErrorCarga(null)
        setBloqueoLocal(null)
        setCargando(false)
      })
      .catch((e: unknown) => {
        if (!activo) return
        setErrorCarga(e instanceof ApiError ? e.message : 'No se pudo cargar el formulario.')
        setCargando(false)
      })
    return () => {
      activo = false
    }
  }, [intento, aplicar])

  const recargar = useCallback(() => {
    setCargando(true)
    setIntento((n) => n + 1)
  }, [])

  const actualizar = useCallback(
    (parcial: Partial<PostulacionReq>) => {
      const p = actual.current
      if (p !== null) aplicar({ ...p, ...parcial })
    },
    [aplicar],
  )

  const resultadoError = useCallback((e: unknown): ResultadoAccion => {
    if (e instanceof ApiError) {
      if (e.codigo === 'nacionalidad_no_permitida' || e.codigo === 'cargo_no_habilitado') {
        setBloqueoLocal({ codigo: e.codigo, mensaje: e.message })
      }
      return { ok: false, mensaje: e.message, codigo: e.codigo }
    }
    return { ok: false, mensaje: 'Ocurrió un error inesperado.', codigo: 'desconocido' }
  }, [])

  const guardar = useCallback(async (): Promise<ResultadoAccion> => {
    const p = actual.current
    if (p === null) return { ok: false, mensaje: 'El formulario no está cargado.', codigo: 'sin_datos' }
    setGuardando(true)
    try {
      const guardada = await guardarPostulacion(aReq(p))
      aplicar(guardada)
      setBloqueoLocal(null)
      return { ok: true }
    } catch (e) {
      return resultadoError(e)
    } finally {
      setGuardando(false)
    }
  }, [aplicar, resultadoError])

  const enviar = useCallback(
    async (consentimiento: boolean): Promise<ResultadoAccion> => {
      setGuardando(true)
      try {
        await enviarPostulacion(consentimiento)
        aplicar(await obtenerPostulacion())
        return { ok: true }
      } catch (e) {
        return resultadoError(e)
      } finally {
        setGuardando(false)
      }
    },
    [aplicar, resultadoError],
  )

  const bloquear = useCallback((b: Bloqueo) => setBloqueoLocal(b), [])

  const valor = useMemo<FormularioContextValue>(
    () => ({
      cargando,
      errorCarga,
      postulacion,
      opciones,
      bloqueo: bloqueoLocal ?? postulacion?.bloqueo ?? null,
      enviada: postulacion?.estado === 'ENVIADA',
      guardando,
      recargar,
      actualizar,
      guardar,
      enviar,
      bloquear,
    }),
    [cargando, errorCarga, postulacion, opciones, bloqueoLocal, guardando, recargar, actualizar, guardar, enviar, bloquear],
  )

  return <FormularioContext.Provider value={valor}>{children}</FormularioContext.Provider>
}