package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"registro-ventas/internal/api"
	"registro-ventas/internal/service"
)

const vendedor = "22222222-2222-4222-8222-222222222222"
const cuerpoVenta = `{"medio_pago":"debito","productos":[{"producto_id":"33333333-3333-4333-8333-333333333333","cantidad":2}]}`

func identidadPrueba(t *testing.T) (api.Autenticacion, func(...string) string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	auth := api.Autenticacion{Emisor: "prueba", Audiencia: "ventas", Origen: "http://frontend.test",
		Llaves: func(*jwt.Token) (any, error) { return pub, nil }}
	return auth, func(permisos ...string) string {
		token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{
			"sub": vendedor, "iss": auth.Emisor, "aud": auth.Audiencia,
			"exp": time.Now().Add(time.Minute).Unix(), "permisos": permisos,
		}).SignedString(priv)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
}

type lecturasPrueba struct{ service.Datos }

func (lecturasPrueba) PorID(context.Context, string) (service.Venta, error) {
	return service.Venta{}, service.ErrNoExiste
}
func (lecturasPrueba) Listar(context.Context, int32, int32) ([]service.Venta, error) {
	return []service.Venta{}, nil
}

func TestAcceso(t *testing.T) {
	auth, firmar := identidadPrueba(t)
	mux := http.NewServeMux()
	configurar(mux, false, service.Servicio{Datos: lecturasPrueba{}}, auth)
	// Un detalle ausente llega al store (404) y un alta sin cuerpo llega al validador (422).
	for _, rol := range []struct {
		nombre, token          string
		listar, detalle, crear int
	}{
		{"anonimo", "", 401, 401, 401},
		{"sin-permisos", firmar(), 403, 403, 403},
		{"consulta", firmar("ventas:consultar"), 200, 404, 403},
		{"operacion", firmar("ventas:operar"), 403, 403, 422},
		{"ambos", firmar("ventas:consultar", "ventas:operar"), 200, 404, 422},
	} {
		for _, ruta := range []struct {
			metodo, path string
			estado       int
		}{
			{"GET", "/api/ventas", rol.listar}, {"GET", "/api/ventas/" + vendedor, rol.detalle},
			{"POST", "/api/ventas", rol.crear},
		} {
			t.Run(rol.nombre+"/"+ruta.metodo+ruta.path, func(t *testing.T) {
				r := httptest.NewRequest(ruta.metodo, ruta.path, strings.NewReader("{}"))
				r.Header.Set("Idempotency-Key", "prueba")
				r.AddCookie(&http.Cookie{Name: "sesion", Value: rol.token})
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, r)
				if w.Code != ruta.estado {
					t.Fatalf("HTTP %d, esperado %d: %s", w.Code, ruta.estado, w.Body.String())
				}
			})
		}
	}
}
