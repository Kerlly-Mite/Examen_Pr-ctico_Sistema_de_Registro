interface Props {
  etiqueta: string
  name: string
  valor: string
  onCambio: (valor: string) => void
  maxLength: number
  error?: string
}

export function AreaTexto({ etiqueta, name, valor, onCambio, maxLength, error }: Props) {
  return (
    <div>
      <label htmlFor={name} className="block text-sm font-medium text-slate-700">
        {etiqueta}
      </label>
      <textarea
        id={name}
        name={name}
        rows={4}
        maxLength={maxLength}
        value={valor}
        onChange={(e) => onCambio(e.target.value)}
        aria-invalid={error ? true : undefined}
        className={`mt-1 w-full rounded border px-3 py-2 outline-none focus:ring-2 focus:ring-blue-500 ${
          error ? 'border-red-500' : 'border-slate-300'
        }`}
      />
      <p className="mt-1 text-right text-xs text-slate-500">
        {valor.length} / {maxLength}
      </p>
      {error && <p className="text-xs text-red-600">{error}</p>}
    </div>
  )
}