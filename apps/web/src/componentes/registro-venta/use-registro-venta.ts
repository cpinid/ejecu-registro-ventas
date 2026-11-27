import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { registrar, type Entrada } from "./registro-venta.api"
export function useRegistroVenta(urlBase: string) {
  const [pendiente, setPendiente] = useState<{ entrada: Entrada; clave: string }>()
  const mutacion = useMutation({
    mutationFn: (p: { entrada: Entrada; clave: string }) =>
      registrar(urlBase, p.entrada, p.clave),
    onSuccess: () => setPendiente(undefined),
  })
  function enviar(entrada: Entrada) {
    if (mutacion.isPending) return
    const p = pendiente ?? { entrada, clave: crypto.randomUUID() }
    setPendiente(p)
    mutacion.mutate(p)
  }
  return { enviar, pendiente, mutacion }
}
