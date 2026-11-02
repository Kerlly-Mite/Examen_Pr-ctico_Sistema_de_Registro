// TEMPORAL: guarda el progreso del formulario en el navegador (localStorage).
// Se reemplaza por la API de la integrante B cuando exista el endpoint de postulaciones.
const PREFIJO = 'biess_borrador_'

export function leerSeccion<T>(cedula: string, seccion: string): T | null {
  try {
    const crudo = localStorage.getItem(`${PREFIJO}${cedula}_${seccion}`)
    return crudo === null ? null : (JSON.parse(crudo) as T)
  } catch {
    return null
  }
}

export function guardarSeccion<T>(cedula: string, seccion: string, datos: T): boolean {
  try {
    localStorage.setItem(`${PREFIJO}${cedula}_${seccion}`, JSON.stringify(datos))
    return true
  } catch {
    return false
  }
}