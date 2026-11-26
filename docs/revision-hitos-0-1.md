# Revisión de Cumplimiento — Hitos 0 y 1 (`ejecu-registro-ventas`)

Revisión del estado actual del repositorio (rama `feat/implementacion-inicial`) frente al documento **DAS — Desarrollo de Microservicios**, verificando todos los requisitos del **Hito 0** y hasta el **Hito 1**.

## Estado comprobado el 26 de noviembre de 2026

| Hito                           | Estado                                  | Evidencia                                                                                                                                                                                                                                                                                                                                                                                     |
| ------------------------------ | --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **0 · Entorno y seguridad**    | **Pendiente de cierre**                 | Pasaron la verificación local, el E2E con PostgreSQL y el arranque en Compose. Se generaron ambos reportes de cobertura. Los tres digests están fijados, dos ejecuciones del seed dejaron una sola venta y Gitleaks detectó el secreto ficticio en el hook. GitHub muestra `main` sin protección. Quedan pendientes el origen desde plantilla, el pipeline remoto en verde y su Quality Gate. |
| **1 · Implementación inicial** | **Validado localmente con simuladores** | El contrato generado declara el permiso de las tres operaciones. Pasaron los 15 casos de la matriz de acceso, la prueba de alta y reintento con PostgreSQL y la prueba del formulario con MSW en Chrome. La API usa clientes oapi-codegen de Productos y Stock basados en el simulador. Quedan pendientes los contratos oficiales y la evaluación de arquitectura del curso.                  |

