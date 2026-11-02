interface Props {
  name: string
  etiqueta: string
  valor: boolean | null
  onCambio: (valor: boolean) => void
  error?: string
  disabled?: boolean
}

export function GrupoSiNo({ name, etiqueta, valor, onCambio, error, disabled }: Props) {
  return (
    <fieldset>
      <legend className="text-sm font-medium text-slate-700">{etiqueta}</legend>
      <div className="mt-1 flex gap-6 text-sm">
        <label className="flex items-center gap-1">
          <input
            type="radio"
            name={name}
            checked={valor === true}
            onChange={() => onCambio(true)}
            disabled={disabled}
          />
          Sí
        </label>
        <label className="flex items-center gap-1">
          <input
            type="radio"
            name={name}
            checked={valor === false}
            onChange={() => onCambio(false)}
            disabled={disabled}
          />
          No
        </label>
      </div>
      {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
    </fieldset>
  )
}