import { useState } from 'react'
import type { FormEvent } from 'react'
import { Boton } from './Boton'
import { Campo } from './Campo'
import type { PruebaIdentidad } from '../api/auth'
import { fechaHoyISO } from '../lib/formato'
import { validarCedula, validarCodigoDactilar, validarFecha } from '../lib/validaciones'

type Modo = 'fecha' | 'dactilar'

interface Props {
  textoBoton: string
  enviando: boolean
  deshabilitado: boolean
  onEnviar: (cedula: string, prueba: PruebaIdentidad) => void
}

export function FormularioIdentidad({ textoBoton, enviando, deshabilitado, onEnviar }: Props) {
  const [cedula, setCedula] = useState('')
  const [modo, setModo] = useState<Modo>('fecha')
  const [fecha, setFecha] = useState('')
  const [codigo, setCodigo] = useState('')
  const [errores, setErrores] = useState<{ cedula?: string; prueba?: string }>({})

  function enviar(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const errorCedula = validarCedula(cedula)
    const errorPrueba = modo === 'fecha' ? validarFecha(fecha) : validarCodigoDactilar(codigo)
    setErrores({ cedula: errorCedula, prueba: errorPrueba })
    if (errorCedula || errorPrueba) return
    onEnviar(cedula, modo === 'fecha' ? { fecha_expedicion: fecha } : { codigo_dactilar: codigo })
  }

  return (
    <form onSubmit={enviar} noValidate className="space-y-4">
      <Campo
        etiqueta="Cédula de identidad"
        name="cedula"
        inputMode="numeric"
        maxLength={10}
        value={cedula}
        onChange={(e) => setCedula(e.target.value.replace(/\D/g, ''))}
        error={errores.cedula}
        disabled={deshabilitado}
        autoComplete="username"
      />

      <fieldset className="space-y-2">
        <legend className="text-sm font-medium text-slate-700">Para comprobar su identidad ingrese:</legend>
        <div className="flex gap-4 text-sm">
          <label className="flex items-center gap-1">
            <input
              type="radio"
              name="modo"
              checked={modo === 'fecha'}
              onChange={() => setModo('fecha')}
              disabled={deshabilitado}
            />
            Fecha de expedición
          </label>
          <label className="flex items-center gap-1">
            <input
              type="radio"
              name="modo"
              checked={modo === 'dactilar'}
              onChange={() => setModo('dactilar')}
              disabled={deshabilitado}
            />
            Código dactilar
          </label>
        </div>

        {modo === 'fecha' ? (
          <Campo
            etiqueta="Fecha de expedición de la cédula"
            name="fecha_expedicion"
            type="date"
            max={fechaHoyISO()}
            value={fecha}
            onChange={(e) => setFecha(e.target.value)}
            error={errores.prueba}
            disabled={deshabilitado}
          />
        ) : (
          <Campo
            etiqueta="Código dactilar (10 dígitos)"
            name="codigo_dactilar"
            inputMode="numeric"
            maxLength={10}
            value={codigo}
            onChange={(e) => setCodigo(e.target.value.replace(/\D/g, ''))}
            error={errores.prueba}
            disabled={deshabilitado}
          />
        )}
      </fieldset>

      <Boton type="submit" cargando={enviando} disabled={deshabilitado}>
        {textoBoton}
      </Boton>
    </form>
  )
}