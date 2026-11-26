import createClient from 'openapi-fetch'
import type { paths } from '../../tipos'
export type Entrada = paths['/api/ventas']['post']['requestBody']['content']['application/json']
export function clienteVentas(urlBase: string) {
  return createClient<paths>({ baseUrl: urlBase, credentials: 'include' })
}
export async function registrar(urlBase: string, body: Entrada, clave: string) {
  const { data, error } = await clienteVentas(urlBase).POST('/api/ventas', { body, params: { header: { 'Idempotency-Key': clave } } })
  if (error || !data) throw Error('No se confirmó la venta. Reintenta con los mismos datos o verifica el estado.')
  return data
}
