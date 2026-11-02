import type { InputHTMLAttributes } from 'react'

interface Props extends InputHTMLAttributes<HTMLInputElement> {
  etiqueta: string
  error?: string
  ayuda?: string
}

export function Campo({ etiqueta, error, ayuda, id, ...resto }: Props) {
  const idCampo = id ?? resto.name ?? etiqueta
  return (
    <div>
      <label htmlFor={idCampo} className="block text-sm font-medium text-slate-700">
        {etiqueta}
      </label>
      <input
        id={idCampo}
        {...resto}
        aria-invalid={error ? true : undefined}
        className={`mt-1 w-full rounded border px-3 py-2 outline-none focus:ring-2 focus:ring-oro-500 disabled:bg-slate-100 read-only:bg-slate-50 read-only:text-slate-600 ${
          error ? 'border-red-500' : 'border-slate-300'
        }`}
      />
      {ayuda && !error && <p className="mt-1 text-xs text-slate-500">{ayuda}</p>}
      {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
    </div>
  )
}