[Pipeline del repositorio](https://github.com/cpinid/ejecu-registro-ventas/actions/workflows/ci.yml) · [Contrato de ventas](../contracts/openapi.yaml) · [Nota de arquitectura](../README.md#revisión-de-arquitectura)

### Bloqueo de seguridad con Go 1.26.5

El equipo decidió conservar el parche exigido por el DAS. Con las dependencias
transitivas corregidas, Trivy v0.76.1 encontró **0 hallazgos** en el repositorio
con los filtros de CI (`HIGH,CRITICAL`, `--ignore-unfixed`, escáneres de
vulnerabilidades, configuración y secretos).

Al analizar las imágenes construidas, Trivy encontró **12 vulnerabilidades de la
biblioteca estándar de Go en cada imagen**, tanto en la API como en el servidor
Go que sirve la web. Son los mismos 12 identificadores. Por separado,
`govulncheck@v1.9.0 ./...` identificó **17 vulnerabilidades alcanzables** de la
biblioteca estándar en la API. Una de ellas, GO-2026-6859, se corrige en Go 1.26.11.

Con estos resultados, el job `seguridad` fallaría en govulncheck y GitHub omitiría
el job dependiente `imagen`. El escaneo local de imágenes también supera el umbral vigente.
El cierre del Hito 0 requiere resolver la incompatibilidad entre el parche del DAS
y los controles de seguridad, además de la protección de rama y el Quality Gate remoto.

## Auditoría inicial

Las secciones siguientes conservan el diagnóstico anterior a estas correcciones y las observaciones adicionales del DAS. La tabla anterior recoge las verificaciones posteriores.

## 1. Qué ya se cumple en el repositorio

- **Estructura base del servicio:** `apps/api/` (`cmd/api`, `internal/api`, `internal/service`, `internal/store`, `internal/repository`, `db/migrations`, `db/queries`, `e2e/testdata`), `apps/web/`, `contracts/openapi.yaml`, `docs/modos-de-falla.md`, `docker-compose.yaml`, `.env.example`, `lefthook.yml`, `.github/workflows/ci.yml` y `sonar-project.properties`.
- **Clean Architecture (backend):** Separación de capas `Handler` (`apps/api/internal/api/ventas.go`) → `Service` (`apps/api/internal/service/ventas.go`, sin HTTP ni SQL) → `Store` (`apps/api/internal/store/ventas.go`) → `Repository` generado con `sqlc` (`apps/api/internal/repository/`).
- **Contrato OpenAPI 3.1 (API-first):** Generado con Huma en `contracts/openapi.yaml` (endpoints en español y plural: `POST /api/ventas`, `GET /api/ventas`, `GET /api/ventas/{id}`), tipos TypeScript en `apps/web/src/tipos.ts`, y `/api/docs` habilitado cuando `API_ENV != prod` (`apps/api/cmd/api/main.go`).
- **Reglas de datos, comunicación y observabilidad:**
  - Migración `apps/api/db/migrations/00001_ventas.sql` con `uuid`, `bigint` para montos, `timestamptz` para fechas y tabla `auditoria`.
  - Logs JSON con `log/slog` incluyendo `servicio`, `version`, `request_id`, `metodo`, `ruta`, `estado` y `duracion_ms` (`apps/api/cmd/api/main.go`).
  - `GET /healthz` y `GET /readyz` (con ping a PostgreSQL) sin autenticación en la API, y `GET /healthz` en `apps/web/spa-server/main.go`.
  - Validación JWT (`RS256` / `EdDSA`) vía `AUTH_JWKS_URL`, soporte de cookie `sesion` `HttpOnly` y reenvío de `X-Request-Id`, `Idempotency-Key` y timeout de `5s` en `apps/api/internal/clientes/clientes.go`.
  - Errores siguiendo RFC 9457 (`application/problem+json`).
- **Versionado JS/TS y Dockerfiles `api` / `web`:**
  - Versiones exactas fijadas en `pnpm-workspace.yaml` (`catalog:`), `.nvmrc` (`22.22.0`), `.npmrc` (`save-exact=true`, `ignore-scripts=true`), `.prettierrc.json` (`{"semi":false,"printWidth":90}`), `.editorconfig` y `.dockerignore`.
  - `apps/api/Dockerfile` y `apps/web/Dockerfile` cumplen multi-stage, bases con digest (`golang:1.26.5-alpine3.23`, `node:22.22.0-alpine3.23`, `distroless/static-debian13:nonroot`), `USER nonroot:nonroot`, `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, puerto `8080`, y `spa-server` con cabeceras de seguridad y proxy `/api/`.
- **Pipeline definido:** `.github/workflows/ci.yml` define los 3 jobs (`calidad`, `seguridad`, `imagen`) con `ubuntu-24.04`, versiones fijadas (`TRIVY_VERSION: v0.76.1`, `GOVULNCHECK_VERSION: v1.9.0`, `OASDIFF_VERSION: v1.35.0`, `DOCKLE_VERSION: v0.4.15`) y GitHub Actions fijadas por SHA.
- **Modos de falla iniciales:** `docs/modos-de-falla.md` documentado.

---

## 2. Qué falta para lograr el Hito 0 (Entorno y seguridad)

### A. Bloqueantes en el Pipeline de CI (`.github/workflows/ci.yml`)

1. **Falta el script `test:e2e:api` en `package.json`:** `.github/workflows/ci.yml` ejecuta `pnpm test:e2e:api`, pero ese script no está definido en `package.json`, por lo que el job `calidad` fallará al correr.
2. **Falta generación de reportes de cobertura para SonarQube:**
   - `sonar-project.properties` declara `sonar.go.coverage.reportPaths=coverage.out` y `sonar.javascript.lcov.reportPaths=apps/web/coverage/lcov.info`, pero ninguno de los dos se genera actualmente.
   - En Go falta compilar la API con `go build -cover`, correr las pruebas con `GOCOVERDIR` y convertir el resultado con `go tool covdata textfmt -i "$GOCOVERDIR" -o coverage.out`.
   - En `apps/web/package.json`, `vitest run` no emite `lcov` (requiere configurar cobertura y ejecutar con `--coverage`).
3. **Verificar `sonar.projectKey` en `sonar-project.properties`:** Actualmente usa `sonar.projectKey=registro-ventas`; confirmar que coincida exactamente con la clave del proyecto creado en el servidor de SonarQube (de lo contrario fallará el scan / Quality Gate).

### B. Contenedorización y `docker-compose.yaml`

4. **Fijar digests `sha256` obligatorios del Hito 0 (pág. 5 y 8 del DAS):**
   - `docker-compose.yaml` deja sin digest fijado `${POSTGRES_IMAGE:?fijar digest}`, `${ALPINE_IMAGE:?fijar digest}` y `${MAILPIT_IMAGE:?fijar digest}`, y `.env.example` mantiene `replace-with-*-image-and-digest`.
   - `apps/api/Dockerfile.migrate` usa `FROM $ALPINE_IMAGE` (lo que además puede fallar en Hadolint/Trivy por tag no explícito).
   - Se debe obtener el digest multi-arquitectura (`docker buildx imagetools inspect <imagen:tag> --format '{{json .Manifest.Digest}}'`) y fijarlo directamente en:
     - `apps/api/Dockerfile.migrate`: `alpine:3.23.7@sha256:...`
     - `docker-compose.yaml`: `postgres:17.11@sha256:...` y `axllent/mailpit:v1.31.3@sha256:...`
5. **Falta servicio `seed` en el perfil `dev` de `docker-compose.yaml`:** El DAS (pág. 9) exige en el perfil `dev` los servicios `seed` (datos de demo) y `mailpit` (SMTP falso en `:8025`). Actualmente solo está `mailpit`.
6. **Confirmar número de servicio `N` (puertos `51N0`/`51N1`):** `docker-compose.yaml` usa `5140` (web) y `5141` (api) (`N=4`). Confirmar que `4` es el número asignado al servicio en el documento de componentes de Ejecución.
7. **Caché de compilación Go en `apps/web/Dockerfile`:** La etapa `servidor` (`spa-server`) no incluye `--mount=type=cache,target=/root/.cache/go-build` (regla de caché de dependencias de la pág. 7).

### C. Entorno local, hooks y evidencia documental

8. **Instalar hooks locales y probar Gitleaks:**
   - `.git/hooks/` solo tiene archivos `.sample` (no se ha ejecutado `pnpm install` ni `pnpm exec lefthook install`).
   - Instalar `gitleaks` `v8.30.1` localmente y verificar que un commit con un secreto falso queda bloqueado.
9. **Crear `.env` local y probar `docker compose up -d --build`:** Actualmente no existe `.env` ni se ha validado el levantamiento completo en Docker.
10. **Agregar enlaces obligatorios en `README.md` (pág. 20):** Falta agregar el enlace al pipeline de GitHub Actions y el enlace al proyecto de SonarQube (además de limpiar las notas de borrador provisional en `README.md`).
11. **Configuraciones remotas en GitHub y SonarQube (por verificar):**
    - Protección de rama `main`: PR obligatorio, al menos 1 aprobación y checks `calidad`, `seguridad` e `imagen` requeridos.
    - Secrets `SONAR_TOKEN` y `SONAR_HOST_URL` cargados en GitHub Actions Secrets.
    - Definición de **"New Code"** configurada en SonarQube antes del primer análisis.
    - Primer pipeline en verde (`calidad`, `seguridad`, `imagen`).

---

## 3. Qué falta para lograr el Hito 1 (Implementación inicial y revisión de arquitectura)

### A. Arquitectura, Autorización y Contrato API

1. **Declaración de permisos por endpoint y matriz de acceso `acceso_test.go` (pág. 16):**
   - Actualmente los permisos (`ventas:consultar`, `ventas:operar`) se deducen en un middleware de `apps/api/cmd/api/main.go` según `r.Method == http.MethodPost`, pero no se declaran por endpoint en `apps/api/internal/api/ventas.go`.
   - **Falta crear `acceso_test.go`** con la matriz de acceso probada por rol para cada endpoint.
2. **Ubicación y ejemplos de structs del contrato (pág. 2):**
   - El DAS indica definir los structs de request y response en `internal/api/` con tags `json`, validaciones y **ejemplos**.
   - Actualmente `Solicitud`, `Item`, `Venta` y `Detalle` están definidos en `apps/api/internal/service/ventas.go` (acoplando dominio y contrato HTTP) y no tienen tags `example:"..."` para OpenAPI.

### B. Dependencias exactas de Go faltantes (`go.mod` y `Makefile`, pág. 3)

3. **En `apps/api/go.mod`:** Faltan dependencias exactas de la tabla del DAS:
   - `github.com/pressly/goose/v3 v3.24.1` (el DAS pide fijarla en `go.mod`, `Makefile` y `Dockerfile.migrate`).
   - `github.com/testcontainers/testcontainers-go/modules/postgres v0.43.0`.
4. **En `Makefile`:**
   - Falta comando para `github.com/pressly/goose/v3@v3.24.1`.
   - Falta comando para `github.com/oapi-codegen/oapi-codegen/v2@v2.8.2`.

### C. Contratos externos e integración con mocks

5. **Contratos externos en `contracts/externos/` y clientes `oapi-codegen` (pág. 2 y 3):**
   - `contracts/externos/` solo contiene `README.md`. Al disponer de los `openapi.yaml` de proveedores (Identidad, Productos, Stock), deben copiarse a `contracts/externos/<servicio>.yaml` y generar los clientes Go con `oapi-codegen v2.8.2` para sustituir el adaptador provisional `apps/api/internal/clientes/clientes.go` (y usar `prism mock contracts/externos/<servicio>.yaml` en lugar de rutas provisionales en `dev/mocks/main.go`).
   - Confirmar con los grupos proveedores si las llamadas internas esperan `Authorization: Bearer <jwt>` o cookie `sesion` (hoy `main.go` extrae la cookie `sesion` y `clientes.go` la reenvía como `Bearer`), y definir la operación de reversa acordada (pendiente según `docs/modos-de-falla.md`).

### D. Frontend: Microcomponentes, MSW y UI (`apps/web/`)

6. **Mocks MSW en la app de desarrollo local (pág. 7):**
   - El DAS indica que la app de desarrollo local monta los microcomponentes con **mocks MSW**.
   - Actualmente `apps/web/src/` no tiene handlers ni worker de MSW (`msw` está instalado en `devDependencies`, pero `apps/web/src/main.tsx` no inicializa MSW).
7. **Configuración de shadcn / `@base-ui/react` (`components.json` y `src/components/ui/`, pág. 4 y 22):**
   - Falta `apps/web/components.json` (estilo `base-nova` con `@base-ui/react` `1.6.0`) y `src/components/ui/`.
   - El microcomponente `apps/web/src/componentes/registro-venta/registro-venta.tsx` usa elementos HTML nativos (`<input>`, `<button>`, `<select>`) sin componentes UI de `@base-ui/react`.
8. **Cobertura del microcomponente y empaquetado npm:**
   - La API define `POST /api/ventas`, `GET /api/ventas` y `GET /api/ventas/{id}`, pero `apps/web/src/componentes/registro-venta/` solo implementa la creación (`POST /api/ventas`), sin vistas/componentes para listar ni consultar ventas.
   - `apps/web/package.json` usa `"name": "registro-ventas-web"` y `"private": true`; falta preparar la estructura de paquete npm versionado `@<org>/registro-ventas-ui` declarando la versión del contrato API que consume (separado de la app local que no se publica).
9. **Formato Prettier y reglas `oxlint` (pág. 11 y 22):**
   - Varios archivos en `apps/web/src/` (`registro-venta.tsx`, `registro-venta.api.ts`, `use-registro-venta.ts`, `main.tsx`) están escritos en líneas únicas muy largas (>90 caracteres), incumpliendo `printWidth: 90` de `.prettierrc.json`.
   - No existe configuración explícita de `oxlint` (`.oxlintrc.json`) para detectar `fetch` fuera de `*.api.ts`, uso de `localStorage`/`sessionStorage` e imports entre features.

### E. Pruebas E2E y validación técnica ejecutada (pág. 11, 22, 23)

10. **Pruebas E2E reales con `Testcontainers` y reporte golden:**
    - `apps/api/e2e/` solo contiene `testdata/README.md`. Aunque la implementación completa sin mocks se evalúa en Hito 2, el CI actual ya invoca `pnpm test:e2e:api` y exige la cobertura para SonarQube.
    - Faltan también pruebas de lógica con fechas usando `TZ=UTC` y `TZ=Pacific/Auckland` (pág. 22).
11. **Ejecución real pendiente de comandos de verificación:**
    - Ejecutar en entorno con Go 1.26.5, Node 22.22.0, pnpm 10.13.1 y Docker:
      - `pnpm install --frozen-lockfile`
      - `pnpm exec lefthook install`
      - `pnpm generate` (`sqlc`)
      - `pnpm codegen`
      - `pnpm exec prettier --write .`
      - `pnpm verify` (`oxlint`, `typecheck`, `vitest`, `go test ./...`, `vite build`)
      - `cd apps/api && go mod verify`
      - Escaneos locales con `trivy` (`v0.76.1`), `gitleaks` (`v8.30.1`) y `docker compose up -d --build`
