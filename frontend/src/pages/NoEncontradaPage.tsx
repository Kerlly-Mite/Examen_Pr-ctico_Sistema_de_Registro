import { Link } from 'react-router-dom'

export function NoEncontradaPage() {
  return (
    <section className="rounded-lg bg-white p-6 text-center shadow">
      <h1 className="text-2xl font-semibold">404</h1>
      <p className="mt-2 text-slate-600">La página que busca no existe.</p>
      <Link to="/" className="mt-4 inline-block text-biess-700 underline">
        Volver al inicio
      </Link>
    </section>
  )
}