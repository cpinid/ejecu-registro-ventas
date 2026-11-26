package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"registro-ventas/internal/service"
)

type Entrada struct {
	Clave string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	Body  service.Solicitud
}
type Salida struct{ Body service.Venta }
type Consulta struct {
	ID string `path:"id" format:"uuid"`
}
type Paginacion struct {
	Limite         int32 `query:"limite" default:"50" minimum:"1" maximum:"100"`
	Desplazamiento int32 `query:"desplazamiento" default:"0" minimum:"0"`
}
type Lista struct{ Body []service.Venta }
type claveUsuario struct{}

func ConUsuario(ctx context.Context, u string) context.Context {
	return context.WithValue(ctx, claveUsuario{}, u)
}
func problema(e error) error {
	switch {
	case errors.Is(e, service.ErrInvalido):
		return huma.Error422UnprocessableEntity(e.Error())
	case errors.Is(e, service.ErrNoExiste):
		return huma.Error404NotFound(e.Error())
	case errors.Is(e, service.ErrConflicto), errors.Is(e, service.ErrStock):
		return huma.Error409Conflict(e.Error())
	case errors.Is(e, service.ErrExterno):
		return huma.Error503ServiceUnavailable(e.Error())
	default:
		return huma.Error500InternalServerError("error interno")
	}
}
func Registrar(a huma.API, s service.Servicio, auth Autenticacion) {
	crear := auth.proteger(a, "ventas:operar", huma.Operation{
		OperationID: "crear-venta", Method: http.MethodPost, Path: "/api/ventas",
		DefaultStatus: http.StatusCreated, Errors: []int{401, 403, 404, 409, 422, 503},
	})
	huma.Register(a, crear, func(ctx context.Context, in *Entrada) (*Salida, error) {
		u, _ := ctx.Value(claveUsuario{}).(string)
		v, e := s.Crear(ctx, u, in.Clave, in.Body)
		if e != nil {
			return nil, problema(e)
		}
		return &Salida{v}, nil
	})
	consultar := auth.proteger(a, "ventas:consultar", huma.Operation{
		OperationID: "consultar-venta", Method: http.MethodGet, Path: "/api/ventas/{id}",
		Errors: []int{401, 403, 404},
	})
	huma.Register(a, consultar, func(ctx context.Context, in *Consulta) (*Salida, error) {
		v, e := s.PorID(ctx, in.ID)
		if e != nil {
			return nil, problema(e)
		}
		return &Salida{v}, nil
	})
	listar := auth.proteger(a, "ventas:consultar", huma.Operation{
		OperationID: "listar-ventas", Method: http.MethodGet, Path: "/api/ventas",
		Errors: []int{401, 403},
	})
	huma.Register(a, listar, func(ctx context.Context, in *Paginacion) (*Lista, error) {
		v, e := s.Listar(ctx, in.Limite, in.Desplazamiento)
		if e != nil {
			return nil, problema(e)
		}
		return &Lista{v}, nil
	})
}
