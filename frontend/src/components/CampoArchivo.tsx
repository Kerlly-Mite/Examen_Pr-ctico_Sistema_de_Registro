import { useRef, useState } from 'react'
import type { ChangeEvent } from 'react'
import { ApiError } from '../api/client'
import { abrirDocumento, nombreDocumento, subirDocumento } from '../api/documentos'
import type { TipoDocumento } from '../api/documentos'

interface Props {
  etiqueta: string
  maxMB: number
  nombre: string
  onCambio: (id: string) => void
  tipo?: TipoDocumento
  error?: string
}

// Valida el PDF en el navegador, lo sube a la API y guarda el id que devuelve
export function CampoArchivo({ etiqueta, maxMB, nombre, onCambio, tipo = 'respaldo', error }: Props) {
  const ref = useRef<HTMLInputElement>(null)
  const [subiendo, setSubiendo] = useState(false)
  const [errorLocal, setErrorLocal] = useState<string | null>(null)
  const limite = Math.min(maxMB, tipo === 'organizacion' ? 4 : 2)

  async function elegir(e: ChangeEvent<HTMLInputElement>) {
    const archivo = e.target.files?.[0]
    e.target.value = ''
    if (!archivo) return

    const esPdf = archivo.type === 'application/pdf' || archivo.name.toLowerCase().endsWith('.pdf')
    if (!esPdf) {
      setErrorLocal('El archivo debe estar en formato PDF')
      return
    }
    if (archivo.size > limite * 1024 * 1024) {
      setErrorLocal(`El archivo no debe superar ${limite} MB`)
      return
    }

    setErrorLocal(null)
    setSubiendo(true)
    try {
      const r = await subirDocumento(archivo, tipo)
      onCambio(r.id)
    } catch (err) {
      setErrorLocal(err instanceof ApiError ? err.message : 'No se pudo subir el documento.')
    } finally {
      setSubiendo(false)
    }
  }

  async function ver() {
    try {
      await abrirDocumento(nombre)
    } catch (err) {
      setErrorLocal(err instanceof ApiError ? err.message : 'No se pudo abrir el documento.')
    }
  }

  const mensaje = errorLocal ?? error

  return (
    <div>
      <p className="block text-sm font-medium text-slate-700">{etiqueta}</p>
      <div className="mt-1 flex flex-wrap items-center gap-2">
        <button
          type="button"
          disabled={subiendo}
          onClick={() => ref.current?.click()}
          className="rounded border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {subiendo ? 'Subiendo...' : nombre !== '' ? 'Cambiar PDF' : 'Seleccionar PDF'}
        </button>
        <span className="truncate text-sm text-slate-600">
          {nombre !== '' ? nombreDocumento(nombre) : `PDF de máximo ${limite} MB`}
        </span>
        {nombre !== '' && !subiendo && (
          <>
            <button type="button" onClick={() => void ver()} className="text-sm text-biess-700 underline">
              Ver
            </button>
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
          </>
        )}
      </div>
      <input ref={ref} type="file" accept="application/pdf" onChange={(e) => void elegir(e)} className="hidden" />
      {mensaje && <p className="mt-1 text-xs text-red-600">{mensaje}</p>}
    </div>
  )
}