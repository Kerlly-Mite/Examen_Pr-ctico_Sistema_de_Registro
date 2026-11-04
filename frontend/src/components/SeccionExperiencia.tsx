import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import { useListaEditable } from '../hooks/useListaEditable'
import { MAX_DESCRIPCION, calcularDuracion, formatearDuracion } from '../lib/experiencia'
import { fechaHoyISO, formatearFecha } from '../lib/formato'
import { AREAS, TIPOS_EXPERIENCIA, etiquetaDe } from '../lib/opciones'
import { EXPERIENCIA_VACIA, validarExperiencia } from '../lib/secciones'
import type { Duracion, Experiencia } from '../types/postulacion'
import { AreaTexto } from './AreaTexto'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'
import { nombreDocumento } from '../api/documentos'
import { CampoArchivo } from './CampoArchivo'
import { PanelLista } from './PanelLista'
import { SelectorTexto } from './SelectorTexto'

// Usa el tiempo que calculó el servidor; si todavía no se guardó, una vista previa local
function textoDuracion(d: Duracion | null | undefined, inicio: string | null, fin: string | null): string | null {
  if (d) return formatearDuracion(d)
  if (inicio === null || fin === null || fin < inicio) return null
  const calculada = calcularDuracion(inicio, fin)
  return calculada === null ? null : formatearDuracion(calculada)
}

export function SeccionExperiencia() {
  const { postulacion: p, actualizar } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()

  const lista = useListaEditable<Experiencia>(EXPERIENCIA_VACIA, validarExperiencia, p.experiencia, (items) => {
    actualizar({ experiencia: items })
    limpiar()
  })
  const { form, errores, cambiar } = lista

  function eliminar(indice: number) {
    const e = p.experiencia[indice]
    if (e === undefined || !window.confirm(`¿Eliminar la experiencia en "${e.institucion}"?`)) return
    lista.quitar(indice)
  }

  const fechaFutura = [form.fecha_inicio, form.fecha_fin].some((f) => f !== null && f > fechaHoyISO())
  const vistaPrevia = fechaFutura ? null : textoDuracion(null, form.fecha_inicio, form.fecha_fin)
  const pendiente = lista.hayContenido() ? 'Tiene una experiencia sin agregar. Pulse Agregar o Cancelar.' : null

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold text-biess-800">Experiencia profesional</h2>

      <div className="mt-4 space-y-4">
        <PanelLista
          nombreItem="experiencia"
          editando={lista.editando !== null}
          onAgregar={lista.agregar}
          onCancelar={lista.cancelar}
          items={p.experiencia}
          textoVacio="Aún no ha agregado experiencias."
          onEditar={lista.editar}
          onEliminar={eliminar}
          resumen={(e) => (
            <>
              <p className="font-medium">
                {e.cargo} · {e.institucion}
              </p>
              <p className="text-slate-600">
                {etiquetaDe(TIPOS_EXPERIENCIA, e.tipo)} · {etiquetaDe(AREAS, e.actividades)} ·{' '}
                {formatearFecha(e.fecha_inicio)} al {formatearFecha(e.fecha_fin)}
              </p>
              <p className="text-slate-600">
                Tiempo: {textoDuracion(e.duracion, e.fecha_inicio, e.fecha_fin) ?? 'no calculable'} · PDF:{' '}
                {nombreDocumento(e.documento_respaldo)}
              </p>
            </>
          )}
          campos={
            <>
              <SelectorTexto
                etiqueta="Tipo de experiencia *"
                name="experienciaTipo"
                opciones={TIPOS_EXPERIENCIA}
                valor={form.tipo}
                onCambio={(v) => cambiar('tipo', v)}
                error={errores.tipo}
              />
              <Campo
                etiqueta="Cargo o puesto *"
                name="experienciaCargo"
                value={form.cargo}
                onChange={(e) => cambiar('cargo', e.target.value)}
                error={errores.cargo}
              />
              <Campo
                etiqueta="Institución o empresa *"
                name="experienciaInstitucion"
                value={form.institucion}
                onChange={(e) => cambiar('institucion', e.target.value)}
                error={errores.institucion}
              />
              <SelectorTexto
                etiqueta="Actividades relacionadas a *"
                name="experienciaActividades"
                opciones={AREAS}
                valor={form.actividades}
                onCambio={(v) => cambiar('actividades', v)}
                error={errores.actividades}
              />
              <Campo
                etiqueta="Fecha de inicio *"
                name="experienciaInicio"
                type="date"
                max={fechaHoyISO()}
                value={form.fecha_inicio ?? ''}
                onChange={(e) => cambiar('fecha_inicio', e.target.value === '' ? null : e.target.value)}
                error={errores.fecha_inicio}
              />
              <Campo
                etiqueta="Fecha de finalización *"
                name="experienciaFin"
                type="date"
                max={fechaHoyISO()}
                value={form.fecha_fin ?? ''}
                onChange={(e) => cambiar('fecha_fin', e.target.value === '' ? null : e.target.value)}
                error={errores.fecha_fin}
              />
              {vistaPrevia && (
                <p className="text-sm text-slate-600 md:col-span-2">
                  Tiempo de experiencia: <strong>{vistaPrevia}</strong> (el servidor lo confirma al guardar)
                </p>
              )}
              <div className="md:col-span-2">
                <AreaTexto
                  etiqueta="Descripción de funciones *"
                  name="experienciaDescripcion"
                  valor={form.descripcion_funciones}
                  onCambio={(v) => cambiar('descripcion_funciones', v)}
                  maxLength={MAX_DESCRIPCION}
                  error={errores.descripcion_funciones}
                />
              </div>
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

        {p.experiencia_total && p.experiencia.length > 0 && (
          <p className="text-sm text-slate-700">
            Experiencia total (sin contar dos veces los periodos que se cruzan):{' '}
            <strong>{formatearDuracion(p.experiencia_total)}</strong>
          </p>
        )}

        <BarraGuardar aviso={aviso} guardando={guardando} onGuardar={() => void ejecutar({ pendiente })} />
      </div>
    </section>
  )
}