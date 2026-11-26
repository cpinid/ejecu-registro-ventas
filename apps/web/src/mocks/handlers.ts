import { http, HttpResponse } from "msw"
import type { components } from "../tipos"

type Venta = components["schemas"]["Venta"]
type Solicitud = components["schemas"]["Solicitud"]
export const productoDemo = "33333333-3333-4333-8333-333333333333"

function problema(status: number, title: string) {
  return HttpResponse.json(
    { status, title },
    {
      status,
      headers: { "Content-Type": "application/problem+json" },
    },
  )
}

export function handlers() {
  const ventas = new Map<string, Venta>()
  const claves = new Map<string, { huella: string; venta: Venta }>()
  let stock = 10
  return [
    http.post("*/api/ventas", async ({ request }) => {
      const clave = request.headers.get("Idempotency-Key")
      const body = (await request.json().catch(() => null)) as Solicitud | null
      if (
        !clave ||
        clave.length > 128 ||
        !body ||
        !["efectivo", "debito", "credito", "transferencia"].includes(body.medio_pago) ||
        !Array.isArray(body.productos) ||
        !body.productos.length ||
        body.productos.length > 100 ||
        body.productos.some(
          (p) =>
            !p || !Number.isInteger(p.cantidad) || p.cantidad < 1 || p.cantidad > 1000000,
        )
      ) {
        return problema(422, "Solicitud inválida")
      }
      if (body.productos.some((p) => p.producto_id !== productoDemo))
        return problema(404, "Producto desconocido")
      const cantidad = body.productos.reduce((total, p) => total + p.cantidad, 0)
      const huella = JSON.stringify([body.medio_pago, cantidad])
      const previa = claves.get(clave)
      if (previa)
        return previa.huella === huella
          ? HttpResponse.json(previa.venta, { status: 201 })
          : problema(409, "Clave reutilizada con otros datos")
      if (cantidad > stock) return problema(409, "Stock insuficiente")
      const total = cantidad * 1500
      const venta: Venta = {
        id: crypto.randomUUID(),
        vendedor_id: "22222222-2222-4222-8222-222222222222",
        fecha: new Date().toISOString(),
        medio_pago: body.medio_pago,
        estado: "confirmada",
        total,
        detalles: [
          {
            producto_id: productoDemo,
            nombre: "Cuaderno de prueba",
            cantidad,
            precio_unitario: 1500,
            costo_unitario: 900,
            subtotal: total,
          },
        ],
      }
      stock -= cantidad
      ventas.set(venta.id, venta)
      claves.set(clave, { huella, venta })
      return HttpResponse.json(venta, { status: 201 })
    }),
    http.get("*/api/ventas", ({ request }) => {
      const query = new URL(request.url).searchParams
      const limite = Number(query.get("limite") ?? 50)
      const desde = Number(query.get("desplazamiento") ?? 0)
      if (
        !Number.isInteger(limite) ||
        limite < 1 ||
        limite > 100 ||
        !Number.isInteger(desde) ||
        desde < 0
      ) {
        return problema(422, "Paginación inválida")
      }
      return HttpResponse.json(
        [...ventas.values()].reverse().slice(desde, desde + limite),
      )
    }),
    http.get("*/api/ventas/:id", ({ params }) => {
      const venta = ventas.get(String(params.id))
      return venta ? HttpResponse.json(venta) : problema(404, "Venta desconocida")
    }),
  ]
}
