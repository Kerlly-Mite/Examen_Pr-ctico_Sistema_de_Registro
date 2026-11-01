import { api } from './client'
import { RUTAS } from './rutas'
import type {
  LoginRespuesta,
  MensajeRespuesta,
  VerificarIdentidadRespuesta,
} from '../types/api'

export type PruebaIdentidad = { fecha_expedicion: string } | { codigo_dactilar: string }

export function verificarIdentidad(cedula: string, prueba: PruebaIdentidad) {
  return api.post<VerificarIdentidadRespuesta>(RUTAS.verificarIdentidad, { cedula, ...prueba })
}

export function registrar(tokenRegistro: string, correo: string, password: string) {
  return api.post<MensajeRespuesta>(RUTAS.registro, {
    token_registro: tokenRegistro,
    correo,
    password,
  })
}

export function iniciarSesionApi(cedula: string, password: string) {
  return api.post<LoginRespuesta>(RUTAS.login, { cedula, password })
}

export function solicitarRecuperacion(cedula: string, prueba: PruebaIdentidad) {
  return api.post<MensajeRespuesta>(RUTAS.recuperarSolicitar, { cedula, ...prueba })
}

export function restablecerPassword(token: string, password: string) {
  return api.post<MensajeRespuesta>(RUTAS.recuperarRestablecer, { token, password })
}