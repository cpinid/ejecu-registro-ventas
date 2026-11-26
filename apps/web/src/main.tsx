import { createRoot } from "react-dom/client"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RegistroVenta } from "./componentes/registro-venta/registro-venta"
import "./estilos.css"
const cliente = new QueryClient()
async function iniciar() {
  if (import.meta.env.DEV && import.meta.env.VITE_MSW !== "false") {
    const { worker } = await import("./mocks/browser")
    await worker.start({ onUnhandledRequest: "bypass" })
  }
  createRoot(document.getElementById("root")!).render(
    <QueryClientProvider client={cliente}>
      <RegistroVenta urlBase="" puedeOperar />
    </QueryClientProvider>,
  )
}
void iniciar()
