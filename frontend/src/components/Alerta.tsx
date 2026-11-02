import type { ReactNode } from 'react'

interface Props {
  tipo: 'error' | 'exito' | 'info' | 'aviso'
  children: ReactNode
}

const estilos: Record<Props['tipo'], string> = {
  error: 'border-red-300 bg-red-50 text-red-800',
  exito: 'border-green-300 bg-green-50 text-green-800',
  info: 'border-biess-100 bg-biess-50 text-biess-800',
  aviso: 'border-amber-300 bg-amber-50 text-amber-900',
}

export function Alerta({ tipo, children }: Props) {
  return (
    <div
      role={tipo === 'error' ? 'alert' : 'status'}
      className={`rounded border px-3 py-2 text-sm ${estilos[tipo]}`}
    >
      {children}
    </div>
  )
}