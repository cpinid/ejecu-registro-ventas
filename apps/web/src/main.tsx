import { createRoot } from "react-dom/client"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RegistroVenta } from "./componentes/registro-venta/registro-venta"
import "./estilos.css"
const cliente = new QueryClient()
createRoot(document.getElementById("root")!).render(
  <QueryClientProvider client={cliente}>
    <RegistroVenta urlBase="" puedeOperar />
  </QueryClientProvider>,
)
