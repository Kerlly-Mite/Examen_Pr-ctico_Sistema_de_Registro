import { api } from './client'
import { RUTAS } from './rutas'
import type {
  OpcionesPostulacion,
  Postulacion,
  PostulacionReq,
  VerificacionCargo,
} from '../types/postulacion'

export function obtenerOpciones() {
  return api.get<OpcionesPostulacion>(RUTAS.postulacionOpciones)
}

export function obtenerPostulacion() {
  return api.get<Postulacion>(RUTAS.postulacion)
}

export function guardarPostulacion(req: PostulacionReq) {
  return api.put<Postulacion>(RUTAS.postulacion, req)
}

export function verificarCargo(codigo: string) {
  return api.get<VerificacionCargo>(RUTAS.verificarCargo(codigo))
}

export function enviarPostulacion(consentimiento: boolean) {
  return api.post<unknown>(RUTAS.postulacionEnviar, { consentimiento_aceptado: consentimiento })
}