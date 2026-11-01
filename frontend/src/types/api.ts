export interface ApiErrorBody {
  error: string
  message: string
  intentos_restantes?: number
  segundos_restantes?: number
}

export interface Postulante {
  id: number
  cedula: string
  correo: string
  nombres: string
  apellidos: string
}

export interface LoginRespuesta {
  token: string
  postulante: Postulante
}

export interface VerificarIdentidadRespuesta {
  token_registro: string
  nombres: string
  apellidos: string
  expira_en_segundos: number
}

export interface MensajeRespuesta {
  message: string
}

export interface Pais {
  id: number
  nombre: string
}

export interface Provincia {
  id: number
  pais_id: number
  nombre: string
}

export interface Canton {
  id: number
  provincia_id: number
  nombre: string
}

export interface RespuestaLista<T> {
  data: T[]
}