export interface Opcion {
  valor: string
  etiqueta: string
}

export const PARENTESCOS: readonly Opcion[] = [
  { valor: 'CONVIVIENTE', etiqueta: 'Conviviente' },
  { valor: 'CONYUGE', etiqueta: 'Cónyuge' },
  { valor: 'CUNADO', etiqueta: 'Cuñado/a' },
  { valor: 'HERMANO', etiqueta: 'Hermano/a' },
  { valor: 'HIJO_POLITICO', etiqueta: 'Hijo político' },
  { valor: 'HIJO', etiqueta: 'Hijo/a' },
  { valor: 'MADRE', etiqueta: 'Madre' },
  { valor: 'NIETO', etiqueta: 'Nieto/a' },
  { valor: 'NUERA', etiqueta: 'Nuera' },
  { valor: 'PADRE', etiqueta: 'Padre' },
  { valor: 'PRIMO_HERMANO', etiqueta: 'Primo hermano/a' },
  { valor: 'SOBRINO', etiqueta: 'Sobrino/a' },
  { valor: 'SUEGRO', etiqueta: 'Suegro/a' },
  { valor: 'TIO', etiqueta: 'Tío/a' },
  { valor: 'YERNO', etiqueta: 'Yerno' },
]

export const INSTITUCIONES: readonly Opcion[] = [
  { valor: 'IESS', etiqueta: 'IESS' },
  { valor: 'BIESS', etiqueta: 'BIESS' },
]

export const NIVELES: readonly Opcion[] = [
  { valor: 'TERCER_NIVEL', etiqueta: 'Tercer Nivel de Grado' },
  { valor: 'CUARTO_ESPECIALIDAD', etiqueta: 'Cuarto Nivel - Especialidad Académica' },
  { valor: 'CUARTO_MAESTRIA', etiqueta: 'Cuarto Nivel - Maestría Académica' },
  { valor: 'CUARTO_PHD', etiqueta: 'Cuarto Nivel PhD - Doctorado' },
]

export const TIPOS_INSTITUCION: readonly Opcion[] = [
  { valor: 'NACIONAL', etiqueta: 'Nacional' },
  { valor: 'EXTRANJERA', etiqueta: 'Extranjera' },
]

export const AREAS: readonly Opcion[] = [
  { valor: 'FINANZAS', etiqueta: 'Finanzas' },
  { valor: 'MERCADO_CAPITALES', etiqueta: 'Mercado de capitales' },
  { valor: 'ADMINISTRACION', etiqueta: 'Administración de empresas' },
  { valor: 'ECONOMIA', etiqueta: 'Economía' },
  { valor: 'DERECHO', etiqueta: 'Derecho' },
  { valor: 'AFINES', etiqueta: 'Materias afines' },
]

export const TIPOS_EXPERIENCIA: readonly Opcion[] = [
  { valor: 'GENERAL', etiqueta: 'General' },
  { valor: 'DIRECCION', etiqueta: 'Dirección' },
]

export const TIPOS_PUBLICACION: readonly Opcion[] = [
  { valor: 'PUBLICACION', etiqueta: 'Publicación' },
  { valor: 'TRABAJO_INVESTIGACION', etiqueta: 'Trabajo de investigación' },
]

export const TIPOS_EVENTO: readonly Opcion[] = [
  { valor: 'CURSO', etiqueta: 'Curso' },
  { valor: 'SEMINARIO', etiqueta: 'Seminario' },
  { valor: 'TALLER', etiqueta: 'Taller' },
]

export function etiquetaDe(opciones: readonly Opcion[], valor: string): string {
  return opciones.find((o) => o.valor === valor)?.etiqueta ?? valor
}