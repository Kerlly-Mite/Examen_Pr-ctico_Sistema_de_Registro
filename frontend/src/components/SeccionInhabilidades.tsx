import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import { fechaHoyISO } from '../lib/formato'
import type { RespuestaInhabilidad } from '../types/postulacion'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'
import { GrupoSiNo } from './GrupoSiNo'

const CODIGO_REMOVIDO = '1.9.7'

export function SeccionInhabilidades() {
  const { postulacion: p, opciones, actualizar } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()
  const respuestas = p.inhabilidades

  function buscar(codigo: string): RespuestaInhabilidad | undefined {
    return respuestas.find((r) => r.codigo === codigo)
  }

  function responder(codigo: string, valor: boolean) {
    const previa = buscar(codigo)
    const resto = respuestas.filter((r) => r.codigo !== codigo)
    const nueva: RespuestaInhabilidad =
      codigo === CODIGO_REMOVIDO && valor && previa ? { ...previa, respuesta: true } : { codigo, respuesta: valor }
    actualizar({ inhabilidades: [...resto, nueva] })
    limpiar()
  }

  function cambiarRemocion(parcial: Partial<RespuestaInhabilidad>) {
    actualizar({
      inhabilidades: respuestas.map((r) => (r.codigo === CODIGO_REMOVIDO ? { ...r, ...parcial } : r)),
    })
    limpiar()
  }

  const removido = buscar(CODIGO_REMOVIDO)
  const pendiente =
    removido?.respuesta === true &&
    ((removido.remocion_institucion ?? '').trim() === '' || (removido.remocion_motivo ?? '').trim() === '')
      ? 'Complete la institución y el motivo de finalización de la pregunta 1.9.7.'
      : null

  return (
    <section className="rounded-lg bg-white p-6 shadow">
      <h2 className="text-lg font-semibold">Inhabilidades</h2>

      <div className="mt-4 space-y-5">
        {opciones.inhabilidades.map((q) => {
          const r = buscar(q.codigo)
          const inhabilitante = r !== undefined && r.respuesta !== q.respuesta_habilitante
          return (
            <div key={q.codigo} className="space-y-3">
              <GrupoSiNo
                name={`inhabilidad-${q.codigo}`}
                etiqueta={`${q.codigo} ${q.texto} *`}
                valor={r === undefined ? null : r.respuesta}
                onCambio={(v) => responder(q.codigo, v)}
              />
              {inhabilitante && (
                <p className="text-xs text-amber-700">
                  Con esta respuesta no podrá enviar la postulación (impedimento según las reglas del concurso).
                </p>
              )}

              {q.codigo === CODIGO_REMOVIDO && r?.respuesta === true && (
                <div className="grid gap-4 rounded border border-slate-200 p-4 md:grid-cols-2">
                  <Campo
                    etiqueta="Institución *"
                    name="removidoInstitucion"
                    value={r.remocion_institucion ?? ''}
                    onChange={(e) => cambiarRemocion({ remocion_institucion: e.target.value })}
                  />
                  <Campo
                    etiqueta="Cargo"
                    name="removidoCargo"
                    value={r.remocion_cargo ?? ''}
                    onChange={(e) => cambiarRemocion({ remocion_cargo: e.target.value })}
                  />
                  <Campo
                    etiqueta="Fecha de finalización"
                    name="removidoFecha"
                    type="date"
                    max={fechaHoyISO()}
                    value={r.remocion_fecha_fin ?? ''}
                    onChange={(e) =>
                      cambiarRemocion({ remocion_fecha_fin: e.target.value === '' ? null : e.target.value })
                    }
                  />
                  <Campo
                    etiqueta="Motivo de finalización *"
                    name="removidoMotivo"
                    value={r.remocion_motivo ?? ''}
                    onChange={(e) => cambiarRemocion({ remocion_motivo: e.target.value })}
                  />
                </div>
              )}
            </div>
          )
        })}

        <BarraGuardar aviso={aviso} guardando={guardando} onGuardar={() => void ejecutar({ pendiente })} />
      </div>
    </section>
  )
}