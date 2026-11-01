import { useCallback, useEffect, useState } from 'react'

export function useCuentaRegresiva() {
  const [finaliza, setFinaliza] = useState<number | null>(null)
  const [ahora, setAhora] = useState<number>(() => Date.now())

  useEffect(() => {
    if (finaliza === null) return
    const id = window.setInterval(() => {
      const t = Date.now()
      setAhora(t)
      if (t >= finaliza) setFinaliza(null)
    }, 500)
    return () => window.clearInterval(id)
  }, [finaliza])

  const iniciar = useCallback((segundos: number) => {
    const t = Date.now()
    setAhora(t)
    setFinaliza(t + segundos * 1000)
  }, [])

  const restantes = finaliza === null ? 0 : Math.max(0, Math.ceil((finaliza - ahora) / 1000))
  return { restantes, bloqueado: restantes > 0, iniciar }
}