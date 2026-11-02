import { useState } from 'react'
import type { ReactNode } from 'react'
import { Alerta } from '../components/Alerta'
import { Boton } from '../components/Boton'
import { Cargando } from '../components/Cargando'
import { SeccionCapacitacion } from '../components/SeccionCapacitacion'
import { SeccionCargo } from '../components/SeccionCargo'
import { SeccionDomicilio } from '../components/SeccionDomicilio'
import { SeccionEnvio } from '../components/SeccionEnvio'
import { SeccionExperiencia } from '../components/SeccionExperiencia'
import { SeccionFormacion } from '../components/SeccionFormacion'
import { SeccionInformacionAdicional } from '../components/SeccionInformacionAdicional'
import { SeccionInhabilidades } from '../components/SeccionInhabilidades'
import { SeccionPersonal } from '../components/SeccionPersonal'
import { SeccionPublicaciones } from '../components/SeccionPublicaciones'
import { FormularioProvider } from '../context/FormularioProvider'
import { useFormulario } from '../context/useFormulario'
import { formatearTexto } from '../lib/formato'
import { estadoSecciones } from '../lib/progreso'

const VISTAS: Record<string, ReactNode> = {
  personal: <SeccionPersonal />,
  domicilio: <SeccionDomicilio />,
  familiares: <SeccionInformacionAdicional />,
  cargo: <SeccionCargo />,
  formacion: <SeccionFormacion />,
  experiencia: <SeccionExperiencia />,
  publicaciones: <SeccionPublicaciones />,
  capacitacion: <SeccionCapacitacion />,
  inhabilidades: <SeccionInhabilidades />,
  envio: <SeccionEnvio />,
}

function Contenido() {
  const { cargando, errorCarga, recargar, postulacion, opciones, bloqueo, enviada } = useFormulario()
  const [actual, setActual] = useState(0)

  if (cargando) {
    return <Cargando texto="Cargando formulario..." />
  }

  if (errorCarga !== null || postulacion === null || opciones === null) {
    return (
      <div className="mx-auto max-w-xl space-y-3">
        <Alerta tipo="error">{errorCarga ?? 'No se pudo cargar el formulario.'}</Alerta>
        <Boton type="button" variante="borde" autoAncho onClick={recargar}>
          Reintentar
        </Boton>
      </div>
    )
  }

  const secciones = estadoSecciones(postulacion, opciones)
  const total = secciones.length
  const ultima = total - 1
  const completas = secciones.filter((s) => s.completa).length

  function ir(indice: number) {
    setActual(Math.min(Math.max(indice, 0), ultima))
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold text-biess-800">Formulario de postulación</h1>
        <p className="text-sm text-slate-600">
          Código: <strong>{postulacion.codigo !== '' ? postulacion.codigo : 'se asigna al guardar'}</strong> ·
          Estado: <strong>{formatearTexto(postulacion.estado)}</strong>
        </p>
        <p className="text-sm text-slate-600">Los campos con * son obligatorios.</p>
      </header>

      {bloqueo && (
        <Alerta tipo="error">
          <p className="font-medium">{bloqueo.mensaje}</p>
          <p className="mt-1">Su postulación no puede continuar.</p>
          {bloqueo.codigo === 'cargo_no_habilitado' && (
            <p className="mt-1">Si desea elegir otro cargo, recargue la página.</p>
          )}
        </Alerta>
      )}
      {enviada && <Alerta tipo="info">Su postulación ya fue enviada y no puede modificarse.</Alerta>}
      {!bloqueo && !enviada && (
        <Alerta tipo="aviso">Si tu nacionalidad no es ecuatoriana, la postulación se cierra.</Alerta>
      )}

      {bloqueo ? (
        <SeccionPersonal />
      ) : (
        <div className="grid gap-6 lg:grid-cols-[16rem_1fr]">
          <aside className="h-fit rounded-lg border border-slate-200 bg-white p-4 shadow-sm print:hidden lg:sticky lg:top-4">
            <p className="text-sm font-semibold text-biess-800">Secciones</p>
            <p className="mt-1 text-xs text-slate-600">
              {completas} de {total} completas
            </p>
            <div className="mt-2 h-1.5 w-full rounded bg-slate-200">
              <div
                className="h-1.5 rounded bg-oro-500 transition-all"
                style={{ width: `${(completas / total) * 100}%` }}
              />
            </div>
            <ol className="mt-3 flex gap-1 overflow-x-auto lg:flex-col lg:overflow-visible">
              {secciones.map((s, i) => (
                <li key={s.id}>
                  <button
                    type="button"
                    onClick={() => ir(i)}
                    aria-current={i === actual ? 'step' : undefined}
                    className={`flex w-full items-center gap-2 whitespace-nowrap rounded px-2 py-1.5 text-left text-sm ${
                      i === actual
                        ? 'bg-biess-50 font-semibold text-biess-800'
                        : 'text-slate-700 hover:bg-slate-50'
                    }`}
                  >
                    <span className="w-5 text-xs text-slate-500">{i + 1}.</span>
                    <span className="flex-1">{s.titulo}</span>
                    {s.completa && (
                      <span className="text-green-700" aria-label="completa">
                        ✓
                      </span>
                    )}
                  </button>
                </li>
              ))}
            </ol>
          </aside>

          <div className="min-w-0 space-y-4">
            <fieldset disabled={enviada} className="min-w-0">
              {secciones.map((s, i) =>
                s.id === 'envio' ? null : (
                  <div key={s.id} className={i === actual ? 'block' : 'hidden print:block'}>
                    {VISTAS[s.id]}
                  </div>
                ),
              )}
            </fieldset>
            <div className={actual === ultima ? 'block' : 'hidden print:block'}>{VISTAS.envio}</div>

            <div className="flex items-center justify-between print:hidden">
              <Boton type="button" variante="borde" autoAncho disabled={actual === 0} onClick={() => ir(actual - 1)}>
                Anterior
              </Boton>
              {actual < ultima && (
                <Boton type="button" variante="oro" autoAncho onClick={() => ir(actual + 1)}>
                  Siguiente
                </Boton>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export function FormularioPage() {
  return (
    <FormularioProvider>
      <Contenido />
    </FormularioProvider>
  )
}