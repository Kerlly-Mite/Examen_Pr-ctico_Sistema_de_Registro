import type { Duracion } from '../types/postulacion'

export const MAX_DESCRIPCION = 1000

function leerFecha(iso: string): { anio: number; mes: number; dia: number } | null {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso)
  return m ? { anio: Number(m[1]), mes: Number(m[2]), dia: Number(m[3]) } : null
}

function diasDelMes(anio: number, mes: number): number {
  return new Date(anio, mes, 0).getDate()
}

// Vista previa local (cuenta el día de inicio y el de fin). El valor oficial lo calcula el servidor al guardar.
export function calcularDuracion(inicio: string, fin: string): Duracion | null {
  const i = leerFecha(inicio)
  const f = leerFecha(fin)
  if (i === null || f === null) return null

  const finMasUno = new Date(f.anio, f.mes - 1, f.dia + 1)
  let anios = finMasUno.getFullYear() - i.anio
  let meses = finMasUno.getMonth() + 1 - i.mes
  let dias = finMasUno.getDate() - i.dia

  if (dias < 0) {
    meses -= 1
    dias += diasDelMes(finMasUno.getFullYear(), finMasUno.getMonth())
  }
  if (meses < 0) {
    anios -= 1
    meses += 12
  }
  if (anios < 0) return null
  return { anios, meses, dias }
}

function conUnidad(n: number, uno: string, varios: string): string {
  return `${n} ${n === 1 ? uno : varios}`
}

export function formatearDuracion(d: Duracion): string {
  return `${conUnidad(d.anios, 'año', 'años')}, ${conUnidad(d.meses, 'mes', 'meses')} y ${conUnidad(d.dias, 'día', 'días')}`
}