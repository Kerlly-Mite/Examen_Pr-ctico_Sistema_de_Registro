import type { ButtonHTMLAttributes } from 'react'

type Variante = 'primario' | 'oro' | 'borde'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  cargando?: boolean
  variante?: Variante
  autoAncho?: boolean
}

const ESTILOS: Record<Variante, string> = {
  primario: 'bg-biess-800 text-white hover:bg-biess-900 disabled:bg-slate-400 disabled:hover:bg-slate-400',
  oro: 'bg-oro-500 text-slate-900 hover:bg-oro-600 disabled:bg-slate-300 disabled:hover:bg-slate-300',
  borde:
    'border border-biess-800 bg-white text-biess-800 hover:bg-biess-50 disabled:border-slate-300 disabled:text-slate-400 disabled:hover:bg-white',
}

export function Boton({
  cargando = false,
  variante = 'primario',
  autoAncho = false,
  disabled,
  children,
  className = '',
  ...resto
}: Props) {
  return (
    <button
      {...resto}
      disabled={disabled || cargando}
      className={`rounded py-2 font-medium transition-colors disabled:cursor-not-allowed ${
        autoAncho ? 'px-5' : 'w-full px-4'
      } ${ESTILOS[variante]} ${className}`}
    >
      {cargando ? 'Procesando...' : children}
    </button>
  )
}