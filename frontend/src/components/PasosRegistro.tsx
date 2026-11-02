interface Props {
  paso: 1 | 2
}

const TITULOS = ['Verificar identidad', 'Crear cuenta']

export function PasosRegistro({ paso }: Props) {
  return (
    <div>
      <p className="text-xs text-slate-600">
        Paso {paso} de 2 · {TITULOS[paso - 1]}
      </p>
      <div className="mt-1 h-1.5 w-full rounded bg-slate-200">
        <div className="h-1.5 rounded bg-oro-500 transition-all" style={{ width: `${paso * 50}%` }} />
      </div>
    </div>
  )
}