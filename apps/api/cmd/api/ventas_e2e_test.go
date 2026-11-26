//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"registro-ventas/internal/clientes"
	"registro-ventas/internal/service"
	"registro-ventas/internal/store"
)

func TestVentaE2E(t *testing.T) {
	ctx := t.Context()
	pg, err := postgres.Run(ctx, "postgres:17.11@sha256:ff87138949ad9a46f32ff23bd6852b52ce024a13dffe9b4087be98292cdd3b4f", postgres.BasicWaitStrategies())
	testcontainers.CleanupContainer(t, pg)
	if err != nil {
		t.Fatal(err)
	}
	url, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	sqlDB := stdlib.OpenDBFromPool(db)
	defer sqlDB.Close()
	if err := goose.UpContext(ctx, sqlDB, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}

	auth, firmar := identidadPrueba(t)
	token := firmar("ventas:consultar", "ventas:operar")
	var descuentos atomic.Int32
	proveedores := http.NewServeMux()
	proveedores.HandleFunc("GET /api/productos/33333333-3333-4333-8333-333333333333", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"nombre":"Cuaderno","precio_venta":1500,"costo":900,"vendible":true}`)
	})
	proveedores.HandleFunc("POST /api/stock/salidas", func(w http.ResponseWriter, r *http.Request) {
		var salida struct{ Productos []service.Item }
		if !strings.HasPrefix(r.Header.Get("Idempotency-Key"), "venta-") || json.NewDecoder(r.Body).Decode(&salida) != nil ||
			len(salida.Productos) != 1 || salida.Productos[0].Cantidad != 2 {
			http.Error(w, "salida inválida", 400)
			return
		}
		descuentos.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token || r.Header.Get("X-Request-Id") == "" {
			http.Error(w, "faltan credenciales o trazabilidad", 401)
			return
		}
		proveedores.ServeHTTP(w, r)
	}))
	defer mock.Close()
	c := clientes.Nuevo(mock.URL, mock.URL)
	mux := http.NewServeMux()
	configurar(mux, false, service.Servicio{Datos: store.Store{Pool: db}, Productos: c, Stock: c}, auth)
	server := httptest.NewServer(registrarSolicitudes(mux))
	defer server.Close()
	pedir := func(metodo, path, body string, esperado int, destino any) {
		t.Helper()
		r, err := http.NewRequestWithContext(ctx, metodo, server.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Idempotency-Key", "venta-e2e")
		resp, err := server.Client().Do(r)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != esperado {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("%s %s: HTTP %d: %s", metodo, path, resp.StatusCode, b)
		}
		if destino != nil {
			if err := json.NewDecoder(resp.Body).Decode(destino); err != nil {
				t.Fatal(err)
			}
		}
	}
	var venta, repetida, consultada service.Venta
	pedir("POST", "/api/ventas", cuerpoVenta, 201, &venta)
	pedir("POST", "/api/ventas", cuerpoVenta, 201, &repetida)
	pedir("POST", "/api/ventas", strings.Replace(cuerpoVenta, "debito", "credito", 1), 409, nil)
	pedir("GET", "/api/ventas/"+venta.ID, "", 200, &consultada)
	var lista []service.Venta
	pedir("GET", "/api/ventas", "", 200, &lista)
	if venta.ID == "" || venta.Estado != "confirmada" || venta.Total != 3000 || len(venta.Detalles) != 1 ||
		repetida.ID != venta.ID || consultada.ID != venta.ID || len(lista) != 1 || lista[0].ID != venta.ID || descuentos.Load() != 1 {
		t.Fatalf("Venta inconsistente: %+v, repetida=%s, consultada=%s, lista=%v, descuentos=%d", venta, repetida.ID, consultada.ID, lista, descuentos.Load())
	}
}
