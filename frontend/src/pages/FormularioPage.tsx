import { Alerta } from '../components/Alerta'
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

function Contenido() {
  const { cargando, errorCarga, recargar, postulacion, opciones, bloqueo, enviada } = useFormulario()

  if (cargando) {
    return <p className="text-center text-slate-600">Cargando formulario...</p>
  }

  if (errorCarga !== null || postulacion === null || opciones === null) {
    return (
      <div className="space-y-3">
        <Alerta tipo="error">{errorCarga ?? 'No se pudo cargar el formulario.'}</Alerta>
        <button
          type="button"
          onClick={recargar}
          className="rounded border border-slate-300 px-4 py-2 text-sm hover:bg-slate-50"
        >
          Reintentar
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-xl font-semibold">Formulario de postulación</h1>
        <p className="text-sm text-slate-600">
          Código: <strong>{postulacion.codigo !== '' ? postulacion.codigo : 'se asigna al guardar'}</strong> · Estado:{' '}
          <strong>{formatearTexto(postulacion.estado)}</strong>
        </p>
        <p className="text-sm text-slate-600">
          Los campos con * son obligatorios. Puede guardar su progreso y continuar después.
        </p>
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

      {bloqueo ? (
        <SeccionPersonal />
      ) : (
        <>
          <fieldset disabled={enviada} className="min-w-0 space-y-6">
            <SeccionPersonal />
            <SeccionDomicilio />
            <SeccionInformacionAdicional />
            <SeccionCargo />
            <SeccionFormacion />
            <SeccionExperiencia />
            <SeccionPublicaciones />
            <SeccionCapacitacion />
            <SeccionInhabilidades />
          </fieldset>
          <SeccionEnvio />
        </>
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