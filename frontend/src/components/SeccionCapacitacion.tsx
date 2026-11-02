import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import { useListaEditable } from '../hooks/useListaEditable'
import { fechaHoyISO, formatearFecha } from '../lib/formato'
import { TIPOS_EVENTO, etiquetaDe } from '../lib/opciones'
import { CAPACITACION_VACIA, validarCapacitacion } from '../lib/secciones'
import type { Capacitacion } from '../types/postulacion'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'
import { CampoArchivo } from './CampoArchivo'
import { PanelLista } from './PanelLista'
import { SelectorTexto } from './SelectorTexto'

export function SeccionCapacitacion() {
  const { postulacion: p, actualizar } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()

  const lista = useListaEditable<Capacitacion>(CAPACITACION_VACIA, validarCapacitacion, p.capacitaciones, (items) => {
    actualizar({ capacitaciones: items })
    limpiar()
  })
  const { form, errores, cambiar } = lista

  function eliminar(indice: number) {
    const c = p.capacitaciones[indice]
    if (c === undefined || !window.confirm(`¿Eliminar "${c.nombre_evento}" de la lista?`)) return
    lista.quitar(indice)
  }

  const pendiente = lista.hayContenido() ? 'Tiene una capacitación sin agregar. Pulse Agregar o Cancelar.' : null

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold text-biess-800">Capacitación</h2>

      <div className="mt-4 space-y-4">
        <PanelLista
          nombreItem="capacitación"
          editando={lista.editando !== null}
          onAgregar={lista.agregar}
          onCancelar={lista.cancelar}
          items={p.capacitaciones}
          textoVacio="Aún no ha agregado capacitaciones."
          onEditar={lista.editar}
          onEliminar={eliminar}
          resumen={(c) => (
            <>
              <p className="font-medium">{c.nombre_evento}</p>
              <p className="text-slate-600">
                {etiquetaDe(TIPOS_EVENTO, c.tipo_evento)} · {c.institucion} · {formatearFecha(c.fecha_inicio)} al{' '}
                {formatearFecha(c.fecha_fin)} · {c.horas} horas · PDF: {c.documento_respaldo}
              </p>
            </>
          )}
          campos={
            <>
              <SelectorTexto
                etiqueta="Tipo de evento *"
                name="capacitacionTipo"
                opciones={TIPOS_EVENTO}
                valor={form.tipo_evento}
                onCambio={(v) => cambiar('tipo_evento', v)}
                error={errores.tipo_evento}
              />
              <Campo
                etiqueta="Nombre del evento *"
                name="capacitacionNombre"
                value={form.nombre_evento}
                onChange={(e) => cambiar('nombre_evento', e.target.value)}
                error={errores.nombre_evento}
              />
              <Campo
                etiqueta="Institución *"
                name="capacitacionInstitucion"
                value={form.institucion}
                onChange={(e) => cambiar('institucion', e.target.value)}
                error={errores.institucion}
              />
              <Campo
                etiqueta="Número de horas *"
                name="capacitacionHoras"
                inputMode="numeric"
                value={form.horas === 0 ? '' : String(form.horas)}
                onChange={(e) => cambiar('horas', Number(e.target.value.replace(/\D/g, '')))}
                error={errores.horas}
              />
              <Campo
                etiqueta="Fecha de inicio *"
                name="capacitacionInicio"
                type="date"
                max={fechaHoyISO()}
                value={form.fecha_inicio ?? ''}
                onChange={(e) => cambiar('fecha_inicio', e.target.value === '' ? null : e.target.value)}
                error={errores.fecha_inicio}
              />
              <Campo
                etiqueta="Fecha de fin *"
                name="capacitacionFin"
                type="date"
                max={fechaHoyISO()}
                value={form.fecha_fin ?? ''}
                onChange={(e) => cambiar('fecha_fin', e.target.value === '' ? null : e.target.value)}
                error={errores.fecha_fin}
              />
              <CampoArchivo
                etiqueta="Documento de respaldo *"
                maxMB={2}
                nombre={form.documento_respaldo}
                onCambio={(n) => cambiar('documento_respaldo', n)}
                error={errores.documento_respaldo}
              />
            </>
          }
        />

        <BarraGuardar aviso={aviso} guardando={guardando} onGuardar={() => void ejecutar({ pendiente })} />
      </div>
    </section>
  )
}