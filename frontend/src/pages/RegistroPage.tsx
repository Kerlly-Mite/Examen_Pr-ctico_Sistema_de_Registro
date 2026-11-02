import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { registrar, verificarIdentidad } from '../api/auth'
import type { PruebaIdentidad } from '../api/auth'
import { ApiError } from '../api/client'
import { Alerta } from '../components/Alerta'
import { Boton } from '../components/Boton'
import { Campo } from '../components/Campo'
import { FormularioIdentidad } from '../components/FormularioIdentidad'
import { MensajeApi } from '../components/MensajeApi'
import { PasosRegistro } from '../components/PasosRegistro'
import { useErrorApi } from '../hooks/useErrorApi'
import { validarConfirmacion, validarCorreo, validarPassword } from '../lib/validaciones'
import type { VerificarIdentidadRespuesta } from '../types/api'

interface ErroresPaso2 {
  correo?: string
  password?: string
  confirmacion?: string
}

export function RegistroPage() {
  const navigate = useNavigate()
  const { error, intentos, restantes, bloqueado, manejar, limpiar } = useErrorApi()
  const [paso, setPaso] = useState<1 | 2>(1)
  const [verificacion, setVerificacion] = useState<VerificarIdentidadRespuesta | null>(null)
  const [enviando, setEnviando] = useState(false)
  const [correo, setCorreo] = useState('')
  const [password, setPassword] = useState('')
  const [confirmacion, setConfirmacion] = useState('')
  const [errores, setErrores] = useState<ErroresPaso2>({})

  async function verificar(cedula: string, prueba: PruebaIdentidad) {
    limpiar()
    setEnviando(true)
    try {
      const r = await verificarIdentidad(cedula, prueba)
      setVerificacion(r)
      setPaso(2)
    } catch (err) {
      manejar(err)
    } finally {
      setEnviando(false)
    }
  }

  async function crearCuenta(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!verificacion) return
    const errs: ErroresPaso2 = {
      correo: validarCorreo(correo),
      password: validarPassword(password),
      confirmacion: validarConfirmacion(password, confirmacion),
    }
    setErrores(errs)
    if (errs.correo || errs.password || errs.confirmacion) return

    limpiar()
    setEnviando(true)
    try {
      const r = await registrar(verificacion.token_registro, correo.trim(), password)
      navigate('/login', { replace: true, state: { mensaje: r.message } })
    } catch (err) {
      manejar(err)
      if (err instanceof ApiError && err.codigo === 'verificacion_expirada') {
        setPaso(1)
        setVerificacion(null)
      }
    } finally {
      setEnviando(false)
    }
  }

  return (
    <section className="mx-auto w-full max-w-md rounded-lg bg-white p-6 shadow">
      <PasosRegistro paso={paso} />
      <h1 className="mt-3 text-xl font-semibold text-biess-800">Crear cuenta</h1>

      <div className="mt-4">
        <MensajeApi error={error} intentos={intentos} restantes={restantes} />
      </div>

      {paso === 1 && (
        <div className="mt-4">
          <FormularioIdentidad
            textoBoton="Continuar"
            enviando={enviando}
            deshabilitado={bloqueado}
            onEnviar={verificar}
          />
        </div>
      )}

      {paso === 2 && verificacion && (
        <form onSubmit={crearCuenta} noValidate className="mt-4 space-y-4">
          <Alerta tipo="exito">
            Identidad verificada: <strong>{verificacion.nombres} {verificacion.apellidos}</strong>.
            Tiene {Math.round(verificacion.expira_en_segundos / 60)} minutos para completar el registro.
          </Alerta>
          <Campo
            etiqueta="Correo electrónico"
            name="correo"
            type="email"
            value={correo}
            onChange={(e) => setCorreo(e.target.value)}
            error={errores.correo}
            autoComplete="email"
          />
          <Campo
            etiqueta="Contraseña"
            name="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errores.password}
            ayuda="De 8 a 72 caracteres, con al menos una letra y un número."
            autoComplete="new-password"
          />
          <Campo
            etiqueta="Confirmar contraseña"
            name="confirmacion"
            type="password"
            value={confirmacion}
            onChange={(e) => setConfirmacion(e.target.value)}
            error={errores.confirmacion}
            autoComplete="new-password"
          />
          <Boton type="submit" cargando={enviando}>
            Crear cuenta
          </Boton>
        </form>
      )}

      <p className="mt-4 text-sm">
        ¿Ya tienes cuenta?{' '}
        <Link to="/login" className="text-biess-700 underline">
          Inicie sesión
        </Link>
      </p>
    </section>
  )
}