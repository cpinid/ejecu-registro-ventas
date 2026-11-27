package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"registro-ventas/internal/api"
	"registro-ventas/internal/clientes"
	"registro-ventas/internal/service"
	"registro-ventas/internal/store"
	"strings"
	"syscall"
	"time"
)

var version = "dev"

func exigir(k string) string {
	v := os.Getenv(k)
	if v == "" || strings.Contains(v, "replace-with-") {
		slog.Error("configuracion faltante", "variable", k)
		os.Exit(1)
	}
	return v
}
func problema(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"type": "about:blank", "title": http.StatusText(status), "status": status})
}
func configurar(m *http.ServeMux, prod bool, s service.Servicio) huma.API {
	c := huma.DefaultConfig("Registro de ventas", "1.0.0")
	c.DocsPath = "/api/docs"
	c.OpenAPIPath = "/api/openapi"
	c.SchemasPath = "/api/schemas"
	if prod {
		c.DocsPath = ""
		c.OpenAPIPath = ""
		c.SchemasPath = ""
	}
	c.Components.SecuritySchemes = map[string]*huma.SecurityScheme{"sesion": {Type: "apiKey", In: "cookie", Name: "sesion"}}
	a := humago.New(m, c)
	api.Registrar(a, s)
	return a
}

type respuesta struct {
	http.ResponseWriter
	estado int
}

func (r *respuesta) WriteHeader(n int) { r.estado = n; r.ResponseWriter.WriteHeader(n) }
func main() {
	if len(os.Args) > 1 && os.Args[1] == "openapi" {
		a := configurar(http.NewServeMux(), false, service.Servicio{})
		b, e := a.OpenAPI().YAML()
		if e != nil {
			panic(e)
		}
		fmt.Print(string(b))
		return
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	env := exigir("API_ENV")
	if env != "dev" && env != "test" && env != "prod" {
		os.Exit(1)
	}
	if exigir("API_PORT") != "8080" || exigir("TZ") != "America/Santiago" {
		os.Exit(1)
	}
	db, e := pgxpool.New(ctx, exigir("DATABASE_URL"))
	if e != nil {
		os.Exit(1)
	}
	defer db.Close()
	jwksURL := exigir("AUTH_JWKS_URL")
	issuer := exigir("AUTH_ISSUER")
	aud := exigir("AUTH_AUDIENCE")
	origen := exigir("FRONTEND_URL")
	p := exigir("PRODUCTOS_URL")
	st := exigir("STOCK_URL")
	for _, u := range []string{jwksURL, origen, p, st} {
		x, e := url.Parse(u)
		if e != nil || x.Host == "" || (x.Scheme != "http" && x.Scheme != "https") {
			os.Exit(1)
		}
	}
	llaves, e := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if e != nil {
		slog.Error("JWKS no disponible")
		os.Exit(1)
	}
	cliente := clientes.Nuevo(p, st)
	mux := http.NewServeMux()
	configurar(mux, env == "prod", service.Servicio{Datos: store.Store{Pool: db}, Productos: cliente, Stock: cliente})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		c, fin := context.WithTimeout(r.Context(), 2*time.Second)
		defer fin()
		if db.Ping(c) != nil {
			problema(w, 503)
			return
		}
		w.WriteHeader(200)
	})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rw := &respuesta{w, 200}
		id := r.Header.Get("X-Request-Id")
		if len(id) == 0 || len(id) > 128 {
			var b [16]byte
			_, _ = rand.Read(b[:])
			id = hex.EncodeToString(b[:])
		}
		w.Header().Set("X-Request-Id", id)
		defer func() {
			slog.Info("solicitud", "servicio", "registro-ventas", "version", version, "request_id", id, "metodo", r.Method, "ruta", r.URL.Path, "estado", rw.estado, "duracion_ms", time.Since(inicio).Milliseconds())
		}()
		if !strings.HasPrefix(r.URL.Path, "/api/ventas") {
			mux.ServeHTTP(rw, r)
			return
		}
		if r.Method == http.MethodPost && r.Header.Get("Origin") != "" && r.Header.Get("Origin") != origen {
			problema(rw, 403)
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			if c, e := r.Cookie("sesion"); e == nil {
				token = c.Value
			}
		}
		t, e := jwt.Parse(token, llaves.Keyfunc, jwt.WithValidMethods([]string{"RS256", "EdDSA"}), jwt.WithIssuer(issuer), jwt.WithAudience(aud), jwt.WithExpirationRequired())
		if e != nil || !t.Valid {
			problema(rw, 401)
			return
		}
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			problema(rw, 401)
			return
		}
		u, e := claims.GetSubject()
		if e != nil || len(u) != 36 {
			problema(rw, 401)
			return
		}
		permiso := "ventas:consultar"
		if r.Method == http.MethodPost {
			permiso = "ventas:operar"
		}
		permisos, _ := claims["permisos"].([]any)
		autorizado := false
		for _, v := range permisos {
			if v == permiso {
				autorizado = true
			}
		}
		if !autorizado {
			problema(rw, 403)
			return
		}
		c := api.ConUsuario(r.Context(), u)
		c = clientes.ConCredencial(c, clientes.Credencial{JWT: token, RequestID: id})
		mux.ServeHTTP(rw, r.WithContext(c))
	})
	servidor := &http.Server{Addr: ":8080", Handler: h, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		c, fin := context.WithTimeout(context.Background(), 10*time.Second)
		defer fin()
		_ = servidor.Shutdown(c)
	}()
	if e = servidor.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
		slog.Error("servidor detenido")
		os.Exit(1)
	}
}
