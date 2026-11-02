import { useState } from 'react'
import { useFormularioListo } from '../context/useFormulario'
import type { Aviso } from '../lib/lista'
import { calcularPendientes } from '../lib/postulacion'
import { Alerta } from './Alerta'
import { Boton } from './Boton'

const DECLARACION = [
  'Autorizo de manera libre y voluntaria, emitiendo consentimiento explícito al Instituto Ecuatoriano de Seguridad Social, para el uso de mis datos personales dentro del concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores, de conformidad con lo prescrito en el artículo 7 de la Ley Orgánica de Protección de Datos Personales.',
  'Declaro que la información del presente formulario y documentación adjunta corresponde a la verdad y autorizo al Instituto Ecuatoriano de Seguridad Social para que se compruebe ante cualquier organismo público o privado su veracidad.',
  'Acepto expresamente cumplir con las normas aplicables al concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores; así como las resoluciones y disposiciones impartidas por el Instituto Ecuatoriano de Seguridad Social. En caso de incumplimiento de requisitos o encontrarme inmerso/a en las prohibiciones o inhabilidades acepto mi descalificación del proceso.',
  'Acepto que mi nombre como postulante al concurso de méritos y oposición para la selección de los Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los Afiliados, Jubilados y Empleadores; sea publicado en la página web del IESS - BIESS.',
  'El presente formulario debe ser impreso y con firma manuscrita, entregado junto con los requisitos solicitados, en el lugar, fecha y horario establecido por el Instituto Ecuatoriano de Seguridad Social.',
]

export function SeccionEnvio() {
  const { postulacion: p, opciones, guardar, enviar, guardando, enviada } = useFormularioListo()
  const [acepto, setAcepto] = useState(false)
  const [aviso, setAviso] = useState<Aviso | null>(null)
  const { obligatorios, recomendaciones } = calcularPendientes(p, opciones)

  async function alGuardar() {
    const r = await guardar()
    setAviso(r.ok ? { tipo: 'exito', texto: 'Postulación guardada.' } : { tipo: 'error', texto: r.mensaje })
  }

  async function alEnviar() {
    if (!acepto) {
      setAviso({ tipo: 'error', texto: 'Debe aceptar la declaración para enviar la postulación.' })
      return
    }
    const g = await guardar()
    if (!g.ok) {
      setAviso({ tipo: 'error', texto: g.mensaje })
      return
    }
    const r = await enviar(true)
    setAviso(
      r.ok
        ? { tipo: 'exito', texto: 'Postulación enviada correctamente.' }
        : { tipo: 'error', texto: r.mensaje },
    )
  }

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold text-biess-800">Declaración y envío</h2>

      <div className="mt-4 space-y-4">
        <ol className="list-decimal space-y-2 pl-5 text-sm text-slate-700">
          {DECLARACION.map((t) => (
            <li key={t}>{t}</li>
          ))}
        </ol>

        <div className="rounded border border-dashed border-slate-400 p-4 text-sm">
          <p className="h-12 text-slate-400">Firma del postulante</p>
          <p className="font-medium">
            {p.nombres} {p.apellidos}
          </p>
          <p className="text-slate-600">C.I. {p.cedula}</p>
        </div>

        {enviada ? (
          <>
            <Alerta tipo="exito">
              Su postulación fue enviada. Código: <strong>{p.codigo}</strong>. Se envió la confirmación al correo{' '}
              {p.correo_1}.
            </Alerta>
            <button
              type="button"
              onClick={() => window.print()}
              className="rounded border border-slate-300 px-4 py-2 text-sm hover:bg-slate-50"
            >
              Imprimir formulario
            </button>
          </>
        ) : (
          <>
            {obligatorios.length > 0 && (
              <Alerta tipo="error">
                <p className="font-medium">Para enviar la postulación falta:</p>
                <ul className="mt-1 list-disc pl-5">
                  {obligatorios.map((t) => (
                    <li key={t}>{t}</li>
                  ))}
                </ul>
              </Alerta>
            )}
            {recomendaciones.length > 0 && (
              <Alerta tipo="info">
                <ul className="list-disc pl-5">
                  {recomendaciones.map((t) => (
                    <li key={t}>{t}</li>
                  ))}
                </ul>
              </Alerta>
            )}

            <label className="flex items-start gap-2 text-sm">
              <input type="checkbox" checked={acepto} onChange={(e) => setAcepto(e.target.checked)} className="mt-1" />
              <span>Declaro que he leído y acepto lo anterior.</span>
            </label>

            {aviso && <Alerta tipo={aviso.tipo}>{aviso.texto}</Alerta>}

            <div className="flex flex-wrap gap-3">
              <div className="w-56">
                <Boton type="button" variante="borde" cargando={guardando} onClick={() => void alGuardar()}>
                  Guardar Postulación
                </Boton>
              </div>
              <div className="w-56">
                <Boton type="button" variante="oro" cargando={guardando} onClick={() => void alEnviar()}>
                  Enviar Postulación
                </Boton>
              </div>
            </div>
          </>
        )}
      </div>
    </section>
  )
}