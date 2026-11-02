import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import { formatearFecha, formatearTexto } from '../lib/formato'
import { requierePareja } from '../lib/postulacion'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'

export function SeccionPersonal() {
  const { postulacion: p, actualizar } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()
  const pareja = requierePareja(p.estado_civil)

  function pendiente(): string | null {
    if (!pareja) return null
    if (p.pareja_nombres.trim() === '' && p.pareja_cedula === '') return null
    if (!/^\d{10}$/.test(p.pareja_cedula)) return 'La cédula de la pareja debe tener 10 dígitos.'
    if (p.pareja_nombres.trim() === '') return 'Ingrese los nombres y apellidos de la pareja.'
    return null
  }

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold text-biess-800">Información personal</h2>
      <p className="mt-1 text-sm text-slate-600">Estos datos vienen del Registro Civil y no pueden modificarse.</p>

      <div className="mt-4 space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <Campo etiqueta="Nombres" name="nombres" value={p.nombres} readOnly />
          <Campo etiqueta="Apellidos" name="apellidos" value={p.apellidos} readOnly />
          <Campo etiqueta="Cédula" name="cedula" value={p.cedula} readOnly />
          <Campo etiqueta="Género" name="genero" value={formatearTexto(p.genero)} readOnly />
          <Campo etiqueta="Fecha de nacimiento" name="fechaNacimiento" value={formatearFecha(p.fecha_nacimiento)} readOnly />
          <Campo etiqueta="Nacionalidad" name="nacionalidad" value={formatearTexto(p.nacionalidad)} readOnly />
          <Campo etiqueta="Estado civil" name="estadoCivil" value={formatearTexto(p.estado_civil)} readOnly />
        </div>

        {pareja && (
          <div className="grid gap-4 rounded border border-slate-200 p-4 md:grid-cols-2">
            <Campo
              etiqueta="Nombres y apellidos de la pareja"
              name="parejaNombres"
              value={p.pareja_nombres}
              onChange={(e) => {
                actualizar({ pareja_nombres: e.target.value })
                limpiar()
              }}
            />
            <Campo
              etiqueta="Cédula de la pareja"
              name="parejaCedula"
              inputMode="numeric"
              maxLength={10}
              value={p.pareja_cedula}
              onChange={(e) => {
                actualizar({ pareja_cedula: e.target.value.replace(/\D/g, '') })
                limpiar()
              }}
            />
          </div>
        )}

        <BarraGuardar aviso={aviso} guardando={guardando} onGuardar={() => void ejecutar({ pendiente: pendiente() })} />
      </div>
    </section>
  )
}