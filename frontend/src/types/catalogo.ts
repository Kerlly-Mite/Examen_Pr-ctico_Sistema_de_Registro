export interface Pais {
  id: number
  nombre: string
}

export interface Provincia {
  id: number
  pais_id: number
  nombre: string
}

export interface Canton {
  id: number
  provincia_id: number
  nombre: string
}
