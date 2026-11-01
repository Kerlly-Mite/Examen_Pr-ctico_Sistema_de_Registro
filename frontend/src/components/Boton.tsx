import type { ButtonHTMLAttributes } from 'react'

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  cargando?: boolean
}

export function Boton({ cargando = false, disabled, children, className = '', ...resto }: Props) {
  return (
    <button
      {...resto}
      disabled={disabled || cargando}
      className={`w-full rounded bg-blue-800 px-4 py-2 font-medium text-white hover:bg-blue-900 disabled:cursor-not-allowed disabled:opacity-60 ${className}`}
    >
      {cargando ? 'Procesando...' : children}
    </button>
  )
}