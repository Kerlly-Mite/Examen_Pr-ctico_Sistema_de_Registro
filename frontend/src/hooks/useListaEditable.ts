import { useState } from 'react'
import type { Errores } from '../lib/lista'

// Formulario de un elemento + Agregar / Editar / Quitar sobre una lista que vive en el contexto
export function useListaEditable<F extends object>(
  vacio: F,
  validar: (f: F) => Errores<F>,
  items: readonly F[],
  alCambiar: (items: F[]) => void,
) {
  const [form, setForm] = useState<F>(vacio)
  const [editando, setEditando] = useState<number | null>(null)
  const [errores, setErrores] = useState<Errores<F>>({})

  function cambiar<K extends keyof F>(campo: K, valor: F[K]) {
    setForm((prev) => ({ ...prev, [campo]: valor }))
  }

  function cancelar() {
    setForm(vacio)
    setEditando(null)
    setErrores({})
  }

  // true si se agregó o actualizó; false si hay errores de validación
  function agregar(): boolean {
    const errs = validar(form)
    setErrores(errs)
    if (Object.values(errs).some((v) => v !== undefined)) return false
    alCambiar(editando === null ? [...items, form] : items.map((x, i) => (i === editando ? form : x)))
    cancelar()
    return true
  }

  function editar(indice: number) {
    const item = items[indice]
    if (item === undefined) return
    setForm({ ...item })
    setEditando(indice)
    setErrores({})
  }

  function quitar(indice: number) {
    alCambiar(items.filter((_, i) => i !== indice))
    if (editando === indice) cancelar()
    else if (editando !== null && indice < editando) setEditando(editando - 1)
  }

  function hayContenido(): boolean {
    return editando !== null || JSON.stringify(form) !== JSON.stringify(vacio)
  }

  return { form, editando, errores, cambiar, agregar, editar, cancelar, quitar, hayContenido }
}