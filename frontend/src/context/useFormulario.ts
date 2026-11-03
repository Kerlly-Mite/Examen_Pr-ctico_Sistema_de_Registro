import { useContext } from 'react'
import { FormularioContext } from './formulario-context'

export function useFormulario() {
  const contexto = useContext(FormularioContext)
  if (contexto === null) {
    throw new Error('useFormulario debe usarse dentro de <FormularioProvider>')
  }
  return contexto
}

// Para las secciones: se renderizan solo cuando el formulario ya cargó
export function useFormularioListo() {
  const c = useFormulario()
  if (c.postulacion === null || c.opciones === null) {
    throw new Error('El formulario aún no está cargado')
  }
  return { ...c, postulacion: c.postulacion, opciones: c.opciones }
}