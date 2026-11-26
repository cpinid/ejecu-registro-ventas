-- Los detalles y la auditoría solo se insertan al crear la venta de demostración.
WITH demo AS (
  INSERT INTO ventas (vendedor_id, clave, huella, medio_pago, estado, total)
  VALUES ('22222222-2222-4222-8222-222222222222', 'seed-demo', 'demo',
          'efectivo', 'confirmada', 1500)
  ON CONFLICT (vendedor_id, clave) DO NOTHING
  RETURNING id, vendedor_id
), detalle AS (
  INSERT INTO detalles_venta
      (venta_id, producto_id, nombre, cantidad, precio_unitario, costo_unitario, subtotal)
  SELECT id, '33333333-3333-4333-8333-333333333333', 'Cuaderno de prueba', 1, 1500, 900, 1500
  FROM demo
)
INSERT INTO auditoria (venta_id, usuario_id, accion)
SELECT id, vendedor_id, 'confirmada' FROM demo;
