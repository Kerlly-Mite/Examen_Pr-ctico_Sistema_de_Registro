import { useState } from 'react'
import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import { tieneErrores, validarDomicilio } from '../lib/domicilio'
import type { ErroresDomicilio, Ubicacion } from '../lib/domicilio'
import { domicilioDe } from '../lib/postulacion'
import type { PostulacionReq } from '../types/postulacion'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'
import { SelectorUbicacion } from './SelectorUbicacion'

export function SeccionDomicilio() {
  const { postulacion: p, actualizar } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()
  const [errores, setErrores] = useState<ErroresDomicilio>({})
  const dom = domicilioDe(p)

  function cambiar(parcial: Partial<PostulacionReq>) {
    actualizar(parcial)
    limpiar()
  }

  function cambiarUbicacion(u: Ubicacion) {
    cambiar({
      pais_id: u.paisId,
      provincia_id: u.provinciaId,
      provincia_texto: u.provinciaTexto,
      canton_id: u.cantonId,
      ciudad_texto: u.cantonTexto,
    })
  }

  async function guardar() {
    const errs = validarDomicilio(dom)
    setErrores(errs)
    await ejecutar(
      tieneErrores(errs)
        ? { exito: 'Progreso guardado. Faltan campos obligatorios por completar.', tipoExito: 'info' }
        : { exito: 'Sección guardada y completa.' },
    )
  }

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold text-biess-800">Información domiciliaria y de contacto</h2>

      <div className="mt-4 space-y-4">
        <SelectorUbicacion valor={dom} onCambio={cambiarUbicacion} errores={errores} />

        <div className="grid gap-4 md:grid-cols-2">
          <Campo
            etiqueta="Teléfono celular *"
            name="telefonoCelular"
            inputMode="tel"
            value={p.telefono_celular}
            onChange={(e) => cambiar({ telefono_celular: e.target.value.replace(/[^\d+]/g, '') })}
            error={errores.telefonoCelular}
          />
          <Campo
            etiqueta="Correo electrónico 1"
            name="correo1"
            value={p.correo_1}
            readOnly
            ayuda="Es el correo con el que se registró."
          />
          <Campo
            etiqueta="Correo electrónico 2"
            name="correo2"
            type="email"
            value={p.correo_2}
            onChange={(e) => cambiar({ correo_2: e.target.value })}
            error={errores.correo2}
          />
          <Campo
            etiqueta="Teléfono de domicilio"
            name="telefonoDomicilio"
            inputMode="tel"
            value={p.telefono_domicilio}
            onChange={(e) => cambiar({ telefono_domicilio: e.target.value.replace(/[^\d+]/g, '') })}
            error={errores.telefonoDomicilio}
          />
          <Campo
            etiqueta="Calle principal *"
            name="callePrincipal"
            value={p.calle_principal}
            onChange={(e) => cambiar({ calle_principal: e.target.value })}
            error={errores.callePrincipal}
          />
          <Campo
            etiqueta="Número"
            name="numero"
            value={p.numero}
            onChange={(e) => cambiar({ numero: e.target.value })}
          />
          <Campo
            etiqueta="Calle secundaria"
            name="calleSecundaria"
            value={p.calle_secundaria}
            onChange={(e) => cambiar({ calle_secundaria: e.target.value })}
          />
          <Campo
            etiqueta="Sector"
            name="sector"
            value={p.sector}
            onChange={(e) => cambiar({ sector: e.target.value })}
          />
          <div className="md:col-span-2">
            <Campo
              etiqueta="Referencia"
              name="referencia"
              value={p.referencia}
              onChange={(e) => cambiar({ referencia: e.target.value })}
            />
          </div>
        </div>

        <BarraGuardar aviso={aviso} guardando={guardando} onGuardar={() => void guardar()} />
      </div>
    </section>
  )
}