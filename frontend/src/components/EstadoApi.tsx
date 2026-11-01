import { useEffect, useState } from 'react'
import { api, ApiError } from '../api/client'
import { RUTAS } from '../api/rutas'
import type { Pais, RespuestaLista } from '../types/api'

type Estado =
  | { tipo: 'cargando' }
  | { tipo: 'ok'; total: number }
  | { tipo: 'error'; mensaje: string }

export function EstadoApi() {
  const [estado, setEstado] = useState<Estado>({ tipo: 'cargando' })

  useEffect(() => {
    let activo = true
    api
      .get<RespuestaLista<Pais>>(RUTAS.paises)
      .then((r) => {
        if (activo) setEstado({ tipo: 'ok', total: r.data.length })
      })
      .catch((e: unknown) => {
        const mensaje = e instanceof ApiError ? e.message : 'No se pudo verificar la API.'
        if (activo) setEstado({ tipo: 'error', mensaje })
      })
    return () => {
      activo = false
    }
  }, [])

  if (estado.tipo === 'cargando') {
    return <span className="text-slate-500">Verificando API...</span>
  }
  if (estado.tipo === 'error') {
    return <span className="text-red-600">API no disponible: {estado.mensaje}</span>
  }
  return <span className="text-green-700">API conectada: {estado.total} países cargados</span>
}