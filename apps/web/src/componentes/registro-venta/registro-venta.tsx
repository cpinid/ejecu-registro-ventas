import { useForm, useFieldArray } from "react-hook-form"
import { useRegistroVenta } from "./use-registro-venta"
import type { Entrada } from "./registro-venta.api"
export function RegistroVenta({
  urlBase,
  puedeOperar,
}: {
  urlBase: string
  puedeOperar: boolean
}) {
  const { register, control, handleSubmit } = useForm<Entrada>({
    defaultValues: {
      medio_pago: "efectivo",
      productos: [{ producto_id: "", cantidad: 1 }],
    },
  })
  const { fields, append, remove } = useFieldArray({ control, name: "productos" })
  const { enviar, pendiente, mutacion } = useRegistroVenta(urlBase)
  return (
    <main className="mx-auto max-w-3xl p-6">
      <h1 className="text-2xl font-bold">Registrar venta</h1>
      <form className="grid gap-4 mt-4" onSubmit={handleSubmit(enviar)}>
        <fieldset
          disabled={!puedeOperar || !!pendiente || mutacion.isPending}
          className="grid gap-3"
        >
          {fields.map((campo, i) => (
            <div key={campo.id} className="flex flex-wrap gap-2">
              <label>
                UUID producto{" "}
                <input required {...register(`productos.${i}.producto_id`)} />
              </label>
              <label>
                Cantidad{" "}
                <input
                  type="number"
                  min="1"
                  max="1000000"
                  required
                  {...register(`productos.${i}.cantidad`, { valueAsNumber: true })}
                />
              </label>
              <button
                type="button"
                disabled={fields.length === 1}
                onClick={() => remove(i)}
              >
                Quitar
              </button>
            </div>
          ))}
          <button
            type="button"
            disabled={fields.length >= 100}
            onClick={() => append({ producto_id: "", cantidad: 1 })}
          >
            Agregar producto
          </button>
          <label>
            Medio de pago{" "}
            <select {...register("medio_pago")}>
              <option value="efectivo">Efectivo</option>
              <option value="debito">Débito</option>
              <option value="credito">Crédito</option>
              <option value="transferencia">Transferencia</option>
            </select>
          </label>
        </fieldset>
        <button disabled={!puedeOperar || mutacion.isPending} type="submit">
          {pendiente ? "Reintentar misma venta" : "Registrar venta"}
        </button>
      </form>
      {mutacion.isError && <p role="alert">{mutacion.error.message}</p>}
      {mutacion.data && (
        <section className="mt-6">
          <h2>Venta confirmada</h2>
          <p>{mutacion.data.id}</p>
          <ul>
            {(mutacion.data.detalles ?? []).map((d) => (
              <li key={d.producto_id}>
                {d.nombre} × {d.cantidad}: ${d.subtotal}
              </li>
            ))}
          </ul>
          <strong>Total: ${mutacion.data.total}</strong>
        </section>
      )}
    </main>
  )
}
