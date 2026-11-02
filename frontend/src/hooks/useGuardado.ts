import { useCallback, useState } from 'react'
import { useFormularioListo } from '../context/useFormulario'
import type { Aviso } from '../lib/lista'

interface OpcionesGuardado {
  pendiente?: string | null
  exito?: string
  tipoExito?: 'exito' | 'info'
}

// Guarda TODO el formulario (el servidor recibe el formulario completo) y deja el aviso de la sección
export function useGuardado() {
  const { guardar, guardando } = useFormularioListo()
  const [aviso, setAviso] = useState<Aviso | null>(null)

  const ejecutar = useCallback(
    async (op: OpcionesGuardado = {}) => {
      if (op.pendiente) {
        setAviso({ tipo: 'error', texto: op.pendiente })
        return
      }
      const r = await guardar()
      setAviso(
        r.ok
          ? { tipo: op.tipoExito ?? 'exito', texto: op.exito ?? 'Progreso guardado.' }
          : { tipo: 'error', texto: r.mensaje },
      )
    },
    [guardar],
  )

  const limpiar = useCallback(() => setAviso(null), [])

  return { aviso, guardando, ejecutar, limpiar }
}