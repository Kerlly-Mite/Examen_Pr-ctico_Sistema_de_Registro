import { useState } from 'react'
import type { FormEvent } from 'react'
import { guardarSeccion, leerSeccion } from '../api/borrador'
import { useAuth } from '../context/useAuth'
import { DOMICILIO_VACIO, tieneErrores, validarDomicilio } from '../lib/domicilio'
import type { Domicilio, ErroresDomicilio } from '../lib/domicilio'
import { Alerta } from './Alerta'
import { Boton } from './Boton'
import { Campo } from './Campo'
import { SelectorUbicacion } from './SelectorUbicacion'

interface Aviso {
  tipo: 'exito' | 'info' | 'error'
  texto: string
}

export function SeccionDomicilio() {
  const { usuario } = useAuth()
  const cedula = usuario?.cedula ?? ''

  const [datos, setDatos] = useState<Domicilio>(() => {
    const guardado = cedula === '' ? null : leerSeccion<Partial<Domicilio>>(cedula, 'domicilio')
    return { ...DOMICILIO_VACIO, ...guardado }
  })
  const [errores, setErrores] = useState<ErroresDomicilio>({})
  const [aviso, setAviso] = useState<Aviso | null>(null)

  function cambiar<K extends keyof Domicilio>(campo: K, valor: Domicilio[K]) {
    setDatos((prev) => ({ ...prev, [campo]: valor }))
    setAviso(null)
  }

  function guardar(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const errs = validarDomicilio(datos)
    setErrores(errs)

    const guardado = cedula !== '' && guardarSeccion(cedula, 'domicilio', datos)
    if (!guardado) {
      setAviso({ tipo: 'error', texto: 'No se pudo guardar el progreso en este navegador.' })
    } else if (tieneErrores(errs)) {
      setAviso({ tipo: 'info', texto: 'Progreso guardado. Faltan campos obligatorios por completar.' })
    } else {
      setAviso({ tipo: 'exito', texto: 'Sección guardada y completa.' })
    }
  }

  return (
    <section className="rounded-lg bg-white p-6 shadow">
      <h2 className="text-lg font-semibold">Información domiciliaria y de contacto</h2>

      <form onSubmit={guardar} noValidate className="mt-4 space-y-4">
        <SelectorUbicacion
          valor={datos}
          onCambio={(u) => {
            setDatos((prev) => ({ ...prev, ...u }))
            setAviso(null)
          }}
          errores={errores}
        />

        <div className="grid gap-4 md:grid-cols-2">
          <Campo
            etiqueta="Teléfono celular *"
            name="telefonoCelular"
            inputMode="tel"
            value={datos.telefonoCelular}
            onChange={(e) => cambiar('telefonoCelular', e.target.value.replace(/[^\d+]/g, ''))}
            error={errores.telefonoCelular}
          />
          <Campo
            etiqueta="Correo electrónico 1"
            name="correo1"
            value={usuario?.correo ?? ''}
            readOnly
            ayuda="Es el correo con el que se registró."
          />
          <Campo
            etiqueta="Correo electrónico 2"
            name="correo2"
            type="email"
            value={datos.correo2}
            onChange={(e) => cambiar('correo2', e.target.value)}
            error={errores.correo2}
          />
          <Campo
            etiqueta="Teléfono de domicilio"
            name="telefonoDomicilio"
            inputMode="tel"
            value={datos.telefonoDomicilio}
            onChange={(e) => cambiar('telefonoDomicilio', e.target.value.replace(/[^\d+]/g, ''))}
            error={errores.telefonoDomicilio}
          />
          <Campo
            etiqueta="Calle principal *"
            name="callePrincipal"
            value={datos.callePrincipal}
            onChange={(e) => cambiar('callePrincipal', e.target.value)}
            error={errores.callePrincipal}
          />
          <Campo
            etiqueta="Número"
            name="numero"
            value={datos.numero}
            onChange={(e) => cambiar('numero', e.target.value)}
          />
          <Campo
            etiqueta="Calle secundaria"
            name="calleSecundaria"
            value={datos.calleSecundaria}
            onChange={(e) => cambiar('calleSecundaria', e.target.value)}
          />
          <Campo
            etiqueta="Sector"
            name="sector"
            value={datos.sector}
            onChange={(e) => cambiar('sector', e.target.value)}
          />
          <div className="md:col-span-2">
            <Campo
              etiqueta="Referencia"
              name="referencia"
              value={datos.referencia}
              onChange={(e) => cambiar('referencia', e.target.value)}
            />
          </div>
        </div>

        {aviso && <Alerta tipo={aviso.tipo}>{aviso.texto}</Alerta>}

        <div className="max-w-xs">
          <Boton type="submit">Guardar progreso</Boton>
        </div>
      </form>
    </section>
  )
}