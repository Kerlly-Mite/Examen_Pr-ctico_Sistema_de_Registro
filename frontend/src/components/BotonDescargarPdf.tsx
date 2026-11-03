import { useState } from 'react'
import { ApiError } from '../api/client'
import { descargarFormularioPdf } from '../api/documentos'
import { useFormularioListo } from '../context/useFormulario'
import { Alerta } from './Alerta'
import { Boton } from './Boton'

export function BotonDescargarPdf() {
  const { postulacion: p } = useFormularioListo()
  const [descargando, setDescargando] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function descargar() {
    setError(null)
    setDescargando(true)
    try {
      await descargarFormularioPdf(`formulario-${p.codigo !== '' ? p.codigo : 'postulacion'}.pdf`)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo descargar el formulario.')
    } finally {
      setDescargando(false)
    }
  }

  return (
    <div className="space-y-2 print:hidden">
      <div className="w-56">
        <Boton type="button" variante="borde" cargando={descargando} onClick={() => void descargar()}>
          Descargar formulario PDF
        </Boton>
      </div>
      {error && <Alerta tipo="error">{error}</Alerta>}
    </div>
  )
}