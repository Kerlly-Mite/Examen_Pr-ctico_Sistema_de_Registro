export function fechaHoyISO(): string {
  const hoy = new Date()
  const mes = String(hoy.getMonth() + 1).padStart(2, '0')
  const dia = String(hoy.getDate()).padStart(2, '0')
  return `${hoy.getFullYear()}-${mes}-${dia}`
}

export function formatearTiempo(segundos: number): string {
  const min = Math.floor(segundos / 60)
  const seg = segundos % 60
  return `${String(min).padStart(2, '0')}:${String(seg).padStart(2, '0')}`
}

// "2024-03-15" -> "15/03/2024"
export function formatearFecha(iso: string | null): string {
  if (!iso) return ''
  const [anio, mes, dia] = iso.split('-')
  return anio && mes && dia ? `${dia}/${mes}/${anio}` : iso
}

// "UNION_DE_HECHO" -> "Union de hecho"
export function formatearTexto(valor: string): string {
  const t = valor.replace(/_/g, ' ').toLowerCase()
  return t.charAt(0).toUpperCase() + t.slice(1)
}