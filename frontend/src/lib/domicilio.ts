import { validarCorreo } from './validaciones'

export interface Ubicacion {
  paisId: number | null
  provinciaId: number | null
  provinciaTexto: string
  cantonId: number | null
  cantonTexto: string
}

export interface Domicilio extends Ubicacion {
  telefonoCelular: string
  correo2: string
  callePrincipal: string
  numero: string
  calleSecundaria: string
  sector: string
  referencia: string
  telefonoDomicilio: string
}

export const DOMICILIO_VACIO: Domicilio = {
  paisId: null,
  provinciaId: null,
  provinciaTexto: '',
  cantonId: null,
  cantonTexto: '',
  telefonoCelular: '',
  correo2: '',
  callePrincipal: '',
  numero: '',
  calleSecundaria: '',
  sector: '',
  referencia: '',
  telefonoDomicilio: '',
}

export interface ErroresDomicilio {
  pais?: string
  provincia?: string
  canton?: string
  telefonoCelular?: string
  correo2?: string
  callePrincipal?: string
  telefonoDomicilio?: string
}

const RE_TELEFONO = /^\+?\d{7,15}$/

export function validarDomicilio(d: Domicilio): ErroresDomicilio {
  const e: ErroresDomicilio = {}

  if (d.paisId === null) e.pais = 'Seleccione un país'
  if (d.provinciaId === null && d.provinciaTexto.trim() === '') {
    e.provincia = 'Seleccione o escriba la provincia / estado'
  }
  if (d.cantonId === null && d.cantonTexto.trim() === '') {
    e.canton = 'Seleccione o escriba la ciudad / cantón'
  }

  if (d.telefonoCelular.trim() === '') {
    e.telefonoCelular = 'Ingrese su teléfono celular'
  } else if (!RE_TELEFONO.test(d.telefonoCelular.trim())) {
    e.telefonoCelular = 'Use solo números (de 7 a 15 dígitos)'
  }

  if (d.correo2.trim() !== '') {
    const errorCorreo = validarCorreo(d.correo2)
    if (errorCorreo) e.correo2 = errorCorreo
  }

  if (d.callePrincipal.trim() === '') e.callePrincipal = 'Ingrese la calle principal'

  if (d.telefonoDomicilio.trim() !== '' && !RE_TELEFONO.test(d.telefonoDomicilio.trim())) {
    e.telefonoDomicilio = 'Use solo números (de 7 a 15 dígitos)'
  }

  return e
}

export function tieneErrores(e: ErroresDomicilio): boolean {
  return Object.values(e).some((v) => v !== undefined)
}