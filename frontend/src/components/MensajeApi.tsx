import { Alerta } from './Alerta'
import { formatearTiempo } from '../lib/formato'

interface Props {
  error: string | null
  intentos?: number
  restantes: number
}

export function MensajeApi({ error, intentos, restantes }: Props) {
  if (!error) return null
  return (
    <Alerta tipo="error">
      <p>{error}</p>
      {intentos !== undefined && restantes === 0 && (
        <p className="mt-1">
          {intentos === 1 ? 'Le queda 1 intento.' : `Le quedan ${intentos} intentos.`}
        </p>
      )}
      {restantes > 0 && (
        <p className="mt-1">
          Podrás reintentar en <strong>{formatearTiempo(restantes)}</strong>.
        </p>
      )}
    </Alerta>
  )
}