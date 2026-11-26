-- name: CrearVenta :one
INSERT INTO ventas(vendedor_id, clave, huella, medio_pago, total)
VALUES($1,$2,$3,$4,$5) ON CONFLICT(vendedor_id,clave) DO NOTHING RETURNING *;
-- name: PorClave :one
SELECT * FROM ventas WHERE vendedor_id=$1 AND clave=$2;
-- name: PorID :one
SELECT * FROM ventas WHERE id=$1;
-- name: Detalles :many
SELECT * FROM detalles_venta WHERE venta_id=$1 ORDER BY producto_id;
-- name: CrearDetalle :exec
INSERT INTO detalles_venta(venta_id,producto_id,nombre,cantidad,precio_unitario,costo_unitario,subtotal) VALUES($1,$2,$3,$4,$5,$6,$7);
-- name: Listar :many
SELECT * FROM ventas WHERE estado='confirmada' ORDER BY fecha DESC,id DESC LIMIT $1 OFFSET $2;
-- name: CambiarEstado :one
UPDATE ventas SET estado=$2 WHERE id=$1 AND estado='pendiente' RETURNING *;
-- name: Auditar :exec
INSERT INTO auditoria(venta_id,usuario_id,accion) VALUES($1,$2,$3);
