package store

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"registro-ventas/internal/repository"
	"registro-ventas/internal/service"
)

type Store struct{ Pool *pgxpool.Pool }

func cargar(ctx context.Context, q *repository.Queries, r repository.Venta) (service.Venta, error) {
	ds, e := q.Detalles(ctx, r.ID)
	if e != nil {
		return service.Venta{}, e
	}
	v := service.Venta{ID: r.ID, VendedorID: r.VendedorID, Fecha: r.Fecha.Time, MedioPago: r.MedioPago, Estado: r.Estado, Total: r.Total, Huella: r.Huella, Detalles: []service.Detalle{}}
	for _, d := range ds {
		v.Detalles = append(v.Detalles, service.Detalle{ProductoID: d.ProductoID, Nombre: d.Nombre, Cantidad: d.Cantidad, PrecioUnitario: d.PrecioUnitario, CostoUnitario: d.CostoUnitario, Subtotal: d.Subtotal})
	}
	return v, nil
}
func (s Store) PorClave(ctx context.Context, u, k string) (service.Venta, error) {
	q := repository.New(s.Pool)
	r, e := q.PorClave(ctx, repository.PorClaveParams{VendedorID: u, Clave: k})
	if errors.Is(e, pgx.ErrNoRows) {
		return service.Venta{}, service.ErrNoExiste
	}
	if e != nil {
		return service.Venta{}, e
	}
	return cargar(ctx, q, r)
}
func (s Store) PorID(ctx context.Context, id string) (service.Venta, error) {
	q := repository.New(s.Pool)
	r, e := q.PorID(ctx, id)
	if errors.Is(e, pgx.ErrNoRows) {
		return service.Venta{}, service.ErrNoExiste
	}
	if e != nil {
		return service.Venta{}, e
	}
	return cargar(ctx, q, r)
}
func (s Store) Listar(ctx context.Context, lim, off int32) ([]service.Venta, error) {
	q := repository.New(s.Pool)
	rs, e := q.Listar(ctx, repository.ListarParams{Limit: lim, Offset: off})
	if e != nil {
		return nil, e
	}
	out := []service.Venta{}
	for _, r := range rs {
		v, e := cargar(ctx, q, r)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}
func (s Store) Crear(ctx context.Context, v service.Venta, k string) (service.Venta, error) {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return service.Venta{}, e
	}
	defer tx.Rollback(ctx)
	q := repository.New(tx)
	r, e := q.CrearVenta(ctx, repository.CrearVentaParams{VendedorID: v.VendedorID, Clave: k, Huella: v.Huella, MedioPago: v.MedioPago, Total: v.Total})
	if errors.Is(e, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return s.PorClave(ctx, v.VendedorID, k)
	}
	if e != nil {
		return service.Venta{}, e
	}
	for _, d := range v.Detalles {
		e = q.CrearDetalle(ctx, repository.CrearDetalleParams{VentaID: r.ID, ProductoID: d.ProductoID, Nombre: d.Nombre, Cantidad: d.Cantidad, PrecioUnitario: d.PrecioUnitario, CostoUnitario: d.CostoUnitario, Subtotal: d.Subtotal})
		if e != nil {
			return service.Venta{}, e
		}
	}
	if e = q.Auditar(ctx, repository.AuditarParams{VentaID: r.ID, UsuarioID: v.VendedorID, Accion: "pendiente"}); e != nil {
		return service.Venta{}, e
	}
	v, e = cargar(ctx, q, r)
	if e != nil {
		return service.Venta{}, e
	}
	return v, tx.Commit(ctx)
}
func (s Store) Estado(ctx context.Context, v service.Venta, estado string) (service.Venta, error) {
	tx, e := s.Pool.Begin(ctx)
	if e != nil {
		return service.Venta{}, e
	}
	defer tx.Rollback(ctx)
	q := repository.New(tx)
	r, e := q.CambiarEstado(ctx, repository.CambiarEstadoParams{ID: v.ID, Estado: estado})
	if errors.Is(e, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		actual, e := s.PorID(ctx, v.ID)
		if e != nil {
			return service.Venta{}, e
		}
		if actual.Estado != estado {
			return service.Venta{}, service.ErrConflicto
		}
		return actual, nil
	}
	if e != nil {
		return service.Venta{}, e
	}
	if e = q.Auditar(ctx, repository.AuditarParams{VentaID: v.ID, UsuarioID: v.VendedorID, Accion: estado}); e != nil {
		return service.Venta{}, e
	}
	v, e = cargar(ctx, q, r)
	if e != nil {
		return service.Venta{}, e
	}
	return v, tx.Commit(ctx)
}
