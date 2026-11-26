// Simulador exclusivo de desarrollo; no se publica como servicio real.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const producto = "33333333-3333-4333-8333-333333333333"
const usuario = "22222222-2222-4222-8222-222222222222"
const emisor = "registro-ventas-dev"
const audiencia = "registro-ventas"

func responder(w http.ResponseWriter, estado int, datos any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	if datos != nil {
		_ = json.NewEncoder(w).Encode(datos)
	}
}
func main() {
	publica, privada, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	codificar := base64.RawURLEncoding.EncodeToString
	token := func() string {
		cabecera, _ := json.Marshal(map[string]string{"alg": "EdDSA", "kid": "dev", "typ": "JWT"})
		ahora := time.Now()
		datos, _ := json.Marshal(map[string]any{"iss": emisor, "aud": audiencia, "sub": usuario, "iat": ahora.Unix(), "exp": ahora.Add(time.Hour).Unix(), "permisos": []string{"ventas:operar", "ventas:consultar"}})
		mensaje := codificar(cabecera) + "." + codificar(datos)
		return mensaje + "." + codificar(ed25519.Sign(privada, []byte(mensaje)))
	}
	autorizado := func(r *http.Request) bool {
		t := strings.Split(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), ".")
		if len(t) != 3 {
			return false
		}
		firma, e := base64.RawURLEncoding.DecodeString(t[2])
		if e != nil || !ed25519.Verify(publica, []byte(t[0]+"."+t[1]), firma) {
			return false
		}
		b, e := base64.RawURLEncoding.DecodeString(t[1])
		if e != nil {
			return false
		}
		var c struct {
			Iss string `json:"iss"`
			Aud string `json:"aud"`
			Exp int64  `json:"exp"`
		}
		return json.Unmarshal(b, &c) == nil && c.Iss == emisor && c.Aud == audiencia && c.Exp > time.Now().Unix()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("GET /jwks", func(w http.ResponseWriter, r *http.Request) {
		responder(w, 200, map[string]any{"keys": []any{map[string]string{"kty": "OKP", "crv": "Ed25519", "alg": "EdDSA", "use": "sig", "kid": "dev", "x": codificar(publica)}}})
	})
	mux.HandleFunc("GET /token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		responder(w, 200, map[string]string{"token": token()})
	})
	mux.HandleFunc("GET /sesion-dev", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		http.SetCookie(w, &http.Cookie{Name: "sesion", Value: token(), Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: 3600})
		http.Redirect(w, r, "http://localhost:5140", http.StatusSeeOther)
	})
	var mu sync.Mutex
	cantidad := int64(10)
	type resultado struct {
		Huella string
		Estado int
	}
	resultados := map[string]resultado{}
	mux.HandleFunc("GET /api/productos/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !autorizado(r) {
			responder(w, 401, nil)
			return
		}
		if r.PathValue("id") != producto {
			responder(w, 404, nil)
			return
		}
		responder(w, 200, map[string]any{"nombre": "Cuaderno de prueba", "precio_venta": 1500, "costo": 900, "vendible": true})
	})
	mux.HandleFunc("GET /inventario-dev", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		responder(w, 200, map[string]any{"producto_id": producto, "cantidad": cantidad})
	})
	mux.HandleFunc("POST /api/stock/salidas", func(w http.ResponseWriter, r *http.Request) {
		if !autorizado(r) {
			responder(w, 401, nil)
			return
		}
		clave := r.Header.Get("Idempotency-Key")
		if clave == "" || r.Header.Get("X-Request-Id") == "" {
			responder(w, 400, nil)
			return
		}
		var entrada struct {
			Productos []struct {
				ID       string `json:"producto_id"`
				Cantidad int64  `json:"cantidad"`
			} `json:"productos"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 65536)).Decode(&entrada) != nil || len(entrada.Productos) == 0 {
			responder(w, 400, nil)
			return
		}
		var total int64
		for _, i := range entrada.Productos {
			if i.ID != producto || i.Cantidad <= 0 || i.Cantidad > 1000000 || total > 1000000-i.Cantidad {
				responder(w, 422, nil)
				return
			}
			total += i.Cantidad
		}
		huella := fmt.Sprint(total)
		mu.Lock()
		defer mu.Unlock()
		if previo, ok := resultados[clave]; ok {
			if previo.Huella != huella {
				responder(w, 422, nil)
				return
			}
			responder(w, previo.Estado, nil)
			return
		}
		estado := 204
		if total > cantidad {
			estado = 409
		} else {
			cantidad -= total
		}
		resultados[clave] = resultado{huella, estado}
		responder(w, estado, nil)
	})
	log.Print("Mocks locales listos")
	log.Fatal((&http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}).ListenAndServe())
}
