La prueba de integración HTTP está en [ventas_e2e_test.go](../../cmd/api/ventas_e2e_test.go).
Usa PostgreSQL temporal con Testcontainers, las migraciones de producción y
proveedores simulados. Desde la raíz, `pnpm test:e2e:api` genera `coverage.out`.
