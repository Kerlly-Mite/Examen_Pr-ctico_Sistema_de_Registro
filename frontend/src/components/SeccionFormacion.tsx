import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import { useListaEditable } from '../hooks/useListaEditable'
import { fechaHoyISO, formatearFecha } from '../lib/formato'
import { AREAS, NIVELES, TIPOS_INSTITUCION, etiquetaDe } from '../lib/opciones'
import { FORMACION_VACIA, validarFormacion } from '../lib/secciones'
import type { Formacion } from '../types/postulacion'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'
import { CampoArchivo } from './CampoArchivo'
import { PanelLista } from './PanelLista'
import { SelectorTexto } from './SelectorTexto'

export function SeccionFormacion() {
  const { postulacion: p, actualizar } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()

  const lista = useListaEditable<Formacion>(FORMACION_VACIA, validarFormacion, p.formacion, (items) => {
    actualizar({ formacion: items })
    limpiar()
  })
  const { form, errores, cambiar } = lista

  function eliminar(indice: number) {
    const f = p.formacion[indice]
    if (f === undefined || !window.confirm(`¿Eliminar "${f.titulo}" de la lista?`)) return
    lista.quitar(indice)
  }

  const pendiente = lista.hayContenido() ? 'Tiene un título sin agregar. Pulse Agregar o Cancelar.' : null

  return (
    <section className="rounded-lg bg-white p-6 shadow">
      <h2 className="text-lg font-semibold">Formación académica</h2>

      <div className="mt-4 space-y-4">
        <PanelLista
          nombreItem="título académico"
          editando={lista.editando !== null}
          onAgregar={lista.agregar}
          onCancelar={lista.cancelar}
          items={p.formacion}
          textoVacio="Aún no ha agregado títulos."
          onEditar={lista.editar}
          onEliminar={eliminar}
          resumen={(f) => (
            <>
              <p className="font-medium">{f.titulo}</p>
              <p className="text-slate-600">
                {etiquetaDe(NIVELES, f.nivel_instruccion)} · {f.institucion} ({etiquetaDe(TIPOS_INSTITUCION, f.tipo)}) ·{' '}
                {etiquetaDe(AREAS, f.area_estudios)} · SENESCYT {f.registro_senescyt} (
                {formatearFecha(f.fecha_registro_senescyt)}) · PDF: {f.documento_respaldo}
              </p>
            </>
          )}
          campos={
            <>
              <SelectorTexto
                etiqueta="Nivel de instrucción *"
                name="formacionNivel"
                opciones={NIVELES}
                valor={form.nivel_instruccion}
                onCambio={(v) => cambiar('nivel_instruccion', v)}
                error={errores.nivel_instruccion}
              />
              <Campo
                etiqueta="Institución educativa *"
                name="formacionInstitucion"
                value={form.institucion}
                onChange={(e) => cambiar('institucion', e.target.value)}
                error={errores.institucion}
              />
              <SelectorTexto
                etiqueta="Tipo *"
                name="formacionTipo"
                opciones={TIPOS_INSTITUCION}
                valor={form.tipo}
                onCambio={(v) => cambiar('tipo', v)}
                error={errores.tipo}
              />
              <Campo
                etiqueta="Título obtenido *"
                name="formacionTitulo"
                value={form.titulo}
                onChange={(e) => cambiar('titulo', e.target.value)}
                error={errores.titulo}
              />
              <SelectorTexto
                etiqueta="Área de estudios *"
                name="formacionArea"
                opciones={AREAS}
                valor={form.area_estudios}
                onCambio={(v) => cambiar('area_estudios', v)}
                error={errores.area_estudios}
              />
              <Campo
                etiqueta="Número de registro SENESCYT *"
                name="formacionRegistro"
                value={form.registro_senescyt}
                onChange={(e) => cambiar('registro_senescyt', e.target.value)}
                error={errores.registro_senescyt}
              />
              <Campo
                etiqueta="Fecha de registro SENESCYT *"
                name="formacionFecha"
                type="date"
                max={fechaHoyISO()}
                value={form.fecha_registro_senescyt ?? ''}
                onChange={(e) => cambiar('fecha_registro_senescyt', e.target.value === '' ? null : e.target.value)}
                error={errores.fecha_registro_senescyt}
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