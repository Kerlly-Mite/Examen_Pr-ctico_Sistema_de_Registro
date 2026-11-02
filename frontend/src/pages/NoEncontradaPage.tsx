import { Link } from 'react-router-dom'

export function NoEncontradaPage() {
  return (
    <section className="mx-auto w-full max-w-md rounded-lg border border-slate-200 bg-white p-8 text-center shadow-sm">
      <p className="text-5xl font-bold text-oro-500">404</p>
      <h1 className="mt-2 text-xl font-semibold text-biess-800">Página no encontrada</h1>
      <p className="mt-1 text-sm text-slate-600">La página que busca no existe o fue movida.</p>
      <Link
        to="/"
        className="mt-5 inline-block rounded bg-biess-800 px-5 py-2 font-medium text-white hover:bg-biess-900"
      >
        Volver al inicio
      </Link>
    </section>
  )
}