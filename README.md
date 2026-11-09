# Registro de Venta — Grupo 9

Microservicio encargado de gestionar el registro de ventas, administrar el detalle de cada operación y coordinar la actualización del stock de productos.

## 1. Descripción del proyecto

**Registro de Venta** es un microservicio desarrollado como parte del proyecto de Desarrollo de Aplicaciones de Software (DAS).

Su responsabilidad principal es registrar las ventas realizadas, incluyendo los productos involucrados, sus cantidades, los montos correspondientes y el medio de pago utilizado.

El servicio se comunica con otros microservicios mediante APIs REST, sin acceder directamente a sus bases de datos.

## 2. Responsabilidades

Las funcionalidades principales del microservicio son:

- **Detalle de venta:** registrar y consultar la información de cada venta, incluyendo los productos y cantidades.
- **Medio de pago:** registrar el medio de pago utilizado en cada operación.
- **Descuento de stock:** coordinar la actualización de las existencias de productos después de confirmar una venta.

## 3. Integración con otros microservicios

El servicio requiere comunicarse con los siguientes grupos:

| Grupo | Microservicio | Responsabilidad |
|---|---|---|
| Grupo 3 | Productos | Proporcionar información de los productos. |
| Grupo 8 | Stock | Gestionar y actualizar las existencias. |
| Grupo 9 | Registro de Venta | Registrar y consultar las ventas. |

Las integraciones se realizarán mediante contratos OpenAPI y solicitudes HTTP REST.

Durante el desarrollo se utilizarán respuestas simuladas (mocks) para permitir el trabajo independiente de los distintos grupos.

## 4. Tecnologías utilizadas

### Backend
- Go 1.26.5
- Huma v2.34.1
- PostgreSQL 17.11
- pgx v5.10.0
- sqlc v1.29.0
- Goose v3.24.1

### Frontend
- React 19.2.8
- TypeScript 6.0.3
- Vite 8.1.5
- Tailwind CSS 4.3.3
- Base UI 1.6.0
- TanStack Query 5.101.4
- React Hook Form 7.83.0
- MSW 2.15.0

### Infraestructura y herramientas
- Docker y Docker Compose
- Git y GitHub
- pnpm 10.13.1
- Lefthook
- GitHub Actions
- SonarQube

## 5. Arquitectura

El proyecto sigue los lineamientos de Clean Architecture establecidos en el DAS.

### Backend

La organización del backend separa las responsabilidades en las siguientes capas:

- **API / Handler:** recibe las solicitudes HTTP, valida los datos de entrada y construye las respuestas.
- **Service:** contiene las reglas de negocio y coordina las operaciones.
- **Store:** gestiona el acceso a la base de datos.
- **Repository:** contiene las consultas generadas mediante sqlc.

Flujo general:

`Handler → Service → Store → Repository → PostgreSQL`

### Frontend

El frontend utiliza microcomponentes independientes, separando:

- Vista (`.tsx`): presentación de la interfaz.
- API (`.api.ts`): comunicación con los servicios.
- Hook (`use-*.ts`): gestión del estado y operaciones del componente.

## 6. Estructura del repositorio

```text
registro-venta/
├── apps/
│   ├── api/
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── api/
│   │   │   ├── service/
│   │   │   ├── store/
│   │   │   └── repository/
│   │   ├── db/
│   │   │   ├── migrations/
│   │   │   └── queries/
│   │   ├── e2e/
│   │   ├── Dockerfile
│   │   └── Dockerfile.migrate
│   │
│   └── web/
│       └── src/
│           ├── componentes/
│           │   └── registro-venta/
│           └── mocks/
│
├── contracts/
│   └── openapi.yaml
├── docs/
│   └── modos-de-falla.md
├── .github/
│   └── workflows/
│       └── ci.yml
├── docker-compose.yaml
├── .env.example
├── lefthook.yml
├── pnpm-workspace.yaml
├── pnpm-lock.yaml
└── README.md
```

## 7. Configuración del entorno

El proyecto utiliza Docker para ejecutar las herramientas de desarrollo sin requerir la instalación local de Go o pnpm.

Requisitos:

- Docker.
- Docker Compose.
- Git.

### Variables de entorno

Las variables necesarias se documentan en `.env.example`.

Para preparar el archivo local:

```bash
cp .env.example .env
```

Completar los valores correspondientes antes de ejecutar los servicios.

No subir el archivo `.env` al repositorio.

## 8. Desarrollo del frontend

El frontend utiliza Node.js 22.22.0 y pnpm 10.13.1.

Para instalar las dependencias mediante Docker:

```bash
docker run --rm -it \
  -v "$PWD:/workspace" \
  -w /workspace \
  node:22.22.0-bookworm-slim \
  sh -lc "corepack enable && corepack prepare pnpm@10.13.1 --activate && pnpm install"
```

Para ejecutar el entorno de desarrollo:

```bash
docker run --rm -it \
  -p 5190:8080 \
  -v "$PWD:/workspace" \
  -w /workspace/apps/web \
  node:22.22.0-bookworm-slim \
  sh -lc "corepack enable && corepack prepare pnpm@10.13.1 --activate && pnpm dev"
```

El frontend estará disponible en:

`http://localhost:5190`

Para compilar:

```bash
pnpm build
```

Este último comando debe ejecutarse dentro de un entorno que tenga pnpm instalado.

## 9. Contratos API

El contrato de comunicación del microservicio se encuentra en:

`contracts/openapi.yaml`

Este documento define los endpoints, parámetros, esquemas de datos y respuestas de la API.

Los contratos se utilizarán como referencia para implementar los handlers y facilitar la integración con los demás microservicios.

## 10. Estado del desarrollo

El proyecto se encuentra en etapa de configuración y desarrollo inicial.

| Componente | Estado |
|---|---|
| Estructura del repositorio | En preparación |
| Configuración del frontend | Compilación inicial exitosa |
| Interfaz de registro de venta | En desarrollo |
| Contrato OpenAPI | Pendiente de completar |
| Backend Go | En desarrollo |
| Base de datos | Pendiente de implementar |
| Integración con Grupo 3 | Pendiente |
| Integración con Grupo 8 | Pendiente |
| Pruebas automatizadas | Pendientes |

## 11. Consideraciones

- Cada microservicio mantiene su propia base de datos.
- La comunicación entre servicios se realiza mediante APIs REST.
- Los contratos OpenAPI deben definirse antes de implementar las integraciones.
- Las operaciones que modifican datos en otros servicios deben contemplar idempotencia.
- Las dependencias y versiones deben mantenerse alineadas con las especificaciones del DAS.

---

**Grupo 9 — Registro de Venta**

Proyecto desarrollado en el contexto de Desarrollo de Aplicaciones de Software (DAS).