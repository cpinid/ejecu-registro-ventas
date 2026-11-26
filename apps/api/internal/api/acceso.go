package api

import (
	"net/http"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/golang-jwt/jwt/v5"
	"registro-ventas/internal/clientes"
)

type Autenticacion struct {
	Llaves                    jwt.Keyfunc
	Emisor, Audiencia, Origen string
}

type sesion struct {
	jwt.RegisteredClaims
	Permisos []string `json:"permisos"`
}

func (auth Autenticacion) proteger(a huma.API, permiso string, op huma.Operation) huma.Operation {
	op.Security = []map[string][]string{{"sesion": {}}}
	op.Extensions = map[string]any{"x-permiso": permiso}
	op.Middlewares = huma.Middlewares{func(ctx huma.Context, next func(huma.Context)) {
		r, _ := humago.Unwrap(ctx)
		if r.Method == http.MethodPost && r.Header.Get("Origin") != "" && r.Header.Get("Origin") != auth.Origen {
			_ = huma.WriteErr(a, ctx, http.StatusForbidden, "Origen no permitido")
			return
		}
		var token string
		if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
			token = strings.TrimPrefix(header, "Bearer ")
		} else if cookie, err := r.Cookie("sesion"); header == "" && err == nil {
			token = cookie.Value
		}
		claims := &sesion{}
		_, err := jwt.ParseWithClaims(token, claims, auth.Llaves, jwt.WithValidMethods([]string{"RS256", "EdDSA"}),
			jwt.WithIssuer(auth.Emisor), jwt.WithAudience(auth.Audiencia), jwt.WithExpirationRequired())
		if err != nil || len(claims.Subject) != 36 {
			_ = huma.WriteErr(a, ctx, http.StatusUnauthorized, "Sesión inválida")
			return
		}
		if !slices.Contains(claims.Permisos, permiso) {
			_ = huma.WriteErr(a, ctx, http.StatusForbidden, "Permiso insuficiente")
			return
		}
		c := ConUsuario(ctx.Context(), claims.Subject)
		next(huma.WithContext(ctx, clientes.ConCredencial(c, clientes.Credencial{JWT: token, RequestID: r.Header.Get("X-Request-Id")})))
	}}
	return op
}
