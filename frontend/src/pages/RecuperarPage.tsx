import { useState } from 'react'
import { Link } from 'react-router-dom'
import { solicitarRecuperacion } from '../api/auth'
import type { PruebaIdentidad } from '../api/auth'
import { Alerta } from '../components/Alerta'
import { FormularioIdentidad } from '../components/FormularioIdentidad'
import { MensajeApi } from '../components/MensajeApi'
import { useErrorApi } from '../hooks/useErrorApi'

export function RecuperarPage() {
  const { error, intentos, restantes, bloqueado, manejar, limpiar } = useErrorApi()
  const [enviando, setEnviando] = useState(false)
  const [exito, setExito] = useState<string | null>(null)

  async function solicitar(cedula: string, prueba: PruebaIdentidad) {
    limpiar()
    setEnviando(true)
    try {
      const r = await solicitarRecuperacion(cedula, prueba)
      setExito(r.message)
    } catch (err) {
      manejar(err)
    } finally {
      setEnviando(false)
    }
  }

  return (
    <section className="mx-auto w-full max-w-md rounded-lg bg-white p-6 shadow">
      <h1 className="text-xl font-semibold">Recuperar contraseña</h1>
      <p className="mt-1 text-sm text-slate-600">
        Compruebe su identidad y le enviaremos un enlace a su correo.
      </p>

      <div className="mt-4 space-y-3">
        <MensajeApi error={error} intentos={intentos} restantes={restantes} />
        {exito && <Alerta tipo="exito">{exito}</Alerta>}
      </div>

      {!exito && (
        <div className="mt-4">
          <FormularioIdentidad
            textoBoton="Enviar enlace"
            enviando={enviando}
            deshabilitado={bloqueado}
            onEnviar={solicitar}
          />
        </div>
      )}

      <p className="mt-4 text-sm">
        <Link to="/login" className="text-blue-700 underline">
          Volver a iniciar sesión
        </Link>
      </p>
    </section>
  )
}