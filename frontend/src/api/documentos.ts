import { ApiError, obtenerToken } from './client'

export type TipoDocumento = 'respaldo' | 'organizacion'

interface RespuestaSubida {
  id: string | number
  nombre: string
  tamano: number
}

// El id que devuelve la API trae el nombre original después de "__"
export function nombreDocumento(id: string): string {
  if (id === '') return ''
  const partes = id.split('__')
  return partes.length > 1 ? partes.slice(1).join('__') : id
}

async function leerError(respuesta: Response): Promise<ApiError> {
  const cuerpo: unknown = await respuesta.json().catch(() => null)
  const dato = typeof cuerpo === 'object' && cuerpo !== null ? (cuerpo as Record<string, unknown>) : {}
  const codigo = typeof dato.error === 'string' ? dato.error : 'error_desconocido'
  const mensaje = typeof dato.message === 'string' ? dato.message : 'Ocurrió un error inesperado.'
  return new ApiError(respuesta.status, { error: codigo, message: mensaje })
}

async function pedir(ruta: string, init: RequestInit): Promise<Response> {
  const headers = new Headers(init.headers)
  const token = obtenerToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)

  let respuesta: Response
  try {
    respuesta = await fetch(ruta, { ...init, headers })
  } catch {
    throw new ApiError(0, {
      error: 'sin_conexion',
      message: 'No se pudo conectar con el servidor. Intente nuevamente.',
    })
  }
  if (!respuesta.ok) throw await leerError(respuesta)
  return respuesta
}

function descargar(url: string, nombre: string): void {
  const enlace = document.createElement('a')
  enlace.href = url
  enlace.download = nombre
  document.body.appendChild(enlace)
  enlace.click()
  enlace.remove()
}

// POST multipart: campos "archivo" (el PDF) y "tipo". Devuelve el id que se guarda en la postulación.
export async function subirDocumento(archivo: File, tipo: TipoDocumento) {
  const datos = new FormData()
  datos.append('archivo', archivo)
  datos.append('tipo', tipo)
  const respuesta = await pedir('/api/postulacion/documentos', { method: 'POST', body: datos })
  const r = (await respuesta.json()) as RespuestaSubida
  return { id: String(r.id), nombre: r.nombre || archivo.name, tamano: r.tamano }
}

// La ruta pide el token: se descarga con fetch y se abre como blob
export async function abrirDocumento(id: string): Promise<void> {
  const ventana = window.open('', '_blank')
  try {
    const respuesta = await pedir(`/api/postulacion/documentos/${encodeURIComponent(id)}`, { method: 'GET' })
    const url = URL.createObjectURL(await respuesta.blob())
    if (ventana) ventana.location.href = url
    else descargar(url, nombreDocumento(id))
    window.setTimeout(() => URL.revokeObjectURL(url), 60000)
  } catch (err) {
    ventana?.close()
    throw err
  }
}

export async function descargarFormularioPdf(nombreArchivo: string): Promise<void> {
  const respuesta = await pedir('/api/postulacion/formulario.pdf', { method: 'GET' })
  const url = URL.createObjectURL(await respuesta.blob())
  descargar(url, nombreArchivo)
  window.setTimeout(() => URL.revokeObjectURL(url), 10000)
}