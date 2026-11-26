# Registro de ventas — Grupo 9
Base de desarrollo para Windows/PowerShell. NO es una entrega validada. Requiere la plantilla oficial y los contratos externos.

## Generación
1. Instalar Go 1.26.5, Node 22.22.0, pnpm 10.13.1.
2. `pnpm install` (primera vez para generar lockfile), `pnpm generate`.
3. En apps/api: `go mod tidy`, `go test ./...`.
4. En raíz: `pnpm codegen`, `pnpm typecheck`, `pnpm test`.
5. Completar .env, fijar digests Docker y ejecutar `docker compose up -d --build`.

## Decisiones provisionales
Puerto 5140/5141; cookie sesion, claims permisos/sub, permisos ventas:operar y ventas:consultar; rutas de proveedores y precio de venta; medio de pago; nombre de módulo. Confirmar con grupos 3, 6 y 8. No almacenar JWT en el navegador.

## Pendiente obligatorio
Plantilla oficial; clientes generados oapi-codegen; lockfiles y contrato generados en entorno real; pruebas E2E, golden, acceso_test.go y cobertura; MSW y app local, paquete npm publicable; CI oficial, SonarQube, Trivy, publicación, SBOM/firma; seed y healthchecks de contenedor; reconciliación y reversa robustas; confirmar versiones y digests. No se pudo compilar el proyecto completo ni ejecutar Docker en el entorno de generación. No presentar como entrega final.
