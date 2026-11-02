import { useEffect, useRef, useState } from 'react'
import { ApiError } from '../api/client'

interface Resultado<T> {
  clave: string
  datos: T[]
  error: string | null
}

// Carga una lista cuando cambia la clave. Con clave null no pide nada.
export function useCatalogo<T>(clave: string | null, cargar: () => Promise<T[]>) {
  const [resultado, setResultado] = useState<Resultado<T> | null>(null)
  const [intento, setIntento] = useState(0)
  const cargarRef = useRef(cargar)

  useEffect(() => {
    cargarRef.current = cargar
  })

  useEffect(() => {
    if (clave === null) return
    const claveCompleta = `${clave}#${intento}`
    let activo = true
    cargarRef
      .current()
      .then((datos) => {
        if (activo) setResultado({ clave: claveCompleta, datos, error: null })
      })
      .catch((e: unknown) => {
        const mensaje = e instanceof ApiError ? e.message : 'No se pudo cargar la lista.'
        if (activo) setResultado({ clave: claveCompleta, datos: [], error: mensaje })
      })
    return () => {
      activo = false
    }
  }, [clave, intento])

  const claveActual = clave === null ? null : `${clave}#${intento}`
  const vigente = resultado !== null && resultado.clave === claveActual ? resultado : null

  return {
    datos: vigente?.datos ?? [],
    cargando: clave !== null && vigente === null,
    error: vigente?.error ?? null,
    reintentar: () => setIntento((n) => n + 1),
  }
}