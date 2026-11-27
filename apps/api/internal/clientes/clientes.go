// Adaptadores PROVISIONALES: sustituir por clientes oapi-codegen de los contratos oficiales.
package clientes

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"registro-ventas/internal/service"
	"time"
)

type claveContexto int

const claveCredencial claveContexto = 0

type Credencial struct{ JWT, RequestID string }

func ConCredencial(ctx context.Context, c Credencial) context.Context {
	return context.WithValue(ctx, claveCredencial, c)
}

type Cliente struct {
	ProductosURL, StockURL string
	HTTP                   *http.Client
}

func Nuevo(p, s string) Cliente {
	return Cliente{p, s, &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (c Cliente) llamar(ctx context.Context, metodo, url, clave string, body []byte) (*http.Response, error) {
	r, e := http.NewRequestWithContext(ctx, metodo, url, bytes.NewReader(body))
	if e != nil {
		return nil, service.ErrExterno
	}
	a, _ := ctx.Value(claveCredencial).(Credencial)
	r.Header.Set("Authorization", "Bearer "+a.JWT)
	r.Header.Set("X-Request-Id", a.RequestID)
	r.Header.Set("Content-Type", "application/json")
	if clave != "" {
		r.Header.Set("Idempotency-Key", clave)
	}
	resp, e := c.HTTP.Do(r)
	if e != nil {
		return nil, service.ErrExterno
	}
	return resp, nil
}
func (c Cliente) Obtener(ctx context.Context, id string) (service.Producto, error) {
	var p service.Producto
	r, e := c.llamar(ctx, "GET", c.ProductosURL+"/api/productos/"+id, "", nil)
	if e != nil {
		return p, e
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
	b, e := json.Marshal(map[string]any{"productos": items})
	if e != nil {
		return e
	}
	r, e := c.llamar(ctx, "POST", c.StockURL+"/api/stock/salidas", clave, b)
	if e != nil {
		return e
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
