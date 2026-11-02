import { createContext } from 'react'
import type { Bloqueo, OpcionesPostulacion, Postulacion, PostulacionReq } from '../types/postulacion'

export type ResultadoAccion = { ok: true } | { ok: false; mensaje: string; codigo: string }

export interface FormularioContextValue {
  cargando: boolean
  errorCarga: string | null
  postulacion: Postulacion | null
  opciones: OpcionesPostulacion | null
  bloqueo: Bloqueo | null
  enviada: boolean
  guardando: boolean
  recargar: () => void
  actualizar: (parcial: Partial<PostulacionReq>) => void
  guardar: () => Promise<ResultadoAccion>
  enviar: (consentimiento: boolean) => Promise<ResultadoAccion>
  bloquear: (bloqueo: Bloqueo) => void
}

export const FormularioContext = createContext<FormularioContextValue | null>(null)