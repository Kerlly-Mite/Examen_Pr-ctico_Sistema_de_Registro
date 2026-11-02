export interface Duracion {
  anios: number
  meses: number
  dias: number
}

export interface Familiar {
  nombres_apellidos: string
  parentesco: string
  institucion: string
  area: string
  cargo: string
}

export interface Formacion {
  nivel_instruccion: string
  institucion: string
  tipo: string
  titulo: string
  area_estudios: string
  registro_senescyt: string
  fecha_registro_senescyt: string | null
  documento_respaldo: string
}

export interface Experiencia {
  tipo: string
  cargo: string
  institucion: string
  fecha_inicio: string | null
  fecha_fin: string | null
  actividades: string
  descripcion_funciones: string
  documento_respaldo: string
  duracion?: Duracion | null
}

export interface Publicacion {
  titulo: string
  tipo: string
  medio_publicacion: string
  fecha_publicacion: string | null
  relacionada_formacion: boolean
  documento_respaldo: string
}

export interface Capacitacion {
  tipo_evento: string
  nombre_evento: string
  institucion: string
  fecha_inicio: string | null
  fecha_fin: string | null
  horas: number
  documento_respaldo: string
}

export interface RespuestaInhabilidad {
  codigo: string
  respuesta: boolean
  remocion_institucion?: string
  remocion_cargo?: string
  remocion_fecha_fin?: string | null
  remocion_motivo?: string
}

export interface PostulacionReq {
  pareja_nombres: string
  pareja_cedula: string
  pais_id: number | null
  provincia_id: number | null
  canton_id: number | null
  provincia_texto: string
  ciudad_texto: string
  telefono_celular: string
  correo_2: string
  calle_principal: string
  numero: string
  calle_secundaria: string
  sector: string
  referencia: string
  telefono_domicilio: string
  tiene_familiares_iess: boolean
  familiares: Familiar[]
  cargo_codigo: string
  respaldado_organizacion: boolean | null
  organizacion_nombre: string
  organizacion_documento: string
  formacion: Formacion[]
  experiencia: Experiencia[]
  publicaciones: Publicacion[]
  capacitaciones: Capacitacion[]
  inhabilidades: RespuestaInhabilidad[]
}

export interface VerificacionCargo {
  cargo_codigo: string
  cumple: boolean
  estado: string
  mensaje: string
  pregunta: string
}

export interface Bloqueo {
  codigo: string
  mensaje: string
}

export interface Postulacion extends PostulacionReq {
  id: number
  codigo: string
  estado: 'NUEVA' | 'BORRADOR' | 'ENVIADA'
  enviada_en: string | null
  actualizado_en: string | null
  cedula: string
  nombres: string
  apellidos: string
  genero: string
  fecha_nacimiento: string | null
  nacionalidad: string
  estado_civil: string
  correo_1: string
  experiencia_total: Duracion | null
  verificacion_cargo: VerificacionCargo | null
  bloqueo: Bloqueo | null
}

export interface Cargo {
  codigo: string
  nombre: string
  sector: string
}

export interface PreguntaInhabilidad {
  codigo: string
  texto: string
  respuesta_habilitante: boolean
}

export interface OpcionesPostulacion {
  cargos: Cargo[]
  inhabilidades: PreguntaInhabilidad[]
}