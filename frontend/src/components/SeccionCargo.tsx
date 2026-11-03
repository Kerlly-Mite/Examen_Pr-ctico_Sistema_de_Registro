import { useRef, useState } from 'react'
import { verificarCargo } from '../api/postulacion'
import { ApiError } from '../api/client'
import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import type { VerificacionCargo } from '../types/postulacion'
import { Alerta } from './Alerta'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'
import { CampoArchivo } from './CampoArchivo'
import { GrupoSiNo } from './GrupoSiNo'
import { SelectorTexto } from './SelectorTexto'

const PREGUNTA_RESPALDO =
  '¿Se encuentra respaldado por una organización legalmente constituida ante autoridad competente?'

export function SeccionCargo() {
  const { postulacion: p, opciones, actualizar, bloquear } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()
  const [verif, setVerif] = useState<VerificacionCargo | null>(p.verificacion_cargo)
  const [consultando, setConsultando] = useState(false)
  const [errorVerif, setErrorVerif] = useState<string | null>(null)
  const peticion = useRef(0)

  async function elegir(codigo: string) {
    actualizar({
      cargo_codigo: codigo,
      respaldado_organizacion: null,
      organizacion_nombre: '',
      organizacion_documento: '',
    })
    limpiar()
    setVerif(null)
    setErrorVerif(null)
    peticion.current += 1
    const numero = peticion.current

    if (codigo === '') {
      setConsultando(false)
      return
    }

    setConsultando(true)
    try {
      const r = await verificarCargo(codigo)
      if (numero !== peticion.current) return
      setVerif(r)
      if (!r.cumple) bloquear({ codigo: 'cargo_no_habilitado', mensaje: r.mensaje })
    } catch (e) {
      if (numero === peticion.current) {
        setErrorVerif(e instanceof ApiError ? e.message : 'No se pudo verificar el cargo.')
      }
    } finally {
      if (numero === peticion.current) setConsultando(false)
    }
  }

  const mostrarRespaldo = p.cargo_codigo !== '' && !consultando && (verif === null || verif.cumple)

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold text-biess-800">Selección del cargo al que postula</h2>

      <div className="mt-4 space-y-4">
        <SelectorTexto
          etiqueta="Cargo *"
          name="cargo"
          opciones={opciones.cargos.map((c) => ({ valor: c.codigo, etiqueta: c.nombre }))}
          valor={p.cargo_codigo}
          onCambio={(v) => {
            void elegir(v)
          }}
        />

        {consultando && <p className="text-sm text-slate-600">Verificando su calidad en el Historial Laboral...</p>}
        {errorVerif && <Alerta tipo="error">{errorVerif}</Alerta>}
        {verif?.cumple && <Alerta tipo="exito">{verif.estado}</Alerta>}

        {mostrarRespaldo && (
          <>
            <GrupoSiNo
              name="respaldoOrganizacion"
              etiqueta={`${verif?.pregunta ?? PREGUNTA_RESPALDO} *`}
              valor={p.respaldado_organizacion}
              onCambio={(v) => {
                actualizar(
                  v
                    ? { respaldado_organizacion: true }
                    : { respaldado_organizacion: false, organizacion_nombre: '', organizacion_documento: '' },
                )
                limpiar()
              }}
            />

            {p.respaldado_organizacion === true && (
              <div className="grid gap-4 rounded border border-slate-200 p-4 md:grid-cols-2">
                <Campo
                  etiqueta="Nombre de la organización *"
                  name="organizacionNombre"
                  value={p.organizacion_nombre}
                  onChange={(e) => {
                    actualizar({ organizacion_nombre: e.target.value })
                    limpiar()
                  }}
                />
                <CampoArchivo
                  etiqueta="Documento de respaldo *"
                  maxMB={4}
                  nombre={p.organizacion_documento} tipo="organizacion"
                  onCambio={(n) => {
                    actualizar({ organizacion_documento: n })
                    limpiar()
                  }}
                />
              </div>
            )}
          </>
        )}

        <BarraGuardar aviso={aviso} guardando={guardando} onGuardar={() => void ejecutar()} />
      </div>
    </section>
  )
}