-- +goose Up
CREATE TABLE ventas (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 vendedor_id uuid NOT NULL,
 clave text NOT NULL,
 huella text NOT NULL,
 fecha timestamptz NOT NULL DEFAULT now(),
 medio_pago text NOT NULL CHECK (medio_pago IN ('efectivo','debito','credito','transferencia')),
 estado text NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente','confirmada','rechazada')),
 total bigint NOT NULL CHECK (total >= 0),
 UNIQUE(vendedor_id, clave)
);
CREATE TABLE detalles_venta (
 venta_id uuid NOT NULL REFERENCES ventas(id),
 producto_id uuid NOT NULL,
 nombre text NOT NULL,
 cantidad integer NOT NULL CHECK (cantidad > 0),
 precio_unitario bigint NOT NULL CHECK (precio_unitario >= 0),
 costo_unitario bigint NOT NULL CHECK (costo_unitario >= 0),
 subtotal bigint NOT NULL CHECK (subtotal = precio_unitario * cantidad),
 PRIMARY KEY(venta_id, producto_id)
);
CREATE TABLE auditoria (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 venta_id uuid NOT NULL REFERENCES ventas(id),
 usuario_id uuid NOT NULL,
 accion text NOT NULL,
 fecha timestamptz NOT NULL DEFAULT now()
);
-- +goose Down
DROP TABLE auditoria;
DROP TABLE detalles_venta;
DROP TABLE ventas;
