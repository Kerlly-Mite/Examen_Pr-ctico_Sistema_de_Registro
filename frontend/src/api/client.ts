import type { ApiErrorBody } from '../types/api'

// Ruta relativa: Vite reenvía /api al backend (proxy)
const BASE_URL = ''
const TOKEN_KEY = 'biess_token'

export function obtenerToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function guardarToken(token: string | null): void {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    else localStorage.removeItem(TOKEN_KEY)
  } catch {
    // si el navegador bloquea el almacenamiento, la sesión solo dura en memoria
  }
}

export class ApiError extends Error {
  status: number
  codigo: string
  intentosRestantes?: number
  segundosRestantes?: number

  constructor(status: number, cuerpo: ApiErrorBody) {
    super(cuerpo.message)
    this.name = 'ApiError'
    this.status = status
    this.codigo = cuerpo.error
    this.intentosRestantes = cuerpo.intentos_restantes
    this.segundosRestantes = cuerpo.segundos_restantes
  }
}

type ManejadorNoAutorizado = () => void
let manejadorNoAutorizado: ManejadorNoAutorizado | null = null

export function registrarManejadorNoAutorizado(fn: ManejadorNoAutorizado | null): void {
  manejadorNoAutorizado = fn
}

function esCuerpoDeError(valor: unknown): valor is ApiErrorBody {
  return (
    typeof valor === 'object' &&
    valor !== null &&
    'error' in valor &&
    'message' in valor
  )
}

// Si el servidor no responde en este tiempo se cancela la petición, para que la pantalla no quede cargando sin fin
export const TIEMPO_LIMITE_MS = 20_000
export const TIEMPO_LIMITE_ARCHIVOS_MS = 60_000

// Convierte el fallo de fetch (sin red o tiempo agotado) en un ApiError con mensaje para el usuario
export function errorDeRed(err: unknown): ApiError {
  if (err instanceof DOMException && (err.name === 'TimeoutError' || err.name === 'AbortError')) {
    return new ApiError(0, {
      error: 'tiempo_agotado',
      message: 'El servidor tardó demasiado en responder. Intente nuevamente.',
    })
  }
  return new ApiError(0, {
    error: 'sin_conexion',
    message: 'No se pudo conectar con el servidor. Intente nuevamente.',
  })
}

async function peticion<T>(metodo: string, ruta: string, cuerpo?: unknown): Promise<T> {
  const headers: Record<string, string> = {}
  if (cuerpo !== undefined) headers['Content-Type'] = 'application/json'
  const token = obtenerToken()
  if (token) headers.Authorization = `Bearer ${token}`

  let respuesta: Response
  try {
    respuesta = await fetch(`${BASE_URL}${ruta}`, {
      method: metodo,
      headers,
      body: cuerpo !== undefined ? JSON.stringify(cuerpo) : undefined,
      signal: AbortSignal.timeout(TIEMPO_LIMITE_MS),
    })
  } catch (err) {
    throw errorDeRed(err)
  }

  let datos: unknown
  try {
    datos = await respuesta.json()
  } catch {
    datos = null
  }

  if (!respuesta.ok) {
    const cuerpoError: ApiErrorBody = esCuerpoDeError(datos)
      ? datos
      : { error: 'error_desconocido', message: 'Ocurrió un error inesperado.' }

    // Solo cerramos sesión si el token no es válido (no cuando fallan credenciales en el login)
    if (respuesta.status === 401 && cuerpoError.error === 'no_autorizado') {
      manejadorNoAutorizado?.()
    }
    throw new ApiError(respuesta.status, cuerpoError)
  }

  return datos as T
}

export const api = {
  get: <T>(ruta: string) => peticion<T>('GET', ruta),
  post: <T>(ruta: string, cuerpo?: unknown) => peticion<T>('POST', ruta, cuerpo),
  put: <T>(ruta: string, cuerpo?: unknown) => peticion<T>('PUT', ruta, cuerpo),
  delete: <T>(ruta: string) => peticion<T>('DELETE', ruta),
}