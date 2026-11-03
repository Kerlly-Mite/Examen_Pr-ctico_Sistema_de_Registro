import { fechaHoyISO } from './formato'

export function validarCedula(v: string): string | undefined {
  if (v.trim() === '') return 'Ingrese su cédula'
  if (!/^\d{10}$/.test(v)) return 'La cédula debe tener 10 dígitos'
  return undefined
}

export function validarCodigoDactilar(v: string): string | undefined {
  if (v.trim() === '') return 'Ingrese su código dactilar'
  if (!/^[A-Za-z0-9]{10}$/.test(v)) return 'El código dactilar debe tener 10 caracteres (letras y números, ej. E3333I1222)'
  return undefined
}

export function validarFecha(v: string): string | undefined {
  if (v === '') return 'Ingrese la fecha de expedición'
  if (v > fechaHoyISO()) return 'La fecha no puede ser futura'
  return undefined
}

export function validarCorreo(v: string): string | undefined {
  if (v.trim() === '') return 'Ingrese su correo electrónico'
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v.trim())) return 'El correo no tiene un formato válido'
  return undefined
}

export function validarPassword(v: string): string | undefined {
  if (v.length < 8 || v.length > 72) return 'La contraseña debe tener entre 8 y 72 caracteres'
  if (!/[A-Za-z]/.test(v) || !/\d/.test(v)) return 'La contraseña debe tener al menos una letra y un número'
  return undefined
}

export function validarConfirmacion(password: string, confirmacion: string): string | undefined {
  if (confirmacion !== password) return 'Las contraseñas no coinciden'
  return undefined
}