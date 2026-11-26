# Registro de ventas — Grupo 9

El servicio registra ventas y consulta su detalle. Productos entrega el precio y el
costo de cada ítem. Stock recibe un descuento idempotente por venta.

[Pipeline de GitHub Actions](https://github.com/cpinid/ejecu-registro-ventas/actions/workflows/ci.yml) · [Contrato OpenAPI 3.1](contracts/openapi.yaml) · [Modos de falla](docs/modos-de-falla.md)

## Desarrollo local

Se usan Go 1.26.5, Node 22.22.0, pnpm 10.13.1 y Docker Compose. En la raíz del
repositorio, ejecutar `pnpm install --frozen-lockfile` y `pnpm exec lefthook install`.
El pre-commit usa Gitleaks v8.30.1. Se puede instalar con
`go install github.com/zricethezav/gitleaks/v8@v8.30.1` si el directorio de
binarios de Go está en el PATH.

Para levantar el entorno completo, copiar `.env.example` a `.env` en la primera
instalación y asignar una contraseña local alfanumérica a `POSTGRES_PASSWORD`.
El archivo de ejemplo deriva `DATABASE_URL` de esa contraseña. En una instalación
existente, verificar que la URL apunte a la base `ventas` del servicio `postgres`
con el usuario `ventas`. El archivo Compose de desarrollo configura Identidad,
Productos y Stock contra el simulador.

```sh
docker compose -f docker-compose.yaml -f docker-compose.dev.yaml --profile dev up -d --build
```

Abrir la [sesión de desarrollo](http://localhost:5142/sesion-dev) para entrar a la
[web](http://localhost:5140). El producto de prueba es
`33333333-3333-4333-8333-333333333333`, con precio unitario de 1500 y stock inicial
de 10. El perfil `dev` también carga una venta histórica de ejemplo y habilita
[Mailpit](http://localhost:8025). La [documentación interactiva](http://localhost:5141/api/docs)
y el [contrato servido por la API](http://localhost:5141/api/openapi.yaml) están
disponibles en desarrollo.

Para trabajar solo en la interfaz, ejecutar `pnpm --filter registro-ventas-web dev`.
MSW simula las tres operaciones en el navegador. Sus datos duran hasta recargar la
página. Para usar la API desde Vite, iniciar Vite con `VITE_MSW=false`.

## Verificación

| Comando desde la raíz            | Resultado                                                                                                                                  |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `pnpm verify`                    | Lint, TypeScript, pruebas web con LCOV, pruebas Go y compilación web                                                                       |
| `pnpm test:e2e:api`              | Arranca PostgreSQL temporal con Testcontainers, aplica Goose y verifica alta, reintento, conflicto y consultas HTTP. Genera `coverage.out` |
| `pnpm codegen`                   | Actualiza el contrato de ventas y los tipos TypeScript                                                                                     |
| `make clientes`                  | Regenera los clientes Go de Productos y Stock con oapi-codegen v2.8.2                                                                      |
| `pnpm generate`                  | Regenera el repositorio Go a partir de las consultas SQL                                                                                   |
| `./scripts/probar-venta-dev.ps1` | Verifica el flujo contra el stack Compose y comprueba el descuento de stock                                                                |

El job `calidad` entrega `coverage.out` y `apps/web/coverage/lcov.info` al escáner.
La clave de SonarQube se mantiene en [sonar-project.properties](sonar-project.properties).
La ruta del panel en el servidor del equipo es
`/dashboard?id=cpinid_ejecu-registro-ventas_aebd780b-ce4f-42e4-b388-f81b08e90a55`.
Para el análisis remoto, GitHub Actions requiere `SONAR_HOST_URL` y `SONAR_TOKEN`.

Go 1.26.5 sigue fijado por el DAS. El escaneo de archivos pasó, pero Trivy detectó
12 vulnerabilidades de la biblioteca estándar por imagen y govulncheck encontró 17
alcanzables en la API. El detalle y el efecto en CI figuran en la
[revisión de hitos](docs/revision-hitos-0-1.md#bloqueo-de-seguridad-con-go-1265).

## Revisión de arquitectura

El handler Huma declara el permiso de cada operación en OpenAPI como `x-permiso`.
Crear requiere `ventas:operar`. Listar y consultar requieren `ventas:consultar`.
La matriz de acceso contrasta las tres rutas con cinco perfiles. El servicio coordina
Productos y Stock mediante interfaces. Store controla las transacciones y delega
las consultas en el repositorio generado por sqlc.

La prueba E2E ejercita el handler, el servicio, los clientes HTTP y PostgreSQL.
Verifica que repetir una clave conserva el ID de venta y produce una sola salida
de stock. Los [contratos de desarrollo](contracts/externos/README.md) describen el
simulador. La revisión con los equipos proveedores queda pendiente de sus contratos
oficiales, en particular para autenticación y reversa de stock.
