import { obtenerCantones, obtenerPaises, obtenerProvincias } from '../api/catalogos'
import { useCatalogo } from '../hooks/useCatalogo'
import type { Ubicacion } from '../lib/domicilio'
import { Campo } from './Campo'
import { Selector } from './Selector'

interface Props {
  valor: Ubicacion
  onCambio: (nuevo: Ubicacion) => void
  errores: { pais?: string; provincia?: string; canton?: string }
}

function ErrorCarga({ mensaje, onReintentar }: { mensaje: string; onReintentar: () => void }) {
  return (
    <p className="mt-1 text-xs text-red-600">
      {mensaje}{' '}
      <button type="button" onClick={onReintentar} className="underline">
        Reintentar
      </button>
    </p>
  )
}

export function SelectorUbicacion({ valor, onCambio, errores }: Props) {
  const paisId = valor.paisId
  const provinciaId = valor.provinciaId

  const paises = useCatalogo('paises', obtenerPaises)
  const provincias = useCatalogo(
    paisId === null ? null : `provincias:${paisId}`,
    () => obtenerProvincias(paisId ?? 0),
  )
  const cantones = useCatalogo(
    provinciaId === null ? null : `cantones:${provinciaId}`,
    () => obtenerCantones(provinciaId ?? 0),
  )

  // Si la API no tiene provincias para el país, se escribe a mano
  const provinciaManual =
    paisId !== null && !provincias.cargando && provincias.error === null && provincias.datos.length === 0
  const cantonManual =
    provinciaManual ||
    (provinciaId !== null && !cantones.cargando && cantones.error === null && cantones.datos.length === 0)

  return (
    <div className="grid gap-4 md:grid-cols-3">
      <div>
        <Selector
          etiqueta="País *"
          name="pais"
          opciones={paises.datos}
          valor={paisId}
          onCambio={(id) =>
            onCambio({ paisId: id, provinciaId: null, provinciaTexto: '', cantonId: null, cantonTexto: '' })
          }
          placeholder={paises.cargando ? 'Cargando países...' : 'Seleccione un país'}
          disabled={paises.cargando || paises.error !== null}
          error={errores.pais}
        />
        {paises.error && <ErrorCarga mensaje={paises.error} onReintentar={paises.reintentar} />}
      </div>

      {paisId === null ? (
        <Selector
          etiqueta="Provincia / Estado *"
          name="provincia"
          opciones={[]}
          valor={null}
          onCambio={() => undefined}
          placeholder="Elija un país"
          disabled
          error={errores.provincia}
        />
      ) : provinciaManual ? (
        <Campo
          etiqueta="Provincia / Estado *"
          name="provincia"
          value={valor.provinciaTexto}
          onChange={(e) => onCambio({ ...valor, provinciaTexto: e.target.value })}
          error={errores.provincia}
        />
      ) : (
        <div>
          <Selector
            etiqueta="Provincia / Estado *"
            name="provincia"
            opciones={provincias.datos}
            valor={provinciaId}
            onCambio={(id) => onCambio({ ...valor, provinciaId: id, cantonId: null, cantonTexto: '' })}
            placeholder={provincias.cargando ? 'Cargando provincias...' : 'Seleccione una provincia'}
            disabled={provincias.cargando || provincias.error !== null}
            error={errores.provincia}
          />
          {provincias.error && (
            <ErrorCarga mensaje={provincias.error} onReintentar={provincias.reintentar} />
          )}
        </div>
      )}

      {cantonManual ? (
        <Campo
          etiqueta="Ciudad / Cantón *"
          name="canton"
          value={valor.cantonTexto}
          onChange={(e) => onCambio({ ...valor, cantonTexto: e.target.value })}
          error={errores.canton}
        />
      ) : (
        <div>
          <Selector
            etiqueta="Ciudad / Cantón *"
            name="canton"
            opciones={cantones.datos}
            valor={valor.cantonId}
            onCambio={(id) => onCambio({ ...valor, cantonId: id })}
            placeholder={
              provinciaId === null
                ? 'Elija una provincia'
                : cantones.cargando
                  ? 'Cargando cantones...'
                  : 'Seleccione un cantón'
            }
            disabled={provinciaId === null || cantones.cargando || cantones.error !== null}
            error={errores.canton}
          />
          {cantones.error && <ErrorCarga mensaje={cantones.error} onReintentar={cantones.reintentar} />}
        </div>
      )}
    </div>
  )
}