import type { Aviso } from '../lib/lista'
import { Alerta } from './Alerta'
import { Boton } from './Boton'

interface Props {
  aviso: Aviso | null
  guardando: boolean
  onGuardar: () => void
  texto?: string
}

export function BarraGuardar({ aviso, guardando, onGuardar, texto = 'Guardar progreso' }: Props) {
  return (
    <div className="space-y-3">
      {aviso && <Alerta tipo={aviso.tipo}>{aviso.texto}</Alerta>}
      <Boton type="button" variante="borde" autoAncho cargando={guardando} onClick={onGuardar}>
        {texto}
      </Boton>
    </div>
  )
}