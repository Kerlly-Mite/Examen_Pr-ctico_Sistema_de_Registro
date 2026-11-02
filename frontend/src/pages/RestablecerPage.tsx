import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { restablecerPassword } from '../api/auth'
import { Alerta } from '../components/Alerta'
import { Boton } from '../components/Boton'
import { Campo } from '../components/Campo'
import { MensajeApi } from '../components/MensajeApi'
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
    <section className="mx-auto w-full max-w-md rounded-lg bg-white p-6 shadow">
      <h1 className="text-xl font-semibold">Restablecer contraseña</h1>

      {!token ? (
        <div className="mt-4 space-y-3">
          <Alerta tipo="error">El enlace no es válido: falta el token.</Alerta>
          <Link to="/recuperar" className="text-sm text-biess-700 underline">
            Solicitar un nuevo enlace
          </Link>
        </div>
      ) : (
        <>
          <div className="mt-4">
            <MensajeApi error={error} intentos={intentos} restantes={restantes} />
            {codigo === 'enlace_invalido' && (
              <Link to="/recuperar" className="mt-2 inline-block text-sm text-biess-700 underline">
                Solicitar un nuevo enlace
              </Link>
            )}
          </div>

          <form onSubmit={enviar} noValidate className="mt-4 space-y-4">
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
    </section>
  )
}