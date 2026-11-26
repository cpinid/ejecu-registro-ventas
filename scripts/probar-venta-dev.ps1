$ErrorActionPreference = 'Stop'
$api = 'http://localhost:5141'
$mock = 'http://localhost:5142'
$token = (Invoke-RestMethod "$mock/token").token
$antes = (Invoke-RestMethod "$mock/inventario-dev").cantidad
if ($antes -lt 2) { throw 'El inventario de prueba tiene menos de dos unidades.' }
$headers = @{ Authorization = "Bearer $token"; 'Idempotency-Key' = [guid]::NewGuid().ToString() }
$body = @{ medio_pago = 'debito'; productos = @(@{ producto_id = '33333333-3333-4333-8333-333333333333'; cantidad = 2 }) } | ConvertTo-Json -Depth 5
$venta = Invoke-RestMethod "$api/api/ventas" -Method Post -Headers $headers -ContentType 'application/json' -Body $body
$repetida = Invoke-RestMethod "$api/api/ventas" -Method Post -Headers $headers -ContentType 'application/json' -Body $body
$despues = (Invoke-RestMethod "$mock/inventario-dev").cantidad
$consulta = Invoke-RestMethod "$api/api/ventas/$($venta.id)" -Headers @{ Authorization = "Bearer $token" }
if ($venta.estado -ne 'confirmada' -or $venta.total -ne 3000 -or $venta.id -ne $repetida.id -or $consulta.id -ne $venta.id -or ($antes - $despues) -ne 2) { throw 'Fallo la comprobacion de venta o idempotencia.' }
[pscustomobject]@{ Estado = $venta.estado; Total = $venta.total; StockAntes = $antes; StockDespues = $despues; ReintentoMismaVenta = $true; ConsultaCorrecta = $true }
