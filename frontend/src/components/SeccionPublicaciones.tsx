import { useState } from 'react'
import { useFormularioListo } from '../context/useFormulario'
import { useGuardado } from '../hooks/useGuardado'
import { useListaEditable } from '../hooks/useListaEditable'
import { fechaHoyISO, formatearFecha } from '../lib/formato'
import { TIPOS_PUBLICACION, etiquetaDe } from '../lib/opciones'
import { PUBLICACION_VACIA, validarPublicacion } from '../lib/secciones'
import type { Publicacion } from '../types/postulacion'
import { BarraGuardar } from './BarraGuardar'
import { Campo } from './Campo'
import { nombreDocumento } from '../api/documentos'
import { CampoArchivo } from './CampoArchivo'
import { GrupoSiNo } from './GrupoSiNo'
import { PanelLista } from './PanelLista'
import { SelectorTexto } from './SelectorTexto'

export function SeccionPublicaciones() {
  const { postulacion: p, actualizar } = useFormularioListo()
  const { aviso, guardando, ejecutar, limpiar } = useGuardado()
  const [tiene, setTiene] = useState<boolean | null>(p.publicaciones.length > 0 ? true : null)

  const lista = useListaEditable<Publicacion>(PUBLICACION_VACIA, validarPublicacion, p.publicaciones, (items) => {
    actualizar({ publicaciones: items })
    limpiar()
  })
  const { form, errores, cambiar } = lista

  function cambiarRespuesta(valor: boolean) {
    if (
      !valor &&
      p.publicaciones.length > 0 &&
      !window.confirm('Se eliminarán las publicaciones agregadas. ¿Desea continuar?')
    ) {
      return
    }
    setTiene(valor)
    if (!valor) {
      actualizar({ publicaciones: [] })
      lista.cancelar()
    }
    limpiar()
  }

  function eliminar(indice: number) {
    const x = p.publicaciones[indice]
    if (x === undefined || !window.confirm(`¿Eliminar "${x.titulo}" de la lista?`)) return
    lista.quitar(indice)
  }

  const pendiente = tiene && lista.hayContenido() ? 'Tiene una publicación sin agregar. Pulse Agregar o Cancelar.' : null

  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="text-lg font-semibold text-biess-800">Publicaciones</h2>

      <div className="mt-4 space-y-4">
        <GrupoSiNo
          name="tienePublicaciones"
          etiqueta="¿Cuenta con publicaciones e investigaciones académicas? *"
          valor={tiene}
          onCambio={cambiarRespuesta}
        />

        {tiene === true && (
          <PanelLista
            nombreItem="publicación"
            editando={lista.editando !== null}
            onAgregar={lista.agregar}
            onCancelar={lista.cancelar}
            items={p.publicaciones}
            textoVacio="Aún no ha agregado publicaciones."
            onEditar={lista.editar}
            onEliminar={eliminar}
            resumen={(x) => (
              <>
                <p className="font-medium">{x.titulo}</p>
                <p className="text-slate-600">
                  {etiquetaDe(TIPOS_PUBLICACION, x.tipo)} · {x.medio_publicacion} ·{' '}
                  {formatearFecha(x.fecha_publicacion)} · Relacionada con su campo:{' '}
                  {x.relacionada_formacion ? 'Sí' : 'No'} · PDF: {nombreDocumento(x.documento_respaldo)}
                </p>
              </>
            )}
            campos={
              <>
                <Campo
                  etiqueta="Título de la publicación o trabajo de investigación *"
                  name="publicacionTitulo"
                  value={form.titulo}
                  onChange={(e) => cambiar('titulo', e.target.value)}
                  error={errores.titulo}
                />
                <SelectorTexto
                  etiqueta="Tipo *"
                  name="publicacionTipo"
                  opciones={TIPOS_PUBLICACION}
                  valor={form.tipo}
                  onCambio={(v) => cambiar('tipo', v)}
                  error={errores.tipo}
                />
                <Campo
                  etiqueta="Medio de publicación *"
                  name="publicacionMedio"
                  value={form.medio_publicacion}
                  onChange={(e) => cambiar('medio_publicacion', e.target.value)}
                  error={errores.medio_publicacion}
                />
                <Campo
                  etiqueta="Fecha de publicación *"
                  name="publicacionFecha"
                  type="date"
                  max={fechaHoyISO()}
                  value={form.fecha_publicacion ?? ''}
                  onChange={(e) => cambiar('fecha_publicacion', e.target.value === '' ? null : e.target.value)}
                  error={errores.fecha_publicacion}
                />
                <GrupoSiNo
                  name="publicacionRelacion"
                  etiqueta="¿Guarda relación con el campo de su ejercicio o formación profesional?"
                  valor={form.relacionada_formacion}
                  onCambio={(v) => cambiar('relacionada_formacion', v)}
                />
                <CampoArchivo
                  etiqueta="Documento de respaldo *"
                  maxMB={2}
                  nombre={form.documento_respaldo}
                  onCambio={(n) => cambiar('documento_respaldo', n)}
                  error={errores.documento_respaldo}
                />
              </>
            }
          />
        )}

        <BarraGuardar aviso={aviso} guardando={guardando} onGuardar={() => void ejecutar({ pendiente })} />
      </div>
    </section>
  )
}