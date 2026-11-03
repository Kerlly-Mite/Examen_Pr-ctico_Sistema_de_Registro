import { useState } from 'react'
import { Link } from 'react-router-dom'
import { solicitarRecuperacion } from '../api/auth'
import type { PruebaIdentidad } from '../api/auth'
import { Alerta } from '../components/Alerta'
import { FormularioIdentidad } from '../components/FormularioIdentidad'
import { MensajeApi } from '../components/MensajeApi'
import { TarjetaAuth } from '../components/TarjetaAuth'
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
    <TarjetaAuth
      titulo="Recuperar contraseña"
      subtitulo="Comprueba tu identidad y te enviaremos un enlace a tu correo."
      pie={
        <Link to="/login" className="font-medium text-biess-700 underline">
          Volver a iniciar sesión
        </Link>
      }
    >
      <MensajeApi error={error} intentos={intentos} restantes={restantes} />
      {exito && <Alerta tipo="exito">{exito}</Alerta>}

      {!exito && (
        <FormularioIdentidad
          textoBoton="Enviar enlace"
          enviando={enviando}
          deshabilitado={bloqueado}
          onEnviar={solicitar}
        />
      )}
    </TarjetaAuth>
  )
}