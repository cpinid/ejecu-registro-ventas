package service

import (
	"context"
	"errors"
	"testing"
)

func TestEntradaInvalida(t *testing.T) {
	s := Servicio{}
	_, e := s.Crear(context.Background(), "usuario", "clave", Solicitud{})
	if !errors.Is(e, ErrInvalido) {
		t.Fatal(e)
	}
}
