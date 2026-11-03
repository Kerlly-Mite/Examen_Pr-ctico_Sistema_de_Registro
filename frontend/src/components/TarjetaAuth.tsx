import type { ReactNode } from 'react'

interface Props {
  titulo: string
  subtitulo?: string
  cabecera?: ReactNode
  pie?: ReactNode
  children: ReactNode
}

export function TarjetaAuth({ titulo, subtitulo, cabecera, pie, children }: Props) {
  return (
    <section className="mx-auto w-full max-w-md rounded-lg border border-slate-200 bg-white p-6 shadow-sm sm:p-8">
      {cabecera}
      <h1 className={`${cabecera ? 'mt-4' : ''} text-xl font-bold text-biess-800`}>{titulo}</h1>
      {subtitulo && <p className="mt-1 text-sm text-slate-600">{subtitulo}</p>}
      <div className="mt-5 space-y-4">{children}</div>
      {pie && <div className="mt-6 border-t border-slate-100 pt-4 text-center text-sm text-slate-600">{pie}</div>}
    </section>
  )
}