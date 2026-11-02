export const PARENTESCOS = [
  'Conviviente',
  'Cónyuge',
  'Cuñado/a',
  'Hermano/a',
  'Hijo político',
  'Hijo/a',
  'Madre',
  'Nieto/a',
  'Nuera',
  'Padre',
  'Primo hermano/a',
  'Sobrino/a',
  'Suegro/a',
  'Tío/a',
  'Yerno',
] as const

export const INSTITUCIONES = ['IESS', 'BIESS'] as const

export interface Familiar {
  id: string
  nombres: string
  parentesco: string
  institucion: string
  area: string
  cargo: string
}

export type FamiliarForm = Omit<Familiar, 'id'>

export const FAMILIAR_VACIO: FamiliarForm = {
  nombres: '',
  parentesco: '',
  institucion: '',
  area: '',
  cargo: '',
}

export interface InformacionAdicional {
  tieneFamiliares: boolean | null
  familiares: Familiar[]
}

export const ADICIONAL_VACIO: InformacionAdicional = {
  tieneFamiliares: null,
  familiares: [],
}

export type ErroresFamiliar = Partial<Record<keyof FamiliarForm, string>>

export function validarFamiliar(f: FamiliarForm): ErroresFamiliar {
  const e: ErroresFamiliar = {}
  if (f.nombres.trim() === '') e.nombres = 'Ingrese los nombres y apellidos del familiar'
  if (f.parentesco === '') e.parentesco = 'Seleccione el parentesco'
  if (f.institucion === '') e.institucion = 'Seleccione la institución'
  if (f.area.trim() === '') e.area = 'Ingrese el área donde trabaja'
  if (f.cargo.trim() === '') e.cargo = 'Ingrese el cargo'
  return e
}

export function hayErroresFamiliar(e: ErroresFamiliar): boolean {
  return Object.values(e).some((v) => v !== undefined)
}

export function formularioFamiliarConContenido(f: FamiliarForm): boolean {
  return Object.values(f).some((v) => v.trim() !== '')
}