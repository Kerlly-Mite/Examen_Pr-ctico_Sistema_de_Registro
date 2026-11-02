import { api } from './client'
import { RUTAS } from './rutas'
import type { Canton, Pais, Provincia } from '../types/api'

// Go puede devolver null cuando una lista está vacía, por eso data admite null
interface RespuestaCatalogo<T> {
  data: T[] | null
}

export async function obtenerPaises(): Promise<Pais[]> {
  const r = await api.get<RespuestaCatalogo<Pais>>(RUTAS.paises)
  return r.data ?? []
}

export async function obtenerProvincias(paisId: number): Promise<Provincia[]> {
  const r = await api.get<RespuestaCatalogo<Provincia>>(RUTAS.provincias(paisId))
  return r.data ?? []
}

export async function obtenerCantones(provinciaId: number): Promise<Canton[]> {
  const r = await api.get<RespuestaCatalogo<Canton>>(RUTAS.cantones(provinciaId))
  return r.data ?? []
}