import type { SelectHTMLAttributes } from 'react'

interface Opcion {
  id: number
  nombre: string
}

interface Props extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'onChange' | 'value'> {
  etiqueta: string
  opciones: Opcion[]
  valor: number | null
  onCambio: (id: number | null) => void
  placeholder?: string
  error?: string
}

export function Selector({
  etiqueta,
  opciones,
  valor,
  onCambio,
  placeholder = 'Seleccione...',
  error,
  id,
  ...resto
}: Props) {
  const idCampo = id ?? resto.name ?? etiqueta
  return (
    <div>
      <label htmlFor={idCampo} className="block text-sm font-medium text-slate-700">
        {etiqueta}
      </label>
      <select
        id={idCampo}
        {...resto}
        value={valor ?? ''}
        onChange={(e) => onCambio(e.target.value === '' ? null : Number(e.target.value))}
        aria-invalid={error ? true : undefined}
        className={`mt-1 w-full rounded border bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-oro-500 disabled:bg-slate-100 ${
          error ? 'border-red-500' : 'border-slate-300'
        }`}
      >
        <option value="">{placeholder}</option>
        {opciones.map((o) => (
          <option key={o.id} value={o.id}>
            {o.nombre}
          </option>
        ))}
      </select>
      {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
    </div>
  )
}