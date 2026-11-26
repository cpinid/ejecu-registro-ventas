import { afterAll, afterEach, beforeAll, expect, it } from "vitest"
import { setupServer } from "msw/node"
import { handlers, productoDemo } from "../../mocks/handlers"
import { clienteVentas, registrar, type Entrada } from "./registro-venta.api"

const servidor = setupServer()
const base = "http://ventas.test"
const entrada: Entrada = {
  medio_pago: "debito",
  productos: [{ producto_id: productoDemo, cantidad: 2 }],
}
beforeAll(() => servidor.listen({ onUnhandledRequest: "error" }))
afterEach(() => servidor.resetHandlers())
afterAll(() => servidor.close())

it("registra, reintenta y consulta la misma venta", async () => {
  servidor.use(...handlers())
  const venta = await registrar(base, entrada, "compra")
  expect(venta.total).toBe(3000)
  expect((await registrar(base, entrada, "compra")).id).toBe(venta.id)
  const cliente = clienteVentas(base)
  expect((await cliente.GET("/api/ventas")).data).toEqual([venta])
  expect(
    (await cliente.GET("/api/ventas/{id}", { params: { path: { id: venta.id } } })).data,
  ).toEqual(venta)
  await expect(
    registrar(base, { ...entrada, medio_pago: "credito" }, "compra"),
  ).rejects.toThrow()
})

it("informa el error cuando el producto no existe", async () => {
  servidor.use(...handlers())
  await expect(
    registrar(
      base,
      { ...entrada, productos: [{ producto_id: crypto.randomUUID(), cantidad: 1 }] },
      "compra",
    ),
  ).rejects.toThrow()
})
