import { useRef, useState } from 'react'
import type { ChangeEvent } from 'react'

interface Props {
  etiqueta: string
  maxMB: number
  nombre: string
  onCambio: (nombre: string) => void
  error?: string
}

// Valida que sea PDF y que no pase del tamaño máximo. Por ahora solo se guarda el nombre del archivo
// (el contrato de la API deja la subida real para una fase posterior).
export function CampoArchivo({ etiqueta, maxMB, nombre, onCambio, error }: Props) {
  const ref = useRef<HTMLInputElement>(null)
  const [errorLocal, setErrorLocal] = useState<string | null>(null)

  function elegir(e: ChangeEvent<HTMLInputElement>) {
    const archivo = e.target.files?.[0]
    e.target.value = ''
    if (!archivo) return

    const esPdf = archivo.type === 'application/pdf' || archivo.name.toLowerCase().endsWith('.pdf')
    if (!esPdf) {
      setErrorLocal('El archivo debe estar en formato PDF')
      return
    }
    if (archivo.size > maxMB * 1024 * 1024) {
      setErrorLocal(`El archivo no debe superar ${maxMB} MB`)
      return
    }
    setErrorLocal(null)
    onCambio(archivo.name)
  }

  const mensaje = errorLocal ?? error

  return (
    <div>
      <p className="block text-sm font-medium text-slate-700">{etiqueta}</p>
      <div className="mt-1 flex items-center gap-2">
        <button
          type="button"
          onClick={() => ref.current?.click()}
          className="rounded border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50"
        >
          Seleccionar PDF
        </button>
        <span className="truncate text-sm text-slate-600">
          {nombre !== '' ? nombre : `PDF de máximo ${maxMB} MB`}
        </span>
        {nombre !== '' && (
          <button
            type="button"
            onClick={() => {
              setErrorLocal(null)
              onCambio('')
            }}
            className="text-sm text-red-700 underline"
          >
            Quitar
          </button>
        )}
      </div>
      <input ref={ref} type="file" accept="application/pdf" onChange={elegir} className="hidden" />
      {mensaje && <p className="mt-1 text-xs text-red-600">{mensaje}</p>}
    </div>
  )
}