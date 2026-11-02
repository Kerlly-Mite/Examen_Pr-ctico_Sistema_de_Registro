import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { restablecerPassword } from '../api/auth'
import { Alerta } from '../components/Alerta'
import { Boton } from '../components/Boton'
import { Campo } from '../components/Campo'
import { MensajeApi } from '../components/MensajeApi'
import { TarjetaAuth } from '../components/TarjetaAuth'
import { useErrorApi } from '../hooks/useErrorApi'
import { validarConfirmacion, validarPassword } from '../lib/validaciones'

export function RestablecerPage() {
  const [params] = useSearchParams()
  const token = params.get('token')
  const navigate = useNavigate()
  const { error, codigo, intentos, restantes, manejar, limpiar } = useErrorApi()
  const [password, setPassword] = useState('')
  const [confirmacion, setConfirmacion] = useState('')
  const [errores, setErrores] = useState<{ password?: string; confirmacion?: string }>({})
  const [enviando, setEnviando] = useState(false)

  async function enviar(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!token) return
    const errs = {
      password: validarPassword(password),
      confirmacion: validarConfirmacion(password, confirmacion),
    }
    setErrores(errs)
    if (errs.password || errs.confirmacion) return

    limpiar()
    setEnviando(true)
    try {
      const r = await restablecerPassword(token, password)
      navigate('/login', { replace: true, state: { mensaje: r.message } })
    } catch (err) {
      manejar(err)
    } finally {
      setEnviando(false)
    }
  }

  return (
    <TarjetaAuth
      titulo="Restablecer contraseña"
      subtitulo={token ? 'Escribe tu nueva contraseña.' : undefined}
      pie={
        <Link to="/login" className="font-medium text-biess-700 underline">
          Volver a iniciar sesión
        </Link>
      }
    >
      {!token ? (
        <>
          <Alerta tipo="error">El enlace no es válido: falta el token.</Alerta>
          <Link to="/recuperar" className="text-sm font-medium text-biess-700 underline">
            Solicitar un nuevo enlace
          </Link>
        </>
      ) : (
        <>
          <MensajeApi error={error} intentos={intentos} restantes={restantes} />
          {codigo === 'enlace_invalido' && (
            <Link to="/recuperar" className="block text-sm font-medium text-biess-700 underline">
              Solicitar un nuevo enlace
            </Link>
          )}

          <form onSubmit={enviar} noValidate className="space-y-4">
            <Campo
              etiqueta="Nueva contraseña"
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
              Guardar nueva contraseña
            </Boton>
          </form>
        </>
      )}
    </TarjetaAuth>
  )
}