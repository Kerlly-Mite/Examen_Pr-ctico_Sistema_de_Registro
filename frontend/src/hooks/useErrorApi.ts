import { useCallback, useState } from 'react'
import { ApiError } from '../api/client'
import { useCuentaRegresiva } from './useCuentaRegresiva'

export function useErrorApi() {
  const { restantes, bloqueado, iniciar } = useCuentaRegresiva()
  const [error, setError] = useState<string | null>(null)
  const [codigo, setCodigo] = useState<string | null>(null)
  const [intentos, setIntentos] = useState<number | undefined>(undefined)

  const manejar = useCallback(
    (e: unknown) => {
      if (e instanceof ApiError) {
        setError(e.message)
        setCodigo(e.codigo)
        setIntentos(e.intentosRestantes)
        if (e.codigo === 'cuenta_bloqueada') iniciar(e.segundosRestantes ?? 600)
      } else {
        setError('Ocurrió un error inesperado.')
        setCodigo(null)
        setIntentos(undefined)
      }
    },
    [iniciar],
  )

  const limpiar = useCallback(() => {
    setError(null)
    setCodigo(null)
    setIntentos(undefined)
  }, [])

  // Cuando termina el bloqueo, el mensaje de bloqueo desaparece solo
  const bloqueoTerminado = codigo === 'cuenta_bloqueada' && !bloqueado

  return {
    error: bloqueoTerminado ? null : error,
    codigo,
    intentos,
    restantes,
    bloqueado,
    manejar,
    limpiar,
  }
}