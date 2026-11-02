import { tieneErrores, validarDomicilio } from './domicilio'
import { domicilioDe, requierePareja } from './postulacion'
import type { OpcionesPostulacion, Postulacion } from '../types/postulacion'

export interface InfoSeccion {
  id: string
  titulo: string
  completa: boolean
}

// El orden de esta lista es el orden del menú lateral y de las secciones (sigue el PDF)
export function estadoSecciones(p: Postulacion, o: OpcionesPostulacion): InfoSeccion[] {
  const parejaOk =
    !requierePareja(p.estado_civil) || (p.pareja_nombres.trim() !== '' && /^\d{10}$/.test(p.pareja_cedula))

  const cargoOk =
    p.cargo_codigo !== '' &&
    p.respaldado_organizacion !== null &&
    (!p.respaldado_organizacion || (p.organizacion_nombre.trim() !== '' && p.organizacion_documento !== ''))

  const familiaresOk = p.tiene_familiares_iess ? p.familiares.length > 0 : p.estado !== 'NUEVA'

  const inhabilidadesOk =
    o.inhabilidades.length > 0 &&
    o.inhabilidades.every((q) => p.inhabilidades.some((r) => r.codigo === q.codigo))

  return [
    { id: 'personal', titulo: 'Información personal', completa: parejaOk },
    { id: 'domicilio', titulo: 'Domicilio y contacto', completa: !tieneErrores(validarDomicilio(domicilioDe(p))) },
    { id: 'familiares', titulo: 'Familiares', completa: familiaresOk },
    { id: 'cargo', titulo: 'Cargo', completa: cargoOk },
    { id: 'formacion', titulo: 'Formación', completa: p.formacion.length > 0 },
    { id: 'experiencia', titulo: 'Experiencia', completa: p.experiencia.length > 0 },
    { id: 'publicaciones', titulo: 'Publicaciones', completa: p.publicaciones.length > 0 },
    { id: 'capacitacion', titulo: 'Capacitación', completa: p.capacitaciones.length > 0 },
    { id: 'inhabilidades', titulo: 'Inhabilidades', completa: inhabilidadesOk },
    { id: 'envio', titulo: 'Declaración y envío', completa: p.estado === 'ENVIADA' },
  ]
}