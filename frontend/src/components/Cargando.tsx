export function Cargando({ texto = 'Cargando...' }: { texto?: string }) {
  return (
    <div role="status" className="flex items-center justify-center gap-3 py-10 text-slate-600">
      <span className="h-5 w-5 animate-spin rounded-full border-2 border-biess-100 border-t-biess-800" />
      <span>{texto}</span>
    </div>
  )
}