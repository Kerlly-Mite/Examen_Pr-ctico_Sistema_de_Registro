import { tieneErrores, validarDomicilio } from './domicilio'
import type { Domicilio } from './domicilio'
import type { OpcionesPostulacion, Postulacion, PostulacionReq } from '../types/postulacion'

// Arma el cuerpo del PUT con solo lo que el servidor acepta (sin los datos de solo lectura ni los calculados)
export function aReq(p: Postulacion): PostulacionReq {
  return {
    pareja_nombres: p.pareja_nombres,
    pareja_cedula: p.pareja_cedula,
    pais_id: p.pais_id,
    provincia_id: p.provincia_id,
    canton_id: p.canton_id,
    provincia_texto: p.provincia_texto,
    ciudad_texto: p.ciudad_texto,
    telefono_celular: p.telefono_celular,
    correo_2: p.correo_2,
    calle_principal: p.calle_principal,
    numero: p.numero,
    calle_secundaria: p.calle_secundaria,
    sector: p.sector,
    referencia: p.referencia,
    telefono_domicilio: p.telefono_domicilio,
    tiene_familiares_iess: p.tiene_familiares_iess,
    familiares: p.familiares,
    cargo_codigo: p.cargo_codigo,
    respaldado_organizacion: p.respaldado_organizacion,
    organizacion_nombre: p.organizacion_nombre,
    organizacion_documento: p.organizacion_documento,
    formacion: p.formacion,
    experiencia: p.experiencia.map((e) => ({
      tipo: e.tipo,
      cargo: e.cargo,
      institucion: e.institucion,
      fecha_inicio: e.fecha_inicio,
      fecha_fin: e.fecha_fin,
      actividades: e.actividades,
      descripcion_funciones: e.descripcion_funciones,
      documento_respaldo: e.documento_respaldo,
    })),
    publicaciones: p.publicaciones,
    capacitaciones: p.capacitaciones,
    inhabilidades: p.inhabilidades,
  }
}

export function domicilioDe(p: Postulacion): Domicilio {
  return {
    paisId: p.pais_id,
    provinciaId: p.provincia_id,
    provinciaTexto: p.provincia_texto,
    cantonId: p.canton_id,
    cantonTexto: p.ciudad_texto,
    telefonoCelular: p.telefono_celular,
    correo2: p.correo_2,
    callePrincipal: p.calle_principal,
    numero: p.numero,
    calleSecundaria: p.calle_secundaria,
    sector: p.sector,
    referencia: p.referencia,
    telefonoDomicilio: p.telefono_domicilio,
  }
}

// 1.1.7: casado/a o unión de hecho piden los datos de la pareja
export function requierePareja(estadoCivil: string): boolean {
  return /CASAD|UNION/.test(estadoCivil.toUpperCase())
}

export interface Pendientes {
  obligatorios: string[]
  recomendaciones: string[]
}

// Mismas reglas que valida el servidor al enviar (POST /postulacion/enviar)
export function calcularPendientes(p: Postulacion, o: OpcionesPostulacion): Pendientes {
  const obligatorios: string[] = []
  const recomendaciones: string[] = []

  if (requierePareja(p.estado_civil) && (p.pareja_nombres.trim() === '' || !/^\d{10}$/.test(p.pareja_cedula))) {
    obligatorios.push('Datos de la pareja (información personal)')
  }
  if (tieneErrores(validarDomicilio(domicilioDe(p)))) {
    obligatorios.push('Información domiciliaria y de contacto')
  }
  if (p.tiene_familiares_iess && p.familiares.length === 0) {
    obligatorios.push('Familiares en el IESS o BIESS (agregue al menos uno)')
  }
  if (p.cargo_codigo === '') {
    obligatorios.push('Selección del cargo al que postula')
  } else if (p.respaldado_organizacion === null) {
    obligatorios.push('Respuesta sobre el respaldo de una organización')
  } else if (
    p.respaldado_organizacion &&
    (p.organizacion_nombre.trim() === '' || p.organizacion_documento === '')
  ) {
    obligatorios.push('Nombre y documento de la organización que lo respalda')
  }

  const sinResponder = o.inhabilidades.filter((q) => !p.inhabilidades.some((r) => r.codigo === q.codigo))
  if (sinResponder.length > 0) {
    obligatorios.push(`Inhabilidades sin responder (${sinResponder.length} de ${o.inhabilidades.length})`)
  }
  const inhabilitantes = o.inhabilidades.filter((q) => {
    const r = p.inhabilidades.find((x) => x.codigo === q.codigo)
    return q.impide_postular && r !== undefined && r.respuesta !== q.respuesta_habilitante
  })
  if (inhabilitantes.length > 0) {
    obligatorios.push(
      `Respuestas de inhabilidades que impiden postular: ${inhabilitantes.map((q) => q.codigo).join(', ')}`,
    )
  }

  if (p.formacion.length === 0) obligatorios.push('Formación académica (agregue al menos una)')
  if (p.experiencia.length === 0) obligatorios.push('Experiencia profesional (agregue al menos una)')

  return { obligatorios, recomendaciones }
}