export type Errores<F> = Partial<Record<keyof F, string>>

export interface Aviso {
  tipo: 'exito' | 'info' | 'error'
  texto: string
}