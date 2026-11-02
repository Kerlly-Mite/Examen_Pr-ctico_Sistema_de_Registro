import { useState } from 'react'
import { guardarSeccion, leerSeccion } from '../api/borrador'
import { useAuth } from '../context/useAuth'
import {
  ADICIONAL_VACIO,
  FAMILIAR_VACIO,
  INSTITUCIONES,
  PARENTESCOS,
  formularioFamiliarConContenido,
  hayErroresFamiliar,
  validarFamiliar,
} from '../lib/informacionAdicional'
import type { ErroresFamiliar, Familiar, FamiliarForm, InformacionAdicional } from '../lib/informacionAdicional'
import { Alerta } from './Alerta'
import { Boton } from './Boton'
import { Campo } from './Campo'
import { GrupoSiNo } from './GrupoSiNo'
import { SelectorTexto } from './SelectorTexto'

interface Aviso {
  tipo: 'exito' | 'info' | 'error'
  texto: string
}

const botonPrincipal =
  'rounded bg-blue-800 px-4 py-2 text-sm font-medium text-white hover:bg-blue-900'
const botonSecundario = 'rounded border border-slate-300 px-3 py-1 text-sm hover:bg-slate-50'

export function SeccionInformacionAdicional() {
  const { usuario } = useAuth()
  const cedula = usuario?.cedula ?? ''

  const [datos, setDatos] = useState<InformacionAdicional>(() => {
    const guardado =
      cedula === '' ? null : leerSeccion<Partial<InformacionAdicional>>(cedula, 'adicional')
    return { ...ADICIONAL_VACIO, ...guardado }
  })
  const [form, setForm] = useState<FamiliarForm>(FAMILIAR_VACIO)
  const [editandoId, setEditandoId] = useState<string | null>(null)
  const [erroresForm, setErroresForm] = useState<ErroresFamiliar>({})
  const [errorSeccion, setErrorSeccion] = useState<string | undefined>(undefined)
  const [aviso, setAviso] = useState<Aviso | null>(null)

  function cambiarCampo<K extends keyof FamiliarForm>(campo: K, valor: FamiliarForm[K]) {
    setForm((prev) => ({ ...prev, [campo]: valor }))
    setAviso(null)
  }

  function cambiarRespuesta(valor: boolean) {
    if (
      !valor &&
      datos.familiares.length > 0 &&
      !window.confirm('Se eliminarán los familiares agregados. ¿Desea continuar?')
    ) {
      return
    }
    setDatos((prev) => ({ tieneFamiliares: valor, familiares: valor ? prev.familiares : [] }))
    if (!valor) {
      setForm(FAMILIAR_VACIO)
      setEditandoId(null)
      setErroresForm({})
    }
    setErrorSeccion(undefined)
    setAviso(null)
  }

  function agregar() {
    const errs = validarFamiliar(form)
    setErroresForm(errs)
    if (hayErroresFamiliar(errs)) return

    if (editandoId === null) {
      const nuevo: Familiar = { id: crypto.randomUUID(), ...form }
      setDatos((prev) => ({ ...prev, familiares: [...prev.familiares, nuevo] }))
    } else {
      setDatos((prev) => ({
        ...prev,
        familiares: prev.familiares.map((f) => (f.id === editandoId ? { id: f.id, ...form } : f)),
      }))
    }
    setForm(FAMILIAR_VACIO)
    setEditandoId(null)
    setErrorSeccion(undefined)
    setAviso(null)
  }

  function editar(f: Familiar) {
    setForm({
      nombres: f.nombres,
      parentesco: f.parentesco,
      institucion: f.institucion,
      area: f.area,
      cargo: f.cargo,
    })
    setEditandoId(f.id)
    setErroresForm({})
    setAviso(null)
  }

  function cancelarEdicion() {
    setForm(FAMILIAR_VACIO)
    setEditandoId(null)
    setErroresForm({})
  }

  function eliminar(f: Familiar) {
    if (!window.confirm(`¿Eliminar a ${f.nombres} de la lista?`)) return
    setDatos((prev) => ({ ...prev, familiares: prev.familiares.filter((x) => x.id !== f.id) }))
    if (editandoId === f.id) cancelarEdicion()
    setAviso(null)
  }

  function guardar() {
    let falta: string | undefined
    if (datos.tieneFamiliares === null) {
      falta = 'Responda si tiene familiares trabajando en el IESS o BIESS.'
    } else if (datos.tieneFamiliares && datos.familiares.length === 0) {
      falta = 'Agregue al menos un familiar o seleccione No.'
    } else if (
      datos.tieneFamiliares &&
      (editandoId !== null || formularioFamiliarConContenido(form))
    ) {
      falta = 'Tiene un familiar sin agregar. Pulse Agregar o Cancelar.'
    }
    setErrorSeccion(falta)

    const guardado = cedula !== '' && guardarSeccion(cedula, 'adicional', datos)
    if (!guardado) {
      setAviso({ tipo: 'error', texto: 'No se pudo guardar el progreso en este navegador.' })
    } else if (falta) {
      setAviso({ tipo: 'info', texto: 'Progreso guardado. Falta completar la sección.' })
    } else {
      setAviso({ tipo: 'exito', texto: 'Sección guardada y completa.' })
    }
  }

  return (
    <section className="rounded-lg bg-white p-6 shadow">
      <h2 className="text-lg font-semibold">Información adicional</h2>

      <div className="mt-4 space-y-4">
        <GrupoSiNo
          name="tieneFamiliares"
          etiqueta="¿Tiene familiares dentro del segundo grado de consanguinidad y cuarto de afinidad trabajando actualmente en el IESS o BIESS? *"
          valor={datos.tieneFamiliares}
          onCambio={cambiarRespuesta}
        />

        {datos.tieneFamiliares === true && (
          <div className="space-y-4 rounded border border-slate-200 p-4">
            <h3 className="text-sm font-semibold">
              {editandoId === null ? 'Agregar familiar' : 'Editar familiar'}
            </h3>

            <div className="grid gap-4 md:grid-cols-2">
              <Campo
                etiqueta="Nombres y apellidos del familiar *"
                name="familiarNombres"
                value={form.nombres}
                onChange={(e) => cambiarCampo('nombres', e.target.value)}
                error={erroresForm.nombres}
              />
              <SelectorTexto
                etiqueta="Parentesco *"
                name="familiarParentesco"
                opciones={PARENTESCOS}
                valor={form.parentesco}
                onCambio={(v) => cambiarCampo('parentesco', v)}
                error={erroresForm.parentesco}
              />
              <SelectorTexto
                etiqueta="Institución donde trabaja *"
                name="familiarInstitucion"
                opciones={INSTITUCIONES}
                valor={form.institucion}
                onCambio={(v) => cambiarCampo('institucion', v)}
                error={erroresForm.institucion}
              />
              <Campo
                etiqueta="Área donde trabaja *"
                name="familiarArea"
                value={form.area}
                onChange={(e) => cambiarCampo('area', e.target.value)}
                error={erroresForm.area}
              />
              <Campo
                etiqueta="Cargo *"
                name="familiarCargo"
                value={form.cargo}
                onChange={(e) => cambiarCampo('cargo', e.target.value)}
                error={erroresForm.cargo}
              />
            </div>

            <div className="flex gap-2">
              <button type="button" onClick={agregar} className={botonPrincipal}>
                {editandoId === null ? 'Agregar' : 'Actualizar'}
              </button>
              {editandoId !== null && (
                <button type="button" onClick={cancelarEdicion} className={botonSecundario}>
                  Cancelar
                </button>
              )}
            </div>

            {datos.familiares.length > 0 ? (
              <ul className="divide-y divide-slate-200 rounded border border-slate-200">
                {datos.familiares.map((f) => (
                  <li key={f.id} className="flex flex-wrap items-center justify-between gap-2 p-3">
                    <div className="text-sm">
                      <p className="font-medium">{f.nombres}</p>
                      <p className="text-slate-600">
                        {f.parentesco} · {f.institucion} · {f.area} · {f.cargo}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <button type="button" onClick={() => editar(f)} className={botonSecundario}>
                        Editar
                      </button>
                      <button
                        type="button"
                        onClick={() => eliminar(f)}
                        className={`${botonSecundario} text-red-700`}
                      >
                        Eliminar
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-slate-500">Aún no ha agregado familiares.</p>
            )}
          </div>
        )}

        {errorSeccion && <p className="text-sm text-red-600">{errorSeccion}</p>}
        {aviso && <Alerta tipo={aviso.tipo}>{aviso.texto}</Alerta>}

        <div className="max-w-xs">
          <Boton type="button" onClick={guardar}>
            Guardar información
          </Boton>
        </div>
      </div>
    </section>
  )
}