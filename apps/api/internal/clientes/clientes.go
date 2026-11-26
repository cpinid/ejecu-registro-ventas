package clientes

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"registro-ventas/internal/clientes/productos"
	"registro-ventas/internal/clientes/stock"
	"registro-ventas/internal/service"
)

type claveContexto int

const claveCredencial claveContexto = 0

type Credencial struct{ JWT, RequestID string }

func ConCredencial(ctx context.Context, c Credencial) context.Context {
	return context.WithValue(ctx, claveCredencial, c)
}

type Cliente struct {
	productos productos.Client
	stock     stock.Client
}

func Nuevo(p, s string) Cliente {
	httpClient := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return Cliente{
		productos: productos.Client{Server: strings.TrimRight(p, "/") + "/", Client: httpClient},
		stock:     stock.Client{Server: strings.TrimRight(s, "/") + "/", Client: httpClient},
	}
}
func autenticar(ctx context.Context, r *http.Request) error {
	a, _ := ctx.Value(claveCredencial).(Credencial)
	r.Header.Set("Authorization", "Bearer "+a.JWT)
	r.Header.Set("X-Request-Id", a.RequestID)
	return nil
}
func (c Cliente) Obtener(ctx context.Context, id string) (service.Producto, error) {
	var p service.Producto
	r, e := c.productos.ObtenerProducto(ctx, id, autenticar)
	if e != nil {
		return p, service.ErrExterno
	}
	defer r.Body.Close()
	if r.StatusCode == 404 {
		return p, service.ErrNoExiste
	}
	if r.StatusCode != 200 {
		return p, service.ErrExterno
	}
	if e = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&p); e != nil {
		return p, service.ErrExterno
	}
	return p, nil
}
func (c Cliente) Descontar(ctx context.Context, clave string, items []service.Item) error {
	body := stock.DescontarJSONRequestBody{Productos: make([]stock.Item, 0, len(items))}
	for _, item := range items {
		body.Productos = append(body.Productos, stock.Item{ProductoId: item.ProductoID, Cantidad: item.Cantidad})
	}
	a, _ := ctx.Value(claveCredencial).(Credencial)
	r, e := c.stock.Descontar(ctx, &stock.DescontarParams{IdempotencyKey: clave, XRequestId: a.RequestID}, body, autenticar)
	if e != nil {
		return service.ErrExterno
	}
	defer r.Body.Close()
	if r.StatusCode == 409 {
		return service.ErrStock
	}
	if r.StatusCode != 200 && r.StatusCode != 201 && r.StatusCode != 204 {
		return service.ErrExterno
	}
	return nil
}
