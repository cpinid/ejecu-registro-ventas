import { describe, it, expect } from "vitest"
import { clienteVentas } from "./registro-venta.api"
describe("cliente de ventas", () => {
  it("crea cliente tipado", () => {
    expect(typeof clienteVentas("http://localhost:5141").POST).toBe("function")
  })
})
