import { SeccionDomicilio } from '../components/SeccionDomicilio'
import { SeccionInformacionAdicional } from '../components/SeccionInformacionAdicional'

export function FormularioPage() {
  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-xl font-semibold">Formulario de postulación</h1>
        <p className="mt-1 text-sm text-slate-600">
          Los campos con * son obligatorios. Puede guardar su progreso y continuar después.
        </p>
      </header>

      <SeccionDomicilio />
      <SeccionInformacionAdicional />
    </div>
  )
}