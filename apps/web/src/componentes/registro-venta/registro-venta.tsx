export default function RegistroVenta() {
  return (
    <main className="min-h-screen bg-slate-100 p-6">
      <div className="mx-auto max-w-7xl">
        <header className="mb-8">
          <p className="text-sm font-semibold uppercase tracking-wider text-blue-600">
            Grupo 9 · Sistema de ventas
          </p>

          <h1 className="mt-2 text-3xl font-bold text-slate-900">
            Registro de venta
          </h1>

          <p className="mt-2 text-slate-500">
            Gestiona los productos y registra nuevas ventas.
          </p>
        </header>

        <section className="rounded-xl border border-slate-200 bg-white p-8 shadow-sm">
          <h2 className="text-xl font-semibold text-slate-800">
            Bienvenido
          </h2>

          <p className="mt-2 text-slate-600">
            La aplicación de registro de ventas está iniciada.
            Próximamente podrás agregar productos, gestionar cantidades,
            seleccionar el medio de pago y confirmar una venta.
          </p>

          <div className="mt-6 rounded-lg bg-blue-50 p-4 text-sm text-blue-800">
            Entorno de desarrollo preparado. Los servicios de productos
            y stock se integrarán mediante sus respectivas APIs.
          </div>
        </section>
      </div>
    </main>
  );
}