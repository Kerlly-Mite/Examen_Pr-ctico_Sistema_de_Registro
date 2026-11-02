import type { ReactNode } from 'react'

interface Props<T> {
  nombreItem: string
  editando: boolean
  campos: ReactNode
  onAgregar: () => void
  onCancelar: () => void
  items: readonly T[]
  resumen: (item: T) => ReactNode
  textoVacio: string
  onEditar: (indice: number) => void
  onEliminar: (indice: number) => void
}

const botonPrincipal = 'rounded bg-biess-800 px-4 py-2 text-sm font-medium text-white hover:bg-biess-900'
const botonSecundario = 'rounded border border-slate-300 px-3 py-1 text-sm hover:bg-slate-50'

export function PanelLista<T>({
  nombreItem,
  editando,
  campos,
  onAgregar,
  onCancelar,
  items,
  resumen,
  textoVacio,
  onEditar,
  onEliminar,
}: Props<T>) {
  return (
    <div className="space-y-4 rounded border border-slate-200 p-4">
      <h3 className="text-sm font-semibold">
        {editando ? 'Editar' : 'Agregar'} {nombreItem}
      </h3>

      <div className="grid gap-4 md:grid-cols-2">{campos}</div>

      <div className="flex gap-2">
        <button type="button" onClick={onAgregar} className={botonPrincipal}>
          {editando ? 'Actualizar' : 'Agregar'}
        </button>
        {editando && (
          <button type="button" onClick={onCancelar} className={botonSecundario}>
            Cancelar
          </button>
        )}
      </div>

      {items.length > 0 ? (
        <ul className="divide-y divide-slate-200 rounded border border-slate-200">
          {items.map((item, indice) => (
            <li key={indice} className="flex flex-wrap items-center justify-between gap-2 p-3">
              <div className="text-sm">{resumen(item)}</div>
              <div className="flex gap-2">
                <button type="button" onClick={() => onEditar(indice)} className={botonSecundario}>
                  Editar
                </button>
                <button
                  type="button"
                  onClick={() => onEliminar(indice)}
                  className={`${botonSecundario} text-red-700`}
                >
                  Eliminar
                </button>
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-sm text-slate-500">{textoVacio}</p>
      )}
    </div>
  )
}