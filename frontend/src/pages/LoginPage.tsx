import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, Navigate, useLocation } from 'react-router-dom'
import { iniciarSesionApi } from '../api/auth'
import { Alerta } from '../components/Alerta'
import { Boton } from '../components/Boton'
import { Campo } from '../components/Campo'
import { MensajeApi } from '../components/MensajeApi'
import { useAuth } from '../context/useAuth'
import { useErrorApi } from '../hooks/useErrorApi'
import { validarCedula } from '../lib/validaciones'

interface EstadoNavegacion {
  desde?: string
  mensaje?: string
}

export function LoginPage() {
  const { estaAutenticado, iniciarSesion } = useAuth()
  const location = useLocation()
  const estado = location.state as EstadoNavegacion | null
  const destino = estado?.desde ?? '/formulario'

  const { error, intentos, restantes, bloqueado, manejar, limpiar } = useErrorApi()
  const [cedula, setCedula] = useState('')
  const [password, setPassword] = useState('')
  const [errores, setErrores] = useState<{ cedula?: string; password?: string }>({})
  const [enviando, setEnviando] = useState(false)

  if (estaAutenticado) return <Navigate to={destino} replace />

  async function enviar(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const errs = {
      cedula: validarCedula(cedula),
      password: password === '' ? 'Ingrese su contraseña' : undefined,
    }
    setErrores(errs)
    if (errs.cedula || errs.password) return

    limpiar()
    setEnviando(true)
    try {
      const r = await iniciarSesionApi(cedula, password)
      iniciarSesion(r.token, r.postulante)
    } catch (err) {
      manejar(err)
    } finally {
      setEnviando(false)
    }
  }

  return (
    <section className="mx-auto w-full max-w-md rounded-lg bg-white p-6 shadow">
      <h1 className="text-xl font-semibold">Iniciar sesión</h1>
      <p className="mt-1 text-sm text-slate-600">Ingrese con su cédula y contraseña.</p>

      <div className="mt-4 space-y-3">
        {estado?.mensaje && <Alerta tipo="exito">{estado.mensaje}</Alerta>}
        <MensajeApi error={error} intentos={intentos} restantes={restantes} />
      </div>

      <form onSubmit={enviar} noValidate className="mt-4 space-y-4">
        <Campo
          etiqueta="Cédula"
          name="cedula"
          inputMode="numeric"
          maxLength={10}
          value={cedula}
          onChange={(e) => setCedula(e.target.value.replace(/\D/g, ''))}
          error={errores.cedula}
          disabled={bloqueado}
          autoComplete="username"
        />
        <Campo
          etiqueta="Contraseña"
          name="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={errores.password}
          disabled={bloqueado}
          autoComplete="current-password"
        />
        <Boton type="submit" cargando={enviando} disabled={bloqueado}>
          Ingresar
        </Boton>
      </form>

      <div className="mt-4 flex justify-between text-sm">
        <Link to="/recuperar" className="text-blue-700 underline">
          ¿Olvidaste tu contraseña?
        </Link>
        <Link to="/registro" className="text-blue-700 underline">
          Registrarse
        </Link>
      </div>
    </section>
  )
}