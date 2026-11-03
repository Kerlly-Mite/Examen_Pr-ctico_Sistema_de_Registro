import { MAX_DESCRIPCION } from './experiencia'
import { fechaHoyISO } from './formato'
import type { Errores } from './lista'
import type { Capacitacion, Experiencia, Familiar, Formacion, Publicacion } from '../types/postulacion'

export const FAMILIAR_VACIO: Familiar = {
  nombres_apellidos: '',
  parentesco: '',
  institucion: '',
  area: '',
  cargo: '',
}

export const FORMACION_VACIA: Formacion = {
  nivel_instruccion: '',
  institucion: '',
  tipo: '',
  titulo: '',
  area_estudios: '',
  registro_senescyt: '',
  fecha_registro_senescyt: null,
  documento_respaldo: '',
}

export const EXPERIENCIA_VACIA: Experiencia = {
  tipo: '',
  cargo: '',
  institucion: '',
  fecha_inicio: null,
  fecha_fin: null,
  actividades: '',
  descripcion_funciones: '',
  documento_respaldo: '',
}

export const PUBLICACION_VACIA: Publicacion = {
  titulo: '',
  tipo: '',
  medio_publicacion: '',
  fecha_publicacion: null,
  relacionada_formacion: false,
  documento_respaldo: '',
}

export const CAPACITACION_VACIA: Capacitacion = {
  tipo_evento: '',
  nombre_evento: '',
  institucion: '',
  fecha_inicio: null,
  fecha_fin: null,
  horas: 0,
  documento_respaldo: '',
}

function errorFecha(v: string | null, que: string): string | undefined {
  if (v === null || v === '') return `Ingrese ${que}`
  if (v > fechaHoyISO()) return 'La fecha no puede ser futura'
  return undefined
}

export function validarFamiliar(f: Familiar): Errores<Familiar> {
  const e: Errores<Familiar> = {}
  if (f.nombres_apellidos.trim() === '') e.nombres_apellidos = 'Ingrese los nombres y apellidos del familiar'
  if (f.parentesco === '') e.parentesco = 'Seleccione el parentesco'
  if (f.institucion === '') e.institucion = 'Seleccione la institución'
  if (f.area.trim() === '') e.area = 'Ingrese el área donde trabaja'
  if (f.cargo.trim() === '') e.cargo = 'Ingrese el cargo'
  return e
}

export function validarFormacion(f: Formacion): Errores<Formacion> {
  const e: Errores<Formacion> = {}
  if (f.nivel_instruccion === '') e.nivel_instruccion = 'Seleccione el nivel de instrucción'
  if (f.institucion.trim() === '') e.institucion = 'Ingrese la institución educativa'
  if (f.tipo === '') e.tipo = 'Seleccione si es nacional o extranjera'
  if (f.titulo.trim() === '') e.titulo = 'Ingrese el título obtenido'
  if (f.area_estudios === '') e.area_estudios = 'Seleccione el área de estudios'
  if (f.registro_senescyt.trim() === '') e.registro_senescyt = 'Ingrese el número de registro SENESCYT'
  const fecha = errorFecha(f.fecha_registro_senescyt, 'la fecha de registro')
  if (fecha) e.fecha_registro_senescyt = fecha
  if (f.documento_respaldo === '') e.documento_respaldo = 'Cargue el documento de respaldo (PDF)'
  return e
}

export function validarExperiencia(f: Experiencia): Errores<Experiencia> {
  const e: Errores<Experiencia> = {}
  if (f.tipo === '') e.tipo = 'Seleccione el tipo de experiencia'
  if (f.cargo.trim() === '') e.cargo = 'Ingrese el cargo o puesto'
  if (f.institucion.trim() === '') e.institucion = 'Ingrese la institución o empresa'
  const inicio = errorFecha(f.fecha_inicio, 'la fecha de inicio')
  if (inicio) e.fecha_inicio = inicio
  const fin = errorFecha(f.fecha_fin, 'la fecha de finalización')
  if (fin) e.fecha_fin = fin
  else if (f.fecha_inicio !== null && f.fecha_fin !== null && f.fecha_fin < f.fecha_inicio) {
    e.fecha_fin = 'La fecha de finalización no puede ser anterior al inicio'
  }
  if (f.actividades === '') e.actividades = 'Seleccione el área de las actividades'
  if (f.descripcion_funciones.trim() === '') e.descripcion_funciones = 'Describa sus funciones'
  else if (f.descripcion_funciones.length > MAX_DESCRIPCION) {
    e.descripcion_funciones = `No puede superar ${MAX_DESCRIPCION} caracteres`
  }
  if (f.documento_respaldo === '') e.documento_respaldo = 'Cargue el documento de respaldo (PDF)'
  return e
}

export function validarPublicacion(f: Publicacion): Errores<Publicacion> {
  const e: Errores<Publicacion> = {}
  if (f.titulo.trim() === '') e.titulo = 'Ingrese el título'
  if (f.tipo === '') e.tipo = 'Seleccione el tipo'
  if (f.medio_publicacion.trim() === '') e.medio_publicacion = 'Ingrese el medio de publicación'
  const fecha = errorFecha(f.fecha_publicacion, 'la fecha de publicación')
  if (fecha) e.fecha_publicacion = fecha
  if (f.documento_respaldo === '') e.documento_respaldo = 'Cargue el documento de respaldo (PDF)'
  return e
}

export function validarCapacitacion(f: Capacitacion): Errores<Capacitacion> {
  const e: Errores<Capacitacion> = {}
  if (f.tipo_evento === '') e.tipo_evento = 'Seleccione el tipo de evento'
  if (f.nombre_evento.trim() === '') e.nombre_evento = 'Ingrese el nombre del evento'
  if (f.institucion.trim() === '') e.institucion = 'Ingrese la institución'
  const inicio = errorFecha(f.fecha_inicio, 'la fecha de inicio')
  if (inicio) e.fecha_inicio = inicio
  const fin = errorFecha(f.fecha_fin, 'la fecha de fin')
  if (fin) e.fecha_fin = fin
  else if (f.fecha_inicio !== null && f.fecha_fin !== null && f.fecha_fin < f.fecha_inicio) {
    e.fecha_fin = 'La fecha de fin no puede ser anterior al inicio'
  }
  if (f.horas <= 0) e.horas = 'Ingrese un número de horas mayor que 0'
  if (f.documento_respaldo === '') e.documento_respaldo = 'Cargue el documento de respaldo (PDF)'
  return e
}