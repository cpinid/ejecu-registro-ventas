# Contratos de desarrollo

[Productos](productos.yaml) y [Stock](stock.yaml) describen las rutas que implementa el
[simulador](../../dev/mocks/main.go). El backend usa los clientes Go generados con
oapi-codegen v2.8.2. Para regenerarlos, ejecutar `make clientes` desde la raíz.

Identidad se simula con un JWKS en `/jwks` y tokens de desarrollo en `/token`.
La API valida la firma, el emisor, la audiencia y la caducidad con JWT y keyfunc.
El adaptador reenvía el token como Bearer y mantiene la clave de idempotencia al
descontar stock. Los tres equipos proveedores todavía deben entregar sus contratos
y confirmar la autenticación y la operación de reversa.
