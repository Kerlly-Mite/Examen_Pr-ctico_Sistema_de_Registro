import { useState } from 'react'
import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import { useListaEditable } from '../hooks/useListaEditable'
import { INSTITUCIONES, PARENTESCOS, etiquetaDe } from '../lib/opciones'
import { FAMILIAR_VACIO, validarFamiliar } from '../lib/secciones'
import type { Familiar } from '../types/postulacion'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'
import { GrupoSiNo } from './GrupoSiNo'
import { PanelLista } from './PanelLista'
import { SelectorTexto } from './SelectorTexto'

export function SeccionInformacionAdicional() {
  const { postulacion: p, actualizar } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()
  const [respondido, setRespondido] = useState(p.estado !== 'NUEVA')

  const lista = useListaEditable<Familiar>(FAMILIAR_VACIO, validarFamiliar, p.familiares, (items) => {
    actualizar({ familiares: items })
    limpiar()
  })
  const { form, errores, cambiar } = lista

  function cambiarRespuesta(valor: boolean) {
    if (
      !valor &&
      p.familiares.length > 0 &&
      !window.confirm('Se eliminarán los familiares agregados. ¿Desea continuar?')
    ) {
      return
    }
    setRespondido(true)
    actualizar({ tiene_familiares_iess: valor, familiares: valor ? p.familiares : [] })
    if (!valor) lista.cancelar()
    limpiar()
  }

  function eliminar(indice: number) {
    const f = p.familiares[indice]
    if (f === undefined || !window.confirm(`¿Eliminar a ${f.nombres_apellidos} de la lista?`)) return
    lista.quitar(indice)
  }

  const pendiente =
    p.tiene_familiares_iess && lista.hayContenido()
      ? 'Tiene un familiar sin agregar. Pulse Agregar o Cancelar.'
      : null

  return (
    <section className="rounded-lg bg-white p-6 shadow">
      <h2 className="text-lg font-semibold">Información adicional</h2>

      <div className="mt-4 space-y-4">
        <GrupoSiNo
          name="tieneFamiliares"
          etiqueta="¿Tiene familiares dentro del segundo grado de consanguinidad y cuarto de afinidad trabajando actualmente en el IESS o BIESS? *"
          valor={respondido ? p.tiene_familiares_iess : null}
          onCambio={cambiarRespuesta}
        />

        {respondido && p.tiene_familiares_iess && (
          <PanelLista
            nombreItem="familiar"
            editando={lista.editando !== null}
            onAgregar={lista.agregar}
            onCancelar={lista.cancelar}
            items={p.familiares}
            textoVacio="Aún no ha agregado familiares."
            onEditar={lista.editar}
            onEliminar={eliminar}
            resumen={(f) => (
              <>
                <p className="font-medium">{f.nombres_apellidos}</p>
                <p className="text-slate-600">
                  {etiquetaDe(PARENTESCOS, f.parentesco)} · {f.institucion} · {f.area} · {f.cargo}
                </p>
              </>
            )}
            campos={
              <>
                <Campo
                  etiqueta="Nombres y apellidos del familiar *"
                  name="familiarNombres"
                  value={form.nombres_apellidos}
                  onChange={(e) => cambiar('nombres_apellidos', e.target.value)}
                  error={errores.nombres_apellidos}
                />
                <SelectorTexto
                  etiqueta="Parentesco *"
                  name="familiarParentesco"
                  opciones={PARENTESCOS}
                  valor={form.parentesco}
                  onCambio={(v) => cambiar('parentesco', v)}
                  error={errores.parentesco}
                />
                <SelectorTexto
                  etiqueta="Institución donde trabaja *"
                  name="familiarInstitucion"
                  opciones={INSTITUCIONES}
                  valor={form.institucion}
                  onCambio={(v) => cambiar('institucion', v)}
                  error={errores.institucion}
                />
                <Campo
                  etiqueta="Área donde trabaja *"
                  name="familiarArea"
                  value={form.area}
                  onChange={(e) => cambiar('area', e.target.value)}
                  error={errores.area}
                />
                <Campo
                  etiqueta="Cargo *"
                  name="familiarCargo"
                  value={form.cargo}
                  onChange={(e) => cambiar('cargo', e.target.value)}
                  error={errores.cargo}
                />
              </>
            }
          />
        )}

        <BarraGuardar aviso={aviso} guardando={guardando} onGuardar={() => void ejecutar({ pendiente })} />
      </div>
    </section>
  )
}