# Modos de falla — Grupo 9

| Falla                              | Protección                                     | Verificación pendiente   |
| ---------------------------------- | ---------------------------------------------- | ------------------------ |
| Cantidades inválidas               | Huma, service, CHECK                           | E2E 422                  |
| Precio alterado por cliente        | Fuente Productos                               | E2E total real           |
| Duplicación                        | UNIQUE vendedor/clave y clave estable en Stock | E2E concurrencia         |
| Misma clave, cuerpo distinto       | Huella canónica                                | E2E 409                  |
| Stock insuficiente                 | Operación atómica externa                      | E2E 409                  |
| Caída de Stock o respuesta perdida | Estado pendiente y reintento misma clave       | E2E 503 y reconciliación |
| Falla local después de descuento   | Reintento idempotente                          | E2E falla intermedia     |
| JWT inválido o permisos faltantes  | JWKS y permisos                                | E2E 401/403              |
| Desbordamiento                     | Validación entera                              | E2E extremos             |
| Caída de base                      | readyz                                         | E2E 503                  |

Pendiente: reconciliación automática, reversa acordada, matriz de roles oficial, pruebas E2E reales y golden. Verificar retirando deliberadamente cada protección.
