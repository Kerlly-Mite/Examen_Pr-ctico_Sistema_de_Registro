// Cliente HTTP único del frontend: todas las llamadas a la API pasan por aquí.
// Base "/api" -> en desarrollo Vite lo reenvía al backend (ver vite.config.ts).

export class ApiError extends Error {
  status: number
  code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

export async function http<T>(ruta: string, init: RequestInit = {}): Promise<T> {
  let res: Response
  try {
    res = await fetch(`/api${ruta}`, {
      ...init,
      headers: { 'Content-Type': 'application/json', ...(init.headers ?? {}) },
    })
  } catch {
    // fetch solo rechaza ante fallos de red (servidor apagado, sin conexión).
    throw new ApiError(0, 'sin_conexion', 'No se pudo conectar con el servidor. Verifique su conexión.')
  }

  const cuerpo = await res.json().catch(() => null)
  if (!res.ok) {
    throw new ApiError(res.status, cuerpo?.error ?? 'error', cuerpo?.message ?? 'Ocurrió un error inesperado')
  }
  return cuerpo as T
}
